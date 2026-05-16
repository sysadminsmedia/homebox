import type { CurrenciesCurrency } from "~/lib/api/types/data-contracts";

export function validDate(dt: Date | string | null | undefined): boolean {
  if (!dt) {
    return false;
  }

  // If it's a string, try to parse it
  if (typeof dt === "string") {
    if (dt.startsWith("0001")) {
      return false;
    }

    const parsed = new Date(dt);
    if (isNaN(parsed.getTime())) {
      return false;
    }
  }

  // If it's a date, check if it's valid
  if (dt instanceof Date) {
    if (dt.getFullYear() < 1000) {
      return false;
    }
  }

  return true;
}

// Currency cache to store decimal places information
export const currencyDecimalsCache: Record<string, number> = {};

// Promise to track in-flight loading to coalesce concurrent calls
let currencyLoadingPromise: Promise<void> | null = null;

// Safe range for server-provided decimals
const SAFE_MIN_DECIMALS = 0;
const SAFE_MAX_DECIMALS = 4;

// Helper function to clamp decimal places to safe range
function clampDecimals(currency: string, decimals: number): number {
  const truncated = Math.trunc(decimals);
  return Math.max(SAFE_MIN_DECIMALS, Math.min(SAFE_MAX_DECIMALS, truncated));
}

// Type guard to validate currency response shape with strict validation
function isValidCurrencyItem(item: CurrenciesCurrency): boolean {
  if (
    typeof item !== "object" ||
    item === null ||
    typeof item.code !== "string" ||
    item.code.trim() === "" ||
    typeof item.decimals !== "number" ||
    !Number.isFinite(item.decimals)
  ) {
    return false;
  }

  // Truncate decimals to integer and check range
  const truncatedDecimals = Math.trunc(item.decimals);
  return truncatedDecimals >= SAFE_MIN_DECIMALS && truncatedDecimals <= SAFE_MAX_DECIMALS;
}

// Function to load currency decimals from API
function loadCurrencyDecimals(): Promise<void> {
  // Check environment variable to see if remote decimals are disabled
  if (process.env.USE_REMOTE_DECIMALS === "false") {
    return Promise.resolve();
  }

  // Return early if already loaded
  if (Object.keys(currencyDecimalsCache).length > 0) {
    return Promise.resolve();
  }

  // Coalesce concurrent calls - return existing promise if loading
  if (currencyLoadingPromise) {
    return currencyLoadingPromise;
  }

  // Create new loading promise
  currencyLoadingPromise = (async () => {
    try {
      const api = useUserApi();
      const { data, error } = await api.group.currencies();

      if (!error && data) {
        // Validate that data is an array
        if (!Array.isArray(data)) {
          // Log generic message without server details
          console.warn("Currency API returned invalid data format");
          return;
        }

        // Process and validate each currency item
        for (const currency of data) {
          // Strict validation: only process items that pass all checks
          if (!isValidCurrencyItem(currency)) {
            // Skip invalid items without caching - no clamping for out-of-range values
            continue;
          }

          // Only cache strictly validated items with truncated and clamped decimals
          const code = currency.code.trim().toUpperCase();
          const truncatedDecimals = Math.trunc(currency.decimals);
          const clampedDecimals = Math.max(SAFE_MIN_DECIMALS, Math.min(SAFE_MAX_DECIMALS, truncatedDecimals));
          currencyDecimalsCache[code] = clampedDecimals;
        }
      } else if (error) {
        // Generic error logging without exposing server error details
        console.warn("Currency API request failed, using default formatting");
      }
    } catch (e) {
      // Generic error without sensitive details - no raw error logging
      console.warn("Currency data loading failed, using default formatting");
    } finally {
      // Clear loading promise when done (success or failure)
      currencyLoadingPromise = null;
    }
  })();

  return currencyLoadingPromise;
}

export function fmtCurrency(value: number | string, currency = "USD", locale = "en-Us"): string {
  if (typeof value === "string") {
    value = parseFloat(value);
  }

  // Normalize and validate currency code
  const normalizedCurrency = String(currency).toUpperCase();
  const safeCurrency = /^[A-Z]{3}$/.test(normalizedCurrency) ? normalizedCurrency : "USD";
  // Derive fraction digits using the same clamp helper
  const fractionDigits = clampDecimals(safeCurrency, currencyDecimalsCache[safeCurrency] ?? 2);

  const formatter = new Intl.NumberFormat(locale, {
    style: "currency",
    currency: safeCurrency,
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits,
  });
  return formatter.format(value);
}

export async function fmtCurrencyAsync(value: number | string, currency = "USD", locale = "en-Us"): Promise<string> {
  await loadCurrencyDecimals();
  return fmtCurrency(value, currency, locale);
}

// Matches an input that already declares a URL scheme like "https://" or
// "file://". When this matches we pass the input to the URL parser as-is
// and accept only http(s); anything else is rejected without trying to
// coerce it.
const EXPLICIT_SCHEME_RE = /^[a-z][a-z0-9+.-]*:\/\//i;

// Parses a scanner payload into a URL. Accepts full URLs as well as
// protocol-less host+path payloads such as "example.com/a/1" or
// "localhost:3000/a/1", which let pre-printed asset stickers carry less
// data. Returns null for inputs that aren't URL-shaped (e.g. EAN/UPC
// barcode digits, file://, mailto:) so callers can fall back to other
// handling.
export function parseScanResult(rawValue: string): URL | null {
  if (EXPLICIT_SCHEME_RE.test(rawValue)) {
    try {
      const url = new URL(rawValue);
      if ((url.protocol === "http:" || url.protocol === "https:") && url.host) {
        return url;
      }
    } catch {
      // Malformed URL.
    }
    return null;
  }

  // No "scheme://", so treat as a protocol-less host+path payload. Require a
  // "/" so plain barcode payloads (EAN/UPC digits) and arbitrary text return
  // null and the caller can fall back to the barcode handler.
  if (rawValue.startsWith("/") || !rawValue.includes("/")) {
    return null;
  }
  // Use the current page's protocol so http-only deployments still pass
  // their own origin checks.
  const protocol = globalThis.location?.protocol ?? "https:";
  try {
    const url = new URL(`${protocol}//${rawValue}`);
    if ((url.protocol !== "http:" && url.protocol !== "https:") || !url.host) {
      return null;
    }
    return url;
  } catch {
    return null;
  }
}

export type MaybeUrlResult = {
  isUrl: boolean;
  url: string;
  text: string;
};

export function maybeUrl(str: string): MaybeUrlResult {
  const result: MaybeUrlResult = {
    isUrl: str.startsWith("http://") || str.startsWith("https://"),
    url: "",
    text: "",
  };

  if (!result.isUrl && !str.startsWith("[")) {
    return result;
  }

  if (str.startsWith("[")) {
    const match = str.match(/\[(.*)\]\((.*)\)/);
    if (match && match.length === 3) {
      result.isUrl = true;
      result.text = match[1]!;
      result.url = match[2]!;
    }
  } else {
    result.url = str;
    result.text = str;
  }

  return result;
}
