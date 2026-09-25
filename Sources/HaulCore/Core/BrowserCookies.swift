import Foundation
#if canImport(CryptoKit)
import CryptoKit
#endif
#if os(macOS)
import CommonCrypto
import Security
import SQLite3
#endif

/// Reads login cookies straight out of a Chromium-based browser's profile (macOS).
///
/// Chromium stores cookie values AES-128-CBC encrypted with a key derived (PBKDF2, salt `saltysalt`, 1003 rounds)
/// from a random password kept in the login keychain as "<Browser> Safe Storage". Newer builds also prefix the
/// plaintext with SHA-256(host_key). Reading the profile needs Full Disk Access for the terminal; reading the
/// password makes macOS ask once.
public enum BrowserCookies {
    public enum Browser: String, CaseIterable, Sendable {
        case edge, chrome

        var displayName: String { self == .edge ? "Microsoft Edge" : "Google Chrome" }
        var dataDirectory: String {
            let base = FileManager.default.homeDirectoryForCurrentUser.appendingPathComponent("Library/Application Support").path
            return self == .edge ? base + "/Microsoft Edge" : base + "/Google/Chrome"
        }
        var keychainService: String { self == .edge ? "Microsoft Edge Safe Storage" : "Chrome Safe Storage" }
        var keychainAccount: String { self == .edge ? "Microsoft Edge" : "Chrome" }
    }

    public struct Cookie: Sendable, Equatable {
        public let host: String
        public let name: String
        public let value: String
    }

    /// The cookie header for bilibili from the given browser profile, e.g. `SESSDATA=…; bili_jct=…`.
    public static func bilibiliCookieHeader(from browser: Browser, profile: String = "Default") throws -> String {
        let cookies = try cookies(from: browser, profile: profile, hostSuffix: "bilibili.com")
        guard cookies.contains(where: { $0.name == "SESSDATA" && !$0.value.isEmpty }) else {
            throw HaulError.auth("The \(browser.displayName) profile \"\(profile)\" is not logged in to bilibili (no SESSDATA cookie); log in there first")
        }
        return header(cookies)
    }

    public static func header(_ cookies: [Cookie]) -> String {
        var seen: Set<String> = []
        return cookies.filter { !$0.value.isEmpty && seen.insert($0.name).inserted }
            .map { "\($0.name)=\($0.value)" }.joined(separator: "; ")
    }

    public static func cookies(from browser: Browser, profile: String, hostSuffix: String) throws -> [Cookie] {
        #if os(macOS)
        let profileDir = browser.dataDirectory + "/" + profile
        let candidates = [profileDir + "/Network/Cookies", profileDir + "/Cookies"]
        let fm = FileManager.default
        guard fm.fileExists(atPath: browser.dataDirectory) else {
            throw HaulError.auth("\(browser.displayName) data not found: \(browser.dataDirectory)")
        }
        guard let dbPath = candidates.first(where: { fm.fileExists(atPath: $0) }) else {
            if !fm.isReadableFile(atPath: profileDir) { throw fullDiskAccessError(browser) }
            throw HaulError.auth("No cookie database in the \(browser.displayName) profile \"\(profile)\"")
        }
        // Probe readability before touching the keychain, so a permissions problem doesn't trigger a pointless prompt.
        guard let probe = FileHandle(forReadingAtPath: dbPath) else { throw fullDiskAccessError(browser) }
        try? probe.close()
        let password = try safeStoragePassword(browser)
        do {
            return try cookies(databaseAt: dbPath, password: password, hostSuffix: hostSuffix)
        } catch let e as CocoaError where (e.underlying as? POSIXError)?.code == .EPERM || e.code == .fileReadNoPermission {
            throw fullDiskAccessError(browser)
        }
        #else
        throw HaulError.auth("Reading browser cookies works on macOS only")
        #endif
    }

    #if os(macOS)
    static func fullDiskAccessError(_ browser: Browser) -> HaulError {
        HaulError.auth("Cannot read the \(browser.displayName) profile (Operation not permitted).\n"
            + "macOS protects browser data: allow your terminal app under System Settings → Privacy & Security → Full Disk Access, then run again.")
    }

    /// The random password Chromium keeps in the login keychain. macOS prompts the user the first time.
    static func safeStoragePassword(_ browser: Browser) throws -> Data {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: browser.keychainService,
            kSecAttrAccount as String: browser.keychainAccount,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne,
        ]
        var item: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &item)
        guard status == errSecSuccess, let data = item as? Data else {
            let reason = SecCopyErrorMessageString(status, nil).map { String($0) } ?? "\(status)"
            throw HaulError.auth("Cannot read \"\(browser.keychainService)\" from the keychain: \(reason)\n"
                + "Choose \"Always Allow\" when macOS asks.")
        }
        return data
    }

    /// Decrypts every cookie whose host ends with `hostSuffix`. The database is copied first because the browser holds a lock on it.
    public static func cookies(databaseAt path: String, password: Data, hostSuffix: String) throws -> [Cookie] {
        let key = try deriveKey(password)
        let copy = FileManager.default.temporaryDirectory.appendingPathComponent("haul-cookies-\(UUID().uuidString).db").path
        try FileManager.default.copyItem(atPath: path, toPath: copy)
        defer { try? FileManager.default.removeItem(atPath: copy) }

        var db: OpaquePointer?
        guard sqlite3_open_v2(copy, &db, SQLITE_OPEN_READONLY, nil) == SQLITE_OK, let db else {
            throw HaulError("Cannot open the cookie database: \(path)")
        }
        defer { sqlite3_close(db) }

        var stmt: OpaquePointer?
        let sql = "SELECT host_key, name, value, encrypted_value FROM cookies WHERE host_key LIKE ? ORDER BY host_key, name"
        guard sqlite3_prepare_v2(db, sql, -1, &stmt, nil) == SQLITE_OK, let stmt else {
            throw HaulError("Cannot read the cookie database: \(String(cString: sqlite3_errmsg(db)))")
        }
        defer { sqlite3_finalize(stmt) }
        sqlite3_bind_text(stmt, 1, "%" + hostSuffix, -1, unsafeBitCast(-1, to: sqlite3_destructor_type.self))

        var out: [Cookie] = []
        while sqlite3_step(stmt) == SQLITE_ROW {
            let host = String(cString: sqlite3_column_text(stmt, 0))
            let name = String(cString: sqlite3_column_text(stmt, 1))
            var value = String(cString: sqlite3_column_text(stmt, 2))
            if value.isEmpty, let blob = sqlite3_column_blob(stmt, 3) {
                let data = Data(bytes: blob, count: Int(sqlite3_column_bytes(stmt, 3)))
                value = (try? decrypt(data, key: key, host: host)) ?? ""
            }
            out.append(Cookie(host: host, name: name, value: value))
        }
        return out
    }

    static func deriveKey(_ password: Data) throws -> Data {
        var key = Data(count: 16)
        let salt = Data("saltysalt".utf8)
        let status = key.withUnsafeMutableBytes { keyBuf in
            password.withUnsafeBytes { pwBuf in
                salt.withUnsafeBytes { saltBuf in
                    CCKeyDerivationPBKDF(CCPBKDFAlgorithm(kCCPBKDF2),
                                         pwBuf.baseAddress!.assumingMemoryBound(to: Int8.self), password.count,
                                         saltBuf.baseAddress!.assumingMemoryBound(to: UInt8.self), salt.count,
                                         CCPseudoRandomAlgorithm(kCCPRFHmacAlgSHA1), 1003,
                                         keyBuf.baseAddress!.assumingMemoryBound(to: UInt8.self), 16)
                }
            }
        }
        guard status == kCCSuccess else { throw HaulError("Key derivation failed: \(status)") }
        return key
    }

    /// `v10` + AES-128-CBC(PKCS7, IV = 16 spaces); the plaintext may start with SHA-256(host).
    static func decrypt(_ encrypted: Data, key: Data, host: String) throws -> String {
        guard encrypted.count > 3, encrypted.starts(with: Data("v10".utf8)) else {
            throw HaulError("Unsupported cookie encryption")
        }
        let body = encrypted.dropFirst(3)
        let iv = Data(repeating: 0x20, count: 16)
        var plain = Data(count: body.count + 16)
        let capacity = plain.count
        var moved = 0
        let status = plain.withUnsafeMutableBytes { outBuf in
            body.withUnsafeBytes { inBuf in
                key.withUnsafeBytes { keyBuf in
                    iv.withUnsafeBytes { ivBuf in
                        CCCrypt(CCOperation(kCCDecrypt), CCAlgorithm(kCCAlgorithmAES128), CCOptions(kCCOptionPKCS7Padding),
                                keyBuf.baseAddress, 16, ivBuf.baseAddress,
                                inBuf.baseAddress, body.count, outBuf.baseAddress, capacity, &moved)
                    }
                }
            }
        }
        guard status == kCCSuccess else { throw HaulError("Cookie decryption failed: \(status)") }
        plain.count = moved
        let hostHash = Data(SHA256.hash(data: Data(host.utf8)))
        if plain.count >= 32, plain.prefix(32) == hostHash { plain = plain.dropFirst(32) }
        return String(decoding: plain, as: UTF8.self)
    }

    /// Test helper: the inverse of `decrypt`, in the newer (host-hashed) layout.
    static func encrypt(_ value: String, key: Data, host: String) throws -> Data {
        var plain = Data(SHA256.hash(data: Data(host.utf8)))
        plain.append(Data(value.utf8))
        let iv = Data(repeating: 0x20, count: 16)
        var out = Data(count: plain.count + 16)
        let capacity = out.count
        var moved = 0
        let status = out.withUnsafeMutableBytes { outBuf in
            plain.withUnsafeBytes { inBuf in
                key.withUnsafeBytes { keyBuf in
                    iv.withUnsafeBytes { ivBuf in
                        CCCrypt(CCOperation(kCCEncrypt), CCAlgorithm(kCCAlgorithmAES128), CCOptions(kCCOptionPKCS7Padding),
                                keyBuf.baseAddress, 16, ivBuf.baseAddress,
                                inBuf.baseAddress, plain.count, outBuf.baseAddress, capacity, &moved)
                    }
                }
            }
        }
        guard status == kCCSuccess else { throw HaulError("encrypt failed: \(status)") }
        out.count = moved
        return Data("v10".utf8) + out
    }
    #endif
}
