import Foundation

/// Everything in a download that differs between sites. The page flow (`PageDownloader`) asks its source and never
/// checks which site it is on, so a new site is a new source.
///
/// The defaults below fit sources whose info already carries each page's streams and subtitles (`Page.media`).
protocol MediaSource: Sendable {
    var site: Site { get }
    var http: HTTPClient { get }

    /// Info for a resolved link. The id comes back too: bilibili may find an episode id is really a course.
    func fetchInfo(_ id: MediaID) async throws -> (MediaID, VideoInfo)

    /// Chapters beyond those the page already carries.
    func chapters(for r: PageRequest) async -> [ViewPoint]
    func subtitles(for r: PageRequest) async -> [SubtitleInfo]
    func saveSubtitle(_ subtitle: SubtitleInfo) async throws
    /// `quality` asks again for one FLV quality (bilibili's interactive choice); other sources ignore it.
    func tracks(for r: PageRequest, quality: String?) async throws -> ParsedTracks

    /// Cookie, headers and range rules for the page's media, on top of the run's options.
    func configure(_ config: inout DownloadConfig, for r: PageRequest)
    /// Servers that answer only closed byte ranges (googlevideo) are read this many bytes at a time; nil for others.
    var maxRange: Int64? { get }
    /// Whether a size the site states is exact enough to drive ranged downloads.
    var trustsStatedSizes: Bool { get }
    /// Last changes to the chosen streams' URLs before downloading (bilibili's CDN host rewriting).
    func finalizeURLs(video: inout VideoTrack?, audio: inout AudioTrack?)

    var hasDanmaku: Bool { get }
    func downloadDanmaku(for r: PageRequest, savePath: String, config: DownloadConfig) async throws
    /// A page that is audio by nature (a podcast episode): saved as audio in its codec's container, with its cover.
    func isAudioOnly(_ page: Page) -> Bool
    /// Written into the comment tag.
    func pageURL(_ page: Page) -> String
    /// How `--archive` remembers the page: `<site>:<id>`, so ids of different sites never collide.
    func archiveKey(_ page: Page) -> String
}

/// What a source needs to know about the page being downloaded.
struct PageRequest: Sendable {
    var page: Page
    var id: MediaID
    var api: APIType
}

extension MediaSource {
    func chapters(for r: PageRequest) async -> [ViewPoint] { [] }
    func subtitles(for r: PageRequest) async -> [SubtitleInfo] { r.page.media?.subtitles ?? [] }

    /// A fresh session: a bilibili cookie stays with bilibili.
    func saveSubtitle(_ subtitle: SubtitleInfo) async throws {
        try await Subtitles.save(subtitle, http: http, session: Session())
    }

    func tracks(for r: PageRequest, quality: String?) async throws -> ParsedTracks {
        guard let media = r.page.media else { throw HaulError("No streams found for \"\(r.page.title)\"") }
        return media.tracks
    }

    /// The site's own headers; no bilibili cookie, no forced HTTP.
    func configure(_ config: inout DownloadConfig, for r: PageRequest) {
        config.cookie = ""
        config.forceHTTP = false
        config.headers = r.page.media?.headers ?? [:]
        config.maxRange = maxRange
        if maxRange != nil && config.useAria2c {
            Log.warn("This server answers only byte ranges; not using aria2c")
            config.useAria2c = false
        }
    }

    var maxRange: Int64? { nil }
    var trustsStatedSizes: Bool { true }
    func finalizeURLs(video: inout VideoTrack?, audio: inout AudioTrack?) {}
    var hasDanmaku: Bool { false }
    func downloadDanmaku(for r: PageRequest, savePath: String, config: DownloadConfig) async throws {}
    func isAudioOnly(_ page: Page) -> Bool { false }
    func pageURL(_ page: Page) -> String { page.media?.webpageURL ?? "" }
    func archiveKey(_ page: Page) -> String { "\(site.rawValue):\(page.aid)" }
}

/// What a page is saved as.
public enum Container: String, Sendable {
    case mp4, m4a, mp3

    /// Audio alone, in the container its codec fits: MP3 stays MP3, everything else goes into M4A.
    public static func audio(codec: String) -> Container { codec == "MP3" ? .mp3 : .m4a }
}
