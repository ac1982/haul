import Foundation

/// Output path templates: `<title>/[P<pageNumberWithZero>]<pageTitle>` and friends.
public enum FilePattern {
    public static let singleDefault = "<title>"
    public static let multiDefault = "<title>/[P<pageNumberWithZero>]<pageTitle>"

    public static let variables: [(String, String)] = [
        ("title", "title of the video, post, show or list"), ("pageNumber", "page / episode number"),
        ("pageNumberWithZero", "page number, zero-padded"), ("pageTitle", "page / episode title"),
        ("id", "the page's id on its site (YouTube id, BV…, episode id)"), ("site", "youtube, x, bilibili, xiaoyuzhou, applePodcasts"),
        ("uploader", "uploader / channel / host name"), ("uploaderId", "uploader id"),
        ("quality", "video quality label (1080p, 4K…)"), ("resolution", "video resolution"), ("fps", "video frame rate"),
        ("videoCodec", "video codec"), ("videoBitrate", "video bitrate, kbps"), ("audioCodec", "audio codec"),
        ("audioBitrate", "audio bitrate, kbps"), ("publishDate", "publish date of the item or list"),
        ("pageDate", "publish date of the page"),
        ("bvid", "bilibili BV id"), ("aid", "bilibili aid"), ("cid", "bilibili cid"), ("api", "bilibili API: WEB / TV / APP / INTL"),
    ]

    public struct Context: Sendable {
        public var title: String
        public var video: VideoTrack?
        public var audio: AudioTrack?
        public var page: Page
        public var pageCount: Int
        public var apiType: String
        public var pubTime: Int64
        public var site: Site

        public init(title: String, video: VideoTrack?, audio: AudioTrack?, page: Page, pageCount: Int, apiType: String, pubTime: Int64,
                    site: Site = .bilibili) {
            self.title = title; self.video = video; self.audio = audio; self.page = page
            self.pageCount = pageCount; self.apiType = apiType; self.pubTime = pubTime; self.site = site
        }
    }

    private static let defaultDateFormat = "yyyy-MM-dd_HH-mm-ss"

    public static func render(_ pattern: String, _ c: Context) -> String {
        var result = pattern.replacingOccurrences(of: "\\", with: "/")
        for m in pattern.matches(of: /<([\w:\-.]+?)>/) {
            var key = String(m.1)
            var dateFormat = defaultDateFormat
            for prefix in ["publishDate:", "pageDate:"] where key.hasPrefix(prefix) {
                dateFormat = String(key.dropFirst(prefix.count))
                key = String(prefix.dropLast())
            }
            let value: String
            switch key {
            case "title": value = Format.cleanName(c.title)
            case "pageNumber": value = String(c.page.index)
            case "pageNumberWithZero":
                let width = String(c.pageCount).count
                value = String(repeating: "0", count: max(0, width - String(c.page.index).count)) + String(c.page.index)
            case "pageTitle": value = Format.cleanName(c.page.title)
            case "id": value = Format.cleanName(Report.pageID(c.page, site: c.site))
            case "site": value = c.site.rawValue
            case "uploader": value = c.page.ownerName.map(Format.cleanName) ?? ""
            case "uploaderId": value = c.page.ownerMid ?? ""
            case "quality": value = c.video?.quality ?? ""
            case "resolution": value = c.video?.resolution ?? ""
            case "fps": value = c.video?.fps ?? ""
            case "videoCodec": value = c.video?.codec ?? ""
            case "videoBitrate": value = c.video.map { String($0.bandwidth) } ?? ""
            case "audioCodec": value = c.audio?.codec ?? ""
            case "audioBitrate": value = c.audio.map { String($0.bandwidth) } ?? ""
            case "publishDate": value = Format.timestamp(c.pubTime, pattern: dateFormat)
            case "pageDate": value = Format.timestamp(c.page.pubTime, pattern: dateFormat)
            case "bvid": value = c.page.bvid
            case "aid": value = c.page.aid
            case "cid": value = c.page.cid
            case "api": value = c.apiType
            default: value = "<\(key)>"
            }
            result = result.replacingOccurrences(of: String(m.output.0), with: value)
        }
        if !result.hasSuffix(".mp4") { result += ".mp4" }
        return result
    }

    /// Swap the extension of an output path (`a/b.mp4` → `a/b.xml`).
    public static func changeExtension(_ path: String, to ext: String) -> String {
        ((path as NSString).deletingPathExtension as NSString).appendingPathExtension(ext) ?? path + "." + ext
    }
}
