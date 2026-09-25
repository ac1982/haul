package bilibili

import (
	"crypto/md5"
	"encoding/hex"
	"path"
	"strings"

	"github.com/ac1982/haul/internal/errs"
)

// App keys and secrets for the appkey-signed endpoints: the TV client's, and BiliPlus-style proxies'.
const (
	tvAppKey       = "4409e2ce8ffd12b8"
	tvAppSecret    = "59b43e04ad6965f34319062b478f83dd"
	biliPlusAppKey = "7d089525d3611b1c"
	biliPlusSecret = "acd495b248ec528c2eed1e862d393126"
)

// mixinKeyTable picks the WBI key's characters out of the two image keys the nav endpoint hands out.
var mixinKeyTable = [...]int{
	46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35,
	27, 43, 5, 49, 33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13,
}

// wbiMixinKey is the WBI signing key made from nav's img and sub keys.
func wbiMixinKey(imgKey, subKey string) string {
	source := imgKey + subKey
	var b strings.Builder
	for _, i := range mixinKeyTable {
		if i < len(source) {
			b.WriteByte(source[i])
		}
	}
	return b.String()
}

// wbiSign appends w_rid to a query that already ends in wts=.
func wbiSign(query, key string) string { return query + "&w_rid=" + md5Hex(query+key) }

// appSign is the appkey-style signature: md5(query + secret).
func appSign(query, secret string) string { return md5Hex(query + secret) }

func md5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

// imageKey is a bfs image URL's file name without its extension: …/abc.png → abc.
func imageKey(url string) string {
	name := path.Base(url)
	return strings.TrimSuffix(name, path.Ext(name))
}

// av ⇄ BV, after https://github.com/Colerar/abv.
const (
	bvXor      = 23_442_827_791_579
	bvMask     = 1<<51 - 1
	bvMaxAid   = bvMask + 1
	bvBase     = 58
	bvLength   = 9
	bvAlphabet = "FcwAPNKTMug3GV5Lj7EJnHpWsx4tb8haYeviqBz6rkCy12mUSDQX9RdoZf"
)

// bvEncode is the BV id of an av number.
func bvEncode(aid int64) (string, error) {
	if aid < 1 {
		return "", errs.NewInput("av%d is below 1", aid)
	}
	if aid >= bvMaxAid {
		return "", errs.NewInput("av%d is out of range", aid)
	}
	out := []byte(strings.Repeat("0", bvLength))
	tmp := (bvMaxAid | aid) ^ bvXor
	for i := bvLength - 1; tmp != 0; i-- {
		out[i] = bvAlphabet[tmp%bvBase]
		tmp /= bvBase
	}
	out[0], out[6] = out[6], out[0]
	out[1], out[4] = out[4], out[1]
	return "BV1" + string(out), nil
}

// bvDecode is the av number of a BV id, given the 9 characters after "BV1".
func bvDecode(body string) (int64, error) {
	if len(body) != bvLength {
		return 0, errs.NewInput("BV1%s must be 12 characters", body)
	}
	chars := []byte(body)
	chars[0], chars[6] = chars[6], chars[0]
	chars[1], chars[4] = chars[4], chars[1]
	var aid int64
	for _, c := range chars {
		v := strings.IndexByte(bvAlphabet, c)
		if v < 0 {
			return 0, errs.NewInput("Invalid character in the BV id: %c", c)
		}
		aid = aid*bvBase + int64(v)
	}
	return (aid & bvMask) ^ bvXor, nil
}

// bvid is the BV id of a numeric aid, or "".
func bvid(aid string) string {
	n, ok := parseInt(aid)
	if !ok {
		return ""
	}
	bv, err := bvEncode(n)
	if err != nil {
		return ""
	}
	return bv
}
