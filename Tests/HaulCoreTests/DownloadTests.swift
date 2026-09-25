import Foundation
import Testing
@testable import HaulCore

@Suite struct DownloaderTests {
    @Test func clipsCoverTheFileWithoutGaps() {
        let clips = Downloader.makeClips(50 * 1024 * 1024)
        #expect(clips.count == 3)
        #expect(clips[0].from == 0 && clips[0].to == 20 * 1024 * 1024)
        #expect(clips[1].from == 20 * 1024 * 1024 + 1)
        #expect(clips[2].to == nil)
        #expect(Downloader.makeClips(10).count == 1)
        #expect(Downloader.makeClips(0).isEmpty)
    }

    /// Every byte in exactly one clip, in order, no inverted ranges — for any size, including a small remainder
    /// after many clips (which used to put the last clip past the end of the file).
    @Test func clipsTileEveryFileSize() {
        for clipSize in Int64(1)...12 {
            for size in Int64(1)...200 {
                for closed in [false, true] {
                    let clips = Downloader.makeClips(size, clipSize: clipSize, closed: closed)
                    var next: Int64 = 0
                    for (i, c) in clips.enumerated() {
                        let end = c.to ?? size - 1
                        #expect(c.index == i && c.from == next && end >= c.from && end - c.from <= clipSize,
                                "size \(size) clip \(clipSize): \(c.from)-\(String(describing: c.to))")
                        #expect(closed || i == clips.count - 1 ? true : c.to != nil)
                        next = end + 1
                    }
                    #expect(next == size, "size \(size) clip \(clipSize)")
                    #expect(clips.last?.to == (closed ? size - 1 : nil))
                }
            }
        }
    }

    @Test func forceHTTPLeavesMcdnAlone() {
        #expect(Downloader.forceHTTP("https://upos-sz.bilivideo.com/a") == "http://upos-sz.bilivideo.com/a")
        #expect(Downloader.forceHTTP("https://x.mcdn.bilivideo.cn:4483/a") == "https://x.mcdn.bilivideo.cn:4483/a")
    }
}

@Suite struct Aria2cTests {
    @Test func argumentSplitting() {
        #expect(Aria2c.splitArguments(#"--all-proxy="http://127.0.0.1:7890" -x 4 'a b'"#) == ["--all-proxy=http://127.0.0.1:7890", "-x", "4", "a b"])
        #expect(Aria2c.splitArguments("") == [])
        #expect(Aria2c.splitArguments(#"a\ b c"#) == ["a b", "c"])
    }

    @Test func headersMatchTheBuiltInDownloader() {
        var bili = DownloadConfig()
        bili.cookie = "SESSDATA=x"
        let a = Aria2c.arguments("https://upos.bilivideo.com/v.m4s", to: "d/v.mp4", config: bili)
        #expect(a.contains("--header=Referer: https://www.bilibili.com") && a.contains("--header=User-Agent: Mozilla/5.0"))
        #expect(a.contains("--header=Cookie: SESSDATA=x"))
        #expect(a.suffix(5) == ["https://upos.bilivideo.com/v.m4s", "-d", "d", "-o", "v.mp4"])

        var other = DownloadConfig()
        other.headers = ["User-Agent": "UA-X", "Accept": "*/*"]
        let b = Aria2c.arguments("https://video.twimg.com/a.mp4", to: "v.mp4", config: other)
        #expect(b.filter { $0.hasPrefix("--header=") } == ["--header=Accept: */*", "--header=User-Agent: UA-X"])
        #expect(b.suffix(4) == ["-d", ".", "-o", "v.mp4"])
    }
}
