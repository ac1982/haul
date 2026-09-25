package subtitle

import (
	"testing"

	"github.com/ac1982/haul/internal/media"
)

func TestTime(t *testing.T) {
	for in, want := range map[float64]string{64.13: "00:01:04,130", 3661.5: "01:01:01,500", 0: "00:00:00,000"} {
		if got := Time(in); got != want {
			t.Errorf("Time(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestBilibiliJSON(t *testing.T) {
	got, err := ToSRT(media.BilibiliJSON, []byte(`{"body":[{"from":1.0,"to":2.5,"content":"你好"},{"to":4.0,"content":"再见"}]}`))
	want := "1\n00:00:01,000 --> 00:00:02,500\n你好\n\n2\n00:00:00,000 --> 00:00:04,000\n再见\n\n"
	if err != nil || got != want {
		t.Errorf("got %q, %v", got, err)
	}
}

func TestJSON3DropsNewlineOnlyEvents(t *testing.T) {
	in := `{"events": [
	  {"tStartMs": 0, "dDurationMs": 5000, "id": 1},
	  {"tStartMs": 805, "dDurationMs": 2368, "segs": [{"utf8": "How much"}, {"utf8": " is left"}]},
	  {"tStartMs": 3100, "dDurationMs": 10, "aAppend": 1, "segs": [{"utf8": "\n"}]},
	  {"tStartMs": 61000, "dDurationMs": 1500, "segs": [{"utf8": "Done."}]}]}`
	want := "1\n00:00:00,805 --> 00:00:03,173\nHow much is left\n\n2\n00:01:01,000 --> 00:01:02,500\nDone.\n\n"
	for _, f := range []media.SubtitleFormat{media.YouTubeJSON3, media.BilibiliJSON} {
		if got, err := ToSRT(f, []byte(in)); err != nil || got != want {
			t.Errorf("%s: got %q, %v", f, got, err)
		}
	}
	if _, err := ToSRT(media.ASS, nil); err == nil {
		t.Error("ASS cannot become SRT")
	}
}

func TestLanguages(t *testing.T) {
	cases := []struct {
		tag  string
		auto bool
		want Language
	}{
		{"zh-Hans", false, Language{"chi", "中文（简体）"}},
		{"zh-CN", false, Language{"chi", "中文（简体）"}},
		{"en", true, Language{"eng", "English, auto-generated"}},
		{"en-US", false, Language{"eng", "English (US)"}},
		{"pt_BR", false, Language{"por", "Português (BR)"}},
		{"xx", false, Language{"und", "xx"}},
	}
	for _, c := range cases {
		if got := LanguageOf(c.tag, c.auto); got != c.want {
			t.Errorf("LanguageOf(%q) = %+v, want %+v", c.tag, got, c.want)
		}
	}
}
