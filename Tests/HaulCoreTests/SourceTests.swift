import Foundation
import Testing
@testable import HaulCore

/// The per-site rules, one source at a time.
@Suite struct SourceTests {
    init() { _ = TestEnv.home }

    func request(_ page: Page) -> PageRequest {
        PageRequest(page: page, id: .video(aid: page.aid), api: .web)
    }

    func externalPage(video: Bool = true) -> Page {
        var tracks = ParsedTracks()
        if video { tracks.video = [VideoTrack(id: "1", quality: "", baseURL: "https://v.test/v", codec: "AVC", bandwidth: 1, duration: 1)] }
        tracks.audio = [AudioTrack(id: "a", quality: "", baseURL: "https://v.test/a", codec: "MP3", bandwidth: 1, duration: 1)]
        var page = Page(index: 1, aid: "p1", cid: "p1", epid: "", title: "T", duration: 1, resolution: "", pubTime: 0)
        page.media = PageMedia(tracks: tracks, subtitles: [SubtitleInfo(language: "en", url: "https://v.test/s", path: "p1/p1.en.srt")],
                                   headers: ["User-Agent": "UA"], webpageURL: "https://v.test/page")
        return page
    }

    @Test func pipelinePicksTheSourceBySite() async throws {
        var o = DownloadOptions()
        o.skipMux = true
        o.api = .tv
        o.url = "av170001"
        let p = try await DownloadPipeline(options: o, http: Stub.client)
        #expect(p.source(for: .bilibili) is BilibiliSource)
        #expect(p.source(for: .youtube) is YtDlpSource && p.source(for: .x) is YtDlpSource)
        #expect(p.source(for: .xiaoyuzhou) is PodcastSource && p.source(for: .applePodcasts) is PodcastSource)
        #expect(MediaID.link(site: .x, url: "u").site == .x && MediaID.video(aid: "1").site == .bilibili)
    }

    @Test func everySiteIsRecognisedByItsLinks() {
        for link in ["BV1qt4y1X7TW", "av170001", "ep123", "ss33073", "md28223066", "cheese/ep1234",
                     "https://www.bilibili.com/video/BV1qt4y1X7TW?p=2", "https://b23.tv/abc", "https://space.bilibili.com/1/favlist?fid=2",
                     "https://www.bilibili.tv/en/play/1/2", "m.bilibili.com/video/av1"] {
            #expect(Site.match(link)?.site == .bilibili, "\(link)")
        }
        #expect(Site.match("https://youtu.be/DdCEmlAydcw")?.site == .youtube)
        #expect(Site.match("https://x.com/a/status/1")?.site == .x)
        #expect(Site.match("https://www.xiaoyuzhoufm.com/episode/abcdef123456")?.site == .xiaoyuzhou)
        #expect(Site.match("https://podcasts.apple.com/us/podcast/id1")?.site == .applePodcasts)
        for link in ["", "hello", "https://example.com/video/BV1qt4y1X7TW", "https://notbilibili.com/video/1", "BV1short"] {
            #expect(Site.match(link) == nil, "\(link)")
        }
    }

    @Test func unsupportedLinksFailAsInput() async {
        await #expect { try await IDResolver.resolve("https://vimeo.com/1", session: Session()) }
            throws: { ($0 as? HaulError)?.kind == .input && $0.readableMessage.contains("Supported sites") }
    }

    @Test func bilibiliRewritesCDNHosts() {
        func finalize(_ configure: (inout DownloadOptions) -> Void, _ url: String, area: String = "") -> String? {
            var o = DownloadOptions()
            o.forceReplaceHost = false   // on by default: then every stream goes to the mirror
            configure(&o)
            var session = Session()
            session.area = area
            let source = BilibiliSource(http: Stub.client, session: session, options: o, preferredCodec: "")
            var video: VideoTrack? = VideoTrack(id: "80", quality: "", baseURL: url, codec: "AVC", bandwidth: 1, duration: 1)
            var audio: AudioTrack?
            source.finalizeURLs(video: &video, audio: &audio)
            return video?.baseURL
        }
        let backup = BilibiliSource.backupHost
        #expect(finalize({ _ in }, "http://1.2.3.4:4480/upgcxcode/v.m4s") == "http://\(backup)/upgcxcode/v.m4s")
        #expect(finalize({ $0.allowPCDN = true }, "http://1.2.3.4:4480/v.m4s") == "http://1.2.3.4:4480/v.m4s")
        #expect(finalize({ _ in }, "https://upos-hz.akamaized.net/v.m4s", area: "hk") == "https://\(backup)/v.m4s")
        #expect(finalize({ _ in }, "https://upos-hz.akamaized.net/v.m4s") == "https://upos-hz.akamaized.net/v.m4s")
        #expect(finalize({ $0.uposHost = "cn.test" }, "https://a.bilivideo.com/v.m4s") == "https://cn.test/v.m4s")
        #expect(finalize({ $0.forceReplaceHost = true }, "https://a.bilivideo.com/v.m4s") == "https://\(backup)/v.m4s")
        #expect(finalize({ $0.forceReplaceHost = true; $0.allowPCDN = true }, "http://1.2.3.4:4480/v.m4s") == "http://\(backup)/v.m4s")
    }

    @Test func bilibiliSendsItsCookie() {
        var session = Session()
        session.cookie = "SESSDATA=x"
        var o = DownloadOptions()
        o.forceHTTP = true
        let source = BilibiliSource(http: Stub.client, session: session, options: o, preferredCodec: "")
        var config = DownloadConfig()
        source.configure(&config, for: request(Page(index: 1, aid: "1", cid: "2", epid: "", title: "", duration: 0, resolution: "", pubTime: 0)))
        #expect(config.cookie == "SESSDATA=x" && config.forceHTTP && config.headers.isEmpty && config.maxRange == nil)
        #expect(source.hasDanmaku && !source.trustsStatedSizes)
        #expect(source.pageURL(Page(index: 1, aid: "626497566", cid: "", epid: "", title: "", duration: 0, resolution: "", pubTime: 0))
                == "https://www.bilibili.com/video/BV1qt4y1X7TW/")
    }

    @Test func otherSitesUseTheirOwnHeaders() async throws {
        let page = externalPage()
        for source in [YtDlpSource(site: .x, ytdlp: nil, http: Stub.client), YtDlpSource(site: .youtube, ytdlp: nil, http: Stub.client)] as [any MediaSource] {
            var config = DownloadConfig()
            config.cookie = "SESSDATA=x"
            config.forceHTTP = true
            config.useAria2c = true
            source.configure(&config, for: request(page))
            #expect(config.cookie.isEmpty && !config.forceHTTP && config.headers == ["User-Agent": "UA"])
            #expect(!source.hasDanmaku && source.trustsStatedSizes && !source.isAudioOnly(page))
            #expect(source.pageURL(page) == "https://v.test/page")
            #expect(await source.subtitles(for: request(page)).map(\.language) == ["en"])
            #expect(try await source.tracks(for: request(page), quality: nil).video.map(\.baseURL) == ["https://v.test/v"])
        }
        // googlevideo: closed ranges only, so no aria2c.
        var yt = DownloadConfig()
        yt.useAria2c = true
        YtDlpSource(site: .youtube, ytdlp: nil, http: Stub.client).configure(&yt, for: request(page))
        #expect(yt.maxRange == YtDlpSource.googlevideoRange && !yt.useAria2c)
        var x = DownloadConfig()
        x.useAria2c = true
        YtDlpSource(site: .x, ytdlp: nil, http: Stub.client).configure(&x, for: request(page))
        #expect(x.maxRange == nil && x.useAria2c)
        // Without yt-dlp there is no info, and the error says what to install.
        await #expect(throws: HaulError.self) {
            _ = try await YtDlpSource(site: .youtube, ytdlp: nil, http: Stub.client).fetchInfo(.link(site: .youtube, url: "u"))
        }
    }

    @Test func podcastEpisodesAreAudio() {
        let source = PodcastSource(site: .applePodcasts, http: Stub.client)
        #expect(source.isAudioOnly(externalPage(video: false)))
        // A video podcast is downloaded like any video.
        #expect(!source.isAudioOnly(externalPage(video: true)))
        #expect(!source.hasDanmaku)
        // Archive keys carry the site, so numeric ids of different sites never collide.
        #expect(source.archiveKey(externalPage()) == "applePodcasts:p1")
        #expect(YtDlpSource(site: .youtube, ytdlp: nil, http: Stub.client).archiveKey(externalPage()) == "youtube:p1")
    }
}
