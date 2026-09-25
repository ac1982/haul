import Foundation

public struct MuxJob: Sendable {
    public var videoPath = ""
    public var audioPath = ""
    public var extraAudio: [AudioMaterial] = []
    public var outputPath = ""
    public var desc = ""
    public var title = ""
    public var author = ""
    /// Set for multi-page/episode output; becomes the track title, with `title` as album.
    public var episodeTitle = ""
    public var coverPath = ""
    public var language = ""
    public var subtitles: [SubtitleInfo] = []
    public var audioOnly = false
    public var videoOnly = false
    public var chapters: [ViewPoint] = []
    public var pubTime: Int64 = 0
    public var simpleMux = false
    public var isHEVC = false
    /// MP4 and M4A are written the same way; MP3 gets ID3 tags.
    public var container: Container = .mp4

    public init() {}

    /// Written into the comment tag.
    public var pageURL = ""
    var workDir: String { ((videoPath.isEmpty ? audioPath : videoPath) as NSString).deletingLastPathComponent }
    /// Where the chapter list is written for the muxer to read.
    public var chaptersFile: String { (workDir as NSString).appendingPathComponent("chapters") }

    /// Subtitles that actually have content; empty files break the muxers.
    var usableSubtitles: [SubtitleInfo] {
        subtitles.filter { Downloader.fileSize($0.path) > 0 }
    }
}

/// Wraps ffmpeg / MP4Box. Arguments go through argv, so titles with quotes or backslashes need no escaping.
public enum Muxer {
    public static func mux(_ input: MuxJob, useMP4Box: Bool, ffmpeg: String, mp4box: String) async throws -> Int32 {
        let job = normalized(input)
        if useMP4Box { return try await muxMP4Box(job, mp4box: mp4box) }
        return try await muxFFmpeg(job, ffmpeg: ffmpeg)
    }

    /// --audio-only drops a separate video file, --video-only the audio file.
    static func normalized(_ input: MuxJob) -> MuxJob {
        var job = input
        if job.audioOnly && !job.audioPath.isEmpty { job.videoPath = "" }
        if job.videoOnly { job.audioPath = "" }
        return job
    }

    static func muxFFmpeg(_ job: MuxJob, ffmpeg: String) async throws -> Int32 {
        let outDir = (job.outputPath as NSString).deletingLastPathComponent
        if !outDir.isEmpty { try FileManager.default.createDirectory(atPath: outDir, withIntermediateDirectories: true) }
        if !job.chapters.isEmpty {
            try ffmpegChapters(job.chapters).write(toFile: job.chaptersFile, atomically: true, encoding: .utf8)
        }
        let args = ffmpegArguments(job)
        Log.debug("ffmpeg \(args.joined(separator: " "))")
        return try await Shell.run(ffmpeg, args).status
    }

    /// The ffmpeg command line for a job; the chapters file, if any, is expected at `job.chaptersFile`.
    static func ffmpegArguments(_ job: MuxJob) -> [String] {
        let mp3 = job.container == .mp3
        var inputs: [String] = []
        var meta: [String] = []
        var inputCount = 0

        for path in [job.videoPath, job.audioPath] where !path.isEmpty {
            inputs += ["-i", path]
            inputCount += 1
        }
        if !job.extraAudio.isEmpty {
            meta += ["-metadata:s:a:0", "title=Original audio"]
            for (n, audio) in job.extraAudio.enumerated() {
                inputs += ["-i", audio.path]
                inputCount += 1
                let idx = n + 1
                if !audio.title.trimmingCharacters(in: .whitespaces).isEmpty { meta += ["-metadata:s:a:\(idx)", "title=\(audio.title)"] }
                if !audio.personName.trimmingCharacters(in: .whitespaces).isEmpty { meta += ["-metadata:s:a:\(idx)", "artist=\(audio.personName)"] }
            }
        }
        if !job.coverPath.isEmpty {
            inputs += ["-i", job.coverPath]
            inputCount += 1
        }
        let subs = job.usableSubtitles
        for (n, sub) in subs.enumerated() {
            inputs += ["-i", sub.path]
            inputCount += 1
            let info = Subtitles.languageInfo(sub.language)
            meta += ["-metadata:s:s:\(n)", "title=\(info.name)", "-metadata:s:s:\(n)", "language=\(info.code)"]
        }
        if !job.coverPath.isEmpty {
            meta += ["-disposition:v:\(job.audioOnly ? 0 : 1)", "attached_pic"]
            // ID3 pictures carry a type; players look for the front cover.
            if mp3 { meta += ["-metadata:s:v:0", "comment=Cover (front)"] }
        }
        if !job.chapters.isEmpty {
            inputs += ["-i", job.chaptersFile, "-map_chapters", "\(inputCount)"]
        }
        // Audio-only: take just the audio of the first file, so our cover stays the only video stream —
        // X and FLV files carry video, podcast MP3s / M4As often an embedded picture of their own.
        for i in 0..<inputCount { inputs += ["-map", i == 0 && job.audioOnly ? "0:a" : "\(i)"] }

        var args = ["-loglevel", Log.debugEnabled ? "verbose" : "warning", "-y"]
        args += inputs
        args += meta
        if !job.simpleMux {
            args += ["-metadata", "title=\(job.episodeTitle.isEmpty ? job.title : job.episodeTitle)"]
            args += ["-metadata", "comment=\(job.pageURL)"]
            if !job.language.isEmpty { args += ["-metadata:s:a:0", "language=\(job.language)"] }
            if !job.desc.trimmingCharacters(in: .whitespaces).isEmpty { args += ["-metadata", "description=\(job.desc)"] }
            if !job.author.isEmpty { args += ["-metadata", "artist=\(job.author)"] }
            if !job.episodeTitle.isEmpty { args += ["-metadata", "album=\(job.title)"] }
            if job.pubTime != 0 { args += ["-metadata", "creation_time=\(Format.ffmpegCreationTime(job.pubTime))"] }
            if mp3 && job.pubTime != 0 { args += ["-metadata", "date=\(Format.timestamp(job.pubTime, pattern: "yyyy-MM-dd"))"] }
        }
        args += ["-c:v", "copy", "-c:a", "copy"]
        if mp3 { return args + ["-f", "mp3", "--", job.outputPath] }
        // A video file may carry its own audio (X).
        if job.videoOnly { args.append("-an") }
        if !subs.isEmpty { args += ["-c:s", "mov_text"] }
        #if os(macOS)
        // QuickTime only plays HEVC tagged hvc1, not hev1.
        if job.isHEVC { args += ["-tag:v:0", "hvc1"] }
        #endif
        args += ["-movflags", "faststart", "-strict", "unofficial", "-strict", "-2", "-f", "mp4", "--", job.outputPath]
        return args
    }

    static func muxMP4Box(_ job: MuxJob, mp4box: String) async throws -> Int32 {
        var args: [String] = []
        if Log.debugEnabled { args.append("-v") }
        args += ["-inter", "500", "-noprog"]
        var trackID = 0
        if !job.videoPath.isEmpty {
            args += ["-add", "\(job.videoPath)#trackID=\(job.audioOnly && job.audioPath.isEmpty ? "2" : "1"):name="]
            trackID += 1
        }
        if !job.audioPath.isEmpty {
            args += ["-add", "\(job.audioPath):lang=\(job.language.isEmpty ? "und" : job.language)"]
            trackID += 1
        }
        if !job.chapters.isEmpty {
            try mp4boxChapters(job.chapters).write(toFile: job.chaptersFile, atomically: true, encoding: .utf8)
            args += ["-chap", job.chaptersFile]
        }
        var tags = ""
        if !job.coverPath.isEmpty { tags += ":cover=\"\(job.coverPath)\"" }
        if !job.episodeTitle.isEmpty { tags += ":album=\"\(job.title)\":title=\"\(job.episodeTitle)\"" } else { tags += ":title=\"\(job.title)\"" }
        tags += ":sdesc=\"\(job.desc)\":comment=\"\(job.pageURL)\":artist=\"\(job.author)\""
        for sub in job.usableSubtitles {
            trackID += 1
            let info = Subtitles.languageInfo(sub.language)
            args += ["-add", "\(sub.path)#trackID=1:name=:hdlr=sbtl:lang=\(info.code)"]
            args += ["-udta", "\(trackID):type=name:str=\"\(info.name)\""]
        }
        args += ["-itags", "tool=" + tags, "-new", job.outputPath]
        Log.debug("mp4box \(args.joined(separator: " "))")
        return try await Shell.run(mp4box, args).status
    }

    /// FLV segments → one MP4: each segment is remuxed to MPEG-TS, the TS files are concatenated.
    public static func mergeFLV(files: [String], to output: String, ffmpeg: String) async throws {
        if files.count == 1 {
            try Downloader.replaceItem(at: output, with: files[0])
            return
        }
        for file in files {
            let ts = ((file as NSString).deletingPathExtension) + ".ts"
            let args = ["-loglevel", "warning", "-y", "-i", file, "-map", "0", "-c", "copy", "-f", "mpegts", "-bsf:v", "h264_mp4toannexb", ts]
            Log.debug("ffmpeg \(args.joined(separator: " "))")
            try await Shell.run(ffmpeg, args)
            try? FileManager.default.removeItem(atPath: file)
        }
        let dir = (files[0] as NSString).deletingLastPathComponent
        let tsFiles = Downloader.files(in: dir, ext: ".ts")
        try Downloader.combine(files: tsFiles, to: output)
        for f in tsFiles { try? FileManager.default.removeItem(atPath: f) }
    }

    /// Dolby Vision needs libavutil ≥ 57.17 (ffmpeg 5.0).
    public static func ffmpegSupportsDolbyVision(_ ffmpeg: String) async -> Bool {
        guard let result = try? await Shell.run(ffmpeg, ["-version"], echo: false, capture: true),
              let m = result.output.firstMatch(of: /libavutil\s+(\d+)\. +(\d+)\./),
              let major = Int(m.1), let minor = Int(m.2) else { return false }
        return major > 57 || (major == 57 && minor >= 17)
    }

    static func ffmpegChapters(_ points: [ViewPoint]) -> String {
        var s = ";FFMETADATA\n"
        for p in points {
            s += "[CHAPTER]\nTIMEBASE=1/1000\nSTART=\(p.start * 1000)\nEND=\(p.end * 1000)\ntitle=\(p.title)\n\n"
        }
        return s
    }

    static func mp4boxChapters(_ points: [ViewPoint]) -> String {
        points.map { "\(Format.duration($0.start, absolute: true)) \($0.title)" }.joined(separator: "\n") + "\n"
    }
}
