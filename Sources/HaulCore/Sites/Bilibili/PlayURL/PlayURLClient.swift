import Foundation

/// What we need to ask a playurl endpoint about one page.
public struct PlayURLRequest: Sendable {
    public var id: MediaID
    public var aid: String
    public var cid: String
    public var epId: String
    public var api: APIType
    /// Preferred codec name (`HEVC`/`AVC`/`AV1`), empty for server default. Only the APP API honours it.
    public var encoding: String

    public init(id: MediaID, aid: String, cid: String, epId: String, api: APIType, encoding: String) {
        self.id = id; self.aid = aid; self.cid = cid; self.epId = epId; self.api = api; self.encoding = encoding
    }

    var isCheese: Bool { id.isCheese }
    /// PGC (bangumi or course): the endpoints and parameters differ.
    var isPGC: Bool { id.isPGC }
    /// Bangumi proper (not a course).
    var isBangumi: Bool { id.isEpisode }
}

/// Talks to the four playurl backends and normalises their answers into `ParsedTracks`.
public enum PlayURLClient {
    // MARK: fetching

    static func fetchPlayJSON(_ r: PlayURLRequest, qn: String, http: HTTPClient, session: Session) async throws -> String {
        Log.debug("aid=\(r.aid),cid=\(r.cid),epId=\(r.epId),api=\(r.api),qn=\(qn),pgc=\(r.isPGC),cheese=\(r.isCheese)")
        switch r.api {
        case .intl:
            return try await fetchIntlJSON(r, qn: qn, code: "0", http: http, session: session)
        case .app:
            return try await AppAPI.playView(aid: r.aid, cid: r.cid, epId: r.epId, qn: qn, isPGC: r.isPGC, encoding: r.encoding, http: http, session: session)
        case .tv:
            let base = r.isPGC ? "\(session.tvHost)/pgc/player/api/playurltv" : "\(session.tvHost)/x/tv/playurl"
            var q = session.hasToken ? "access_key=\(session.token)&" : ""
            q += "appkey=\(Signing.tvAppKey)&build=106500&cid=\(r.cid)&device=android"
            if r.isPGC { q += "&ep_id=\(r.epId)&expire=0" }
            q += "&fnval=4048&fnver=0&fourk=1&mid=0&mobi_app=android_tv_yst"
            q += "&object_id=\(r.aid)&platform=android&playurl_type=1&qn=\(qn)&ts=\(Format.unixSeconds())"
            var api = "https://\(base)?\(q)&sign=\(Signing.appSign(q, secret: Signing.tvAppSecret))"
            if r.isCheese { api = api.replacingOccurrences(of: "/pgc/", with: "/pugv/") }
            return try await http.getString(api, session: session)
        case .web:
            let base = r.isPGC ? "\(session.host)/pgc/player/web/v2/playurl" : "api.bilibili.com/x/player/wbi/playurl"
            var q = "support_multi_audio=true&from_client=BROWSER&avid=\(r.aid)&cid=\(r.cid)&fnval=4048&fnver=0&fourk=1"
            if !session.area.isEmpty { q += "&access_key=\(session.token)&area=\(session.area)" }
            q += "&otype=json&qn=\(qn)"
            if r.isPGC { q += "&module=bangumi&ep_id=\(r.epId)&session=" }
            if !session.hasCookie { q += "&try_look=1" }
            q += "&wts=\(Format.unixSeconds())"
            var api = "https://\(base)?" + (r.isPGC ? q : Signing.wbiSign(q, key: session.wbiKey))
            if r.isCheese { api = api.replacingOccurrences(of: "/pgc/", with: "/pugv/") }
            var json = try await http.getString(api, session: session)
            if json.contains("\"大会员专享限制\"") {
                Log.warn("This video needs a bilibili premium membership; trying the web page")
                let html = try await http.getString("https://www.bilibili.com/bangumi/play/ep\(r.epId)", session: session)
                json = html.firstMatch(of: /window\.__playinfo__=([\s\S]*?)<\/script>/).map { String($0.1) } ?? json
            }
            return json
        }
    }

    static func fetchIntlJSON(_ r: PlayURLRequest, qn: String, code: String, http: HTTPClient, session: Session) async throws -> String {
        let plus = session.isBiliPlus
        var q = session.hasToken ? "access_key=\(session.token)&" : ""
        q += "aid=\(r.aid)"
        if plus { q += "&appkey=\(Signing.biliPlusAppKey)&area=\(session.area.isEmpty ? "th" : session.area)" }
        q += "&cid=\(r.cid)&ep_id=\(r.epId)&platform=android&prefer_code_type=\(code)&qn=\(qn)"
        if plus { q += "&ts=\(Format.unixSeconds())" }
        q += "&s_locale=zh_SG"
        let api = "https://\(plus ? session.host : "api.biliintl.com")/intl/gateway/v2/ogv/playurl?"
            + (plus ? "\(q)&sign=\(Signing.appSign(q, secret: Signing.biliPlusSecret))" : q)
        return try await http.getString(api, session: session)
    }

    // MARK: parsing

    /// Prefer a URL that is not a bare `ip:port` (PCDN) host.
    static func pickURL(_ node: JSON) -> String {
        var urls = [node["base_url"].stringValue]
        if node["backup_url"].isArray { urls += node["backup_url"].array.map(\.stringValue) }
        return urls.first { $0.firstMatch(of: /http.*:\d+/) == nil } ?? urls[0]
    }

    static func audioCodec(_ raw: String) -> String {
        switch raw {
        case "mp4a.40.2", "mp4a.40.5": return "M4A"
        case "ec-3": return "E-AC-3"
        case "fLaC": return "FLAC"
        default: return raw
        }
    }

    private static func audioTrack(_ node: JSON, duration: Int, codec: String? = nil) -> AudioTrack {
        let id = node["id"].stringValue
        return AudioTrack(id: id, quality: id, baseURL: pickURL(node), codec: codec ?? audioCodec(node["codecs"].stringValue),
                          bandwidth: (node["bandwidth"].int64 ?? 0) / 1000, duration: duration)
    }

    /// The node that holds `dash`/`durl`: `data`, `result`, or `result.video_info` (v2).
    private static func payloadRoot(_ root: JSON, raw: String) -> JSON {
        if raw.contains("\"result\":{") {
            return raw.contains("\"video_info\":{") ? root["result"]["video_info"] : root["result"]
        }
        if raw.contains("\"data\":{") { return root["data"] }
        return root
    }

    public static func extractTracks(_ r: PlayURLRequest, qn: String = "0", http: HTTPClient = .shared, session: Session) async throws -> ParsedTracks {
        var result = ParsedTracks()
        result.rawJSON = try await fetchPlayJSON(r, qn: qn, http: http, session: session)
        Log.debug(result.rawJSON)

        // International (bstar) API: stream_list with dash_video; two passes, one per codec preference.
        if result.rawJSON.contains("\"stream_list\"") {
            for code in ["0", "1"] {
                if code == "1" { result.rawJSON = try await fetchIntlJSON(r, qn: qn, code: code, http: http, session: session) }
                let vi = try JSON.parse(result.rawJSON)["data"]["video_info"]
                let dur = (vi["timelength"].int ?? 0) / 1000
                for stream in vi["stream_list"].array {
                    let dv = stream["dash_video"]
                    guard dv.exists, !dv["base_url"].stringValue.isEmpty else { continue }
                    let id = stream["stream_info"]["quality"].stringValue
                    let v = VideoTrack(id: id, quality: Quality.name(id), baseURL: pickURL(dv), codec: Quality.videoCodec(dv["codecid"].stringValue),
                                       bandwidth: (dv["bandwidth"].int64 ?? 0) / 1000, duration: dur, size: dv["size"].double ?? 0)
                    if !result.video.contains(v) { result.video.append(v) }
                }
                for node in vi["dash_audio"].array {
                    let a = audioTrack(node, duration: dur, codec: "M4A")
                    if !result.audio.contains(a) { result.audio.append(a) }
                }
            }
            return result
        }

        var root = try JSON.parse(result.rawJSON)
        var payload = payloadRoot(root, raw: result.rawJSON)
        let isApp = r.api == .app
        let isTV = r.api == .tv

        if result.rawJSON.contains("\"dash\":{") {
            var duration = payload["dash"]["duration"].int ?? 0
            if let t = payload["timelength"].int { duration = t / 1000 }

            // Pass 1 with the requested qn, pass 2 with the maximum qn to surface tracks the first answer hid
            // (the "no re-encode" originals). APP responses already contain everything.
            let passes = isApp ? 1 : 2
            var audioNodes: [JSON] = []
            var backgroundNodes: [JSON] = []
            var roleNodes: [JSON] = []
            for pass in 0..<passes {
                if pass == 1 {
                    result.rawJSON = try await fetchPlayJSON(r, qn: Quality.maxID, http: http, session: session)
                    root = try JSON.parse(result.rawJSON)
                    payload = payloadRoot(root, raw: result.rawJSON)
                }
                let dash = payload["dash"]
                audioNodes = dash["audio"].array
                if isApp && r.isBangumi {
                    backgroundNodes = root["dubbing_info"]["background_audio"].array
                    roleNodes = root["dubbing_info"]["role_audio_list"].array
                }
                if !isTV {
                    if dash["dolby"]["audio"].isArray { audioNodes += dash["dolby"]["audio"].array }
                    if dash["flac"]["audio"].isObject { audioNodes.append(dash["flac"]["audio"]) }
                }
                for node in dash["video"].array {
                    let id = node["id"].stringValue
                    var v = VideoTrack(id: id, quality: Quality.name(id), baseURL: pickURL(node), codec: Quality.videoCodec(node["codecid"].stringValue),
                                       bandwidth: (node["bandwidth"].int64 ?? 0) / 1000, duration: duration, size: node["size"].double ?? 0)
                    if !isTV && !isApp {
                        v.resolution = "\(node["width"].stringValue)x\(node["height"].stringValue)"
                        v.fps = node["frame_rate"].stringValue
                    }
                    if !result.video.contains(v) { result.video.append(v) }
                }
            }

            result.audio = audioNodes.map { audioTrack($0, duration: duration) }
            if !backgroundNodes.isEmpty && !roleNodes.isEmpty {
                result.backgroundAudio = backgroundNodes.map { audioTrack($0, duration: duration) }
                for role in roleNodes {
                    let tracks = role["audio"].array.map { audioTrack($0, duration: duration) }
                    result.roleAudio.append(RoleAudio(title: role["title"].stringValue, personName: role["person_name"].stringValue,
                                                      path: "\(r.aid)/\(r.aid).\(r.cid).\(role["audio_id"].stringValue).m4a", tracks: tracks))
                }
            }
        } else if result.rawJSON.contains("\"durl\":[") {
            // FLV: segments at the highest quality the account can see.
            result.rawJSON = try await fetchPlayJSON(r, qn: Quality.maxID, http: http, session: session)
            root = try JSON.parse(result.rawJSON)
            payload = payloadRoot(root, raw: result.rawJSON)
            var size = 0.0
            var length = 0.0
            for node in payload["durl"].array {
                result.clips.append(node["url"].stringValue)
                size += node["size"].double ?? 0
                length += node["length"].double ?? 0
            }
            if payload["qn_extras"].isArray {
                result.qualities = payload["qn_extras"].array.map { $0["qn"].stringValue }
            } else if payload["accept_quality"].isArray {
                result.qualities = payload["accept_quality"].array.map(\.stringValue).filter { !$0.isEmpty }
            }
            let quality = payload["quality"].stringValue
            let v = VideoTrack(id: quality, quality: Quality.name(quality), baseURL: "", codec: Quality.videoCodec(payload["video_codecid"].stringValue),
                               bandwidth: 0, duration: Int(length) / 1000, size: size)
            if !result.video.contains(v) { result.video.append(v) }
        }

        // Bangumi OP/ED markers become chapters: [main] → opening → main → ending.
        if r.isBangumi, payload["clip_info_list"].isArray {
            let clips = payload["clip_info_list"].array.map {
                ViewPoint(title: $0["toastText"].stringValue.replacingOccurrences(of: "即将跳过", with: ""), start: $0["start"].int ?? 0, end: $0["end"].int ?? 0)
            }.sorted { $0.start < $1.start }
            var points: [ViewPoint] = []
            var lastEnd = 0
            for c in clips {
                if lastEnd < c.start { points.append(ViewPoint(title: "Main", start: lastEnd, end: c.start)) }
                points.append(c)
                lastEnd = c.end
            }
            result.extraPoints = points
        }

        return result
    }
}
