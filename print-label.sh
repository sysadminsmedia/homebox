#!/bin/sh
# print_ptouch.sh — Install dependencies if needed, then print a label on a
# Brother P-Touch printer using labelprinterkit.
#
# Usage: print_ptouch.sh [OPTIONS] <image>
#
# Options:
#   --host HOST      Printer hostname or IP (required)
#   --port PORT      TCP port (default: 9100)
#   --model MODEL    Printer model: P700 or P750W (default: P700)
#   --media MEDIA    Tape width: W3_5 W6 W9 W12 W18 W24 (default: W24)
#   --rotate DEGREES Rotate image CCW before printing (default: 90)
#
# Example HBOX_LABEL_MAKER_PRINT_COMMAND:
#   /config/print_ptouch.sh --host 192.168.1.50 --model P700 --media W24
#
# Available env vars (set automatically by Homebox):
#   LABEL_FILENAME        — path to the generated PNG temp file
#   LABEL_LABELNAME       — name of the item or location
#   LABEL_LABELURL        — URL to the item or location page
#   LABEL_ITEMNAME        — item name (empty for location labels)
#   LABEL_ITEMLOCATION    — parent location name
#   LABEL_ITEMDESCRIPTION — item description text
#   LABEL_ITEMTAGS        — comma-separated tag names
#   LABEL_ITEMASSETID     — asset ID (e.g. 000-001)

set -e

# ---------------------------------------------------------------------------
# Install system dependencies if missing
# ---------------------------------------------------------------------------
install_deps() {
    if command -v apk >/dev/null 2>&1; then
        echo "Installing dependencies via apk..." >&2
        apk add --no-cache python3 python3-dev py3-pip build-base zlib-dev jpeg-dev libjpeg-turbo-dev libpng-dev
    elif command -v apt-get >/dev/null 2>&1; then
        echo "Installing dependencies via apt-get..." >&2
        apt-get update -qq
        apt-get install -y -qq python3 python3-dev python3-pip build-essential zlib1g-dev libjpeg-dev libpng-dev
    else
        echo "ERROR: No supported package manager found (apk, apt-get)" >&2
        exit 1
    fi
}

# Check if we have python3 and necessary build tools
if ! command -v python3 >/dev/null 2>&1; then
    echo "Python3 not found, installing..." >&2
    install_deps
fi

# Check if we have build-base/build-essential (needed for compiling Pillow)
if ! command -v gcc >/dev/null 2>&1; then
    echo "Build tools not found, installing..." >&2
    install_deps
fi

# Check if we have the necessary image library headers (needed for Pillow)
if [ ! -f /usr/include/zlib.h ] || [ ! -f /usr/include/jpeglib.h ] || [ ! -f /usr/include/png.h ] || [ ! -f /usr/include/freetype2/ft2build.h ]; then
    echo "Image library development headers not found, installing..." >&2
    if command -v apk >/dev/null 2>&1; then
        apk add --no-cache zlib-dev jpeg-dev libjpeg-turbo-dev libpng-dev freetype-dev font-liberation
    elif command -v apt-get >/dev/null 2>&1; then
        apt-get update -qq
        apt-get install -y -qq zlib1g-dev libjpeg-dev libpng-dev libfreetype6-dev fonts-liberation
    fi
fi

# ---------------------------------------------------------------------------
# Setup virtual environment and install labelprinterkit if missing
# ---------------------------------------------------------------------------
VENV_DIR="/tmp/labelprinterkit-venv"
if [ ! -d "$VENV_DIR" ]; then
    echo "Creating virtual environment..." >&2
    python3 -m venv "$VENV_DIR"
fi

# Activate venv and install packages
. "$VENV_DIR/bin/activate"
if ! python3 -c "import labelprinterkit" >/dev/null 2>&1; then
    echo "labelprinterkit not found, installing..." >&2
    pip install --quiet "labelprinterkit==0.7.1" numpy qrcode
fi
# Ensure Pillow was compiled with FreeType support (reinstall if not)
if ! python3 -c "from PIL import _imagingft" >/dev/null 2>&1; then
    echo "Reinstalling Pillow with FreeType support..." >&2
    pip install --quiet --force-reinstall --no-binary :all: Pillow
fi

# ---------------------------------------------------------------------------
# Print the label (Python embedded below, arguments forwarded via "$@")
# ---------------------------------------------------------------------------
exec python3 - "$@" <<'PYTHON'
import argparse
import os
import qrcode
from PIL import Image, ImageDraw, ImageFont
from labelprinterkit.backends.network import TCPBackend
from labelprinterkit.printers import P700, P750W
from labelprinterkit.label import Item, Label
from labelprinterkit.job import Job
from labelprinterkit.constants import Media, Resolution

MEDIA_MAP = {
    "W3_5": Media.W3_5,
    "W6":   Media.W6,
    "W9":   Media.W9,
    "W12":  Media.W12,
    "W18":  Media.W18,
    "W24":  Media.W24,
}

PRINTER_MAP = {
    "P700":  P700,
    "P750W": P750W,
}

class ImageItem(Item):
    def __init__(self, image):
        self._image = image
    def render(self):
        return self._image

parser = argparse.ArgumentParser()
parser.add_argument("--host", required=True)
parser.add_argument("--port", type=int, default=9100)
parser.add_argument("--model", default="P750W", choices=list(PRINTER_MAP))
parser.add_argument("--media", default="W18", choices=list(MEDIA_MAP))
args = parser.parse_args()

url = os.environ.get("LABEL_URL", "")
if not url:
    raise SystemExit("ERROR: LABEL_URL env var is not set")

asset_id = os.environ.get("LABEL_AssetID", "")
name = os.environ.get("LABEL_Name", "")

media = MEDIA_MAP[args.media]
printable_width = media.value.printarea  # pixels tall (tape feeds horizontally)

# --- QR code: fill the full tape height, pixel-perfect with NEAREST. ---
qr = qrcode.QRCode(error_correction=qrcode.constants.ERROR_CORRECT_M, box_size=1, border=0)
qr.add_data(url)
qr.make(fit=True)
qr_img = qr.make_image(fill_color="black", back_color="white").convert("RGB")
qr_img = qr_img.resize((printable_width, printable_width), Image.NEAREST)

# --- Fonts. ---
# Name: condensed bold — fits more text in the available width.
# Asset ID: monospaced bold — makes the formatted ID easy to read.
def load_font(path, size):
    try:
        return ImageFont.truetype(path, size)
    except (OSError, ImportError):
        return ImageFont.load_default()

# Supersample scale: render text at 4x then shrink with LANCZOS before
# thresholding — gives antialiased outlines that survive the 1-bit conversion
# cleanly instead of getting dithered into noise.
SS = 4

# Scale font sizes relative to tape height, applied at supersampled resolution.
name_size  = max(10, int(printable_width * 0.30)) * SS
asset_size = max(8,  int(printable_width * 0.22)) * SS
name_font  = load_font("/usr/share/fonts/liberation/LiberationSans-Bold.ttf", name_size)
asset_font = load_font("/usr/share/fonts/liberation/LiberationMono-Bold.ttf", asset_size)

padding_ss  = max(4, int(printable_width * 0.06)) * SS
line_gap_ss = max(2, int(printable_width * 0.04)) * SS
pw_ss       = printable_width * SS

# --- Wrap name into up to 3 lines, shrinking canvas width to fit. ---
# Strategy: try progressively narrower max-widths (fewer chars per line) so the
# label stays compact. We prefer more lines over a wider label.
dummy_draw = ImageDraw.Draw(Image.new("L", (1, 1)))

def wrap_text(text, font, max_px_wide):
    """Word-wrap text to fit max_px_wide pixels. Returns list of lines."""
    words = text.split()
    lines, current = [], ""
    for word in words:
        trial = (current + " " + word).strip()
        if dummy_draw.textlength(trial, font=font) <= max_px_wide:
            current = trial
        else:
            if current:
                lines.append(current)
            current = word
    if current:
        lines.append(current)
    return lines

asset_w = int(dummy_draw.textlength(asset_id, font=asset_font)) if asset_id else 0

def best_wrap(text, font, max_lines=3):
    """
    Find the narrowest canvas width that wraps text into at most max_lines,
    distributing line lengths as evenly as possible.
    Returns (lines, width_px).
    A single word that won't fit is never dropped — the width expands to fit it.
    """
    words = text.split()
    if not words:
        return [], 0

    # Measure each word so we know the minimum possible line width.
    word_widths = [int(dummy_draw.textlength(w, font=font)) for w in words]
    min_word_w  = max(word_widths)  # no line can be narrower than the widest word

    # Single-line width is the upper bound.
    single_w = int(dummy_draw.textlength(text, font=font))

    best_lines = [text]
    best_w     = single_w

    # Binary-search for the narrowest width that still wraps into ≤ max_lines.
    lo, hi = min_word_w, single_w
    while lo < hi:
        mid = (lo + hi) // 2
        candidate = wrap_text(text, font, mid)
        if len(candidate) <= max_lines:
            best_lines = candidate
            best_w     = mid
            hi         = mid - 1
        else:
            lo = mid + 1

    return best_lines, best_w

name_lines, name_content_w = best_wrap(name, name_font) if name else ([], 0)
text_w_ss = max(name_content_w, asset_w) + padding_ss * 2

# --- Render text block at SS resolution. ---
# Layout: asset ID pinned to bottom, name fills the space above it.
text_img = Image.new("L", (text_w_ss, pw_ss), 255)
draw     = ImageDraw.Draw(text_img)

top_margin  = max(padding_ss // 2, line_gap_ss)
asset_y     = pw_ss - padding_ss // 2 - asset_size
name_budget = asset_y - top_margin - line_gap_ss

total_name_h = lambda n, sz: n * sz + (n - 1) * line_gap_ss

# Shrink font size until the name lines fit the vertical budget.
name_size_min     = max(8 * SS, int(name_size * 0.40))
current_name_size = name_size
current_name_font = name_font
while name_lines and total_name_h(len(name_lines), current_name_size) > name_budget:
    next_size = current_name_size - SS
    if next_size < name_size_min:
        break
    current_name_size = next_size
    current_name_font = load_font(
        "/usr/share/fonts/liberation/LiberationSans-Bold.ttf", current_name_size
    )
    # Re-wrap at new size; the optimal width may also change.
    name_lines, _ = best_wrap(name, current_name_font)

y = top_margin
for line in name_lines:
    draw.text((padding_ss, y), line, fill=0, font=current_name_font)
    y += current_name_size + line_gap_ss

if asset_id:
    draw.text((padding_ss, asset_y), asset_id, fill=0, font=asset_font)

# Downscale with LANCZOS then hard-threshold: anything darker than 50% -> black.
text_w  = text_w_ss // SS
text_img = text_img.resize((text_w, printable_width), Image.LANCZOS)
text_img = text_img.point(lambda p: 0 if p < 128 else 255, "L")

# QR stays pixel-perfect (NEAREST); convert to L for uniform mode before merge.
qr_img = qr_img.convert("L")

# --- Combine QR + text on a white grayscale canvas. ---
gap       = max(4, int(printable_width * 0.04))
total_w   = printable_width + gap + text_w
label_img = Image.new("L", (total_w, printable_width), 255)
label_img.paste(qr_img,  (0, 0))
label_img.paste(text_img, (printable_width + gap, 0))

# Hard threshold -> pure 1-bit: no dithering, no noise.
label_img = label_img.point(lambda p: 0 if p < 128 else 255, "L")

# Trim trailing whitespace from the right edge.
import numpy as np
pixels = np.array(label_img)
cols = np.where(pixels.min(axis=0) < 255)[0]
if len(cols) > 0:
    label_img = label_img.crop((0, 0, cols[-1] + 1, printable_width))

label_img = label_img.convert("1")

job = Job(media, resolution=Resolution.LOW)
job.half_cut = True
label = Label(ImageItem(label_img))
job.add_page(label)

printer = PRINTER_MAP[args.model](TCPBackend(args.host, args.port))
printer.print(job)
PYTHON
