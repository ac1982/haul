import Foundation
import Testing
@testable import HaulCore

/// Keeps pipeline tests away from the real `~/.config/haul` (stored cookie, archive).
enum TestEnv {
    static let home: String = {
        let dir = try! makeTempDir("home")
        setenv("HAUL_HOME", dir, 1)
        return dir
    }()

    static let ffmpeg = Shell.findExecutable("ffmpeg")
    static let ffprobe = Shell.findExecutable("ffprobe")
    static var hasFFmpeg: Bool { ffmpeg != nil && ffprobe != nil }
}

/// Small real media files, made by ffmpeg.
struct TestMedia {
    let video: Data      // H.264 only
    let audio: Data      // AAC only
    let combined: Data   // H.264 + AAC in one file
    let cover: Data      // JPEG

    static func make() async throws -> TestMedia {
        let dir = try makeTempDir("media")
        let ffmpeg = try #require(TestEnv.ffmpeg)
        func run(_ args: [String]) async throws {
            let r = try await Shell.run(ffmpeg, ["-v", "error", "-y"] + args, echo: false, capture: true)
            #expect(r.status == 0, "ffmpeg \(args): \(r.errors)")
        }
        let v = ["-f", "lavfi", "-i", "testsrc=size=64x64:rate=10", "-t", "1"]
        let a = ["-f", "lavfi", "-i", "sine=frequency=440:duration=1"]
        try await run(v + ["-c:v", "libx264", "-pix_fmt", "yuv420p", "\(dir)/v.mp4"])
        try await run(a + ["-c:a", "aac", "\(dir)/a.m4a"])
        try await run(v + a + ["-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac", "-shortest", "\(dir)/av.mp4"])
        try await run(["-f", "lavfi", "-i", "color=red:size=32x32", "-frames:v", "1", "\(dir)/c.jpg"])
        func read(_ name: String) throws -> Data { try Data(contentsOf: URL(fileURLWithPath: "\(dir)/\(name)")) }
        return TestMedia(video: try read("v.mp4"), audio: try read("a.m4a"), combined: try read("av.mp4"), cover: try read("c.jpg"))
    }
}

/// What ffprobe sees in a file.
struct Probe {
    let streams: [String]   // "video:h264", "audio:aac", …; not the data track ffmpeg adds for chapters
    let chapters: [String]
    let tags: [String: String]

    init(_ path: String) async throws {
        let r = try await Shell.run(try #require(TestEnv.ffprobe),
                                    ["-v", "error", "-show_entries", "stream=codec_type,codec_name:format_tags", "-show_chapters", "-of", "json", path],
                                    echo: false, capture: true)
        let json = try JSON.parse(r.output)
        streams = json["streams"].array.filter { $0["codec_type"].string != "data" }
            .map { "\($0["codec_type"].stringValue):\($0["codec_name"].stringValue)" }
        chapters = json["chapters"].array.map { $0["tags"]["title"].stringValue }
        tags = json["format"]["tags"].object.compactMapValues(\.string)
    }
}
