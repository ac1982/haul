import Foundation
import Testing
@testable import HaulCore

@Suite struct FilePatternTests {
    let page = Page(index: 3, aid: "170001", cid: "279786", epid: "", title: "第三话/标题?", duration: 100, resolution: "1920x1080",
                    pubTime: 0, cover: nil, desc: nil, ownerName: "UP:主", ownerMid: "1")
    var video: VideoTrack { VideoTrack(id: "80", quality: "1080P", baseURL: "", resolution: "1920x1080", fps: "30", codec: "HEVC", bandwidth: 1200, duration: 100) }

    @Test func multiPageDefault() {
        let ctx = FilePattern.Context(title: "标题.", video: video, audio: nil, page: page, pageCount: 12, apiType: "WEB", pubTime: 0)
        #expect(FilePattern.render(FilePattern.multiDefault, ctx) == "标题/[P03]第三话_标题_.mp4")
    }

    @Test func variables() {
        let ctx = FilePattern.Context(title: "T", video: video, audio: nil, page: page, pageCount: 3, apiType: "TV", pubTime: 0)
        let out = FilePattern.render("<uploader>-<bvid>-<aid>-<cid>-<quality>-<resolution>-<videoCodec>-<videoBitrate>-<api>-<id>-<site>-<unknown>", ctx)
        #expect(out == "UP_主-BV17x411w7KC-170001-279786-1080P-1920x1080-HEVC-1200-TV-BV17x411w7KC-bilibili-<unknown>.mp4")
    }

    @Test func keepsExplicitExtension() {
        let ctx = FilePattern.Context(title: "T", video: nil, audio: nil, page: page, pageCount: 1, apiType: "WEB", pubTime: 0)
        #expect(FilePattern.render("out/<title>.mp4", ctx) == "out/T.mp4")
        #expect(FilePattern.changeExtension("a/b.mp4", to: "xml") == "a/b.xml")
    }
}

@Suite struct PageSelectionTests {
    func info(_ n: Int, index: String? = nil) -> VideoInfo {
        var i = VideoInfo(title: "t", desc: "", cover: "", pubTime: 0,
                          pages: (1...n).map { Page(index: $0, aid: "1", cid: "\($0)", epid: "", title: "p\($0)", duration: 0, resolution: "", pubTime: 0) })
        i.index = index
        return i
    }

    @Test func specs() throws {
        #expect(try PageSelection.parse("ALL", info: info(5), input: "") == nil)
        #expect(try PageSelection.parse("", info: info(5), input: "") == nil)
        #expect(try PageSelection.parse("8", info: info(9), input: "") == [8])
        #expect(try PageSelection.parse("1,2", info: info(5), input: "") == [1, 2])
        #expect(try PageSelection.parse("3-5", info: info(5), input: "") == [3, 4, 5])
        #expect(try PageSelection.parse("last", info: info(5), input: "") == [5])
        #expect(try PageSelection.parse("3,5,LATEST", info: info(9), input: "") == [3, 5, 9])
        #expect(try PageSelection.parse("1-2,4", info: info(5), input: "") == [1, 2, 4])
    }

    @Test func badSpecsAreInputErrors() {
        for spec in ["x", "5-3", "1,,a"] {
            #expect { try PageSelection.parse(spec, info: info(5), input: "") } throws: { ($0 as? HaulError)?.kind == .input }
        }
    }

    @Test func implicitChoice() throws {
        #expect(try PageSelection.parse("", info: info(5, index: "4"), input: "") == [4])
        #expect(try PageSelection.parse("", info: info(5), input: "https://www.bilibili.com/video/BV1?p=2") == [2])
    }
}

@Suite struct OptionsTests {
    @Test func priorities() {
        let (codec, first) = DownloadPipeline.parseCodecPriority("hevc, e-ac-3,av1,hevc")
        #expect(codec == ["HEVC": 0, "EAC3": 1, "AV1": 2])
        #expect(first == "HEVC")
        #expect(DownloadPipeline.parseQualityPriority("1080P+，4k") == ["1080P+": 0, "4K": 1])
        #expect(DownloadPipeline.parseCodecPriority(nil).0.isEmpty)
    }

    @Test func decodeWithDefaults() throws {
        let o = try JSONDecoder().decode(DownloadOptions.self, from: Data(#"{"api":"tv","pages":"ALL","danmakuFormats":["ass"]}"#.utf8))
        #expect(o.api == .tv)
        #expect(o.pages == "ALL")
        #expect(o.danmakuFormats == [.ass])
        #expect(o.multiThread == true && o.forceHTTP == true && o.host == "api.bilibili.com")
    }

    @Test func normalizeResolvesConflicts() {
        var o = DownloadOptions()
        o.interactive = true; o.hideStreams = true
        o.audioOnly = true; o.videoOnly = true
        o.skipSubtitle = true; o.subtitleOnly = true
        o.normalize()
        #expect(o.hideStreams == false && o.audioOnly == false && o.videoOnly == false && o.subtitleOnly == false)
    }
}
