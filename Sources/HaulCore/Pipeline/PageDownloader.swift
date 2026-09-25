import Foundation

/// Downloads one page: cover, subtitles, tracks, danmaku, then muxes and cleans up.
/// Whatever differs between sites comes from `source`; this flow is the same for all of them.
struct PageDownloader {
    let pipeline: DownloadPipeline
    let source: any MediaSource
    let info: VideoInfo
    let id: MediaID
    let api: APIType
    let selectedPages: [Page]
    let pattern: String

    var options: DownloadOptions { pipeline.options }
    var http: HTTPClient { pipeline.http }
    var style: Style { Log.style }
    var report: Report { pipeline.report }

    /// The user's interactive track choice, kept across retries so they are asked once.
    struct Selection { var video = 0; var audio = 0 }

    func run(_ page: Page) async throws {
        var selection: Selection?
        var attempt = 0
        while true {
            do {
                try await download(page, selection: &selection)
                return
            } catch {
                // A bad option, a missing tool or a cancelled choice will not go away by trying again.
                if error.haulKind != .failed { throw error }
                attempt += 1
                if attempt > 2 { throw error }
                Log.error(error.readableMessage)
                Log.warn("Retrying in 3s (\(attempt)/2)…")
                try await Task.sleep(for: .seconds(3))
            }
        }
    }

    // MARK: per-page flow

    private struct Context {
        var page: Page
        var title: String
        var desc: String
        var pageCount: Int
        var videoPath: String
        var audioPath: String
        var coverPath: String
        var coverURL: String
        var subtitles: [SubtitleInfo]
        var config: DownloadConfig
        var started: Date
    }

    private func download(_ input: Page, selection: inout Selection?) async throws {
        let fm = FileManager.default
        var page = input
        let aid = page.aid
        var title = info.title
        if title.hasSuffix(".") { title += "_fix" }
        if title.hasPrefix(".") { title = "_" + title }

        var ctx = Context(
            page: page,
            title: title,
            desc: (page.desc?.isEmpty == false) ? page.desc! : info.desc,
            pageCount: selectedPages.count,
            videoPath: "\(aid)/\(aid).P\(page.index).\(page.cid).mp4",
            audioPath: "\(aid)/\(aid).P\(page.index).\(page.cid).m4a",
            coverPath: "\(aid)/\(aid).jpg",
            coverURL: info.cover.isEmpty ? (page.cover ?? "") : info.cover,
            subtitles: [],
            config: DownloadConfig(),
            started: Date())

        let chapters = await source.chapters(for: request(page))
        if !chapters.isEmpty { page.points = chapters }

        if !options.onlyShowInfo {
            try fm.createDirectory(atPath: aid, withIntermediateDirectories: true)
            if !options.skipCover && !options.subtitleOnly && !fm.fileExists(atPath: ctx.coverPath) && !options.danmakuOnly && !options.coverOnly {
                // A missing cover (YouTube's guessed maxresdefault.jpg can 404) is not worth the video.
                do {
                    try await Downloader.downloadFile(ctx.coverURL, to: ctx.coverPath, config: DownloadConfig(), http: http)
                } catch {
                    Log.warn("Cover download failed, skipping: \(error.readableMessage)")
                }
            }
            if !options.skipSubtitle && !options.danmakuOnly && !options.coverOnly {
                let subs = await wantedSubtitles(page)
                for s in subs {
                    Log.status("Subtitle  \(s.language)  \(Subtitles.languageInfo(s.language).name)")
                    Log.debug("Downloading \(s.url)")
                    do {
                        try await source.saveSubtitle(s)
                    } catch {
                        // Subtitles are extras; a failed one (YouTube rate-limits timedtext) is skipped, not fatal.
                        Log.warn("Subtitle \(s.language) failed, skipping: \(error.readableMessage)")
                        continue
                    }
                    if options.subtitleOnly && Downloader.fileSize(s.path) > 0 {
                        var out = render(ctx, video: nil, audio: nil)
                        let dir = (out as NSString).deletingLastPathComponent
                        if !dir.isEmpty { try fm.createDirectory(atPath: dir, withIntermediateDirectories: true) }
                        out = FilePattern.changeExtension(out, to: "\(s.language).srt")
                        try Downloader.replaceItem(at: out, with: s.path)
                        report.extraFile(page, out)
                        Log.success("Subtitle  " + Terminal.prettyPath(out))
                    }
                }
                ctx.subtitles = subs
                report.subtitles(page, subs)
            }
            if options.subtitleOnly {
                report.finished(page, status: .downloaded, file: nil)
                removeDirectoryIfEmpty(aid)
                return
            }
        } else if !options.skipSubtitle {
            report.subtitles(page, await wantedSubtitles(page))
        }

        var parsed = try await source.tracks(for: request(page), quality: nil)
        if page.points.isEmpty { page.points = parsed.extraPoints }
        ctx.page = page
        if Log.debugEnabled && !parsed.rawJSON.isEmpty {
            try? parsed.rawJSON.write(toFile: "debug_\(Format.compactTimestamp()).json", atomically: true, encoding: .utf8)
        }

        ctx.config.useAria2c = options.useAria2c
        ctx.config.aria2cPath = pipeline.aria2c ?? "aria2c"
        ctx.config.aria2cArgs = options.aria2cArgs
        ctx.config.multiThread = options.multiThread
        source.configure(&ctx.config, for: request(page))

        if parsed.isDash {
            try await downloadDash(&parsed, ctx: ctx, selection: &selection)
        } else if parsed.isFLV {
            try await downloadFLV(&parsed, ctx: ctx, selection: &selection)
        } else {
            Log.debug(parsed.rawJSON)
            let detail = parsed.rawJSON.count < 100 && !parsed.rawJSON.isEmpty ? ": \(parsed.rawJSON)" : " (--debug shows the response)"
            throw HaulError("Could not find the streams of \"\(page.title)\"" + detail)
        }
    }

    // MARK: DASH

    private func downloadDash(_ parsed: inout ParsedTracks, ctx input: Context, selection: inout Selection?) async throws {
        let fm = FileManager.default
        var ctx = input
        let page = ctx.page
        if source.isAudioOnly(page) {
            if options.videoOnly {
                removeDirectoryIfEmpty(page.aid)
                throw HaulError.input("\"\(page.title)\" is audio only; drop --video-only")
            }
        } else if parsed.video.isEmpty {
            if options.videoOnly { throw HaulError("No video stream found for \"\(page.title)\"") }
            Log.warn("No video stream found")
        }
        if parsed.audio.isEmpty && !parsed.videoHasAudio {
            if options.audioOnly { throw HaulError("No audio stream found for \"\(page.title)\"") }
            Log.warn("No audio stream found")
        }
        // Audio inside the video file is split out at mux time.
        if options.audioOnly && !parsed.videoHasAudio { parsed.video = [] }
        if options.videoOnly { parsed.audio = []; parsed.backgroundAudio = []; parsed.roleAudio = [] }

        parsed.video = pipeline.sortVideo(parsed.video)
        parsed.audio = pipeline.sortAudio(parsed.audio)
        parsed.backgroundAudio = pipeline.sortAudio(parsed.backgroundAudio)
        for i in parsed.roleAudio.indices { parsed.roleAudio[i].tracks = pipeline.sortAudio(parsed.roleAudio[i].tracks) }

        // Show everything when the person is choosing or asked to see it; otherwise focus on the chosen codec.
        let showAll = options.interactive || options.showAll || options.onlyShowInfo
        if selection == nil && (options.videoStream != nil || options.audioStream != nil) {
            selection = try chosenStreams(parsed)
        }
        if options.interactive && selection == nil {
            selection = try askSelection(parsed, pageDuration: page.duration)
        } else if !options.hideStreams {
            let marker = selection ?? Selection()
            Log.lines(UI.streamTable(parsed, selectedVideo: marker.video, selectedAudio: marker.audio, pageDuration: page.duration,
                                     collapse: !showAll, showURLs: options.onlyShowInfo && options.showURLs, style: style))
        }
        report.streams(page, parsed, video: selection?.video ?? 0, audio: selection?.audio ?? 0, includeURLs: options.showURLs)
        if options.onlyShowInfo {
            report.finished(page, status: .listed, file: nil)
            return
        }
        let vIndex = selection?.video ?? 0
        let aIndex = selection?.audio ?? 0
        var video = parsed.video.indices.contains(vIndex) ? parsed.video[vIndex] : nil
        var audio = parsed.audio.indices.contains(aIndex) ? parsed.audio[aIndex] : nil
        let background = parsed.backgroundAudio.indices.contains(aIndex) ? parsed.backgroundAudio[aIndex] : nil

        let container = outputContainer(page, audio: audio)
        let savePath = outputPath(ctx, video: video, audio: audio)
        Log.debug("Output: \(savePath) (template: \(pattern))")
        // An audio-only page keeps its format in the work directory too, so ffmpeg reads an MP3 as one.
        if source.isAudioOnly(page) { ctx.audioPath = FilePattern.changeExtension(ctx.audioPath, to: container.rawValue) }

        if options.downloadDanmaku || options.danmakuOnly {
            if source.hasDanmaku {
                try await source.downloadDanmaku(for: request(page), savePath: savePath, config: ctx.config)
                for ext in ["xml", "ass"] {
                    let path = FilePattern.changeExtension(savePath, to: ext)
                    if fm.fileExists(atPath: path) { report.extraFile(page, path) }
                }
            } else {
                Log.warn("\(info.site.name) has no danmaku")
            }
            if options.danmakuOnly {
                report.finished(page, status: .downloaded, file: nil)
                removeDirectoryIfEmpty(page.aid)
                return
            }
        }

        if options.coverOnly {
            let ext = URL(string: ctx.coverURL)?.pathExtension ?? ""
            let coverOut = FilePattern.changeExtension(savePath, to: ext.isEmpty ? "jpg" : ext)
            try await Downloader.downloadFile(ctx.coverURL, to: coverOut, config: ctx.config, http: http)
            removeDirectoryIfEmpty(page.aid)
            report.finished(page, status: .downloaded, file: coverOut)
            Log.success("Cover  " + Terminal.prettyPath(coverOut))
            return
        }

        source.finalizeURLs(video: &video, audio: &audio)

        if fm.fileExists(atPath: savePath) && Downloader.fileSize(savePath) != 0 {
            Log.status("Exists, skipping  " + Terminal.prettyPath(savePath))
            report.finished(page, status: .skipped, file: savePath, reason: "exists")
            try? fm.removeItem(atPath: ctx.coverPath)
            for s in ctx.subtitles { try? fm.removeItem(atPath: s.path) }
            removeDirectoryIfEmpty(page.aid)
            return
        }

        // MP4Box writes only MP4; an MP3 is tagged by ffmpeg.
        var useMP4Box = options.useMP4Box && container != .mp3
        var materials: [AudioMaterial] = []

        if let v = video {
            var needsMP4Box = false
            if v.quality == Quality.dolbyVisionName && !useMP4Box {
                needsMP4Box = !(await Muxer.ffmpegSupportsDolbyVision(pipeline.ffmpeg))
            }
            if needsMP4Box {
                if pipeline.mp4box != nil {
                    Log.warn("Dolby Vision needs ffmpeg ≥ 5.0; muxing with MP4Box")
                    useMP4Box = true
                } else {
                    Log.warn("Dolby Vision needs ffmpeg ≥ 5.0 or MP4Box (brew install gpac); muxing may fail")
                }
            }
            var c = ctx.config; c.label = "video"
            if source.trustsStatedSizes && v.size > 0 { c.size = Int64(v.size) }
            try await Downloader.downloadTrack(v.baseURL, to: ctx.videoPath, config: c, isVideo: true, http: http)
        }
        if let a = audio {
            var c = ctx.config; c.label = "audio"
            if source.trustsStatedSizes && a.size > 0 { c.size = Int64(a.size) }
            try await Downloader.downloadTrack(a.baseURL, to: ctx.audioPath, config: c, isVideo: false, http: http)
        }
        if let bg = background {
            let path = "\(page.aid)/\(page.aid).\(page.cid).P\(page.index).back_ground.m4a"
            var c = ctx.config; c.label = "bg audio"
            try await Downloader.downloadTrack(bg.baseURL, to: path, config: c, isVideo: false, http: http)
            materials.append(AudioMaterial(title: "Background audio", personName: "", path: path))
        }
        for role in parsed.roleAudio {
            guard let track = role.tracks.indices.contains(aIndex) ? role.tracks[aIndex] : role.tracks.first else { continue }
            var c = ctx.config; c.label = "dub"
            Log.status("Dub  \(role.title)")
            try await Downloader.downloadTrack(track.baseURL, to: role.path, config: c, isVideo: false, http: http)
            materials.append(AudioMaterial(title: role.title, personName: role.personName, path: role.path))
        }

        if options.skipMux {
            for path in [ctx.videoPath, ctx.audioPath] + materials.map(\.path) where fm.fileExists(atPath: path) { report.extraFile(page, path) }
            report.finished(page, status: .downloaded, file: nil, reason: "skip-mux")
            return
        }

        var job = muxJob(ctx, output: savePath, container: container, materials: materials)
        job.videoPath = parsed.video.isEmpty ? "" : ctx.videoPath
        job.audioPath = parsed.audio.isEmpty ? "" : ctx.audioPath
        job.isHEVC = video?.codec == "HEVC"
        Log.status("Muxing  \(useMP4Box ? "MP4Box" : "ffmpeg")\(ctx.subtitles.isEmpty ? "" : "  +subtitles")\(job.chapters.isEmpty ? "" : "  +chapters")")
        let code = try await Muxer.mux(job, useMP4Box: useMP4Box, ffmpeg: pipeline.ffmpeg, mp4box: pipeline.mp4box ?? "mp4box")
        guard code == 0, Downloader.fileSize(savePath) > 0 else {
            throw HaulError("Muxing failed (\(useMP4Box ? "MP4Box" : "ffmpeg") exit code \(code)); --debug shows its output")
        }
        cleanup(ctx, job: job, materials: materials, removeVideo: !parsed.video.isEmpty, removeAudio: !parsed.audio.isEmpty)
        report.finished(page, status: .downloaded, file: savePath)
        Log.lines(UI.summary(path: savePath, size: Downloader.fileSize(savePath), video: options.audioOnly ? nil : video, audio: audio,
                             seconds: Date().timeIntervalSince(ctx.started), style: style))
    }

    // MARK: FLV

    private func downloadFLV(_ parsed: inout ParsedTracks, ctx: Context, selection: inout Selection?) async throws {
        let fm = FileManager.default
        let page = ctx.page
        parsed.video = pipeline.sortVideo(parsed.video)

        if options.interactive && selection == nil {
            let names = parsed.qualities.map(Quality.name)
            let idx = try Picker.choose(count: names.count, prompt: "Choose a quality") { cursor in
                names.enumerated().map { i, n in
                    let marker = i == cursor ? style.cyan("▶") : " "
                    return UI.indent + marker + " " + (i == cursor ? style.boldCyan("\(i)  \(n)") : "\(i)  \(n)")
                }
            }
            selection = Selection(video: 0, audio: 0)
            parsed = try await source.tracks(for: request(page), quality: parsed.qualities[idx])
            parsed.video = pipeline.sortVideo(parsed.video)
        }
        let clips = parsed.clips

        if !options.hideStreams { Log.lines(UI.flvTable(parsed, style: style)) }
        report.streams(page, parsed, video: 0, audio: nil, includeURLs: false)
        if options.onlyShowInfo {
            if options.showURLs { clips.forEach { Log.plain("    " + style.dim($0)) } }
            report.finished(page, status: .listed, file: nil)
            return
        }

        let savePath = outputPath(ctx, video: parsed.video.first, audio: nil)
        if fm.fileExists(atPath: savePath) && Downloader.fileSize(savePath) != 0 {
            Log.status("Exists, skipping  " + Terminal.prettyPath(savePath))
            report.finished(page, status: .skipped, file: savePath, reason: "exists")
            try? fm.removeItem(atPath: ctx.coverPath)
            for s in ctx.subtitles { try? fm.removeItem(atPath: s.path) }
            removeDirectoryIfEmpty(page.aid)
            return
        }

        let width = String(clips.count).count
        for (i, link) in clips.enumerated() {
            let n = String(format: "%0\(width)d", i)
            let clipPath = "\(page.aid)/\(page.aid).P\(page.index).\(page.cid).\(n).mp4"
            var c = ctx.config; c.label = "clip \(i + 1)/\(clips.count)"
            try await Downloader.downloadTrack(link, to: clipPath, config: c, isVideo: true, http: http)
        }
        Log.status("Joining \(clips.count) clips")
        let files = Downloader.files(in: page.aid, ext: ".mp4")
        try await Muxer.mergeFLV(files: files, to: ctx.videoPath, ffmpeg: pipeline.ffmpeg)
        if options.skipMux {
            report.extraFile(page, ctx.videoPath)
            report.finished(page, status: .downloaded, file: nil, reason: "skip-mux")
            return
        }

        var job = muxJob(ctx, output: savePath, container: outputContainer(page, audio: nil), materials: [])
        job.videoPath = ctx.videoPath
        job.audioPath = ""
        Log.status("Muxing  ffmpeg\(ctx.subtitles.isEmpty ? "" : "  +subtitles")")
        let code = try await Muxer.mux(job, useMP4Box: false, ffmpeg: pipeline.ffmpeg, mp4box: pipeline.mp4box ?? "mp4box")
        guard code == 0, Downloader.fileSize(savePath) > 0 else {
            throw HaulError("Muxing failed (ffmpeg exit code \(code)); --debug shows its output")
        }
        cleanup(ctx, job: job, materials: [], removeVideo: !parsed.video.isEmpty, removeAudio: false)
        report.finished(page, status: .downloaded, file: savePath)
        Log.lines(UI.summary(path: savePath, size: Downloader.fileSize(savePath), video: parsed.video.first, audio: nil,
                             seconds: Date().timeIntervalSince(ctx.started), style: style))
    }

    // MARK: pieces

    /// The page's subtitles, less AI ones unless asked for, and only the `--sub-lang` languages when given.
    private func wantedSubtitles(_ page: Page) async -> [SubtitleInfo] {
        Log.debug("Fetching subtitles")
        var subs = await source.subtitles(for: request(page))
        if options.skipAISubtitle {
            let ai = subs.filter { $0.language.hasPrefix("ai-") }.count
            if ai > 0 { Log.debug("Skipping \(ai) AI subtitles") }
            subs = subs.filter { !$0.language.hasPrefix("ai-") }
        }
        let wanted = options.subtitleLanguages.lowercased().split(separator: ",").map { $0.trimmingCharacters(in: .whitespaces) }
            .filter { !$0.isEmpty }
        guard !wanted.isEmpty else { return subs }
        return subs.filter { s in
            let lang = s.language.lowercased().replacingOccurrences(of: "ai-", with: "")
            return wanted.contains { lang == $0 || lang.hasPrefix($0 + "-") }
        }
    }

    private func request(_ page: Page) -> PageRequest {
        PageRequest(page: page, id: id, api: api)
    }

    private func render(_ ctx: Context, video: VideoTrack?, audio: AudioTrack?) -> String {
        FilePattern.render(pattern, FilePattern.Context(title: ctx.title, video: video, audio: audio, page: ctx.page,
                                                       pageCount: info.pages.count, apiType: api.label, pubTime: info.pubTime,
                                                       site: info.site))
    }

    /// `.m4a` for --audio-only, so an existing video does not count as done; an audio-only page (a podcast episode)
    /// keeps its audio's format, `.mp3` or `.m4a`, whatever --audio-only says.
    private func outputContainer(_ page: Page, audio: AudioTrack?) -> Container {
        if source.isAudioOnly(page) { return .audio(codec: audio?.codec ?? "") }
        return options.audioOnly ? .m4a : .mp4
    }

    /// The file this page ends up as.
    private func outputPath(_ ctx: Context, video: VideoTrack?, audio: AudioTrack?) -> String {
        let path = render(ctx, video: video, audio: audio)
        let container = outputContainer(ctx.page, audio: audio)
        return container == .mp4 ? path : FilePattern.changeExtension(path, to: container.rawValue)
    }

    private func muxJob(_ ctx: Context, output: String, container: Container, materials: [AudioMaterial]) -> MuxJob {
        var job = MuxJob()
        job.container = container
        job.pageURL = source.pageURL(ctx.page)
        job.extraAudio = materials
        job.outputPath = output
        job.desc = ctx.desc
        job.title = ctx.title
        job.author = ctx.page.ownerName ?? ""
        job.episodeTitle = (ctx.pageCount > 1 || (info.isBangumi && !info.isBangumiEnd)) ? ctx.page.title : ""
        job.coverPath = FileManager.default.fileExists(atPath: ctx.coverPath) ? ctx.coverPath : ""
        job.language = options.language
        job.subtitles = ctx.subtitles
        job.audioOnly = options.audioOnly || source.isAudioOnly(ctx.page)
        job.videoOnly = options.videoOnly
        job.chapters = ctx.page.points
        job.pubTime = ctx.page.pubTime
        job.simpleMux = options.simpleMux
        if let show = ctx.page.media?.album, !show.isEmpty {
            // A podcast episode: the show is the album, whether one episode or the whole show is downloaded.
            job.title = show
            job.episodeTitle = ctx.page.title
        }
        return job
    }

    private func cleanup(_ ctx: Context, job: MuxJob, materials: [AudioMaterial], removeVideo: Bool, removeAudio: Bool) {
        let fm = FileManager.default
        Log.debug("Removing temporary files")
        if removeVideo { try? fm.removeItem(atPath: ctx.videoPath) }
        if removeAudio { try? fm.removeItem(atPath: ctx.audioPath) }
        if !job.chapters.isEmpty { try? fm.removeItem(atPath: job.chaptersFile) }
        for s in ctx.subtitles { try? fm.removeItem(atPath: s.path) }
        for m in materials { try? fm.removeItem(atPath: m.path) }
        let page = ctx.page
        if let last = selectedPages.last, selectedPages.count == 1 || page.index == last.index || page.aid != last.aid {
            try? fm.removeItem(atPath: ctx.coverPath)
        }
        removeDirectoryIfEmpty(page.aid)
    }

    /// Arrow-key pick of the video track, then the audio track, redrawing the same table with ▶ on the cursor.
    private func askSelection(_ parsed: ParsedTracks, pageDuration: Int) throws -> Selection {
        var s = Selection()
        func table(video: Int?, audio: Int?) -> [String] {
            UI.streamTable(parsed, selectedVideo: video, selectedAudio: audio, pageDuration: pageDuration,
                           collapse: false, showURLs: false, style: style)
        }
        do {
            var onScreen = 0
            if !parsed.video.isEmpty {
                s.video = try Picker.choose(count: parsed.video.count, prompt: "Choose the video stream") { table(video: $0, audio: nil) }
                onScreen = table(video: s.video, audio: nil).count
            }
            if !parsed.audio.isEmpty {
                // Draw over the video frame so both choices live in one table.
                s.audio = try Picker.choose(count: parsed.audio.count, prompt: "Choose the audio stream", overwriting: onScreen) {
                    table(video: parsed.video.isEmpty ? nil : s.video, audio: $0)
                }
            }
        } catch is Picker.Cancelled {
            throw HaulError.cancelled
        }
        return s
    }

    /// `--video-stream` / `--audio-stream`: indexes into the sorted tables, checked against what the page has.
    private func chosenStreams(_ parsed: ParsedTracks) throws -> Selection {
        func check(_ index: Int?, _ count: Int, _ flag: String) throws -> Int {
            guard let index else { return 0 }
            guard count > 0 else { return 0 }
            guard (0..<count).contains(index) else {
                throw HaulError.input("\(flag) \(index) is out of range: this page has \(count) (0-\(count - 1)); `haul info` lists them")
            }
            return index
        }
        return Selection(video: try check(options.videoStream, parsed.video.count, "--video-stream"),
                         audio: try check(options.audioStream, parsed.audio.count, "--audio-stream"))
    }

    private func removeDirectoryIfEmpty(_ dir: String) {
        let fm = FileManager.default
        guard fm.fileExists(atPath: dir), let items = try? fm.contentsOfDirectory(atPath: dir), items.isEmpty else { return }
        try? fm.removeItem(atPath: dir)
    }
}
