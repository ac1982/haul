import Foundation
import Testing
@testable import HaulCore

@Suite struct MuxerArgumentTests {
    /// The value after each occurrence of `flag`.
    func values(_ args: [String], _ flag: String) -> [String] {
        args.indices.filter { args[$0] == flag && $0 + 1 < args.count }.map { args[$0 + 1] }
    }

    func job(_ dir: String) throws -> MuxJob {
        var job = MuxJob()
        job.videoPath = "\(dir)/v.mp4"
        job.audioPath = "\(dir)/a.m4a"
        job.outputPath = "\(dir)/out/T.mp4"
        job.title = "T"
        job.author = "UP"
        job.desc = "D"
        job.pageURL = "https://www.bilibili.com/video/BV1qt4y1X7TW/"
        job.pubTime = 1595684326
        return job
    }

    @Test func fullJob() throws {
        let dir = try makeTempDir("mux")
        let sub = "\(dir)/en.srt"
        try "1\n00:00:00,000 --> 00:00:01,000\nhi\n".write(toFile: sub, atomically: true, encoding: .utf8)
        var j = try job(dir)
        j.coverPath = "\(dir)/c.jpg"
        j.subtitles = [SubtitleInfo(language: "en", url: "", path: sub),
                       SubtitleInfo(language: "zh-CN", url: "", path: "\(dir)/missing.srt")]
        j.chapters = [ViewPoint(title: "A", start: 0, end: 10)]
        j.isHEVC = true
        let args = Muxer.ffmpegArguments(Muxer.normalized(j))

        #expect(values(args, "-i") == ["\(dir)/v.mp4", "\(dir)/a.m4a", "\(dir)/c.jpg", sub, j.chaptersFile])
        #expect(values(args, "-map") == ["0", "1", "2", "3"])
        #expect(values(args, "-map_chapters") == ["4"])
        #expect(values(args, "-disposition:v:1") == ["attached_pic"])
        // An empty subtitle file is left out.
        #expect(values(args, "-metadata:s:s:0") == ["title=English", "language=eng"])
        #expect(values(args, "-c:s") == ["mov_text"])
        #expect(values(args, "-tag:v:0") == ["hvc1"])
        #expect(values(args, "-metadata").contains("title=T"))
        #expect(values(args, "-metadata").contains("comment=https://www.bilibili.com/video/BV1qt4y1X7TW/"))
        #expect(values(args, "-metadata").contains("artist=UP"))
        #expect(values(args, "-metadata").contains("description=D"))
        #expect(values(args, "-metadata").contains { $0.hasPrefix("creation_time=2020-07-25") })
        #expect(args.suffix(2) == ["--", "\(dir)/out/T.mp4"])
        #expect(!args.contains("-vn") && !args.contains("-an"))
    }

    @Test func episodeTitleAndAlbum() throws {
        var j = try job(try makeTempDir("mux"))
        j.episodeTitle = "P2"
        let meta = values(Muxer.ffmpegArguments(Muxer.normalized(j)), "-metadata")
        #expect(meta.contains("title=P2") && meta.contains("album=T"))
    }

    @Test func audioOnlyDropsTheSeparateVideo() throws {
        let dir = try makeTempDir("mux")
        var j = try job(dir)
        j.audioOnly = true
        j.coverPath = "\(dir)/c.jpg"
        let args = Muxer.ffmpegArguments(Muxer.normalized(j))
        #expect(values(args, "-i") == ["\(dir)/a.m4a", "\(dir)/c.jpg"])
        #expect(values(args, "-map") == ["0:a", "1"])
        #expect(values(args, "-disposition:v:0") == ["attached_pic"])
        #expect(!args.contains("-vn"))
    }

    @Test func audioOnlyFromAVideoWithAudioInside() throws {
        let dir = try makeTempDir("mux")
        var j = try job(dir)
        j.audioPath = ""
        j.audioOnly = true
        j.coverPath = "\(dir)/c.jpg"
        let args = Muxer.ffmpegArguments(Muxer.normalized(j))
        #expect(values(args, "-i") == ["\(dir)/v.mp4", "\(dir)/c.jpg"])
        // Only the audio of the video file; the cover is then video stream 0.
        #expect(values(args, "-map") == ["0:a", "1"])
        #expect(values(args, "-disposition:v:0") == ["attached_pic"])
        #expect(!args.contains("-vn"))
    }

    @Test func podcastMP3StaysMP3() throws {
        let dir = try makeTempDir("mux")
        var j = try job(dir)
        j.videoPath = ""
        j.audioPath = "\(dir)/a.mp3"
        j.outputPath = "\(dir)/out/E.mp3"
        j.container = .mp3
        j.audioOnly = true
        j.coverPath = "\(dir)/c.jpg"
        j.episodeTitle = "E"
        let args = Muxer.ffmpegArguments(Muxer.normalized(j))
        #expect(values(args, "-i") == ["\(dir)/a.mp3", "\(dir)/c.jpg"])
        // A picture embedded in the MP3 is left behind, so ours is video stream 0.
        #expect(values(args, "-map") == ["0:a", "1"])
        #expect(values(args, "-disposition:v:0") == ["attached_pic"])
        #expect(values(args, "-metadata:s:v:0") == ["comment=Cover (front)"])
        let meta = values(args, "-metadata")
        #expect(meta.contains("title=E") && meta.contains("album=T") && meta.contains { $0.hasPrefix("date=2020-07-2") })
        #expect(values(args, "-f") == ["mp3"])
        #expect(!args.contains("-movflags") && !args.contains("-strict"))
        #expect(args.suffix(2) == ["--", "\(dir)/out/E.mp3"])
    }

    @Test func videoOnlyDropsAllAudio() throws {
        let dir = try makeTempDir("mux")
        var j = try job(dir)
        j.videoOnly = true
        let args = Muxer.ffmpegArguments(Muxer.normalized(j))
        #expect(values(args, "-i") == ["\(dir)/v.mp4"])
        #expect(args.contains("-an"))
    }

    @Test func simpleMuxWritesNoTags() throws {
        var j = try job(try makeTempDir("mux"))
        j.simpleMux = true
        #expect(values(Muxer.ffmpegArguments(Muxer.normalized(j)), "-metadata").isEmpty)
    }

    @Test func extraAudioTracks() throws {
        let dir = try makeTempDir("mux")
        var j = try job(dir)
        j.extraAudio = [AudioMaterial(title: "背景音频", personName: "", path: "\(dir)/bg.m4a"),
                        AudioMaterial(title: "角色", personName: "配音演员", path: "\(dir)/r.m4a")]
        let args = Muxer.ffmpegArguments(Muxer.normalized(j))
        #expect(values(args, "-i") == ["\(dir)/v.mp4", "\(dir)/a.m4a", "\(dir)/bg.m4a", "\(dir)/r.m4a"])
        #expect(values(args, "-metadata:s:a:0") == ["title=Original audio"])
        #expect(values(args, "-metadata:s:a:2") == ["title=角色", "artist=配音演员"])
    }
}
