import Foundation

/// The sites haul downloads from, recognised by their links. How each is extracted and downloaded is its
/// `MediaSource`: yt-dlp for YouTube and X, `Podcast` for podcasts, bilibili's own APIs for bilibili.
public enum Site: String, Sendable, CaseIterable, Codable {
    case youtube, x, bilibili, xiaoyuzhou, applePodcasts

    public var name: String {
        switch self {
        case .youtube: return "YouTube"
        case .x: return "X"
        case .bilibili: return "bilibili"
        case .xiaoyuzhou: return "Xiaoyuzhou"
        case .applePodcasts: return "Apple Podcasts"
        }
    }

    /// How the header names the uploader.
    var ownerLabel: String {
        switch self {
        case .youtube: return "channel"
        case .x: return "by"
        case .bilibili: return "uploader"
        case .xiaoyuzhou, .applePodcasts: return "host"
        }
    }

    /// What one downloadable item is called.
    var unit: String {
        switch self {
        case .bilibili: return "page"
        case .xiaoyuzhou, .applePodcasts: return "episode"
        case .youtube, .x: return "video"
        }
    }

    public var isPodcast: Bool { self == .xiaoyuzhou || self == .applePodcasts }

    /// Podcast pages and APIs are plain JSON / HTML and bilibili has its own clients; only YouTube and X need yt-dlp.
    public var usesYtDlp: Bool { self == .youtube || self == .x }

    /// The links each site accepts, for help texts and error messages.
    public var linkForms: String {
        switch self {
        case .youtube: return "youtube.com/watch?v=…, youtu.be/…, /shorts/…, /embed/…, /live/…"
        case .x: return "x.com/<user>/status/<id>, twitter.com/…; /video/<n> picks one"
        case .bilibili: return "bilibili.com (video, bangumi, course, space, list), b23.tv, BV…, ep…"
        case .xiaoyuzhou: return "xiaoyuzhoufm.com/episode/<id>, /podcast/<id>"
        case .applePodcasts: return "podcasts.apple.com/<cc>/podcast/<name>/id<show>[?i=<episode>]"
        }
    }

    /// Recognise a supported link: the site and the URL its source reads (bilibili links are resolved by `IDResolver`).
    public static func match(_ input: String) -> (site: Site, url: String)? {
        if isBilibili(input) { return (.bilibili, input.trimmingCharacters(in: .whitespacesAndNewlines)) }
        if let id = youtubeID(from: input) { return (.youtube, "https://www.youtube.com/watch?v=\(id)") }
        if tweetID(from: input) != nil {
            // Kept as given: `/video/2` picks one video of a multi-video post.
            let text = input.trimmingCharacters(in: .whitespacesAndNewlines)
            return (.x, text.hasPrefix("http") ? text : "https://" + text)
        }
        if let ref = xiaoyuzhouID(from: input) {
            return (.xiaoyuzhou, "https://www.xiaoyuzhoufm.com/\(ref.isEpisode ? "episode" : "podcast")/\(ref.id)")
        }
        if let ref = applePodcastID(from: input) {
            let episode = ref.episode.map { "?i=\($0)" } ?? ""
            return (.applePodcasts, "https://podcasts.apple.com/\(ref.country ?? "us")/podcast/id\(ref.show)\(episode)")
        }
        return nil
    }

    /// A bilibili link (bilibili.com, b23.tv, bilibili.tv) or a bare `BV…`, `av…`, `ep…`, `ss…`, `md…`, `cheese/ep…` id.
    public static func isBilibili(_ input: String) -> Bool {
        let text = input.trimmingCharacters(in: .whitespacesAndNewlines)
        if text.wholeMatch(of: /(?i)(bv1[0-9a-z]{9}|av\d+|ep\d+|ss\d+|md\d+)/) != nil { return true }
        if text.wholeMatch(of: /cheese\/(ep|ss)\d+/) != nil { return true }
        guard let (_, host, _) = components(text) else { return false }
        return ["bilibili.com", "b23.tv", "bilibili.tv", "biliintl.com"].contains { host == $0 || host.hasSuffix("." + $0) }
    }

    /// A Xiaoyuzhou episode (`xiaoyuzhoufm.com/episode/<id>`) or show (`/podcast/<id>`) link.
    public static func xiaoyuzhouID(from input: String) -> (id: String, isEpisode: Bool)? {
        guard let (_, host, path) = components(input),
              host == "xiaoyuzhoufm.com" || host.hasSuffix(".xiaoyuzhoufm.com"),
              path.count >= 2, path[0] == "episode" || path[0] == "podcast",
              path[1].wholeMatch(of: /[A-Za-z0-9]{8,}/) != nil else { return nil }
        return (path[1], path[0] == "episode")
    }

    /// An Apple Podcasts link: `podcasts.apple.com/<cc>/podcast/<slug>/id<show>`, with `?i=<episode>` for one episode.
    public static func applePodcastID(from input: String) -> (show: String, episode: String?, country: String?)? {
        guard let (url, host, path) = components(input),
              host == "podcasts.apple.com" || host == "itunes.apple.com",
              path.contains("podcast"),
              let show = path.last(where: { $0.wholeMatch(of: /id\d+/) != nil }) else { return nil }
        let episode = Format.queryValue("i", in: url.absoluteString)
        let country = path.first.flatMap { $0.wholeMatch(of: /[A-Za-z]{2}/) != nil ? $0.lowercased() : nil }
        return (String(show.dropFirst(2)), episode.wholeMatch(of: /\d+/) != nil ? episode : nil, country)
    }

    /// The 11-character video id of a YouTube video link, or nil for anything else.
    public static func youtubeID(from input: String) -> String? {
        guard let (url, host, path) = components(input) else { return nil }
        var candidate: String?
        if host == "youtu.be" {
            candidate = path.first
        } else if host == "youtube.com" || host.hasSuffix(".youtube.com") || host.hasSuffix("youtube-nocookie.com") {
            if path.first == "watch" {
                candidate = Format.queryValue("v", in: url.absoluteString)
            } else if path.count >= 2, ["shorts", "embed", "live", "v", "e"].contains(path[0]) {
                candidate = path[1]
            }
        }
        guard let id = candidate, id.wholeMatch(of: /[A-Za-z0-9_-]{11}/) != nil else { return nil }
        return id
    }

    /// The post id of an X (Twitter) status link: `x.com/<user>/status/<id>`, `twitter.com/i/web/status/<id>`…
    public static func tweetID(from input: String) -> String? {
        guard let (_, host, path) = components(input) else { return nil }
        let bare = host.hasPrefix("www.") ? String(host.dropFirst(4)) : host.hasPrefix("mobile.") ? String(host.dropFirst(7)) : host
        guard bare == "x.com" || bare == "twitter.com",
              let i = path.firstIndex(of: "status"), path.indices.contains(i + 1),
              path[i + 1].wholeMatch(of: /\d+/) != nil else { return nil }
        return path[i + 1]
    }

    private static func components(_ input: String) -> (URL, String, [String])? {
        let text = input.trimmingCharacters(in: .whitespacesAndNewlines)
        guard let url = URL(string: text.hasPrefix("http") ? text : "https://" + text),
              let host = url.host()?.lowercased() else { return nil }
        return (url, host, url.pathComponents.filter { $0 != "/" })
    }
}
