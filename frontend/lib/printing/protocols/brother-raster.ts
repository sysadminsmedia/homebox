// Brother Raster Protocol implementation for QL series label printers
// Based on Brother QL series command reference

interface BrotherPrintOptions {
  width: number; // Label width in mm
  height?: number; // Label height in mm (for die-cut labels)
  copies: number;
  highQuality?: boolean;
  cutAtEnd?: boolean;
}

// Brother QL command constants
const ESC = 0x1b;
const NULL = 0x00;

// Media info for common Brother label sizes
const MEDIA_WIDTHS: Record<number, number> = {
  12: 106, // 12mm = 106 dots
  29: 306, // 29mm = 306 dots
  38: 413, // 38mm = 413 dots
  50: 554, // 50mm = 554 dots
  54: 590, // 54mm = 590 dots
  62: 696, // 62mm = 696 dots
  102: 1164, // 102mm = 1164 dots
};

function invalidate(): number[] {
  // Send 200 null bytes to clear any pending data
  return new Array(200).fill(NULL);
}

function initialize(): number[] {
  // ESC @ - Initialize printer
  return [ESC, 0x40];
}

function switchToRasterMode(): number[] {
  // ESC i a {n} - Switch to raster mode (n=1)
  return [ESC, 0x69, 0x61, 0x01];
}

function setMediaInfo(widthMm: number, heightMm?: number): number[] {
  // ESC i z {flags} {mediaType} {width} {height_low} {height_high} {pageCount_low} {pageCount_high} {startingPage} {0}
  const cmd = [ESC, 0x69, 0x7a];

  // Flags: bit 6 = quality, bit 7 = recover
  const flags = 0x86;
  cmd.push(flags);

  // Media type: 0x0A = continuous, 0x0B = die-cut
  const mediaType = heightMm ? 0x0b : 0x0a;
  cmd.push(mediaType);

  // Width (in mm, single byte)
  cmd.push(widthMm);

  // Height in mm (2 bytes, little endian)
  const height = heightMm || 0;
  cmd.push(height & 0xff);
  cmd.push((height >> 8) & 0xff);

  // Page count (2 bytes) - not used, set to 0
  cmd.push(0, 0);

  // Starting page - 0
  cmd.push(0);

  // Reserved - 0
  cmd.push(0);

  return cmd;
}

function setPrintMode(highQuality: boolean): number[] {
  // ESC i M {n} - Set print mode
  // n=0: Standard, n=1: High quality
  return [ESC, 0x69, 0x4d, highQuality ? 0x01 : 0x00];
}

function setCutOptions(cutAtEnd: boolean): number[] {
  // ESC i K {n} - Set cut options
  // bit 3: cut at end
  const flags = cutAtEnd ? 0x08 : 0x00;
  return [ESC, 0x69, 0x4b, flags];
}

function setCopies(copies: number): number[] {
  // ESC i A {n} - Set number of copies
  return [ESC, 0x69, 0x41, Math.min(copies, 255)];
}

function printCommand(): number[] {
  // Print and feed - end of page
  return [0x1a];
}

/**
 * Convert a PNG image to Brother raster format
 * For now, this creates a simple raster representation
 * In production, you'd want to use a proper image processing library
 */
function convertImageToRaster(_imageData: Uint8Array, widthMm: number): number[] {
  const rasterCommands: number[] = [];

  // Get the width in dots (300 DPI)
  const widthDots = MEDIA_WIDTHS[widthMm] || Math.round((widthMm * 300) / 25.4);
  const bytesPerLine = Math.ceil(widthDots / 8);

  // For a real implementation, we would:
  // 1. Decode the PNG image
  // 2. Convert to 1-bit black/white
  // 3. Generate raster lines

  // For now, we'll pass through the image data with raster line commands
  // This is a placeholder - real implementation would process the PNG

  // Each raster line: g {0x00} {0x5a} {length_low} {length_high} {data...}
  // The 'g' command sends a raster line

  // Create a simple test pattern (checkerboard) as placeholder
  // In production, this would be replaced with actual image conversion
  const heightDots = 200; // Placeholder height

  for (let y = 0; y < heightDots; y++) {
    // Raster line command
    rasterCommands.push(0x67, 0x00, 0x5a);

    // Length (2 bytes, little endian)
    rasterCommands.push(bytesPerLine & 0xff);
    rasterCommands.push((bytesPerLine >> 8) & 0xff);

    // Raster data - placeholder pattern
    for (let x = 0; x < bytesPerLine; x++) {
      // This would be actual image data in production
      rasterCommands.push(0x00);
    }
  }

  return rasterCommands;
}

/**
 * Brother Raster Protocol utilities for QL series label printers
 */
export const BrotherRasterProtocol = {
  /**
   * Create a complete print job for Brother QL printers
   * @param imageData PNG image data as Uint8Array
   * @param options Print options
   * @returns Uint8Array containing the complete print job
   */
  createPrintJob(imageData: Uint8Array, options: BrotherPrintOptions): Uint8Array {
    const commands: number[] = [];

    // 1. Invalidate - clear print buffer
    commands.push(...invalidate());

    // 2. Initialize - reset printer
    commands.push(...initialize());

    // 3. Switch to raster mode
    commands.push(...switchToRasterMode());

    // 4. Set media info
    commands.push(...setMediaInfo(options.width, options.height));

    // 5. Set print mode (quality)
    commands.push(...setPrintMode(options.highQuality ?? true));

    // 6. Set cut options
    commands.push(...setCutOptions(options.cutAtEnd ?? true));

    // 7. Set number of copies
    if (options.copies > 1) {
      commands.push(...setCopies(options.copies));
    }

    // 8. Convert image to raster data and add raster lines
    const rasterData = convertImageToRaster(imageData, options.width);
    commands.push(...rasterData);

    // 9. Print command
    commands.push(...printCommand());

    return new Uint8Array(commands);
  },

  /**
   * Parse printer status response
   */
  parseStatus(data: Uint8Array): {
    mediaWidth: number;
    mediaType: string;
    error: string | null;
  } {
    if (data.length < 32) {
      return { mediaWidth: 0, mediaType: "unknown", error: "Invalid status response" };
    }

    // Status byte locations in Brother QL response
    const errorInfo1 = data[8] ?? 0;
    const errorInfo2 = data[9] ?? 0;
    const mediaWidth = data[10] ?? 0;
    const mediaType = data[11] ?? 0;

    let error: string | null = null;

    // Check for errors
    if (errorInfo1 & 0x01) error = "No media";
    else if (errorInfo1 & 0x02) error = "End of media";
    else if (errorInfo1 & 0x04) error = "Tape cutter jam";
    else if (errorInfo1 & 0x10) error = "Transmission error";
    else if (errorInfo2 & 0x01) error = "Cover open";
    else if (errorInfo2 & 0x02) error = "Overheating";

    return {
      mediaWidth,
      mediaType: mediaType === 0x0a ? "continuous" : "die-cut",
      error,
    };
  },
};
