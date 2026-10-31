// Pure logic for the label generator page (frontend/pages/reports/label-generator.vue).
// Kept free of Vue/i18n/toast so it can be unit tested in isolation.

export type Measure = "in" | "cm" | "mm";

export type LabelMode = "sheet" | "maker" | "custom";

export type LabelOptionInput = {
  measure: string;
  page: {
    height: number;
    width: number;
    pageTopPadding: number;
    pageBottomPadding: number;
    pageLeftPadding: number;
    pageRightPadding: number;
  };
  cardHeight: number;
  cardWidth: number;
};

export type GridData = {
  measure: Measure;
  cols: number;
  rows: number;
  gapY: number;
  gapX: number;
  card: {
    width: number;
    height: number;
  };
  page: {
    width: number;
    height: number;
    pt: number;
    pb: number;
    pl: number;
    pr: number;
  };
};

export type GridResult = { ok: true; data: GridData } | { ok: false; error: "page_too_small_card" };

export type LabelMakerInput = {
  measure: string;
  labelWidth: number;
  labelHeight: number;
  labelsPerRow: number;
  labelGap: number;
};

export type MakerPageSize = {
  measure: Measure;
  width: number;
  height: number;
};

export type LabelPreset = {
  measure: Measure;
  cardHeight: number;
  cardWidth: number;
  pageWidth: number;
  pageHeight: number;
  pageTopPadding: number;
  pageBottomPadding: number;
  pageLeftPadding: number;
  pageRightPadding: number;
  labelsPerRow: number;
  labelGap: number;
};

export const DEFAULT_MEASURE: Measure = "in";

// Avery 5260 sheet of labels (the historical default).
export const SHEET_PRESET: LabelPreset = {
  measure: "in",
  cardHeight: 1,
  cardWidth: 2.63,
  pageWidth: 8.5,
  pageHeight: 11,
  pageTopPadding: 0.52,
  pageBottomPadding: 0.42,
  pageLeftPadding: 0.25,
  pageRightPadding: 0.1,
  labelsPerRow: 1,
  labelGap: 0,
};

// Brother QL DK-2205 62mm continuous tape: 2.4" tape width, default 1" length. Dimensions are editable in the UI.
export const MAKER_PRESET: LabelPreset = {
  measure: "in",
  cardHeight: 1,
  cardWidth: 2.4,
  pageWidth: 2.4,
  pageHeight: 1,
  pageTopPadding: 0,
  pageBottomPadding: 0,
  pageLeftPadding: 0,
  pageRightPadding: 0,
  labelsPerRow: 1,
  labelGap: 0,
};

export function normalizeMeasure(measure: string): Measure {
  return /^(in|cm|mm)$/.test(measure) ? (measure as Measure) : DEFAULT_MEASURE;
}

// Returns the geometry preset for a mode, or null when the mode owns no preset (custom).
export function presetFor(mode: LabelMode): LabelPreset | null {
  if (mode === "sheet") {
    return SHEET_PRESET;
  }
  if (mode === "maker") {
    return MAKER_PRESET;
  }
  return null;
}

export function fmtAssetID(aid: number | string): string {
  let aidStr = aid.toString().padStart(6, "0");
  aidStr = aidStr.slice(0, 3) + "-" + aidStr.slice(3);
  return aidStr;
}

// Lays a sheet out into a grid of cards based on the available page area.
export function calculateGridData(input: LabelOptionInput): GridResult {
  const { page, cardHeight, cardWidth } = input;
  const measure = normalizeMeasure(input.measure);

  const availablePageWidth = page.width - page.pageLeftPadding - page.pageRightPadding;
  const availablePageHeight = page.height - page.pageTopPadding - page.pageBottomPadding;

  if (availablePageWidth < cardWidth || availablePageHeight < cardHeight) {
    return { ok: false, error: "page_too_small_card" };
  }

  const cols = Math.floor(availablePageWidth / cardWidth);
  const rows = Math.floor(availablePageHeight / cardHeight);
  // Guard single-column / single-row layouts so the gap stays 0 instead of NaN/Infinity.
  const gapX = cols > 1 ? (availablePageWidth - cols * cardWidth) / (cols - 1) : 0;
  const gapY = rows > 1 ? (availablePageHeight - rows * cardHeight) / (rows - 1) : 0;

  return {
    ok: true,
    data: {
      measure,
      cols,
      rows,
      gapX,
      gapY,
      card: {
        width: cardWidth,
        height: cardHeight,
      },
      page: {
        width: page.width,
        height: page.height,
        pt: page.pageTopPadding,
        pb: page.pageBottomPadding,
        pl: page.pageLeftPadding,
        pr: page.pageRightPadding,
      },
    },
  };
}

// Width (and height) of a single label-maker segment: one row of N labels.
export function makerPageSize(input: LabelMakerInput): MakerPageSize {
  const perRow = Math.max(1, Math.floor(input.labelsPerRow));
  return {
    measure: normalizeMeasure(input.measure),
    width: perRow * input.labelWidth + (perRow - 1) * input.labelGap,
    height: input.labelHeight,
  };
}

// A label maker prints one row of labels per tape segment: a single-row, zero-padding grid.
export function calculateMakerGrid(input: LabelMakerInput): GridData {
  const cols = Math.max(1, Math.floor(input.labelsPerRow));
  const size = makerPageSize(input);

  return {
    measure: size.measure,
    cols,
    rows: 1,
    gapX: cols > 1 ? input.labelGap : 0,
    gapY: 0,
    card: {
      width: input.labelWidth,
      height: input.labelHeight,
    },
    page: {
      width: size.width,
      height: size.height,
      pt: 0,
      pb: 0,
      pl: 0,
      pr: 0,
    },
  };
}

// Rotation applied to maker labels when printing, to match how the printer feeds the tape.
export type PrintRotation = 0 | 90 | 180 | 270;

// CSS @page rule. Sheet/custom keep the historical behavior (no rule, user sets printer margins).
// Label maker sizes each printed page to one tape segment so labels feed correctly.
// 90/270 swap the page to portrait since the rotated label is taller than wide.
export function buildPageCss(mode: LabelMode, size: MakerPageSize, rotation: PrintRotation = 0): string {
  if (mode !== "maker") {
    return "";
  }
  const swap = rotation === 90 || rotation === 270;
  const w = swap ? size.height : size.width;
  const h = swap ? size.width : size.height;
  return `@page { size: ${w}${size.measure} ${h}${size.measure}; margin: 0; }`;
}

// Print-only transform that rotates each maker label to match the chosen rotation, re-centering it on the page.
export function buildRotateCss(mode: LabelMode, size: MakerPageSize, rotation: PrintRotation): string {
  if (mode !== "maker" || rotation === 0) {
    return "";
  }
  if (rotation === 180) {
    return `@media print { .maker-label { transform: rotate(180deg); transform-origin: center center; } }`;
  }
  const m = size.measure;
  const shift = (size.width - size.height) / 2;
  return `@media print { .maker-label { width: ${size.width}${m}; height: ${size.height}${m}; transform: translate(${-shift}${m}, ${shift}${m}) rotate(${rotation}deg); transform-origin: center center; } }`;
}
