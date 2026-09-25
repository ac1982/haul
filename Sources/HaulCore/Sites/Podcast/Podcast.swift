import Foundation

/// Podcast episodes from Xiaoyuzhou (小宇宙) and Apple Podcasts. Both lead to a plain audio file, so there is no yt-dlp here:
/// Xiaoyuzhou pages carry their data as Next.js JSON, and Apple has the public iTunes lookup API plus each show's RSS feed.
public enum Podcast {
    /// Some podcast hosts turn away unknown clients; a browser is always welcome.
    static let userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Safari/605.1.15"

    public static func fetch(_ url: String, site: Site, http: HTTPClient = .shared) async throws -> VideoInfo {
        // A fresh session: the bilibili cookie stays with bilibili.
        let session = Session()
        switch site {
        case .xiaoyuzhou:
            guard let ref = Site.xiaoyuzhouID(from: url) else { throw HaulError.unsupported(url) }
            let html = try await http.getString(url, session: session, userAgent: userAgent)
            return try xiaoyuzhou(html: html, id: ref.id, isEpisode: ref.isEpisode)
        case .applePodcasts:
            return try await apple(url, http: http, session: session)
        case .youtube, .x, .bilibili:
            throw HaulError("\(site.name) links are not podcasts")
        }
    }

    /// One episode, whichever site it came from.
    struct Episode {
        var id: String
        var title: String
        var mediaURL: String
        /// MIME type or file extension, when the site states one.
        var type = ""
        /// As stated by the site; shown, but not trusted to drive ranged downloads.
        var size: Int64 = 0
        var duration = 0
        var pubTime: Int64 = 0
        var cover: String?
        var desc: String?
        var show = ""
        var author: String?
        var showID: String?
        var webpageURL = ""
        var isVideo = false
    }

    // MARK: Xiaoyuzhou

    static let xiaoyuzhouHeaders = ["User-Agent": userAgent, "Referer": "https://www.xiaoyuzhoufm.com/"]

    /// An episode page gives that episode; a show page gives the episodes it lists (the latest ones), oldest first.
    static func xiaoyuzhou(html: String, id: String, isEpisode: Bool) throws -> VideoInfo {
        var episodes: [Episode] = []
        var show = JSON.null
        if let m = html.firstMatch(of: /<script[^>]*id="__NEXT_DATA__"[^>]*>([\s\S]*?)<\/script>/),
           let data = try? JSON.parse(String(m.1)) {
            // Found by shape rather than by path, so a reshuffled page still parses.
            var seen = Set<String>()
            walk(data) { node in
                if let e = xiaoyuzhouEpisode(node), seen.insert(e.id).inserted { episodes.append(e) }
                if !show.isObject, !node.has("eid"), let pid = node["pid"].string, node["title"].string != nil,
                   pid == (isEpisode ? episodes.first(where: { $0.id == id })?.showID : id) {
                    show = node
                }
            }
        }
        // Episodes nested in their show's page do not repeat the show.
        for i in episodes.indices {
            if episodes[i].show.isEmpty { episodes[i].show = show["title"].stringValue }
            if episodes[i].author == nil { episodes[i].author = show["author"].string }
            if episodes[i].cover == nil { episodes[i].cover = xiaoyuzhouImage(show["image"]) }
        }

        if isEpisode {
            if let e = episodes.first(where: { $0.id == id }) {
                return info([e], site: .xiaoyuzhou, headers: xiaoyuzhouHeaders)
            }
            // The page data moved: the Open Graph tags still name the audio.
            if let audio = meta("og:audio", in: html) {
                let e = Episode(id: id, title: meta("og:title", in: html) ?? id, mediaURL: audio,
                                cover: meta("og:image", in: html), desc: meta("og:description", in: html),
                                webpageURL: "https://www.xiaoyuzhoufm.com/episode/\(id)")
                return info([e], site: .xiaoyuzhou, headers: xiaoyuzhouHeaders)
            }
            throw HaulError("No audio on the Xiaoyuzhou page (a paid episode, or the page changed; --debug shows more)")
        }
        let listed = episodes.filter { $0.showID == nil || $0.showID == id }
        guard !listed.isEmpty else { throw HaulError("No episodes on the Xiaoyuzhou show page (the page may have changed; --debug shows more)") }
        Log.status("The show page lists its latest \(listed.count) episodes; download older ones by their episode links")
        return info(listed, site: .xiaoyuzhou, headers: xiaoyuzhouHeaders, show: show["title"].string ?? listed[0].show,
                    desc: show["description"].stringValue)
    }

    static func xiaoyuzhouEpisode(_ e: JSON) -> Episode? {
        guard let eid = e["eid"].string, let title = e["title"].string,
              let audio = e["enclosure"]["url"].string ?? e["media"]["source"]["url"].string, !audio.isEmpty else { return nil }
        let podcast = e["podcast"]
        return Episode(id: eid, title: title.trimmingCharacters(in: .whitespacesAndNewlines), mediaURL: audio,
                       type: e["media"]["mimeType"].stringValue,
                       size: e["media"]["size"].int64 ?? 0,
                       duration: e["duration"].int ?? 0,
                       pubTime: parseISODate(e["pubDate"].stringValue),
                       cover: xiaoyuzhouImage(e["image"]) ?? xiaoyuzhouImage(podcast["image"]),
                       desc: e["description"].string,
                       show: podcast["title"].stringValue,
                       author: podcast["author"].string,
                       showID: e["pid"].string ?? podcast["pid"].string,
                       webpageURL: "https://www.xiaoyuzhoufm.com/episode/\(eid)")
    }

    static func xiaoyuzhouImage(_ image: JSON) -> String? {
        [image["picUrl"], image["largePicUrl"], image["middlePicUrl"]].compactMap(\.string).first { !$0.isEmpty }
    }

    // MARK: Apple Podcasts

    static let appleHeaders = ["User-Agent": userAgent]

    static func apple(_ url: String, http: HTTPClient, session: Session) async throws -> VideoInfo {
        guard let ref = Site.applePodcastID(from: url) else { throw HaulError.unsupported(url) }
        let country = ref.country ?? "us"
        let lookup = try await http.getJSON(
            "https://itunes.apple.com/lookup?id=\(ref.show)&country=\(country)&media=podcast&entity=podcastEpisode&limit=200", session: session)
        if let info = try apple(lookup: lookup, show: ref.show, episode: ref.episode) { return info }

        // The lookup lists only the latest 200 episodes. An older one is found in the show's feed by its title.
        guard let episode = ref.episode else { throw HaulError("Apple Podcasts returned no episodes for this show") }
        let show = appleShow(lookup)
        guard let feed = show["feedUrl"].string else { throw HaulError("Apple Podcasts gave no RSS feed for this show") }
        Log.status("The episode is not among the latest 200; looking in the RSS feed")
        async let pageText = http.getString(url, session: session, userAgent: userAgent)
        async let feedText = http.getString(feed, session: session, userAgent: userAgent)
        let (page, rss) = try await (pageText, feedText)
        let showName = show["collectionName"].stringValue
        let titles = appleEpisodeTitles(page, show: showName)
        guard !titles.isEmpty else { throw HaulError("No episode title on the Apple Podcasts page") }
        guard var e = rssEpisode(rss, matching: titles) else {
            throw HaulError("Episode \"\(titles[0])\" is not in the show's RSS feed (the feed may keep only newer episodes)")
        }
        e.id = episode
        e.show = showName
        e.author = e.author ?? show["artistName"].string
        e.showID = ref.show
        e.cover = e.cover ?? artwork(show)
        e.webpageURL = url
        return info([e], site: .applePodcasts, headers: appleHeaders)
    }

    /// nil when the link names an episode the lookup did not return.
    static func apple(lookup: JSON, show id: String, episode: String?) throws -> VideoInfo? {
        let results = lookup["results"].array
        let show = appleShow(lookup)
        guard show.exists || !results.isEmpty else { throw HaulError("Apple Podcasts has no show id\(id) (the country in the link may be wrong)") }
        let episodes = results.filter { $0["wrapperType"].string == "podcastEpisode" }.compactMap { appleEpisode($0, show: show) }
        if let episode {
            guard let e = episodes.first(where: { $0.id == episode }) else { return nil }
            return info([e], site: .applePodcasts, headers: appleHeaders)
        }
        guard !episodes.isEmpty else { throw HaulError("Apple Podcasts returned no episodes for this show") }
        if appleListIsCut(results) { Log.warn("Apple returns only the latest \(episodes.count) episodes; download older ones by their episode links") }
        return info(episodes, site: .applePodcasts, headers: appleHeaders, show: show["collectionName"].string ?? episodes[0].show, desc: "")
    }

    /// The lookup stops at 200 episodes; the show's own entry does not count.
    static func appleListIsCut(_ results: [JSON]) -> Bool {
        results.count { $0["wrapperType"].string == "podcastEpisode" } >= 200
    }

    static func appleShow(_ lookup: JSON) -> JSON {
        lookup["results"].array.first { $0["wrapperType"].string == "track" && $0["kind"].string == "podcast" } ?? .null
    }

    static func appleEpisode(_ e: JSON, show: JSON) -> Episode? {
        guard let id = e["trackId"].int64, let url = e["episodeUrl"].string ?? e["previewUrl"].string, !url.isEmpty else { return nil }
        return Episode(id: String(id), title: e["trackName"].stringValue.trimmingCharacters(in: .whitespacesAndNewlines), mediaURL: url,
                       type: e["episodeFileExtension"].stringValue,
                       duration: Int((e["trackTimeMillis"].int64 ?? 0) / 1000),
                       pubTime: parseISODate(e["releaseDate"].stringValue),
                       cover: artwork(e) ?? artwork(show),
                       desc: e["description"].string ?? e["shortDescription"].string,
                       show: e["collectionName"].string ?? show["collectionName"].stringValue,
                       author: e["artistName"].string ?? show["artistName"].string,
                       showID: e["collectionId"].int64.map(String.init),
                       webpageURL: e["trackViewUrl"].string ?? "",
                       isVideo: e["episodeContentType"].string == "video")
    }

    /// The artwork at 1400 px; the lookup offers 600 px, and the image server renders any size the path asks for.
    static func artwork(_ node: JSON) -> String? {
        guard let url = node["artworkUrl600"].string ?? node["artworkUrl160"].string, !url.isEmpty else { return nil }
        return url.replacing(/\/\d+x\d+bb\.(jpg|png|webp)$/, with: "/1400x1400bb.jpg")
    }

    // MARK: RSS

    /// The episode's title as its Apple page gives it: `og:title`, and `<title>` without " - <show> - Apple Podcasts".
    static func appleEpisodeTitles(_ html: String, show: String) -> [String] {
        var titles: [String] = []
        if let og = meta("og:title", in: html) { titles.append(og) }
        if let m = html.firstMatch(of: /<title[^>]*>([^<]*)<\/title>/) {
            var title = normalize(Format.unescapeEntities(String(m.1)))
            let suffixes = [" - apple podcasts", " on apple podcasts"] + (show.isEmpty ? [] : [" - " + normalize(show)])
            for suffix in suffixes where title.hasSuffix(suffix) { title = String(title.dropLast(suffix.count)) }
            titles.append(title)
        }
        return titles.filter { !normalize($0).isEmpty }
    }

    /// The feed item titled exactly one of `titles` (case and spacing aside). A near miss is no match:
    /// it would download another episode (a trailer, a bonus) under this one's name.
    static func rssEpisode(_ rss: String, matching titles: [String]) -> Episode? {
        let wanted = Set(titles.map(normalize).filter { !$0.isEmpty })
        for item in rss.matches(of: /<item[\s>][\s\S]*?<\/item>/) {
            let text = String(item.output)
            // The title alone decides; only a match is parsed further.
            guard let title = rssTag("title", in: text), wanted.contains(normalize(title)), let e = rssItem(text) else { continue }
            return e
        }
        return nil
    }

    static func rssTag(_ name: String, in item: String) -> String? {
        guard let r = try? Regex("<\(name)(?:\\s[^>]*)?>([\\s\\S]*?)</\(name)>"),
              let m = item.firstMatch(of: r), let text = m.output[1].substring else { return nil }
        return Format.unescapeEntities(String(text).replacing(/^\s*<!\[CDATA\[([\s\S]*?)\]\]>\s*$/) { String($0.1) })
            .trimmingCharacters(in: .whitespacesAndNewlines)
    }

    static func rssItem(_ item: String) -> Episode? {
        func tag(_ name: String) -> String? { rssTag(name, in: item) }
        guard let enclosure = item.firstMatch(of: /<enclosure\s[^>]*>/).map({ String($0.output) }),
              let url = attribute("url", in: enclosure) else { return nil }
        let type = attribute("type", in: enclosure) ?? ""
        return Episode(id: tag("guid") ?? url, title: tag("title") ?? "", mediaURL: url, type: type,
                       duration: parseDuration(tag("itunes:duration") ?? ""),
                       pubTime: parseRFC822Date(tag("pubDate") ?? ""),
                       cover: item.firstMatch(of: /<itunes:image\s[^>]*>/).flatMap { attribute("href", in: String($0.output)) },
                       // Show notes are HTML; the tag keeps their text.
                       desc: (tag("itunes:summary") ?? tag("description")).map {
                           $0.replacing(/<[^>]+>/, with: "").trimmingCharacters(in: .whitespacesAndNewlines)
                       },
                       author: tag("itunes:author"),
                       isVideo: type.hasPrefix("video/"))
    }

    // MARK: shared

    static func info(_ episodes: [Episode], site: Site, headers: [String: String], show: String? = nil, desc: String = "") -> VideoInfo {
        var info: VideoInfo
        if let show {
            // A show: oldest first, so -p LATEST is the newest episode.
            let ordered = episodes.enumerated().sorted { a, b in
                a.element.pubTime != b.element.pubTime ? a.element.pubTime < b.element.pubTime : a.offset > b.offset
            }.map(\.element)
            let pages = ordered.enumerated().map { page($1, index: $0 + 1, headers: headers) }
            // Each page brings its own cover.
            info = VideoInfo(title: show, desc: desc, cover: "", pubTime: ordered.last?.pubTime ?? 0, pages: pages)
        } else {
            let e = episodes[0]
            info = VideoInfo(title: e.title, desc: e.desc ?? "", cover: e.cover ?? "", pubTime: e.pubTime, pages: [page(e, index: 1, headers: headers)])
        }
        info.site = site
        return info
    }

    static func page(_ e: Episode, index: Int, headers: [String: String]) -> Page {
        var tracks = ParsedTracks()
        let kbps = e.size > 0 && e.duration > 0 ? Int64((Double(e.size) * 8 / 1024 / Double(e.duration)).rounded()) : 0
        if e.isVideo {
            // A video podcast is one file with the audio inside, like an X post.
            tracks.video = [VideoTrack(id: "0", quality: "original", baseURL: e.mediaURL, codec: "MP4", bandwidth: kbps, duration: e.duration)]
            tracks.videoHasAudio = true
        } else {
            tracks.audio = [AudioTrack(id: "0", quality: "", baseURL: e.mediaURL, codec: audioCodec(e.type, url: e.mediaURL),
                                       bandwidth: kbps, duration: e.duration)]
        }
        var page = Page(index: index, aid: e.id, cid: e.id, epid: "", title: e.title, duration: e.duration, resolution: "",
                        pubTime: e.pubTime, cover: e.cover, desc: e.desc, ownerName: e.author, ownerMid: e.showID)
        page.media = PageMedia(tracks: tracks, subtitles: [], headers: headers, webpageURL: e.webpageURL, album: e.show)
        return page
    }

    /// `MP3` or `M4A` from a MIME type or the URL's extension. With no hint at all it is MP3, as most podcasts are;
    /// a format we do not know keeps its own name, and so goes into an M4A (which can carry nearly anything).
    static func audioCodec(_ type: String, url: String) -> String {
        let ext = (URL(string: url)?.pathExtension ?? "").lowercased()
        let hints = [type.lowercased(), ext].filter { !$0.isEmpty }
        for hint in hints {
            if hint.contains("mpeg") || hint.contains("mp3") { return "MP3" }
            if ["mp4", "m4a", "m4b", "aac"].contains(where: { hint.contains($0) }) { return "M4A" }
        }
        guard let hint = hints.last else { return "MP3" }
        return (hint.split(separator: "/").last.map(String.init) ?? hint).uppercased()
    }

    /// Every object in the tree, depth first, keys in order so the result does not depend on hashing.
    static func walk(_ node: JSON, _ visit: (JSON) -> Void) {
        switch node {
        case .object(let o):
            visit(node)
            for key in o.keys.sorted() { walk(o[key]!, visit) }
        case .array(let a):
            for item in a { walk(item, visit) }
        default:
            break
        }
    }

    /// `<meta property="og:audio" content="…">`, attributes in either order.
    static func meta(_ property: String, in html: String) -> String? {
        for m in html.matches(of: /<meta\s[^>]*>/) {
            let tag = String(m.output)
            guard attribute("property", in: tag) == property || attribute("name", in: tag) == property,
                  let content = attribute("content", in: tag), !content.isEmpty else { continue }
            return content
        }
        return nil
    }

    static func attribute(_ name: String, in tag: String) -> String? {
        for m in tag.matches(of: /([\w:-]+)\s*=\s*(?:"([^"]*)"|'([^']*)')/) where m.1 == name {
            return Format.unescapeEntities(String(m.2 ?? m.3 ?? ""))
        }
        return nil
    }

    /// For comparing titles: no invisible marks (Apple prefixes U+200E), single spaces, no case.
    static func normalize(_ text: String) -> String {
        text.replacingOccurrences(of: "\u{200E}", with: "").replacingOccurrences(of: "\u{200F}", with: "")
            .split(whereSeparator: \.isWhitespace).joined(separator: " ").lowercased()
    }

    /// `2024-09-20T09:00:00Z`, with or without fractional seconds.
    static func parseISODate(_ text: String) -> Int64 {
        guard !text.isEmpty else { return 0 }
        let f = ISO8601DateFormatter()
        f.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        if let d = f.date(from: text) { return Int64(d.timeIntervalSince1970) }
        f.formatOptions = [.withInternetDateTime]
        return f.date(from: text).map { Int64($0.timeIntervalSince1970) } ?? 0
    }

    /// RSS dates: `Fri, 20 Sep 2024 09:00:00 +0000` or `… GMT`.
    static func parseRFC822Date(_ text: String) -> Int64 {
        let f = DateFormatter()
        f.locale = Locale(identifier: "en_US_POSIX")
        for pattern in ["EEE, d MMM yyyy HH:mm:ss Z", "EEE, d MMM yyyy HH:mm:ss zzz", "d MMM yyyy HH:mm:ss Z", "EEE, d MMM yyyy HH:mm Z"] {
            f.dateFormat = pattern
            if let d = f.date(from: text.trimmingCharacters(in: .whitespaces)) { return Int64(d.timeIntervalSince1970) }
        }
        return 0
    }

    /// `3600`, `59:30` or `1:02:03` → seconds.
    static func parseDuration(_ text: String) -> Int {
        text.split(separator: ":").reduce(0) { $0 * 60 + Int(Double($1.trimmingCharacters(in: .whitespaces)) ?? 0) }
    }
}
