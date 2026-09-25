import Foundation

public enum Aria2c {
    public static func download(_ url: String, to path: String, config: DownloadConfig) async throws {
        let args = arguments(url, to: path, config: config)
        Log.debug("aria2c \(args.joined(separator: " "))")
        try await Shell.run(config.aria2cPath, args, echo: false)
        if FileManager.default.fileExists(atPath: path + ".aria2") || !FileManager.default.fileExists(atPath: path) {
            throw HaulError("aria2c failed")
        }
        Log.plain()
    }

    /// Same headers as the built-in downloader: the site's own when given, else the bilibili defaults.
    static func arguments(_ url: String, to path: String, config: DownloadConfig) -> [String] {
        var args = ["--auto-file-renaming=false", "--download-result=hide", "--allow-overwrite=true", "--console-log-level=warn",
                    "-x16", "-s16", "-j16", "-k5M"]
        if config.headers.isEmpty {
            if !url.contains("platform=android_tv_yst") && !url.contains("platform=android") {
                args.append("--header=Referer: https://www.bilibili.com")
            }
            args.append("--header=User-Agent: Mozilla/5.0")
        }
        for (k, v) in config.headers.sorted(by: { $0.key < $1.key }) { args.append("--header=\(k): \(v)") }
        if !config.cookie.isEmpty { args.append("--header=Cookie: \(config.cookie)") }
        args += splitArguments(config.aria2cArgs)
        let dir = (path as NSString).deletingLastPathComponent
        args += [url, "-d", dir.isEmpty ? "." : dir, "-o", (path as NSString).lastPathComponent]
        return args
    }

    /// Shell-style word splitting with single/double quotes, for `--aria2c-args`.
    public static func splitArguments(_ text: String) -> [String] {
        var out: [String] = []
        var current = ""
        var quote: Character?
        var hasToken = false
        var escaped = false
        for ch in text {
            if escaped { current.append(ch); escaped = false; continue }
            if ch == "\\" && quote != "'" { escaped = true; continue }
            if let q = quote {
                if ch == q { quote = nil } else { current.append(ch) }
                continue
            }
            if ch == "\"" || ch == "'" { quote = ch; hasToken = true; continue }
            if ch.isWhitespace {
                if hasToken { out.append(current); current = ""; hasToken = false }
                continue
            }
            current.append(ch)
            hasToken = true
        }
        if hasToken { out.append(current) }
        return out
    }
}
