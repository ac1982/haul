import Foundation
import Testing
@testable import HaulCore

@Suite struct GzipAndGRPCTests {
    @Test func gzipRoundTrip() throws {
        let payload = Data(String(repeating: "bilibili ", count: 1000).utf8)
        let packed = try Gzip.compress(payload)
        #expect(packed.count < payload.count)
        #expect(packed[packed.startIndex] == 0x1F && packed[packed.startIndex + 1] == 0x8B)
        #expect(try Gzip.decompress(packed) == payload)
    }

    @Test func grpcFrameRoundTrip() throws {
        let message = Data([1, 2, 3, 4, 5, 6, 7, 8, 9])
        let frame = try GRPCFrame.pack(message)
        #expect(frame[frame.startIndex] == 1)
        let len = frame[(frame.startIndex + 1)..<(frame.startIndex + 5)].withUnsafeBytes { UInt32(bigEndian: $0.loadUnaligned(as: UInt32.self)) }
        #expect(Int(len) == frame.count - 5)
        #expect(try GRPCFrame.unpack(frame) == message)
    }

    @Test func grpcUncompressedFrame() throws {
        var frame = Data([0, 0, 0, 0, 3])
        frame.append(contentsOf: [9, 8, 7, 0xFF])  // trailing byte beyond the declared length is ignored
        #expect(try GRPCFrame.unpack(frame) == Data([9, 8, 7]))
    }
}

@Suite struct JSONTests {
    @Test func traversal() throws {
        let j = try JSON.parse(#"{"data":{"list":[{"id":42,"name":"x","ratio":1.5,"ok":true,"none":null}],"count":"7"}}"#)
        #expect(j["data"]["list"][0]["id"].int == 42)
        #expect(j["data"]["list"][0]["id"].stringValue == "42")
        #expect(j["data"]["list"][0]["ratio"].double == 1.5)
        #expect(j["data"]["list"][0]["ok"].bool == true)
        #expect(j["data"]["list"][0]["none"].isNull)
        #expect(j["data"]["count"].int64 == 7)
        #expect(j["missing"]["deep"].isNull)
        #expect(throws: JSONError.self) { try j.get("nope") }
        #expect(j["data"]["list"].stringValue.contains("\"id\":42"))
    }

    @Test func bigIntegers() throws {
        let j = try JSON.parse(#"{"aid":1145141919810,"ts":1700000000}"#)
        #expect(j["aid"].int64 == 1_145_141_919_810)
        #expect(j["ts"].stringValue == "1700000000")
    }
}

@Suite struct FormatTests {
    @Test func sizesAndDurations() {
        #expect(Format.fileSize(512) == "512 bytes")
        #expect(Format.fileSize(1536) == "1.50 KB")
        #expect(Format.fileSize(3 * 1024 * 1024) == "3.00 MB")
        #expect(Format.duration(65) == "01m05s")
        #expect(Format.duration(3661) == "1h01m01s")
        #expect(Format.duration(3661, absolute: true) == "01:01:01")
        #expect(Format.duration(90000, absolute: true) == "25:00:00")
    }

    @Test func fileNames() {
        #expect(Format.validFileName("a/b:c*d?e\"f<g>h|i\\j") == "a_b_c_d_e_f_g_h_i_j")
        #expect(Format.cleanName(" name... ") == "name")
    }

    @Test func query() {
        #expect(Format.queryValue("p", in: "https://x/y?p=3&q=4") == "3")
        #expect(Format.queryValue("q", in: "https://x/y?p=3&q=4#frag") == "4")
        #expect(Format.queryValue("z", in: "https://x/y?p=3") == "")
    }
}

@Suite struct TerminalTests {
    @Test func displayWidthCountsCJKDouble() {
        #expect(Terminal.displayWidth("abc") == 3)
        #expect(Terminal.displayWidth("4K 超清") == 7)
        #expect(Terminal.displayWidth("1080P 高帧率") == 12)
        #expect(Terminal.displayWidth("é") == 1)
    }

    @Test func padAndTruncateMeasureCells() {
        #expect(Terminal.pad("超", 4) == "超  ")
        #expect(Terminal.pad("ab", 4, right: true) == "  ab")
        #expect(Terminal.pad("toolong", 3) == "toolong")
        #expect(Terminal.truncate("哔哩哔哩下载", to: 5) == "哔哩…")
        #expect(Terminal.truncate("short", to: 10) == "short")
    }

    @Test func prettyPathUsesTilde() {
        let home = FileManager.default.homeDirectoryForCurrentUser.path
        #expect(Terminal.prettyPath(home + "/Movies/a.mp4") == "~/Movies/a.mp4")
        #expect(Terminal.prettyPath("/tmp/x.mp4") == "/tmp/x.mp4")
        #expect(Terminal.prettyPath("rel/x.mp4").hasSuffix("/rel/x.mp4"))
    }

    @Test func plainStyleEmitsNoEscapes() {
        let s = Style.plain
        #expect(s.bold("a") == "a" && s.cyan("b") == "b" && s.dim("c") == "c")
        #expect(Style(color: true).red("x") == "\u{1B}[31mx\u{1B}[0m")
    }
}
