import Foundation

/// Xiaoyuzhou and Apple Podcasts: `Podcast` reads the episodes; an audio episode is saved as audio.
struct PodcastSource: MediaSource {
    let site: Site
    let http: HTTPClient

    func fetchInfo(_ id: MediaID) async throws -> (MediaID, VideoInfo) {
        guard case .link(_, let url) = id else { throw HaulError.input("Not a \(site.name) link: \(id)") }
        return (id, try await Podcast.fetch(url, site: site, http: http))
    }

    /// Video podcasts (Apple) are one MP4 with the audio inside, and download like an X post.
    func isAudioOnly(_ page: Page) -> Bool { page.media?.tracks.video.isEmpty == true }
}
