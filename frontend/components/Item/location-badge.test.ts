import { describe, expect, test } from "vitest";
import { isScopedLocationItem, shouldShowParentLocationBadge, withScopedLocationPath } from "./location-badge";

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

describe("withScopedLocationPath", () => {
  const item = {
    parent: { id: "parent-in-basement", name: "Basement box" },
    locationPath: "Attic / Shelf",
  };

  test("preserves a visible physical path and scoped marker", () => {
    const projected = withScopedLocationPath(item, true);
    expect(projected.locationPath).toBe("Attic / Shelf");
    expect(isScopedLocationItem(projected)).toBe(true);
    expect(shouldShowParentLocationBadge(projected)).toBe(false);
  });

  test("does not leak a hidden path or fall back to another room's parent", () => {
    const projected = withScopedLocationPath(item, false);
    expect(projected.locationPath).toBeUndefined();
    expect(Object.hasOwn(projected, "locationPath")).toBe(true);
    expect(shouldShowParentLocationBadge(projected)).toBe(false);
    expect(item.locationPath).toBe("Attic / Shelf");
  });

  test("marks a scoped result even when the API omitted an empty path", () => {
    const projected = withScopedLocationPath({ parent: item.parent }, true);
    expect(Object.hasOwn(projected, "locationPath")).toBe(true);
    expect(shouldShowParentLocationBadge(projected)).toBe(false);
  });
});
