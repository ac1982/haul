import Foundation

/// Credentials and endpoint overrides for one run. Immutable once the pipeline starts, so it is safe to share across tasks.
public struct Session: Sendable {
    /// Web cookie string (`SESSDATA=...; bili_jct=...`).
    public var cookie = ""
    /// APP/TV access_key.
    public var token = ""
    /// BiliPlus-style proxy host for the main API.
    public var host = "api.bilibili.com"
    /// Proxy host for `/pgc/view/web/season`.
    public var epHost = "api.bilibili.com"
    /// TV API host.
    public var tvHost = "api.snm0516.aisee.tv"
    /// BiliPlus area (`hk`/`tw`/`th`), empty when not proxying.
    public var area = ""
    /// WBI mixin key, filled in by `checkLogin`.
    public var wbiKey = ""
    public var userAgent = HTTPClient.randomUserAgent()

    public init() {}

    public var isBiliPlus: Bool { host != "api.bilibili.com" }
    public var hasCookie: Bool { !cookie.isEmpty }
    public var hasToken: Bool { !token.isEmpty }
}
