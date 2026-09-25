import Foundation

/// YouTube and X: yt-dlp finds the streams; the pages carry them.
struct YtDlpSource: MediaSource {
    /// googlevideo answers only closed byte ranges; plain and open-ended requests get 403.
    static let googlevideoRange: Int64 = 10 * 1024 * 1024

    let site: Site
    /// Resolved only when the run's link is YouTube or X.
    let ytdlp: String?
    let http: HTTPClient

    func fetchInfo(_ id: MediaID) async throws -> (MediaID, VideoInfo) {
        guard case .link(_, let url) = id else { throw HaulError.input("Not a \(site.name) link: \(id)") }
        guard let ytdlp else { throw HaulError.dependency("\(site.name) links need yt-dlp: brew install yt-dlp deno") }
        return (id, try await YtDlp.fetch(url, site: site, ytdlp: ytdlp))
    }

    var maxRange: Int64? { site == .youtube ? Self.googlevideoRange : nil }
}
