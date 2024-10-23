import { useI18n } from "vue-i18n";
import { type UseTimeAgoMessages, type UseTimeAgoUnitNamesDefault } from "@vueuse/core";

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

function ordinalIndicator(num: number) {
  if (num > 3 && num < 21) return "th";
  switch (num % 10) {
    case 1:
      return "st";
    case 2:
      return "nd";
    case 3:
      return "rd";
    default:
      return "th";
  }
}

export function useLocaleTimeAgo(date: Date) {
  const { t } = useI18n();

  const I18N_MESSAGES: UseTimeAgoMessages<UseTimeAgoUnitNamesDefault> = {
    justNow: t("components.global.date_time.just-now"),
    past: (n) => (n.match(/\d/) ? t("components.global.date_time.ago", [n]) : n),
    future: (n) => (n.match(/\d/) ? t("components.global.date_time.in", [n]) : n),
    month: (n, past) =>
      n === 1
        ? past
          ? t("components.global.date_time.last-month")
          : t("components.global.date_time.next-month")
        : `${n} ${t(`components.global.date_time.months`)}`,
    year: (n, past) =>
      n === 1
        ? past
          ? t("components.global.date_time.last-year")
          : t("components.global.date_time.next-year")
        : `${n} ${t(`components.global.date_time.years`)}`,
    day: (n, past) =>
      n === 1
        ? past
          ? t("components.global.date_time.yesterday")
          : t("components.global.date_time.tomorrow")
        : `${n} ${t(`components.global.date_time.days`)}`,
    week: (n, past) =>
      n === 1
        ? past
          ? t("components.global.date_time.last-week")
          : t("components.global.date_time.next-week")
        : `${n} ${t(`components.global.date_time.weeks`)}`,
    hour: (n) => `${n} ${
      n === 1 ? t("components.global.date_time.hour") : t("components.global.date_time.hours")
      }`,
    minute: (n) => `${n} ${
      n === 1 ? t("components.global.date_time.minute") : t("components.global.date_time.minutes")
      }`,
    second: (n) => `${n} ${
      n === 1
        ? t("components.global.date_time.second")
        : t("components.global.date_time.seconds")
    }`,
    invalid: "",
  };

  return useTimeAgo(date, {
    fullDateFormatter: (date: Date) => date.toLocaleDateString(),
    messages: I18N_MESSAGES,
  });
}

export function fmtDate(
  value: string | Date,
  fmt: DateTimeFormat = "human"
): string {
  const months = [
    "January",
    "February",
    "March",
    "April",
    "May",
    "June",
    "July",
    "August",
    "September",
    "October",
    "November",
    "December",
  ];

  const dt = typeof value === "string" ? new Date(value) : value;
  if (!dt) {
    return "";
  }

  if (!validDate(dt)) {
    return "";
  }

  switch (fmt) {
    case "relative":
      return useLocaleTimeAgo(dt).value + useDateFormat(dt, " (YYYY-MM-DD)").value;
    case "long":
      return useDateFormat(dt, "YYYY-MM-DD (dddd)").value;
    case "short":
      return useDateFormat(dt, "YYYY-MM-DD").value;
    case "human":
      // January 1st, 2021
      return `${months[dt.getMonth()]} ${dt.getDate()}${ordinalIndicator(dt.getDate())}, ${dt.getFullYear()}`;
    default:
      return "";
  }
}
