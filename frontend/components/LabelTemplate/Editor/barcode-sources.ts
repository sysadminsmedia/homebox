import { LabelmakerContentType, type LabelmakerBarcodeFormatInfo } from "~~/lib/api/types/data-contracts";

/**
 * Content type compatibility between a barcode symbology and the data it encodes:
 * "any" = URLs and arbitrary text, "alphanumeric" = letters and digits, "numeric" = digits only.
 */
export type SourceContentType = "any" | "alphanumeric" | "numeric";

export interface BarcodeContentSource {
  value: string;
  labelKey: string;
  contentType: SourceContentType;
}

/** Sentinel used by the editor UI for a value typed by the user; the renderer prints it as is. */
export const FIXED_VALUE_SOURCE = "__fixed__";

export const itemBarcodeContentSources: BarcodeContentSource[] = [
  { value: "item_url", labelKey: "item_url", contentType: "any" },
  { value: "asset_id", labelKey: "asset_id", contentType: "alphanumeric" },
  { value: "serial_number", labelKey: "serial_number", contentType: "alphanumeric" },
  { value: "model_number", labelKey: "model_number", contentType: "alphanumeric" },
  { value: "id", labelKey: "item_id", contentType: "any" },
];

export const locationBarcodeContentSources: BarcodeContentSource[] = [
  { value: "location_url", labelKey: "location_url", contentType: "any" },
  { value: "location_name", labelKey: "location_name", contentType: "alphanumeric" },
  { value: "location_path", labelKey: "location_full_path", contentType: "any" },
];

const knownSourceValues = new Set([...itemBarcodeContentSources, ...locationBarcodeContentSources].map(s => s.value));

/** A contentSource that is not one of the named fields is a literal value the renderer encodes verbatim. */
export function isFixedValueSource(source: string | undefined | null): boolean {
  return !knownSourceValues.has(source ?? "");
}

export function isNumericFormat(format: LabelmakerBarcodeFormatInfo | undefined | null): boolean {
  return format?.contentType === LabelmakerContentType.ContentTypeNumeric;
}

export function isSourceCompatible(source: BarcodeContentSource, format: LabelmakerBarcodeFormatInfo): boolean {
  // A URL or a UUID is 36 to 80 characters: hundreds of modules in a linear
  // symbology, far more than a small label has printer dots. Only 2D codes get them.
  if (!format.is2D && source.contentType === "any") return false;
  if (format.contentType === LabelmakerContentType.ContentTypeAny) return true;
  if (format.contentType === LabelmakerContentType.ContentTypeAlphanumeric) {
    return source.contentType === "alphanumeric" || source.contentType === "numeric";
  }
  if (format.contentType === LabelmakerContentType.ContentTypeNumeric) {
    return source.contentType === "numeric";
  }
  return true;
}

export const compatibleItemSources = (format: LabelmakerBarcodeFormatInfo) =>
  itemBarcodeContentSources.filter(s => isSourceCompatible(s, format));
export const compatibleLocationSources = (format: LabelmakerBarcodeFormatInfo) =>
  locationBarcodeContentSources.filter(s => isSourceCompatible(s, format));

/** The named source for a value, or null for a fixed value. */
export function findSource(source: string): BarcodeContentSource | undefined {
  return [...itemBarcodeContentSources, ...locationBarcodeContentSources].find(s => s.value === source);
}

/**
 * Approximate number of modules (the narrowest bar) a linear symbology needs
 * for the content, including start, stop and check symbols. 2D codes grow in
 * both directions and are not estimated.
 */
export function estimateLinearModules(format: string, content: string): number | null {
  const n = content.length;
  if (n === 0) return null;
  switch (format) {
    case "code128": {
      // Start, check and stop take 35 modules; each character 11, or 11 per pair of digits (code set C).
      const symbols = /^\d+$/.test(content) ? Math.ceil(n / 2) : n;
      return 35 + 11 * symbols;
    }
    case "code39":
      // 15 modules per character plus a 1-module gap; start and stop characters included.
      return 16 * (n + 2) - 1;
    case "ean13":
    case "upca":
      return 95;
    case "ean8":
      return 67;
    default:
      return null;
  }
}

/** Representative content for a data source, used to estimate barcode density before printing. */
export function sampleContentForSource(source: string): string {
  const origin = typeof window === "undefined" ? "https://homebox.example" : window.location.origin;
  const uuid = "00000000-0000-0000-0000-000000000000";
  switch (source) {
    case "item_url":
      return `${origin}/item/${uuid}`;
    case "location_url":
      return `${origin}/location/${uuid}`;
    case "id":
      return uuid;
    case "asset_id":
      return "000-001";
    case "serial_number":
    case "model_number":
      return "ABC-12345678";
    case "location_name":
      return "Garage shelf";
    case "location_path":
      return "Garage > Shelf B > Box 3";
    default:
      // Fixed values are known exactly.
      return source;
  }
}

/** Printer dots are the only resolution that matters; below this many per module a 1D code will not scan. */
export const MIN_DOTS_PER_MODULE = 2;

/**
 * Printer dots per module for a linear barcode drawn `widthPx` CSS pixels wide
 * (96 dpi canvas) on a label printed at `dpi`. Null when not estimable.
 */
export function dotsPerModule(format: string, content: string, widthPx: number, dpi: number): number | null {
  const modules = estimateLinearModules(format, content);
  if (!modules || widthPx <= 0 || dpi <= 0) return null;
  const dots = (widthPx * dpi) / 96;
  return dots / modules;
}
