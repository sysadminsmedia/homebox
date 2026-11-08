import { describe, expect, test } from "vitest";
import { isExactAssetIdSearch, orderSelection } from "./label-batch";

describe("isExactAssetIdSearch", () => {
  test.each(["#000-123", "#123"])("accepts %s", q => {
    expect(isExactAssetIdSearch(q)).toBe(true);
  });
  test.each(["#>=100", "#>=100 <=200", "#abc", "#", "books"])("rejects %s", q => {
    expect(isExactAssetIdSearch(q)).toBe(false);
  });
});

describe("orderSelection", () => {
  test("follows the list order", () => {
    expect(orderSelection(["c", "a", "b"], ["a", "b", "c", "d"])).toEqual(["a", "b", "c"]);
  });
  test("keeps selected items outside the list at the end", () => {
    expect(orderSelection(["x", "c", "y", "a"], ["a", "b", "c"])).toEqual(["a", "c", "x", "y"]);
  });
  test("keeps the order when nothing was reordered", () => {
    expect(orderSelection(["a", "b"], ["a", "b"])).toEqual(["a", "b"]);
  });
  test("handles an empty selection", () => {
    expect(orderSelection([], ["a"])).toEqual([]);
  });
});
