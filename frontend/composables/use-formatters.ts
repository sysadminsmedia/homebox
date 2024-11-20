import { useI18n } from "vue-i18n";
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

  return (value: number | string) => fmtCurrency(value, cache.currency);
}

export type DateTimeFormat = "relative" | "long" | "short" | "human";
export type DateTimeType = "date" | "time" | "datetime";

function getLocale() {
  const t = useI18n();
  const localeCode = (t?.locale?.value as string) ?? "en-US";
  const lang = localeCode.length > 1 ? localeCode.substring(0, 2) : localeCode;
  const region = localeCode.length > 2 ? localeCode.substring(3) : "";
  return Locales[(lang + region) as keyof typeof Locales] ?? Locales[lang as keyof typeof Locales] ?? Locales.enUS;
}

export function useLocaleTimeAgo(date: Date) {
  return formatDistance(date, new Date(), { locale: getLocale() });
}

export function fmtDate(value: string | Date, fmt: DateTimeFormat = "human", type: DateTimeType = "date"): string {
  const dt = typeof value === "string" ? new Date(value) : value;

  if (!dt || !validDate(dt)) {
    return "";
  }

  if (fmt === "relative") {
    return useLocaleTimeAgo(dt);
  }

  if (type === "time") {
    return format(dt, "p", { locale: getLocale() });
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

  return format(dt, formatStr, { locale: getLocale() });
}
