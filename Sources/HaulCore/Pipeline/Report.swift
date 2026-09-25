import Foundation
import Synchronization
#if canImport(Glibc)
import Glibc
#endif

/// What a run found and did, for `--json`: the item, every page, the streams of the pages it looked at, and the files
/// it wrote. Filled in while the run goes; `document` is what gets printed.
public final class Report: Sendable {
    public struct Stream: Codable, Sendable, Equatable {
        /// The index `--video-stream` / `--audio-stream` take, in the order haul sorts the streams.
        public var index: Int
        public var quality: String?
        public var resolution: String?
        public var codec: String
        public var fps: Double?
        public var bitrateKbps: Int64?
        /// Stated by the site, or estimated from bitrate × duration; nil when neither is known.
        public var sizeBytes: Int64?
        public var url: String?
    }

    public struct PageEntry: Codable, Sendable {
        public enum Status: String, Codable, Sendable {
            /// Streams listed (`info`), nothing downloaded.
            case listed
            case downloaded
            /// The output file already existed, or `--archive` had the page.
            case skipped
        }

        /// The number `-p` takes.
        public var index: Int
        public var title: String
        public var id: String
        public var durationSeconds: Int?
        /// ISO 8601, when the site dates its pages (episodes, list items).
        public var published: String?
        /// Whether the run worked on this page (`-p`, or the page a link points at).
        public var selected: Bool
        public var status: Status?
        public var reason: String?
        /// The main output file, absolute.
        public var file: String?
        public var sizeBytes: Int64?
        /// Subtitles, danmaku or a cover saved next to it, absolute.
        public var extraFiles: [String]?
        public var video: [Stream]?
        public var audio: [Stream]?
        /// The video streams are whole files with the audio inside (X), so there are no audio streams.
        public var videoHasAudio: Bool?
        public var selectedVideo: Int?
        public var selectedAudio: Int?
        public var subtitles: [String]?
    }

    public struct ErrorDetail: Codable, Sendable {
        public var kind: HaulError.Kind
        public var message: String
        public var exitCode: Int32

        init(_ error: Error) {
            kind = error.haulKind
            message = error.readableMessage
            exitCode = kind.exitCode
        }
    }

    public struct Document: Codable, Sendable {
        public var ok = true
        /// Why the run stopped; the pages keep what was done before it.
        public var error: ErrorDetail?
        public var command: String
        public var input: String
        public var site: Site
        public var title: String
        public var uploader: String?
        /// ISO 8601.
        public var published: String?
        public var description: String?
        /// bilibili only: whether the stored login was valid (logged out means lower qualities); nil when not checked.
        public var loggedIn: Bool?
        public var pageCount: Int
        /// The output files of the pages the run took, absolute, whether written now or found already there (skipped):
        /// each page's file and extra files. Empty for `info`.
        public var files: [String] = []
        public var pages: [PageEntry]
    }

    private struct State {
        var document: Document?
    }

    private let state = Mutex(State())
    let command: String

    public init(command: String) { self.command = command }

    /// The document, with `files` gathered from the pages.
    public var document: Document? {
        guard var doc = state.withLock({ $0.document }) else { return nil }
        doc.files = doc.pages.flatMap { p in
            (p.status == .downloaded || p.status == .skipped ? [p.file].compactMap { $0 } : []) + (p.extraFiles ?? [])
        }
        return doc
    }

    func start(_ info: VideoInfo, input: String, selected: Set<Int>, loggedIn: Bool?) {
        let owner = info.pages.first { !($0.ownerName ?? "").isEmpty }?.ownerName
        let doc = Document(
            command: command, input: input, site: info.site, title: info.title, uploader: owner,
            published: info.pubTime > 0 ? Self.iso(info.pubTime) : nil,
            description: info.desc.isEmpty ? nil : info.desc, loggedIn: loggedIn, pageCount: info.pages.count,
            pages: info.pages.map {
                PageEntry(index: $0.index, title: $0.title, id: Self.pageID($0, site: info.site),
                          durationSeconds: $0.duration > 0 ? $0.duration : nil, published: $0.pubTime > 0 ? Self.iso($0.pubTime) : nil,
                          selected: selected.contains($0.index))
            })
        state.withLock { $0.document = doc }
    }

    func update(_ page: Page, _ change: (inout PageEntry) -> Void) {
        state.withLock { st in
            guard let i = st.document?.pages.firstIndex(where: { $0.index == page.index }) else { return }
            change(&st.document!.pages[i])
        }
    }

    func streams(_ page: Page, _ p: ParsedTracks, video: Int?, audio: Int?, includeURLs: Bool) {
        let v = p.video.enumerated().map { i, t in
            Stream(index: i, quality: t.quality, resolution: t.resolution, codec: t.codec, fps: t.fps.flatMap(Double.init),
                   bitrateKbps: t.bandwidth > 0 ? t.bandwidth : nil, sizeBytes: Self.size(t.estimatedSize(pageDuration: page.duration)),
                   url: includeURLs ? t.baseURL : nil)
        }
        let a = p.audio.enumerated().map { i, t in
            Stream(index: i, quality: nil, resolution: nil, codec: t.codec, fps: nil,
                   bitrateKbps: t.bandwidth > 0 ? t.bandwidth : nil, sizeBytes: Self.size(t.estimatedSize(pageDuration: page.duration)),
                   url: includeURLs ? t.baseURL : nil)
        }
        update(page) {
            $0.video = v
            $0.audio = a
            $0.videoHasAudio = p.videoHasAudio ? true : nil
            $0.selectedVideo = v.isEmpty ? nil : video
            $0.selectedAudio = a.isEmpty ? nil : audio
        }
    }

    func subtitles(_ page: Page, _ subs: [SubtitleInfo]) {
        update(page) { $0.subtitles = subs.map(\.language) }
    }

    func extraFile(_ page: Page, _ path: String) {
        update(page) { $0.extraFiles = ($0.extraFiles ?? []) + [Self.absolute(path)] }
    }

    func finished(_ page: Page, status: PageEntry.Status, file: String?, reason: String? = nil) {
        update(page) {
            $0.status = status
            $0.reason = reason
            if let file {
                $0.file = Self.absolute(file)
                let size = Downloader.fileSize(file)
                $0.sizeBytes = size > 0 ? size : nil
            }
        }
    }

    /// The document as pretty JSON.
    public func json() -> String? {
        guard let doc = document else { return nil }
        return Self.encode(doc)
    }

    public static func encode<T: Encodable>(_ value: T) -> String {
        let encoder = JSONEncoder()
        encoder.outputFormatting = [.prettyPrinted, .sortedKeys, .withoutEscapingSlashes]
        return (try? String(decoding: encoder.encode(value), as: UTF8.self)) ?? "{}"
    }

    /// The document of a run that failed: what it did so far with `ok: false` and the error, or just the error when
    /// it failed before finding the item.
    public func failureJSON(_ error: Error, input: String?) -> String {
        if var doc = document {
            doc.ok = false
            doc.error = ErrorDetail(error)
            return Self.encode(doc)
        }
        struct Failure: Encodable {
            var ok = false
            var command: String
            var input: String?
            var error: ErrorDetail
        }
        return Self.encode(Failure(command: command, input: input, error: ErrorDetail(error)))
    }

    // MARK: helpers

    static func pageID(_ page: Page, site: Site) -> String {
        guard site == .bilibili else { return page.aid }
        let bv = page.bvid
        return page.epid.isEmpty || page.epid == "0" ? (bv.isEmpty ? "av\(page.aid)" : bv) : "ep\(page.epid)"
    }

    static func size(_ bytes: Double) -> Int64? { bytes > 0 ? Int64(bytes) : nil }

    static func iso(_ seconds: Int64) -> String {
        Date(timeIntervalSince1970: TimeInterval(seconds)).formatted(.iso8601)
    }

    /// Absolute and with symlinks resolved (`/tmp` → `/private/tmp`), so it compares equal to what `realpath` gives.
    static func absolute(_ path: String) -> String {
        let p = ((path.hasPrefix("/") ? path : (FileManager.default.currentDirectoryPath as NSString).appendingPathComponent(path))
            as NSString).standardizingPath
        func real(_ p: String) -> String? {
            guard let r = realpath(p, nil) else { return nil }
            defer { free(r) }
            return String(cString: r)
        }
        if let r = real(p) { return r }
        if let dir = real((p as NSString).deletingLastPathComponent) { return (dir as NSString).appendingPathComponent((p as NSString).lastPathComponent) }
        return p
    }
}
