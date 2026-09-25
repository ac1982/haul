import Foundation
import Testing
@testable import HaulCore

@Suite struct UIRenderingTests {
    let plain = Style.plain

    func video(_ id: String, _ q: String, _ codec: String, res: String = "1920x1080", fps: String = "30.000", bw: Int64 = 1000) -> VideoTrack {
        VideoTrack(id: id, quality: q, baseURL: "https://cdn/\(id)-\(codec)", resolution: res, fps: fps, codec: codec, bandwidth: bw, duration: 100)
    }

    var parsed: ParsedTracks {
        var p = ParsedTracks()
        p.video = [video("120", "4K", "HEVC", res: "3840x2160", fps: "60.000", bw: 4493),
                   video("80", "1080P", "HEVC"),
                   video("120", "4K", "AVC", res: "3840x2160", fps: "60.000", bw: 11450),
                   video("80", "1080P", "AV1")]
        p.audio = [AudioTrack(id: "30280", quality: "30280", baseURL: "https://cdn/a", codec: "M4A", bandwidth: 172, duration: 100),
                   AudioTrack(id: "30216", quality: "30216", baseURL: "https://cdn/b", codec: "M4A", bandwidth: 66, duration: 100)]
        return p
    }

    @Test func formatting() {
        #expect(UI.fps("60.000") == "60fps")
        #expect(UI.fps("29.412") == "29.4fps")
        #expect(UI.fps(nil) == "")
        #expect(UI.resolution("1920x1080") == "1920×1080")
        #expect(UI.eta(7) == "7s")
        #expect(UI.eta(65) == "1m05s")
        #expect(UI.eta(3700) == "1h01m")
        #expect(UI.eta(.infinity) == "--")
    }

    @Test func streamTableAlignsAndMarksSelection() {
        let lines = UI.streamTable(parsed, selectedVideo: 0, selectedAudio: 1, pageDuration: 100, collapse: false, showURLs: false, style: plain)
        let videoRows = lines.filter { $0.contains("HEVC") || $0.contains("AVC") || $0.contains("AV1") }
        #expect(videoRows.count == 4)
        #expect(Set(videoRows.map(Terminal.displayWidth)).count == 1, "every video row has the same width")
        #expect(videoRows[0].hasPrefix("  ▶ 0  4K"))
        #expect(videoRows[1].hasPrefix("    1  1080P"))
        #expect(videoRows[0].contains("3840×2160") && videoRows[0].contains("60fps") && videoRows[0].contains("4493 kbps"))
        let audioRows = lines.filter { $0.contains("M4A") }
        #expect(audioRows[1].hasPrefix("  ▶ 1  M4A"))
        #expect(!lines.joined().contains("\u{1B}"))
    }

    @Test func streamTableCollapsesOtherCodecs() {
        let lines = UI.streamTable(parsed, selectedVideo: 0, selectedAudio: 0, pageDuration: 100, collapse: true, showURLs: false, style: plain)
        let videoRows = lines.filter { $0.contains(" kbps") && $0.contains("×") }
        #expect(videoRows.count == 2, "only the selected codec's rows stay")
        #expect(lines.contains { $0.contains("2 more AV1 / AVC streams") })
    }

    @Test func streamTableShowsURLsOnRequest() {
        let lines = UI.streamTable(parsed, selectedVideo: 0, selectedAudio: 0, pageDuration: 100, collapse: false, showURLs: true, style: plain)
        #expect(lines.contains { $0.trimmingCharacters(in: .whitespaces) == "https://cdn/120-HEVC" })
    }

    @Test func progressLines() {
        let line = UI.progressLine(label: "video", downloaded: 5 * 1024 * 1024, total: 10 * 1024 * 1024, bytesPerSecond: 1024 * 1024,
                                   spinner: "|", terminalWidth: 100, style: plain)
        #expect(line.contains(" 50%"))
        #expect(line.contains("5.0 MB / 10.0 MB"))
        #expect(line.contains("1.0 MB/s"))
        #expect(line.contains("5s left"))
        #expect(line.contains("████████████░░░░░░░░░░░░"))

        let unknown = UI.progressLine(label: "cover", downloaded: 2048, total: nil, bytesPerSecond: 0, spinner: "⠋", terminalWidth: 80, style: plain)
        #expect(unknown.contains("⠋") && unknown.contains("2.0 KB") && !unknown.contains("%"))

        let narrow = UI.progressLine(label: "video", downloaded: 1, total: 100, bytesPerSecond: 1, spinner: "|", terminalWidth: 40, style: plain)
        #expect(narrow.filter { $0 == "█" || $0 == "░" }.count == 8, "bar shrinks to its minimum on narrow terminals")

        let done = UI.progressDone(label: "video", total: 10 * 1024 * 1024, seconds: 2, style: plain)
        #expect(done.contains("✓") && done.contains("10.0 MB") && done.contains("5.0 MB/s") && done.contains("2s"))
    }

    @Test func summaryAndHeader() {
        let v = video("120", "4K", "HEVC", res: "3840x2160", fps: "60.000", bw: 4493)
        let a = AudioTrack(id: "30280", quality: "30280", baseURL: "", codec: "M4A", bandwidth: 172, duration: 100)
        let lines = UI.summary(path: "/tmp/out.mp4", size: 300 * 1024 * 1024, video: v, audio: a, seconds: 24, style: plain)
        #expect(lines[0] == "✓ Done  /tmp/out.mp4")
        #expect(lines[1].contains("300.0 MB") && lines[1].contains("4K HEVC 60fps + M4A 172 kbps") && lines[1].contains("24s"))

        var info = VideoInfo(title: "标题", desc: "", cover: "", pubTime: 1_700_000_000,
                             pages: [Page(index: 1, aid: "1", cid: "2", epid: "", title: "p1", duration: 518, resolution: "", pubTime: 0, ownerName: "某UP")])
        var header = UI.header(info, loggedIn: true, style: plain)
        #expect(header[0] == "  标题")
        #expect(header[1].contains("uploader 某UP") && header[1].contains("00:08:38") && !header[1].contains("pages") && header[1].contains("logged in"))
        header = UI.header(info, loggedIn: false, style: plain)
        #expect(header[1].contains("logged out"))

        info.pages = (1...8).map { Page(index: $0, aid: "1", cid: "\($0)", epid: "", title: "第\($0)话", duration: 60, resolution: "", pubTime: 0) }
        let list = UI.pageList(info.pages, showAll: false, style: plain)
        #expect(list.count == 6 && list[0].hasPrefix("  P1  第1话") && list[5].contains("3 more"))
        #expect(UI.pageList(info.pages, showAll: true, style: plain).count == 8)
        #expect(UI.pageHeader(index: 3, count: 12, title: "第三话", style: plain) == "  [03/12] 第三话")
    }
}

@Suite struct PickerTests {
    @Test func keyParsing() {
        #expect(Picker.parse([0x1B, 0x5B, 0x41]) == .up)
        #expect(Picker.parse([0x1B, 0x5B, 0x42]) == .down)
        #expect(Picker.parse([UInt8(ascii: "k")]) == .up)
        #expect(Picker.parse([UInt8(ascii: "j")]) == .down)
        #expect(Picker.parse([0x0D]) == .enter)
        #expect(Picker.parse([0x0A]) == .enter)
        #expect(Picker.parse([UInt8(ascii: "q")]) == .cancel)
        #expect(Picker.parse([0x1B]) == .cancel)
        #expect(Picker.parse([UInt8(ascii: "7")]) == .digit(7))
        #expect(Picker.parse([UInt8(ascii: "x")]) == .other)
    }

    @Test func splitsCoalescedReads() {
        #expect(Picker.split([0x1B, 0x5B, 0x42, 0x0D]) == [[0x1B, 0x5B, 0x42], [0x0D]])
        #expect(Picker.split([0x1B, 0x5B, 0x41, 0x1B, 0x5B, 0x41, UInt8(ascii: "3")]) == [[0x1B, 0x5B, 0x41], [0x1B, 0x5B, 0x41], [UInt8(ascii: "3")]])
        #expect(Picker.split([0x1B]) == [[0x1B]])
        #expect(Picker.split([0x1B, 0x5B, 0x31, 0x7E]) == [[0x1B, 0x5B, 0x31, 0x7E]])
        #expect(Picker.split([]).isEmpty)
    }

    @Test func cursorMovementClampsAndJumps() {
        var typed = ""
        #expect(Picker.move(0, key: .up, count: 5, typed: &typed) == 0)
        #expect(Picker.move(4, key: .down, count: 5, typed: &typed) == 4)
        #expect(Picker.move(2, key: .down, count: 5, typed: &typed) == 3)
        #expect(Picker.move(2, key: .last, count: 5, typed: &typed) == 4)
        #expect(Picker.move(2, key: .first, count: 5, typed: &typed) == 0)
        #expect(Picker.move(2, key: .digit(9), count: 5, typed: &typed) == 2, "out of range digit keeps the cursor")
        // 15 rows: "1" then "4" reaches 14
        typed = ""
        #expect(Picker.move(0, key: .digit(1), count: 15, typed: &typed) == 1)
        #expect(Picker.move(1, key: .digit(4), count: 15, typed: &typed) == 14)
        #expect(typed == "", "buffer clears once it is as long as the largest index")
        #expect(Picker.move(0, key: .other, count: 15, typed: &typed) == 0)
    }
}
