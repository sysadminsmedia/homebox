import { describe, expect, test } from "vitest";
import { clampPageToTotal, isCurrentLocationPage } from "./pagination";

describe("clampPageToTotal", () => {
  test.each([
    [2, 48, 24, 2],
    [2, 24, 24, 1],
    [4, 50, 24, 3],
    [2, 0, 24, 1],
    [1, 1, 24, 1],
    [0, 100, 24, 1],
    [3, 1, 0, 1],
    [3, 0, 0, 1],
  ])("clamps page %i with total %i and size %i to %i", (page, total, size, expected) => {
    expect(clampPageToTotal(page, total, size)).toBe(expected);
  });
});

describe("isCurrentLocationPage", () => {
  test("shows only the successful current page", () => {
    expect(isCurrentLocationPage("success", "page-2", "page-2")).toBe(true);
  });

  test("hides previous-page rows even if cached status says success", () => {
    expect(isCurrentLocationPage("success", "page-2", "page-1")).toBe(false);
  });

  test("hides rows and counts while loading or refreshing", () => {
    expect(isCurrentLocationPage("pending", "page-2", "page-2")).toBe(false);
    expect(isCurrentLocationPage("idle", "page-2", "page-2")).toBe(false);
  });

  test("hides rows after errors or without a keyed result", () => {
    expect(isCurrentLocationPage("error", "page-2", "page-2")).toBe(false);
    expect(isCurrentLocationPage("success", "page-2")).toBe(false);
  });
});
