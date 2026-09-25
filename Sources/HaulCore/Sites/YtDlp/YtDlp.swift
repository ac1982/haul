import Foundation

/// `yt-dlp -J` → our models. YouTube's player challenges need yt-dlp's JavaScript solver, and X changes its API
/// often; letting yt-dlp find the streams keeps those fixes a `brew upgrade` away.
public enum YtDlp {
    /// Run `yt-dlp -J` on the link and turn its answer into our models.
    public static func fetch(_ url: String, site: Site, ytdlp: String) async throws -> VideoInfo {
        let args = ["-J", "--no-playlist", "--no-progress", "--", url]
        Log.debug("\(ytdlp) \(args.joined(separator: " "))")
        let result = try await Shell.run(ytdlp, args, echo: false, capture: true)
        for line in result.errors.split(separator: "\n") { Log.debug("yt-dlp: \(line)") }
        guard result.status == 0, !result.output.isEmpty else {
            let reason = result.errors.split(separator: "\n").last { $0.hasPrefix("ERROR") }
                .map { $0.replacing(/^ERROR:\s*(\[[\w:]+\]\s*\S+:\s*)?/, with: "") } ?? "exit code \(result.status)"
            throw HaulError("yt-dlp failed: \(reason)")
        }
        return try parse(JSON.parse(result.output), site: site)
    }

    /// A post with several videos (X) comes back as a playlist; each entry becomes a page.
    static func parse(_ json: JSON, site: Site) throws -> VideoInfo {
        let entries = json["_type"].string == "playlist" ? json["entries"].array.filter(\.isObject) : [json]
        let pages = try entries.enumerated().map { i, entry in
            try page(entry, index: i + 1, title: entries.count > 1 ? "Video \(i + 1)" : nil, parent: json, site: site)
        }
        guard let first = pages.first else { throw HaulError("yt-dlp found no downloadable video") }
        var info = VideoInfo(title: title(json["title"].string ?? first.title, site: site), desc: json["description"].stringValue,
                             // Several videos have several covers; each page brings its own.
                             cover: pages.count == 1 ? first.cover ?? "" : "", pubTime: first.pubTime, pages: pages)
        info.site = site
        return info
    }

    private static func page(_ json: JSON, index: Int, title: String?, parent: JSON, site: Site) throws -> Page {
        let id = try json.get("id").stringValue
        let duration = Int(json["duration"].double ?? 0)
        var pubTime = json["timestamp"].int64 ?? parent["timestamp"].int64 ?? 0
        if pubTime == 0, let day = json["upload_date"].string { pubTime = parseDay(day) }
        let (tracks, headers) = try streams(json["formats"], duration: duration)

        var page = Page(index: index, aid: id, cid: id, epid: "", title: title ?? Self.title(json["title"].stringValue, site: site), duration: duration,
                        resolution: "", pubTime: pubTime, cover: cover(json, id: id),
                        desc: json["description"].string ?? parent["description"].string,
                        ownerName: json["channel"].string ?? json["uploader"].string ?? parent["uploader"].string,
                        ownerMid: json["channel_id"].string ?? json["uploader_id"].string)
        page.points = json["chapters"].array.map {
            ViewPoint(title: $0["title"].stringValue, start: Int($0["start_time"].double ?? 0), end: Int($0["end_time"].double ?? 0))
        }
        page.media = PageMedia(tracks: tracks, subtitles: subtitles(json, id: id), headers: headers,
                                   webpageURL: json["webpage_url"].string ?? parent["webpage_url"].stringValue)
        return page
    }

    /// yt-dlp shortens X titles (the post text) with a trailing `...`; a file name is better off without it.
    /// Other sites' titles are the uploader's own and stay as they are.
    static func title(_ text: String, site: Site) -> String {
        site == .x && text.hasSuffix("...") ? String(text.dropLast(3)).trimmingCharacters(in: .whitespaces) : text
    }

    /// DASH video and audio when the site has them (YouTube); otherwise whole files with the audio inside (X).
    static func streams(_ formats: JSON, duration: Int) throws -> (ParsedTracks, [String: String]) {
        var tracks = ParsedTracks()
        var progressive: [VideoTrack] = []
        var headers: [String: String] = [:]
        var audio: [(track: AudioTrack, preference: Int)] = []
        for f in formats.array {
            // Plain HTTPS only: no HLS, storyboards, DRC or DRM variants.
            guard f["protocol"].string == "https", let url = f["url"].string, f["has_drm"].bool != true,
                  !f["format_id"].stringValue.contains("drc") else { continue }
            let vcodec = f["vcodec"].string
            let acodec = f["acodec"].string
            // Only exact sizes: X's `filesize_approx` is a guess from the bitrate.
            let size = Double(f["filesize"].int64 ?? 0)
            if vcodec == "none" && acodec != nil && acodec != "none" {
                audio.append((AudioTrack(
                    id: f["format_id"].stringValue,
                    quality: f["format_note"].stringValue,
                    baseURL: url,
                    codec: audioCodec(acodec!),
                    bandwidth: Int64((f["abr"].double ?? f["tbr"].double ?? 0).rounded()),
                    duration: duration,
                    size: size), f["language_preference"].int ?? 0))
            } else if vcodec != "none", f["width"].int != nil {
                let track = videoTrack(f, url: url, vcodec: vcodec, duration: duration, size: size)
                if acodec == "none" { tracks.video.append(track) } else { progressive.append(track) }
            } else {
                continue
            }
            if headers.isEmpty { headers = f["http_headers"].object.compactMapValues(\.string) }
        }
        if tracks.video.isEmpty && audio.isEmpty && !progressive.isEmpty {
            tracks.video = progressive
            tracks.videoHasAudio = true
        }
        // Dubbed videos carry an audio track per language; keep the original.
        let best = audio.map(\.preference).max() ?? 0
        tracks.audio = audio.filter { $0.preference == best }.map(\.track)
        guard !tracks.video.isEmpty || !tracks.audio.isEmpty else { throw HaulError("yt-dlp found no downloadable streams") }
        return (tracks, headers)
    }

    private static func videoTrack(_ f: JSON, url: String, vcodec: String?, duration: Int, size: Double) -> VideoTrack {
        let w = f["width"].int ?? 0, h = f["height"].int ?? 0
        let fps = f["fps"].double ?? 0
        let note = f["format_note"].stringValue
        return VideoTrack(
            // Sorts like bilibili's quality ids: higher is better.
            id: String(min(w, h) * 1000 + Int(fps.rounded())),
            quality: note.isEmpty ? "\(min(w, h))p" : note,
            baseURL: url,
            resolution: w > 0 ? "\(w)x\(h)" : nil,
            fps: fps > 0 ? String(Int(fps.rounded())) : nil,
            // X leaves the codec out; its file paths name it (`/vid/avc1/…`).
            codec: videoCodec(vcodec ?? url),
            bandwidth: Int64((f["vbr"].double ?? f["tbr"].double ?? 0).rounded()),
            duration: duration,
            size: size)
    }

    /// Uploaded subtitles, plus the auto-generated track in the spoken language (as `ai-…`, so --skip-ai-subtitle applies).
    /// Machine translations and the transcripts of auto-dubbed audio are left out.
    static func subtitles(_ json: JSON, id: String) -> [SubtitleInfo] {
        func json3(_ entries: JSON) -> String? {
            entries.array.first { $0["ext"].string == "json3" }?["url"].string
        }
        func base(_ lang: String) -> String { lang.components(separatedBy: "-")[0] }
        let spoken = json["language"].string.map(base)
        var out: [SubtitleInfo] = []
        for (lang, entries) in json["subtitles"].object.sorted(by: { $0.key < $1.key }) where lang != "live_chat" {
            if let url = json3(entries) { out.append(SubtitleInfo(language: lang, url: url, path: "\(id)/\(id).\(lang).srt")) }
        }
        for (key, entries) in json["automatic_captions"].object.sorted(by: { $0.key < $1.key }) where key.hasSuffix("-orig") {
            let lang = String(key.dropLast("-orig".count))
            if let spoken, base(lang) != spoken { continue }
            if let url = json3(entries) { out.append(SubtitleInfo(language: "ai-\(lang)", url: url, path: "\(id)/\(id).ai-\(lang).srt")) }
        }
        return out
    }

    /// The largest JPEG thumbnail; the muxers cannot embed WebP.
    static func cover(_ json: JSON, id: String) -> String? {
        let jpegs = json["thumbnails"].array.filter { ($0["url"].string ?? "").contains(".jpg") }
        let best = jpegs.max { ($0["preference"].int ?? -100, $0["width"].int ?? 0) < ($1["preference"].int ?? -100, $1["width"].int ?? 0) }
        return best?["url"].string ?? json["thumbnail"].string
    }

    static func videoCodec(_ vcodec: String) -> String {
        let c = vcodec.lowercased()
        if c.hasPrefix("avc") || c.contains("/avc1/") { return "AVC" }
        if c.hasPrefix("hev") || c.hasPrefix("hvc") || c.contains("/hevc/") { return "HEVC" }
        if c.hasPrefix("av01") { return "AV1" }
        if c.hasPrefix("vp9") || c.hasPrefix("vp09") { return "VP9" }
        return c.hasPrefix("http") ? "MP4" : vcodec.uppercased()
    }

    static func audioCodec(_ acodec: String) -> String {
        let c = acodec.lowercased()
        if c.hasPrefix("mp4a") { return "M4A" }
        if c.hasPrefix("opus") { return "OPUS" }
        if c.hasPrefix("ec-3") { return "E-AC-3" }
        if c.hasPrefix("ac-3") { return "AC-3" }
        return acodec.uppercased()
    }

    /// `20260910` → midnight UTC that day.
    static func parseDay(_ text: String) -> Int64 {
        let f = DateFormatter()
        f.locale = Locale(identifier: "en_US_POSIX")
        f.timeZone = TimeZone(identifier: "UTC")
        f.dateFormat = "yyyyMMdd"
        return f.date(from: text).map { Int64($0.timeIntervalSince1970) } ?? 0
    }

    // MARK: captions

    /// YouTube `json3` captions → SRT. Auto-generated tracks interleave newline-only events; those are dropped.
    static func srt(fromJSON3 json: JSON) -> String {
        var out = ""
        var n = 0
        for e in json["events"].array {
            let text = e["segs"].array.map { $0["utf8"].stringValue }.joined().trimmingCharacters(in: .whitespacesAndNewlines)
            guard !text.isEmpty, let start = e["tStartMs"].double else { continue }
            let end = start + (e["dDurationMs"].double ?? 0)
            n += 1
            out += "\(n)\n\(Subtitles.srtTime(start / 1000)) --> \(Subtitles.srtTime(end / 1000))\n\(text)\n\n"
        }
        return out
    }
}
