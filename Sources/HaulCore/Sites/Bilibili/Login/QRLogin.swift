import Foundation

/// Scan-to-login for the web account (cookie) and the TV account (access token).
public enum QRLogin {
    public static let pngPath = "qrcode.png"

    public static func loginWeb(http: HTTPClient = .shared, session: Session = Session()) async throws {
        Log.status("Requesting a login QR code")
        let gen = try await http.getJSON("https://passport.bilibili.com/x/passport-login/web/qrcode/generate?source=main-fe-header", session: session)
        let url = try gen.get("data").get("url").stringValue
        let key = Format.queryValue("qrcode_key", in: url)
        try showQRCode(url)
        defer { try? FileManager.default.removeItem(atPath: pngPath) }

        var scanned = false
        while true {
            try await Task.sleep(for: .seconds(1))
            let poll = try await http.getJSON("https://passport.bilibili.com/x/passport-login/web/qrcode/poll?qrcode_key=\(key)&source=main-fe-header", session: session)
            switch poll["data"]["code"].int ?? -1 {
            case 86038:
                throw HaulError.auth("The QR code expired; run the login again")
            case 86101:
                continue
            case 86090:
                if !scanned { Log.status("Scanned; confirm on your phone"); scanned = true }
            default:
                let cc = poll["data"]["url"].stringValue
                Log.debug("SESSDATA=\(Format.queryValue("SESSDATA", in: cc))")
                // Cookie string from the query part; commas would break the header, so they are escaped.
                let query = cc.split(separator: "?", maxSplits: 1).last.map(String.init) ?? ""
                let cookie = query.replacingOccurrences(of: "&", with: ";").replacingOccurrences(of: ",", with: "%2C")
                try Storage.write(Storage.cookieFile, cookie)
                Log.success("Logged in  " + Log.style.dim("cookie saved to \(Terminal.prettyPath(Storage.url(Storage.cookieFile).path))"))
                return
            }
        }
    }

    public static func loginTV(http: HTTPClient = .shared, session: Session = Session()) async throws {
        var params = tvLoginParams()
        Log.status("Requesting a login QR code")
        let auth = try JSON.parse(try await http.postForm("https://passport.snm0516.aisee.tv/x/passport-tv-login/qrcode/auth_code", fields: params, session: session))
        let url = try auth.get("data").get("url").stringValue
        let authCode = auth["data"]["auth_code"].stringValue
        try showQRCode(url)
        defer { try? FileManager.default.removeItem(atPath: pngPath) }

        params = params.filter { $0.0 != "sign" }.map { $0.0 == "auth_code" ? ("auth_code", authCode) : $0.0 == "ts" ? ("ts", Format.unixSeconds()) : $0 }
        params.append(("sign", Signing.appSign(HTTPClient.formEncode(params), secret: Signing.tvAppSecret)))

        while true {
            try await Task.sleep(for: .seconds(1))
            let poll = try JSON.parse(try await http.postForm("https://passport.bilibili.com/x/passport-tv-login/qrcode/poll", fields: params, session: session))
            switch poll["code"].stringValue {
            case "86038":
                throw HaulError.auth("The QR code expired; run the login again")
            case "86039":
                continue
            default:
                let token = try poll.get("data").get("access_token").stringValue
                Log.debug("access_token=\(token)")
                try Storage.write(Storage.tvTokenFile, "access_token=\(token)")
                Log.success("Logged in  " + Log.style.dim("token saved to \(Terminal.prettyPath(Storage.url(Storage.tvTokenFile).path))"))
                return
            }
        }
    }

    /// The TV client's login form, signed. Keys stay in this (sorted) order for the signature.
    static func tvLoginParams() -> [(String, String)] {
        let deviceId = Format.randomString(20)
        let buvid = Format.randomString(37)
        let fingerprint = Format.compactTimestamp() + Format.randomString(45)
        var p: [(String, String)] = [
            ("appkey", Signing.tvAppKey), ("auth_code", ""), ("bili_local_id", deviceId), ("build", "102801"), ("buvid", buvid),
            ("channel", "master"), ("device", "OnePlus"), ("device_id", deviceId), ("device_name", "OnePlus7TPro"),
            ("device_platform", "Android10OnePlusHD1910"), ("fingerprint", fingerprint), ("guid", buvid),
            ("local_fingerprint", fingerprint), ("local_id", buvid), ("mobi_app", "android_tv_yst"), ("networkstate", "wifi"),
            ("platform", "android"), ("sys_ver", "29"), ("ts", Format.unixSeconds()),
        ]
        p.append(("sign", Signing.appSign(HTTPClient.formEncode(p), secret: Signing.tvAppSecret)))
        return p
    }

    static func showQRCode(_ text: String) throws {
        let qr = try QRImage(text: text)
        try qr.writePNG(to: pngPath, scale: 7)
        Log.info("Scan the QR code with the bilibili app " + Log.style.dim("(also saved as \(pngPath))"))
        qr.printToConsole()
    }
}
