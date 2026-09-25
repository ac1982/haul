package bilibili

import (
	"net/http"
	"testing"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/testkit"
)

func TestRecognisesBilibiliLinks(t *testing.T) {
	t.Parallel()
	e := New(nil, DefaultOptions())
	for _, link := range []string{"BV1qt4y1X7TW", "bv1qt4y1x7tw", "av170001", "ep123", "ss33073", "md28223066", "cheese/ep1234", "cheese/ss12",
		"https://www.bilibili.com/video/BV1qt4y1X7TW?p=2", "https://b23.tv/abc", "https://space.bilibili.com/1/favlist?fid=2",
		"https://www.bilibili.tv/en/play/1/2", "https://www.biliintl.com/x", " https://bilibili.com/video/av1 "} {
		if _, ok := e.Match(link); !ok {
			t.Errorf("%q not matched", link)
		}
	}
	for _, link := range []string{"", "hello", "https://example.com/video/BV1qt4y1X7TW", "https://notbilibili.com/video/1", "BV1short",
		"https://youtu.be/DdCEmlAydcw", "cheese/xx1", "https://bilibili.com.evil.test/video/av1"} {
		if _, ok := e.Match(link); ok {
			t.Errorf("%q matched", link)
		}
	}
	// Links without a scheme come back with one, in the form Resolve takes.
	if got, _ := e.Match("m.bilibili.com/video/av1"); got != "https://m.bilibili.com/video/av1" {
		t.Errorf("got %q", got)
	}
	if got, _ := e.Match(" BV1qt4y1X7TW "); got != "BV1qt4y1X7TW" {
		t.Errorf("got %q", got)
	}
}

func TestInfo(t *testing.T) {
	t.Parallel()
	info := New(nil, DefaultOptions()).Info()
	if info.Site != "bilibili" || info.Name != "bilibili" || info.Unit != "page" || info.OwnerLabel != "uploader" || len(info.Requires) != 0 ||
		info.Links != "bilibili.com (video, bangumi, course, space, list), b23.tv, BV…, ep…" {
		t.Fatalf("%+v", info)
	}
}

func resolveAll(t *testing.T, e *Extractor, cases map[string]mediaID) {
	t.Helper()
	for link, want := range cases {
		got, _, err := e.resolveID(ctx, plainSession(), link)
		if err != nil {
			t.Errorf("%s: %v", link, err)
		} else if got != want {
			t.Errorf("%s: got %v, want %v", link, got, want)
		}
	}
}

func TestResolvesVideoLinksAndIDs(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stubCaptured(t, stub)
	resolveAll(t, e, map[string]mediaID{
		"BV1qt4y1X7TW": {kind: video, id: "626497566"},
		"av170001":     {kind: video, id: "170001"},
		"https://www.bilibili.com/video/BV1qt4y1X7TW/?p=2": {kind: video, id: "626497566"},
		"https://www.bilibili.com/video/av170001":          {kind: video, id: "170001"},
		"m.bilibili.com/video/av170001":                    {kind: video, id: "170001"},
	})
	// Every video id is checked for a redirect to an episode.
	if n := len(stub.Requests("www.bilibili.com/video/av626497566/")); n != 2 {
		t.Errorf("%d redirect checks", n)
	}
}

func TestLicensedVideoRedirectsToItsEpisode(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stub.On("www.bilibili.com/video/av840009001", func(*http.Request) testkit.Response {
		return testkit.Response{Redirect: "https://www.bilibili.com/bangumi/play/ep317690"}
	})
	stub.On("www.bilibili.com/bangumi/play/ep317690", status(200))
	resolveAll(t, e, map[string]mediaID{"av840009001": {kind: episode, id: "317690"}})
}

func TestExpandsShortLinks(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stub.On("b23.tv/AbCdEf1", func(*http.Request) testkit.Response {
		return testkit.Response{Redirect: "https://www.bilibili.com/video/BV1qt4y1X7TW?p=1&share=1"}
	})
	stub.On("www.bilibili.com/video/BV1qt4y1X7TW?p=1&share=1", status(200))
	stub.On("www.bilibili.com/video/av", status(200))
	id, link, err := e.resolveID(ctx, plainSession(), "https://b23.tv/AbCdEf1")
	if err != nil || id != (mediaID{kind: video, id: "626497566"}) || link != "https://www.bilibili.com/video/BV1qt4y1X7TW?p=1&share=1" {
		t.Fatalf("%v %q %v", id, link, err)
	}

	stub.On("b23.tv/loop", status(200))
	if _, _, err := e.resolveID(ctx, plainSession(), "https://b23.tv/loop"); err == nil {
		t.Fatal("a b23.tv link that goes nowhere resolved")
	}
}

func TestResolvesSeasonsCoursesAndLists(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stubCaptured(t, stub)
	stub.On("pgc/review/user?media_id=28223066", answer(`{"code":0,"result":{"media":{"new_ep":{"id":317699}}}}`))
	stub.On("pugv/view/web/season?season_id=12", answer(`{"code":0,"data":{"episodes":[{"id":1201},{"id":1202}]}}`))
	stub.On("www.bilibili.com/bangumi/play/xyz", func(*http.Request) testkit.Response {
		return testkit.Text(`<script>window.__INITIAL_STATE__={"epList":[{"id":4455}]};(function(){})</script>`)
	})
	resolveAll(t, e, map[string]mediaID{
		"ep317690": {kind: episode, id: "317690"},
		"EP317690": {kind: episode, id: "317690"},
		"https://www.bilibili.com/bangumi/play/ep317690?from=search": {kind: episode, id: "317690"},
		// A season resolves to its first episode, a media page to its newest.
		"ss33073": {kind: episode, id: "317690"},
		"https://www.bilibili.com/bangumi/play/ss33073": {kind: episode, id: "317690"},
		"md28223066": {kind: episode, id: "317699"},
		"https://www.bilibili.com/bangumi/media/md28223066/": {kind: episode, id: "317699"},
		"https://www.bilibili.com/bangumi/play/xyz":          {kind: episode, id: "4455"},
		"https://m.bilibili.com/bangumi/play?ep_id=77":       {kind: episode, id: "77"},
		"https://www.bilibili.tv/en/play/1001/2002":          {kind: episode, id: "2002"},
		"cheese/ep1234": {kind: cheese, id: "1234"},
		"cheese/ss12":   {kind: cheese, id: "1201"},
		"https://www.bilibili.com/cheese/play/ep1234":                                       {kind: cheese, id: "1234"},
		"https://www.bilibili.com/cheese/play/ss12":                                         {kind: cheese, id: "1201"},
		"https://space.bilibili.com/403748305":                                              {kind: space, id: "403748305"},
		"https://space.bilibili.com/403748305/favlist?fid=66":                               {kind: favorites, id: "66", mid: "403748305"},
		"https://space.bilibili.com/403748305/favlist":                                      {kind: favorites, id: "", mid: "403748305"},
		"https://space.bilibili.com/403748305/lists/4242?type=season":                       {kind: collection, id: "4242"},
		"https://space.bilibili.com/403748305/lists/77/?type=series":                        {kind: series, id: "77"},
		"https://space.bilibili.com/1/channel/collectiondetail?sid=9":                       {kind: collection, id: "9"},
		"https://space.bilibili.com/1/channel/seriesdetail?sid=8":                           {kind: series, id: "8"},
		"https://www.bilibili.com/medialist/play/1?business=space_series&business_id=5":     {kind: series, id: "5"},
		"https://www.bilibili.com/medialist/play/1?business=space_collection&business_id=6": {kind: collection, id: "6"},
	})
}

func TestBadLinksAreInputErrors(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stub.On("www.bilibili.com/bangumi/play/nothing", status(200))
	for _, link := range []string{"not a link", "https://vimeo.com/1", "BV1qt4y1X7T0", "BV1qt4y1X7TWW", "https://www.bilibili.com/cheese/play/x",
		"https://www.bilibili.com/bangumi/play/nothing"} {
		_, _, err := e.resolveID(ctx, plainSession(), link)
		if !errs.Is(err, errs.Input) {
			t.Errorf("%q: %v", link, err)
		}
	}
}
