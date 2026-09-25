import Foundation

/// bilibili: info fetchers, playurl APIs, CDN host rewriting, danmaku.
struct BilibiliSource: MediaSource {
    /// A known-good mirror for PCDN and overseas hosts.
    static let backupHost = "upos-sz-mirrorcoso1.bilivideo.com"

    let http: HTTPClient
    let session: Session
    let options: DownloadOptions
    /// The codec the APP API is asked for.
    let preferredCodec: String

    var site: Site { .bilibili }

    /// Credentials and endpoints for the run, and whether the account is logged in. The login check doubles as the
    /// WBI key fetch the web playurl needs; tv, intl and proxied runs skip it.
    static func session(options: DownloadOptions, http: HTTPClient) async -> (Session, Bool?) {
        var session = Session()
        if !options.userAgent.isEmpty { session.userAgent = options.userAgent }
        session.host = options.host
        session.epHost = options.epHost
        session.tvHost = options.tvHost
        session.area = options.area
        session.cookie = options.cookie
        session.token = options.accessToken.replacingOccurrences(of: "access_token=", with: "")
        Auth.loadStored(into: &session, api: options.api)

        guard options.api != .intl && options.api != .tv && options.area.isEmpty else { return (session, nil) }
        Log.debug("Checking the bilibili login")
        guard let nav = try? await Auth.nav(http: http, session: session) else {
            Log.warn("Could not check the bilibili login; some streams may be missing")
            return (session, nil)
        }
        session.wbiKey = nav.wbiKey
        return (session, nav.isLoggedIn)
    }

    func fetchInfo(_ input: MediaID) async throws -> (MediaID, VideoInfo) {
        var id = input
        var info: VideoInfo
        do {
            info = try await InfoFetcher.fetch(id, useIntl: options.api == .intl, http: http, session: session)
        } catch let e as JSONError {
            // An ep/ss that is not a bangumi may be a course.
            guard case .missingKey = e, case .episode(let ep) = id else { throw e }
            Log.warn("No bangumi has this ep/ss id; trying it as a course")
            id = .cheese(epId: ep)
            Log.debug("Resolved to \(id)")
            info = try await InfoFetcher.fetch(id, useIntl: false, http: http, session: session)
        }
        if let bv = info.pages.first?.bvid, !bv.isEmpty, options.api != .intl {
            Log.debug("https://www.bilibili.com/video/\(bv)/")
        }
        return (id, info)
    }

    func chapters(for r: PageRequest) async -> [ViewPoint] {
        Log.debug("Fetching chapters")
        let url = "https://api.bilibili.com/x/player/wbi/v2?cid=\(r.page.cid)&aid=\(r.page.aid)"
        guard let json = try? await http.getJSON(url, session: session) else { return [] }
        return json["data"]["view_points"].array.map {
            ViewPoint(title: $0["content"].stringValue, start: $0["from"].int ?? 0, end: $0["to"].int ?? 0)
        }
    }

    func subtitles(for r: PageRequest) async -> [SubtitleInfo] {
        await Subtitles.list(aid: r.page.aid, cid: r.page.cid, epId: r.page.epid, index: r.page.index, intl: r.api == .intl,
                             http: http, session: session)
    }

    func saveSubtitle(_ subtitle: SubtitleInfo) async throws {
        try await Subtitles.save(subtitle, http: http, session: session)
    }

    func tracks(for r: PageRequest, quality: String?) async throws -> ParsedTracks {
        let request = PlayURLRequest(id: r.id, aid: r.page.aid, cid: r.page.cid, epId: r.page.epid, api: r.api, encoding: preferredCodec)
        return try await PlayURLClient.extractTracks(request, qn: quality ?? "0", http: http, session: session)
    }

    func configure(_ config: inout DownloadConfig, for r: PageRequest) {
        config.cookie = session.cookie
        config.forceHTTP = options.forceHTTP
    }

    var trustsStatedSizes: Bool { false }

    /// PCDN (`ip:port`) and overseas (akamaized) hosts are swapped for a known-good mirror; `--upos-host` overrides everything.
    func finalizeURLs(video: inout VideoTrack?, audio: inout AudioTrack?) {
        var uposHost = options.uposHost
        if options.forceReplaceHost && uposHost.isEmpty { uposHost = Self.backupHost }
        func rewrite(_ url: String, label: String) -> String {
            if uposHost.isEmpty {
                var u = url
                if !options.allowPCDN, u.firstMatch(of: /:\/\/[^\/]*:\d+\//) != nil {
                    Log.debug("\(label) stream is on a PCDN host; using \(Self.backupHost)")
                    u = u.replacing(/:\/\/[^\/]*:\d+\//, with: "://\(Self.backupHost)/")
                }
                if !session.area.isEmpty, u.contains("akamaized.net") {
                    Log.debug("\(label) stream is on an overseas host; using \(Self.backupHost)")
                    u = u.replacing(/:\/\/[^\/]*akamaized\.net\//, with: "://\(Self.backupHost)/")
                }
                return u
            }
            Log.debug("\(label) stream host set to \(uposHost)")
            return url.replacing(/:\/\/[^\/]+\//, with: "://\(uposHost)/")
        }
        if var v = video { v.baseURL = rewrite(v.baseURL, label: "Video"); video = v }
        if var a = audio { a.baseURL = rewrite(a.baseURL, label: "Audio"); audio = a }
    }

    var hasDanmaku: Bool { true }

    /// Comment XML next to the output, converted to ASS unless only XML was asked for.
    func downloadDanmaku(for r: PageRequest, savePath: String, config: DownloadConfig) async throws {
        let fm = FileManager.default
        let xmlPath = FilePattern.changeExtension(savePath, to: "xml")
        let assPath = FilePattern.changeExtension(savePath, to: "ass")
        try await Downloader.downloadFile("https://comment.bilibili.com/\(r.page.cid).xml", to: xmlPath, config: config, http: http)
        guard let items = Danmaku.parseXML(at: xmlPath) else {
            Log.warn("Could not parse the danmaku XML")
            try? fm.removeItem(atPath: xmlPath)
            return
        }
        if items.isEmpty {
            Log.status("Danmaku  none")
            try? fm.removeItem(atPath: xmlPath)
            return
        }
        var produced: [String] = []
        if options.danmakuFormats.contains(.ass) {
            try Danmaku.writeASS(items, to: assPath)
            produced.append(".ass")
        }
        if options.danmakuFormats.contains(.xml) {
            produced.append(".xml")
        } else {
            try? fm.removeItem(atPath: xmlPath)
        }
        Log.status("Danmaku  \(items.count) comments  →  \(produced.joined(separator: ", "))")
    }

    func pageURL(_ page: Page) -> String { "https://www.bilibili.com/video/\(page.bvid)/" }
}
