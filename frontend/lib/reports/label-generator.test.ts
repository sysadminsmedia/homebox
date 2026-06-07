import { describe, expect, test } from "vitest";
import {
  buildPageCss,
  buildRotateCss,
  calculateGridData,
  calculateMakerGrid,
  fmtAssetID,
  makerPageSize,
  MAKER_PRESET,
  presetFor,
  SHEET_PRESET,
  type LabelMakerInput,
  type LabelOptionInput,
} from "./label-generator";

const averyInput: LabelOptionInput = {
  measure: "in",
  page: {
    width: 8.5,
    height: 11,
    pageTopPadding: 0.52,
    pageBottomPadding: 0.42,
    pageLeftPadding: 0.25,
    pageRightPadding: 0.1,
  },
  cardWidth: 2.63,
  cardHeight: 1,
};

const makerInput: LabelMakerInput = {
  measure: "mm",
  labelWidth: 90,
  labelHeight: 62,
  labelsPerRow: 1,
  labelGap: 0,
};

describe("fmtAssetID", () => {
  test("pads and hyphenates a small number", () => {
    expect(fmtAssetID(1)).toBe("000-001");
  });

  test("formats a full six digit value", () => {
    expect(fmtAssetID(123456)).toBe("123-456");
  });

  test("does not truncate values longer than six digits", () => {
    expect(fmtAssetID(1234567)).toBe("123-4567");
  });

  test("accepts string input", () => {
    expect(fmtAssetID("42")).toBe("000-042");
  });
});

describe("calculateGridData", () => {
  test("lays out the Avery 5260 sheet", () => {
    const res = calculateGridData(averyInput);
    expect(res.ok).toBe(true);
    if (!res.ok) return;
    expect(res.data.cols).toBe(3);
    expect(res.data.rows).toBe(10);
    expect(res.data.measure).toBe("in");
    expect(res.data.card).toEqual({ width: 2.63, height: 1 });
  });

  test("errors when the page is too small for the card", () => {
    const res = calculateGridData({ ...averyInput, cardWidth: 100 });
    expect(res).toEqual({ ok: false, error: "page_too_small_card" });
  });

  test("single column yields gapX of 0, not NaN", () => {
    const res = calculateGridData({ ...averyInput, cardWidth: 7 });
    expect(res.ok).toBe(true);
    if (!res.ok) return;
    expect(res.data.cols).toBe(1);
    expect(res.data.gapX).toBe(0);
  });

  test("single row yields gapY of 0, not NaN", () => {
    const res = calculateGridData({ ...averyInput, cardHeight: 9 });
    expect(res.ok).toBe(true);
    if (!res.ok) return;
    expect(res.data.rows).toBe(1);
    expect(res.data.gapY).toBe(0);
  });

  test("falls back to inches for an invalid measure", () => {
    const res = calculateGridData({ ...averyInput, measure: "furlong" });
    expect(res.ok).toBe(true);
    if (!res.ok) return;
    expect(res.data.measure).toBe("in");
  });
});

describe("makerPageSize", () => {
  test("single label width equals the label width", () => {
    expect(makerPageSize(makerInput)).toEqual({ measure: "mm", width: 90, height: 62 });
  });

  test("row of three includes the gaps between labels", () => {
    const size = makerPageSize({ ...makerInput, labelsPerRow: 3, labelGap: 2 });
    expect(size.width).toBe(3 * 90 + 2 * 2);
    expect(size.height).toBe(62);
  });

  test("normalizes the measure", () => {
    expect(makerPageSize({ ...makerInput, measure: "bogus" }).measure).toBe("in");
  });
});

describe("calculateMakerGrid", () => {
  test("single label is a one-by-one grid with no gaps", () => {
    const grid = calculateMakerGrid(makerInput);
    expect(grid.cols).toBe(1);
    expect(grid.rows).toBe(1);
    expect(grid.gapX).toBe(0);
    expect(grid.gapY).toBe(0);
    expect(grid.page).toEqual({ width: 90, height: 62, pt: 0, pb: 0, pl: 0, pr: 0 });
  });

  test("row of three uses the label gap for gapX", () => {
    const grid = calculateMakerGrid({ ...makerInput, labelsPerRow: 3, labelGap: 2 });
    expect(grid.cols).toBe(3);
    expect(grid.rows).toBe(1);
    expect(grid.gapX).toBe(2);
    expect(grid.page.width).toBe(3 * 90 + 2 * 2);
  });
});

describe("buildPageCss", () => {
  const size = { measure: "mm" as const, width: 90, height: 62 };

  test("maker mode emits a sized, margin-free page rule", () => {
    expect(buildPageCss("maker", size)).toBe("@page { size: 90mm 62mm; margin: 0; }");
  });

  test("sheet mode emits no rule", () => {
    expect(buildPageCss("sheet", size)).toBe("");
  });

  test("custom mode emits no rule", () => {
    expect(buildPageCss("custom", size)).toBe("");
  });

  test("180 rotation keeps the page dimensions", () => {
    expect(buildPageCss("maker", size, 180)).toBe("@page { size: 90mm 62mm; margin: 0; }");
  });

  test("90 rotation swaps the page dimensions", () => {
    expect(buildPageCss("maker", size, 90)).toBe("@page { size: 62mm 90mm; margin: 0; }");
  });

  test("270 rotation swaps the page dimensions", () => {
    expect(buildPageCss("maker", size, 270)).toBe("@page { size: 62mm 90mm; margin: 0; }");
  });
});

describe("buildRotateCss", () => {
  const size = { measure: "mm" as const, width: 90, height: 62 };

  test("no rotation emits no rule", () => {
    expect(buildRotateCss("maker", size, 0)).toBe("");
  });

  test("sheet mode emits no rule", () => {
    expect(buildRotateCss("sheet", size, 90)).toBe("");
  });

  test("180 rotation flips in place without resizing", () => {
    expect(buildRotateCss("maker", size, 180)).toBe(
      "@media print { .maker-label { transform: rotate(180deg); transform-origin: center center; } }"
    );
  });

  test("90 rotation sizes and re-centers the label onto the swapped page", () => {
    expect(buildRotateCss("maker", size, 90)).toBe(
      "@media print { .maker-label { width: 90mm; height: 62mm; transform: translate(-14mm, 14mm) rotate(90deg); transform-origin: center center; } }"
    );
  });

  test("270 rotation sizes and re-centers the label onto the swapped page", () => {
    // Re-centering moves the box center onto the swapped page center, which is independent of rotation direction,
    // so 270 shares the 90 translate offsets.
    expect(buildRotateCss("maker", size, 270)).toBe(
      "@media print { .maker-label { width: 90mm; height: 62mm; transform: translate(-14mm, 14mm) rotate(270deg); transform-origin: center center; } }"
    );
  });
});

describe("presetFor", () => {
  test("returns the sheet preset", () => {
    expect(presetFor("sheet")).toBe(SHEET_PRESET);
  });

  test("returns the maker preset", () => {
    expect(presetFor("maker")).toBe(MAKER_PRESET);
  });

  test("returns null for custom", () => {
    expect(presetFor("custom")).toBeNull();
  });
});
