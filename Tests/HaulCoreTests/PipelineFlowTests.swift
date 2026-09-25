import Foundation
import Testing
@testable import HaulCore

/// Whole page downloads — info, tracks, subtitles, danmaku, cover, download and a real ffmpeg mux — over the stub
/// network. Serialized: the pipeline changes the working directory.
@Suite(.serialized) struct PipelineFlowTests {
    init() { _ = TestEnv.home }

    func pipeline(_ configure: (inout DownloadOptions) -> Void) async throws -> (DownloadPipeline, String) {
        var o = DownloadOptions()
        o.workDir = try makeTempDir("work")
        o.url = "BV1xx411c7mD"   // a bilibili link unless a test gives another
        o.api = .tv   // no login check unless a test asks for the web API
        configure(&o)
        return (try await DownloadPipeline(options: o, http: Stub.client), o.workDir)
    }

    // MARK: stream ordering

    func tracks() -> [VideoTrack] {
        [VideoTrack(id: "80", quality: "1080P", baseURL: "", codec: "AVC", bandwidth: 2000, duration: 1),
         VideoTrack(id: "80", quality: "1080P", baseURL: "", codec: "HEVC", bandwidth: 1000, duration: 1),
         VideoTrack(id: "120", quality: "4K", baseURL: "", codec: "HEVC", bandwidth: 9000, duration: 1),
         VideoTrack(id: "32", quality: "480P", baseURL: "", codec: "AVC", bandwidth: 500, duration: 1)]
    }

    @Test func ordersVideoByQualityThenBandwidth() async throws {
        let (p, _) = try await pipeline { $0.skipMux = true }
        #expect(p.sortVideo(tracks()).map { "\($0.id)/\($0.codec)" } == ["120/HEVC", "80/AVC", "80/HEVC", "32/AVC"])
    }

    @Test func codecAndQualityPriorities() async throws {
        // A codec priority alone outranks resolution: -c avc means "AVC, even if another codec goes higher".
        let (codec, _) = try await pipeline { $0.skipMux = true; $0.codecPriority = "avc,hevc" }
        #expect(codec.sortVideo(tracks()).map { "\($0.id)/\($0.codec)" } == ["80/AVC", "32/AVC", "120/HEVC", "80/HEVC"])
        let (quality, _) = try await pipeline { $0.skipMux = true; $0.qualityPriority = "480p, 1080P"; $0.codecPriority = "hevc" }
        #expect(quality.sortVideo(tracks()).map { "\($0.id)/\($0.codec)" } == ["32/AVC", "80/HEVC", "80/AVC", "120/HEVC"])
        let (codecFirst, _) = try await pipeline {
            $0.skipMux = true; $0.qualityPriority = "1080P"; $0.codecPriority = "avc"; $0.codecPriorityFirst = true
        }
        #expect(codecFirst.sortVideo(tracks()).map { "\($0.id)/\($0.codec)" } == ["80/AVC", "32/AVC", "80/HEVC", "120/HEVC"])
        let (ascending, _) = try await pipeline { $0.skipMux = true; $0.videoAscending = true }
        #expect(ascending.sortVideo(tracks()).map { "\($0.id)/\($0.codec)" } == ["32/AVC", "80/HEVC", "80/AVC", "120/HEVC"])
    }

    @Test func ordersAudioByCodecThenBandwidth() async throws {
        let audio = [AudioTrack(id: "1", quality: "", baseURL: "", codec: "M4A", bandwidth: 130, duration: 1),
                     AudioTrack(id: "2", quality: "", baseURL: "", codec: "E-AC-3", bandwidth: 640, duration: 1),
                     AudioTrack(id: "3", quality: "", baseURL: "", codec: "M4A", bandwidth: 320, duration: 1)]
        let (plain, _) = try await pipeline { $0.skipMux = true }
        #expect(plain.sortAudio(audio).map(\.id) == ["2", "3", "1"])
        let (m4a, _) = try await pipeline { $0.skipMux = true; $0.codecPriority = "m4a,eac3" }
        #expect(m4a.sortAudio(audio).map(\.id) == ["3", "1", "2"])
    }

    // MARK: bilibili

    @Test(.enabled(if: TestEnv.hasFFmpeg)) func bilibiliVideoEndToEnd() async throws {
        let media = try await TestMedia.make()
        try Stub.fixture("api.bilibili.com/x/web-interface/nav", "nav-logged-out.json")
        try Stub.fixture("x/web-interface/view?aid=626497566", "view-single.json")
        try Stub.fixture("x/player/wbi/playurl?", "playurl-web-dash.json")
        Stub.on("www.bilibili.com/video/av626497566") { _ in .status(200) }
        Stub.on("x/player/wbi/v2?cid=220355130&aid=626497566") { _ in
            .json(#"{"code":0,"data":{"view_points":[{"content":"开场","from":0,"to":1}]}}"#)
        }
        Stub.on("comment.bilibili.com/220355130.xml") { _ in
            StubResponse(body: Data(#"<?xml version="1.0" encoding="UTF-8"?><i><d p="0.5,1,25,16777215,0,0,0,0">第一条</d><d p="0.8,1,25,255,0,0,0,0">第二条</d></i>"#.utf8))
        }
        Stub.on("hdslb.com/bfs/archive") { _ in StubResponse(body: media.cover) }
        // Audio ids are 302xx; anything else from the CDN is video.
        Stub.on(specificity: 50, { $0.host?.hasSuffix("bilivideo.com") == true }) { request in
            .ranged(request.url!.path.contains("-1-302") ? media.audio : media.video, for: request)
        }

        let (p, dir) = try await pipeline {
            $0.url = "BV1qt4y1X7TW"; $0.api = .web; $0.codecPriority = "avc"; $0.downloadDanmaku = true
        }
        #expect(p.isLoggedIn == false)
        try await p.run()

        let out = "\(dir)/【4K60帧】咬人猫最新单曲《dududu》有没有戳中你心~【BML2020单品】.mp4"
        let probe = try await Probe(out)
        #expect(probe.streams == ["video:h264", "audio:aac", "video:mjpeg"])
        #expect(probe.chapters == ["开场"])
        #expect(probe.tags["comment"] == "https://www.bilibili.com/video/BV1qt4y1X7TW/")
        #expect(probe.tags["artist"] == "BML制作指挥部")
        #expect(FileManager.default.fileExists(atPath: FilePattern.changeExtension(out, to: "ass")))
        #expect(FileManager.default.fileExists(atPath: FilePattern.changeExtension(out, to: "xml")))
        // Work directory cleaned up; the chosen streams went over the replacement host, as plain HTTP.
        #expect(!FileManager.default.fileExists(atPath: "\(dir)/626497566"))
        let cdn = Stub.requests("upos-sz-mirrorcoso1.bilivideo.com")
        #expect(!cdn.isEmpty && cdn.allSatisfy { $0.url!.scheme == "http" })
        #expect(cdn.contains { $0.url!.path.contains("-1-302") })
    }

    // MARK: other sites

    func youtubeInfo(host: String = "yt.test") -> VideoInfo {
        var tracks = ParsedTracks()
        tracks.video = [VideoTrack(id: "360024", quality: "360p", baseURL: "https://\(host)/videoplayback?itag=134",
                                   resolution: "64x64", fps: "10", codec: "AVC", bandwidth: 200, duration: 1)]
        tracks.audio = [AudioTrack(id: "140", quality: "", baseURL: "https://\(host)/videoplayback?itag=140", codec: "M4A", bandwidth: 130, duration: 1)]
        var page = Page(index: 1, aid: "ytTestVid01", cid: "ytTestVid01", epid: "", title: "Clip", duration: 1, resolution: "",
                        pubTime: 1790186492, cover: nil, desc: "d", ownerName: "Channel", ownerMid: "UC1")
        page.points = [ViewPoint(title: "Intro", start: 0, end: 1)]
        page.media = PageMedia(tracks: tracks,
                                   subtitles: [SubtitleInfo(language: "en", url: "https://\(host)/api/timedtext?lang=en", path: "ytTestVid01/ytTestVid01.en.srt")],
                                   headers: ["User-Agent": "UA-YT"], webpageURL: "https://www.youtube.com/watch?v=ytTestVid01")
        var info = VideoInfo(title: "YouTube Clip", desc: "d", cover: "https://\(host)/vi/ytTestVid01/maxresdefault.jpg", pubTime: 1790186492, pages: [page])
        info.site = .youtube
        return info
    }

    @Test(.enabled(if: TestEnv.hasFFmpeg)) func youtubeEndToEnd() async throws {
        let media = try await TestMedia.make()
        Stub.on("yt.test/videoplayback?itag=134") { DownloaderStubTests.rangeOnly(media.video, limit: 10 << 20, request: $0) }
        Stub.on("yt.test/videoplayback?itag=140") { DownloaderStubTests.rangeOnly(media.audio, limit: 10 << 20, request: $0) }
        Stub.on("yt.test/api/timedtext") { _ in .json(#"{"events":[{"tStartMs":0,"dDurationMs":900,"segs":[{"utf8":"Hello"}]}]}"#) }
        Stub.on("yt.test/vi/") { _ in StubResponse(body: media.cover) }

        let (p, dir) = try await pipeline { $0.cookie = "SESSDATA=secret"; $0.downloadDanmaku = true }
        let info = youtubeInfo()
        try await p.downloadPages(info, id: .link(site: .youtube, url: "https://www.youtube.com/watch?v=ytTestVid01"))

        let probe = try await Probe("\(dir)/YouTube Clip.mp4")
        #expect(probe.streams == ["video:h264", "audio:aac", "subtitle:mov_text", "video:mjpeg"])
        #expect(probe.chapters == ["Intro"])
        #expect(probe.tags["comment"] == "https://www.youtube.com/watch?v=ytTestVid01")
        // googlevideo rules: closed ranges only, the site's headers, never the bilibili cookie.
        let streams = Stub.requests("yt.test/videoplayback")
        #expect(!streams.isEmpty)
        #expect(streams.allSatisfy { Stub.range($0)?.to != nil })
        #expect(streams.allSatisfy { $0.value(forHTTPHeaderField: "User-Agent") == "UA-YT" && $0.value(forHTTPHeaderField: "Cookie") == nil })
        #expect(Stub.requests("yt.test/api/timedtext").allSatisfy { $0.value(forHTTPHeaderField: "Cookie") == nil })
        // No bilibili endpoints for another site's page (chapters, danmaku).
        #expect(Stub.requests("ytTestVid01").allSatisfy { $0.url!.host == "yt.test" })

        // The --json report: what was chosen and where it went.
        let doc = try #require(p.report.document)
        #expect(doc.ok && doc.site == .youtube && doc.title == "YouTube Clip" && doc.uploader == "Channel" && doc.pageCount == 1)
        let entry = try #require(doc.pages.first)
        #expect(entry.status == .downloaded && entry.selected && entry.id == "ytTestVid01")
        let file = Report.absolute("\(dir)/YouTube Clip.mp4")
        #expect(entry.file == file && file.hasPrefix("/private/"))
        #expect(doc.files == [file])
        #expect(entry.sizeBytes.map { $0 > 0 } == true && entry.subtitles == ["en"])
        #expect(entry.video?.map(\.quality) == ["360p"] && entry.audio?.map(\.codec) == ["M4A"] && entry.selectedVideo == 0)
        #expect(entry.video?.first?.url == nil)
        let decoded = try JSONDecoder().decode(Report.Document.self, from: Data(try #require(p.report.json()).utf8))
        #expect(decoded.pages.first?.file == entry.file)

        // Again: skipped, and nothing left behind (no subtitle or cover work files).
        let (again, _) = try await pipeline { $0.workDir = dir }
        try await again.downloadPages(info, id: .link(site: .youtube, url: "https://www.youtube.com/watch?v=ytTestVid01"))
        let skipped = try #require(again.report.document)
        #expect(skipped.pages.first?.status == .skipped && skipped.pages.first?.reason == "exists" && skipped.files == [file])
        #expect(try FileManager.default.contentsOfDirectory(atPath: dir) == ["YouTube Clip.mp4"])
    }

    @Test func infoReportsStreamsWithoutDownloading() async throws {
        let (p, dir) = try await pipeline { $0.onlyShowInfo = true; $0.showURLs = true }
        try await p.downloadPages(youtubeInfo(host: "yt-info.test"), id: .link(site: .youtube, url: "https://www.youtube.com/watch?v=ytTestVid01"))
        let entry = try #require(p.report.document?.pages.first)
        #expect(entry.status == .listed && entry.file == nil)
        #expect(entry.video?.first?.url == "https://yt-info.test/videoplayback?itag=134")
        #expect(entry.audio?.first?.index == 0 && entry.audio?.first?.bitrateKbps == 130)
        #expect(try FileManager.default.contentsOfDirectory(atPath: dir).isEmpty)
        #expect(Stub.requests("yt-info.test").isEmpty)
    }

    @Test func infoListsSubtitlesAndFiltersThemByLanguage() async throws {
        let (all, _) = try await pipeline { $0.onlyShowInfo = true }
        try await all.downloadPages(youtubeInfo(host: "yt-subs.test"), id: .link(site: .youtube, url: "u"))
        #expect(all.report.document?.pages.first?.subtitles == ["en"])
        let (french, _) = try await pipeline { $0.onlyShowInfo = true; $0.subtitleLanguages = "fr, de" }
        try await french.downloadPages(youtubeInfo(host: "yt-subs.test"), id: .link(site: .youtube, url: "u"))
        #expect(french.report.document?.pages.first?.subtitles == [])
        #expect(Stub.requests("yt-subs.test").isEmpty)
    }

    @Test func infoOnAListListsPagesUntilOneIsChosen() async throws {
        let (list, _) = try await pipeline { $0.onlyShowInfo = true }
        try await list.downloadPages(xInfo(pages: 3), id: .link(site: .x, url: "u"))
        let doc = try #require(list.report.document)
        #expect(doc.pageCount == 3 && doc.pages.allSatisfy { !$0.selected && $0.video == nil && $0.status == nil })
        let (one, _) = try await pipeline { $0.onlyShowInfo = true; $0.pages = "2" }
        try await one.downloadPages(xInfo(pages: 3), id: .link(site: .x, url: "u"))
        let pages = try #require(one.report.document?.pages)
        #expect(pages[1].selected && pages[1].status == .listed && pages[1].video?.isEmpty == false && pages[0].video == nil)
    }

    @Test func aPageSelectionOutOfRangeStillReportsThePages() async throws {
        let (p, _) = try await pipeline { $0.pages = "99" }
        await #expect { try await p.downloadPages(xInfo(pages: 2), id: .link(site: .x, url: "u")) }
            throws: { ($0 as? HaulError)?.kind == .input }
        let json = try JSON.parse(p.report.failureJSON(HaulError.input("x"), input: "u"))
        #expect(json["ok"].bool == false && json["pages"].array.count == 2 && json["pageCount"].int == 2)
    }

    @Test func streamIndexesAreChecked() async throws {
        let (p, _) = try await pipeline { $0.onlyShowInfo = true; $0.audioStream = 3 }
        await #expect { try await p.downloadPages(youtubeInfo(host: "yt-index.test"), id: .link(site: .youtube, url: "u")) }
            throws: { ($0 as? HaulError)?.kind == .input && $0.readableMessage.contains("--audio-stream 3") }
        #expect(p.report.failureJSON(HaulError.input("x"), input: "u").contains(#""ok" : false"#))
        let (ok, _) = try await pipeline { $0.onlyShowInfo = true; $0.audioStream = 0; $0.videoStream = 0 }
        try await ok.downloadPages(youtubeInfo(host: "yt-index.test"), id: .link(site: .youtube, url: "u"))
        #expect(ok.report.document?.pages.first?.selectedAudio == 0)
    }

    @Test func unsupportedLinksAreInputErrors() async throws {
        await #expect { try await pipeline { $0.url = "https://example.com/video/1" } } throws: { ($0 as? HaulError)?.kind == .input }
        let failure = Report(command: "info").failureJSON(HaulError.unsupported("https://example.com"), input: "https://example.com")
        let json = try JSON.parse(failure)
        #expect(json["ok"].bool == false && json["error"]["kind"].stringValue == "input" && json["error"]["exitCode"].int == 2)
    }

    /// A missing cover and a rate-limited subtitle cost only themselves.
    @Test(.enabled(if: TestEnv.hasFFmpeg)) func youtubeWithoutCoverOrSubtitle() async throws {
        let media = try await TestMedia.make()
        Stub.on("yt-flaky.test/videoplayback?itag=134") { DownloaderStubTests.rangeOnly(media.video, limit: 10 << 20, request: $0) }
        Stub.on("yt-flaky.test/videoplayback?itag=140") { DownloaderStubTests.rangeOnly(media.audio, limit: 10 << 20, request: $0) }
        Stub.on("yt-flaky.test/api/timedtext") { _ in .status(429) }
        Stub.on("yt-flaky.test/vi/") { _ in .status(404) }

        let (p, dir) = try await pipeline { _ in }
        try await p.downloadPages(youtubeInfo(host: "yt-flaky.test"), id: .link(site: .youtube, url: "https://www.youtube.com/watch?v=ytTestVid01"))
        #expect(try await Probe("\(dir)/YouTube Clip.mp4").streams == ["video:h264", "audio:aac"])
    }

    @Test func danmakuOnlyForAnotherSiteLeavesNothingBehind() async throws {
        let (p, dir) = try await pipeline { $0.danmakuOnly = true; $0.skipMux = true }
        try await p.downloadPages(youtubeInfo(host: "yt-none.test"), id: .link(site: .youtube, url: "https://www.youtube.com/watch?v=ytTestVid01"))
        #expect(try FileManager.default.contentsOfDirectory(atPath: dir).isEmpty)
        #expect(Stub.requests("yt-none.test").isEmpty)
    }

    func xInfo(pages count: Int) -> VideoInfo {
        let pages = (1...count).map { i in
            var tracks = ParsedTracks()
            tracks.video = [VideoTrack(id: "64000", quality: "64p", baseURL: "https://x.test/vid/avc1/64x64/\(i).mp4",
                                       resolution: "64x64", codec: "AVC", bandwidth: 300, duration: 1)]
            tracks.videoHasAudio = true
            var page = Page(index: i, aid: "17000000000000000\(i)", cid: "17000000000000000\(i)", epid: "", title: count > 1 ? "Video \(i)" : "Post",
                            duration: 1, resolution: "", pubTime: 1715756260, cover: "https://x.test/thumb/\(i).jpg", desc: "text", ownerName: "Someone")
            page.media = PageMedia(tracks: tracks, subtitles: [], headers: ["User-Agent": "UA-X"],
                                       webpageURL: "https://x.com/someone/status/1790637656616943991")
            return page
        }
        var info = VideoInfo(title: "Someone - A post", desc: "text", cover: count == 1 ? "https://x.test/thumb/1.jpg" : "",
                             pubTime: 1715756260, pages: pages)
        info.site = .x
        return info
    }

    func stubX(_ media: TestMedia) {
        Stub.on("x.test/vid/") { .ranged(media.combined, for: $0) }
        Stub.on("x.test/thumb/") { _ in StubResponse(body: media.cover) }
    }

    @Test(.enabled(if: TestEnv.hasFFmpeg)) func xVideoWithAudioInside() async throws {
        let media = try await TestMedia.make()
        stubX(media)
        let id = MediaID.link(site: .x, url: "https://x.com/someone/status/1790637656616943991")

        let (p, dir) = try await pipeline { _ in }
        try await p.downloadPages(xInfo(pages: 1), id: id)
        #expect(try await Probe("\(dir)/Someone - A post.mp4").streams == ["video:h264", "audio:aac", "video:mjpeg"])

        // --audio-only next to the finished video must not count as already downloaded.
        let (audioOnly, _) = try await pipeline { $0.workDir = dir; $0.audioOnly = true }
        try await audioOnly.downloadPages(xInfo(pages: 1), id: id)
        #expect(try await Probe("\(dir)/Someone - A post.m4a").streams == ["audio:aac", "video:mjpeg"])

        let (videoOnly, _) = try await pipeline { $0.workDir = dir; $0.videoOnly = true; $0.filePattern = "<title> silent" }
        try await videoOnly.downloadPages(xInfo(pages: 1), id: id)
        #expect(!(try await Probe("\(dir)/Someone - A post silent.mp4").streams.contains { $0.hasPrefix("audio") }))
    }

    @Test(.enabled(if: TestEnv.hasFFmpeg)) func xPostWithSeveralVideos() async throws {
        let media = try await TestMedia.make()
        stubX(media)
        let id = MediaID.link(site: .x, url: "https://x.com/someone/status/1790637656616943991")

        let (all, dir) = try await pipeline { _ in }
        try await all.downloadPages(xInfo(pages: 2), id: id)
        #expect(try FileManager.default.contentsOfDirectory(atPath: "\(dir)/Someone - A post").sorted() == ["[P1]Video 1.mp4", "[P2]Video 2.mp4"])

        let (second, dir2) = try await pipeline { $0.pages = "2" }
        try await second.downloadPages(xInfo(pages: 2), id: id)
        #expect(try FileManager.default.contentsOfDirectory(atPath: "\(dir2)/Someone - A post") == ["[P2]Video 2.mp4"])
    }

    // MARK: podcasts

    /// No ffmpeg needed: --skip-mux leaves the downloaded episode in its work directory.
    @Test func podcastEpisodeDownload() async throws {
        let eid = "6650a1b2c3d4e5f6a7b8c9e1"
        let audio = patternData(300_000)
        Stub.on("www.xiaoyuzhoufm.com/episode/\(eid)") { _ in
            StubResponse(body: Data(PodcastTests.episodeHTML(eid: eid, audio: "https://media.xyz-flow.test/ep.m4a").utf8))
        }
        Stub.on("media.xyz-flow.test/") { .ranged(audio, for: $0) }
        Stub.on("image.xyz.test/") { _ in StubResponse(body: Data("cover".utf8)) }

        let (p, dir) = try await pipeline { $0.url = "https://www.xiaoyuzhoufm.com/episode/\(eid)?s=share"; $0.skipMux = true; $0.cookie = "SESSDATA=secret" }
        #expect(p.isLoggedIn == nil)
        try await p.run()
        #expect(FileManager.default.contents(atPath: "\(dir)/\(eid)/\(eid).P1.\(eid).m4a") == audio)
        // The site's own headers, never the bilibili cookie.
        let media = Stub.requests("media.xyz-flow.test")
        #expect(!media.isEmpty)
        #expect(media.allSatisfy { $0.value(forHTTPHeaderField: "Cookie") == nil && $0.value(forHTTPHeaderField: "Referer") == "https://www.xiaoyuzhoufm.com/" })
        #expect(Stub.requests(eid).allSatisfy { $0.value(forHTTPHeaderField: "Cookie")?.contains("SESSDATA") != true })
    }

    @Test(.enabled(if: TestEnv.hasFFmpeg)) func podcastM4AEndToEnd() async throws {
        let media = try await TestMedia.make()
        let eid = "6650a1b2c3d4e5f6a7b8c9e2"
        Stub.on("www.xiaoyuzhoufm.com/episode/\(eid)") { _ in
            StubResponse(body: Data(PodcastTests.episodeHTML(eid: eid, audio: "https://media.xyz-e2e.test/ep.m4a").utf8))
        }
        Stub.on("media.xyz-e2e.test/") { .ranged(media.audio, for: $0) }
        Stub.on("image.xyz.test/") { _ in StubResponse(body: media.cover) }

        let (p, dir) = try await pipeline { $0.url = "https://www.xiaoyuzhoufm.com/episode/\(eid)" }
        try await p.run()
        let probe = try await Probe("\(dir)/E42 声音的故事.m4a")
        #expect(probe.streams == ["audio:aac", "video:mjpeg"])
        #expect(probe.tags["title"] == "E42 声音的故事" && probe.tags["album"] == "声东击西" && probe.tags["artist"] == "声动活泼")
        #expect(!FileManager.default.fileExists(atPath: "\(dir)/\(eid)"))
    }

    /// An MP3 episode stays MP3: ID3 tags and the cover, no MP4 around it. The picture the MP3 carries itself (PNG here) gives way to ours.
    @Test(.enabled(if: TestEnv.hasFFmpeg)) func podcastMP3EndToEnd() async throws {
        let media = try await TestMedia.make()
        let tmp = try makeTempDir("mp3")
        let r = try await Shell.run(try #require(TestEnv.ffmpeg), ["-v", "error", "-y", "-f", "lavfi", "-i", "sine=frequency=440:duration=1",
                                                                  "-f", "lavfi", "-i", "color=blue:size=16x16", "-frames:v", "1",
                                                                  "-map", "0", "-map", "1", "-c:a", "libmp3lame", "-c:v", "png",
                                                                  "-disposition:v:0", "attached_pic", "\(tmp)/a.mp3"], echo: false, capture: true)
        #expect(r.status == 0, "ffmpeg: \(r.errors)")
        let mp3 = try Data(contentsOf: URL(fileURLWithPath: "\(tmp)/a.mp3"))
        Stub.on("itunes.apple.com/lookup?id=1200361736") { _ in .json(PodcastTests.lookup) }
        Stub.on("podcast.test/redirect.mp3/") { .ranged(mp3, for: $0) }
        Stub.on("mzstatic.test/") { _ in StubResponse(body: media.cover) }

        let (p, dir) = try await pipeline { $0.url = "https://podcasts.apple.com/us/podcast/the-daily/id1200361736?i=1000671234567" }
        try await p.run()
        let probe = try await Probe("\(dir)/Newer.mp3")
        #expect(probe.streams == ["audio:mp3", "video:mjpeg"])
        #expect(probe.tags["title"] == "Newer" && probe.tags["album"] == "The Daily" && probe.tags["artist"] == "The New York Times")
        #expect(Stub.requests("podcast.test/redirect.mp3/").allSatisfy { $0.value(forHTTPHeaderField: "User-Agent") == Podcast.userAgent })
    }
}
