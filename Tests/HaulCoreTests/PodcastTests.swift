import Foundation
import Testing
@testable import HaulCore

@Suite struct PodcastTests {
    static let eid = "6650a1b2c3d4e5f6a7b8c9d0"
    static let pid = "6021f949a789fca4eff4492c"

    // MARK: links

    @Test func recognisesXiaoyuzhouLinks() throws {
        let episode = try #require(Site.xiaoyuzhouID(from: "https://www.xiaoyuzhoufm.com/episode/\(Self.eid)?s=eyJ1"))
        #expect(episode.id == Self.eid && episode.isEpisode)
        let show = try #require(Site.xiaoyuzhouID(from: "xiaoyuzhoufm.com/podcast/\(Self.pid)"))
        #expect(show.id == Self.pid && !show.isEpisode)
        #expect(Site.xiaoyuzhouID(from: "https://www.xiaoyuzhoufm.com/") == nil)
        #expect(Site.xiaoyuzhouID(from: "https://www.xiaoyuzhoufm.com/about") == nil)
        #expect(Site.xiaoyuzhouID(from: "https://notxiaoyuzhoufm.com/episode/\(Self.eid)") == nil)
    }

    @Test func recognisesApplePodcastLinks() throws {
        let ep = try #require(Site.applePodcastID(from: "https://podcasts.apple.com/cn/podcast/%E5%A3%B0%E4%B8%9C/id1487143507?i=1000655000001&l=en"))
        #expect(ep.show == "1487143507" && ep.episode == "1000655000001" && ep.country == "cn")
        let show = try #require(Site.applePodcastID(from: "podcasts.apple.com/us/podcast/the-daily/id1200361736"))
        #expect(show.show == "1200361736" && show.episode == nil && show.country == "us")
        let bare = try #require(Site.applePodcastID(from: "https://podcasts.apple.com/podcast/id1200361736"))
        #expect(bare.show == "1200361736" && bare.country == nil)
        #expect(Site.applePodcastID(from: "https://podcasts.apple.com/us/browse") == nil)
        #expect(Site.applePodcastID(from: "https://music.apple.com/us/album/x/id123") == nil)
    }

    @Test func resolverRoutesPodcasts() async throws {
        #expect(try await IDResolver.resolve("https://www.xiaoyuzhoufm.com/episode/\(Self.eid)?s=abc", session: Session())
                == .link(site: .xiaoyuzhou, url: "https://www.xiaoyuzhoufm.com/episode/\(Self.eid)"))
        #expect(try await IDResolver.resolve("https://podcasts.apple.com/cn/podcast/slug/id1487143507?i=1000655000001", session: Session())
                == .link(site: .applePodcasts, url: "https://podcasts.apple.com/cn/podcast/id1487143507?i=1000655000001"))
        #expect(try await IDResolver.resolve("https://podcasts.apple.com/podcast/the-daily/id1200361736", session: Session())
                == .link(site: .applePodcasts, url: "https://podcasts.apple.com/us/podcast/id1200361736"))
        #expect(Site.xiaoyuzhou.isPodcast && !Site.xiaoyuzhou.usesYtDlp && Site.youtube.usesYtDlp)
    }

    // MARK: Xiaoyuzhou

    static func episodeHTML(eid: String = PodcastTests.eid, audio: String = "https://media.xyz.test/\(PodcastTests.pid)/ep.m4a", data: Bool = true) -> String {
        let next = #"""
        {"props":{"pageProps":{"episode":{"type":"EPISODE","eid":"\#(eid)","pid":"\#(pid)","title":" E42 声音的故事 ",
        "description":"节目简介","duration":2712,"enclosure":{"url":"\#(audio)"},
        "media":{"id":"m1","size":43392000,"mimeType":"audio/mp4","source":{"mode":"PUBLIC","url":"\#(audio)"}},
        "pubDate":"2024-05-24T22:00:00.000Z","image":{"picUrl":"https://image.xyz.test/ep.jpg"},
        "podcast":{"type":"PODCAST","pid":"\#(pid)","title":"声东击西","author":"声动活泼","image":{"picUrl":"https://image.xyz.test/show.jpg"}}}},
        "__N_SSP":true},"page":"/episode/[id]","query":{"id":"\#(eid)"}}
        """#
        return """
        <!DOCTYPE html><html><head><meta charset="utf-8"/><meta property="og:title" content="E42 声音的故事 &amp; 更多"/>
        <meta content="https://media.xyz.test/og.mp3" property="og:audio"/></head>
        <body><div id="__next"></div>\(data ? "<script id=\"__NEXT_DATA__\" type=\"application/json\">\(next)</script>" : "")</body></html>
        """
    }

    static let showHTML = #"""
    <html><body><script id="__NEXT_DATA__" type="application/json">{"props":{"pageProps":{"podcast":{"type":"PODCAST","pid":"\#(pid)",
    "title":"声东击西","author":"声动活泼","description":"关于节目","image":{"picUrl":"https://image.xyz.test/show.jpg"},
    "episodes":[
      {"eid":"e3","pid":"\#(pid)","title":"第三期","duration":60,"enclosure":{"url":"https://media.xyz.test/3.mp3"},"media":{"mimeType":"audio/mpeg"},"pubDate":"2024-03-01T00:00:00.000Z"},
      {"eid":"e2","pid":"\#(pid)","title":"第二期","duration":60,"enclosure":{"url":"https://media.xyz.test/2.m4a"},"pubDate":"2024-02-01T00:00:00.000Z","image":{"picUrl":"https://image.xyz.test/e2.jpg"}},
      {"eid":"e1","pid":"\#(pid)","title":"第一期","duration":60,"enclosure":{"url":"https://media.xyz.test/1.m4a"},"pubDate":"2024-01-01T00:00:00.000Z"}
    ]}}}}</script></body></html>
    """#

    @Test func parsesXiaoyuzhouEpisode() throws {
        let info = try Podcast.xiaoyuzhou(html: Self.episodeHTML(), id: Self.eid, isEpisode: true)
        #expect(info.site == .xiaoyuzhou)
        #expect(info.title == "E42 声音的故事" && info.desc == "节目简介" && info.cover == "https://image.xyz.test/ep.jpg")
        #expect(info.pubTime == 1716588000)
        let page = try #require(info.pages.first)
        #expect(info.pages.count == 1)
        #expect(page.aid == Self.eid && page.duration == 2712 && page.ownerName == "声动活泼" && page.ownerMid == Self.pid)
        let media = try #require(page.media)
        #expect(media.album == "声东击西")
        #expect(media.webpageURL == "https://www.xiaoyuzhoufm.com/episode/\(Self.eid)")
        #expect(media.headers["Referer"] == "https://www.xiaoyuzhoufm.com/")
        #expect(media.tracks.video.isEmpty && !media.tracks.videoHasAudio)
        let audio = try #require(media.tracks.audio.first)
        #expect(audio.baseURL == "https://media.xyz.test/\(Self.pid)/ep.m4a" && audio.codec == "M4A")
        // The stated size gives a bitrate for the table, but never drives the downloader.
        #expect(audio.bandwidth == 125 && audio.size == 0)
    }

    @Test func xiaoyuzhouFallsBackToOpenGraph() throws {
        let info = try Podcast.xiaoyuzhou(html: Self.episodeHTML(data: false), id: Self.eid, isEpisode: true)
        #expect(info.title == "E42 声音的故事 & 更多")
        let audio = try #require(info.pages.first?.media?.tracks.audio.first)
        #expect(audio.baseURL == "https://media.xyz.test/og.mp3" && audio.codec == "MP3")
        #expect(throws: HaulError.self) { try Podcast.xiaoyuzhou(html: "<html></html>", id: Self.eid, isEpisode: true) }
    }

    @Test func parsesXiaoyuzhouShowOldestFirst() throws {
        let info = try Podcast.xiaoyuzhou(html: Self.showHTML, id: Self.pid, isEpisode: false)
        #expect(info.title == "声东击西" && info.desc == "关于节目" && info.cover.isEmpty)
        #expect(info.pages.map(\.title) == ["第一期", "第二期", "第三期"])
        #expect(info.pages.map(\.index) == [1, 2, 3])
        #expect(info.pages.map(\.cover) == ["https://image.xyz.test/show.jpg", "https://image.xyz.test/e2.jpg", "https://image.xyz.test/show.jpg"])
        #expect(info.pages.allSatisfy { $0.ownerName == "声动活泼" && $0.media?.album == "声东击西" })
        #expect(info.pages.compactMap { $0.media?.tracks.audio.first?.codec } == ["M4A", "M4A", "MP3"])
        #expect(UI.header(info, loggedIn: nil, style: .plain)[1].contains("3 episodes"))
    }

    // MARK: Apple Podcasts

    static let lookup = #"""
    {"resultCount":4,"results":[
     {"wrapperType":"track","kind":"podcast","collectionId":1200361736,"trackId":1200361736,"artistName":"The New York Times",
      "collectionName":"The Daily","feedUrl":"https://feeds.test/daily","artworkUrl600":"https://is1-ssl.mzstatic.test/image/thumb/show/600x600bb.jpg"},
     {"wrapperType":"podcastEpisode","kind":"podcast-episode","trackId":1000671234567,"trackName":"Newer","collectionId":1200361736,
      "collectionName":"The Daily","releaseDate":"2024-09-20T09:45:00Z","trackTimeMillis":1800000,
      "episodeUrl":"https://podcast.test/redirect.mp3/new.mp3?dest-id=1","episodeFileExtension":"mp3","episodeContentType":"audio",
      "description":"About the newer one","artworkUrl600":"https://is1-ssl.mzstatic.test/image/thumb/ep/600x600bb.jpg",
      "trackViewUrl":"https://podcasts.apple.com/us/podcast/newer/id1200361736?i=1000671234567&uo=4"},
     {"wrapperType":"podcastEpisode","kind":"podcast-episode","trackId":1000671000001,"trackName":"Older","collectionId":1200361736,
      "collectionName":"The Daily","releaseDate":"2024-09-19T09:45:00Z","trackTimeMillis":1500000,
      "episodeUrl":"https://podcast.test/old.m4a","episodeFileExtension":"m4a","episodeContentType":"audio"},
     {"wrapperType":"podcastEpisode","kind":"podcast-episode","trackId":1000671000002,"trackName":"On camera","collectionId":1200361736,
      "collectionName":"The Daily","releaseDate":"2024-09-18T09:45:00Z","trackTimeMillis":600000,
      "episodeUrl":"https://podcast.test/video.mp4","episodeFileExtension":"mp4","episodeContentType":"video"}
    ]}
    """#

    @Test func parsesAppleEpisode() throws {
        let info = try #require(try Podcast.apple(lookup: JSON.parse(Self.lookup), show: "1200361736", episode: "1000671234567"))
        #expect(info.site == .applePodcasts && info.title == "Newer" && info.desc == "About the newer one")
        #expect(info.cover == "https://is1-ssl.mzstatic.test/image/thumb/ep/1400x1400bb.jpg")
        #expect(info.pubTime == 1726825500)
        let page = try #require(info.pages.first)
        #expect(page.aid == "1000671234567" && page.duration == 1800 && page.ownerName == "The New York Times")
        let media = try #require(page.media)
        #expect(media.album == "The Daily")
        #expect(media.headers["User-Agent"] == Podcast.userAgent)
        #expect(media.webpageURL == "https://podcasts.apple.com/us/podcast/newer/id1200361736?i=1000671234567&uo=4")
        let audio = try #require(media.tracks.audio.first)
        #expect(audio.codec == "MP3" && audio.baseURL == "https://podcast.test/redirect.mp3/new.mp3?dest-id=1")
    }

    @Test func parsesAppleShow() throws {
        let lookup = try JSON.parse(Self.lookup)
        let info = try #require(try Podcast.apple(lookup: lookup, show: "1200361736", episode: nil))
        #expect(info.title == "The Daily")
        #expect(info.pages.map(\.title) == ["On camera", "Older", "Newer"])
        // A video episode is one file with its audio inside.
        let video = try #require(info.pages.first?.media?.tracks)
        #expect(video.videoHasAudio && video.audio.isEmpty && video.video.first?.baseURL == "https://podcast.test/video.mp4")
        #expect(info.pages[1].media?.tracks.audio.first?.codec == "M4A")
        // Older episodes than the lookup returns are looked up elsewhere; an unknown show is an error.
        #expect(try Podcast.apple(lookup: lookup, show: "1200361736", episode: "42") == nil)
        #expect(throws: HaulError.self) { try Podcast.apple(lookup: JSON.parse(#"{"resultCount":0,"results":[]}"#), show: "1", episode: nil) }
        // 199 episodes plus the show is the whole list; 200 episodes may not be.
        let show = lookup["results"].array[0]
        let episode = lookup["results"].array[1]
        #expect(!Podcast.appleListIsCut([show] + Array(repeating: episode, count: 199)))
        #expect(Podcast.appleListIsCut([show] + Array(repeating: episode, count: 200)))
    }

    static let rss = """
    <?xml version="1.0" encoding="UTF-8"?><rss version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd"><channel>
    <title>The Daily</title>
    <item><title>The Daily</title><enclosure url="https://podcast.test/trailer.mp3" length="1" type="audio/mpeg"/></item>
    <item><title>Intro</title><enclosure url="https://podcast.test/intro.mp3" length="1" type="audio/mpeg"/></item>
    <item>
      <title><![CDATA[Bits &amp; Pieces: The Long One]]></title><itunes:title>Long One</itunes:title>
      <guid isPermaLink="false">abc-123</guid><pubDate>Fri, 20 Sep 2024 09:45:00 +0000</pubDate>
      <itunes:duration>1:02:03</itunes:duration><itunes:author>Host</itunes:author>
      <itunes:image href="https://podcast.test/long.jpg"/>
      <description><![CDATA[<p>What the <b>long</b> one is about.</p>]]></description>
      <enclosure length="123" type="audio/x-m4a" url="https://podcast.test/long.m4a?a=1&amp;b=2"/>
    </item>
    </channel></rss>
    """

    @Test func findsAnOlderEpisodeInTheFeed() throws {
        // Apple titles its pages with an invisible mark and the show's name around the episode title.
        let page = "<html><head><title>\u{200E}Bits &amp; Pieces: The Long One - The Daily - Apple Podcasts</title></head></html>"
        let titles = Podcast.appleEpisodeTitles(page, show: "The Daily")
        let e = try #require(Podcast.rssEpisode(Self.rss, matching: titles))
        #expect(e.title == "Bits & Pieces: The Long One" && e.desc == "What the long one is about.")
        #expect(e.mediaURL == "https://podcast.test/long.m4a?a=1&b=2" && e.type == "audio/x-m4a")
        #expect(e.duration == 3723 && e.pubTime == 1726825500 && e.cover == "https://podcast.test/long.jpg" && e.author == "Host")
    }

    /// An episode the feed no longer has is an error, never the trailer or another item whose title fits inside.
    @Test func aMissingEpisodeIsNotReplacedByANearMatch() {
        let page = "<html><head><title>An Old Episode - The Daily - Apple Podcasts</title></head></html>"
        #expect(Podcast.rssEpisode(Self.rss, matching: Podcast.appleEpisodeTitles(page, show: "The Daily")) == nil)
        #expect(Podcast.rssEpisode(Self.rss, matching: ["Intro to everything"]) == nil)
        #expect(Podcast.rssEpisode(Self.rss, matching: ["  INTRO "])?.mediaURL == "https://podcast.test/intro.mp3")
    }

    // MARK: helpers

    @Test func codecsAndDates() {
        #expect(Podcast.audioCodec("audio/mpeg", url: "https://a.test/x") == "MP3")
        #expect(Podcast.audioCodec("", url: "https://a.test/x.m4a?t=1") == "M4A")
        #expect(Podcast.audioCodec("audio/mp4", url: "https://a.test/x.mp3") == "M4A")
        #expect(Podcast.audioCodec("", url: "https://a.test/redirect") == "MP3")
        // A format we do not know must not be muxed as MP3: it keeps its name and goes into an M4A.
        #expect(Podcast.audioCodec("audio/ogg", url: "https://a.test/x") == "OGG")
        #expect(Podcast.audioCodec("", url: "https://a.test/x.opus") == "OPUS")
        #expect(Podcast.audioCodec("audio/x-m4b", url: "https://a.test/x") == "M4A")
        #expect(Container.audio(codec: "OGG") == .m4a)
        #expect(Container.audio(codec: "MP3") == .mp3 && Container.audio(codec: "M4A") == .m4a && Container.audio(codec: "OPUS") == .m4a)
        #expect(Podcast.parseISODate("2024-05-24T22:00:00.000Z") == 1716588000)
        #expect(Podcast.parseISODate("2024-05-24T22:00:00Z") == 1716588000)
        #expect(Podcast.parseRFC822Date("Fri, 24 May 2024 22:00:00 GMT") == 1716588000)
        #expect(Podcast.parseDuration("3600") == 3600 && Podcast.parseDuration("59:30") == 3570)
        #expect(Format.unescapeEntities("a &amp; b &#26159; &#x4E2D;") == "a & b 是 中")
        #expect(Podcast.artwork(JSON.object(["artworkUrl600": .string("https://x.test/a/600x600bb.png")])) == "https://x.test/a/1400x1400bb.jpg")
    }

    // MARK: over the stub network

    @Test func fetchesOverTheNetwork() async throws {
        Stub.on("www.xiaoyuzhoufm.com/episode/\(Self.eid)") { _ in StubResponse(body: Data(Self.episodeHTML().utf8)) }
        let info = try await Podcast.fetch("https://www.xiaoyuzhoufm.com/episode/\(Self.eid)", site: .xiaoyuzhou, http: Stub.client)
        #expect(info.title == "E42 声音的故事")
        #expect(Stub.requests("xiaoyuzhoufm.com/episode").allSatisfy { $0.value(forHTTPHeaderField: "User-Agent") == Podcast.userAgent })

        Stub.on("itunes.apple.com/lookup?id=1200361736") { _ in .json(Self.lookup) }
        let apple = try await Podcast.fetch("https://podcasts.apple.com/us/podcast/id1200361736?i=1000671000001", site: .applePodcasts, http: Stub.client)
        #expect(apple.title == "Older")
        #expect(Stub.requests("itunes.apple.com/lookup?id=1200361736").first?.url?.query?.contains("country=us") == true)
    }

    @Test func olderAppleEpisodeComesFromTheFeed() async throws {
        Stub.on("itunes.apple.com/lookup?id=1200361736") { _ in .json(Self.lookup) }
        Stub.on("podcasts.apple.com/us/podcast/id1200361736?i=1000000000009") { _ in
            StubResponse(body: Data(#"<html><head><meta property="og:title" content="Bits &amp; Pieces: The Long One"></head></html>"#.utf8))
        }
        Stub.on("feeds.test/daily") { _ in StubResponse(body: Data(Self.rss.utf8)) }
        let info = try await Podcast.fetch("https://podcasts.apple.com/us/podcast/id1200361736?i=1000000000009", site: .applePodcasts, http: Stub.client)
        let page = try #require(info.pages.first)
        #expect(info.title == "Bits & Pieces: The Long One" && page.aid == "1000000000009")
        #expect(page.media?.album == "The Daily" && page.ownerName == "Host")
        #expect(page.media?.tracks.audio.first?.baseURL == "https://podcast.test/long.m4a?a=1&b=2")
    }
}
