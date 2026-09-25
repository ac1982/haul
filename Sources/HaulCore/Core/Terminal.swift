import Foundation
#if canImport(Glibc)
import Glibc
#elseif canImport(Darwin)
import Darwin
#endif

/// ANSI styling that switches itself off when the output is not a colour terminal.
public struct Style: Sendable {
    public var color: Bool

    public init(color: Bool) { self.color = color }

    public static let plain = Style(color: false)

    func wrap(_ s: String, _ code: String) -> String { color ? "\u{1B}[\(code)m\(s)\u{1B}[0m" : s }
    public func bold(_ s: String) -> String { wrap(s, "1") }
    public func dim(_ s: String) -> String { wrap(s, "2") }
    public func red(_ s: String) -> String { wrap(s, "31") }
    public func green(_ s: String) -> String { wrap(s, "32") }
    public func yellow(_ s: String) -> String { wrap(s, "33") }
    public func magenta(_ s: String) -> String { wrap(s, "35") }
    public func cyan(_ s: String) -> String { wrap(s, "36") }
    public func gray(_ s: String) -> String { wrap(s, "90") }
    public func boldCyan(_ s: String) -> String { wrap(s, "1;36") }
}

public enum Terminal {
    /// Whether log output (stdout, or stderr under `--json`) is a terminal.
    public static var isTTY: Bool { isatty(Log.fd) != 0 }

    public static var supportsColor: Bool {
        let env = ProcessInfo.processInfo.environment
        return isTTY && env["NO_COLOR"] == nil && env["TERM"] != "dumb"
    }

    /// Rows of the attached terminal, 24 when unknown.
    public static var height: Int {
        var w = winsize()
        if ioctl(Log.fd, TIOCGWINSZ, &w) == 0, w.ws_row > 0 { return Int(w.ws_row) }
        if let r = ProcessInfo.processInfo.environment["LINES"], let n = Int(r), n > 0 { return n }
        return 24
    }

    /// Columns of the attached terminal, 80 when unknown.
    public static var width: Int {
        var w = winsize()
        if ioctl(Log.fd, TIOCGWINSZ, &w) == 0, w.ws_col > 0 { return Int(w.ws_col) }
        if let c = ProcessInfo.processInfo.environment["COLUMNS"], let n = Int(c), n > 0 { return n }
        return 80
    }

    /// Number of terminal cells a string occupies: CJK and emoji count double, combining marks zero.
    public static func displayWidth(_ s: String) -> Int {
        s.unicodeScalars.reduce(0) { $0 + cellWidth($1) }
    }

    static func cellWidth(_ u: Unicode.Scalar) -> Int {
        let v = u.value
        if v < 0x20 || (0x7F...0x9F).contains(v) || v == 0x200B || v == 0x200D { return 0 }
        switch u.properties.generalCategory {
        case .nonspacingMark, .enclosingMark, .format: return 0
        default: break
        }
        let wide: [ClosedRange<UInt32>] = [
            0x1100...0x115F, 0x2E80...0x303E, 0x3041...0x33FF, 0x3400...0x4DBF, 0x4E00...0x9FFF,
            0xA000...0xA4CF, 0xAC00...0xD7A3, 0xF900...0xFAFF, 0xFE30...0xFE4F, 0xFF00...0xFF60,
            0xFFE0...0xFFE6, 0x1F300...0x1F64F, 0x1F900...0x1F9FF, 0x20000...0x3FFFD,
        ]
        return wide.contains { $0.contains(v) } ? 2 : 1
    }

    /// Pad with spaces to `width` cells (measured, not counted), left-aligned unless `right`.
    public static func pad(_ s: String, _ width: Int, right: Bool = false) -> String {
        let w = displayWidth(s)
        guard w < width else { return s }
        let fill = String(repeating: " ", count: width - w)
        return right ? fill + s : s + fill
    }

    /// Cut to `width` cells with an ellipsis.
    public static func truncate(_ s: String, to width: Int) -> String {
        guard displayWidth(s) > width, width > 1 else { return s }
        var out = ""
        var used = 0
        for u in s.unicodeScalars {
            let w = cellWidth(u)
            if used + w > width - 1 { break }
            out.unicodeScalars.append(u)
            used += w
        }
        return out + "…"
    }

    /// Absolute path with the home directory shown as `~`.
    public static func prettyPath(_ path: String) -> String {
        let absolute = path.hasPrefix("/") ? path : (FileManager.default.currentDirectoryPath as NSString).appendingPathComponent(path)
        let standard = (absolute as NSString).standardizingPath
        let home = FileManager.default.homeDirectoryForCurrentUser.path
        if standard == home { return "~" }
        if standard.hasPrefix(home + "/") { return "~" + standard.dropFirst(home.count) }
        return standard
    }
}
