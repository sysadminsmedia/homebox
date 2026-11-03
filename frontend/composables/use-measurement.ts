import { getLocaleCode } from "~~/composables/use-formatters";
import { useViewPreferences } from "~~/composables/use-preferences";

export type MeasurementSystem = "metric" | "imperial";

export const MM_PER_INCH = 25.4;

// Regions that measure everyday lengths in inches rather than millimetres
const IMPERIAL_REGIONS = new Set(["US", "LR", "MM"]);

/**
 * localeMeasurementSystem guesses the measurement system for a locale from its
 * region, e.g. "en-US" -> imperial, "en-GB" -> metric. Locales without a region
 * resolve to their most likely one ("en" -> "en-US").
 */
export function localeMeasurementSystem(locale: string): MeasurementSystem {
  let region: string | undefined;
  try {
    region = new Intl.Locale(locale).maximize().region;
  } catch {
    region = undefined;
  }
  return region && IMPERIAL_REGIONS.has(region) ? "imperial" : "metric";
}

/**
 * measurementLocale returns the locale used to pick a default measurement
 * system. UI locales are often region-less ("en"), which would resolve to the
 * language's most likely region; prefer the browser's regional locale when it
 * is the same language, so en-GB browsers get metric with the English UI.
 */
export function measurementLocale(): string {
  const locale = getLocaleCode();
  if (!locale.includes("-") && import.meta.client) {
    const browser = navigator.language;
    if (browser?.toLowerCase().startsWith(`${locale.toLowerCase()}-`)) {
      return browser;
    }
  }
  return locale;
}

/**
 * useMeasurement converts label dimensions between the stored unit (always
 * millimetres) and the unit the user has chosen to see. A null preference
 * follows the user's locale.
 */
export function useMeasurement() {
  const preferences = useViewPreferences();

  const system = computed<MeasurementSystem>(
    () => preferences.value.measurementSystem ?? localeMeasurementSystem(measurementLocale())
  );
  const isImperial = computed(() => system.value === "imperial");

  /** Unit suffix shown next to values and inputs */
  const unit = computed(() => (isImperial.value ? "in" : "mm"));
  /** Input step that matches the unit's useful precision */
  const step = computed(() => (isImperial.value ? 0.01 : 0.1));
  /** Decimal places kept when converting for display */
  const precision = computed(() => (isImperial.value ? 3 : 1));

  function round(value: number, places: number): number {
    const factor = 10 ** places;
    return Math.round(value * factor) / factor;
  }

  /** Converts millimetres to the display unit */
  function toDisplay(mm: number): number {
    return round(isImperial.value ? mm / MM_PER_INCH : mm, precision.value);
  }

  /** Converts a value in the display unit back to millimetres */
  function fromDisplay(value: number): number {
    return isImperial.value ? round(value * MM_PER_INCH, 2) : value;
  }

  function formatNumber(mm: number): string {
    return new Intl.NumberFormat(getLocaleCode(), {
      maximumFractionDigits: isImperial.value ? 2 : 1,
    }).format(isImperial.value ? mm / MM_PER_INCH : mm);
  }

  /** Formats a single length, e.g. "62 mm" or "2.44 in" */
  function format(mm: number): string {
    return `${formatNumber(mm)} ${unit.value}`;
  }

  /** Formats a width × height pair, e.g. "62 × 29 mm" or "2.44 × 1.14 in" */
  function formatSize(widthMm: number, heightMm: number): string {
    return `${formatNumber(widthMm)} × ${formatNumber(heightMm)} ${unit.value}`;
  }

  return { system, isImperial, unit, step, toDisplay, fromDisplay, format, formatSize };
}
