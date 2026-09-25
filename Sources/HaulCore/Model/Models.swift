import Foundation

/// One downloadable unit: a page of a video, an episode of a season, an item of a list.
public struct Page: Sendable, Hashable {
    public var index: Int
    public var aid: String
    public var cid: String
    public var epid: String
    public var title: String
    public var duration: Int
    public var resolution: String
    public var pubTime: Int64
    public var cover: String?
    public var desc: String?
    public var ownerName: String?
    public var ownerMid: String?
    public var points: [ViewPoint] = []
    /// Set when the site's info already carried the page's streams and subtitles (every site but bilibili).
    public var media: PageMedia?

    public init(index: Int, aid: String, cid: String, epid: String, title: String, duration: Int, resolution: String,
                pubTime: Int64, cover: String? = nil, desc: String? = nil, ownerName: String? = nil, ownerMid: String? = nil) {
        self.index = index; self.aid = aid; self.cid = cid; self.epid = epid; self.title = title
        self.duration = duration; self.resolution = resolution; self.pubTime = pubTime
        self.cover = cover; self.desc = desc; self.ownerName = ownerName; self.ownerMid = ownerMid
    }

    public var bvid: String {
        guard let n = Int64(aid), let bv = try? BVConverter.encode(n) else { return "" }
        return bv
    }

    /// Identity is the (aid, cid, epid) triple; index and titles are presentation.
    public static func == (a: Page, b: Page) -> Bool { a.aid == b.aid && a.cid == b.cid && a.epid == b.epid }
    public func hash(into h: inout Hasher) { h.combine(aid); h.combine(cid); h.combine(epid) }
}

/// What a site's info carried for one page beyond the `Page` fields: its streams, subtitles and request headers.
public struct PageMedia: Sendable {
    public var tracks: ParsedTracks
    public var subtitles: [SubtitleInfo]
    /// Request headers the stream URLs were issued for.
    public var headers: [String: String]
    public var webpageURL: String
    /// A podcast episode's show; it becomes the album tag, with the episode as the title.
    public var album: String

    public init(tracks: ParsedTracks, subtitles: [SubtitleInfo], headers: [String: String], webpageURL: String, album: String = "") {
        self.tracks = tracks; self.subtitles = subtitles; self.headers = headers; self.webpageURL = webpageURL; self.album = album
    }
}

public struct ViewPoint: Sendable, Equatable {
    public var title: String
    public var start: Int
    public var end: Int
    public init(title: String, start: Int, end: Int) { self.title = title; self.start = start; self.end = end }
}

public struct VideoTrack: Sendable {
    public var id: String
    public var quality: String
    public var baseURL: String
    public var resolution: String?
    public var fps: String?
    public var codec: String
    public var bandwidth: Int64
    public var duration: Int
    public var size: Double

    public init(id: String, quality: String, baseURL: String, resolution: String? = nil, fps: String? = nil,
                codec: String, bandwidth: Int64, duration: Int, size: Double = 0) {
        self.id = id; self.quality = quality; self.baseURL = baseURL; self.resolution = resolution; self.fps = fps
        self.codec = codec; self.bandwidth = bandwidth; self.duration = duration; self.size = size
    }

    /// Estimated size when the API did not state one.
    public func estimatedSize(pageDuration: Int) -> Double {
        if size > 0 { return size }
        let d = pageDuration == 0 ? duration : pageDuration
        return Double(d) * Double(bandwidth) * 1024 / 8
    }
}

extension VideoTrack: Equatable {
    public static func == (a: VideoTrack, b: VideoTrack) -> Bool {
        a.id == b.id && a.quality == b.quality && a.resolution == b.resolution && a.fps == b.fps
            && a.codec == b.codec && a.bandwidth == b.bandwidth && a.duration == b.duration
    }
}

public struct AudioTrack: Sendable {
    public var id: String
    public var quality: String
    public var baseURL: String
    public var codec: String
    public var bandwidth: Int64
    public var duration: Int
    public var size: Double

    public init(id: String, quality: String, baseURL: String, codec: String, bandwidth: Int64, duration: Int, size: Double = 0) {
        self.id = id; self.quality = quality; self.baseURL = baseURL; self.codec = codec
        self.bandwidth = bandwidth; self.duration = duration; self.size = size
    }

    /// `E-AC-3` → `EAC3`, the form priorities are written in.
    public var shortCodec: String { codec.uppercased().replacingOccurrences(of: "-", with: "") }

    public func estimatedSize(pageDuration: Int) -> Double {
        if size > 0 { return size }
        let d = pageDuration == 0 ? duration : pageDuration
        return Double(d) * Double(bandwidth) * 1024 / 8
    }
}

extension AudioTrack: Equatable {
    public static func == (a: AudioTrack, b: AudioTrack) -> Bool {
        a.id == b.id && a.quality == b.quality && a.codec == b.codec && a.bandwidth == b.bandwidth && a.duration == b.duration
    }
}

public struct SubtitleInfo: Sendable {
    public var language: String
    public var url: String
    public var path: String
    public init(language: String, url: String, path: String) { self.language = language; self.url = url; self.path = path }
}

/// A dubbing track (PGC only): background audio or one voice role.
public struct RoleAudio: Sendable {
    public var title: String
    public var personName: String
    public var path: String
    public var tracks: [AudioTrack]
    public init(title: String, personName: String, path: String, tracks: [AudioTrack]) {
        self.title = title; self.personName = personName; self.path = path; self.tracks = tracks
    }
}

/// A downloaded extra audio file to mux in.
public struct AudioMaterial: Sendable {
    public var title: String
    public var personName: String
    public var path: String
    public init(title: String, personName: String, path: String) { self.title = title; self.personName = personName; self.path = path }
}

public struct VideoInfo: Sendable {
    public var title: String
    public var desc: String
    public var cover: String
    public var pubTime: Int64
    public var pages: [Page]
    public var isBangumi = false
    public var isCheese = false
    public var isBangumiEnd = false
    public var isInteractive = false
    /// Which page the user's input pointed at (episode/season links), as a 1-based index string.
    public var index: String?
    public var site: Site = .bilibili

    public init(title: String, desc: String, cover: String, pubTime: Int64, pages: [Page]) {
        self.title = title; self.desc = desc; self.cover = cover; self.pubTime = pubTime; self.pages = pages
    }
}

/// Everything a playurl response yielded.
public struct ParsedTracks: Sendable {
    public var rawJSON: String = ""
    public var video: [VideoTrack] = []
    public var audio: [AudioTrack] = []
    public var backgroundAudio: [AudioTrack] = []
    public var roleAudio: [RoleAudio] = []
    public var extraPoints: [ViewPoint] = []
    /// FLV mode: segment URLs and the quality ids the server offers.
    public var clips: [String] = []
    public var qualities: [String] = []
    /// The video tracks are whole files with the audio inside (X); there are no separate audio tracks.
    public var videoHasAudio = false
    public init() {}

    public var isDash: Bool { (!video.isEmpty || !audio.isEmpty) && clips.isEmpty }
    public var isFLV: Bool { !clips.isEmpty && !qualities.isEmpty }
}
