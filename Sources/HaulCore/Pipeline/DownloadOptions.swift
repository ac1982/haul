import Foundation

/// Every knob for one download. Decoded from the config file and filled from the command line,
/// so both speak the same vocabulary.
public struct DownloadOptions: Codable, Sendable {
    public var url = ""
    public var api: APIType = .web
    public var useMP4Box = false
    /// e.g. `hevc,av1,avc,flac,eac3,m4a`
    public var codecPriority: String?
    /// e.g. `4K, 1080P60, 1080p` — the quality labels the stream table shows.
    public var qualityPriority: String?
    /// When both priorities are set, sort by codec first (otherwise quality first).
    public var codecPriorityFirst = false
    public var onlyShowInfo = false
    /// `info` only: print each stream's download URL under its row.
    public var showURLs = false
    public var showAll = false
    public var useAria2c = false
    public var interactive = false
    /// Pick a stream by its index in the (sorted) table instead of the first one.
    public var videoStream: Int?
    public var audioStream: Int?
    public var hideStreams = false
    public var multiThread = true
    public var simpleMux = false
    public var videoOnly = false
    public var audioOnly = false
    public var danmakuOnly = false
    public var coverOnly = false
    public var subtitleOnly = false
    public var debug = false
    public var skipMux = false
    public var skipSubtitle = false
    public var skipCover = false
    public var forceHTTP = true
    public var downloadDanmaku = false
    public var danmakuFormats: [DanmakuFormat] = [.xml, .ass]
    public var skipAISubtitle = true
    public var videoAscending = false
    public var audioAscending = false
    public var allowPCDN = false
    public var forceReplaceHost = true
    public var saveArchive = false
    public var filePattern = ""
    public var multiFilePattern = ""
    /// Page selection: `8`, `1,2`, `3-5`, `ALL`, `LAST`, `3,5,LATEST`.
    public var pages = ""
    /// Audio language code written into the container (`chi`, `jpn`).
    public var language = ""
    public var userAgent = ""
    public var cookie = ""
    public var accessToken = ""
    public var aria2cArgs = ""
    public var workDir = ""
    public var ffmpegPath = ""
    public var mp4boxPath = ""
    public var aria2cPath = ""
    public var ytdlpPath = ""
    public var uposHost = ""
    public var delayPerPage = 0
    public var host = "api.bilibili.com"
    public var epHost = "api.bilibili.com"
    public var tvHost = "api.snm0516.aisee.tv"
    public var area = ""

    public init() {}

    /// Missing keys keep their defaults, so a config file only has to mention what it changes.
    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        func v<T: Decodable>(_ key: CodingKeys, _ current: inout T) throws {
            if let x = try c.decodeIfPresent(T.self, forKey: key) { current = x }
        }
        func o<T: Decodable>(_ key: CodingKeys, _ current: inout T?) throws {
            if let x = try c.decodeIfPresent(T.self, forKey: key) { current = x }
        }
        try v(.url, &url); try v(.api, &api); try v(.useMP4Box, &useMP4Box)
        try o(.codecPriority, &codecPriority); try o(.qualityPriority, &qualityPriority); try v(.codecPriorityFirst, &codecPriorityFirst)
        try v(.onlyShowInfo, &onlyShowInfo); try v(.showURLs, &showURLs); try v(.showAll, &showAll); try v(.useAria2c, &useAria2c)
        try v(.interactive, &interactive); try o(.videoStream, &videoStream); try o(.audioStream, &audioStream); try v(.hideStreams, &hideStreams); try v(.multiThread, &multiThread)
        try v(.simpleMux, &simpleMux); try v(.videoOnly, &videoOnly); try v(.audioOnly, &audioOnly)
        try v(.danmakuOnly, &danmakuOnly); try v(.coverOnly, &coverOnly); try v(.subtitleOnly, &subtitleOnly)
        try v(.debug, &debug); try v(.skipMux, &skipMux); try v(.skipSubtitle, &skipSubtitle); try v(.skipCover, &skipCover)
        try v(.forceHTTP, &forceHTTP); try v(.downloadDanmaku, &downloadDanmaku); try v(.danmakuFormats, &danmakuFormats)
        try v(.skipAISubtitle, &skipAISubtitle); try v(.videoAscending, &videoAscending); try v(.audioAscending, &audioAscending)
        try v(.allowPCDN, &allowPCDN); try v(.forceReplaceHost, &forceReplaceHost); try v(.saveArchive, &saveArchive)
        try v(.filePattern, &filePattern); try v(.multiFilePattern, &multiFilePattern); try v(.pages, &pages)
        try v(.language, &language); try v(.userAgent, &userAgent); try v(.cookie, &cookie); try v(.accessToken, &accessToken)
        try v(.aria2cArgs, &aria2cArgs); try v(.workDir, &workDir); try v(.ffmpegPath, &ffmpegPath)
        try v(.mp4boxPath, &mp4boxPath); try v(.aria2cPath, &aria2cPath); try v(.ytdlpPath, &ytdlpPath); try v(.uposHost, &uposHost)
        try v(.delayPerPage, &delayPerPage); try v(.host, &host); try v(.epHost, &epHost); try v(.tvHost, &tvHost); try v(.area, &area)
    }

    /// Resolve mutually exclusive switches the way a person would expect.
    public mutating func normalize() {
        if interactive { hideStreams = false }
        if audioOnly && videoOnly { audioOnly = false; videoOnly = false }
        if skipSubtitle { subtitleOnly = false }
    }

    public static func load(configFile: URL) throws -> DownloadOptions {
        let data = try Data(contentsOf: configFile)
        return try JSONDecoder().decode(DownloadOptions.self, from: data)
    }
}
