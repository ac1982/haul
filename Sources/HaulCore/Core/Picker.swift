import Foundation
#if canImport(Glibc)
import Glibc
#elseif canImport(Darwin)
import Darwin
#endif

/// An in-place arrow-key chooser. The caller renders the block for a given cursor; the picker redraws it on every key.
public enum Picker {
    public enum Key: Equatable, Sendable {
        case up, down, first, last, enter, cancel, digit(Int), other
    }

    public struct Cancelled: Error {}

    public static let hint = "↑↓ move   Enter choose   digits jump   q cancel"

    /// Decode one key press from raw terminal bytes.
    public static func parse(_ bytes: [UInt8]) -> Key {
        guard let first = bytes.first else { return .other }
        switch bytes {
        case [0x1B, 0x5B, 0x41], [0x1B, 0x4F, 0x41]: return .up
        case [0x1B, 0x5B, 0x42], [0x1B, 0x4F, 0x42]: return .down
        case [0x1B, 0x5B, 0x48], [0x1B, 0x5B, 0x31, 0x7E]: return .first
        case [0x1B, 0x5B, 0x46], [0x1B, 0x5B, 0x34, 0x7E]: return .last
        default: break
        }
        switch first {
        case 0x0D, 0x0A: return .enter
        case 0x1B, UInt8(ascii: "q"), 0x03: return .cancel
        case UInt8(ascii: "k"): return .up
        case UInt8(ascii: "j"): return .down
        case UInt8(ascii: "g"): return .first
        case UInt8(ascii: "G"): return .last
        case UInt8(ascii: "0")...UInt8(ascii: "9"): return .digit(Int(first - UInt8(ascii: "0")))
        default: return .other
        }
    }

    /// Cursor after a key; typed digits accumulate into `typed` so `1` then `4` reaches row 14.
    public static func move(_ cursor: Int, key: Key, count: Int, typed: inout String) -> Int {
        guard count > 0 else { return 0 }
        switch key {
        case .up: typed = ""; return max(0, cursor - 1)
        case .down: typed = ""; return min(count - 1, cursor + 1)
        case .first: typed = ""; return 0
        case .last: typed = ""; return count - 1
        case .digit(let d):
            let candidate = typed + String(d)
            if let n = Int(candidate), n < count {
                typed = candidate
                if candidate.count >= String(count - 1).count { typed = "" }
                return n
            }
            typed = ""
            return d < count ? d : cursor
        default: return cursor
        }
    }

    /// Run the chooser. `render` draws the whole block (a fixed number of lines) for a cursor position.
    /// `overwriting` is how many lines of a previous frame are already on screen and should be drawn over,
    /// so consecutive choosers share one block. Returns the chosen index; throws `Cancelled` on q / Esc.
    /// Without a terminal on both ends there is nobody to ask: it throws an input error rather than wait on stdin.
    public static func choose(count: Int, initial: Int = 0, prompt: String, overwriting: Int = 0, render: (Int) -> [String]) throws -> Int {
        guard count > 0 else { return 0 }
        guard isatty(STDIN_FILENO) != 0 && Terminal.isTTY else {
            throw HaulError.input("-i needs a terminal. Pick streams with --video-stream / --audio-stream instead "
                + "(the indexes `haul info` lists)")
        }
        // Too tall to redraw in place: ask for the number instead.
        guard render(initial).count + 1 < Terminal.height else {
            if overwriting == 0 { Log.lines(render(initial)) }
            Log.prompt("\(prompt) [\(initial)]: ")
            let n = Int(readLine() ?? "") ?? initial
            return (0..<count).contains(n) ? n : initial
        }

        var cursor = min(max(0, initial), count - 1)
        var typed = ""
        var drawnLines = overwriting
        func draw() {
            if drawnLines > 0 { Log.write("\u{1B}[\(drawnLines)A") }
            let lines = render(cursor) + [UI.indent + Log.style.dim(prompt + "   " + hint)]
            for l in lines { Log.write("\r\u{1B}[K" + l + "\n") }
            drawnLines = lines.count
        }

        let raw = RawMode()
        defer { raw.restore() }
        draw()
        while true {
            for key in readKeys().map(parse) {
                switch key {
                case .enter:
                    // Leave the final state on screen without the hint line.
                    Log.write("\u{1B}[1A\r\u{1B}[K")
                    return cursor
                case .cancel:
                    Log.write("\u{1B}[1A\r\u{1B}[K")
                    throw Cancelled()
                default:
                    let next = move(cursor, key: key, count: count, typed: &typed)
                    if next != cursor { cursor = next; draw() }
                }
            }
        }
    }

    private static func readKeys() -> [[UInt8]] {
        var buf = [UInt8](repeating: 0, count: 64)
        let n = read(STDIN_FILENO, &buf, buf.count)
        return n > 0 ? split(Array(buf[0..<n])) : [[0x1B]]
    }

    /// Split one read into individual key presses: ESC-sequences stay together, everything else is one byte each.
    public static func split(_ bytes: [UInt8]) -> [[UInt8]] {
        var keys: [[UInt8]] = []
        var i = 0
        while i < bytes.count {
            if bytes[i] == 0x1B, i + 1 < bytes.count, bytes[i + 1] == 0x5B || bytes[i + 1] == 0x4F {
                // CSI / SS3: ESC [ ... final byte in 0x40...0x7E
                var j = i + 2
                while j < bytes.count, !(0x40...0x7E).contains(bytes[j]) { j += 1 }
                keys.append(Array(bytes[i...min(j, bytes.count - 1)]))
                i = j + 1
            } else {
                keys.append([bytes[i]])
                i += 1
            }
        }
        return keys
    }
}

/// Puts the terminal into raw mode (no line buffering, no echo) and restores it on exit and on
/// INT / TERM / HUP / TSTP, re-raising the signal so the shell still sees it; raw mode is re-applied on CONT.
final class RawMode {
    nonisolated(unsafe) private static var saved: termios?
    nonisolated(unsafe) private static var raw: termios?
    private static let handled: [Int32] = [SIGINT, SIGTERM, SIGHUP, SIGTSTP]

    init() {
        var t = termios()
        guard tcgetattr(STDIN_FILENO, &t) == 0 else { return }
        Self.saved = t
        var r = t
        r.c_lflag &= ~tcflag_t(ICANON | ECHO)
        Self.raw = r
        tcsetattr(STDIN_FILENO, TCSANOW, &r)
        for sig in Self.handled {
            signal(sig) { sig in
                // Signal context: only async-signal-safe calls here.
                RawMode.restoreSaved()
                if sig != SIGTSTP { _ = write(Log.fd, "\n", 1) }
                signal(sig, SIG_DFL)
                raise(sig)
            }
        }
        signal(SIGCONT) { _ in
            RawMode.reapplyRaw()
            signal(SIGTSTP) { sig in
                RawMode.restoreSaved()
                signal(sig, SIG_DFL)
                raise(sig)
            }
        }
    }

    func restore() {
        Self.restoreSaved()
        for sig in Self.handled { signal(sig, SIG_DFL) }
        signal(SIGCONT, SIG_DFL)
    }

    static func restoreSaved() {
        guard var t = saved else { return }
        tcsetattr(STDIN_FILENO, TCSANOW, &t)
    }

    static func reapplyRaw() {
        guard var r = raw else { return }
        tcsetattr(STDIN_FILENO, TCSANOW, &r)
    }
}
