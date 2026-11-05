package repo

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type AssetID int64

// AssetIDRange is a range of asset IDs, both ends inclusive. A range with Min
// above Max matches nothing.
type AssetIDRange struct {
	Min, Max AssetID
}

// assetIDSearchRe matches a search term of one or two asset ID comparisons,
// such as "#>=100" or "asset-id >= 000-100 <= 000-200".
var assetIDSearchRe = regexp.MustCompile(`(?i)^(?:#|asset[- ]?id)\s*(>=|<=|>|<)\s*([0-9-]+)(?:\s*(>=|<=|>|<)\s*([0-9-]+))?$`)

func (aid AssetID) Nil() bool {
	return int64(aid) <= 0
}

func ParseAssetIDBytes(d []byte) (aid AssetID, ok bool) {
	d = bytes.ReplaceAll(d, []byte(`"`), []byte(``))
	d = bytes.ReplaceAll(d, []byte(`-`), []byte(``))

	aidInt, err := strconv.Atoi(string(d))
	if err != nil {
		return AssetID(-1), false
	}

	return AssetID(aidInt), true
}

func ParseAssetID(s string) (aid AssetID, ok bool) {
	return ParseAssetIDBytes([]byte(s))
}

// ParseAssetIDSearch reads a search term of one or two asset ID comparisons.
// Any other term returns false. Min is always at least 1, so entities without
// an asset ID never match.
func ParseAssetIDSearch(s string) (AssetIDRange, bool) {
	m := assetIDSearchRe.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return AssetIDRange{}, false
	}

	r := AssetIDRange{Min: 1, Max: math.MaxInt64}
	if !r.tighten(m[1], m[2]) {
		return AssetIDRange{}, false
	}
	if m[3] != "" && !r.tighten(m[3], m[4]) {
		return AssetIDRange{}, false
	}
	return r, true
}

// tighten applies one comparison to the range and never loosens it. It returns
// false when the value isn't a valid asset ID.
func (r *AssetIDRange) tighten(op, value string) bool {
	n, ok := ParseAssetID(value)
	if !ok {
		return false
	}

	switch op {
	case ">=":
		r.Min = max(r.Min, n)
	case ">":
		if n == math.MaxInt64 {
			// Nothing is above the highest ID, and n+1 would overflow.
			r.Max = 0
		} else {
			r.Min = max(r.Min, n+1)
		}
	case "<=":
		r.Max = min(r.Max, n)
	case "<":
		r.Max = min(r.Max, n-1)
	}
	return true
}

// ResolveAssetIDSearch turns a search term that names asset IDs into the
// matching filter and clears the term. Any other term stays a text search.
// Every place that reads a search term resolves it here, so they can't disagree.
func (q *EntityQuery) ResolveAssetIDSearch() {
	if aidStr, ok := strings.CutPrefix(q.Search, "#"); ok {
		if aid, ok := ParseAssetID(aidStr); ok {
			q.Search = ""
			q.AssetID = aid
			return
		}
	}

	if r, ok := ParseAssetIDSearch(q.Search); ok {
		q.Search = ""
		q.AssetIDRange = &r
	}
}

func (aid AssetID) String() string {
	if aid.Nil() {
		return ""
	}

	aidStr := fmt.Sprintf("%06d", aid)
	aidStr = fmt.Sprintf("%s-%s", aidStr[:3], aidStr[3:])
	return aidStr
}

func (aid AssetID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + aid.String() + `"`), nil
}

func (aid *AssetID) UnmarshalJSON(d []byte) error {
	if len(d) == 0 || bytes.Equal(d, []byte(`""`)) {
		*aid = -1
		return nil
	}

	d = bytes.ReplaceAll(d, []byte(`"`), []byte(``))
	d = bytes.ReplaceAll(d, []byte(`-`), []byte(``))

	aidInt, err := strconv.Atoi(string(d))
	if err != nil {
		return err
	}

	*aid = AssetID(aidInt)
	return nil
}

func (aid AssetID) MarshalCSV() (string, error) {
	return aid.String(), nil
}

func (aid *AssetID) UnmarshalCSV(d string) error {
	return aid.UnmarshalJSON([]byte(d))
}
