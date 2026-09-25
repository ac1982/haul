import Foundation

/// Resolves a `MediaID` into the list of pages to download.
public enum InfoFetcher {
    public static func fetch(_ id: MediaID, useIntl: Bool, http: HTTPClient = .shared, session: Session) async throws -> VideoInfo {
        switch id {
        case .video(let aid): return try await NormalFetcher.fetch(aid: aid, http: http, session: session)
        case .episode(let ep):
            return useIntl
                ? try await IntlBangumiFetcher.fetch(epId: ep, http: http, session: session)
                : try await BangumiFetcher.fetch(epId: ep, http: http, session: session)
        case .cheese(let ep): return try await CheeseFetcher.fetch(epId: ep, http: http, session: session)
        case .space(let mid): return try await SpaceFetcher.fetch(mid: mid, http: http, session: session)
        case .collection(let biz): return try await MediaListFetcher.fetch(bizId: biz, kind: .collection, http: http, session: session)
        case .series(let biz): return try await MediaListFetcher.fetch(bizId: biz, kind: .series, http: http, session: session)
        case .favorites(let fid, let mid): return try await FavListFetcher.fetch(fid: fid, mid: mid, http: http, session: session)
        case .link(let site, _): throw HaulError("\(site.name) links are not read by the bilibili fetcher")
        }
    }

    static func dimension(_ node: JSON) -> String {
        let d = node["dimension"]
        guard d.exists, d["width"].exists else { return "" }
        return "\(d["width"].stringValue)x\(d["height"].stringValue)"
    }
}

enum NormalFetcher {
    static func fetch(aid: String, http: HTTPClient, session: Session) async throws -> VideoInfo {
        let json = try await http.getJSON("https://api.bilibili.com/x/web-interface/view?aid=\(aid)", session: session)
        let data = try json.get("data")
        let title = data["title"].stringValue
        let owner = data["owner"]
        let ownerMid = owner["mid"].stringValue
        let ownerName = owner["name"].stringValue
        let pubTime = data["pubdate"].int64 ?? 0
        let bvid = data["bvid"].stringValue
        let cid = data["cid"].stringValue
        let isInteractive = data["rights"]["is_stein_gate"].int == 1

        var pages: [Page] = []
        let pageNodes = try data.get("pages").array
        for p in pageNodes {
            pages.append(Page(index: p["page"].int ?? pages.count + 1, aid: aid, cid: p["cid"].stringValue, epid: "",
                              title: p["part"].stringValue.trimmingCharacters(in: .whitespaces),
                              duration: p["duration"].int ?? 0, resolution: InfoFetcher.dimension(p),
                              pubTime: pubTime, cover: "", desc: "", ownerName: ownerName, ownerMid: ownerMid))
        }

        if isInteractive {
            pages += try await interactivePages(aid: aid, bvid: bvid, cid: cid, pubTime: pubTime, ownerName: ownerName, ownerMid: ownerMid, http: http, session: session)
        }

        var bangumi = false
        let redirect = data["redirect_url"].stringValue
        if redirect.contains("bangumi"), let m = redirect.firstMatch(of: /ep(\d+)/) {
            bangumi = true
            // Licensed content normally has one page; then the episode id is what the playurl wants.
            if pageNodes.count == 1 {
                for i in pages.indices { pages[i].epid = String(m.1) }
            }
        }

        var info = VideoInfo(title: title.trimmingCharacters(in: .whitespaces), desc: data["desc"].stringValue.trimmingCharacters(in: .whitespaces),
                             cover: data["pic"].stringValue, pubTime: pubTime, pages: pages)
        info.isBangumi = bangumi
        info.isInteractive = isInteractive
        return info
    }

    /// Interactive videos hide their branches behind the player graph; each choice becomes a page, numbered from 2.
    private static func interactivePages(aid: String, bvid: String, cid: String, pubTime: Int64, ownerName: String, ownerMid: String,
                                         http: HTTPClient, session: Session) async throws -> [Page] {
        let playerSo = try await http.getString("https://api.bilibili.com/x/player.so?bvid=\(bvid)&id=cid:\(cid)", session: session)
        guard let m = playerSo.firstMatch(of: /<interaction>([\s\S]*?)<\/interaction>/), !m.1.isEmpty else {
            throw HaulError("Could not read the pages of this interactive video")
        }
        let interaction = try JSON.parse(Format.unescapeEntities(String(m.1)))
        guard let graph = interaction["graph_version"].int64 else { throw HaulError("Could not read the pages of this interactive video") }
        let edge = try await http.getJSON("https://api.bilibili.com/x/stein/edgeinfo_v2?graph_version=\(graph)&bvid=\(bvid)", session: session)
        var pages: [Page] = []
        var index = 2
        for question in edge["data"]["edges"]["questions"].array {
            for choice in question["choices"].array {
                pages.append(Page(index: index, aid: aid, cid: choice["cid"].stringValue, epid: "",
                                  title: choice["option"].stringValue.trimmingCharacters(in: .whitespaces), duration: 0, resolution: "",
                                  pubTime: pubTime, cover: "", desc: "", ownerName: ownerName, ownerMid: ownerMid))
                index += 1
            }
        }
        return pages
    }
}

enum BangumiFetcher {
    static func fetch(epId: String, http: HTTPClient, session: Session) async throws -> VideoInfo {
        let json = try await http.getJSON("https://\(session.epHost)/pgc/view/web/season?ep_id=\(epId)", session: session)
        let result = try json.get("result")
        var title = result["title"].stringValue
        let pubTimeText = result["publish"]["pub_time"].stringValue
        let pubTime = pubTimeText.isEmpty ? 0 : (Format.parseDateTime(pubTimeText) ?? 0)
        var episodes = try result.get("episodes")

        // The episode may live in a section (extras, PVs) rather than the main list.
        if episodes.array.isEmpty || !episodes.stringValue.contains("/ep\(epId)") {
            for section in result["section"].array where section.stringValue.contains("/ep\(epId)") {
                title += "[\(section["title"].stringValue)]"
                episodes = section["episodes"]
                break
            }
        }

        var pages: [Page] = []
        var index: String?
        var i = 1
        for ep in episodes.array {
            if ep["badge"].stringValue == "预告" { continue }
            let t = "\(ep["title"].stringValue) \(ep["long_title"].stringValue)".trimmingCharacters(in: .whitespaces)
            let p = Page(index: i, aid: ep["aid"].stringValue, cid: ep["cid"].stringValue, epid: ep["id"].stringValue,
                         title: t, duration: 0, resolution: InfoFetcher.dimension(ep), pubTime: ep["pub_time"].int64 ?? 0)
            if p.epid == epId { index = String(i) }
            pages.append(p)
            i += 1
        }

        var info = VideoInfo(title: title.trimmingCharacters(in: .whitespaces), desc: result["evaluate"].stringValue.trimmingCharacters(in: .whitespaces),
                             cover: result["cover"].stringValue, pubTime: pubTime, pages: pages)
        info.isBangumi = true
        info.isBangumiEnd = result["publish"]["is_finish"].int == 1
        info.index = index
        return info
    }
}

enum CheeseFetcher {
    static func fetch(epId: String, http: HTTPClient, session: Session) async throws -> VideoInfo {
        let json = try await http.getJSON("https://api.bilibili.com/pugv/view/web/season?ep_id=\(epId)", session: session)
        let data = try json.get("data")
        let ownerName = data["up_info"]["uname"].stringValue
        let ownerMid = data["up_info"]["mid"].stringValue
        var pages: [Page] = []
        var index: String?
        for ep in try data.get("episodes").array {
            let p = Page(index: ep["index"].int ?? pages.count + 1, aid: ep["aid"].stringValue, cid: ep["cid"].stringValue, epid: ep["id"].stringValue,
                         title: ep["title"].stringValue.trimmingCharacters(in: .whitespaces), duration: ep["duration"].int ?? 0, resolution: "",
                         pubTime: ep["release_date"].int64 ?? 0, cover: "", desc: "", ownerName: ownerName, ownerMid: ownerMid)
            if p.epid == epId { index = String(p.index) }
            pages.append(p)
        }
        var info = VideoInfo(title: data["title"].stringValue.trimmingCharacters(in: .whitespaces), desc: data["subtitle"].stringValue.trimmingCharacters(in: .whitespaces),
                             cover: data["cover"].stringValue, pubTime: pages.first?.pubTime ?? 0, pages: pages)
        info.isBangumi = true
        info.isCheese = true
        info.index = index
        return info
    }
}

enum IntlBangumiFetcher {
    static func fetch(epId: String, http: HTTPClient, session: Session) async throws -> VideoInfo {
        let host = session.isBiliPlus ? session.host : "api.bilibili.tv"
        var api = "https://\(host)/intl/gateway/v2/ogv/view/app/season?ep_id=\(epId)&platform=android&s_locale=zh_SG&mobi_app=bstar_a"
        if session.hasToken { api += "&access_key=\(session.token)" }
        let raw = try await http.getString(api, session: session).replacingOccurrences(of: "\\/", with: "/")
        let result = try JSON.parse(raw).get("result")
        var cover = result["cover"].stringValue
        var title = result["title"].stringValue
        var desc = result["evaluate"].stringValue

        if cover.isEmpty {
            let html = try await http.getString("https://bangumi.bilibili.com/anime/\(result["season_id"].stringValue)", session: session)
            if let m = html.firstMatch(of: /window\.__INITIAL_STATE__=([\s\S]*?);\(function\(\)/), let state = try? JSON.parse(String(m.1)) {
                cover = state["mediaInfo"]["cover"].stringValue
                title = state["mediaInfo"]["title"].stringValue
                desc = state["mediaInfo"]["evaluate"].stringValue
            }
        }

        let pubTimeText = result["publish"]["pub_time"].stringValue
        let pubTime = pubTimeText.isEmpty ? 0 : (Format.parseDateTime(pubTimeText) ?? 0)
        var episodes = result["episodes"].array
        for module in result["modules"].array where module.stringValue.contains("/\(epId)") {
            episodes = module["data"]["episodes"].array
            break
        }

        var pages: [Page] = []
        var index: String?
        var i = 1
        for ep in episodes {
            if ep["badge"].stringValue == "预告" { continue }
            let t = "\(ep["title"].stringValue) \(ep["long_title"].stringValue)".trimmingCharacters(in: .whitespaces)
            let p = Page(index: i, aid: ep["aid"].stringValue, cid: ep["cid"].stringValue, epid: ep["id"].stringValue,
                         title: t, duration: 0, resolution: InfoFetcher.dimension(ep), pubTime: ep["pub_time"].int64 ?? 0)
            if p.epid == epId { index = String(i) }
            pages.append(p)
            i += 1
        }

        var info = VideoInfo(title: title.trimmingCharacters(in: .whitespaces), desc: desc.trimmingCharacters(in: .whitespaces), cover: cover, pubTime: pubTime, pages: pages)
        info.isBangumi = true
        info.index = index
        return info
    }
}

enum FavListFetcher {
    static func fetch(fid: String, mid: String, http: HTTPClient, session: Session) async throws -> VideoInfo {
        var favId = fid
        if favId.isEmpty {
            let list = try await http.getJSON("https://api.bilibili.com/x/v3/fav/folder/created/list-all?up_mid=\(mid)", session: session)
            favId = try list.get("data")["list"][0]["id"].stringValue
            if favId.isEmpty { throw HaulError("This user has no public favorites list") }
        }
        let pageSize = 20
        func url(_ pn: Int) -> String {
            "https://api.bilibili.com/x/v3/fav/resource/list?media_id=\(favId)&pn=\(pn)&ps=\(pageSize)&order=mtime&type=2&tid=0&platform=web"
        }
        let first = try await http.getJSON(url(1), session: session)
        let data = try first.get("data")
        let total = data["info"]["media_count"].int ?? 0
        let totalPages = Int((Double(total) / Double(pageSize)).rounded(.up))
        var medias = data["medias"].array
        if totalPages > 1 {
            for pn in 2...totalPages {
                medias += try await http.getJSON(url(pn), session: session)["data"]["medias"].array
            }
        }

        var pages: [Page] = []
        var index = 1
        for m in medias {
            if m["attr"].int != 0 { continue }  // unavailable
            if (m["page"].int ?? 1) > 1 {
                let sub = try await NormalFetcher.fetch(aid: m["id"].stringValue, http: http, session: session)
                for item in sub.pages {
                    var p = item
                    p.index = index
                    p.title = "\(m["title"].stringValue)_P\(item.index)_\(item.title)"
                    p.cover = sub.cover
                    p.desc = m["intro"].stringValue
                    if !pages.contains(p) { pages.append(p); index += 1 }
                }
            } else {
                let p = Page(index: index, aid: m["id"].stringValue, cid: m["ugc"]["first_cid"].stringValue, epid: "",
                             title: m["title"].stringValue, duration: m["duration"].int ?? 0, resolution: "",
                             pubTime: m["pubtime"].int64 ?? 0, cover: m["cover"].stringValue, desc: m["intro"].stringValue,
                             ownerName: m["upper"]["name"].stringValue, ownerMid: m["upper"]["mid"].stringValue)
                if !pages.contains(p) { pages.append(p); index += 1 }
            }
        }

        return VideoInfo(title: data["info"]["title"].stringValue.trimmingCharacters(in: .whitespaces),
                         desc: data["info"]["intro"].stringValue.trimmingCharacters(in: .whitespaces),
                         cover: "", pubTime: data["info"]["ctime"].int64 ?? 0, pages: pages)
    }
}

enum MediaListFetcher {
    enum Kind { case collection, series
        var type: Int { self == .collection ? 8 : 5 }
    }

    static func fetch(bizId: String, kind: Kind, http: HTTPClient, session: Session) async throws -> VideoInfo {
        let info = try await http.getJSON("https://api.bilibili.com/x/v1/medialist/info?type=\(kind.type)&biz_id=\(bizId)&tid=0", session: session)
        let data = info["data"]
        guard data.isObject else {
            // A deleted/private collection, or a series that was mistaken for one — try the other kind before giving up.
            if kind == .collection, let asSeries = try? await fetch(bizId: bizId, kind: .series, http: http, session: session) { return asSeries }
            throw HaulError("Could not read the \(kind == .collection ? "collection" : "series") (code \(info["code"].stringValue)): \(info["message"].stringValue)")
        }

        var pages: [Page] = []
        var hasMore = true
        var oid = ""
        var index = 1
        while hasMore {
            let desc = kind == .collection ? "false" : "true"
            let extra = kind == .series ? "&bvid=" : ""
            let listURL = "https://api.bilibili.com/x/v2/medialist/resource/list?type=\(kind.type)&oid=\(oid)&otype=2&biz_id=\(bizId)\(extra)&with_current=true&mobi_app=web&ps=20&direction=false&sort_field=1&tid=0&desc=\(desc)"
            let list = try await http.getJSON(listURL, session: session)
            let ldata = list["data"]
            guard ldata.isObject else {
                throw HaulError("Could not read the video list (code \(list["code"].stringValue)): \(list["message"].stringValue)")
            }
            hasMore = ldata["has_more"].bool ?? false
            for m in ldata["media_list"].array {
                if m["attr"].exists, m["attr"].int != 0 { continue }
                let pageCount = m["page"].int ?? 1
                for pg in m["pages"].array {
                    let title = pageCount == 1 ? m["title"].stringValue : "\(m["title"].stringValue)_P\(pg["page"].stringValue)_\(pg["title"].stringValue)"
                    let p = Page(index: index, aid: m["id"].stringValue, cid: pg["id"].stringValue, epid: "", title: title,
                                 duration: pg["duration"].int ?? 0, resolution: InfoFetcher.dimension(pg), pubTime: m["pubtime"].int64 ?? 0,
                                 cover: m["cover"].stringValue, desc: m["intro"].stringValue,
                                 ownerName: m["upper"]["name"].stringValue, ownerMid: m["upper"]["mid"].stringValue)
                    if !pages.contains(p) { pages.append(p); index += 1 }
                }
                oid = m["id"].stringValue
            }
        }

        return VideoInfo(title: data["title"].stringValue.trimmingCharacters(in: .whitespaces),
                         desc: data["intro"].stringValue.trimmingCharacters(in: .whitespaces),
                         cover: "", pubTime: data["ctime"].int64 ?? 0, pages: pages)
    }
}

/// All uploads of one user. Each video costs one extra request (the list has no cids), so this is paced.
enum SpaceFetcher {
    static func fetch(mid: String, http: HTTPClient, session s: Session) async throws -> VideoInfo {
        var session = s
        if session.wbiKey.isEmpty { session.wbiKey = try await Auth.fetchWbiKey(http: http, session: session) }
        let userInfo = try await http.getJSON("https://api.live.bilibili.com/live_user/v1/Master/info?uid=\(mid)", session: session)
        let userName = userInfo["data"]["info"]["uname"].stringValue

        let pageSize = 50
        func listURL(_ pn: Int) -> String {
            let q = Signing.wbiSign("mid=\(mid)&order=pubdate&pn=\(pn)&ps=\(pageSize)&tid=0&wts=\(Format.unixSeconds())", key: session.wbiKey)
            return "https://api.bilibili.com/x/space/wbi/arc/search?\(q)"
        }
        let first = try await http.getJSON(listURL(1), session: session)
        var items = try first.get("data")["list"]["vlist"].array
        let total = first["data"]["page"]["count"].int ?? items.count
        let totalPages = Int((Double(total) / Double(pageSize)).rounded(.up))
        if totalPages > 1 {
            for pn in 2...totalPages {
                items += try await http.getJSON(listURL(pn), session: session)["data"]["list"]["vlist"].array
            }
        }
        Log.status("\(items.count) uploads; reading their pages")

        var pages: [Page] = []
        var index = 1
        for (n, item) in items.enumerated() {
            let aid = item["aid"].stringValue
            if n > 0 { try await Task.sleep(for: .milliseconds(200)) }
            if (n + 1) % 20 == 0 || n + 1 == items.count { Log.status("Uploads read \(n + 1)/\(items.count)") }
            guard let sub = try? await NormalFetcher.fetch(aid: aid, http: http, session: session) else {
                Log.warn("Skipping av\(aid): could not read it")
                continue
            }
            let multi = sub.pages.count > 1
            for item in sub.pages {
                var p = item
                p.index = index
                p.title = multi ? "\(sub.title)_P\(item.index)_\(item.title)" : sub.title
                p.cover = sub.cover
                p.desc = sub.desc
                if !pages.contains(p) { pages.append(p); index += 1 }
            }
        }

        return VideoInfo(title: "\(userName)'s uploads", desc: "", cover: "", pubTime: 0, pages: pages)
    }
}
