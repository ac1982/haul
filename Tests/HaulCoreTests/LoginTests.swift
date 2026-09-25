import Foundation
import Testing
@testable import HaulCore

#if os(macOS)
import SQLite3

@Suite struct BrowserCookieTests {
    /// Build a Chromium-style Cookies database with values encrypted the way Edge/Chrome do, then read it back.
    @Test func decryptsChromiumCookies() throws {
        let password = Data("fake-safe-storage-password".utf8)
        let key = try BrowserCookies.deriveKey(password)
        #expect(key.count == 16)

        let dir = FileManager.default.temporaryDirectory.appendingPathComponent("bilidl-cookies-\(UUID().uuidString)")
        try FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        defer { try? FileManager.default.removeItem(at: dir) }
        let dbPath = dir.appendingPathComponent("Cookies").path

        var db: OpaquePointer?
        #expect(sqlite3_open(dbPath, &db) == SQLITE_OK)
        defer { sqlite3_close(db) }
        #expect(sqlite3_exec(db, "CREATE TABLE cookies (host_key TEXT, name TEXT, value TEXT, encrypted_value BLOB)", nil, nil, nil) == SQLITE_OK)
        func insert(host: String, name: String, value: String) throws {
            var stmt: OpaquePointer?
            #expect(sqlite3_prepare_v2(db, "INSERT INTO cookies VALUES (?, ?, '', ?)", -1, &stmt, nil) == SQLITE_OK)
            defer { sqlite3_finalize(stmt) }
            let transient = unsafeBitCast(-1, to: sqlite3_destructor_type.self)
            sqlite3_bind_text(stmt, 1, host, -1, transient)
            sqlite3_bind_text(stmt, 2, name, -1, transient)
            let blob = try BrowserCookies.encrypt(value, key: key, host: host)
            _ = blob.withUnsafeBytes { sqlite3_bind_blob(stmt, 3, $0.baseAddress, Int32(blob.count), transient) }
            #expect(sqlite3_step(stmt) == SQLITE_DONE)
        }
        try insert(host: ".bilibili.com", name: "SESSDATA", value: "abc%2Cdef")
        try insert(host: ".bilibili.com", name: "bili_jct", value: "csrf123")
        try insert(host: ".bilibili.com", name: "DedeUserID", value: "42")
        try insert(host: ".example.com", name: "other", value: "no")

        let cookies = try BrowserCookies.cookies(databaseAt: dbPath, password: password, hostSuffix: "bilibili.com")
        #expect(cookies.map(\.name) == ["DedeUserID", "SESSDATA", "bili_jct"])
        #expect(cookies.first { $0.name == "SESSDATA" }?.value == "abc%2Cdef")
        #expect(BrowserCookies.header(cookies) == "DedeUserID=42; SESSDATA=abc%2Cdef; bili_jct=csrf123")
    }

    @Test func decryptWithoutHostHashPrefix() throws {
        // Older Chromium builds do not prefix SHA-256(host); the raw plaintext must come through untouched.
        let key = try BrowserCookies.deriveKey(Data("pw".utf8))
        let encrypted = try BrowserCookies.encrypt("plain", key: key, host: "h")
        #expect(try BrowserCookies.decrypt(encrypted, key: key, host: "h") == "plain")
        #expect(try BrowserCookies.decrypt(encrypted, key: key, host: "other-host").hasSuffix("plain"))
    }
}
#endif
