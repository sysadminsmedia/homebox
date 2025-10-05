# Custom Font Support for Label Maker

Homebox label maker now supports external font files. **CJK (Chinese, Japanese, Korean) characters require a custom font** - the default bundled fonts only support Latin characters.

## Quick Start

### Docker/Podman Setup

1. **Download custom fonts** (e.g., Noto Sans KR):
   - Download from [Google Fonts](https://fonts.google.com/noto/specimen/Noto+Sans+KR)
   - Or use the Variable Font from [GitHub](https://github.com/notofonts/noto-cjk)

2. **Create a fonts directory**:
```bash
mkdir -p ./fonts
# Place your font files in this directory
# e.g., NotoSansKR-VF.ttf
```

3. **Mount the fonts directory and set environment variables**:
```yaml
# docker-compose.yml
services:
  homebox:
    image: homebox:latest
    volumes:
      - ./data:/data
      - ./fonts:/fonts:ro  # Mount fonts directory as read-only
    environment:
      - HBOX_LABEL_MAKER_REGULAR_FONT_PATH=/fonts/NotoSansKR-VF.ttf
      - HBOX_LABEL_MAKER_BOLD_FONT_PATH=/fonts/NotoSansKR-VF.ttf
    ports:
      - 3100:7745
```

Or with podman:
```bash
podman run -d \
  --name homebox \
  -p 3100:7745 \
  -v ./data:/data \
  -v ./fonts:/fonts:ro \
  -e HBOX_LABEL_MAKER_REGULAR_FONT_PATH=/fonts/NotoSansKR-VF.ttf \
  -e HBOX_LABEL_MAKER_BOLD_FONT_PATH=/fonts/NotoSansKR-VF.ttf \
  homebox:latest
```

4. **Restart the container** and test label generation with Chinese, Japanese, Korean text!

## Supported Fonts

- **Format**: TTF (TrueType Font)
- **Recommended Fonts**:
  - Noto Sans KR (Korean)
  - Noto Sans CJK (Chinese, Japanese, Korean)
  - Noto Sans SC (Simplified Chinese)
  - Noto Sans JP (Japanese)

## Fallback Behavior

1. **External font specified** → Tries to load from `HBOX_LABEL_MAKER_*_FONT_PATH`
2. **External font fails or not specified** → Falls back to bundled Go fonts (Latin-only, **does not support CJK characters**)

## Troubleshooting

### Labels still show squares (□□□)
- Check if the font file exists at the specified path
- Verify the font file format (must be TTF, not OTF)
- Check container logs: `podman logs homebox | grep -i font`

### Font file not found
- Ensure the volume is correctly mounted
- Check file permissions (font files should be readable)
- Use absolute paths in environment variables

## Why External Fonts?

- **Smaller base image**: No need to embed large font files (~10MB per font)
- **Flexibility**: Easily switch fonts without rebuilding the image
- **Multi-language support**: Add support for any language by mounting appropriate fonts
