package bilibili

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/jsonv"
)

// The Android app's gRPC services, spoken over plain HTTPS POSTs carrying one gRPC frame each way.
const (
	appUGCEndpoint    = "https://grpc.biliapi.net/bilibili.app.playurl.v1.PlayURL/PlayView"
	appPGCEndpoint    = "https://app.bilibili.com/bilibili.pgc.gateway.player.v2.PlayURL/PlayView"
	appDmViewEndpoint = "https://app.biliapi.net/bilibili.community.service.dm.v1.DM/DmView"
)

// The device the APP API is told about. The build must be recent enough for the dubbing fields to be sent.
const (
	appDalvik    = "2.1.0"
	appOSVersion = "11"
	appBrand     = "M2012K11AC"
	appModel     = "Build/RKQ1.200826.002"
	appVersion   = "7.32.0"
	appBuild     = 7_320_200
	appChannel   = "xiaomi_cn_tv.danmaku.bili_zm20200902"
	appNetOID    = "46007"
	appCronet    = "1.36.1"
	appBuvid     = ""
	appMobiApp   = "android"
	appKey       = "android64"
	appSessionID = "dedf8669"
	appPlatform  = "android"
	appEnv       = "prod"
	appID        = 1
	appRegion    = "CN"
	appLanguage  = "zh"
	// androidUserAgent is what the DmView call sends.
	androidUserAgent = "Dalvik/2.1.0 (Linux; U; Android 6.0.1; oneplus a5010 Build/V417IR) 6.10.0 os/android model/oneplus a5010 mobi_app/android build/6100500 channel/bili innerVer/6100500 osVer/6.0.1 network/2"
)

// PlayViewReq.CodeType.
const (
	code264 = 1
	code265 = 2
	codeAV1 = 3
)

func codeType(codec string) int64 {
	switch codec {
	case "AVC":
		return code264
	case "AV1":
		return codeAV1
	}
	return code265
}

// playViewRequest is a PlayViewReq. For videos the epId field carries the aid.
func playViewRequest(epID, cid, codec int64) []byte {
	var w pbWriter
	w.int(1, epID)
	w.int(2, cid)
	w.int(3, 127)  // qn
	w.int(5, 4048) // fnval
	w.uint(6, 0)   // download: 0 play, 1 flv download, 2 dash download
	w.int(7, 2)    // forceHost: 0 allow ip, 1 http, 2 https
	w.bool(8, true)
	w.string(9, "main.ugc-video-detail.0.0") // spmid
	w.string(10, "main.my-history.0.0")      // fromSpmid
	w.int(12, codec)                         // preferCodecType
	return w.buf
}

// dmViewRequest is a DmViewReq for a video page.
func dmViewRequest(aid, cid int64) []byte {
	var w pbWriter
	w.int(1, aid).int(2, cid).int(3, 1).string(4, "main.ugc-video-detail.0.0")
	return w.buf
}

func deviceBin() []byte {
	var w pbWriter
	w.int(1, appID).int(2, appBuild).string(3, appBuvid).string(4, appMobiApp).string(5, appPlatform).
		string(7, appChannel).string(8, appBrand).string(9, appModel).string(10, appOSVersion)
	return w.buf
}

func metadataBin(token string) []byte {
	var w pbWriter
	w.string(1, token).string(2, appMobiApp).int(4, appBuild).string(5, appChannel).string(6, appBuvid).string(7, appPlatform)
	return w.buf
}

func networkBin() []byte {
	var w pbWriter
	w.int(1, 1) // type: WIFI
	w.string(3, appNetOID)
	return w.buf
}

func localeBin() []byte {
	var ids, w pbWriter
	ids.string(1, appLanguage).string(3, appRegion)
	w.message(1, &ids) // cLocale
	return w.buf
}

func fawkesBin() []byte {
	var w pbWriter
	w.string(1, appKey).string(2, appEnv).string(3, appSessionID)
	return w.buf
}

// appHeader is what the app sends with PlayView: its user agent, the device and account as base64 protobuf
// headers, and gRPC's own. Host is grpc.biliapi.net for the bangumi endpoint as well, as the app does.
func appHeader(token string) http.Header {
	b64 := base64.StdEncoding.EncodeToString
	h := http.Header{}
	for k, v := range map[string]string{
		"Host":                   "grpc.biliapi.net",
		"User-Agent":             fmt.Sprintf("Dalvik/%s (Linux; U; Android %s; %s %s) %s os/android model/%s mobi_app/android build/%d channel/%s innerVer/%d osVer/%s network/2 grpc-java-cronet/%s", appDalvik, appOSVersion, appBrand, appModel, appVersion, appBrand, appBuild, appChannel, appBuild, appOSVersion, appCronet),
		"Content-Type":           "application/grpc",
		"Te":                     "trailers",
		"X-Bili-Fawkes-Req-Bin":  b64(fawkesBin()),
		"X-Bili-Metadata-Bin":    b64(metadataBin(token)),
		"Authorization":          "identify_v1 " + token,
		"X-Bili-Device-Bin":      b64(deviceBin()),
		"X-Bili-Network-Bin":     b64(networkBin()),
		"X-Bili-Restriction-Bin": "",
		"X-Bili-Locale-Bin":      b64(localeBin()),
		"X-Bili-Exps-Bin":        "",
		"Grpc-Encoding":          "gzip",
		"Grpc-Accept-Encoding":   "identity,gzip",
		"Grpc-Timeout":           "17996161u",
	} {
		h[k] = []string{v}
	}
	return h
}

func (e *Extractor) grpc(ctx context.Context, endpoint string, header http.Header, message []byte) (pbMessage, error) {
	body, err := e.http.Post(ctx, endpoint, header, packFrame(message))
	if err != nil {
		return nil, err
	}
	payload, err := unpackFrame(body)
	if err != nil {
		return nil, err
	}
	m, err := decodeProto(payload)
	if err != nil {
		return nil, errs.New("Invalid APP API answer: %v", err)
	}
	return m, nil
}

// appPlayView asks the APP API for an entry's streams, reshaped into the web API's JSON so one parser reads both.
func (e *Extractor) appPlayView(ctx context.Context, s *session, r *ref) (string, error) {
	aid, _ := strconv.ParseInt(r.aid, 10, 64)
	cid, _ := strconv.ParseInt(r.cid, 10, 64)
	endpoint, req := appUGCEndpoint, playViewRequest(aid, cid, codeType(e.opts.PreferredCodec))
	if r.kind.pgc() {
		if c := e.opts.PreferredCodec; c != "" && c != "HEVC" {
			console.Warn("The APP API offers bangumi only in HEVC")
		}
		epID, _ := strconv.ParseInt(r.epid, 10, 64)
		endpoint, req = appPGCEndpoint, playViewRequest(epID, cid, code265)
	}
	reply, err := e.grpc(ctx, endpoint, appHeader(s.token), req)
	if err != nil {
		return "", err
	}
	return playViewJSON(reply).Serialized(), nil
}

// dmView asks the danmaku service for a page's subtitles; it answers without a web cookie.
func (e *Extractor) dmView(ctx context.Context, aid, cid string) (pbMessage, error) {
	a, _ := strconv.ParseInt(aid, 10, 64)
	c, _ := strconv.ParseInt(cid, 10, 64)
	h := http.Header{}
	h.Set("Content-Type", "application/grpc")
	h.Set("User-Agent", androidUserAgent)
	h.Set("Grpc-Encoding", "gzip")
	return e.grpc(ctx, appDmViewEndpoint, h, dmViewRequest(a, c))
}

// dashItemJSON is a DashItem (id 1, baseUrl 2, backupUrl 3, bandwidth 4, size 7) as a web DASH node.
func dashItemJSON(item pbMessage, codec string) map[string]any {
	return map[string]any{
		"id":         int64(item.uint(1)),
		"base_url":   item.string(2),
		"backup_url": stringsJSON(item.strings(3)),
		"bandwidth":  int64(item.uint(4)),
		"size":       int64(item.uint(7)),
		"codecs":     codec,
	}
}

func stringsJSON(s []string) []any {
	out := make([]any, len(s))
	for i, v := range s {
		out[i] = v
	}
	return out
}

// playViewJSON reshapes a PlayViewReply into the web playurl's JSON: data.dash.{video,audio}, the bangumi
// opening / ending markers as data.clip_info_list, and dubbing as dubbing_info.
func playViewJSON(reply pbMessage) jsonv.Value {
	info := reply.message(1)          // videoInfo
	timelength := int64(info.uint(3)) // milliseconds
	secs := max(timelength/1000, 1)

	videos := []any{}
	for _, stream := range info.messages(5) { // streamList
		if !stream.has(2) {
			continue // segment video only
		}
		dv := stream.message(2) // dashVideo: baseUrl 1, backupUrl 2, codecid 4, size 6
		videos = append(videos, map[string]any{
			"id":         int64(stream.message(1).uint(1)), // streamInfo.quality
			"base_url":   dv.string(1),
			"backup_url": stringsJSON(dv.strings(2)),
			"bandwidth":  int64(dv.uint(6)) * 8 / secs,
			"codecid":    int64(dv.uint(4)),
			"size":       int64(dv.uint(6)),
		})
	}
	audios := []any{}
	for _, a := range info.messages(6) { // dashAudio
		audios = append(audios, dashItemJSON(a, "M4A"))
	}
	if flac := info.message(9); flac.has(2) {
		audios = append(audios, dashItemJSON(flac.message(2), "FLAC"))
	}
	if dolby := info.message(7); dolby.has(2) {
		audios = append(audios, dashItemJSON(dolby.message(2), "E-AC-3"))
	}

	clips := []any{}
	for _, c := range reply.message(3).messages(6) { // business.clipInfo: start 2, end 3, toastText 5
		clips = append(clips, map[string]any{"start": c.int(2), "end": c.int(3), "toastText": c.string(5)})
	}

	background, roles := []any{}, []any{}
	dub := reply.message(7).message(1) // playExtInfo.playDubbingInfo
	if dub.has(1) {
		for _, a := range dub.message(1).messages(7) { // backgroundAudio.audio
			background = append(background, dashItemJSON(a, "M4A"))
		}
		for _, role := range dub.messages(2) { // roleAudioList
			for _, m := range role.messages(4) { // audioMaterialList: audioId 1, title 2, edition 3, personName 5, audio 7
				title := m.string(1)
				if m.has(2) {
					title = m.string(2)
				}
				person := ""
				if m.has(5) {
					person = m.string(5)
				} else if m.has(3) {
					person = m.string(3)
				}
				tracks := []any{}
				for _, a := range m.messages(7) {
					tracks = append(tracks, dashItemJSON(a, "M4A"))
				}
				roles = append(roles, map[string]any{"audio_id": m.string(1), "title": title, "person_name": person, "audio": tracks})
			}
		}
	}

	return jsonv.Of(map[string]any{
		"code":    int64(0),
		"message": "0",
		"ttl":     int64(1),
		"data": map[string]any{
			"timelength":     timelength,
			"dash":           map[string]any{"video": videos, "audio": audios},
			"clip_info_list": clips,
		},
		"dubbing_info": map[string]any{"background_audio": background, "role_audio_list": roles},
	})
}
