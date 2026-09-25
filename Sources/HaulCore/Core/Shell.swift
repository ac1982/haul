import Foundation
import Synchronization

/// Runs external tools (ffmpeg, mp4box, aria2c).
public enum Shell {
    /// Search order: current directory, the directory the binary lives in, then `$PATH`.
    public static func findExecutable(_ name: String) -> String? {
        let fm = FileManager.default
        let exeDir = URL(fileURLWithPath: CommandLine.arguments[0]).resolvingSymlinksInPath().deletingLastPathComponent().path
        var dirs = [fm.currentDirectoryPath, exeDir]
        if let path = ProcessInfo.processInfo.environment["PATH"] {
            dirs += path.split(separator: ":").map(String.init)
        }
        for dir in dirs {
            let candidate = (dir as NSString).appendingPathComponent(name)
            if fm.isExecutableFile(atPath: candidate) { return candidate }
        }
        return nil
    }

    /// Resolve a user-supplied path or bare name to an executable path.
    public static func resolve(_ pathOrName: String) -> String? {
        if pathOrName.contains("/") {
            return FileManager.default.isExecutableFile(atPath: pathOrName) ? pathOrName : nil
        }
        return findExecutable(pathOrName)
    }

    public struct Result: Sendable {
        public let status: Int32
        /// stdout, when captured.
        public let output: String
        /// stderr, when captured.
        public let errors: String
    }

    /// Run and wait. stderr lines are echoed through `Log.info` when `echo` is set (ffmpeg talks on stderr).
    @discardableResult
    public static func run(_ executable: String, _ arguments: [String], echo: Bool = true, capture: Bool = false) async throws -> Result {
        let process = Process()
        process.executableURL = URL(fileURLWithPath: executable)
        process.arguments = arguments
        process.standardInput = FileHandle.nullDevice

        let errPipe = Pipe()
        let outPipe = Pipe()
        process.standardError = errPipe
        process.standardOutput = capture ? outPipe : FileHandle.standardOutput

        let collected = Mutex<String>("")
        let collectedErrors = Mutex<String>("")
        let group = DispatchGroup()

        func pump(_ handle: FileHandle, isStderr: Bool, echoLines: Bool) {
            group.enter()
            DispatchQueue.global().async {
                var pending = ""
                while true {
                    let chunk = handle.availableData
                    if chunk.isEmpty { break }
                    let text = String(decoding: chunk, as: UTF8.self)
                    if capture {
                        if isStderr { collectedErrors.withLock { $0 += text } } else { collected.withLock { $0 += text } }
                    }
                    guard echoLines else { continue }
                    pending += text
                    while let nl = pending.firstIndex(where: { $0 == "\n" || $0 == "\r" }) {
                        let line = String(pending[..<nl]).trimmingCharacters(in: .whitespaces)
                        pending = String(pending[pending.index(after: nl)...])
                        if !line.isEmpty { Log.status(line) }
                    }
                }
                let rest = pending.trimmingCharacters(in: .whitespacesAndNewlines)
                if echoLines && !rest.isEmpty { Log.status(rest) }
                group.leave()
            }
        }

        pump(errPipe.fileHandleForReading, isStderr: true, echoLines: echo && !capture)
        if capture { pump(outPipe.fileHandleForReading, isStderr: false, echoLines: false) }

        try process.run()
        await withCheckedContinuation { (cont: CheckedContinuation<Void, Never>) in
            process.terminationHandler = { _ in cont.resume() }
        }
        await withCheckedContinuation { (cont: CheckedContinuation<Void, Never>) in
            group.notify(queue: .global()) { cont.resume() }
        }
        return Result(status: process.terminationStatus, output: collected.withLock { $0 }, errors: collectedErrors.withLock { $0 })
    }
}
