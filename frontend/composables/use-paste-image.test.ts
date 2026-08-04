import { describe, expect, test } from "vitest";
import { imageFilesFromClipboard, pastedImageName } from "./use-paste-image";

function file(name: string, type: string): File {
  return new File(["x"], name, { type });
}

/** Minimal stand-in for the DataTransfer a real ClipboardEvent carries. */
function clipboard(files: File[]): DataTransfer {
  return { files, items: [] } as unknown as DataTransfer;
}

describe("pastedImageName", () => {
  test("maps mime types to the expected extension", () => {
    expect(pastedImageName("image/png")).toMatch(/^pasted-image-\d{8}-\d{6}\.png$/);
    expect(pastedImageName("image/jpeg")).toMatch(/\.jpg$/);
    expect(pastedImageName("image/gif")).toMatch(/\.gif$/);
    expect(pastedImageName("image/webp")).toMatch(/\.webp$/);
    expect(pastedImageName("image/avif")).toMatch(/\.avif$/);
  });

  test("falls back to the mime subtype for unknown image types", () => {
    expect(pastedImageName("image/bmp")).toMatch(/\.bmp$/);
    expect(pastedImageName("image/svg+xml")).toMatch(/\.svg$/);
  });

  test("suffixes all but the first image so names stay unique", () => {
    expect(pastedImageName("image/png", 0)).not.toMatch(/-\d\.png$/);
    expect(pastedImageName("image/png", 1)).toMatch(/-1\.png$/);
    expect(pastedImageName("image/png", 2)).toMatch(/-2\.png$/);
  });
});

describe("imageFilesFromClipboard", () => {
  test("returns nothing when there is no clipboard payload", () => {
    expect(imageFilesFromClipboard(null)).toEqual([]);
    expect(imageFilesFromClipboard(undefined)).toEqual([]);
    expect(imageFilesFromClipboard(clipboard([]))).toEqual([]);
  });

  test("keeps images and drops everything else", () => {
    const result = imageFilesFromClipboard(
      clipboard([
        file("notes.txt", "text/plain"),
        file("image.png", "image/png"),
        file("manual.pdf", "application/pdf"),
      ])
    );

    expect(result).toHaveLength(1);
    expect(result[0]!.type).toBe("image/png");
  });

  test("renames the generic name browsers hand over", () => {
    const [image] = imageFilesFromClipboard(clipboard([file("image.png", "image/png")]));

    expect(image!.name).not.toBe("image.png");
    expect(image!.name).toMatch(/^pasted-image-\d{8}-\d{6}\.png$/);
  });

  test("gives every image of a multi-image paste a distinct name", () => {
    const names = imageFilesFromClipboard(
      clipboard([file("image.png", "image/png"), file("image.png", "image/png"), file("", "image/jpeg")])
    ).map(image => image.name);

    expect(names).toHaveLength(3);
    expect(new Set(names).size).toBe(3);
    expect(names[2]).toMatch(/-2\.jpg$/);
  });

  test("reads from items when files is empty, as Safari does", () => {
    const image = file("image.png", "image/png");
    const data = {
      files: [],
      items: [
        { kind: "string", getAsFile: () => null },
        { kind: "file", getAsFile: () => image },
      ],
    } as unknown as DataTransfer;

    expect(imageFilesFromClipboard(data)).toHaveLength(1);
  });
});
