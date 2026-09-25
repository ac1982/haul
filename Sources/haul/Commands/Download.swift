import ArgumentParser
import HaulCore
import Foundation

struct Download: AsyncParsableCommand {
    static let configuration = CommandConfiguration(
        abstract: "Download a link. The default: `haul <url>` is the same.",
        discussion: """
            Picks the best streams (or those -q / -c / --video-stream / --audio-stream ask for), downloads them and muxes \
            video, audio, subtitles, chapters and cover into one file. Pages that already exist are skipped. \
            With --json, stdout gets one JSON document: every page, the streams chosen and the files written.
            """)

    @Argument(help: "A YouTube, X, bilibili, Xiaoyuzhou or Apple Podcasts link, or a bilibili id (BV…, av…, ep…, ss…).")
    var url: String?

    @OptionGroup var flags: DownloadFlags

    mutating func run() async throws {
        try await execute(command: "download", url: url, flags: flags)
    }
}

struct Info: AsyncParsableCommand {
    static let configuration = CommandConfiguration(
        abstract: "Show the item, its pages and their streams; download nothing.",
        discussion: """
            Streams are listed in the order haul would choose them under the same -q / -c / --*-ascending options; \
            the index of each is what --video-stream / --audio-stream take, so pass the same options to both. \
            For an item with several pages (a playlist, season, show, multi-video post) info lists the pages only; \
            -p <n> lists the streams and subtitles of page n, -p ALL of every page. Download options are accepted \
            and ignored. With --json, stdout gets one JSON document.
            """)

    @Argument(help: "A YouTube, X, bilibili, Xiaoyuzhou or Apple Podcasts link, or a bilibili id (BV…, av…, ep…, ss…).")
    var url: String

    @Flag(name: .long, help: "Include each stream's download URL.")
    var urls = false

    @OptionGroup var flags: DownloadFlags

    mutating func run() async throws {
        let urls = urls
        try await execute(command: "info", url: url, flags: flags) {
            $0.onlyShowInfo = true
            if urls { $0.showURLs = true }
        }
    }
}

struct Templates: AsyncParsableCommand {
    static let configuration = CommandConfiguration(abstract: "List the variables of -o / --multi-output file-name templates.")

    mutating func run() async throws {
        Log.plain("File-name template variables (-o, --multi-output):")
        for (k, v) in FilePattern.variables { Log.plain("  <\(k)>".padding(toLength: 24, withPad: " ", startingAt: 0) + v) }
        Log.plain("  <publishDate:yyyy-MM-dd>  a date in any format; also <pageDate:…>")
        Log.plain()
        Log.plain("Single page default:   \(FilePattern.singleDefault)")
        Log.plain("Several pages default: \(FilePattern.multiDefault)")
        Log.plain("The extension (.mp4, .m4a, .mp3) is added; a / makes folders.")
    }
}

struct Login: AsyncParsableCommand {
    enum LoginSite: String, ExpressibleByArgument, CaseIterable {
        case bilibili
    }

    static let configuration = CommandConfiguration(
        abstract: "Log in to a site. bilibili: higher qualities and members-only content.",
        discussion: """
            haul login bilibili              scan a QR code with the bilibili app
            haul login bilibili --from-edge  reuse Microsoft Edge's login
                                             (--from-chrome; --profile "Profile 1")
            haul login bilibili --tv         a TV access token (--api tv / app)

            Reading a browser's login needs Full Disk Access for the terminal;
            macOS asks once for the keychain. The login is saved under
            \(Terminal.prettyPath(Storage.home.path))/.
            """)

    @Argument(help: "The site: bilibili.")
    var site: LoginSite

    @Flag(name: .long, help: "Log in the TV account instead and save its access token.")
    var tv = false

    @Flag(name: .customLong("from-edge"), help: "Copy the login from Microsoft Edge.")
    var fromEdge = false

    @Flag(name: .customLong("from-chrome"), help: "Copy the login from Google Chrome.")
    var fromChrome = false

    @Option(name: .long, help: "Browser profile. Default: Default; e.g. \"Profile 1\".")
    var profile: String = "Default"

    @Flag(name: .long, help: "Debug log.")
    var debug = false

    mutating func run() async throws {
        Log.debugEnabled = debug
        printBanner()
        do {
            if fromEdge || fromChrome {
                try await importFromBrowser(fromEdge ? .edge : .chrome)
            } else if tv {
                try await QRLogin.loginTV()
            } else {
                try await QRLogin.loginWeb()
            }
        } catch {
            Log.error(Log.debugEnabled ? "\(error)" : error.readableMessage)
            throw ExitCode(error.haulKind.exitCode)
        }
    }

    private func importFromBrowser(_ browser: BrowserCookies.Browser) async throws {
        Log.status("Reading the bilibili cookie of \(browser.rawValue) (\(profile))")
        let cookie = try BrowserCookies.bilibiliCookieHeader(from: browser, profile: profile)
        var session = Session()
        session.cookie = cookie
        Log.status("Checking the login")
        let nav = try await Auth.nav(session: session)
        guard nav.isLoggedIn else {
            throw HaulError.auth("The browser's bilibili login has expired; log in again there and retry")
        }
        try Storage.write(Storage.cookieFile, cookie)
        let uid = cookie.split(separator: ";").map { $0.trimmingCharacters(in: .whitespaces) }
            .first { $0.hasPrefix("DedeUserID=") }?.dropFirst("DedeUserID=".count) ?? ""
        Log.success("Logged in  uid \(uid)  " + Log.style.dim("cookie saved to \(Terminal.prettyPath(Storage.url(Storage.cookieFile).path))"))
    }
}
