import Foundation
import Synchronization
#if canImport(FoundationNetworking)
import FoundationNetworking
#endif

public struct DownloadConfig: Sendable {
    public var useAria2c = false
    public var aria2cPath = "aria2c"
    public var aria2cArgs = ""
    public var forceHTTP = false
    public var multiThread = false
    public var cookie = ""
    /// Request headers that replace the bilibili defaults (Referer, User-Agent).
    public var headers: [String: String] = [:]
    /// For servers that only answer closed byte ranges (googlevideo): always fetch in ranges of at most this many bytes.
    public var maxRange: Int64?
    /// The track size when the API stated it, saving a probe.
    public var size: Int64?
    /// Progress label (video / audio …); nil downloads silently.
    public var label: String?

    public init() {}
}

/// HTTP range downloads: single stream with resume, or segmented parallel download for CDN media.
public enum Downloader {
    static let clipSize: Int64 = 20 * 1024 * 1024
    static let maxParallel = 16

    struct RangeNotSupported: Error {}

    // MARK: public entry points

    public static func downloadFile(_ url: String, to path: String, config: DownloadConfig, http: HTTPClient = .shared) async throws {
        guard !url.isEmpty else { return }
        let url = config.forceHTTP ? forceHTTP(url) : url
        Log.debug("Start downloading: \(url)")
        let dir = (path as NSString).deletingLastPathComponent
        if !dir.isEmpty { try FileManager.default.createDirectory(atPath: dir, withIntermediateDirectories: true) }
        if config.useAria2c {
            try await Aria2c.download(url, to: path, config: config)
            return
        }
        let tmp = (dir as NSString).appendingPathComponent(((path as NSString).lastPathComponent as NSString).deletingPathExtension + ".tmp")
        var attempt = 0
        while true {
            attempt += 1
            let bar = config.label.map { ProgressBar(label: $0, total: nil) }
            do {
                try await rangeDownload(url, to: tmp, from: 0, to: nil, cookie: config.cookie, headers: config.headers, http: http) { done, total in
                    bar?.report(downloaded: done, total: total)
                }
                bar?.finish(success: true)
                try replaceItem(at: path, with: tmp)
                return
            } catch {
                bar?.finish(success: false)
                if attempt >= 3 { throw error }
                Log.debug("Download failed, retrying (\(attempt)/3): \(error.readableMessage)")
            }
        }
    }

    /// A media track: segmented when allowed (always, for range-only servers), otherwise one stream.
    public static func downloadTrack(_ url: String, to path: String, config: DownloadConfig, isVideo: Bool, http: HTTPClient = .shared) async throws {
        var config = config
        if config.multiThread && url.contains("-cmcc-") {
            Log.debug("cmcc CDN host; single connection")
            config.forceHTTP = false
            config.multiThread = false
        }
        guard config.multiThread || config.maxRange != nil else {
            try await downloadFile(url, to: path, config: config, http: http)
            return
        }
        do {
            try await multiThreadDownload(url, to: path, config: config, http: http)
        } catch is RangeNotSupported where config.maxRange == nil {
            Log.warn("The server refuses byte ranges; downloading over one connection")
            let dir = (path as NSString).deletingLastPathComponent
            for f in files(in: dir, ext: ".vclip") + files(in: dir, ext: ".aclip") { try? FileManager.default.removeItem(atPath: f) }
            try await downloadFile(url, to: path, config: config, http: http)
            return
        }
        let dir = (path as NSString).deletingLastPathComponent
        let ext = isVideo ? ".vclip" : ".aclip"
        Log.debug("Joining the \(isVideo ? "video" : "audio") clips")
        try combine(files: files(in: dir, ext: ext), to: path)
        for f in files(in: dir, ext: ".vclip") + files(in: dir, ext: ".aclip") { try? FileManager.default.removeItem(atPath: f) }
    }

    // MARK: segmented

    static func multiThreadDownload(_ rawURL: String, to path: String, config: DownloadConfig, http: HTTPClient) async throws {
        let url = config.forceHTTP ? forceHTTP(rawURL) : rawURL
        Log.debug("Start downloading: \(url)")
        let dir = (path as NSString).deletingLastPathComponent
        if !dir.isEmpty { try FileManager.default.createDirectory(atPath: dir, withIntermediateDirectories: true) }
        if config.useAria2c {
            try await Aria2c.download(url, to: path, config: config)
            return
        }
        let size: Int64
        if let known = config.size, known > 0 {
            size = known
        } else {
            size = try await contentLength(url, config: config, http: http)
        }
        Log.debug("Size: \(size) bytes")
        guard size > 0 else {
            if config.maxRange != nil { throw HaulError("The server did not state the file size: \(url)") }
            throw RangeNotSupported()  // falls back to one stream
        }
        if let attrs = try? FileManager.default.attributesOfItem(atPath: path), (attrs[.size] as? Int64) == size {
            Log.debug("Already downloaded, skipping")
            return
        }
        let clips = config.maxRange.map { makeClips(size, clipSize: $0 - 1, closed: true) } ?? makeClips(size)
        // Range-only servers get one range at a time: googlevideo cuts off or refuses parallel ranges.
        let parallel = config.multiThread && config.maxRange == nil ? maxParallel : 1
        Log.debug("Clips: \(clips.count)")
        let progress = ClipProgress()
        let bar = config.label.map { ProgressBar(label: $0, total: size) }
        var succeeded = false
        defer { bar?.finish(success: succeeded) }
        let base = ((path as NSString).lastPathComponent as NSString).deletingPathExtension
        let ext = path.hasSuffix(".mp4") ? ".vclip" : ".aclip"

        try await withThrowingTaskGroup(of: Void.self) { group in
            var pending = clips.makeIterator()
            var running = 0
            func launch(_ clip: Clip) {
                group.addTask {
                    let tmp = (dir as NSString).appendingPathComponent(String(format: "%05d_", clip.index) + base + ext)
                    var attempt = 0
                    while true {
                        attempt += 1
                        do {
                            try await rangeDownload(url, to: tmp, from: clip.from, to: clip.to, cookie: config.cookie, headers: config.headers,
                                                    http: http, strictRange: true) { done, _ in
                                let total = progress.update(clip: clip.index, bytes: done)
                                bar?.report(downloaded: total, total: size)
                            }
                            return
                        } catch is RangeNotSupported {
                            if attempt >= 3 { throw RangeNotSupported() }
                        } catch {
                            if attempt >= 3 { throw HaulError("Clip \(clip.index) failed: \(error.readableMessage)") }
                        }
                    }
                }
            }
            while running < parallel, let c = pending.next() { launch(c); running += 1 }
            while running > 0 {
                try await group.next()
                running -= 1
                if let c = pending.next() { launch(c); running += 1 }
            }
        }
        succeeded = true
    }

    struct Clip { let index: Int; let from: Int64; let to: Int64? }

    /// Bytes received per clip, summed for the progress bar. A class so task closures capture a reference, not the lock itself.
    final class ClipProgress: Sendable {
        private let bytes = Mutex<[Int: Int64]>([:])
        func update(clip: Int, bytes done: Int64) -> Int64 {
            bytes.withLock { $0[clip] = done; return $0.values.reduce(0, +) }
        }
    }

    /// Clips of `clipSize + 1` bytes (`bytes=a-(a+clipSize)`); the last one reads to EOF, or ends at the last byte when `closed`.
    static func makeClips(_ fileSize: Int64, clipSize: Int64 = clipSize, closed: Bool = false) -> [Clip] {
        var clips: [Clip] = []
        var cursor: Int64 = 0
        while cursor < fileSize {
            let end = cursor + clipSize
            if end >= fileSize - 1 {
                clips.append(Clip(index: clips.count, from: cursor, to: closed ? fileSize - 1 : nil))
                break
            }
            clips.append(Clip(index: clips.count, from: cursor, to: end))
            cursor = end + 1
        }
        return clips
    }

    // MARK: single range stream

    private static func mediaRequest(_ url: String, cookie: String, headers: [String: String] = [:]) throws -> URLRequest {
        guard let u = URL(string: url) else { throw HaulError("Invalid URL: \(url)") }
        var req = URLRequest(url: u)
        if headers.isEmpty {
            if !url.contains("platform=android_tv_yst") && !url.contains("platform=android") {
                req.setValue("https://www.bilibili.com", forHTTPHeaderField: "Referer")
            }
            req.setValue("Mozilla/5.0", forHTTPHeaderField: "User-Agent")
        }
        for (k, v) in headers { req.setValue(v, forHTTPHeaderField: k) }
        if !cookie.isEmpty { req.setValue(cookie, forHTTPHeaderField: "Cookie") }
        return req
    }

    /// Range-only servers are asked for the first byte and report the size in `Content-Range`.
    static func contentLength(_ url: String, config: DownloadConfig, http: HTTPClient) async throws -> Int64 {
        var req = try mediaRequest(url, cookie: config.cookie, headers: config.headers)
        if config.maxRange != nil { req.setValue("bytes=0-0", forHTTPHeaderField: "Range") }
        let resp = try await http.stream(req)
        resp.cancel()
        guard (200..<300).contains(resp.head.statusCode) else { throw HaulError("HTTP \(resp.head.statusCode): \(url)") }
        // The Content-Length of a one-byte probe is 1, not the size.
        if config.maxRange != nil { return resp.head.rangeTotal ?? 0 }
        return resp.head.contentLength ?? 0
    }

    /// Download `[from, to]` of `url` into `tmp`, resuming whatever is already there.
    static func rangeDownload(_ url: String, to tmp: String, from: Int64, to: Int64?, cookie: String, headers: [String: String] = [:], http: HTTPClient,
                              strictRange: Bool = false, onProgress: @Sendable (Int64, Int64?) -> Void) async throws {
        let fm = FileManager.default
        if !fm.fileExists(atPath: tmp) { fm.createFile(atPath: tmp, contents: nil) }
        let handle = try FileHandle(forWritingTo: URL(fileURLWithPath: tmp))
        defer { try? handle.close() }
        var position = Int64(try handle.seekToEnd())

        if let to, to > 0, position == to - from + 1 {
            onProgress(position, position)
            return
        }
        var downloaded = from + position

        var req = try mediaRequest(url, cookie: cookie, headers: headers)
        req.setValue("bytes=\(downloaded)-\(to.map(String.init) ?? "")", forHTTPHeaderField: "Range")

        let resp = try await http.stream(req)
        // Leaving early for any reason (bad status, write error, range refused) must stop the transfer.
        defer { resp.cancel() }
        guard (200..<300).contains(resp.head.statusCode) else {
            throw HaulError("HTTP \(resp.head.statusCode): \(url)")
        }
        if resp.head.statusCode == 200 {
            // Server ignored the range: start over.
            if strictRange && (downloaded > 0 || to != nil) { throw RangeNotSupported() }
            downloaded = 0
            position = 0
            try handle.truncate(atOffset: 0)
        }
        let total: Int64? = resp.head.contentLength.map { downloaded + $0 }
        var received: Int64 = 0
        for try await chunk in resp.body {
            try handle.write(contentsOf: chunk)
            received += Int64(chunk.count)
            downloaded += Int64(chunk.count)
            onProgress(downloaded - from, total.map { $0 - from })
        }
        try handle.synchronize()
        if let expected = resp.head.contentLength, expected != received {
            throw HaulError("Incomplete download (\(received)/\(expected) bytes)")
        }
    }

    // MARK: files

    /// CDN media over plain HTTP; the mcdn hosts only speak what they advertise.
    static func forceHTTP(_ url: String) -> String {
        if url.contains(".mcdn.bilivideo.cn:") {
            Log.debug("Keeping https for *.mcdn.bilivideo.cn:port")
            return url
        }
        Log.debug("Switching https to http")
        return url.replacingOccurrences(of: "https:", with: "http:")
    }

    public static func files(in dir: String, ext: String) -> [String] {
        let d = dir.isEmpty ? "." : dir
        let names = (try? FileManager.default.contentsOfDirectory(atPath: d)) ?? []
        return names.filter { $0.lowercased().hasSuffix(ext.lowercased()) }.sorted().map { (d as NSString).appendingPathComponent($0) }
    }

    /// Concatenate `files` into `output` (move when there is only one).
    public static func combine(files: [String], to output: String) throws {
        guard !files.isEmpty else { return }
        if files.count == 1 {
            try replaceItem(at: output, with: files[0])
            return
        }
        let dir = (output as NSString).deletingLastPathComponent
        if !dir.isEmpty { try FileManager.default.createDirectory(atPath: dir, withIntermediateDirectories: true) }
        FileManager.default.createFile(atPath: output, contents: nil)
        let out = try FileHandle(forWritingTo: URL(fileURLWithPath: output))
        defer { try? out.close() }
        for f in files {
            let input = try FileHandle(forReadingFrom: URL(fileURLWithPath: f))
            defer { try? input.close() }
            while let chunk = try input.read(upToCount: 4 * 1024 * 1024), !chunk.isEmpty {
                try out.write(contentsOf: chunk)
            }
        }
    }

    public static func replaceItem(at path: String, with source: String) throws {
        let fm = FileManager.default
        if fm.fileExists(atPath: path) { try fm.removeItem(atPath: path) }
        try fm.moveItem(atPath: source, toPath: path)
    }

    public static func fileSize(_ path: String) -> Int64 {
        ((try? FileManager.default.attributesOfItem(atPath: path))?[.size] as? Int64) ?? 0
    }
}
