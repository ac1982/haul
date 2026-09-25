package bilibili

import (
	"slices"
	"testing"

	"github.com/ac1982/haul/internal/jsonv"
	"github.com/ac1982/haul/internal/media"
)

func dmViewReply(subs ...[2]string) []byte {
	var subtitle, reply pbWriter
	subtitle.string(1, "zh-CN")
	for _, s := range subs {
		var item pbWriter
		item.string(1, "id").int(2, 1).string(3, s[0]).string(4, "doc").string(5, s[1])
		subtitle.message(3, &item)
	}
	reply.bool(1, false).message(3, &subtitle).bool(8, true)
	return reply.buf
}

func langs(subs []media.Subtitle) []string {
	var out []string
	for _, s := range subs {
		tag := s.Lang
		if s.Auto {
			tag += "(auto)"
		}
		out = append(out, tag+" "+s.Source.URL+" "+string(s.Format))
	}
	return out
}

func TestLoggedOutSubtitlesComeFromTheApp(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	var sent []byte
	stub.On(appDmViewEndpoint, grpcAnswer(t, dmViewReply([2]string{"ai-zh", "//aisubtitle.hdslb.com/a.json"}, [2]string{"en-US", "https://s.test/b.json"}), &sent))
	r := &ref{kind: video, aid: "1", cid: "2", api: Web}
	subs := e.subtitles(ctx, plainSession(), r, jsonv.Null)
	want := []string{"zh(auto) https://aisubtitle.hdslb.com/a.json bilijson", "en-US https://s.test/b.json bilijson"}
	if !slices.Equal(langs(subs), want) {
		t.Fatalf("%v", langs(subs))
	}
	if req, _ := decodeProto(sent); req.int(1) != 1 || req.int(2) != 2 || req.int(3) != 1 {
		t.Fatalf("%+v", req)
	}
	if call := stub.Requests(appDmViewEndpoint)[0]; call.Header.Get("User-Agent") != androidUserAgent || call.Header.Get("Content-Type") != "application/grpc" {
		t.Fatalf("%v", call.Header)
	}
	if h := subs[0].Source.Header; h.Get("Referer") != "https://www.bilibili.com" || h.Get("User-Agent") != "Mozilla/5.0" {
		t.Fatalf("%v", h)
	}
	// No subtitle message: no subtitles, and no error.
	stub.On(appDmViewEndpoint, grpcAnswer(t, nil, nil))
	if subs := e.subtitles(ctx, plainSession(), r, jsonv.Null); len(subs) != 0 {
		t.Fatalf("%v", langs(subs))
	}
}

func TestLoggedInSubtitlesTryThePlayerThenTheViewThenTheApp(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	s := plainSession()
	s.cookie = "SESSDATA=x"
	r := &ref{kind: video, aid: "1", cid: "2", api: Web}
	player, _ := jsonv.ParseString(`{"data":{"subtitle":{"subtitles":[{"lan":"zh-Hans","subtitle_url":"//s.test/zh.json"}]}}}`)
	if got := langs(e.subtitles(ctx, s, r, player)); !slices.Equal(got, []string{"zh-Hans https://s.test/zh.json bilijson"}) {
		t.Fatalf("%v", got)
	}
	if n := len(stub.Requests("")); n != 0 {
		t.Fatalf("%d requests for subtitles the player listed", n)
	}

	// A list with a hole is no list: the view is asked, and then the app.
	holey, _ := jsonv.ParseString(`{"data":{"subtitle":{"subtitles":[{"lan":"zh-Hans","subtitle_url":""}]}}}`)
	stub.On("x/web-interface/view?aid=1&cid=2", answer(`{"data":{"subtitle":{"list":[{"lan":"ja","subtitle_url":"https://s.test/ja.json"}]}}}`))
	if got := langs(e.subtitles(ctx, s, r, holey)); !slices.Equal(got, []string{"ja https://s.test/ja.json bilijson"}) {
		t.Fatalf("%v", got)
	}
	stub.On("x/web-interface/view?aid=1&cid=2", status(500))
	stub.On(appDmViewEndpoint, grpcAnswer(t, dmViewReply([2]string{"ko", "https://s.test/ko.json"}), nil))
	if got := langs(e.subtitles(ctx, s, r, jsonv.Null)); !slices.Equal(got, []string{"ko https://s.test/ko.json bilijson"}) {
		t.Fatalf("%v", got)
	}
	if h := e.subtitles(ctx, s, r, player)[0].Source.Header; h.Get("Cookie") != "SESSDATA=x" {
		t.Fatalf("%v", h)
	}
}

func TestIntlSubtitles(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(func(o *Options) { o.API = Intl })
	r := &ref{kind: episode, aid: "1", cid: "2", epid: "3", api: Intl, position: 2}
	stub.On("api.biliintl.com/intl/gateway/web/v2/subtitle?episode_id=3", answer(`{"data":{"subtitles":[
	  {"lang_key":"th","url":"https://s.test/th.json"},{"lang_key":"zh-Hant","url":"https://s.test/zh.ass"}]}}`))
	if got := langs(e.subtitles(ctx, plainSession(), r, jsonv.Null)); !slices.Equal(got, []string{"th https://s.test/th.json bilijson", "zh-Hant https://s.test/zh.ass ass"}) {
		t.Fatalf("%v", got)
	}
	// The web list failing, the app season's episode (by position) is read.
	stub.On("api.biliintl.com/intl/gateway/web/v2/subtitle?episode_id=3", status(404))
	stub.On("api.bilibili.tv/intl/gateway/v2/ogv/view/app/season?ep_id=3&platform=android&s_locale=zh_SG", answer(`{"result":{"modules":[{"data":{"episodes":[
	  {"subtitles":[{"key":"en","url":"https://s.test/first.json"}]},{"subtitles":[{"key":"id","url":"https:\/\/s.test\/id.ass"}]}]}}]}}`))
	if got := langs(e.subtitles(ctx, plainSession(), r, jsonv.Null)); !slices.Equal(got, []string{"id https://s.test/id.ass ass"}) {
		t.Fatalf("%v", got)
	}
}

func TestSubtitleLanguageTags(t *testing.T) {
	t.Parallel()
	for key, want := range map[string]string{
		"zh-Hans": "zh-Hans", "zh-hans": "zh-Hans", "ZH-HANT": "zh-Hant", "en-US": "en-US", "en-us": "en-US", "ai-zh": "zh(auto)",
		"AI-en": "en(auto)", "zh_CN": "zh-CN", "es-419": "es-419", "ja": "ja", "ai": "ai",
	} {
		lang, auto := subtitleLanguage(key)
		if auto {
			lang += "(auto)"
		}
		if lang != want {
			t.Errorf("%s → %s, want %s", key, lang, want)
		}
	}
}
