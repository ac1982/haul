package format

import (
	"testing"
	"time"
)

func TestSizesAndDurations(t *testing.T) {
	cases := map[string]string{
		FileSize(512, 2):               "512 bytes",
		FileSize(1536, 2):              "1.50 KB",
		FileSize(3*1024*1024, 2):       "3.00 MB",
		FileSize(300*1024*1024, 1):     "300.0 MB",
		Duration(65, false):            "01m05s",
		Duration(3661, false):          "1h01m01s",
		Duration(3661, true):           "01:01:01",
		Duration(90000, true):          "25:00:00",
		FFmpegCreationTime(1595684326): "2020-07-25T13:38:46.000000Z",
		ISO(1595684326):                "2020-07-25T13:38:46Z",
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestDatePatterns(t *testing.T) {
	d := time.Date(2024, 3, 7, 9, 5, 4, 12_000_000, time.Local)
	for pattern, want := range map[string]string{
		"yyyy-MM-dd_HH-mm-ss": "2024-03-07_09-05-04",
		"yyyyMMddHHmmssSSS":   "20240307090504012",
		"yy/M/d H:m":          "24/3/7 9:5",
		"'at' HH'h'":          "at 09h",
		"dd MMM yyyy":         "07 Mar 2024",
	} {
		if got := FormatDate(d, pattern); got != want {
			t.Errorf("%s: got %q, want %q", pattern, got, want)
		}
	}
	if Timestamp(0, "yyyy") != "null" {
		t.Error("0 is null")
	}
	ts, ok := ParseDateTime("2020-04-06 07:05:00")
	if !ok || Timestamp(ts, "yyyy-MM-dd HH:mm:ss") != "2020-04-06 07:05:00" {
		t.Errorf("ParseDateTime round trip: %d %v", ts, ok)
	}
}

func TestFileNames(t *testing.T) {
	if got := ValidFileName(`a/b:c*d?e"f<g>h|i\j`); got != "a_b_c_d_e_f_g_h_i_j" {
		t.Errorf("ValidFileName = %q", got)
	}
	if got := CleanName(" name... "); got != "name" {
		t.Errorf("CleanName = %q", got)
	}
	if got := ChangeExtension("a/b.mp4", "xml"); got != "a/b.xml" {
		t.Errorf("ChangeExtension = %q", got)
	}
	if got := ChangeExtension("a/b.c/d", "srt"); got != "a/b.c/d.srt" {
		t.Errorf("ChangeExtension without extension = %q", got)
	}
}

func TestQueryAndEntities(t *testing.T) {
	if QueryValue("p", "https://x/y?p=3&q=4") != "3" || QueryValue("q", "https://x/y?p=3&q=4#frag") != "4" || QueryValue("z", "https://x/y?p=3") != "" {
		t.Error("QueryValue")
	}
	if got := UnescapeEntities("a &amp; b &#26159; &#x4E2D; &lt;i&gt;"); got != "a & b 是 中 <i>" {
		t.Errorf("UnescapeEntities = %q", got)
	}
	if got := PercentEncode("a b&c=d~é"); got != "a%20b%26c%3Dd~%C3%A9" {
		t.Errorf("PercentEncode = %q", got)
	}
	if len(RandomString(20)) != 20 {
		t.Error("RandomString length")
	}
}
