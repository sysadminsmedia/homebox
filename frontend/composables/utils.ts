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

// Function to load currency decimals from API
async function loadCurrencyDecimals(): Promise<void> {
  if (Object.keys(currencyDecimalsCache).length > 0) {
    return; // Already loaded
  }

  try {
    const api = useUserApi();
    const { data, error } = await api.group.currencies();

    if (!error && data) {
      for (const currency of data) {
        currencyDecimalsCache[currency.code] = currency.decimals;
      }
    }
  } catch (e) {
    // Fallback to default behavior if API fails
    console.warn("Failed to load currency decimals, using defaults");
  }
}

export function fmtCurrency(value: number | string, currency = "USD", locale = "en-Us"): string {
  if (typeof value === "string") {
    value = parseFloat(value);
  }

  // Get decimal places from cache or default to 2
  const fractionDigits = currencyDecimalsCache[currency] ?? 2;

  const formatter = new Intl.NumberFormat(locale, {
    style: "currency",
    currency,
    minimumFractionDigits: fractionDigits,
  });
  return formatter.format(value);
}

export async function fmtCurrencyAsync(value: number | string, currency = "USD", locale = "en-Us"): Promise<string> {
  await loadCurrencyDecimals();
  return fmtCurrency(value, currency, locale);
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
      result.text = match[1];
      result.url = match[2];
    }
  } else {
    result.url = str;
    result.text = str;
  }

  return result;
}
