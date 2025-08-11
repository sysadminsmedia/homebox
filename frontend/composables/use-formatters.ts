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
  const dt = typeof value === "string" || typeof value === "number" ? new Date(value) : value;

  if (!dt || !validDate(dt)) {
    return "";
  }

  const localeOptions = { locale: getLocaleForDate() };

  if (fmt === "relative") {
    return `${formatDistance(dt, new Date(), { ...localeOptions, addSuffix: true })} (${fmtDate(dt, "short", "date")})`;
  }

  if (type === "time") {
    return format(dt, "p", localeOptions);
  }

  let formatStr = "";

  // Get runtime config for custom date formats
  const config = useRuntimeConfig();

  switch (fmt) {
    case "human":
      formatStr = (config.public.hboxDateFormatHuman as string) || "PPP";
      break;
    case "long":
      formatStr = (config.public.hboxDateFormatLong as string) || "PP";
      break;
    case "short":
      formatStr = (config.public.hboxDateFormatShort as string) || "P";
      break;
    default:
      return "";
  }
  if (type === "datetime") {
    formatStr += "p";
  }

  return format(dt, formatStr, localeOptions);
}
