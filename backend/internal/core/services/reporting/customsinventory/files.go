package customsinventory

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"
)

// FormatAssetID renders an asset ID like repo.AssetID.String: 123 → "000-123".
func FormatAssetID(id int64) string {
	if id <= 0 {
		return ""
	}
	s := fmt.Sprintf("%06d", id)
	return s[:3] + "-" + s[3:]
}

const maxSlugLen = 60

var slugReplacer = strings.NewReplacer("ä", "ae", "ö", "oe", "ü", "ue", "ß", "ss")

// Slug turns a title into a short ASCII file name part (FR-20).
func Slug(title string) string {
	s := slugReplacer.Replace(strings.ToLower(title))
	var b strings.Builder
	dash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
		} else if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	out := b.String()
	if len(out) > maxSlugLen {
		out = out[:maxSlugLen]
	}
	out = strings.Trim(out, "-")
	if out == "" {
		return "export"
	}
	return out
}

// BaseName is the file name without extension: Inventarliste_{slug}_{yyyy-mm-dd}, in UTC.
func BaseName(title string, generated time.Time) string {
	return "Inventarliste_" + Slug(title) + "_" + generated.UTC().Format(time.DateOnly)
}

var (
	extPattern = regexp.MustCompile(`^[a-z0-9]{1,5}$`)
	mimeExts   = map[string]string{
		"application/pdf": ".pdf",
		"image/jpeg":      ".jpg",
		"image/png":       ".png",
		"image/heic":      ".heic",
	}
)

// FileExt picks a receipt's extension: from its title if that has a short
// alphanumeric one, else from the MIME type, else ".bin" (FR-19).
func FileExt(title, mime string) string {
	if ext := strings.ToLower(strings.TrimPrefix(path.Ext(title), ".")); extPattern.MatchString(ext) {
		return "." + ext
	}
	mime, _, _ = strings.Cut(mime, ";")
	if ext, ok := mimeExts[strings.ToLower(strings.TrimSpace(mime))]; ok {
		return ext
	}
	return ".bin"
}

// dupLetters renders n ≥ 1 in bijective base 26: 1 → "a", 2 → "b", 27 → "aa".
func dupLetters(n int) string {
	var s []byte
	for n > 0 {
		n--
		s = append([]byte{byte('a' + n%26)}, s...)
		n /= 26
	}
	return string(s)
}

// ReceiptName builds "Belege/{assetLabel}{dup}_Rechnung{suffix}{ext}".
// dupIndex 0 → "", 1 → "_b", 2 → "_c"…; index 0 → "", 1 → "_2"….
func ReceiptName(assetLabel string, dupIndex, index int, title, mime string) string {
	dup, suffix := "", ""
	if dupIndex > 0 {
		dup = "_" + dupLetters(dupIndex+1)
	}
	if index > 0 {
		suffix = fmt.Sprintf("_%d", index+1)
	}
	return "Belege/" + assetLabel + dup + "_Rechnung" + suffix + FileExt(title, mime)
}
