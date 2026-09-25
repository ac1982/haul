import Foundation

/// Renders the blocks the tool prints: header, page list, stream table, progress, summary.
/// Pure functions of their inputs and a `Style`, so they are testable and degrade to plain text.
public enum UI {
    public static let indent = "  "

    // MARK: header

    public static func header(_ info: VideoInfo, loggedIn: Bool?, style: Style) -> [String] {
        var meta: [String] = []
        if let owner = info.pages.first(where: { !($0.ownerName ?? "").isEmpty })?.ownerName {
            meta.append("\(info.site.ownerLabel) \(owner)")
        }
        if info.pubTime != 0 { meta.append(Format.timestamp(info.pubTime, pattern: "yyyy-MM-dd")) }
        let durations = info.pages.map(\.duration)
        if info.pages.count == 1, let d = durations.first, d > 0 {
            meta.append(Format.duration(d, absolute: true))
        } else if info.pages.count > 1, durations.allSatisfy({ $0 > 0 }) {
            meta.append("\(Format.duration(durations.reduce(0, +))) total")
        }
        if info.pages.count > 1 {
            meta.append("\(info.pages.count) \(info.site.unit)s")
        }
        if let loggedIn {
            meta.append(loggedIn ? style.green("logged in") : style.yellow("logged out (lower qualities only)"))
        }
        return [
            indent + style.bold(info.title),
            indent + style.dim(meta.joined(separator: "  ·  ")),
        ]
    }

    public static func pageList(_ pages: [Page], showAll: Bool, style: Style) -> [String] {
        guard pages.count > 1 else { return [] }
        let width = String(pages.count).count
        var lines: [String] = []
        let shown = showAll ? pages : Array(pages.prefix(5))
        for p in shown {
            let n = String(repeating: "0", count: max(0, width - String(p.index).count)) + String(p.index)
            let dur = p.duration > 0 ? "  " + style.dim(Format.duration(p.duration, absolute: true)) : ""
            lines.append(indent + style.dim("P\(n)") + "  " + p.title + dur)
        }
        if shown.count < pages.count {
            lines.append(indent + style.dim("…  \(pages.count - shown.count) more (--show-all lists them)"))
        }
        return lines
    }

    public static func pageHeader(index: Int, count: Int, title: String, style: Style) -> String {
        let width = String(count).count
        let n = String(repeating: "0", count: max(0, width - String(index).count)) + String(index)
        return indent + style.bold("[\(n)/\(count)] \(title)")
    }

    // MARK: stream table

    public static func fps(_ raw: String?) -> String {
        guard let raw, let v = Double(raw) else { return "" }
        return v == v.rounded() ? "\(Int(v))fps" : String(format: "%.1ffps", v)
    }

    public static func resolution(_ raw: String?) -> String {
        raw?.replacingOccurrences(of: "x", with: "×") ?? ""
    }

    /// Video and audio tables. `collapse` hides video codecs other than the selected one behind a one-line summary.
    public static func streamTable(_ p: ParsedTracks, selectedVideo: Int?, selectedAudio: Int?, pageDuration: Int,
                                   collapse: Bool, showURLs: Bool, style: Style) -> [String] {
        var out: [String] = []

        if !p.video.isEmpty {
            out.append(indent + style.dim("Video"))
            var rows: [(index: Int, cells: [String], url: String)] = []
            var hidden: [VideoTrack] = []
            let keepCodec = collapse ? selectedVideo.flatMap { p.video.indices.contains($0) ? p.video[$0].codec : nil } : nil
            for (i, v) in p.video.enumerated() {
                if let keepCodec, v.codec != keepCodec { hidden.append(v); continue }
                rows.append((i, [v.quality, resolution(v.resolution), v.codec, fps(v.fps), "\(v.bandwidth) kbps",
                                 Format.fileSize(v.estimatedSize(pageDuration: pageDuration), decimals: 1)], v.baseURL))
            }
            out += table(rows, selected: selectedVideo, rightAligned: [4, 5], showURLs: showURLs, style: style)
            if !hidden.isEmpty {
                let codecs = Array(Set(hidden.map(\.codec))).sorted().joined(separator: " / ")
                out.append(indent + "  " + style.dim("…  \(hidden.count) more \(codecs) streams (-i or --show-all lists them)"))
            }
        }

        if p.videoHasAudio {
            out.append(indent + style.dim("Audio  inside the video file"))
        }
        if !p.audio.isEmpty {
            out.append(indent + style.dim("Audio"))
            // A podcast may not state its bitrate or size.
            let rows = p.audio.enumerated().map { i, a -> (index: Int, cells: [String], url: String) in
                let size = a.estimatedSize(pageDuration: pageDuration)
                return (index: i, cells: [a.codec, a.bandwidth > 0 ? "\(a.bandwidth) kbps" : "-", size > 0 ? Format.fileSize(size, decimals: 1) : "-"],
                        url: a.baseURL)
            }
            out += table(rows, selected: selectedAudio, rightAligned: [1, 2], showURLs: showURLs, style: style)
        }

        if !p.backgroundAudio.isEmpty && !p.roleAudio.isEmpty {
            out.append(indent + style.dim("Background audio"))
            out += table(p.backgroundAudio.enumerated().map { i, a in
                (index: i, cells: [a.codec, "\(a.bandwidth) kbps", Format.fileSize(a.estimatedSize(pageDuration: pageDuration), decimals: 1)], url: a.baseURL)
            }, selected: selectedAudio, rightAligned: [1, 2], showURLs: showURLs, style: style)
            out.append(indent + style.dim("Dubs  \(p.roleAudio.count) roles, \(p.roleAudio[0].tracks.count) streams each: ")
                + p.roleAudio.map(\.title).joined(separator: ", "))
        }
        return out
    }

    /// FLV mode offers whole-file qualities rather than DASH tracks.
    public static func flvTable(_ p: ParsedTracks, style: Style) -> [String] {
        var out = [indent + style.dim("Video (FLV, \(p.clips.count) clips)")]
        let rows = p.video.enumerated().map { i, v -> (index: Int, cells: [String], url: String) in
            let kbps = v.duration > 0 ? Int(v.size / 1024 / Double(v.duration) * 8) : 0
            return (i, [v.quality, v.codec, "\(kbps) kbps", Format.fileSize(v.size, decimals: 1)], "")
        }
        out += table(rows, selected: 0, rightAligned: [2, 3], showURLs: false, style: style)
        return out
    }

    private static func table(_ rows: [(index: Int, cells: [String], url: String)], selected: Int?, rightAligned: Set<Int>,
                              showURLs: Bool, style: Style) -> [String] {
        guard let columns = rows.first?.cells.count else { return [] }
        var widths = [Int](repeating: 0, count: columns)
        for r in rows { for (c, cell) in r.cells.enumerated() { widths[c] = max(widths[c], Terminal.displayWidth(cell)) } }
        let indexWidth = String(rows.map(\.index).max() ?? 0).count
        var out: [String] = []
        for r in rows {
            let isSelected = r.index == selected
            let marker = isSelected ? style.cyan("▶") : " "
            let index = Terminal.pad(String(r.index), indexWidth, right: true)
            let cells = r.cells.enumerated().map { Terminal.pad($1, widths[$0], right: rightAligned.contains($0)) }
            let body = index + "  " + cells.joined(separator: "  ")
            out.append(indent + marker + " " + (isSelected ? style.boldCyan(body) : body))
            if showURLs, !r.url.isEmpty { out.append(indent + "    " + style.dim(r.url)) }
        }
        return out
    }

    // MARK: progress

    public static func bar(_ fraction: Double, width: Int, style: Style) -> String {
        let f = min(1, max(0, fraction))
        let filled = Int((f * Double(width)).rounded(.down))
        return style.cyan(String(repeating: "█", count: filled)) + style.dim(String(repeating: "░", count: width - filled))
    }

    public static func eta(_ seconds: Double) -> String {
        guard seconds.isFinite, seconds >= 0 else { return "--" }
        let s = Int(seconds.rounded())
        if s < 60 { return "\(s)s" }
        if s < 3600 { return String(format: "%dm%02ds", s / 60, s % 60) }
        return String(format: "%dh%02dm", s / 3600, (s % 3600) / 60)
    }

    public static func progressLine(label: String, downloaded: Int64, total: Int64?, bytesPerSecond: Double,
                                    spinner: Character, terminalWidth: Int, style: Style) -> String {
        let name = Terminal.pad(label, 5)
        let speed = bytesPerSecond > 0 ? Format.fileSize(bytesPerSecond, decimals: 1) + "/s" : "--"
        guard let total, total > 0 else {
            return indent + name + "  " + style.dim(String(spinner)) + "  " + Format.fileSize(Double(downloaded), decimals: 1) + "   " + style.dim(speed)
        }
        let fraction = Double(downloaded) / Double(total)
        let percent = Terminal.pad(String(format: "%3.0f%%", fraction * 100), 4, right: true)
        let sizes = "\(Format.fileSize(Double(downloaded), decimals: 1)) / \(Format.fileSize(Double(total), decimals: 1))"
        let remaining = bytesPerSecond > 0 ? eta(Double(total - downloaded) / bytesPerSecond) + " left" : ""
        let tail = "  " + percent + "   " + sizes + "   " + speed + (remaining.isEmpty ? "" : "   " + remaining)
        let barWidth = min(24, max(8, terminalWidth - Terminal.displayWidth(indent + name + "  " + tail) - 2))
        return indent + name + "  " + bar(fraction, width: barWidth, style: style) + tail
    }

    public static func progressDone(label: String, total: Int64, seconds: Double, style: Style) -> String {
        let speed = seconds > 0 ? Format.fileSize(Double(total) / seconds, decimals: 1) + "/s" : "--"
        return indent + Terminal.pad(label, 5) + "  " + style.green("✓") + "  " + Format.fileSize(Double(total), decimals: 1)
            + style.dim("  ·  \(speed)  ·  \(eta(seconds))")
    }

    // MARK: summary

    public static func summary(path: String, size: Int64, video: VideoTrack?, audio: AudioTrack?, seconds: Double, style: Style) -> [String] {
        var parts: [String] = [Format.fileSize(Double(size), decimals: 1)]
        var streams: [String] = []
        if let v = video { streams.append([v.quality, v.codec, fps(v.fps)].filter { !$0.isEmpty }.joined(separator: " ")) }
        if let a = audio { streams.append(a.bandwidth > 0 ? "\(a.codec) \(a.bandwidth) kbps" : a.codec) }
        if !streams.isEmpty { parts.append(streams.joined(separator: " + ")) }
        parts.append(eta(seconds))
        return [
            style.green("✓ ") + style.bold("Done") + "  " + Terminal.prettyPath(path),
            indent + "  " + style.dim(parts.joined(separator: "  ·  ")),
        ]
    }
}
