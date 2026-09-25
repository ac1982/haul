package bilibili

import (
	"context"
	"regexp"
	"strings"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/format"
	"github.com/ac1982/haul/internal/jsonv"
)

// maxQuality is the qn asked for in the second pass: the highest there is.
const maxQuality = "127"

var playInfo = regexp.MustCompile(`window\.__playinfo__=([\s\S]*?)</script>`)

// playJSON asks the entry's API for its streams at quality qn and returns the answer as JSON text.
func (e *Extractor) playJSON(ctx context.Context, s *session, r *ref, qn string) (string, error) {
	console.Debugf("aid=%s,cid=%s,epId=%s,api=%s,qn=%s,pgc=%t,cheese=%t", r.aid, r.cid, r.epid, r.api, qn, r.kind.pgc(), r.kind == cheese)
	switch r.api {
	case Intl:
		return e.intlJSON(ctx, s, r, qn, "0")
	case App:
		return e.appPlayView(ctx, s, r)
	case TV:
		return e.getText(ctx, s, tvPlayURL(s, r, qn, format.UnixSeconds()))
	}
	text, err := e.getText(ctx, s, webPlayURL(s, r, qn, format.UnixSeconds()))
	if err != nil {
		return "", err
	}
	if strings.Contains(text, `"大会员专享限制"`) {
		console.Warn("This video needs a bilibili premium membership; trying the web page")
		html, err := e.getText(ctx, s, "https://www.bilibili.com/bangumi/play/ep"+r.epid)
		if err != nil {
			return "", err
		}
		if m := playInfo.FindStringSubmatch(html); m != nil {
			text = m[1]
		}
	}
	return text, nil
}

// webPlayURL is the web player's playurl: WBI-signed for videos, the pgc (or, for courses, pugv) endpoint for
// episodes. ts is the current unix time.
func webPlayURL(s *session, r *ref, qn, ts string) string {
	base := "api.bilibili.com/x/player/wbi/playurl"
	if r.kind.pgc() {
		base = s.host + "/pgc/player/web/v2/playurl"
	}
	q := "support_multi_audio=true&from_client=BROWSER&avid=" + r.aid + "&cid=" + r.cid + "&fnval=4048&fnver=0&fourk=1"
	if s.area != "" {
		q += "&access_key=" + s.token + "&area=" + s.area
	}
	q += "&otype=json&qn=" + qn
	if r.kind.pgc() {
		q += "&module=bangumi&ep_id=" + r.epid + "&session="
	}
	if s.cookie == "" {
		q += "&try_look=1"
	}
	q += "&wts=" + ts
	if !r.kind.pgc() {
		q = wbiSign(q, s.currentWBI())
	}
	api := "https://" + base + "?" + q
	if r.kind == cheese {
		api = strings.ReplaceAll(api, "/pgc/", "/pugv/")
	}
	return api
}

// tvPlayURL is the TV client's playurl, appkey-signed.
func tvPlayURL(s *session, r *ref, qn, ts string) string {
	base := s.tvHost + "/x/tv/playurl"
	if r.kind.pgc() {
		base = s.tvHost + "/pgc/player/api/playurltv"
	}
	q := ""
	if s.token != "" {
		q = "access_key=" + s.token + "&"
	}
	q += "appkey=" + tvAppKey + "&build=106500&cid=" + r.cid + "&device=android"
	if r.kind.pgc() {
		q += "&ep_id=" + r.epid + "&expire=0"
	}
	q += "&fnval=4048&fnver=0&fourk=1&mid=0&mobi_app=android_tv_yst"
	q += "&object_id=" + r.aid + "&platform=android&playurl_type=1&qn=" + qn + "&ts=" + ts
	api := "https://" + base + "?" + q + "&sign=" + appSign(q, tvAppSecret)
	if r.kind == cheese {
		api = strings.ReplaceAll(api, "/pgc/", "/pugv/")
	}
	return api
}

// intlJSON asks the international API, directly or through a BiliPlus-style proxy (then signed), preferring the
// codec type code ("0" or "1").
func (e *Extractor) intlJSON(ctx context.Context, s *session, r *ref, qn, code string) (string, error) {
	return e.getText(ctx, s, intlPlayURL(s, r, qn, code, format.UnixSeconds()))
}

func intlPlayURL(s *session, r *ref, qn, code, ts string) string {
	plus := s.biliPlus()
	q := ""
	if s.token != "" {
		q = "access_key=" + s.token + "&"
	}
	q += "aid=" + r.aid
	if plus {
		area := s.area
		if area == "" {
			area = "th"
		}
		q += "&appkey=" + biliPlusAppKey + "&area=" + area
	}
	q += "&cid=" + r.cid + "&ep_id=" + r.epid + "&platform=android&prefer_code_type=" + code + "&qn=" + qn
	if plus {
		q += "&ts=" + ts
	}
	q += "&s_locale=zh_SG"
	if plus {
		return "https://" + s.host + "/intl/gateway/v2/ogv/playurl?" + q + "&sign=" + appSign(q, biliPlusSecret)
	}
	return "https://api.biliintl.com/intl/gateway/v2/ogv/playurl?" + q
}

// payloadRoot is the node that holds dash or durl: data, result, or result.video_info.
func payloadRoot(root jsonv.Value, raw string) jsonv.Value {
	if strings.Contains(raw, `"result":{`) {
		if strings.Contains(raw, `"video_info":{`) {
			return root.Path("result", "video_info")
		}
		return root.Get("result")
	}
	if strings.Contains(raw, `"data":{`) {
		return root.Get("data")
	}
	return root
}

// noStreamsError says why an answer had no streams, with bilibili's reason when it gave one.
func noStreamsError(title, raw string) error {
	if j, err := jsonv.ParseString(raw); err == nil {
		if code := j.Get("code").String(); code != "" && code != "0" {
			return errs.New("Could not find the streams of %q: bilibili answered %s: %s", title, code, j.Get("message").String())
		}
	}
	if raw != "" && len(raw) < 100 {
		return errs.New("Could not find the streams of %q: %s", title, raw)
	}
	return errs.New("Could not find the streams of %q (--debug shows the response)", title)
}
