import { describe, expect, test } from "vitest";
import { resolveWeekStart } from "./weekStart";

describe("resolveWeekStart", () => {
  test("passes explicit day values through unchanged", () => {
    expect(resolveWeekStart(0)).toBe(0);
    expect(resolveWeekStart(1)).toBe(1);
    expect(resolveWeekStart(6)).toBe(6);
  });

  test("'auto' derives the locale's first day (ISO 1-7 → VueDatePicker 0-6)", () => {
    expect(resolveWeekStart("auto", "en-US")).toBe(0); // ISO 7 (Sunday) → 0
    expect(resolveWeekStart("auto", "en-GB")).toBe(1); // ISO 1 (Monday) → 1
    expect(resolveWeekStart("auto", "ar-EG")).toBe(6); // ISO 6 (Saturday) → 6
  });

  test("'auto' falls back to Monday (1) for an invalid locale tag", () => {
    expect(resolveWeekStart("auto", "!!!")).toBe(1);
  });
});
