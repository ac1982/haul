import Foundation

/// Where credentials, config and the download archive live: `$HAUL_HOME`, else `$XDG_CONFIG_HOME/haul`, else `~/.config/haul`.
public enum Storage {
    public static let cookieFile = "cookie.txt"
    public static let tvTokenFile = "tv-token.txt"
    public static let appTokenFile = "app-token.txt"
    public static let configFile = "config.json"
    public static let archiveFile = "archives.txt"

    public static var home: URL {
        let env = ProcessInfo.processInfo.environment
        if let h = env["HAUL_HOME"], !h.isEmpty { return URL(fileURLWithPath: h) }
        if let x = env["XDG_CONFIG_HOME"], !x.isEmpty { return URL(fileURLWithPath: x).appendingPathComponent("haul") }
        return FileManager.default.homeDirectoryForCurrentUser.appendingPathComponent(".config/haul")
    }

    public static func url(_ name: String) -> URL { home.appendingPathComponent(name) }

    public static func read(_ name: String) -> String? {
        guard let s = try? String(contentsOf: url(name), encoding: .utf8) else { return nil }
        let t = s.trimmingCharacters(in: .whitespacesAndNewlines)
        return t.isEmpty ? nil : t
    }

    public static func write(_ name: String, _ text: String) throws {
        try FileManager.default.createDirectory(at: home, withIntermediateDirectories: true)
        try text.write(to: url(name), atomically: true, encoding: .utf8)
    }

    public static func append(_ name: String, _ text: String) throws {
        try FileManager.default.createDirectory(at: home, withIntermediateDirectories: true)
        let u = url(name)
        if !FileManager.default.fileExists(atPath: u.path) { try "".write(to: u, atomically: true, encoding: .utf8) }
        let h = try FileHandle(forWritingTo: u)
        defer { try? h.close() }
        try h.seekToEnd()
        try h.write(contentsOf: Data(text.utf8))
    }
}
