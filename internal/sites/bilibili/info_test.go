package bilibili

import (
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ac1982/haul/internal/format"
	"github.com/ac1982/haul/internal/media"
	"github.com/ac1982/haul/internal/testkit"
)

func fetchItem(t *testing.T, e *Extractor, id mediaID, link string) *media.Item {
	t.Helper()
	id, in, err := e.fetchInfo(ctx, plainSession(), id)
	if err != nil {
		t.Fatal(err)
	}
	return buildItem(link, link, id, in, e.opts.API)
}

func titles(item *media.Item) []string {
	var out []string
	for _, e := range item.Entries {
		out = append(out, e.Title)
	}
	return out
}

func fields(item *media.Item, name string) []string {
	var out []string
	for _, e := range item.Entries {
		out = append(out, e.Fields[name])
	}
	return out
}

func TestSinglePageVideo(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stubCaptured(t, stub)
	item := fetchItem(t, e, mediaID{kind: video, id: "626497566"}, "BV1qt4y1X7TW")
	if item.Title != "【4K60帧】咬人猫最新单曲《dududu》有没有戳中你心~【BML2020单品】" || item.Published.Unix() != 1595684326 ||
		!strings.HasPrefix(item.Thumbnail, "http") || item.Collection || item.Focus != 0 {
		t.Fatalf("%+v", item)
	}
	if item.ID != "BV1qt4y1X7TW" || item.URL != "https://www.bilibili.com/video/BV1qt4y1X7TW/" || item.Uploader != (media.Person{Name: "BML制作指挥部", ID: "403748305"}) {
		t.Fatalf("%+v", item)
	}
	if len(item.Entries) != 1 {
		t.Fatalf("%d entries", len(item.Entries))
	}
	p := item.Entries[0]
	if p.Index != 1 || p.ID != "BV1qt4y1X7TW" || p.Title != "dududu" || p.Duration != 226*time.Second || p.URL != "https://www.bilibili.com/video/BV1qt4y1X7TW/" {
		t.Fatalf("%+v", p)
	}
	if p.Uploader.Name != "BML制作指挥部" || p.Thumbnail != item.Thumbnail || p.Description != item.Description || p.Published.Unix() != 1595684326 {
		t.Fatalf("%+v", p)
	}
	want := map[string]string{"bvid": "BV1qt4y1X7TW", "aid": "626497566", "cid": "220355130", "api": "WEB"}
	for k, v := range want {
		if p.Fields[k] != v {
			t.Errorf("field %s = %q", k, p.Fields[k])
		}
	}
	if r := p.Ref.(*ref); *r != (ref{kind: video, aid: "626497566", cid: "220355130", api: Web, position: 1}) {
		t.Fatalf("ref %+v", r)
	}
}

func TestMultiPageVideo(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stubCaptured(t, stub)
	item := fetchItem(t, e, mediaID{kind: video, id: "246993280"}, "https://www.bilibili.com/video/av246993280?p=3")
	if item.Title != "(强推)李宏毅2021/2022春机器学习课程" || len(item.Entries) != 5 || item.Collection {
		t.Fatalf("%q %d", item.Title, len(item.Entries))
	}
	if cids := fields(item, "cid"); !slices.Equal(cids, []string{"513702855", "303819228", "303813984", "303814306", "513703312"}) {
		t.Fatalf("cids %v", cids)
	}
	if item.Entries[2].Title != "第一节 2021 - (上) - 机器学习基本概念简介" || item.Entries[4].Index != 5 {
		t.Fatalf("%+v", item.Entries[2])
	}
	for _, en := range item.Entries {
		if en.Fields["aid"] != "246993280" || en.Uploader.Name != "啥都会一点的研究生" || en.ID != "BV1Wv411h7kN" {
			t.Fatalf("%+v", en)
		}
	}
	// ?p=3 points at the third page.
	if item.Focus != 3 {
		t.Fatalf("focus %d", item.Focus)
	}
}

func TestBangumiSeason(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stubCaptured(t, stub)
	item := fetchItem(t, e, mediaID{kind: episode, id: "317690"}, "ep317690")
	want, _ := format.ParseDateTime("2020-04-06 07:05:00")
	if item.Title != "女学。～圣女斯克威尔学院～" || len(item.Entries) != 50 || item.Focus != 1 || item.Published.Unix() != want {
		t.Fatalf("%q %d %d %v", item.Title, len(item.Entries), item.Focus, item.Published)
	}
	// A finished season is no collection of its own; the engine still folders its 50 entries.
	if item.Collection || item.ID != "ss33073" || item.URL != "https://www.bilibili.com/bangumi/play/ss33073" {
		t.Fatalf("%+v", item)
	}
	first := item.Entries[0]
	if first.ID != "ep317690" || first.Fields["aid"] != "840009001" || first.Fields["cid"] != "178175633" || first.Fields["bvid"] != "BV1X54y1R7aZ" ||
		first.Title != "1 我的梦想从这里开始！托娅的女演员之路" || first.URL != "https://www.bilibili.com/bangumi/play/ep317690" ||
		first.Duration != 170*time.Second || first.Published.Unix() != 1586127900 {
		t.Fatalf("%+v %v", first, first.Fields)
	}
	if r := first.Ref.(*ref); r.kind != episode || r.epid != "317690" {
		t.Fatalf("%+v", r)
	}
}

func TestBangumiSkipsTrailersAndFindsSectionEpisodes(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stub.On("pgc/view/web/season?ep_id=9002", answer(`{"code": 0, "result": {"title": "番剧", "evaluate": "简介", "cover": "c.jpg", "season_id": 90,
	  "publish": {"pub_time": "", "is_finish": 0},
	  "episodes": [{"id": 9001, "aid": 1, "cid": 11, "title": "1", "long_title": "正片"}],
	  "section": [{"title": "PV", "episodes": [
	    {"id": 9003, "aid": 3, "cid": 33, "title": "预告", "long_title": "", "badge": "预告"},
	    {"id": 9002, "aid": 2, "cid": 22, "title": "PV1", "long_title": "先导", "link": "https://www.bilibili.com/bangumi/play/ep9002"}]}]}}`))
	item := fetchItem(t, e, mediaID{kind: episode, id: "9002"}, "ep9002")
	if item.Title != "番剧[PV]" || !slices.Equal(titles(item), []string{"PV1 先导"}) || item.Entries[0].ID != "ep9002" || item.Focus != 1 {
		t.Fatalf("%q %v %d", item.Title, titles(item), item.Focus)
	}
	// An ongoing season is saved as a collection even for one episode.
	if !item.Collection || !item.Published.IsZero() || item.Entries[0].Thumbnail != "c.jpg" || item.Entries[0].Description != "简介" {
		t.Fatalf("%+v", item)
	}
}

func TestCourse(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stub.On("pugv/view/web/season?ep_id=502", answer(`{"code": 0, "data": {"title": " 课程 ", "subtitle": "副标题", "cover": "c.jpg", "season_id": 50,
	  "up_info": {"uname": "老师", "mid": 7},
	  "episodes": [
	    {"id": 501, "aid": 10, "cid": 100, "index": 1, "title": "第一课", "duration": 600, "release_date": 1600000000},
	    {"id": 502, "aid": 20, "cid": 200, "index": 2, "title": "第二课 ", "duration": 700, "release_date": 1600100000}]}}`))
	item := fetchItem(t, e, mediaID{kind: cheese, id: "502"}, "cheese/ep502")
	if item.Title != "课程" || item.Description != "副标题" || item.Focus != 2 || item.Published.Unix() != 1600000000 || item.Collection {
		t.Fatalf("%+v", item)
	}
	if !slices.Equal(titles(item), []string{"第一课", "第二课"}) || item.ID != "ss50" || item.URL != "https://www.bilibili.com/cheese/play/ss50" {
		t.Fatalf("%v %s %s", titles(item), item.ID, item.URL)
	}
	second := item.Entries[1]
	if second.ID != "ep502" || second.Duration != 700*time.Second || second.Uploader.Name != "老师" || second.URL != "https://www.bilibili.com/cheese/play/ep502" {
		t.Fatalf("%+v", second)
	}
	if r := second.Ref.(*ref); r.kind != cheese || r.epid != "502" || r.bangumi() {
		t.Fatalf("%+v", r)
	}
}

func TestAnEpisodeIDThatIsNoBangumiIsTriedAsACourse(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stub.On("pgc/view/web/season?ep_id=777", answer(`{"code": -404, "message": "啥都木有"}`))
	stub.On("pugv/view/web/season?ep_id=777", answer(`{"code": 0, "data": {"title": "课", "episodes": [{"id": 777, "aid": 1, "cid": 2, "index": 1, "title": "一"}]}}`))
	id, in, err := e.fetchInfo(ctx, plainSession(), mediaID{kind: episode, id: "777"})
	if err != nil || id.kind != cheese || in.title != "课" || in.focus != 1 {
		t.Fatalf("%v %+v %v", id, in, err)
	}
	// A video that is not there says why, in bilibili's words.
	stub.On("x/web-interface/view?aid=5", answer(`{"code": -404, "message": "啥都木有"}`))
	if _, _, err := e.fetchInfo(ctx, plainSession(), mediaID{kind: video, id: "5"}); err == nil || !strings.Contains(err.Error(), "-404: 啥都木有") {
		t.Fatalf("err = %v", err)
	}
}

func TestInteractiveVideoBranches(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(func(o *Options) { o.API = TV })
	stub.On("x/web-interface/view?aid=10", answer(`{"code":0,"data":{"bvid":"BV1xx411c7mD","cid":100,"title":"互动","pubdate":5,
	  "owner":{"name":"UP","mid":1},"rights":{"is_stein_gate":1},"pages":[{"page":1,"cid":100,"part":"开始"}]}}`))
	stub.On("x/player.so?bvid=BV1xx411c7mD&id=cid:100", func(*http.Request) testkit.Response {
		return testkit.Text(`<interaction>{&quot;graph_version&quot;:42}</interaction>`)
	})
	stub.On("x/stein/edgeinfo_v2?graph_version=42&bvid=BV1xx411c7mD", answer(`{"data":{"edges":{"questions":[
	  {"choices":[{"cid":101,"option":" 左 "},{"cid":102,"option":"右"}]},{"choices":[{"cid":103,"option":"结局"}]}]}}}`))
	stub.On("www.bilibili.com/video/av", status(200))
	item, err := e.Resolve(ctx, "av10")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(titles(item), []string{"开始", "左", "右", "结局"}) || !slices.Equal(fields(item, "cid"), []string{"100", "101", "102", "103"}) {
		t.Fatalf("%v %v", titles(item), fields(item, "cid"))
	}
	// The TV API has no interactive videos: the web one is used, and the file names say so.
	if item.Entries[3].Fields["api"] != "WEB" || item.Entries[3].Ref.(*ref).api != Web || item.LoggedIn != nil {
		t.Fatalf("%v", item.Entries[3].Fields)
	}
}

func TestLicensedVideoKeepsItsEpisodeForThePlayurl(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stub.On("x/web-interface/view?aid=11", answer(`{"code":0,"data":{"title":"番","redirect_url":"https://www.bilibili.com/bangumi/play/ep99",
	  "pages":[{"page":1,"cid":110,"part":"p"}]}}`))
	item := fetchItem(t, e, mediaID{kind: video, id: "11"}, "av11")
	if r := item.Entries[0].Ref.(*ref); r.epid != "99" || r.kind != video || item.Entries[0].ID != bvid("11") || item.Collection {
		t.Fatalf("%+v", r)
	}
}

func TestCollectionPagesThroughTheList(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stub.On("x/v1/medialist/info?type=8&biz_id=4243", answer(`{"code": 0, "data": {"title": "合集", "intro": "介绍", "ctime": 1700000000, "upper": {"name": "UP", "mid": 9}}}`))
	stub.OnFunc(100, func(u string) bool {
		return strings.Contains(u, "resource/list?type=8&oid=&") && strings.Contains(u, "biz_id=4243")
	}, answer(`
	  {"code": 0, "data": {"has_more": true, "media_list": [
	    {"id": 1, "title": "一", "page": 1, "attr": 0, "pubtime": 1, "upper": {"name": "UP", "mid": 9}, "pages": [{"id": 11, "page": 1, "duration": 60}]},
	    {"id": 2, "title": "失效", "page": 1, "attr": 9, "pages": [{"id": 22, "page": 1}]}]}}`))
	stub.OnFunc(100, func(u string) bool {
		return strings.Contains(u, "resource/list?type=8&oid=2&") && strings.Contains(u, "biz_id=4243")
	}, answer(`
	  {"code": 0, "data": {"has_more": false, "media_list": [
	    {"id": 3, "title": "三", "page": 2, "attr": 0, "pages": [{"id": 31, "page": 1, "title": "上"}, {"id": 32, "page": 2, "title": "下"}]}]}}`))
	item := fetchItem(t, e, mediaID{kind: collection, id: "4243"}, "https://space.bilibili.com/9/lists/4243?type=season")
	if item.Title != "合集" || item.Published.Unix() != 1700000000 || item.Uploader.Name != "UP" || item.ID != "collection:4243" {
		t.Fatalf("%+v", item)
	}
	if !slices.Equal(fields(item, "cid"), []string{"11", "31", "32"}) || !slices.Equal(titles(item), []string{"一", "三_P1_上", "三_P2_下"}) {
		t.Fatalf("%v %v", fields(item, "cid"), titles(item))
	}
	if item.Entries[0].Duration != time.Minute || item.Entries[2].Index != 3 {
		t.Fatalf("%+v", item.Entries[0])
	}
}

func TestACollectionThatIsASeries(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stub.On("x/v1/medialist/info?type=8&biz_id=5", answer(`{"code": -404, "message": "not found"}`))
	stub.On("x/v1/medialist/info?type=5&biz_id=5", answer(`{"code": 0, "data": {"title": "系列"}}`))
	stub.On("resource/list?type=5&oid=&otype=2&biz_id=5&bvid=", answer(`{"code": 0, "data": {"has_more": false, "media_list": [
	  {"id": 1, "title": "一", "page": 1, "pages": [{"id": 11, "page": 1}]}]}}`))
	item := fetchItem(t, e, mediaID{kind: collection, id: "5"}, "l")
	if item.Title != "系列" || len(item.Entries) != 1 {
		t.Fatalf("%+v", item)
	}
	stub.On("x/v1/medialist/info?type=5&biz_id=6", answer(`{"code": -404, "message": "gone"}`))
	stub.On("x/v1/medialist/info?type=8&biz_id=6", answer(`{"code": -404, "message": "gone"}`))
	if _, _, err := e.fetchInfo(ctx, plainSession(), mediaID{kind: collection, id: "6"}); err == nil || !strings.Contains(err.Error(), "collection (code -404): gone") {
		t.Fatalf("err = %v", err)
	}
}

func TestFavoritesList(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stub.On("fav/folder/created/list-all?up_mid=9", answer(`{"code":0,"data":{"list":[{"id":66}]}}`))
	stub.On("fav/resource/list?media_id=66&pn=1&", answer(`{"code":0,"data":{"info":{"title":"收藏","intro":"i","ctime":7,"media_count":21,"upper":{"name":"我","mid":9}},
	  "medias":[{"id":1,"title":"一","page":1,"attr":0,"ugc":{"first_cid":11},"duration":5,"pubtime":3,"cover":"c1","intro":"简介","upper":{"name":"A","mid":2}},
	            {"id":2,"title":"没了","page":1,"attr":9,"ugc":{"first_cid":22}}]}}`))
	stub.On("fav/resource/list?media_id=66&pn=2&", answer(`{"code":0,"data":{"medias":[{"id":246993280,"title":"课","page":5,"attr":0,"intro":"多P"},
	  {"id":1,"title":"一","page":1,"attr":0,"ugc":{"first_cid":11}}]}}`))
	stubCaptured(t, stub)
	item := fetchItem(t, e, mediaID{kind: favorites, mid: "9"}, "https://space.bilibili.com/9/favlist")
	if item.Title != "收藏" || item.ID != "fav66" || item.URL != "https://space.bilibili.com/9/favlist?fid=66" || item.Uploader.Name != "我" || len(item.Entries) != 6 {
		t.Fatalf("%+v %v", item, titles(item))
	}
	if first := item.Entries[0]; first.Title != "一" || first.Fields["cid"] != "11" || first.Thumbnail != "c1" || first.Uploader.Name != "A" || first.Duration != 5*time.Second {
		t.Fatalf("%+v", first)
	}
	if second := item.Entries[1]; second.Title != "课_P1_2022-机器学习相关规定" || second.Description != "多P" || second.Uploader.Name != "啥都会一点的研究生" {
		t.Fatalf("%+v", second)
	}
}

func TestSpaceUploads(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stub.On("x/web-interface/nav", fixture(t, "nav-logged-out.json"))
	stub.On("live_user/v1/Master/info?uid=403748305", answer(`{"data":{"info":{"uname":"BML制作指挥部"}}}`))
	stub.On("x/space/wbi/arc/search?mid=403748305&order=pubdate&pn=1&", answer(`{"code":0,"data":{"list":{"vlist":[{"aid":626497566},{"aid":404}]},"page":{"count":2}}}`))
	stub.On("x/web-interface/view?aid=404", status(404))
	stubCaptured(t, stub)
	item := fetchItem(t, e, mediaID{kind: space, id: "403748305"}, "https://space.bilibili.com/403748305")
	if item.Title != "BML制作指挥部's uploads" || item.URL != "https://space.bilibili.com/403748305/upload/video" || !slices.Equal(titles(item), []string{"【4K60帧】咬人猫最新单曲《dududu》有没有戳中你心~【BML2020单品】"}) {
		t.Fatalf("%+v %v", item, titles(item))
	}
	// The upload list is WBI-signed with the key nav gives.
	req := stub.Requests("x/space/wbi/arc/search")[0].URL.String()
	query := req[strings.Index(req, "?")+1 : strings.Index(req, "&w_rid=")]
	if !strings.HasSuffix(req, "&w_rid="+md5Hex(query+"ea1db124af3c7062474693fa704f4ff8")) {
		t.Fatalf("unsigned: %s", req)
	}
}

func TestIntlSeason(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(func(o *Options) { o.API = Intl })
	stub.On("api.bilibili.tv/intl/gateway/v2/ogv/view/app/season?ep_id=2002", answer(`{"code":0,"result":{"title":"T","season_id":1001,"cover":"",
	  "publish":{"pub_time":"2021-01-02 03:04:05"},
	  "episodes":[],"modules":[{"data":{"episodes":[{"id":2001,"aid":1,"cid":10,"title":"1"},{"id":2002,"aid":2,"cid":20,"title":"2","long_title":"二","link":"https://www.bilibili.tv/play/1001/2002"}]}}]}}`))
	stub.On("bangumi.bilibili.com/anime/1001", func(*http.Request) testkit.Response {
		return testkit.Text(`window.__INITIAL_STATE__={"mediaInfo":{"cover":"p.jpg","title":"标题","evaluate":"简介"}};(function(){`)
	})
	item := fetchItem(t, e, mediaID{kind: episode, id: "2002"}, "https://www.bilibili.tv/en/play/1001/2002")
	if item.Title != "标题" || item.ID != "ss1001" || item.URL != "https://www.bilibili.tv/play/1001" || item.Thumbnail != "p.jpg" || !item.Collection || item.Focus != 2 || len(item.Entries) != 2 {
		t.Fatalf("%+v", item)
	}
	if en := item.Entries[1]; en.URL != "https://www.bilibili.tv/play/1001/2002" || en.Title != "2 二" || en.Fields["api"] != "INTL" || en.Ref.(*ref).position != 2 {
		t.Fatalf("%+v", en)
	}
}
