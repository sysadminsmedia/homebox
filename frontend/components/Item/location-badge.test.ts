import { describe, expect, test } from "vitest";
import { isScopedLocationItem, shouldShowParentLocationBadge } from "./location-badge";

describe("shouldShowParentLocationBadge", () => {
  const parent = { id: "parent-in-another-room", name: "Remote container" };

  test("shows the original parent badge for ordinary item lists", () => {
    expect(shouldShowParentLocationBadge({ parent })).toBe(true);
  });

  test("does not show a parent's location when the scoped path is visible", () => {
    expect(shouldShowParentLocationBadge({ parent, locationPath: "Attic / Shelf" })).toBe(false);
  });

  test("does not show a parent's location when the scoped path is hidden", () => {
    expect(shouldShowParentLocationBadge({ parent, locationPath: undefined })).toBe(false);
  });

  test("does not show a location badge when there is no parent", () => {
    expect(shouldShowParentLocationBadge({})).toBe(false);
  });

  test("distinguishes scoped rows with hidden paths from unscoped rows", () => {
    expect(isScopedLocationItem({ locationPath: undefined })).toBe(true);
    expect(isScopedLocationItem({ locationPath: "Attic / Shelf" })).toBe(true);
    expect(isScopedLocationItem({})).toBe(false);
  });
});
