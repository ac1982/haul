import Foundation
import Testing
@testable import HaulCore

/// Against the real sites. Off by default (slow, and the answers change); run with
/// `HAUL_LIVE=1 swift test --filter LiveTests`. No stored credentials are used.
@Suite(.enabled(if: ProcessInfo.processInfo.environment["HAUL_LIVE"] == "1"), .serialized)
struct LiveTests {
    @Test func bilibili() async throws {
        var session = Session()
        session.wbiKey = try await Auth.fetchWbiKey(session: session)
        let id = try await IDResolver.resolve("https://www.bilibili.com/video/BV1qt4y1X7TW", session: session)
        #expect(id == .video(aid: "626497566"))
        let info = try await InfoFetcher.fetch(id, useIntl: false, session: session)
        let page = try #require(info.pages.first)
        #expect(page.cid == "220355130")
        let tracks = try await PlayURLClient.extractTracks(
            PlayURLRequest(id: id, aid: page.aid, cid: page.cid, epId: "", api: .web, encoding: ""), session: session)
        #expect(tracks.isDash && !tracks.video.isEmpty && !tracks.audio.isEmpty)
    }

    @Test func youtube() async throws {
        let ytdlp = try #require(Shell.findExecutable("yt-dlp"), "brew install yt-dlp deno")
        let (site, url) = try #require(Site.match("https://youtu.be/DdCEmlAydcw"))
        let info = try await YtDlp.fetch(url, site: site, ytdlp: ytdlp)
        let media = try #require(info.pages.first?.media)
        #expect(info.title == "Inside Anthropic's molecular biology lab")
        #expect(!media.tracks.video.isEmpty && !media.tracks.audio.isEmpty && !media.tracks.videoHasAudio)
        #expect(media.subtitles.contains { $0.language == "en" })

        // googlevideo still serves the whole file through closed ranges.
        let audio = try #require(media.tracks.audio.min { $0.size < $1.size })
        var config = DownloadConfig()
        config.headers = media.headers
        config.maxRange = YtDlpSource.googlevideoRange
        config.size = Int64(audio.size)
        let out = (try makeTempDir("live-yt") as NSString).appendingPathComponent("a.m4a")
        try await Downloader.downloadTrack(audio.baseURL, to: out, config: config, isVideo: false)
        #expect(Downloader.fileSize(out) == Int64(audio.size))
    }

    @Test func x() async throws {
        let ytdlp = try #require(Shell.findExecutable("yt-dlp"), "brew install yt-dlp")
        let (site, url) = try #require(Site.match("https://x.com/historyinmemes/status/1790637656616943991"))
        let info = try await YtDlp.fetch(url, site: site, ytdlp: ytdlp)
        let media = try #require(info.pages.first?.media)
        #expect(info.pages.count == 1 && info.pages[0].ownerName == "Historic Vids")
        #expect(media.tracks.videoHasAudio && !media.tracks.video.isEmpty)
    }

    @Test func applePodcasts() async throws {
        let (site, url) = try #require(Site.match("https://podcasts.apple.com/us/podcast/the-daily/id1200361736"))
        let info = try await Podcast.fetch(url, site: site)
        #expect(info.title == "The Daily" && info.pages.count > 10)
        let newest = try #require(info.pages.last?.media)
        #expect(newest.album == "The Daily" && newest.tracks.audio.first?.baseURL.hasPrefix("http") == true)
    }
}
