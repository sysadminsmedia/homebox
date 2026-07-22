// WeekStart is the first day shown in date pickers: "auto" (derive from the
// user's locale) or an explicit day where 0 = Sunday … 6 = Saturday, matching
// @vuepic/vue-datepicker's `week-start` prop.
export type WeekStart = "auto" | 0 | 1 | 2 | 3 | 4 | 5 | 6;

// isWeekDay is a runtime guard for a VueDatePicker week-start (0 = Sunday …
// 6 = Saturday). The preference is read back from schemaless storage (the server
// settings blob and localStorage), so the WeekStart type gives no runtime
// protection — a malformed value must not leak into the picker.
function isWeekDay(n: unknown): n is 0 | 1 | 2 | 3 | 4 | 5 | 6 {
  return typeof n === "number" && Number.isInteger(n) && n >= 0 && n <= 6;
}

// resolveWeekStart converts a stored firstDayOfWeek preference into a concrete
// VueDatePicker `week-start` (0 = Sunday … 6 = Saturday). "auto" derives it from
// the supplied locale (or the browser locale) via Intl.Locale week info. Any
// value it can't validate — a malformed preference, an unknown locale, or an
// out-of-range firstDay — falls back to Monday.
//
// Kept as a side-effect-free module (no Nuxt auto-imports) so it stays unit
// testable in isolation from the stateful use-preferences composable.
export function resolveWeekStart(pref: WeekStart, locale?: string | null): number {
  if (pref !== "auto") {
    return isWeekDay(pref) ? pref : 1; // clamp malformed persisted values
  }
  try {
    const tag = locale || (typeof navigator !== "undefined" ? navigator.language : "en-US");
    const loc = new Intl.Locale(tag);
    // getWeekInfo() is the current spec API; older engines (including current
    // Node/V8) expose a `weekInfo` getter instead — support both.
    const info =
      (loc as unknown as { getWeekInfo?: () => { firstDay?: number } }).getWeekInfo?.() ??
      (loc as unknown as { weekInfo?: { firstDay?: number } }).weekInfo;
    const firstDay = info?.firstDay; // 1 = Monday … 7 = Sunday (ISO-8601)
    // Map ISO (1-7) → VueDatePicker (0 = Sunday); ignore missing/out-of-range.
    if (typeof firstDay === "number" && firstDay >= 1 && firstDay <= 7) {
      return firstDay === 7 ? 0 : firstDay;
    }
  } catch {
    // unknown/invalid locale tag — fall through to the default
  }
  return 1; // Monday
}
