import Foundation
import Synchronization
@testable import HaulCore

/// One canned HTTP answer.
struct StubResponse: Sendable {
    var status = 200
    var headers: [String: String] = [:]
    var body = Data()
    /// Send only this many body bytes, then end the response (a dropped connection).
    var truncateAt: Int?
    /// Answer with a 302 to this URL instead.
    var redirect: String?

    static func json(_ text: String) -> StubResponse {
        StubResponse(headers: ["Content-Type": "application/json"], body: Data(text.utf8))
    }

    static func status(_ code: Int) -> StubResponse { StubResponse(status: code) }

    /// Serve `data` honouring a `Range: bytes=a-b` header, the way a CDN does.
    static func ranged(_ data: Data, for request: URLRequest) -> StubResponse {
        guard let range = Stub.range(request), range.from < data.count else {
            return StubResponse(status: request.value(forHTTPHeaderField: "Range") == nil ? 200 : 416, body: request.value(forHTTPHeaderField: "Range") == nil ? data : Data())
        }
        let end = min(range.to ?? data.count - 1, data.count - 1)
        return StubResponse(status: 206, headers: ["Content-Range": "bytes \(range.from)-\(end)/\(data.count)"],
                            body: data.subdata(in: range.from..<(end + 1)))
    }
}

/// A `URLProtocol` that answers every request of the stub `HTTPClient` from registered routes.
/// Routes are process-wide; each test registers URLs of its own (distinct ids or hosts), so suites can run in parallel.
final class StubProtocol: URLProtocol, @unchecked Sendable {
    override class func canInit(with request: URLRequest) -> Bool { true }
    override class func canonicalRequest(for request: URLRequest) -> URLRequest { request }
    override func stopLoading() {}

    override func startLoading() {
        guard let url = request.url, let client else { return }
        Stub.log.withLock { $0.append(request) }
        guard let respond = Stub.handler(for: url) else {
            client.urlProtocol(self, didFailWithError: URLError(.cannotConnectToHost, userInfo: [NSURLErrorFailingURLStringErrorKey: url.absoluteString]))
            return
        }
        let r = respond(request)
        if let target = r.redirect, let next = URL(string: target) {
            let resp = HTTPURLResponse(url: url, statusCode: 302, httpVersion: "HTTP/1.1", headerFields: ["Location": target])!
            var follow = request
            follow.url = next
            client.urlProtocol(self, wasRedirectedTo: follow, redirectResponse: resp)
            client.urlProtocolDidFinishLoading(self)
            return
        }
        var headers = r.headers
        headers["Content-Length"] = headers["Content-Length"] ?? String(r.body.count)
        let resp = HTTPURLResponse(url: url, statusCode: r.status, httpVersion: "HTTP/1.1", headerFields: headers)!
        client.urlProtocol(self, didReceive: resp, cacheStoragePolicy: .notAllowed)
        if request.httpMethod != "HEAD" {
            let body = r.truncateAt.map { r.body.prefix($0) } ?? r.body
            var offset = 0
            while offset < body.count {
                let n = min(64 * 1024, body.count - offset)
                client.urlProtocol(self, didLoad: body.subdata(in: (body.startIndex + offset)..<(body.startIndex + offset + n)))
                offset += n
            }
        }
        client.urlProtocolDidFinishLoading(self)
    }
}

enum Stub {
    typealias Handler = @Sendable (URLRequest) -> StubResponse

    /// The `HTTPClient` whose requests never leave the process.
    static let client = HTTPClient(configure: { $0.protocolClasses = [StubProtocol.self] })

    fileprivate static let routes = Mutex<[(specificity: Int, matches: @Sendable (URL) -> Bool, respond: Handler)]>([])
    fileprivate static let log = Mutex<[URLRequest]>([])

    /// Answer URLs containing `fragment`. The most specific (longest) matching fragment wins, so a test's own
    /// route beats a shared one whatever order parallel tests register them in.
    static func on(_ fragment: String, _ respond: @escaping Handler) {
        on(specificity: fragment.count, { $0.absoluteString.contains(fragment) }, respond)
    }

    static func on(specificity: Int = 1000, _ matches: @escaping @Sendable (URL) -> Bool, _ respond: @escaping Handler) {
        routes.withLock { $0.insert((specificity, matches, respond), at: 0) }
    }

    /// Answer URLs containing `fragment` with a fixture file.
    static func fixture(_ fragment: String, _ name: String) throws {
        let data = try Fixture.data(name)
        on(fragment) { _ in StubResponse(headers: ["Content-Type": "application/json"], body: data) }
    }

    fileprivate static func handler(for url: URL) -> Handler? {
        routes.withLock { list in
            // max(by:) keeps the first of equals, and newer routes are in front.
            list.filter { $0.matches(url) }.max { $0.specificity < $1.specificity }?.respond
        }
    }

    /// Every request so far whose URL contains `fragment`.
    static func requests(_ fragment: String) -> [URLRequest] {
        log.withLock { $0.filter { $0.url?.absoluteString.contains(fragment) == true } }
    }

    static func range(_ request: URLRequest) -> (from: Int, to: Int?)? {
        guard let header = request.value(forHTTPHeaderField: "Range"),
              let m = header.wholeMatch(of: /bytes=(\d+)-(\d*)/), let from = Int(m.1) else { return nil }
        return (from, Int(m.2))
    }
}

enum Fixture {
    static func data(_ name: String) throws -> Data {
        guard let url = Bundle.module.url(forResource: name, withExtension: nil, subdirectory: "Fixtures") else {
            throw HaulError("fixture not found: \(name)")
        }
        return try Data(contentsOf: url)
    }

    static func json(_ name: String) throws -> JSON { try JSON.parse(data(name)) }
}

/// A fresh directory under the system temp dir.
func makeTempDir(_ label: String = "test") throws -> String {
    let dir = (NSTemporaryDirectory() as NSString).appendingPathComponent("haul-\(label)-\(UUID().uuidString.prefix(8))")
    try FileManager.default.createDirectory(atPath: dir, withIntermediateDirectories: true)
    return dir
}

/// Bytes with a pattern, so a misplaced range shows up as a mismatch.
func patternData(_ count: Int) -> Data {
    Data((0..<count).map { UInt8(truncatingIfNeeded: $0 &* 31 &+ $0 >> 8) })
}
