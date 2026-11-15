import { describe, expect, test } from "vitest";
import { useShowItems, useTreeState } from "./tree-state";

// Leaving the Locations page and coming back mounts the page again, and every
// mount asks for its state by the same key. Expand/Collapse survived that and
// Show/Hide Items did not, because the second was a ref local to the page
// (#1577).
describe("the location tree remembers its view across a revisit", () => {
  test("a hidden-items choice survives a second mount", () => {
    const first = useShowItems("revisit-tree");
    first.value = false;

    const second = useShowItems("revisit-tree");

    expect(second.value).toBe(false);
  });

  test("it is kept exactly as long as the expanded state beside it", () => {
    useTreeState("paired-tree").value["abc12345"] = true;
    useShowItems("paired-tree").value = false;

    expect(useTreeState("paired-tree").value["abc12345"]).toBe(true);
    expect(useShowItems("paired-tree").value).toBe(false);
  });

  test("a tree nobody has touched still shows its items", () => {
    expect(useShowItems("untouched-tree").value).toBe(true);
  });

  test("two trees keep separate choices", () => {
    useShowItems("tree-a").value = false;

    expect(useShowItems("tree-b").value).toBe(true);
  });
});
