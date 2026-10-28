import { describe, expect, test } from "vitest";
import { clampPageToTotal } from "./pagination";

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
