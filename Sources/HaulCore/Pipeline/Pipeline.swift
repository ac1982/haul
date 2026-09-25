import Foundation

/// One download run, from input string to muxed files. Construct, then `run()`.
public final class DownloadPipeline: Sendable {
    public let options: DownloadOptions
    public let site: Site
    public let session: Session
    let http: HTTPClient
    let codecPriority: [String: Int]
    let qualityPriority: [String: Int]
    /// The codec bilibili's APP API is asked for.
    let preferredCodec: String
    let ffmpeg: String
    let mp4box: String?
    let aria2c: String?
    /// Resolved only for YouTube / X links.
    let ytdlp: String?
    /// nil when the login state was not checked (every site but bilibili, and bilibili's tv / intl / proxy modes).
    public let isLoggedIn: Bool?
    /// What the run found and did, for `--json`.
    public let report: Report

    public init(options rawOptions: DownloadOptions, report: Report = Report(command: "download"), http: HTTPClient = .shared) async throws {
        var options = rawOptions
        options.normalize()
        Log.debugEnabled = options.debug
        guard let site = Site.match(options.url)?.site else { throw HaulError.unsupported(options.url) }
        self.site = site

        // External tools; `info` downloads nothing, so it needs only what reads the link.
        let muxes = !options.onlyShowInfo && !options.skipMux
        ffmpeg = try Self.tool(options.ffmpegPath, name: "ffmpeg", required: muxes && !options.useMP4Box) ?? "ffmpeg"
        mp4box = try Self.tool(options.mp4boxPath, name: "mp4box", required: muxes && options.useMP4Box)
            ?? Shell.findExecutable("MP4Box")
        aria2c = try Self.tool(options.aria2cPath, name: "aria2c", required: options.useAria2c && !options.onlyShowInfo)
        ytdlp = site.usesYtDlp ? try Self.tool(options.ytdlpPath, name: "yt-dlp", required: false) : nil
        if site.usesYtDlp {
            if ytdlp == nil { throw HaulError.dependency("\(site.name) links need yt-dlp: brew install yt-dlp deno") }
            // yt-dlp solves YouTube's player challenges in deno; without it most formats are missing.
            if site == .youtube && Shell.findExecutable("deno") == nil {
                Log.warn("deno not found; yt-dlp may miss most YouTube formats: brew install deno")
            }
        }

        if !options.workDir.isEmpty {
            let dir = Self.expand(options.workDir)
            try FileManager.default.createDirectory(atPath: dir, withIntermediateDirectories: true)
            FileManager.default.changeCurrentDirectoryPath(dir)
            Log.debug("Working directory: \(dir)")
        }

        let codec = Self.parseCodecPriority(options.codecPriority)
        codecPriority = codec.0
        preferredCodec = codec.1
        qualityPriority = Self.parseQualityPriority(options.qualityPriority)

        if site == .bilibili {
            (session, isLoggedIn) = await BilibiliSource.session(options: options, http: http)
        } else {
            session = Session()
            isLoggedIn = nil
        }

        self.options = options
        self.http = http
        self.report = report
        Log.debug("Options: \((try? String(decoding: JSONEncoder().encode(options), as: UTF8.self)) ?? "")")
    }

    // MARK: setup helpers

    private static func tool(_ explicit: String, name: String, required: Bool) throws -> String? {
        if !explicit.isEmpty {
            if let p = Shell.resolve(explicit) { return p }
            Log.warn("\(name) not found at \(explicit); looking on PATH")
        }
        if let found = Shell.findExecutable(name) { return found }
        if required { throw HaulError.dependency("\(name) not found on PATH: brew install \(name == "mp4box" ? "gpac" : name)") }
        return nil
    }

    private static func expand(_ path: String) -> String {
        var p = (path as NSString).expandingTildeInPath
        for (k, v) in ProcessInfo.processInfo.environment {
            p = p.replacingOccurrences(of: "${\(k)}", with: v).replacingOccurrences(of: "$\(k)", with: v)
        }
        return (p as NSString).standardizingPath
    }

    static func parseCodecPriority(_ text: String?) -> ([String: Int], String) {
        guard let text else { return ([:], "") }
        let items = text.uppercased().replacingOccurrences(of: "，", with: ",").replacingOccurrences(of: "-", with: "")
            .split(separator: ",").map { $0.trimmingCharacters(in: .whitespaces) }.filter { !$0.isEmpty }
        var map: [String: Int] = [:]
        for (i, c) in items.enumerated() where map[c] == nil { map[c] = i }
        return (map, items.first ?? "")
    }

    static func parseQualityPriority(_ text: String?) -> [String: Int] {
        guard let text else { return [:] }
        let items = text.replacingOccurrences(of: "，", with: ",").split(separator: ",")
            .map { $0.uppercased().trimmingCharacters(in: .whitespaces) }.filter { !$0.isEmpty }
        var map: [String: Int] = [:]
        for (i, q) in items.enumerated() where map[q] == nil { map[q] = i }
        return map
    }

    // MARK: run

    public func run() async throws {
        let (id, info) = try await fetchInfo()
        try await downloadPages(info, id: id)
    }

    public func fetchInfo() async throws -> (MediaID, VideoInfo) {
        Log.debug("Resolving the link")
        let resolved = try await IDResolver.resolve(options.url, http: http, session: session)
        Log.debug("Resolved to \(resolved)")
        let (id, info) = try await source(for: resolved.site).fetchInfo(resolved)

        let style = Log.style
        Log.lines(UI.header(info, loggedIn: isLoggedIn, style: style))
        Log.lines(UI.pageList(info.pages, showAll: options.showAll, style: style))
        Log.plain()
        return (id, info)
    }

    /// Where a site's info and media come from.
    func source(for site: Site) -> any MediaSource {
        switch site {
        case .bilibili: return BilibiliSource(http: http, session: session, options: options, preferredCodec: preferredCodec)
        case .youtube, .x: return YtDlpSource(site: site, ytdlp: ytdlp, http: http)
        case .xiaoyuzhou, .applePodcasts: return PodcastSource(site: site, http: http)
        }
    }

    public func downloadPages(_ info: VideoInfo, id: MediaID) async throws {
        var api = options.api
        if info.isInteractive && api == .tv {
            Log.warn("Interactive videos are not available through the TV API; using the web API")
            api = .web
        }
        let selection = try PageSelection.parse(options.pages, info: info, input: options.url)
        let totalPages = info.pages.count
        let unit = info.site.unit
        // `info` on a list without -p lists the pages only: reading every page's streams can take minutes.
        if options.onlyShowInfo && selection == nil && totalPages > 1 && options.pages.isEmpty {
            report.start(info, input: options.url, selected: [], loggedIn: isLoggedIn)
            Log.status("\(totalPages) \(unit)s; -p <n> lists the streams of one, -p ALL of all")
            return
        }
        let pages = selection.map { s in info.pages.filter { s.contains($0.index) } } ?? info.pages
        report.start(info, input: options.url, selected: Set(pages.map(\.index)), loggedIn: isLoggedIn)
        if let selection, pages.isEmpty {
            throw HaulError.input("-p \(options.pages) matches none of the \(totalPages) \(unit)s (\(selection.map(String.init).joined(separator: ",")))")
        }
        if totalPages > 1 {
            let verb = options.onlyShowInfo ? "Listing" : "Downloading"
            Log.status(selection == nil ? "\(verb) all \(totalPages) \(unit)s"
                : "\(verb) \(pages.count) of \(totalPages) \(unit)s: \(pages.map { String($0.index) }.joined(separator: ", "))")
        }
        let started = Date()
        let mediaSource = source(for: info.site)

        var pattern = options.filePattern.isEmpty ? FilePattern.singleDefault : options.filePattern
        // Multi-page, or an ongoing series, gets the folder layout.
        if totalPages > 1 || (info.isBangumi && !info.isBangumiEnd) {
            pattern = options.multiFilePattern.isEmpty ? FilePattern.multiDefault : options.multiFilePattern
        }

        for (n, page) in pages.enumerated() {
            if pages.count > 1 && options.delayPerPage > 0 && n > 0 {
                Log.status("Waiting \(options.delayPerPage)s")
                try await Task.sleep(for: .seconds(options.delayPerPage))
            }
            if pages.count > 1 {
                if n > 0 { Log.plain() }
                Log.plain(UI.pageHeader(index: n + 1, count: pages.count, title: page.title, style: Log.style))
            }
            let key = mediaSource.archiveKey(page)
            if options.saveArchive && !options.onlyShowInfo && Archive.contains(key) {
                Log.status("Already downloaded (\(key)), skipping")
                report.finished(page, status: .skipped, file: nil, reason: "archive")
                continue
            }
            try await PageDownloader(pipeline: self, source: mediaSource, info: info, id: id, api: api, selectedPages: pages, pattern: pattern).run(page)
            if options.saveArchive && !options.onlyShowInfo { Archive.add(key) }
        }
        if pages.count > 1 && !options.onlyShowInfo {
            Log.plain()
            Log.success("All done  \(pages.count) \(unit)s  " + Log.style.dim("·  \(UI.eta(Date().timeIntervalSince(started)))"))
        }
    }

    // MARK: track ordering

    func sortVideo(_ tracks: [VideoTrack]) -> [VideoTrack] {
        let byCodecFirst = !qualityPriority.isEmpty && !codecPriority.isEmpty && options.codecPriorityFirst
        return tracks.sorted { a, b in
            let ka = key(a, codecFirst: byCodecFirst), kb = key(b, codecFirst: byCodecFirst)
            return ka.lexicographicallyPrecedes(kb)
        }
    }

    private func key(_ v: VideoTrack, codecFirst: Bool) -> [Int64] {
        let q = Int64(qualityPriority[v.quality.uppercased()] ?? 100)
        let c = Int64(codecPriority[v.codec] ?? 100)
        // Ascending turns the whole order around: lowest quality first, then the smallest stream.
        let id = options.videoAscending ? Int64(v.id) ?? 0 : -(Int64(v.id) ?? 0)
        let bw = options.videoAscending ? v.bandwidth : -v.bandwidth
        return codecFirst ? [c, q, id, bw] : [q, c, id, bw]
    }

    func sortAudio(_ tracks: [AudioTrack]) -> [AudioTrack] {
        tracks.sorted { a, b in
            let ka = [Int64(codecPriority[a.shortCodec] ?? 100), options.audioAscending ? a.bandwidth : -a.bandwidth]
            let kb = [Int64(codecPriority[b.shortCodec] ?? 100), options.audioAscending ? b.bandwidth : -b.bandwidth]
            return ka.lexicographicallyPrecedes(kb)
        }
    }
}

/// `-p 8`, `-p 1,2`, `-p 3-5`, `-p ALL`, `-p LAST`, `-p 3,5,LATEST`. `nil` means everything.
public enum PageSelection {
    public static func parse(_ spec: String, info: VideoInfo, input: String) throws -> [Int]? {
        let text = spec.uppercased().trimmingCharacters(in: .whitespaces).trimmingCharacters(in: CharacterSet(charactersIn: ","))
        if text.isEmpty {
            // No explicit choice: an episode link or `?p=` picks its own page.
            if let idx = info.index, let n = Int(idx) {
                Log.status("The link points at episode \(n) (-p ALL for all)")
                return [n]
            }
            if let n = Int(Format.queryValue("p", in: input)) {
                Log.status("The link points at page \(n) (-p ALL for all)")
                return [n]
            }
            return nil
        }
        if text == "ALL" { return nil }
        var t = text
        for k in ["LAST", "NEW", "LATEST"] { t = t.replacingOccurrences(of: k, with: String(info.pages.count)) }
        var out: [Int] = []
        let invalid = HaulError.input("Invalid page selection -p \(spec): use 8, 1,2, 3-5, ALL, LAST or LATEST")
        for token in t.split(separator: ",") {
            let part = token.trimmingCharacters(in: .whitespaces)
            if part.contains("-") {
                let ends = part.split(separator: "-", maxSplits: 1).map { Int($0.trimmingCharacters(in: .whitespaces)) }
                guard ends.count == 2, let a = ends[0], let b = ends[1], a <= b else { throw invalid }
                out += Array(a...b)
            } else if let n = Int(part) {
                out.append(n)
            } else {
                throw invalid
            }
        }
        return out
    }
}

/// The "already downloaded" list used by `--archive`.
enum Archive {
    static func contains(_ aid: String) -> Bool {
        guard let text = Storage.read(Storage.archiveFile) else { return false }
        return text.split(separator: "|").contains { $0 == aid }
    }

    static func add(_ aid: String) {
        try? Storage.append(Storage.archiveFile, "\(aid)|")
    }
}
