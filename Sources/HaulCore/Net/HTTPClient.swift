import Foundation
import Synchronization
#if canImport(FoundationNetworking)
import FoundationNetworking
#endif

/// One shared URLSession for the whole tool. Cookies are handled by hand because the API wants them verbatim.
public final class HTTPClient: @unchecked Sendable {
    public static let shared = HTTPClient()

    private let session: URLSession
    private let streamSession: URLSession
    private let streamDelegate: StreamDelegate

    private convenience init() { self.init(configure: { _ in }) }

    /// `configure` lets tests route requests through a stub `URLProtocol`.
    init(configure: (URLSessionConfiguration) -> Void) {
        let cfg = URLSessionConfiguration.ephemeral
        cfg.httpCookieStorage = nil
        cfg.httpShouldSetCookies = false
        cfg.httpCookieAcceptPolicy = .never
        cfg.timeoutIntervalForRequest = 120
        cfg.timeoutIntervalForResource = 6 * 3600
        cfg.httpMaximumConnectionsPerHost = 64
        cfg.requestCachePolicy = .reloadIgnoringLocalCacheData
        configure(cfg)
        session = URLSession(configuration: cfg)
        streamDelegate = StreamDelegate()
        streamSession = URLSession(configuration: cfg, delegate: streamDelegate, delegateQueue: nil)
    }

    // MARK: user agent

    private static let platforms = ["Windows NT 10.0; Win64", "Macintosh; Intel Mac OS X 10_15", "X11; Linux x86_64"]

    public static func randomUserAgent() -> String {
        func version() -> String { String(format: "%.3f", Double.random(in: 80...110)) }
        let browsers = [
            "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/\(version()) Safari/537.36",
            "Gecko/20100101 Firefox/\(version())",
        ]
        return "Mozilla/5.0 (\(platforms.randomElement()!)) \(browsers.randomElement()!)"
    }

    public static let androidUserAgent =
        "Dalvik/2.1.0 (Linux; U; Android 6.0.1; oneplus a5010 Build/V417IR) 6.10.0 os/android model/oneplus a5010 mobi_app/android build/6100500 channel/bili innerVer/6100500 osVer/6.0.1 network/2"

    // MARK: helpers

    private func makeRequest(_ url: String, method: String = "GET") throws -> URLRequest {
        guard let u = URL(string: url) else { throw HaulError("Invalid URL: \(url)") }
        var r = URLRequest(url: u)
        r.httpMethod = method
        r.cachePolicy = .reloadIgnoringLocalCacheData
        return r
    }

    private static func ensureSuccess(_ response: URLResponse, _ url: String) throws {
        guard let http = response as? HTTPURLResponse else { return }
        guard (200..<300).contains(http.statusCode) else {
            throw HaulError("HTTP \(http.statusCode): \(url)")
        }
    }

    /// The `Cookie` header for an API call. Episode/season endpoints also need CURRENT_FNVAL to get DASH back.
    private static func cookieHeader(for url: String, session: Session) -> String? {
        let base = session.cookie
        let value = (url.contains("/ep") || url.contains("/ss")) ? base + ";CURRENT_FNVAL=4048;" : base
        return value.isEmpty ? nil : value
    }

    // MARK: text / JSON

    public func getString(_ url: String, session s: Session, userAgent: String? = nil) async throws -> String {
        var req = try makeRequest(url)
        req.setValue(userAgent ?? s.userAgent, forHTTPHeaderField: "User-Agent")
        if let cookie = Self.cookieHeader(for: url, session: s) { req.setValue(cookie, forHTTPHeaderField: "Cookie") }
        if url.contains("api.bilibili.com") { req.setValue("https://www.bilibili.com/", forHTTPHeaderField: "Referer") }
        if url.contains("api.bilibili.tv") {
            req.setValue("\"Google Chrome\";v=\"131\", \"Chromium\";v=\"131\", \"Not_A Brand\";v=\"24\"", forHTTPHeaderField: "sec-ch-ua")
        }
        req.setValue("no-cache", forHTTPHeaderField: "Cache-Control")
        Log.debug("GET \(url)")
        let (data, resp) = try await session.data(for: req)
        try Self.ensureSuccess(resp, url)
        let text = String(decoding: data, as: UTF8.self)
        Log.debug("Response: \(text)")
        return text
    }

    public func getJSON(_ url: String, session s: Session, userAgent: String? = nil) async throws -> JSON {
        try JSON.parse(await getString(url, session: s, userAgent: userAgent))
    }

    /// Follow redirects with a HEAD and report where we ended up.
    public func finalURL(_ url: String, session s: Session) async throws -> String {
        var req = try makeRequest(url, method: "HEAD")
        req.setValue(s.userAgent, forHTTPHeaderField: "User-Agent")
        req.setValue("no-cache", forHTTPHeaderField: "Cache-Control")
        Log.debug("HEAD \(url)")
        let (_, resp) = try await session.data(for: req)
        try Self.ensureSuccess(resp, url)
        let location = resp.url?.absoluteString ?? url
        Log.debug("Location: \(location)")
        return location
    }

    // MARK: POST

    public func postGRPC(_ url: String, body: Data, headers: [String: String]? = nil) async throws -> Data {
        var req = try makeRequest(url, method: "POST")
        req.httpBody = body
        req.setValue("application/grpc", forHTTPHeaderField: "Content-Type")
        if let headers {
            for (k, v) in headers { req.setValue(v, forHTTPHeaderField: k) }
        } else {
            req.setValue(Self.androidUserAgent, forHTTPHeaderField: "User-Agent")
            req.setValue("gzip", forHTTPHeaderField: "grpc-encoding")
        }
        Log.debug("POST(grpc) \(url) body=\(body.base64EncodedString())")
        let (data, _) = try await session.data(for: req)
        return data
    }

    public func postForm(_ url: String, fields: [(String, String)], session s: Session) async throws -> Data {
        var req = try makeRequest(url, method: "POST")
        req.setValue("application/x-www-form-urlencoded", forHTTPHeaderField: "Content-Type")
        req.setValue(s.userAgent, forHTTPHeaderField: "User-Agent")
        req.httpBody = Data(Self.formEncode(fields).utf8)
        Log.debug("POST(form) \(url)")
        let (data, resp) = try await session.data(for: req)
        try Self.ensureSuccess(resp, url)
        return data
    }

    public func postJSON(_ url: String, json: Data) async throws {
        var req = try makeRequest(url, method: "POST")
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        req.httpBody = json
        _ = try await session.data(for: req)
    }

    public static func formEncode(_ fields: [(String, String)]) -> String {
        fields.map { "\(Format.percentEncode($0.0))=\(Format.percentEncode($0.1))" }.joined(separator: "&")
    }

    // MARK: streaming (downloads)

    public struct ResponseHead: Sendable {
        public let statusCode: Int
        public let contentLength: Int64?
        /// The full size from `Content-Range: bytes a-b/total`.
        public var rangeTotal: Int64? = nil
    }

    public struct StreamingResponse: Sendable {
        public let head: ResponseHead
        public let body: AsyncThrowingStream<Data, Error>
        public let cancel: @Sendable () -> Void
    }

    /// Start a request and hand back the response head plus a stream of body chunks.
    public func stream(_ request: URLRequest) async throws -> StreamingResponse {
        let task = streamSession.dataTask(with: request)
        let (bodyStream, bodyContinuation) = AsyncThrowingStream<Data, Error>.makeStream()
        let cancel: @Sendable () -> Void = { task.cancel() }
        bodyContinuation.onTermination = { _ in cancel() }
        // Cancelling the calling Task stops the transfer whether we are waiting for the head or streaming the body.
        return try await withTaskCancellationHandler {
            let head: ResponseHead = try await withCheckedThrowingContinuation { cont in
                streamDelegate.register(task.taskIdentifier, head: cont, body: bodyContinuation)
                task.resume()
            }
            return StreamingResponse(head: head, body: bodyStream, cancel: cancel)
        } onCancel: {
            cancel()
        }
    }

    public func cancelAllStreams() { streamSession.getAllTasks { $0.forEach { $0.cancel() } } }
}

/// Multiplexes URLSession delegate callbacks into per-task continuations.
final class StreamDelegate: NSObject, URLSessionDataDelegate, @unchecked Sendable {
    private struct Entry {
        var head: CheckedContinuation<HTTPClient.ResponseHead, Error>?
        var body: AsyncThrowingStream<Data, Error>.Continuation
    }

    private let entries = Mutex<[Int: Entry]>([:])

    func register(_ id: Int, head: CheckedContinuation<HTTPClient.ResponseHead, Error>, body: AsyncThrowingStream<Data, Error>.Continuation) {
        entries.withLock { $0[id] = Entry(head: head, body: body) }
    }

    func urlSession(_ session: URLSession, dataTask: URLSessionDataTask, didReceive response: URLResponse,
                    completionHandler: @escaping @Sendable (URLSession.ResponseDisposition) -> Void) {
        let http = response as? HTTPURLResponse
        let length = http.flatMap { $0.value(forHTTPHeaderField: "Content-Length") }.flatMap { Int64($0) }
        let total = http.flatMap { $0.value(forHTTPHeaderField: "Content-Range") }
            .flatMap { $0.split(separator: "/").last }.flatMap { Int64($0) }
        let head = HTTPClient.ResponseHead(statusCode: http?.statusCode ?? 200, contentLength: length, rangeTotal: total)
        let cont = entries.withLock { e -> CheckedContinuation<HTTPClient.ResponseHead, Error>? in
            let c = e[dataTask.taskIdentifier]?.head
            e[dataTask.taskIdentifier]?.head = nil
            return c
        }
        cont?.resume(returning: head)
        completionHandler(.allow)
    }

    func urlSession(_ session: URLSession, dataTask: URLSessionDataTask, didReceive data: Data) {
        let body = entries.withLock { $0[dataTask.taskIdentifier]?.body }
        body?.yield(data)
    }

    func urlSession(_ session: URLSession, task: URLSessionTask, didCompleteWithError error: Error?) {
        let entry = entries.withLock { $0.removeValue(forKey: task.taskIdentifier) }
        guard let entry else { return }
        if let error {
            entry.head?.resume(throwing: error)
            entry.body.finish(throwing: error)
        } else {
            entry.head?.resume(throwing: HaulError("The connection closed before the response headers arrived"))
            entry.body.finish()
        }
    }
}
