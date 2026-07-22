// WeekStart is the first day shown in date pickers: "auto" (derive from the
// user's locale) or an explicit day where 0 = Sunday … 6 = Saturday, matching
// @vuepic/vue-datepicker's `week-start` prop.
export type WeekStart = "auto" | 0 | 1 | 2 | 3 | 4 | 5 | 6;

// resolveWeekStart converts a stored firstDayOfWeek preference into a concrete
// VueDatePicker `week-start` (0 = Sunday … 6 = Saturday). "auto" derives it from
// the supplied locale (or the browser locale) via Intl.Locale week info, falling
// back to Monday when the engine doesn't expose it.
//
// Kept as a side-effect-free module (no Nuxt auto-imports) so it stays unit
// testable in isolation from the stateful use-preferences composable.
export function resolveWeekStart(pref: WeekStart, locale?: string | null): number {
  if (pref !== "auto") {
    return pref;
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
    if (typeof firstDay === "number") {
      return firstDay === 7 ? 0 : firstDay; // map ISO → VueDatePicker (0 = Sunday)
    }
  } catch {
    // unknown/invalid locale tag — fall through to the default
  }
  return 1; // Monday
}
