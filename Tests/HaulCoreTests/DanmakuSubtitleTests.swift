import Foundation
import Testing
@testable import HaulCore

@Suite struct DanmakuTests {
    @Test func colourAndTime() {
        #expect(Danmaku.assColor("FE0302") == "0203FE")
        #expect(Danmaku.assTime(65.5) == "0:01:05.50")
        #expect(Danmaku.assTime(3725.0) == "1:02:05.00")
    }

    @Test func parseAndWrite() throws {
        let xml = """
        <?xml version="1.0" encoding="UTF-8"?><i><chatserver>chat.bilibili.com</chatserver>
        <d p="1.5,1,25,16646914,1700000000,0,abc,1">第一条</d>
        <d p="2.0,5,25,16777215,1700000000,0,abc,2">顶部</d>
        <d p="bad">忽略</d>
        </i>
        """
        let dir = FileManager.default.temporaryDirectory.appendingPathComponent("bilidl-\(UUID().uuidString)")
        try FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        defer { try? FileManager.default.removeItem(at: dir) }
        let xmlPath = dir.appendingPathComponent("d.xml").path
        try xml.write(toFile: xmlPath, atomically: true, encoding: .utf8)
        let items = try #require(Danmaku.parseXML(at: xmlPath))
        #expect(items.count == 2)
        #expect(items[0].color == "FE0302" && items[0].mode == .scroll)
        #expect(items[1].mode == .top)
        let assPath = dir.appendingPathComponent("d.ass").path
        try Danmaku.writeASS(items, to: assPath)
        let ass = try String(contentsOfFile: assPath, encoding: .utf8)
        #expect(ass.contains("\\move(1920, 0, -120, 0)"))
        #expect(ass.contains("\\c&H0203FE&"))
        #expect(ass.contains("\\an8\\pos(960, 0)}顶部"))
    }
}

@Suite struct SubtitleTests {
    @Test func languageTable() {
        #expect(Subtitles.languageInfo("zh-hans").code == "chi")
        #expect(Subtitles.languageInfo("ai-zh").name.contains("AI"))
        #expect(Subtitles.languageInfo("xx").code == "und")
    }

    @Test func srtTiming() {
        #expect(Subtitles.srtTime(64.13) == "00:01:04,130")
        #expect(Subtitles.srtTime(3661.5) == "01:01:01,500")
    }

    @Test func jsonToSRT() throws {
        let json = #"{"body":[{"from":1.0,"to":2.5,"content":"你好"},{"to":4.0,"content":"再见"}]}"#
        let srt = try Subtitles.srt(fromJSON: json)
        #expect(srt == "1\n00:00:01,000 --> 00:00:02,500\n你好\n\n2\n00:00:00,000 --> 00:00:04,000\n再见\n\n")
    }
}
