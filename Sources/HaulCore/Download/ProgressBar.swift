import Foundation
import Synchronization

/// One progress line per transfer: bar, sizes, rate, ETA. Redraws in place on a terminal; on a pipe it prints
/// nothing until `finish`, which leaves one summary line either way.
public final class ProgressBar: Sendable {
    private struct State {
        var downloaded: Int64 = 0
        var total: Int64?
        var samples: [(time: Date, bytes: Int64)] = []
        var finished = false
        var frame = 0
    }

    private static let spinner: [Character] = Array("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")
    private let label: String
    private let started = Date()
    private let state: Mutex<State>
    private let ticker = Mutex<Task<Void, Never>?>(nil)

    public init(label: String, total: Int64?) {
        self.label = label
        state = Mutex(State(total: total))
        guard Terminal.isTTY else { return }
        let task = Task.detached(priority: .utility) { [self] in
            while !Task.isCancelled {
                try? await Task.sleep(for: .milliseconds(125))
                let line: String? = self.state.withLock { st in
                    guard !st.finished else { return nil }
                    st.frame += 1
                    return UI.progressLine(label: self.label, downloaded: st.downloaded, total: st.total,
                                           bytesPerSecond: Self.speed(st.samples), spinner: Self.spinner[st.frame % Self.spinner.count],
                                           terminalWidth: Terminal.width, style: Log.style)
                }
                guard let line else { break }
                Log.write("\r" + line + "\u{1B}[K")
            }
        }
        ticker.withLock { $0 = task }
    }

    public func report(downloaded: Int64, total: Int64?) {
        let now = Date()
        state.withLock { st in
            st.downloaded = downloaded
            if let total, total > 0 { st.total = total }
            st.samples.append((now, downloaded))
            // Keep a three-second window for the rate.
            while st.samples.count > 2, now.timeIntervalSince(st.samples[1].time) > 3 { st.samples.removeFirst() }
        }
    }

    /// Bytes per second over the recent window.
    private static func speed(_ samples: [(time: Date, bytes: Int64)]) -> Double {
        guard let first = samples.first, let last = samples.last, last.time > first.time else { return 0 }
        return Double(last.bytes - first.bytes) / last.time.timeIntervalSince(first.time)
    }

    public func finish(success: Bool) {
        ticker.withLock { $0?.cancel() }
        let downloaded = state.withLock { st -> Int64 in
            st.finished = true
            return st.downloaded
        }
        if Terminal.isTTY { Log.write("\r\u{1B}[K") }
        if success {
            Log.plain(UI.progressDone(label: label, total: downloaded, seconds: Date().timeIntervalSince(started), style: Log.style))
        }
    }
}
