package bilibili

import (
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ac1982/haul/internal/extract"
	"github.com/ac1982/haul/internal/media"
	"github.com/ac1982/haul/internal/testkit"
)

func TestResolveAndFormatsOfAVideo(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stubCaptured(t, stub)
	stub.On("api.bilibili.com/x/web-interface/nav", fixture(t, "nav-logged-out.json"))
	stub.On("x/player/wbi/v2?cid=220355130&aid=626497566", answer(`{"code":0,"data":{"view_points":[{"content":"开场","from":0,"to":1},{"content":"正片","from":1,"to":226}]}}`))
	stub.On(appDmViewEndpoint, grpcAnswer(t, dmViewReply([2]string{"ai-zh", "//aisubtitle.hdslb.com/bfs/ai_subtitle/x.json"}), nil))
	stub.On("comment.bilibili.com/220355130.xml", func(*http.Request) testkit.Response {
		return testkit.Text(`<?xml version="1.0" encoding="UTF-8"?><i><d p="0.5,1,25,16777215,0,0,0,0">第一条</d><d p="0.8,1,25,255,0,0,0,0">第二条</d></i>`)
	})

	router := extract.NewRouter(e)
	x, link, err := router.Route("https://www.bilibili.com/video/BV1qt4y1X7TW/?spm_id_from=333")
	if err != nil || x != e {
		t.Fatal(err)
	}
	item, err := e.Resolve(ctx, link)
	if err != nil {
		t.Fatal(err)
	}
	if item.LoggedIn == nil || *item.LoggedIn || item.Site != Site || item.ID != "BV1qt4y1X7TW" || len(item.Entries) != 1 || item.Focus != 0 || item.Collection {
		t.Fatalf("%+v", item)
	}
	entry := item.Entries[0]
	f, err := extract.LoadFormats(ctx, e, item, entry)
	if err != nil {
		t.Fatal(err)
	}

	// The web API with the WBI key from the login check, both passes.
	playurls := stub.Requests("x/player/wbi/playurl")
	if len(playurls) != 2 || !strings.Contains(playurls[0].URL.RawQuery, "&w_rid=") {
		t.Fatalf("%d playurl calls", len(playurls))
	}
	query := playurls[0].URL.RawQuery
	if !strings.HasSuffix(query, "&w_rid="+md5Hex(query[:strings.Index(query, "&w_rid=")]+"ea1db124af3c7062474693fa704f4ff8")) {
		t.Fatalf("signed with the wrong key: %s", query)
	}
	if len(f.Video) != 4 || len(f.Audio) != 3 || f.Raw == "" {
		t.Fatalf("%d video, %d audio", len(f.Video), len(f.Audio))
	}
	for _, r := range []media.Resource{f.Video[0].Source, f.Audio[2].Source} {
		if !strings.HasPrefix(r.URL, "http://"+backupHost+"/upgcxcode/30/51/220355130/") || r.Header.Get("Referer") != "https://www.bilibili.com" || r.Header.Get("Cookie") != "" {
			t.Fatalf("%+v", r)
		}
	}
	if !strings.Contains(f.Audio[2].Source.URL, "-1-302") {
		t.Fatalf("audio %s", f.Audio[2].Source.URL)
	}
	if !slices.Equal(f.Chapters, []media.Chapter{{Title: "开场", Start: 0, End: time.Second}, {Title: "正片", Start: time.Second, End: 226 * time.Second}}) {
		t.Fatalf("%+v", f.Chapters)
	}
	if len(f.Subtitles) != 1 || f.Subtitles[0].Lang != "zh" || !f.Subtitles[0].Auto || f.Subtitles[0].Source.URL != "https://aisubtitle.hdslb.com/bfs/ai_subtitle/x.json" {
		t.Fatalf("%+v", f.Subtitles)
	}
	if len(f.Sidecars) != 1 {
		t.Fatal("no danmaku")
	}
	base := filepath.Join(t.TempDir(), "【4K60帧】咬人猫最新单曲《dududu》")
	files, err := f.Sidecars[0].Write(ctx, base)
	if err != nil || len(files) != 2 {
		t.Fatalf("%v %v", files, err)
	}
	ass, _ := os.ReadFile(base + ".ass")
	if !strings.Contains(string(ass), `{\move(1920, 40, -120, 40)\c&HFF0000&}第二条`) {
		t.Fatalf("%s", ass)
	}
}

func TestResolveFocusesTheLinkedEpisode(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(func(o *Options) { o.API = TV })
	stubCaptured(t, stub)
	item, err := e.Resolve(ctx, "https://www.bilibili.com/bangumi/play/ep317690")
	if err != nil {
		t.Fatal(err)
	}
	if item.Focus != 1 || item.LoggedIn != nil || item.Entries[0].Fields["api"] != "TV" {
		t.Fatalf("%+v", item)
	}
	if _, err := e.Resolve(ctx, "https://example.com/video/1"); err == nil || !strings.Contains(err.Error(), "Unsupported bilibili link") {
		t.Fatalf("err = %v", err)
	}
}
