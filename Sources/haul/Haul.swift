import ArgumentParser
import HaulCore
import Foundation

let toolVersion = "1.0.0"

@main
struct Haul: AsyncParsableCommand {
    static let configuration = CommandConfiguration(
        commandName: "haul",
        abstract: "Download video and audio from YouTube, X, bilibili, Xiaoyuzhou and Apple Podcasts.",
        usage: "haul <url> [options]  |  haul <subcommand> [options]",
        discussion: Self.guide,
        version: toolVersion,
        subcommands: [Download.self, Info.self, Login.self, Templates.self],
        defaultSubcommand: Download.self
    )

    static let guide = """
        SITES
        \(Site.allCases.map { "  \($0.name)\n      \($0.linkForms)" }.joined(separator: "\n"))
          YouTube and X need yt-dlp and deno: brew install yt-dlp deno
          Muxing needs ffmpeg: brew install ffmpeg

        FOR SCRIPTS AND AI AGENTS
          haul info <url> --json   inspect first: the item, its pages, and
                                   each page's streams with their indexes
          haul <url> --json        download; prints one JSON document with
                                   the files written
          With --json, stdout carries only that document; progress and logs
          go to stderr. haul never prompts without a terminal: choose streams
          with --video-stream N / --audio-stream N, or let -q / -c decide.

        EXIT CODES
          0 done              1 download failed    2 bad link or option
          3 missing tool      4 login needed       64 bad command line
          130 cancelled

        EXAMPLES
          haul "https://youtu.be/DdCEmlAydcw"
          haul -q 720p -c avc,m4a "https://youtu.be/DdCEmlAydcw"
          haul --audio-only "https://youtu.be/DdCEmlAydcw"
          haul -p ALL "https://x.com/<user>/status/<id>"
          haul -p 1-3 -w ~/Movies "https://www.bilibili.com/video/BV1qt4y1X7TW"
          haul -p LATEST "https://podcasts.apple.com/us/podcast/id1200361736"
          haul info --json --urls "https://youtu.be/DdCEmlAydcw"

        FILES
          \(Terminal.prettyPath(Storage.home.path))/ (or $HAUL_HOME) holds config.json (defaults for
          any option, camelCase keys), the bilibili login and the archive.
          Files are saved in the current directory, or -w <dir>.

          `haul download --help` lists every option; `--help-hidden` adds
          the rarely needed bilibili network options.
        """
}

/// Runs a pipeline for `download` and `info`: human output, or with `--json` one JSON document on stdout.
func execute(command: String, url: String?, flags: DownloadFlags, adjust: (inout DownloadOptions) -> Void = { _ in }) async throws {
    let json = flags.general.json
    if json { Log.toStderr = true }
    let report = Report(command: command)
    do {
        guard let url, !url.isEmpty else { throw HaulError.input("No link given. Usage: haul \(command == "info" ? "info " : "")<url>; see haul --help") }
        if !json { printBanner() }
        var options = try flags.resolve(url: url)
        adjust(&options)
        let pipeline = try await DownloadPipeline(options: options, report: report)
        try await pipeline.run()
        if json, let doc = report.json() { Log.output(doc) }
    } catch {
        if json { Log.output(report.failureJSON(error, input: url)) }
        Log.error(Log.debugEnabled ? "\(error)" : error.readableMessage)
        throw ExitCode(error.haulKind.exitCode)
    }
}

func printBanner() {
    Log.banner("haul " + Log.style.dim("v\(toolVersion)"))
    Log.plain()
}
