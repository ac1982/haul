import ArgumentParser
import HaulCore
import Foundation

extension APIType: ExpressibleByArgument {}
extension DanmakuFormat: ExpressibleByArgument {}

struct GeneralFlags: ParsableArguments {
    @Flag(name: .long, help: ArgumentHelp(
        "Print one JSON document on stdout; logs go to stderr.",
        discussion: "The item, its pages, their streams and, for a download, the output files (top-level `files`)."))
    var json = false

    @Option(name: .long, help: ArgumentHelp("Config file.", discussion: "Default: \(Storage.url(Storage.configFile).path)", valueName: "file"))
    var config: String?

    @Flag(name: .long, help: "Debug log: timestamps, requests, tool command lines.")
    var debug = false
}

struct StreamFlags: ParsableArguments {
    @Option(name: [.customShort("q"), .long], help: ArgumentHelp(
        "Quality priority, comma separated, as the stream table labels them.",
        discussion: "e.g. \"1080p,720p\" (YouTube), \"4K,1080P+,1080P\" (bilibili). Unlisted qualities come after, best first.",
        valueName: "list"))
    var quality: String?

    @Option(name: [.customShort("c"), .long], help: ArgumentHelp(
        "Codec priority, comma separated.", discussion: "Video: av1 vp9 hevc avc. Audio: m4a opus flac eac3 mp3.", valueName: "list"))
    var codec: String?

    @Option(name: .customLong("video-stream"), help: ArgumentHelp(
        "Take this video stream: its index in `haul info`.",
        discussion: "Indexes follow the order -q, -c and --*-ascending give: pass the same ones to info and download.", valueName: "n"))
    var videoStream: Int?

    @Option(name: .customLong("audio-stream"), help: ArgumentHelp("Take this audio stream: its index in `haul info`.", valueName: "n"))
    var audioStream: Int?

    @Flag(name: [.customShort("i"), .long], help: "Choose the streams with the arrow keys (needs a terminal).")
    var interactive = false

    @Flag(name: .long, help: "Prefer the lowest quality and the smallest video stream.")
    var videoAscending = false

    @Flag(name: .long, help: "Prefer the smallest audio stream.")
    var audioAscending = false
}

struct PageFlags: ParsableArguments {
    @Option(name: [.customShort("p"), .long], help: ArgumentHelp(
        "Pages / episodes / videos of a post to take: 8, 1,2, 3-5, ALL, LAST.",
        discussion: "1-based, as `info` numbers them; LAST (or LATEST) is the last page, the newest episode of a show. "
            + "Default: the page the link points at, else all (`info`: a list of the pages only).", valueName: "spec"))
    var pages: String?

    @Flag(name: .long, help: "List every page and every stream instead of a summary.")
    var showAll = false

    @Flag(name: .long, help: "Do not print the stream table.")
    var hideStreams = false
}

struct ContentFlags: ParsableArguments {
    @Flag(name: .long, help: ArgumentHelp("Audio only, as .m4a.",
                                          discussion: "The best audio stream, which may be Opus; -c m4a gets AAC. Podcasts keep .mp3 / .m4a."))
    var audioOnly = false
    @Flag(name: .long, help: "Video only, no audio.")
    var videoOnly = false
    @Flag(name: .long, help: "Only the subtitles, as .srt next to the output.")
    var subtitleOnly = false
    @Flag(name: .long, help: "Only the cover image.")
    var coverOnly = false
    @Flag(name: .long, help: "Do not embed subtitles.")
    var skipSubtitle = false
    @Option(name: .customLong("sub-lang"), help: ArgumentHelp(
        "Only these subtitle languages, comma separated, e.g. en,zh.", discussion: "en also matches en-US. `haul info` lists them.",
        valueName: "list"))
    var subLang: String?
    @Flag(name: .long, inversion: .prefixedNo, help: "Skip auto-generated (AI) subtitles. Default: skip.")
    var skipAiSubtitle: Bool?
    @Flag(name: .long, help: "Do not embed the cover.")
    var skipCover = false
    @Flag(name: .long, help: "Keep the downloaded tracks as they are, without muxing.")
    var skipMux = false
}

struct OutputFlags: ParsableArguments {
    @Option(name: [.customShort("o"), .long], help: ArgumentHelp(
        "File-name template for an item with one page.",
        discussion: "Default: \(FilePattern.singleDefault). Variables: `haul templates`. No extension: haul adds it.", valueName: "template"))
    var output: String?
    @Option(name: .long, help: ArgumentHelp("File-name template for an item with several pages, even when -p takes one.",
                                            discussion: "Default: \(FilePattern.multiDefault)", valueName: "template"))
    var multiOutput: String?
    @Option(name: [.customShort("w"), .long], help: ArgumentHelp("Directory to download into. Default: the current directory.", valueName: "dir"))
    var workDir: String?
    @Option(name: .long, help: ArgumentHelp("Audio language code written into the file, e.g. eng, jpn, chi.", valueName: "code"))
    var lang: String?
    @Flag(name: .long, help: "Do not write description, author and date tags.")
    var simpleMux = false
    @Flag(name: .long, help: "Remember downloaded pages in the archive and skip them next time.")
    var archive = false
    @Option(name: .long, help: ArgumentHelp("Seconds to wait between pages.", valueName: "seconds"))
    var delay: Int?
}

struct BilibiliFlags: ParsableArguments {
    @Option(name: .long, help: "bilibili API: web, tv, app or intl (bilibili.tv). Default: web.")
    var api: APIType?
    @Flag(name: .long, help: "Also save the danmaku (bullet comments).")
    var danmaku = false
    @Flag(name: .long, help: "Only the danmaku.")
    var danmakuOnly = false
    @Option(name: .long, parsing: .upToNextOption, help: "Danmaku formats: xml ass. Default: both.")
    var danmakuFormat: [DanmakuFormat] = []
    @Option(name: .long, help: ArgumentHelp("Web cookie (SESSDATA=…) instead of the stored login.", valueName: "cookie"))
    var cookie: String?
    @Option(name: .long, help: ArgumentHelp("TV / APP access token instead of the stored login.", valueName: "token"))
    var token: String?
    @Option(name: .long, help: ArgumentHelp("User-Agent for bilibili requests. Default: a random desktop browser.", visibility: .hidden))
    var userAgent: String?
    @Option(name: .long, help: ArgumentHelp("Download from this upos CDN host.", visibility: .hidden))
    var uposHost: String?
    @Flag(name: .long, inversion: .prefixedNo, help: ArgumentHelp("Rewrite the CDN host to a known-good mirror. Default: on.", visibility: .hidden))
    var replaceHost: Bool?
    @Flag(name: .long, help: ArgumentHelp("Keep PCDN hosts.", visibility: .hidden))
    var allowPcdn = false
    @Flag(name: .long, inversion: .prefixedNo, help: ArgumentHelp("Fetch media over HTTP instead of HTTPS. Default: on.", visibility: .hidden))
    var forceHttp: Bool?
    @Option(name: .long, help: ArgumentHelp("BiliPlus-style API proxy host (needs a token).", visibility: .hidden))
    var host: String?
    @Option(name: .long, help: ArgumentHelp("Proxy host for /pgc/view/web/season.", visibility: .hidden))
    var epHost: String?
    @Option(name: .long, help: ArgumentHelp("TV API host.", visibility: .hidden))
    var tvHost: String?
    @Option(name: .long, help: ArgumentHelp("Proxy area: hk, tw or th.", visibility: .hidden))
    var area: String?
}

struct ToolFlags: ParsableArguments {
    @Option(name: .long, help: ArgumentHelp("ffmpeg to use. Default: from PATH.", valueName: "path"))
    var ffmpeg: String?
    @Option(name: .customLong("yt-dlp"), help: ArgumentHelp("yt-dlp to use (YouTube, X). Default: from PATH.", valueName: "path"))
    var ytdlp: String?
    @Flag(name: .long, help: "Mux with MP4Box instead of ffmpeg.")
    var useMp4box = false
    @Option(name: .long, help: ArgumentHelp("MP4Box to use.", valueName: "path"))
    var mp4box: String?
    @Flag(name: .long, help: "Download with aria2c.")
    var useAria2c = false
    @Option(name: .long, help: ArgumentHelp("aria2c to use.", valueName: "path"))
    var aria2c: String?
    @Option(name: .long, help: ArgumentHelp("Extra aria2c arguments (on top of -x16 -s16 -j16 -k5M).", valueName: "args"))
    var aria2cArgs: String?
    @Flag(name: .long, inversion: .prefixedNo, help: "Download over several connections. Default: on.")
    var multiThread: Bool?
}

/// Every download switch, grouped as `--help` shows them. Only the ones given on the command line override the config file.
struct DownloadFlags: ParsableArguments {
    @OptionGroup(title: "General") var general: GeneralFlags
    @OptionGroup(title: "Streams") var streams: StreamFlags
    @OptionGroup(title: "Pages") var pageFlags: PageFlags
    @OptionGroup(title: "Content") var content: ContentFlags
    @OptionGroup(title: "Output") var output: OutputFlags
    @OptionGroup(title: "bilibili") var bilibili: BilibiliFlags
    @OptionGroup(title: "Tools") var tools: ToolFlags

    /// Config file (if any) first, then whatever was said on the command line.
    func resolve(url: String?) throws -> DownloadOptions {
        var o: DownloadOptions
        let path = general.config ?? Storage.url(Storage.configFile).path
        if FileManager.default.fileExists(atPath: path) {
            Log.status("Config  \(Terminal.prettyPath(path))")
            do { o = try DownloadOptions.load(configFile: URL(fileURLWithPath: path)) }
            catch { throw HaulError.input("Cannot read the config file \(path): \(error.readableMessage)") }
        } else {
            if general.config != nil { throw HaulError.input("Config file not found: \(path)") }
            o = DownloadOptions()
        }
        if let url { o.url = url }
        if general.debug { o.debug = true }

        let s = streams
        if let v = s.quality { o.qualityPriority = v }
        if let v = s.codec { o.codecPriority = v }
        o.codecPriorityFirst = Self.appearsFirst(["-c", "--codec"], before: ["-q", "--quality"])
        if let v = s.videoStream { o.videoStream = v }
        if let v = s.audioStream { o.audioStream = v }
        if s.interactive { o.interactive = true }
        if s.videoAscending { o.videoAscending = true }
        if s.audioAscending { o.audioAscending = true }

        let p = pageFlags
        if let v = p.pages { o.pages = v }
        if p.showAll { o.showAll = true }
        if p.hideStreams { o.hideStreams = true }

        let c = content
        if c.audioOnly { o.audioOnly = true }
        if c.videoOnly { o.videoOnly = true }
        if c.subtitleOnly { o.subtitleOnly = true }
        if c.coverOnly { o.coverOnly = true }
        if c.skipSubtitle { o.skipSubtitle = true }
        if let v = c.subLang { o.subtitleLanguages = v }
        if let v = c.skipAiSubtitle { o.skipAISubtitle = v }
        if c.skipCover { o.skipCover = true }
        if c.skipMux { o.skipMux = true }

        let out = output
        if let v = out.output { o.filePattern = v }
        if let v = out.multiOutput { o.multiFilePattern = v }
        if let v = out.workDir { o.workDir = v }
        if let v = out.lang { o.language = v }
        if out.simpleMux { o.simpleMux = true }
        if out.archive { o.saveArchive = true }
        if let v = out.delay { o.delayPerPage = v }

        let b = bilibili
        if let v = b.api { o.api = v }
        if b.danmaku { o.downloadDanmaku = true }
        if b.danmakuOnly { o.danmakuOnly = true }
        if !b.danmakuFormat.isEmpty { o.danmakuFormats = b.danmakuFormat }
        if let v = b.cookie { o.cookie = v }
        if let v = b.token { o.accessToken = v }
        if let v = b.userAgent { o.userAgent = v }
        if let v = b.uposHost { o.uposHost = v }
        if let v = b.replaceHost { o.forceReplaceHost = v }
        if b.allowPcdn { o.allowPCDN = true }
        if let v = b.forceHttp { o.forceHTTP = v }
        if let v = b.host { o.host = v }
        if let v = b.epHost { o.epHost = v }
        if let v = b.tvHost { o.tvHost = v }
        if let v = b.area { o.area = v }

        let t = tools
        if let v = t.ffmpeg { o.ffmpegPath = v }
        if let v = t.ytdlp { o.ytdlpPath = v }
        if t.useMp4box { o.useMP4Box = true }
        if let v = t.mp4box { o.mp4boxPath = v }
        if t.useAria2c { o.useAria2c = true }
        if let v = t.aria2c { o.aria2cPath = v }
        if let v = t.aria2cArgs { o.aria2cArgs = v }
        if let v = t.multiThread { o.multiThread = v }
        return o
    }

    private static func appearsFirst(_ a: [String], before b: [String]) -> Bool {
        let args = CommandLine.arguments
        guard let ia = args.firstIndex(where: { a.contains($0) }), let ib = args.firstIndex(where: { b.contains($0) }) else { return false }
        return ia < ib
    }
}
