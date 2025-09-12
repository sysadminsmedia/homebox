import { format, formatDistance } from "date-fns";
/* eslint import/namespace: ['error', { allowComputed: true }] */
import * as Locales from "date-fns/locale";

const cache = {
  currency: "",
};

export function resetCurrency() {
  cache.currency = "";
}

export async function useFormatCurrency() {
  if (cache.currency === "") {
    const client = useUserApi();

    const { data: group } = await client.group.get();

    if (group) {
      cache.currency = group.currency;
    }
  }

  return (value: number | string) => fmtCurrency(value, cache.currency, getLocaleCode());
}

export type DateTimeFormat = "relative" | "long" | "short" | "human";
export type DateTimeType = "date" | "time" | "datetime";

export function getLocaleCode() {
  const { $i18nGlobal } = useNuxtApp();
  const preferences = useViewPreferences();
  // TODO: make reactive
  if (preferences.value.overrideFormatLocale) {
    return preferences.value.overrideFormatLocale;
  }
  return ($i18nGlobal?.locale?.value as string) ?? "en-US";
}

function getLocaleForDate() {
  const localeCode = getLocaleCode();
  const lang = localeCode.length > 1 ? localeCode.substring(0, 2) : localeCode;
  const region = localeCode.length > 2 ? localeCode.substring(3) : "";
  return Locales[(lang + region) as keyof typeof Locales] ?? Locales[lang as keyof typeof Locales] ?? Locales.enUS;
}

export function fmtDate(
  value: string | Date | number,
  fmt: DateTimeFormat = "human",
  type: DateTimeType = "date"
): string {
  let dt: Date | null = null;
  
  if (typeof value === "string" || typeof value === "number") {
    dt = new Date(value);
  } else {
    dt = value;
  }

  if (!dt || !validDate(dt)) {
    return "";
  }

  // Always use UTC dates internally
  const utcDate = new Date(Date.UTC(
    dt.getUTCFullYear(),
    dt.getUTCMonth(),
    dt.getUTCDate(),
    dt.getUTCHours(),
    dt.getUTCMinutes(),
    dt.getUTCSeconds()
  ));

  const localeOptions = { locale: getLocaleForDate() };

  if (fmt === "relative") {
    return `${formatDistance(utcDate, new Date(), { ...localeOptions, addSuffix: true })} (${fmtDate(utcDate, "short", "date")})`;
  }

  if (type === "time") {
    return format(utcDate, "p", localeOptions);
  }

  let formatStr = "";

  switch (fmt) {
    case "human":
      formatStr = "PPP";
      break;
    case "long":
      formatStr = "PP";
      break;
    case "short":
      formatStr = "P";
      break;
    default:
      return "";
  }
  if (type === "datetime") {
    formatStr += "p";
  }

  return format(utcDate, formatStr, localeOptions);
}
