package labelmaker

import (
	"bytes"
	"image"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/golang/freetype/truetype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gomedium"
)

func TestLoadFont_WithNilConfig(t *testing.T) {
	font, err := loadFont(nil, FontTypeRegular)
	require.NoError(t, err)
	assert.NotNil(t, font)

	font, err = loadFont(nil, FontTypeBold)
	require.NoError(t, err)
	assert.NotNil(t, font)
}

func TestLoadFont_WithEmptyConfig(t *testing.T) {
	cfg := &config.Config{}

	font, err := loadFont(cfg, FontTypeRegular)
	require.NoError(t, err)
	assert.NotNil(t, font)

	font, err = loadFont(cfg, FontTypeBold)
	require.NoError(t, err)
	assert.NotNil(t, font)
}

func TestLoadFont_WithCustomFontPath(t *testing.T) {
	tempDir := t.TempDir()
	fontPath := filepath.Join(tempDir, "test-font.ttf")

	err := os.WriteFile(fontPath, gomedium.TTF, 0644)
	require.NoError(t, err)

	cfg := &config.Config{
		LabelMaker: config.LabelMakerConf{
			RegularFontPath: &fontPath,
		},
	}

	font, err := loadFont(cfg, FontTypeRegular)
	require.NoError(t, err)
	assert.NotNil(t, font)
}

func TestLoadFont_WithNonExistentFontPath(t *testing.T) {
	cfg := &config.Config{
		LabelMaker: config.LabelMakerConf{
			RegularFontPath: new("/non/existent/path/font.ttf"),
		},
	}

	font, err := loadFont(cfg, FontTypeRegular)
	require.NoError(t, err)
	assert.NotNil(t, font)
}

func TestLoadFont_UnknownFontType(t *testing.T) {
	cfg := &config.Config{}

	_, err := loadFont(cfg, FontType(999))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown font type")
}

func TestLoadFont_BoldFontWithCustomPath(t *testing.T) {
	tempDir := t.TempDir()
	fontPath := filepath.Join(tempDir, "test-bold-font.ttf")

	err := os.WriteFile(fontPath, gobold.TTF, 0644)
	require.NoError(t, err)

	cfg := &config.Config{
		LabelMaker: config.LabelMakerConf{
			BoldFontPath: &fontPath,
		},
	}

	font, err := loadFont(cfg, FontTypeBold)
	require.NoError(t, err)
	assert.NotNil(t, font)
}

func TestLoadFont_EmptyStringPath(t *testing.T) {
	cfg := &config.Config{
		LabelMaker: config.LabelMakerConf{
			RegularFontPath: new(""),
		},
	}

	font, err := loadFont(cfg, FontTypeRegular)
	require.NoError(t, err)
	assert.NotNil(t, font)
}

func TestLoadFont_CJKRendering(t *testing.T) {
	cjkFontPath := filepath.Join("testdata", "NotoSansKR-VF.ttf")

	tests := []struct {
		name            string
		text            string
		fontPath        string
		shouldHaveGlyph bool
	}{
		{
			name:            "Korean with default font",
			text:            "한글",
			fontPath:        "",
			shouldHaveGlyph: false,
		},
		{
			name:            "Chinese with default font",
			text:            "中文",
			fontPath:        "",
			shouldHaveGlyph: false,
		},
		{
			name:            "Japanese with default font",
			text:            "ひらがなカタカナ",
			fontPath:        "",
			shouldHaveGlyph: false,
		},
		{
			name:            "Korean with Noto Sans CJK",
			text:            "한글",
			fontPath:        cjkFontPath,
			shouldHaveGlyph: true,
		},
		{
			name:            "Chinese with Noto Sans CJK",
			text:            "中文",
			fontPath:        cjkFontPath,
			shouldHaveGlyph: true,
		},
		{
			name:            "Japanese with Noto Sans CJK",
			text:            "ひらがなカタカナ",
			fontPath:        cjkFontPath,
			shouldHaveGlyph: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg *config.Config
			if tt.fontPath != "" {
				if _, err := os.Stat(tt.fontPath); os.IsNotExist(err) {
					t.Skipf("Font file not found: %s", tt.fontPath)
				}
				cfg = &config.Config{
					LabelMaker: config.LabelMakerConf{
						RegularFontPath: &tt.fontPath,
					},
				}
			}

			font, err := loadFont(cfg, FontTypeRegular)
			require.NoError(t, err)
			require.NotNil(t, font)

			hasAllGlyphs := true
			for _, r := range tt.text {
				if font.Index(r) == 0 {
					hasAllGlyphs = false
					break
				}
			}

			if tt.shouldHaveGlyph {
				assert.True(t, hasAllGlyphs, "Font should render %s characters", tt.name)
			} else {
				assert.False(t, hasAllGlyphs, "Default font should not render %s characters", tt.name)
			}
		})
	}
}

func TestSplitCommandTemplate(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", nil},
		{"whitespace only", "   \t ", nil},
		{"simple", "lp -d printer file.png", []string{"lp", "-d", "printer", "file.png"}},
		{"collapses runs", "lp    -d   printer", []string{"lp", "-d", "printer"}},
		{
			"placeholder without inner space",
			"lp -t {{.TitleText}} {{.FileName}}",
			[]string{"lp", "-t", "{{.TitleText}}", "{{.FileName}}"},
		},
		{
			"placeholder with inner space stays one token",
			"lp -o {{ .FileName }}",
			[]string{"lp", "-o", "{{ .FileName }}"},
		},
		{
			"placeholder glued to literal stays one token",
			"convert prefix-{{.TitleText}}-suffix",
			[]string{"convert", "prefix-{{.TitleText}}-suffix"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, splitCommandTemplate(tt.in))
		})
	}
}

// TestWrapText_EmptyTextWithNoHeightBudget is the regression test for a panic
// ("slice bounds out of range [1:0]") that fired when wrapText was given an
// empty string and a maxHeight too small to fit even one line. This happens
// in practice when HBOX_LABEL_MAKER_HIDE_ASSET_ID leaves the description
// empty (no location) and the promoted item-name title wraps to enough lines
// to exceed the QR code's height, leaving zero (or negative) budget for the
// description text beside it.
func TestWrapText_EmptyTextWithNoHeightBudget(t *testing.T) {
	fnt, err := loadFont(nil, FontTypeRegular)
	require.NoError(t, err)

	face := truetype.NewFace(fnt, &truetype.Options{Size: 24, DPI: 72})
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	ctx := createContext(fnt, 24, img, 72)

	require.NotPanics(t, func() {
		lines, remaining := wrapText("", face, 200, -100, 24, ctx)
		assert.Empty(t, lines)
		assert.Empty(t, remaining)
	})
}

// TestGenerateLabel_EmptyDescriptionWithTallTitle reproduces the real crash
// end to end: an asset label with HideAssetID promotes a long item name to
// the bold title slot with an empty description, which used to panic inside
// GenerateLabel -> wrapText.
func TestGenerateLabel_EmptyDescriptionWithTallTitle(t *testing.T) {
	params := NewGenerateParams(300, 80, 10, 10, 32,
		"A Really Rather Long Item Name That Wraps Several Times",
		"", "http://localhost/item/1", true, nil)

	var buf bytes.Buffer
	require.NotPanics(t, func() {
		require.NoError(t, GenerateLabel(&buf, &params, nil))
	})
	assert.NotZero(t, buf.Len())
}

// TestPrintLabel_ArgumentInjection is the regression test for the argument
// injection where a user-controlled item name containing whitespace could
// smuggle extra argv entries into the print binary. The fix renders each argv
// token independently, so a malicious TitleText must arrive as a single
// argument.
func TestPrintLabel_ArgumentInjection(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("recorder script relies on a POSIX shell")
	}

	dir := t.TempDir()
	argsFile := filepath.Join(dir, "args.txt")
	recorder := filepath.Join(dir, "recorder.sh")
	script := "#!/bin/sh\n: > '" + argsFile + "'\nfor a in \"$@\"; do printf '%s\\n' \"$a\" >> '" + argsFile + "'; done\n"
	require.NoError(t, os.WriteFile(recorder, []byte(script), 0o700))

	// The admin template separates two arguments: the (attacker-controlled)
	// title and the generated file name.
	printCommand := recorder + " {{.TitleText}} {{.FileName}}"

	cfg := &config.Config{}
	cfg.LabelMaker.PrintCommand = &printCommand

	const maliciousTitle = "evil title --output /etc/passwd"
	params := NewGenerateParams(400, 200, 10, 5, 32, maliciousTitle, "desc", "http://localhost/item/1", false, nil)

	require.NoError(t, PrintLabel(cfg, &params))

	raw, err := os.ReadFile(argsFile)
	require.NoError(t, err)
	got := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")

	// Exactly two arguments: the whole title as one entry, then the file name.
	require.Len(t, got, 2, "argv must not be split on whitespace inside the title")
	assert.Equal(t, maliciousTitle, got[0])
	assert.True(t, strings.HasSuffix(got[1], ".png"), "second arg should be the generated label file, got %q", got[1])
}
