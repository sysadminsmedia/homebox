// Date-only helpers for fields that represent a calendar date with no
// time-of-day or timezone semantics (e.g. purchaseDate, scheduledDate).
//
// The frontend must keep these fields as YYYY-MM-DD strings end-to-end. A JS
// Date object always carries a UTC instant, and JSON.stringify uses UTC, so
// passing Dates across the API boundary causes the day to shift for every
// user not on UTC. Use these helpers at the picker/display boundary instead.

const DATE_ONLY_RE = /^\d{4}-\d{2}-\d{2}$/;

export function isDateOnlyString(value: unknown): value is string {
  return typeof value === "string" && DATE_ONLY_RE.test(value);
}

// toDateOnlyString returns YYYY-MM-DD using LOCAL year/month/day components.
// Anything else (zero/empty/null/invalid) collapses to "".
export function toDateOnlyString(value: Date | string | null | undefined): string {
  if (value == null || value === "") return "";

  if (typeof value === "string") {
    if (isDateOnlyString(value)) return value;
    // Tolerate ISO timestamps coming from older code paths — interpret in
    // local time so the visible day is preserved.
    const d = new Date(value);
    if (isNaN(d.getTime()) || d.getFullYear() < 1000) return "";
    return formatLocal(d);
  }

  if (value instanceof Date) {
    if (isNaN(value.getTime()) || value.getFullYear() < 1000) return "";
    return formatLocal(value);
  }

  return "";
}

// parseDateOnly turns YYYY-MM-DD into a Date constructed from LOCAL
// components, so the resulting Date's local getters return the same Y/M/D.
// Anything else returns null.
export function parseDateOnly(value: string | null | undefined): Date | null {
  if (!value || !isDateOnlyString(value)) return null;
  const [y, m, d] = value.split("-").map(n => parseInt(n, 10)) as [number, number, number];
  return new Date(y, m - 1, d);
}

function formatLocal(d: Date): string {
  const y = d.getFullYear().toString().padStart(4, "0");
  const m = (d.getMonth() + 1).toString().padStart(2, "0");
  const day = d.getDate().toString().padStart(2, "0");
  return `${y}-${m}-${day}`;
}
