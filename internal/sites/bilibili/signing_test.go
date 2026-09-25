package bilibili

import (
	"strings"
	"testing"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/httpx"
)

func TestBVKnownPairs(t *testing.T) {
	t.Parallel()
	for aid, bv := range map[int64]string{2: "BV1xx411c7mD", 170001: "BV17x411w7KC", 626497566: "BV1qt4y1X7TW"} {
		if got, err := bvEncode(aid); got != bv || err != nil {
			t.Errorf("encode %d = %s %v", aid, got, err)
		}
		if got, err := bvDecode(bv[3:]); got != aid || err != nil {
			t.Errorf("decode %s = %d %v", bv, got, err)
		}
	}
}

func TestBVRoundTrip(t *testing.T) {
	t.Parallel()
	for _, aid := range []int64{1, 12345, 999_999_999, 1_000_000_000_000, bvMaxAid - 1} {
		bv, err := bvEncode(aid)
		if err != nil || !strings.HasPrefix(bv, "BV1") || len(bv) != 12 {
			t.Errorf("%d → %s %v", aid, bv, err)
		}
		if got, _ := bvDecode(bv[3:]); got != aid {
			t.Errorf("%d → %s → %d", aid, bv, got)
		}
	}
}

func TestBVErrorsAreInputErrors(t *testing.T) {
	t.Parallel()
	for _, aid := range []int64{0, -1, bvMaxAid} {
		if _, err := bvEncode(aid); !errs.Is(err, errs.Input) {
			t.Errorf("encode %d: %v", aid, err)
		}
	}
	for _, body := range []string{"", "xx411c7m", "xx411c7mDD", "xx411c7m0"} {
		if _, err := bvDecode(body); !errs.Is(err, errs.Input) {
			t.Errorf("decode %q: %v", body, err)
		}
	}
	if bvid("abc") != "" || bvid("0") != "" || bvid("2") != "BV1xx411c7mD" {
		t.Error("bvid")
	}
}

func TestWBIMixinKeyMatchesTheDocumentedExample(t *testing.T) {
	t.Parallel()
	if key := wbiMixinKey("7cd084941338484aae1ad9425b84077c", "4932caff0ff746eab6f01bf08b70ac45"); key != "ea1db124af3c7062474693fa704f4ff8" {
		t.Fatalf("key = %s", key)
	}
	// The documented signature of bilibili-API-collect's example query.
	signed := wbiSign("foo=114&bar=514&zab=1919810&wts=1702204169", "ea1db124af3c7062474693fa704f4ff8")
	if signed != "foo=114&bar=514&zab=1919810&wts=1702204169&w_rid="+md5Hex("foo=114&bar=514&zab=1919810&wts=1702204169ea1db124af3c7062474693fa704f4ff8") {
		t.Fatalf("signed = %s", signed)
	}
	if wbiMixinKey("", "") != "" {
		t.Fatal("empty keys gave characters")
	}
}

func TestDigests(t *testing.T) {
	t.Parallel()
	if md5Hex("") != "d41d8cd98f00b204e9800998ecf8427e" || md5Hex("abc") != "900150983cd24fb0d6963f7d28e17f72" {
		t.Fatal("md5")
	}
	if appSign("a=1", "secret") != md5Hex("a=1secret") {
		t.Fatal("appSign")
	}
	if k := imageKey("https://i0.hdslb.com/bfs/wbi/7cd084941338484aae1ad9425b84077c.png"); k != "7cd084941338484aae1ad9425b84077c" {
		t.Fatalf("imageKey = %s", k)
	}
}

func TestTVLoginParamsAreSortedAndSigned(t *testing.T) {
	t.Parallel()
	p := tvLoginParams()
	last := p[len(p)-1]
	if last[0] != "sign" || last[1] != appSign(httpx.FormEncode(p[:len(p)-1]), tvAppSecret) {
		t.Fatalf("sign = %v", last)
	}
	for i := 1; i < len(p)-1; i++ {
		if p[i-1][0] >= p[i][0] {
			t.Fatalf("keys out of order: %s, %s", p[i-1][0], p[i][0])
		}
	}
	values := map[string]string{}
	for _, f := range p {
		values[f[0]] = f[1]
	}
	if values["appkey"] != tvAppKey || values["mobi_app"] != "android_tv_yst" || len(values["device_id"]) != 20 ||
		values["guid"] != values["buvid"] || len(values["buvid"]) != 37 || values["fingerprint"] != values["local_fingerprint"] {
		t.Fatalf("%v", values)
	}
	// Re-signing replaces the signature rather than adding a second one.
	again := signTV(p)
	if len(again) != len(p) || again[len(again)-1] != last {
		t.Fatal("re-signing changed the form")
	}
}

func TestPlayURLSigning(t *testing.T) {
	t.Parallel()
	s := plainSession()
	s.wbi = "ea1db124af3c7062474693fa704f4ff8"
	r := &ref{kind: video, aid: "626497566", cid: "220355130", api: Web}
	web := webPlayURL(s, r, "0", "1700000000")
	q := "support_multi_audio=true&from_client=BROWSER&avid=626497566&cid=220355130&fnval=4048&fnver=0&fourk=1&otype=json&qn=0&try_look=1&wts=1700000000"
	if web != "https://api.bilibili.com/x/player/wbi/playurl?"+wbiSign(q, s.wbi) {
		t.Fatalf("web = %s", web)
	}

	// Episodes are not WBI-signed; courses go to pugv; a cookie drops try_look; an area adds the proxy's key.
	s.cookie, s.token, s.area, s.host = "SESSDATA=x", "tok", "hk", "proxy.test"
	ep := webPlayURL(s, &ref{kind: cheese, aid: "1", cid: "2", epid: "3"}, "127", "9")
	if ep != "https://proxy.test/pugv/player/web/v2/playurl?support_multi_audio=true&from_client=BROWSER&avid=1&cid=2&fnval=4048&fnver=0&fourk=1&access_key=tok&area=hk&otype=json&qn=127&module=bangumi&ep_id=3&session=&wts=9" {
		t.Fatalf("pgc = %s", ep)
	}

	tv := tvPlayURL(plainSession(), &ref{kind: episode, aid: "1", cid: "2", epid: "3"}, "0", "9")
	tq := "appkey=" + tvAppKey + "&build=106500&cid=2&device=android&ep_id=3&expire=0&fnval=4048&fnver=0&fourk=1&mid=0&mobi_app=android_tv_yst&object_id=1&platform=android&playurl_type=1&qn=0&ts=9"
	if tv != "https://"+defaultTVHost+"/pgc/player/api/playurltv?"+tq+"&sign="+appSign(tq, tvAppSecret) {
		t.Fatalf("tv = %s", tv)
	}

	intl := intlPlayURL(plainSession(), &ref{aid: "1", cid: "2", epid: "3"}, "0", "1", "9")
	if intl != "https://api.biliintl.com/intl/gateway/v2/ogv/playurl?aid=1&cid=2&ep_id=3&platform=android&prefer_code_type=1&qn=0&s_locale=zh_SG" {
		t.Fatalf("intl = %s", intl)
	}
	proxy := plainSession()
	proxy.host = "proxy.test"
	iq := "aid=1&appkey=" + biliPlusAppKey + "&area=th&cid=2&ep_id=3&platform=android&prefer_code_type=0&qn=0&ts=9&s_locale=zh_SG"
	if got := intlPlayURL(proxy, &ref{aid: "1", cid: "2", epid: "3"}, "0", "0", "9"); got != "https://proxy.test/intl/gateway/v2/ogv/playurl?"+iq+"&sign="+appSign(iq, biliPlusSecret) {
		t.Fatalf("intl proxy = %s", got)
	}
}
