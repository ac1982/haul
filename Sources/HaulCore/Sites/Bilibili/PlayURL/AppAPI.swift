import Foundation
import SwiftProtobuf

/// The Android app's gRPC PlayView, spoken over plain HTTPS POSTs with gRPC framing.
/// The reply is reshaped into the web API's JSON so one parser serves both.
enum AppAPI {
    static let ugcEndpoint = "https://grpc.biliapi.net/bilibili.app.playurl.v1.PlayURL/PlayView"
    static let pgcEndpoint = "https://app.bilibili.com/bilibili.pgc.gateway.player.v2.PlayURL/PlayView"
    static let dmViewEndpoint = "https://app.biliapi.net/bilibili.community.service.dm.v1.DM/DmView"

    // The device we claim to be. `build` must be recent enough for the dubbing fields to be present.
    private static let dalvikVersion = "2.1.0"
    private static let osVersion = "11"
    private static let brand = "M2012K11AC"
    private static let model = "Build/RKQ1.200826.002"
    private static let appVersion = "7.32.0"
    private static let build: Int32 = 7_320_200
    private static let channel = "xiaomi_cn_tv.danmaku.bili_zm20200902"
    private static let networkOid = "46007"
    private static let cronet = "1.36.1"
    private static let buvid = ""
    private static let mobiApp = "android"
    private static let appKey = "android64"
    private static let sessionID = "dedf8669"
    private static let platform = "android"
    private static let env = "prod"
    private static let appID: Int32 = 1
    private static let region = "CN"
    private static let language = "zh"

    static func codeType(_ encoding: String) -> BiliPlayViewReq.CodeType {
        switch encoding {
        case "AVC": return .code264
        case "HEVC": return .code265
        case "AV1": return .codeav1
        default: return .code265
        }
    }

    static func playView(aid: String, cid: String, epId: String, qn: String, isPGC: Bool, encoding: String,
                         http: HTTPClient, session: Session) async throws -> String {
        let headers = try headers(token: session.token)
        Log.debug("App-Req-Headers: \(headers)")
        let data: Data
        if isPGC {
            if !(encoding.isEmpty || encoding == "HEVC") { Log.warn("The APP API offers bangumi only in HEVC") }
            let body = try payload(epId: Int64(epId) ?? 0, cid: Int64(cid) ?? 0, codec: .code265)
            data = try await http.postGRPC(pgcEndpoint, body: body, headers: headers)
        } else {
            // For UGC the ep_id field carries the aid.
            let body = try payload(epId: Int64(aid) ?? 0, cid: Int64(cid) ?? 0, codec: codeType(encoding))
            data = try await http.postGRPC(ugcEndpoint, body: body, headers: headers)
        }
        let reply = try BiliPlayViewReply(serializedBytes: GRPCFrame.unpack(data))
        Log.debug("PlayViewReply: \(reply.textFormatString())")
        return convertToDashJSON(reply).serialized
    }

    /// Subtitle list via the danmaku view service (works without a web cookie).
    static func dmView(aid: String, cid: String, http: HTTPClient) async throws -> BiliDmViewReply {
        var req = BiliDmViewReq()
        req.pid = Int64(aid) ?? 0
        req.oid = Int64(cid) ?? 0
        req.type = 1
        req.spmid = "main.ugc-video-detail.0.0"
        let body = try GRPCFrame.pack(try req.serializedBytes())
        let data = try await http.postGRPC(dmViewEndpoint, body: body)
        return try BiliDmViewReply(serializedBytes: GRPCFrame.unpack(data))
    }

    private static func payload(epId: Int64, cid: Int64, codec: BiliPlayViewReq.CodeType) throws -> Data {
        var req = BiliPlayViewReq()
        req.epID = epId
        req.cid = cid
        req.qn = 127
        req.fnval = 4048
        req.fourk = true
        req.spmid = "main.ugc-video-detail.0.0"
        req.fromSpmid = "main.my-history.0.0"
        req.preferCodecType = codec
        req.download = 0      // 0 play, 1 flv download, 2 dash download
        req.forceHost = 2     // 0 allow ip, 1 http, 2 https
        Log.debug("PayLoad: \(req.textFormatString())")
        return try GRPCFrame.pack(try req.serializedBytes())
    }

    // MARK: headers

    private static func headers(token: String) throws -> [String: String] {
        [
            "Host": "grpc.biliapi.net",
            "user-agent": "Dalvik/\(dalvikVersion) (Linux; U; Android \(osVersion); \(brand) \(model)) \(appVersion) os/android model/\(brand) mobi_app/android build/\(build) channel/\(channel) innerVer/\(build) osVer/\(osVersion) network/2 grpc-java-cronet/\(cronet)",
            "te": "trailers",
            "x-bili-fawkes-req-bin": try fawkesBin(),
            "x-bili-metadata-bin": try metadataBin(token: token),
            "authorization": "identify_v1 \(token)",
            "x-bili-device-bin": try deviceBin(),
            "x-bili-network-bin": try networkBin(),
            "x-bili-restriction-bin": "",
            "x-bili-locale-bin": try localeBin(),
            "x-bili-exps-bin": "",
            "grpc-encoding": "gzip",
            "grpc-accept-encoding": "identity,gzip",
            "grpc-timeout": "17996161u",
        ]
    }

    private static func base64(_ m: some SwiftProtobuf.Message) throws -> String {
        let bytes: Data = try m.serializedBytes()
        return bytes.base64EncodedString()
    }

    private static func localeBin() throws -> String {
        var ids = BiliLocale.LocaleIds()
        ids.language = language
        ids.region = region
        var l = BiliLocale()
        l.cLocale = ids
        return try base64(l)
    }

    private static func networkBin() throws -> String {
        var n = BiliNetwork()
        n.type = .wifi
        n.oid = networkOid
        return try base64(n)
    }

    private static func deviceBin() throws -> String {
        var d = BiliDevice()
        d.appID = appID
        d.build = build
        d.buvid = buvid
        d.mobiApp = mobiApp
        d.platform = platform
        d.channel = channel
        d.brand = brand
        d.model = model
        d.osver = osVersion
        return try base64(d)
    }

    private static func metadataBin(token: String) throws -> String {
        var m = BiliMetadata()
        m.accessKey = token
        m.mobiApp = mobiApp
        m.build = build
        m.channel = channel
        m.buvid = buvid
        m.platform = platform
        return try base64(m)
    }

    private static func fawkesBin() throws -> String {
        var f = BiliFawkesReq()
        f.appkey = appKey
        f.env = env
        f.sessionID = sessionID
        return try base64(f)
    }

    // MARK: protobuf → web-shaped JSON

    private static func audioJSON(_ item: BiliDashItem, codec: String) -> JSON {
        .object([
            "id": .int(Int64(item.id)),
            "base_url": .string(item.baseURL),
            "backup_url": .array(item.backupURL.map(JSON.string)),
            "bandwidth": .int(Int64(item.bandwidth)),
            "codecs": .string(codec),
        ])
    }

    private static func convertToDashJSON(_ reply: BiliPlayViewReply) -> JSON {
        let info = reply.videoInfo
        let seconds = max(info.timelength / 1000, 1)

        var videos: [JSON] = []
        for item in info.streamList where item.hasDashVideo {
            let dv = item.dashVideo
            videos.append(.object([
                "id": .int(Int64(item.streamInfo.quality)),
                "base_url": .string(dv.baseURL),
                "backup_url": .array(dv.backupURL.map(JSON.string)),
                "bandwidth": .int(Int64(dv.size * 8 / seconds)),
                "codecid": .int(Int64(dv.codecid)),
            ]))
        }

        var audios = info.dashAudio.map { audioJSON($0, codec: "M4A") }
        if info.hasFlac, info.flac.hasAudio { audios.append(audioJSON(info.flac.audio, codec: "FLAC")) }
        if info.hasDolby, info.dolby.hasAudio { audios.append(audioJSON(info.dolby.audio, codec: "E-AC-3")) }

        var clips: [JSON] = []
        if reply.hasBusiness {
            clips = reply.business.clipInfo.map {
                .object(["start": .int(Int64($0.start)), "end": .int(Int64($0.end)), "toastText": .string($0.toastText)])
            }
        }

        var background: [JSON] = []
        var roles: [JSON] = []
        if reply.hasPlayExtInfo, reply.playExtInfo.hasPlayDubbingInfo, reply.playExtInfo.playDubbingInfo.hasBackgroundAudio {
            let dub = reply.playExtInfo.playDubbingInfo
            background = dub.backgroundAudio.audio.map { audioJSON($0, codec: "M4A") }
            for role in dub.roleAudioList {
                for material in role.audioMaterialList {
                    roles.append(.object([
                        "audio_id": .string(material.audioID),
                        "title": .string(material.hasTitle ? material.title : material.audioID),
                        "person_name": .string(material.hasPersonName ? material.personName : (material.hasEdition ? material.edition : "")),
                        "audio": .array(material.audio.map { audioJSON($0, codec: "M4A") }),
                    ]))
                }
            }
        }

        return .object([
            "code": .int(0),
            "message": .string("0"),
            "ttl": .int(1),
            "data": .object([
                "timelength": .int(Int64(info.timelength)),
                "dash": .object(["video": .array(videos), "audio": .array(audios)]),
                "clip_info_list": .array(clips),
            ]),
            "dubbing_info": .object([
                "background_audio": .array(background),
                "role_audio_list": .array(roles),
            ]),
        ])
    }
}
