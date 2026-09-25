package bilibili

import (
	"regexp"
	"strings"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/media"
)

var (
	pcdnHost   = regexp.MustCompile(`://[^/]*:\d+/`)
	akamaiHost = regexp.MustCompile(`://[^/]*akamaized\.net/`)
	anyHost    = regexp.MustCompile(`://[^/]+/`)
)

// rewriteHost moves a stream off hosts that tend to fail: PCDN (ip:port) hosts unless allowed, and, behind an
// area proxy, the overseas akamaized hosts. UposHost, or the backup mirror when ReplaceHost is on, replaces
// every host.
func rewriteHost(u string, o *Options, area string) string {
	upos := o.UposHost
	if upos == "" && o.ReplaceHost {
		upos = backupHost
	}
	if upos != "" {
		console.Debugf("Stream host set to %s", upos)
		return anyHost.ReplaceAllString(u, "://"+upos+"/")
	}
	if !o.AllowPCDN && pcdnHost.MatchString(u) {
		console.Debugf("Stream is on a PCDN host; using %s", backupHost)
		u = pcdnHost.ReplaceAllString(u, "://"+backupHost+"/")
	}
	if area != "" && strings.Contains(u, "akamaized.net") {
		console.Debugf("Stream is on an overseas host; using %s", backupHost)
		u = akamaiHost.ReplaceAllString(u, "://"+backupHost+"/")
	}
	return u
}

// forceHTTP switches a stream to plain HTTP, except on *.mcdn.bilivideo.cn:port hosts, which only speak what
// they advertise.
func forceHTTP(u string) string {
	if strings.Contains(u, ".mcdn.bilivideo.cn:") {
		return u
	}
	if rest, ok := strings.CutPrefix(u, "https:"); ok {
		return "http:" + rest
	}
	return u
}

// resource prepares a stream for download: its host (FLV segments keep theirs), its scheme, its headers and how
// it may be fetched. cmcc hosts break on parallel ranges and on the switch to HTTP, so they get one plain request.
// bilibili's stated sizes are not trusted to drive ranges, so Size stays 0.
func (e *Extractor) resource(s *session, u string, rewrite bool) media.Resource {
	if u == "" {
		return media.Resource{}
	}
	if rewrite {
		u = rewriteHost(u, &e.opts, s.area)
	}
	policy := media.Parallel
	switch {
	case strings.Contains(u, "-cmcc-"):
		policy = media.Whole
	case e.opts.ForceHTTP:
		u = forceHTTP(u)
	}
	return media.Resource{URL: u, Header: s.mediaHeader(u), Policy: policy}
}

// prepare turns every stream URL of f into a ready resource.
func (e *Extractor) prepare(s *session, f *media.Formats) {
	for i := range f.Video {
		v := &f.Video[i]
		if v.Source.URL != "" {
			v.Source = e.resource(s, v.Source.URL, true)
		}
		for j := range v.Parts {
			v.Parts[j] = e.resource(s, v.Parts[j].URL, false)
		}
	}
	for i := range f.Audio {
		f.Audio[i].Source = e.resource(s, f.Audio[i].Source.URL, true)
	}
	for i := range f.ExtraAudio {
		f.ExtraAudio[i].Audio.Source = e.resource(s, f.ExtraAudio[i].Audio.Source.URL, true)
	}
}
