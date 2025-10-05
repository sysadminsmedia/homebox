package labelmaker

import (
	"os"
	"path/filepath"
	"testing"

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
	nonExistentPath := "/non/existent/path/font.ttf"
	cfg := &config.Config{
		LabelMaker: config.LabelMakerConf{
			RegularFontPath: &nonExistentPath,
		},
	}

	font, err := loadFont(cfg, FontTypeRegular)
	require.NoError(t, err)
	assert.NotNil(t, font)
}

func TestLoadFont_UnknownFontType(t *testing.T) {
	cfg := &config.Config{}

	_, err := loadFont(cfg, FontType(999))
	assert.Error(t, err)
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
	emptyPath := ""
	cfg := &config.Config{
		LabelMaker: config.LabelMakerConf{
			RegularFontPath: &emptyPath,
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
			text:            "中문",
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
