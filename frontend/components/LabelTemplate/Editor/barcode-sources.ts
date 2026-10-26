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
