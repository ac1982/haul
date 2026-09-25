import Foundation
import Testing
@testable import HaulCore

/// bilibili link resolution, info and playurl parsing, against responses captured from the live API (logged out).
@Suite struct BilibiliParsingTests {
    let http = Stub.client
    let session = Session()

    init() throws {
        try Stub.fixture("x/web-interface/view?aid=626497566", "view-single.json")
        try Stub.fixture("x/web-interface/view?aid=246993280", "view-multi.json")
        try Stub.fixture("pgc/view/web/season?ep_id=317690", "pgc-season.json")
        try Stub.fixture("pgc/view/web/season?season_id=33073", "pgc-season.json")
        try Stub.fixture("x/player/wbi/playurl?", "playurl-web-dash.json")
        try Stub.fixture("pgc/player/web/v2/playurl?", "pgc-playurl.json")
        // A plain av page that is not licensed content stays where it is.
        Stub.on("www.bilibili.com/video/av") { _ in .status(200) }
    }

    // MARK: links

    @Test func resolvesVideoLinksAndIds() async throws {
        #expect(try await IDResolver.resolve("BV1qt4y1X7TW", http: http, session: session) == .video(aid: "626497566"))
        #expect(try await IDResolver.resolve("av170001", http: http, session: session) == .video(aid: "170001"))
        #expect(try await IDResolver.resolve("https://www.bilibili.com/video/BV1qt4y1X7TW/?p=2", http: http, session: session)
                == .video(aid: "626497566"))
        #expect(try await IDResolver.resolve("https://www.bilibili.com/video/av170001", http: http, session: session) == .video(aid: "170001"))
    }

    @Test func licensedVideoRedirectsToItsEpisode() async throws {
        Stub.on("www.bilibili.com/video/av840009001") { _ in StubResponse(redirect: "https://www.bilibili.com/bangumi/play/ep317690") }
        Stub.on("www.bilibili.com/bangumi/play/ep317690") { _ in .status(200) }
        #expect(try await IDResolver.resolve("av840009001", http: http, session: session) == .episode(epId: "317690"))
    }

    @Test func expandsShortLinks() async throws {
        Stub.on("b23.tv/AbCdEf1") { _ in StubResponse(redirect: "https://www.bilibili.com/video/BV1qt4y1X7TW?share=1") }
        Stub.on("www.bilibili.com/video/BV1qt4y1X7TW?share=1") { _ in .status(200) }
        #expect(try await IDResolver.resolve("https://b23.tv/AbCdEf1", http: http, session: session) == .video(aid: "626497566"))
    }

    @Test func resolvesSeasonsCoursesAndLists() async throws {
        func resolve(_ s: String) async throws -> MediaID { try await IDResolver.resolve(s, http: http, session: session) }
        #expect(try await resolve("ep317690") == .episode(epId: "317690"))
        #expect(try await resolve("https://www.bilibili.com/bangumi/play/ep317690?from=search") == .episode(epId: "317690"))
        // A season resolves to its first episode.
        #expect(try await resolve("ss33073") == .episode(epId: "317690"))
        #expect(try await resolve("https://www.bilibili.com/bangumi/play/ss33073") == .episode(epId: "317690"))
        #expect(try await resolve("cheese/ep1234") == .cheese(epId: "1234"))
        #expect(try await resolve("https://www.bilibili.com/cheese/play/ep1234") == .cheese(epId: "1234"))
        #expect(try await resolve("https://space.bilibili.com/403748305") == .space(mid: "403748305"))
        #expect(try await resolve("https://space.bilibili.com/403748305/favlist?fid=66&ftype=create") == .favorites(fid: "66", mid: "403748305"))
        #expect(try await resolve("https://space.bilibili.com/403748305/lists/4242?type=season") == .collection(bizId: "4242"))
        #expect(try await resolve("https://space.bilibili.com/403748305/lists/77?type=series") == .series(bizId: "77"))
        #expect(try await resolve("https://space.bilibili.com/1/channel/collectiondetail?sid=9") == .collection(bizId: "9"))
        #expect(try await resolve("https://www.bilibili.com/medialist/play/1?business=space_series&business_id=5") == .series(bizId: "5"))
        await #expect(throws: HaulError.self) { try await resolve("not a link") }
    }

    // MARK: info

    @Test func singlePageVideo() async throws {
        let info = try await InfoFetcher.fetch(.video(aid: "626497566"), useIntl: false, http: http, session: session)
        #expect(info.title == "【4K60帧】咬人猫最新单曲《dududu》有没有戳中你心~【BML2020单品】")
        #expect(info.pubTime == 1595684326 && !info.isBangumi && !info.isInteractive)
        #expect(info.cover.hasPrefix("http"))
        let page = try #require(info.pages.first)
        #expect(info.pages.count == 1)
        #expect(page.index == 1 && page.aid == "626497566" && page.cid == "220355130" && page.title == "dududu")
        #expect(page.duration == 226 && page.resolution == "3840x2160")
        #expect(page.ownerName == "BML制作指挥部" && page.ownerMid == "403748305")
        #expect(page.bvid == "BV1qt4y1X7TW")
    }

    @Test func multiPageVideo() async throws {
        let info = try await InfoFetcher.fetch(.video(aid: "246993280"), useIntl: false, http: http, session: session)
        #expect(info.title == "(强推)李宏毅2021/2022春机器学习课程")
        #expect(info.pages.map(\.index) == [1, 2, 3, 4, 5])
        #expect(info.pages.map(\.cid) == ["513702855", "303819228", "303813984", "303814306", "513703312"])
        #expect(info.pages[2].title == "第一节 2021 - (上) - 机器学习基本概念简介")
        #expect(info.pages.allSatisfy { $0.aid == "246993280" && $0.ownerName == "啥都会一点的研究生" })
    }

    @Test func bangumiSeason() async throws {
        let info = try await InfoFetcher.fetch(.episode(epId: "317690"), useIntl: false, http: http, session: session)
        #expect(info.title == "女学。～圣女斯克威尔学院～")
        #expect(info.isBangumi && info.isBangumiEnd && !info.isCheese)
        #expect(info.pages.count == 50)
        #expect(info.index == "1")
        #expect(info.pubTime == Format.parseDateTime("2020-04-06 07:05:00"))
        let first = info.pages[0]
        #expect(first.epid == "317690" && first.aid == "840009001" && first.cid == "178175633")
        #expect(first.title == "1 我的梦想从这里开始！托娅的女演员之路")
    }

    @Test func bangumiSkipsTrailersAndFindsSectionEpisodes() async throws {
        Stub.on("pgc/view/web/season?ep_id=9002") { _ in .json(#"""
        {"code": 0, "result": {"title": "番剧", "evaluate": "简介", "cover": "c.jpg", "publish": {"pub_time": "", "is_finish": 0},
          "episodes": [{"id": 9001, "aid": 1, "cid": 11, "title": "1", "long_title": "正片"}],
          "section": [{"title": "PV", "episodes": [
            {"id": 9003, "aid": 3, "cid": 33, "title": "预告", "long_title": "", "badge": "预告"},
            {"id": 9002, "aid": 2, "cid": 22, "title": "PV1", "long_title": "先导", "link": "https://www.bilibili.com/bangumi/play/ep9002"}]}]}}
        """#) }
        let info = try await InfoFetcher.fetch(.episode(epId: "9002"), useIntl: false, http: http, session: session)
        #expect(info.title == "番剧[PV]")
        #expect(info.pages.map(\.epid) == ["9002"])
        #expect(info.pages[0].title == "PV1 先导" && info.index == "1" && !info.isBangumiEnd)
    }

    @Test func course() async throws {
        Stub.on("pugv/view/web/season?ep_id=502") { _ in .json(#"""
        {"code": 0, "data": {"title": " 课程 ", "subtitle": "副标题", "cover": "c.jpg", "up_info": {"uname": "老师", "mid": 7},
          "episodes": [
            {"id": 501, "aid": 10, "cid": 100, "index": 1, "title": "第一课", "duration": 600, "release_date": 1600000000},
            {"id": 502, "aid": 20, "cid": 200, "index": 2, "title": "第二课 ", "duration": 700, "release_date": 1600100000}]}}
        """#) }
        let info = try await InfoFetcher.fetch(.cheese(epId: "502"), useIntl: false, http: http, session: session)
        #expect(info.title == "课程" && info.desc == "副标题" && info.isCheese && info.isBangumi)
        #expect(info.index == "2" && info.pubTime == 1600000000)
        #expect(info.pages.map(\.title) == ["第一课", "第二课"])
        #expect(info.pages[1].epid == "502" && info.pages[1].duration == 700 && info.pages[1].ownerName == "老师")
    }

    @Test func collectionPagesThroughTheList() async throws {
        Stub.on("x/v1/medialist/info?type=8&biz_id=4243") { _ in .json(#"""
        {"code": 0, "data": {"title": "合集", "intro": "介绍", "ctime": 1700000000}}
        """#) }
        Stub.on({ $0.absoluteString.contains("x/v2/medialist/resource/list?type=8&oid=&") && $0.absoluteString.contains("biz_id=4243") }) { _ in .json(#"""
        {"code": 0, "data": {"has_more": true, "media_list": [
          {"id": 1, "title": "一", "page": 1, "attr": 0, "pubtime": 1, "upper": {"name": "UP", "mid": 9}, "pages": [{"id": 11, "page": 1, "duration": 60}]},
          {"id": 2, "title": "失效", "page": 1, "attr": 9, "pages": [{"id": 22, "page": 1}]}]}}
        """#) }
        Stub.on({ $0.absoluteString.contains("resource/list?type=8&oid=1&") && $0.absoluteString.contains("biz_id=4243") }) { _ in .json(#"""
        {"code": 0, "data": {"has_more": false, "media_list": [
          {"id": 3, "title": "三", "page": 2, "attr": 0, "pages": [{"id": 31, "page": 1, "title": "上"}, {"id": 32, "page": 2, "title": "下"}]}]}}
        """#) }
        let info = try await InfoFetcher.fetch(.collection(bizId: "4243"), useIntl: false, http: http, session: session)
        #expect(info.title == "合集" && info.pubTime == 1700000000)
        #expect(info.pages.map(\.cid) == ["11", "31", "32"])
        #expect(info.pages.map(\.index) == [1, 2, 3])
        #expect(info.pages.map(\.title) == ["一", "三_P1_上", "三_P2_下"])
    }

    // MARK: playurl

    @Test func webDashTracks() async throws {
        var s = session
        s.wbiKey = "ea1db124af3c7062474693fa704f4ff8"
        let r = PlayURLRequest(id: .video(aid: "626497566"), aid: "626497566", cid: "220355130", epId: "", api: .web, encoding: "")
        let t = try await PlayURLClient.extractTracks(r, http: http, session: s)
        #expect(t.isDash && !t.isFLV && !t.videoHasAudio)
        // Two passes (qn=0, then max) are merged without duplicates.
        #expect(t.video.map(\.id) == ["32", "32", "16", "16"])
        #expect(t.video.map(\.codec) == ["AVC", "HEVC", "HEVC", "AVC"])
        #expect(t.video[0].quality == "480P" && t.video[0].resolution == "852x480" && t.video[0].fps == "29.412")
        #expect(t.video[0].bandwidth == 756 && t.video[0].duration == 225)
        #expect(t.audio.map(\.id) == ["30216", "30232", "30280"])
        #expect(t.audio.allSatisfy { $0.codec == "M4A" })
        #expect(t.audio.map(\.bandwidth) == [67, 132, 319])
        // WBI-signed request, both passes.
        let calls = Stub.requests("avid=626497566&cid=220355130")
        #expect(calls.contains { $0.url!.absoluteString.contains("qn=0") } && calls.contains { $0.url!.absoluteString.contains("qn=127") })
        #expect(calls.allSatisfy { $0.url!.absoluteString.contains("w_rid=") })
    }

    @Test func bangumiPreviewIsFLVWithChapters() async throws {
        let r = PlayURLRequest(id: .episode(epId: "317690"), aid: "840009001", cid: "178175633", epId: "317690", api: .web, encoding: "")
        let t = try await PlayURLClient.extractTracks(r, http: http, session: session)
        #expect(t.isFLV && !t.isDash)
        #expect(t.clips.count == 1 && t.clips[0].contains("bilivideo.com"))
        #expect(t.qualities == ["125", "112", "80", "64", "32", "16"])
        #expect(t.video.map(\.id) == ["32"] && t.video[0].codec == "AVC" && t.video[0].size == 10228474)
        // OP / ED markers become chapters with the main part in between.
        #expect(t.extraPoints == [
            ViewPoint(title: "片头", start: 0, end: 24),
            ViewPoint(title: "Main", start: 24, end: 145),
            ViewPoint(title: "片尾", start: 145, end: 169),
        ])
    }

    @Test func pcdnHostsAreAvoided() {
        let node = try! JSON.parse(#"{"base_url": "http://1.2.3.4:4480/v.m4s", "backup_url": ["https://upos-sz-mirror.bilivideo.com/v.m4s"]}"#)
        #expect(PlayURLClient.pickURL(node) == "https://upos-sz-mirror.bilivideo.com/v.m4s")
        let only = try! JSON.parse(#"{"base_url": "http://1.2.3.4:4480/v.m4s"}"#)
        #expect(PlayURLClient.pickURL(only) == "http://1.2.3.4:4480/v.m4s")
    }
}
