import Foundation
import Testing
@testable import HaulCore

@Suite struct YtDlpTests {
    @Test func recognisesVideoLinks() {
        #expect(Site.youtubeID(from: "https://youtu.be/DdCEmlAydcw") == "DdCEmlAydcw")
        #expect(Site.youtubeID(from: "https://youtu.be/DdCEmlAydcw?si=abc&t=10") == "DdCEmlAydcw")
        #expect(Site.youtubeID(from: "https://www.youtube.com/watch?v=DdCEmlAydcw&list=PL1") == "DdCEmlAydcw")
        #expect(Site.youtubeID(from: "m.youtube.com/watch?feature=share&v=-abc_DEF123") == "-abc_DEF123")
        #expect(Site.youtubeID(from: "https://www.youtube.com/shorts/DdCEmlAydcw") == "DdCEmlAydcw")
        #expect(Site.youtubeID(from: "https://www.youtube-nocookie.com/embed/DdCEmlAydcw") == "DdCEmlAydcw")
        #expect(Site.youtubeID(from: "https://www.youtube.com/playlist?list=PL1") == nil)
        #expect(Site.youtubeID(from: "https://www.bilibili.com/video/BV1qt4y1X7TW") == nil)
        #expect(Site.youtubeID(from: "BV1qt4y1X7TW") == nil)
    }

    @Test func recognisesXLinks() {
        #expect(Site.tweetID(from: "https://x.com/historyinmemes/status/1790637656616943991") == "1790637656616943991")
        #expect(Site.tweetID(from: "https://twitter.com/CTVJLaidlaw/status/1600649710662213632/video/2") == "1600649710662213632")
        #expect(Site.tweetID(from: "mobile.twitter.com/i/web/status/910031516746514432?s=20") == "910031516746514432")
        #expect(Site.tweetID(from: "https://x.com/historyinmemes") == nil)
        #expect(Site.tweetID(from: "https://notx.com/a/status/1") == nil)
    }

    @Test func resolverRoutesOtherSites() async throws {
        #expect(try await IDResolver.resolve("https://youtu.be/DdCEmlAydcw?si=x", session: Session())
                == .link(site: .youtube, url: "https://www.youtube.com/watch?v=DdCEmlAydcw"))
        #expect(try await IDResolver.resolve("x.com/a/status/12/video/2", session: Session())
                == .link(site: .x, url: "https://x.com/a/status/12/video/2"))
    }

    static let sample = #"""
    {
      "id": "DdCEmlAydcw", "title": "Lab tour", "description": "desc", "channel": "Anthropic", "channel_id": "UC1",
      "timestamp": 1790186492, "duration": 75, "language": "en-US",
      "webpage_url": "https://www.youtube.com/watch?v=DdCEmlAydcw",
      "chapters": [{"title": "Intro", "start_time": 0.0, "end_time": 12.5}],
      "thumbnails": [
        {"url": "https://i.ytimg.com/vi/DdCEmlAydcw/hqdefault.jpg?sqp=x", "preference": -7, "width": 336},
        {"url": "https://i.ytimg.com/vi/DdCEmlAydcw/maxresdefault.jpg", "preference": -1},
        {"url": "https://i.ytimg.com/vi_webp/DdCEmlAydcw/maxresdefault.webp", "preference": 0}
      ],
      "formats": [
        {"format_id": "sb0", "protocol": "mhtml", "url": "u", "vcodec": "none", "acodec": "none"},
        {"format_id": "620", "protocol": "m3u8_native", "url": "u", "vcodec": "vp09.00.50.08", "acodec": "none"},
        {"format_id": "140-0", "protocol": "https", "url": "a-dub", "vcodec": "none", "acodec": "mp4a.40.2", "abr": 129.5,
         "filesize": 1222496, "language_preference": -1},
        {"format_id": "140-19", "protocol": "https", "url": "a-orig", "vcodec": "none", "acodec": "mp4a.40.2", "abr": 129.6,
         "filesize": 1220994, "language_preference": 10, "http_headers": {"User-Agent": "UA"}},
        {"format_id": "251-drc", "protocol": "https", "url": "a-drc", "vcodec": "none", "acodec": "opus", "abr": 131, "language_preference": 10},
        {"format_id": "18", "protocol": "https", "url": "muxed", "vcodec": "avc1.42001E", "acodec": "mp4a.40.2"},
        {"format_id": "271", "protocol": "https", "url": "v", "vcodec": "vp9", "acodec": "none", "format_note": "1080p",
         "width": 2048, "height": 1240, "fps": 24, "vbr": 5826.7, "filesize": 54868477}
      ],
      "subtitles": {"en": [{"ext": "vtt", "url": "s-vtt"}, {"ext": "json3", "url": "s-en"}], "live_chat": [{"ext": "json3", "url": "chat"}]},
      "automatic_captions": {
        "en-orig": [{"ext": "json3", "url": "asr-en"}], "fr-FR-orig": [{"ext": "json3", "url": "asr-fr"}],
        "de": [{"ext": "json3", "url": "tr-de"}]
      }
    }
    """#

    @Test func parsesYtDlpJSON() throws {
        let info = try YtDlp.parse(JSON.parse(Self.sample), site: .youtube)
        #expect(info.site == .youtube)
        #expect(info.title == "Lab tour")
        #expect(info.pubTime == 1790186492)
        #expect(info.cover == "https://i.ytimg.com/vi/DdCEmlAydcw/maxresdefault.jpg")
        let page = try #require(info.pages.first)
        #expect(page.aid == "DdCEmlAydcw" && page.ownerName == "Anthropic" && page.duration == 75)
        #expect(page.points == [ViewPoint(title: "Intro", start: 0, end: 12)])

        let yt = try #require(page.media)
        #expect(yt.headers == ["User-Agent": "UA"])
        #expect(yt.tracks.video.map(\.baseURL) == ["v"])
        let v = yt.tracks.video[0]
        #expect(v.codec == "VP9" && v.quality == "1080p" && v.resolution == "2048x1240" && v.fps == "24" && v.bandwidth == 5827)
        #expect(v.id == "1240024" && v.size == 54868477)
        // DASH present: the muxed format 18 is ignored.
        #expect(!yt.tracks.videoHasAudio)
        // Only the original-language audio, no DRC variant.
        #expect(yt.tracks.audio.map(\.baseURL) == ["a-orig"])
        #expect(yt.tracks.audio[0].codec == "M4A" && yt.tracks.audio[0].bandwidth == 130)
        // Uploaded track plus the spoken-language auto track; no chat, translations or dub transcripts.
        #expect(yt.subtitles.map(\.language) == ["en", "ai-en"])
        #expect(yt.subtitles.map(\.url) == ["s-en", "asr-en"])
        #expect(yt.subtitles[1].path == "DdCEmlAydcw/DdCEmlAydcw.ai-en.srt")
    }

    @Test func onlyXTitlesLoseTheEllipsis() throws {
        let json = #"{"id": "DdCEmlAydcw", "title": "Wait for it...", "formats": [{"format_id": "18", "protocol": "https", "url": "u", "vcodec": "avc1", "acodec": "mp4a.40.2", "width": 640, "height": 360}]}"#
        #expect(try YtDlp.parse(JSON.parse(json), site: .youtube).title == "Wait for it...")
        #expect(try YtDlp.parse(JSON.parse(json), site: .x).title == "Wait for it")
    }

    @Test func noDownloadableStreamsIsAnError() {
        #expect(throws: HaulError.self) {
            try YtDlp.parse(JSON.parse(#"{"id": "DdCEmlAydcw", "formats": []}"#), site: .youtube)
        }
    }

    static let xPost = #"""
    {
      "_type": "playlist", "id": "1600649710662213632", "title": "Jocelyn Laidlaw - How her diagnosis changed...",
      "description": "full text", "uploader": "Jocelyn Laidlaw", "timestamp": 1670459604,
      "webpage_url": "https://twitter.com/CTVJLaidlaw/status/1600649710662213632",
      "entries": [
        {"id": "1600649511827038209", "title": "Jocelyn Laidlaw - How her diagnosis changed...", "duration": 113.49,
         "uploader": "Jocelyn Laidlaw", "uploader_id": "JocelynVLaidlaw",
         "thumbnails": [{"url": "https://pbs.twimg.com/a.jpg?name=small", "width": 680}, {"url": "https://pbs.twimg.com/a.jpg?name=orig", "width": 886}],
         "formats": [
           {"format_id": "hls-audio-128000-Audio", "protocol": "m3u8_native", "url": "hls", "vcodec": "none", "acodec": null},
           {"format_id": "http-632", "protocol": "https", "url": "https://video.twimg.com/ext_tw_video/1/pu/vid/320x568/a.mp4",
            "width": 320, "height": 568, "tbr": 632, "filesize_approx": 8965710},
           {"format_id": "http-2176", "protocol": "https", "url": "https://video.twimg.com/amplify_video/1/vid/avc1/720x1280/b.mp4",
            "width": 720, "height": 1280, "tbr": 2176, "http_headers": {"User-Agent": "UA"}}
         ]},
        {"id": "1600649511827013632", "duration": 102.2,
         "formats": [{"format_id": "http-950", "protocol": "https", "url": "c.mp4", "width": 480, "height": 852, "tbr": 950}]}
      ]
    }
    """#

    @Test func parsesMultiVideoXPost() throws {
        let info = try YtDlp.parse(JSON.parse(Self.xPost), site: .x)
        #expect(info.title == "Jocelyn Laidlaw - How her diagnosis changed")
        #expect(info.cover == "" && info.pubTime == 1670459604)
        #expect(info.pages.map(\.index) == [1, 2])
        #expect(info.pages.map(\.title) == ["Video 1", "Video 2"])
        #expect(info.pages.map(\.aid) == ["1600649511827038209", "1600649511827013632"])

        let first = info.pages[0]
        #expect(first.cover == "https://pbs.twimg.com/a.jpg?name=orig")
        #expect(first.ownerName == "Jocelyn Laidlaw" && first.duration == 113)
        let media = try #require(first.media)
        #expect(media.webpageURL == "https://twitter.com/CTVJLaidlaw/status/1600649710662213632")
        #expect(media.headers == ["User-Agent": "UA"])
        // Whole files with the audio inside, no HLS; X's approximate sizes are not trusted.
        #expect(media.tracks.videoHasAudio && media.tracks.audio.isEmpty)
        #expect(media.tracks.video.map(\.quality) == ["320p", "720p"])
        #expect(media.tracks.video.map(\.codec) == ["MP4", "AVC"])
        #expect(media.tracks.video.allSatisfy { $0.size == 0 })
        #expect(info.pages[1].media?.tracks.video.first?.bandwidth == 950)
    }

    @Test func json3CaptionsBecomeSRT() throws {
        let json = try JSON.parse(#"""
        {"events": [
          {"tStartMs": 0, "dDurationMs": 5000, "id": 1},
          {"tStartMs": 805, "dDurationMs": 2368, "segs": [{"utf8": "How much"}, {"utf8": " is left"}]},
          {"tStartMs": 3100, "dDurationMs": 10, "aAppend": 1, "segs": [{"utf8": "\n"}]},
          {"tStartMs": 61000, "dDurationMs": 1500, "segs": [{"utf8": "Done."}]}
        ]}
        """#)
        #expect(try Subtitles.srt(fromJSON: json.serialized) == """
        1
        00:00:00,805 --> 00:00:03,173
        How much is left

        2
        00:01:01,000 --> 00:01:02,500
        Done.


        """)
    }

    @Test func rangeOnlyClipsAreClosed() {
        let clips = Downloader.makeClips(25, clipSize: 9, closed: true)
        #expect(clips.map(\.from) == [0, 10, 20])
        #expect(clips.map(\.to) == [9, 19, 24])
    }
}
