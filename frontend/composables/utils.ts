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

export function fmtCurrency(value: number | string, currency = "USD", locale = "en-Us"): string {
  if (typeof value === "string") {
    value = parseFloat(value);
  }

  const formatter = new Intl.NumberFormat(locale, {
    style: "currency",
    currency,
    minimumFractionDigits: 2,
  });
  return formatter.format(value);
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
    const match = str.match(/\[(.*)\]\((.*)\)/) as [string, string, string];
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
