import Foundation

public enum Format {
    public static func fileSize(_ bytes: Double, decimals: Int = 2) -> String {
        let b = max(0, bytes)
        let kb = 1024.0, mb = kb * 1024, gb = mb * 1024
        let f = "%.\(decimals)f"
        if b >= gb { return String(format: f + " GB", b / gb) }
        if b >= mb { return String(format: f + " MB", b / mb) }
        if b >= kb { return String(format: f + " KB", b / kb) }
        return "\(Int(b)) bytes"
    }

    /// `1h02m03s` / `02m03s`, or `01:02:03` when `absolute`.
    public static func duration(_ seconds: Int, absolute: Bool = false) -> String {
        let h = seconds / 3600
        let m = (seconds % 3600) / 60
        let s = seconds % 60
        if absolute { return String(format: "%02d:%02d:%02d", h, m, s) }
        return h == 0 ? String(format: "%02dm%02ds", m, s) : String(format: "%dh%02dm%02ds", h, m, s)
    }

    /// Unix seconds → local time in a Unicode date pattern (e.g. `yyyy-MM-dd_HH-mm-ss`). `0` renders as `null`, like the original.
    public static func timestamp(_ ts: Int64, pattern: String) -> String {
        guard ts != 0 else { return "null" }
        let f = DateFormatter()
        f.locale = Locale(identifier: "en_US_POSIX")
        f.timeZone = .current
        f.dateFormat = pattern
        return f.string(from: Date(timeIntervalSince1970: TimeInterval(ts)))
    }

    /// ISO-8601 UTC with microseconds, the shape ffmpeg accepts for `creation_time`.
    public static func ffmpegCreationTime(_ ts: Int64) -> String {
        let f = DateFormatter()
        f.locale = Locale(identifier: "en_US_POSIX")
        f.timeZone = TimeZone(identifier: "UTC")
        f.dateFormat = "yyyy-MM-dd'T'HH:mm:ss.SSSSSS'Z'"
        return f.string(from: Date(timeIntervalSince1970: TimeInterval(ts)))
    }

    /// Parse `yyyy-MM-dd HH:mm:ss` (bilibili's pgc pub_time) as local time.
    public static func parseDateTime(_ text: String) -> Int64? {
        let f = DateFormatter()
        f.locale = Locale(identifier: "en_US_POSIX")
        f.timeZone = .current
        f.dateFormat = "yyyy-MM-dd HH:mm:ss"
        return f.date(from: text).map { Int64($0.timeIntervalSince1970) }
    }

    private static let invalidFileChars: Set<Character> = {
        var s: Set<Character> = ["\"", "<", ">", "|", ":", "*", "?", "\\", "/"]
        for i in 0..<32 { s.insert(Character(UnicodeScalar(UInt8(i)))) }
        return s
    }()

    /// Replace characters no filesystem likes. Slashes are always replaced; they are directory separators, not name characters.
    public static func validFileName(_ input: String, replacement: String = "_") -> String {
        var out = ""
        out.reserveCapacity(input.count)
        for ch in input { out += invalidFileChars.contains(ch) ? replacement : String(ch) }
        return out
    }

    /// Title cleanup applied to every name component: strip, drop trailing dots, strip again.
    public static func cleanName(_ input: String) -> String {
        var s = validFileName(input).trimmingCharacters(in: .whitespaces)
        while s.hasSuffix(".") { s.removeLast() }
        return s.trimmingCharacters(in: .whitespaces)
    }

    /// `&amp;`, `&lt;`… and numeric entities (`&#26159;`, `&#x4E2D;`) back to characters.
    public static func unescapeEntities(_ text: String) -> String {
        guard text.contains("&") else { return text }
        var s = text
        s = s.replacing(/&#(\d+);/) { m in Int(m.1).flatMap { Unicode.Scalar($0) }.map { String(Character($0)) } ?? String(m.0) }
        s = s.replacing(/&#[xX]([0-9a-fA-F]+);/) { m in Int(m.1, radix: 16).flatMap { Unicode.Scalar($0) }.map { String(Character($0)) } ?? String(m.0) }
        for (entity, char) in [("&quot;", "\""), ("&apos;", "'"), ("&lt;", "<"), ("&gt;", ">"), ("&nbsp;", " "), ("&amp;", "&")] {
            s = s.replacingOccurrences(of: entity, with: char)
        }
        return s
    }

    public static func queryValue(_ name: String, in url: String) -> String {
        guard let q = url.firstIndex(of: "?") else { return "" }
        let query = url[url.index(after: q)...].split(separator: "#").first ?? ""
        for pair in query.split(separator: "&") {
            let kv = pair.split(separator: "=", maxSplits: 1)
            if kv.count == 2, kv[0] == name { return String(kv[1]) }
        }
        return ""
    }

    public static func randomString(_ length: Int) -> String {
        let chars = Array("ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_0123456789")
        return String((0..<length).map { _ in chars.randomElement()! })
    }

    public static func unixSeconds() -> String { String(Int64(Date().timeIntervalSince1970)) }
    public static func unixMillis() -> String { String(Int64(Date().timeIntervalSince1970 * 1000)) }

    /// `yyyyMMddHHmmssfff` — used to seed the TV login fingerprint.
    public static func compactTimestamp() -> String {
        let f = DateFormatter()
        f.locale = Locale(identifier: "en_US_POSIX")
        f.dateFormat = "yyyyMMddHHmmssSSS"
        return f.string(from: Date())
    }

    public static func percentEncode(_ s: String) -> String {
        var allowed = CharacterSet.alphanumerics
        allowed.insert(charactersIn: "-._~")
        return s.addingPercentEncoding(withAllowedCharacters: allowed) ?? s
    }
}
