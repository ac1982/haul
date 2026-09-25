import Foundation
import Synchronization
#if canImport(Glibc)
import Glibc
#elseif canImport(Darwin)
import Darwin
#endif

/// Console output. Normal mode prints clean status lines with a leading glyph; `--debug` adds timestamps
/// and shows the internal chatter. Colour follows the terminal (`NO_COLOR` respected).
///
/// Everything goes to stdout, unless `toStderr` is set (`--json`): then stdout carries only the JSON document.
public enum Log {
    private static let debugFlag = Atomic<Bool>(false)
    private static let stderrFlag = Atomic<Bool>(false)
    public static var style: Style { Style(color: Terminal.supportsColor) }

    public static var debugEnabled: Bool {
        get { debugFlag.load(ordering: .relaxed) }
        set { debugFlag.store(newValue, ordering: .relaxed) }
    }

    /// Send every log line to stderr, keeping stdout for data.
    public static var toStderr: Bool {
        get { stderrFlag.load(ordering: .relaxed) }
        set { stderrFlag.store(newValue, ordering: .relaxed) }
    }

    /// The file descriptor log lines go to.
    public static var fd: Int32 { toStderr ? STDERR_FILENO : STDOUT_FILENO }

    public static var isTerminal: Bool { Terminal.isTTY }

    static func write(_ s: String) {
        if toStderr {
            FileHandle.standardError.write(Data(s.utf8))
        } else {
            print(s, terminator: "")
            fflush(stdout)
        }
    }

    /// Data for stdout, whatever `toStderr` says.
    public static func output(_ s: String) {
        print(s)
        fflush(stdout)
    }

    public static func timestamp(_ date: Date = Date()) -> String {
        let c = Calendar.current.dateComponents([.year, .month, .day, .hour, .minute, .second, .nanosecond], from: date)
        let ms = (c.nanosecond ?? 0) / 1_000_000
        return String(format: "[%04d-%02d-%02d %02d:%02d:%02d.%03d]",
                      c.year ?? 0, c.month ?? 0, c.day ?? 0, c.hour ?? 0, c.minute ?? 0, c.second ?? 0, ms)
    }

    private static func emit(_ glyph: String, _ text: String, newline: Bool = true) {
        let prefix = debugEnabled ? style.gray(timestamp()) + " " : ""
        write(prefix + glyph + text + (newline ? "\n" : ""))
    }

    /// A neutral line.
    public static func info(_ text: String, newline: Bool = true) { emit("  ", text, newline: newline) }
    /// Low-key activity (loading a cookie, a subtitle found…).
    public static func status(_ text: String) { emit("  ", style.dim("· " + text)) }
    public static func success(_ text: String) { emit(style.green("✓ "), text) }
    public static func warn(_ text: String) { emit(style.yellow("! "), style.yellow(text)) }
    public static func error(_ text: String) { emit(style.red("✗ "), style.red(text)) }
    public static func highlight(_ text: String) { emit("  ", style.cyan(text)) }
    public static func title(_ text: String) { emit("  ", style.bold(text)) }
    public static func banner(_ text: String) { write(style.bold(text) + "\n") }
    public static func plain(_ text: String = "") { write(text + "\n") }
    public static func lines(_ lines: [String]) { for l in lines { plain(l) } }
    /// A question; the cursor stays on the line.
    public static func prompt(_ text: String) { emit("  ", style.cyan(text), newline: false) }

    public static func debug(_ text: @autoclosure () -> String) {
        guard debugEnabled else { return }
        write(style.gray(timestamp() + " " + text()) + "\n")
    }
}
