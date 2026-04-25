import { describe, expect, test } from "vitest";
import { isDateOnlyString, parseDateOnly, toDateOnlyString } from "./dateOnly";

describe("isDateOnlyString", () => {
  test("matches YYYY-MM-DD", () => {
    expect(isDateOnlyString("2026-04-17")).toBe(true);
  });
  test("rejects timestamps and junk", () => {
    expect(isDateOnlyString("2026-04-17T22:00:00.000Z")).toBe(false);
    expect(isDateOnlyString("2026/04/17")).toBe(false);
    expect(isDateOnlyString("")).toBe(false);
    expect(isDateOnlyString(null as unknown as string)).toBe(false);
  });
});

describe("toDateOnlyString", () => {
  test("passes YYYY-MM-DD through unchanged", () => {
    expect(toDateOnlyString("2026-04-17")).toBe("2026-04-17");
  });

  test("uses LOCAL components for Date objects (no UTC shift)", () => {
    // April 18 00:00 in the local timezone — the canonical value emitted by
    // VueDatePicker. Used to be ISO-stringified to '2026-04-17T22:00:00Z' in
    // UTC+2 and saved a day early on the backend; must now produce
    // '2026-04-18' regardless of timezone.
    const d = new Date(2026, 3, 18);
    expect(toDateOnlyString(d)).toBe("2026-04-18");
  });

  test("returns empty string for nullish/zero values", () => {
    expect(toDateOnlyString(null)).toBe("");
    expect(toDateOnlyString(undefined)).toBe("");
    expect(toDateOnlyString("")).toBe("");
    const zero = new Date();
    zero.setFullYear(1);
    expect(toDateOnlyString(zero)).toBe("");
  });

  test("normalizes ISO timestamps via local components", () => {
    // Pick a Date from an ISO string and re-extract — the returned date must
    // reflect what the user *sees* locally, not the underlying UTC instant.
    const iso = new Date(2026, 3, 18, 9, 30).toISOString();
    expect(toDateOnlyString(iso)).toBe("2026-04-18");
  });
});

describe("parseDateOnly", () => {
  test("constructs a Date from local components", () => {
    const d = parseDateOnly("2026-04-18");
    expect(d).not.toBeNull();
    expect(d!.getFullYear()).toBe(2026);
    expect(d!.getMonth()).toBe(3);
    expect(d!.getDate()).toBe(18);
    expect(d!.getHours()).toBe(0);
  });

  test("returns null for invalid input", () => {
    expect(parseDateOnly("")).toBeNull();
    expect(parseDateOnly(null)).toBeNull();
    expect(parseDateOnly("2026-04-17T22:00:00Z")).toBeNull();
    expect(parseDateOnly("not a date")).toBeNull();
  });

  test("returns null for impossible calendar dates (no JS rollover)", () => {
    // Day overflow — JS would roll these into the next month.
    expect(parseDateOnly("2026-02-30")).toBeNull();
    expect(parseDateOnly("2026-04-31")).toBeNull();
    // Non-leap Feb 29 — JS would roll into March 1.
    expect(parseDateOnly("2025-02-29")).toBeNull();
    // Month out of range — JS would roll into next/prev year.
    expect(parseDateOnly("2026-13-01")).toBeNull();
    expect(parseDateOnly("2026-00-15")).toBeNull();
    // Day = 0 — JS would land on the last day of the previous month.
    expect(parseDateOnly("2026-04-00")).toBeNull();
  });

  test("accepts genuine leap-year Feb 29", () => {
    const d = parseDateOnly("2024-02-29");
    expect(d).not.toBeNull();
    expect(d!.getFullYear()).toBe(2024);
    expect(d!.getMonth()).toBe(1);
    expect(d!.getDate()).toBe(29);
  });
});

describe("Date object round-trip preserves the calendar day", () => {
  test("toDateOnlyString(parseDateOnly(s)) === s", () => {
    for (const s of ["2020-01-01", "2024-02-29", "2026-04-18", "2026-12-31"]) {
      expect(toDateOnlyString(parseDateOnly(s))).toBe(s);
    }
  });
});
