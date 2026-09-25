import Foundation

public enum Auth {
    /// Reads stored credentials into the session for the API in use. Explicit values already in the session win.
    public static func loadStored(into session: inout Session, api: APIType) {
        if session.cookie.isEmpty, let c = Storage.read(Storage.cookieFile) {
            Log.debug("Using the stored cookie: \(Storage.url(Storage.cookieFile).path)")
            session.cookie = c
        }
        if session.token.isEmpty, api == .tv, let t = Storage.read(Storage.tvTokenFile) {
            Log.debug("Using the stored TV token")
            session.token = t.replacingOccurrences(of: "access_token=", with: "")
        }
        if session.token.isEmpty, api == .app, let t = Storage.read(Storage.appTokenFile) ?? Storage.read(Storage.tvTokenFile) {
            Log.debug("Using the stored APP token")
            session.token = t.replacingOccurrences(of: "access_token=", with: "")
        }
    }

    public struct NavInfo: Sendable {
        public var isLoggedIn: Bool
        public var wbiKey: String
    }

    /// `x/web-interface/nav`: login state plus the WBI image keys.
    public static func nav(http: HTTPClient = .shared, session: Session) async throws -> NavInfo {
        let json = try await http.getJSON("https://api.bilibili.com/x/web-interface/nav", session: session)
        let data = try json.get("data")
        let img = data["wbi_img"]
        let key = Signing.wbiMixinKey(imgKey: Signing.imageKey(from: img["img_url"].stringValue),
                                      subKey: Signing.imageKey(from: img["sub_url"].stringValue))
        Log.debug("wbi: \(key)")
        return NavInfo(isLoggedIn: data["isLogin"].bool ?? false, wbiKey: key)
    }

    public static func fetchWbiKey(http: HTTPClient = .shared, session: Session) async throws -> String {
        try await nav(http: http, session: session).wbiKey
    }
}
