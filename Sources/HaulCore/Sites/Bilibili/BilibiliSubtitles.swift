import Foundation

/// bilibili subtitle discovery across the web, app and international endpoints.
extension Subtitles {
    public static func list(aid: String, cid: String, epId: String, index: Int, intl: Bool,
                            http: HTTPClient = .shared, session: Session) async -> [SubtitleInfo] {
        var found: [SubtitleInfo]?
        if intl {
            found = await intlWeb(aid: aid, cid: cid, epId: epId, http: http, session: session)
            if found == nil { found = await intlApp(aid: aid, cid: cid, epId: epId, index: index, http: http, session: session) }
        } else if !session.hasCookie {
            // Logged out, only the app endpoint still answers.
            found = await app(aid: aid, cid: cid, http: http)
        } else {
            found = await webPlayer(aid: aid, cid: cid, http: http, session: session)
            if found == nil { found = await webView(aid: aid, cid: cid, http: http, session: session) }
            if found == nil { found = await app(aid: aid, cid: cid, http: http) }
        }
        return (found ?? []).map { s in
            var s = s
            if s.url.hasPrefix("//") { s.url = "https:" + s.url }
            return s
        }
    }

    private static func validate(_ subs: [SubtitleInfo]) -> [SubtitleInfo]? {
        subs.contains { $0.url.isEmpty } ? nil : subs
    }

    private static func intlWeb(aid: String, cid: String, epId: String, http: HTTPClient, session: Session) async -> [SubtitleInfo]? {
        let host = session.epHost == "api.bilibili.com" ? "api.biliintl.com" : session.epHost
        guard let json = try? await http.getJSON("https://\(host)/intl/gateway/web/v2/subtitle?episode_id=\(epId)", session: session),
              let subs = try? json.get("data").get("subtitles").array else { return nil }
        return validate(subs.map {
            let lan = $0["lang_key"].stringValue
            let url = $0["url"].stringValue
            return SubtitleInfo(language: lan, url: url, path: "\(aid)/\(aid).\(cid).\(lan)\(url.contains(".json") ? ".srt" : ".ass")")
        })
    }

    private static func intlApp(aid: String, cid: String, epId: String, index: Int, http: HTTPClient, session: Session) async -> [SubtitleInfo]? {
        let host = session.isBiliPlus ? session.host : "api.bilibili.tv"
        var api = "https://\(host)/intl/gateway/v2/ogv/view/app/season?ep_id=\(epId)&platform=android&s_locale=zh_SG"
        if session.hasToken { api += "&access_key=\(session.token)" }
        guard let json = try? await http.getJSON(api, session: session),
              let subs = try? json.get("result").get("modules")[0].get("data").get("episodes")[index - 1].get("subtitles").array else { return nil }
        return validate(subs.map {
            let lan = $0["key"].stringValue
            let url = $0["url"].stringValue.replacingOccurrences(of: "\\/", with: "/")
            return SubtitleInfo(language: lan, url: url, path: "\(aid)/\(aid).\(cid).\(lan)\(url.contains(".json") ? ".srt" : ".ass")")
        })
    }

    private static func webView(aid: String, cid: String, http: HTTPClient, session: Session) async -> [SubtitleInfo]? {
        guard let json = try? await http.getJSON("https://api.bilibili.com/x/web-interface/view?aid=\(aid)&cid=\(cid)", session: session),
              let subs = try? json.get("data").get("subtitle").get("list").array else { return nil }
        return validate(subs.map {
            let lan = $0["lan"].stringValue
            return SubtitleInfo(language: lan, url: $0["subtitle_url"].stringValue, path: "\(aid)/\(aid).\(cid).\(lan).srt")
        })
    }

    private static func webPlayer(aid: String, cid: String, http: HTTPClient, session: Session) async -> [SubtitleInfo]? {
        guard let json = try? await http.getJSON("https://api.bilibili.com/x/player/wbi/v2?cid=\(cid)&aid=\(aid)", session: session),
              let subs = try? json.get("data").get("subtitle").get("subtitles").array else { return nil }
        return validate(subs.map {
            let lan = $0["lan"].stringValue
            return SubtitleInfo(language: lan, url: $0["subtitle_url"].stringValue, path: "\(aid)/\(aid).\(cid).\(lan).srt")
        })
    }

    private static func app(aid: String, cid: String, http: HTTPClient) async -> [SubtitleInfo]? {
        guard let reply = try? await AppAPI.dmView(aid: aid, cid: cid, http: http) else { return nil }
        guard reply.hasSubtitle else { return [] }
        return validate(reply.subtitle.subtitles.map {
            SubtitleInfo(language: $0.lan, url: $0.subtitleURL, path: "\(aid)/\(aid).\(cid).\($0.lan).srt")
        })
    }
}
