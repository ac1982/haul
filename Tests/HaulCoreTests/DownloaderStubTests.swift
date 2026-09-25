import Foundation
import Synchronization
import Testing
@testable import HaulCore

/// The downloader against simulated servers: a normal CDN, one that ignores ranges, one that serves only closed
/// ranges (googlevideo), and connections that drop.
@Suite struct DownloaderStubTests {
    let http = Stub.client

    /// A file bigger than `n` MB, built from a 1 MB pattern so it is quick to make and a misplaced byte still shows.
    static func media(megabytes n: Int, extra: Int = 12345) -> Data {
        let block = patternData(1 << 20)
        var d = Data(capacity: n << 20 + extra)
        for i in 0..<n { d.append(block); d[d.count - 1] = UInt8(i & 0xff) }
        d.append(patternData(extra))
        return d
    }

    func url(_ name: String) -> String { "https://dl-\(name).test/media/\(name).m4s" }

    @Test func singleStream() async throws {
        let data = patternData(300_000)
        Stub.on(url("single")) { .ranged(data, for: $0) }
        let out = (try makeTempDir("single") as NSString).appendingPathComponent("a.m4a")
        try await Downloader.downloadTrack(url("single"), to: out, config: DownloadConfig(), isVideo: false, http: http)
        #expect(FileManager.default.contents(atPath: out) == data)
        #expect(Stub.requests(url("single")).map { $0.value(forHTTPHeaderField: "Range") } == ["bytes=0-"])
    }

    @Test func segmentedAcrossClips() async throws {
        let data = Self.media(megabytes: 41)
        Stub.on(url("segmented")) { .ranged(data, for: $0) }
        var config = DownloadConfig()
        config.multiThread = true
        let dir = try makeTempDir("segmented")
        let out = (dir as NSString).appendingPathComponent("v.mp4")
        try await Downloader.downloadTrack(url("segmented"), to: out, config: config, isVideo: true, http: http)
        #expect(FileManager.default.contents(atPath: out) == data)
        let ranges = Set(Stub.requests(url("segmented")).compactMap { $0.value(forHTTPHeaderField: "Range") })
        #expect(ranges.isSuperset(of: ["bytes=0-20971520", "bytes=20971521-41943041", "bytes=41943042-"]))
        // Clip files are merged and removed.
        #expect(try FileManager.default.contentsOfDirectory(atPath: dir) == ["v.mp4"])
    }

    @Test func serverIgnoringRangesFallsBackToOneStream() async throws {
        let data = Self.media(megabytes: 41)
        Stub.on(url("norange")) { _ in StubResponse(body: data) }
        var config = DownloadConfig()
        config.multiThread = true
        let out = (try makeTempDir("norange") as NSString).appendingPathComponent("v.mp4")
        try await Downloader.downloadTrack(url("norange"), to: out, config: config, isVideo: true, http: http)
        #expect(FileManager.default.contents(atPath: out) == data)
    }

    /// googlevideo: plain and open-ended requests get 403, and so does a range over the limit.
    static func rangeOnly(_ data: Data, limit: Int, request: URLRequest) -> StubResponse {
        guard let r = Stub.range(request), let to = r.to, to - r.from < limit else { return .status(403) }
        return .ranged(data, for: request)
    }

    @Test func rangeOnlyServerGetsClosedRangesOneAtATime() async throws {
        let data = patternData(2_500_000)
        Stub.on(url("rangeonly")) { Self.rangeOnly(data, limit: 1_000_000, request: $0) }
        var config = DownloadConfig()
        config.multiThread = true
        config.maxRange = 1_000_000
        let out = (try makeTempDir("rangeonly") as NSString).appendingPathComponent("v.mp4")
        try await Downloader.downloadTrack(url("rangeonly"), to: out, config: config, isVideo: true, http: http)
        #expect(FileManager.default.contents(atPath: out) == data)
        // The size comes from a one-byte probe, then three closed ranges in order.
        #expect(Stub.requests(url("rangeonly")).map { $0.value(forHTTPHeaderField: "Range") }
                == ["bytes=0-0", "bytes=0-999999", "bytes=1000000-1999999", "bytes=2000000-2499999"])
    }

    @Test func rangeOnlyServerWithoutASizeFails() async throws {
        Stub.on(url("nosize")) { request in
            Stub.range(request)?.to == nil ? .status(403) : StubResponse(status: 206, headers: ["Content-Range": "bytes 0-0/*"], body: Data([0]))
        }
        var config = DownloadConfig()
        config.maxRange = 1_000_000
        let dir = try makeTempDir("nosize")
        await #expect(throws: HaulError.self) {
            try await Downloader.downloadTrack(url("nosize"), to: (dir as NSString).appendingPathComponent("v.mp4"), config: config, isVideo: true, http: http)
        }
        #expect(!FileManager.default.fileExists(atPath: (dir as NSString).appendingPathComponent("v.mp4")))
    }

    @Test func unknownSizeFallsBackToOneStream() async throws {
        let data = patternData(50_000)
        // No Content-Length on the probe: the length is unknown, so no clips can be planned.
        Stub.on(url("nolength")) { request in
            var r = StubResponse.ranged(data, for: request)
            if Stub.range(request) == nil { r.headers["Content-Length"] = "" }
            return r
        }
        var config = DownloadConfig()
        config.multiThread = true
        let out = (try makeTempDir("nolength") as NSString).appendingPathComponent("v.mp4")
        try await Downloader.downloadTrack(url("nolength"), to: out, config: config, isVideo: true, http: http)
        #expect(FileManager.default.contents(atPath: out) == data)
    }

    @Test func knownSizeSkipsTheProbe() async throws {
        let data = patternData(1_500_000)
        Stub.on(url("knownsize")) { Self.rangeOnly(data, limit: 1_000_000, request: $0) }
        var config = DownloadConfig()
        config.maxRange = 1_000_000
        config.size = Int64(data.count)
        let out = (try makeTempDir("knownsize") as NSString).appendingPathComponent("a.m4a")
        try await Downloader.downloadTrack(url("knownsize"), to: out, config: config, isVideo: false, http: http)
        #expect(FileManager.default.contents(atPath: out) == data)
        #expect(Stub.requests(url("knownsize")).map { $0.value(forHTTPHeaderField: "Range") } == ["bytes=0-999999", "bytes=1000000-1499999"])
    }

    @Test func droppedConnectionResumes() async throws {
        let data = patternData(400_000)
        let calls = Mutex(0)
        Stub.on(url("dropped")) { request in
            var r = StubResponse.ranged(data, for: request)
            if calls.withLock({ $0 += 1; return $0 }) == 1 {
                // Promise everything, deliver half.
                r.headers["Content-Length"] = String(r.body.count)
                r.truncateAt = 150_000
            }
            return r
        }
        let out = (try makeTempDir("dropped") as NSString).appendingPathComponent("a.m4a")
        try await Downloader.downloadFile(url("dropped"), to: out, config: DownloadConfig(), http: http)
        #expect(FileManager.default.contents(atPath: out) == data)
        #expect(Stub.requests(url("dropped")).map { $0.value(forHTTPHeaderField: "Range") } == ["bytes=0-", "bytes=150000-"])
    }

    @Test func partialFileIsResumed() async throws {
        let data = patternData(200_000)
        Stub.on(url("partial")) { .ranged(data, for: $0) }
        let dir = try makeTempDir("partial")
        let out = (dir as NSString).appendingPathComponent("a.m4a")
        try data.prefix(1000).write(to: URL(fileURLWithPath: (dir as NSString).appendingPathComponent("a.tmp")))
        try await Downloader.downloadFile(url("partial"), to: out, config: DownloadConfig(), http: http)
        #expect(FileManager.default.contents(atPath: out) == data)
        #expect(Stub.requests(url("partial")).map { $0.value(forHTTPHeaderField: "Range") } == ["bytes=1000-"])
    }

    @Test func httpErrorsFail() async throws {
        Stub.on(url("missing")) { _ in .status(404) }
        let out = (try makeTempDir("missing") as NSString).appendingPathComponent("a.m4a")
        await #expect(throws: HaulError.self) {
            try await Downloader.downloadTrack(url("missing"), to: out, config: DownloadConfig(), isVideo: false, http: http)
        }
        #expect(!FileManager.default.fileExists(atPath: out))
    }

    @Test func requestHeaders() async throws {
        let data = patternData(1000)
        Stub.on(url("headers")) { .ranged(data, for: $0) }
        let dir = try makeTempDir("headers")

        var bili = DownloadConfig()
        bili.cookie = "SESSDATA=x"
        try await Downloader.downloadFile(url("headers"), to: (dir as NSString).appendingPathComponent("1.bin"), config: bili, http: http)
        var other = DownloadConfig()
        other.headers = ["User-Agent": "UA-X"]
        try await Downloader.downloadFile(url("headers"), to: (dir as NSString).appendingPathComponent("2.bin"), config: other, http: http)

        let reqs = Stub.requests(url("headers"))
        #expect(reqs.count == 2)
        #expect(reqs[0].value(forHTTPHeaderField: "Referer") == "https://www.bilibili.com")
        #expect(reqs[0].value(forHTTPHeaderField: "Cookie") == "SESSDATA=x")
        #expect(reqs[0].value(forHTTPHeaderField: "User-Agent") == "Mozilla/5.0")
        // Another site's headers replace the bilibili ones entirely.
        #expect(reqs[1].value(forHTTPHeaderField: "Referer") == nil)
        #expect(reqs[1].value(forHTTPHeaderField: "Cookie") == nil)
        #expect(reqs[1].value(forHTTPHeaderField: "User-Agent") == "UA-X")
    }
}
