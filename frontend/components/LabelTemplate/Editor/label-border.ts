/**
 * Optional frame around the printable area. It is a template-level setting that
 * lives inside canvasData under `labelBorder` (all sizes in mm) so it travels
 * with the design; the server renderer draws it right after the background.
 */
export interface LabelBorder {
  enabled: boolean;
  /** Line width in mm */
  width: number;
  /** Distance from the label edge in mm */
  inset: number;
  color: string;
}

export const LABEL_BORDER_KEY = "labelBorder";

export const DEFAULT_LABEL_BORDER: LabelBorder = {
  enabled: false,
  width: 0.3,
  inset: 1,
  color: "#000000",
};

/** CSS pixels per millimetre on the editor canvas (96 dpi). */
export const MM_TO_PX = 96 / 25.4;

export function readLabelBorder(canvasData: Record<string, unknown> | null | undefined): LabelBorder {
  const raw = canvasData?.[LABEL_BORDER_KEY];
  if (!raw || typeof raw !== "object") return { ...DEFAULT_LABEL_BORDER };
  const o = raw as Record<string, unknown>;
  const num = (v: unknown, fallback: number) => (typeof v === "number" && Number.isFinite(v) && v >= 0 ? v : fallback);
  return {
    enabled: o.enabled === true,
    width: num(o.width, DEFAULT_LABEL_BORDER.width),
    inset: num(o.inset, DEFAULT_LABEL_BORDER.inset),
    color: typeof o.color === "string" && o.color ? o.color : DEFAULT_LABEL_BORDER.color,
  };
}

/** Returns canvas data carrying the border when it is enabled, and without the key otherwise. */
export function withLabelBorder(canvasData: Record<string, unknown>, border: LabelBorder): Record<string, unknown> {
  const { [LABEL_BORDER_KEY]: _previous, ...rest } = canvasData;
  if (!border.enabled) return rest;
  return { ...rest, [LABEL_BORDER_KEY]: { ...border } };
}
