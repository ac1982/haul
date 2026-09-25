import Foundation

/// What a user's input resolved to.
public enum MediaID: Sendable, CustomStringConvertible, Equatable {
    case video(aid: String)
    case episode(epId: String)
    case cheese(epId: String)
    case space(mid: String)
    case collection(bizId: String)
    case series(bizId: String)
    case favorites(fid: String, mid: String)
    /// Any other site's link, as its source reads it.
    case link(site: Site, url: String)

    public var description: String {
        switch self {
        case .video(let a): return a
        case .episode(let e): return "ep:\(e)"
        case .cheese(let e): return "cheese:\(e)"
        case .space(let m): return "mid:\(m)"
        case .collection(let b): return "listBizId:\(b)"
        case .series(let b): return "seriesBizId:\(b)"
        case .favorites(let f, let m): return "favId:\(f):\(m)"
        case .link(let site, let url): return "\(site.rawValue):\(url)"
        }
    }

    public var site: Site { if case .link(let site, _) = self { return site }; return .bilibili }
    public var isEpisode: Bool { if case .episode = self { return true }; return false }
    public var isCheese: Bool { if case .cheese = self { return true }; return false }
    /// PGC content (bangumi or course) — the playurl endpoints differ.
    public var isPGC: Bool { isEpisode || isCheese }
}

/// Turns a supported link into a `MediaID`: bilibili links and ids (`BV…`, `av…`, `ep…`, `ss…`, `md…`, `cheese/ep…`)
/// into bilibili ids, every other site's link into `.link`.
public enum IDResolver {
    public static func resolve(_ rawInput: String, http: HTTPClient = .shared, session: Session) async throws -> MediaID {
        guard let (site, url) = Site.match(rawInput) else { throw HaulError.unsupported(rawInput) }
        if site != .bilibili { return .link(site: site, url: url) }
        var input = rawInput.trimmingCharacters(in: .whitespacesAndNewlines)
        let lower = input.lowercased()

        if input.hasPrefix("http") {
            if input.contains("b23.tv") {
                let expanded = try await http.finalURL(input, session: session)
                if expanded == input { throw HaulError("b23.tv link redirects to itself") }
                input = expanded
            }
            return try await fixVideoID(resolveURL(input, http: http, session: session), http: http, session: session)
        }
        if lower.hasPrefix("bv") {
            return try await fixVideoID(.video(aid: String(try BVConverter.decode(String(input.dropFirst(3))))), http: http, session: session)
        }
        if lower.hasPrefix("av") {
            return try await fixVideoID(.video(aid: String(input.dropFirst(2))), http: http, session: session)
        }
        if input.hasPrefix("cheese/") {
            if let ep = input.firstMatch(of: /\/ep(\d+)/) { return .cheese(epId: String(ep.1)) }
            if let ss = input.firstMatch(of: /\/ss(\d+)/) { return .cheese(epId: try await cheeseFirstEp(seasonId: String(ss.1), http: http, session: session)) }
            throw HaulError.unsupported(rawInput)
        }
        if input.hasPrefix("ep") { return .episode(epId: String(input.dropFirst(2))) }
        if input.hasPrefix("ss") { return .episode(epId: try await bangumiFirstEp(seasonId: String(input.dropFirst(2)), http: http, session: session)) }
        if input.hasPrefix("md"), let md = input.firstMatch(of: /md(\d+)/) {
            return .episode(epId: try await bangumiNewEp(mediaId: String(md.1), http: http, session: session))
        }
        throw HaulError.unsupported(rawInput)
    }

    private static func resolveURL(_ input: String, http: HTTPClient, session: Session) async throws -> MediaID {
        if input.contains("video/av"), let m = input.firstMatch(of: /av(\d+)/) {
            return .video(aid: String(m.1))
        }
        if input.lowercased().contains("video/bv"), let m = input.firstMatch(of: /[Bb][Vv]1(\w+)/) {
            return .video(aid: String(try BVConverter.decode(String(m.1))))
        }
        if input.contains("/cheese/") {
            if let ep = input.firstMatch(of: /\/ep(\d+)/) { return .cheese(epId: String(ep.1)) }
            if let ss = input.firstMatch(of: /\/ss(\d+)/) { return .cheese(epId: try await cheeseFirstEp(seasonId: String(ss.1), http: http, session: session)) }
            throw HaulError.input("Unrecognised bilibili course link: \(input)")
        }
        if let ep = input.firstMatch(of: /\/ep(\d+)/) { return .episode(epId: String(ep.1)) }
        if let ss = input.firstMatch(of: /\/ss(\d+)/) {
            return .episode(epId: try await bangumiFirstEp(seasonId: String(ss.1), http: http, session: session))
        }
        if input.contains("/medialist/"), input.contains("business_id=") {
            let biz = Format.queryValue("business_id", in: input)
            if input.contains("business=space_collection") { return .collection(bizId: biz) }
            if input.contains("business=space_series") { return .series(bizId: biz) }
        }
        if input.contains("/channel/collectiondetail?sid=") { return .collection(bizId: Format.queryValue("sid", in: input)) }
        if input.contains("/channel/seriesdetail?sid=") { return .series(bizId: Format.queryValue("sid", in: input)) }
        if input.contains("/space.bilibili.com/"), input.contains("/lists/") {
            // New-style space lists: .../lists/<sid>?type=season|series
            let path = input.split(whereSeparator: { $0 == "?" || $0 == "#" }).first.map(String.init) ?? input
            let sid = path.split(separator: "/").last.map(String.init) ?? ""
            return Format.queryValue("type", in: input).lowercased() == "series" ? .series(bizId: sid) : .collection(bizId: sid)
        }
        if input.contains("/space.bilibili.com/"), let m = input.firstMatch(of: /space\.bilibili\.com\/(\d+)/) {
            let mid = String(m.1)
            if input.contains("/favlist") { return .favorites(fid: Format.queryValue("fid", in: input), mid: mid) }
            return .space(mid: mid)
        }
        if input.contains("ep_id=") { return .episode(epId: Format.queryValue("ep_id", in: input)) }
        if let m = input.firstMatch(of: /\.bilibili\.tv\/\w+\/play\/\d+\/(\d+)/) { return .episode(epId: String(m.1)) }
        if let m = input.firstMatch(of: /bangumi\/media\/md(\d+)/) {
            return .episode(epId: try await bangumiNewEp(mediaId: String(m.1), http: http, session: session))
        }
        // Last resort: scrape the page's initial state for an episode list.
        let html = try await http.getString(input, session: session)
        guard let m = html.firstMatch(of: /window\.__INITIAL_STATE__=([\s\S]*?);\(function\(\)/) else { throw HaulError.unsupported(input) }
        let state = try JSON.parse(String(m.1))
        guard let ep = state["epList"][0]["id"].int64 else { throw HaulError("No episode found on the page") }
        return .episode(epId: String(ep))
    }

    /// A plain av id may actually be licensed content; follow the video page's redirect to find its episode.
    private static func fixVideoID(_ id: MediaID, http: HTTPClient, session: Session) async throws -> MediaID {
        guard case .video(let aid) = id, aid.allSatisfy(\.isNumber), !aid.isEmpty else { return id }
        let location = try await http.finalURL("https://www.bilibili.com/video/av\(aid)/", session: session)
        if let m = location.firstMatch(of: /\/ep(\d+)/) { return .episode(epId: String(m.1)) }
        return id
    }


    static func cheeseFirstEp(seasonId: String, http: HTTPClient, session: Session) async throws -> String {
        let json = try await http.getJSON("https://api.bilibili.com/pugv/view/web/season?season_id=\(seasonId)", session: session)
        guard let ep = json["data"]["episodes"][0]["id"].int64 else { throw HaulError("Course ss\(seasonId) has no episodes") }
        return String(ep)
    }

    static func bangumiFirstEp(seasonId: String, http: HTTPClient, session: Session) async throws -> String {
        let json = try await http.getJSON("https://\(session.epHost)/pgc/view/web/season?season_id=\(seasonId)", session: session)
        guard let ep = json["result"]["episodes"][0]["id"].int64 else { throw HaulError("Season ss\(seasonId) has no episodes") }
        return String(ep)
    }

    static func bangumiNewEp(mediaId: String, http: HTTPClient, session: Session) async throws -> String {
        let json = try await http.getJSON("https://api.bilibili.com/pgc/review/user?media_id=\(mediaId)", session: session)
        guard let ep = json["result"]["media"]["new_ep"]["id"].int64 else { throw HaulError("md\(mediaId) has no episodes") }
        return String(ep)
    }
}
