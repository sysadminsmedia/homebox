import { expect, test, type Locator, type Page } from "@playwright/test";
import { faker } from "@faker-js/faker";
import { registerAndLogin } from "./helpers/auth";

function getCreateTagDialog(page: Page): Locator {
  return page.getByRole("dialog").filter({ has: page.getByText("Create Tag", { exact: true }) });
}

async function openCreateTagDialog(page: Page) {
  await expect(page.getByTestId("logout-button")).toBeVisible();
  // Blur any focused input (seeded sidebar cards, search boxes, etc.) so the
  // Shift+Digit2 hotkey is handled by the global dispatcher.
  await page.keyboard.press("Escape");
  await page.keyboard.press("Shift+Digit2");
  await expect(getCreateTagDialog(page).first()).toBeVisible();
}

async function createTag(page: Page, name: string, parentName?: string) {
  await openCreateTagDialog(page);
  const dialog = getCreateTagDialog(page).first();

  await dialog.getByLabel("Tag Name", { exact: false }).first().fill(name);

  if (parentName) {
    // Parent selector is a Popover whose trigger has role="combobox". The option
    // list renders in a portal. reka-ui's @select handler doesn't fire on a
    // direct option click, so drive it via the CommandInput + Enter. The popover's
    // CommandInput uses the tag-selector "Select Tags" placeholder.
    await dialog.getByRole("combobox").first().click();
    const search = page.getByPlaceholder("Select Tags");
    await search.fill(parentName);
    await expect(page.getByRole("option").filter({ hasText: parentName }).first()).toBeVisible();
    await search.press("Enter");
  }

  // Submit form — the page will navigate to /tag/[id] on success.
  await dialog.getByRole("button", { name: "Create", exact: true }).click();
  await expect(page).toHaveURL(/\/tag\/[0-9a-f-]+/i);
}

/**
 * Locator for the link-text of a tag node in the tree (the <NuxtLink> rendered
 * inside Tag/Tree/Node.vue). This is the most stable hook available — the tree
 * components don't yet expose data-testid attributes.
 */
function treeNodeLink(page: Page, name: string): Locator {
  return page.locator(`a[href^="/tag/"]`).filter({ hasText: new RegExp(`^${escapeRegExp(name)}$`) });
}

function escapeRegExp(s: string): string {
  return s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

/**
 * The clickable row (div wrapping the chevron + icon + link) in a tree node.
 * Returned locator toggles the node when clicked OFF the inner link.
 */
function treeNodeRow(page: Page, name: string): Locator {
  // Parent of the link: the flex row with @click handler.
  return treeNodeLink(page, name).locator("xpath=..");
}

/**
 * Click the chevron area of a node row to toggle it (avoids the NuxtLink which
 * uses @click.stop and would navigate instead of toggle).
 */
async function toggleTreeNode(page: Page, name: string) {
  // The chevron wrapper has [data-swap] only when the node has children.
  const chevron = treeNodeRow(page, name).locator("[data-swap]").first();
  await expect(chevron).toBeVisible();
  await chevron.click();
}

test.describe("tags tree view", () => {
  test.beforeEach(async ({ page }) => {
    test.slow();
    await registerAndLogin(page);
  });

  test("creates nested tags and displays hierarchy in tree", async ({ page }) => {
    const parentName = `Parent-${faker.string.alphanumeric(6)}`;
    const childName = `Child-${faker.string.alphanumeric(6)}`;

    // Create parent tag first.
    await createTag(page, parentName);
    await expect(page.getByRole("heading", { name: parentName })).toBeVisible();

    // Create child tag with parent selected.
    await createTag(page, childName, parentName);

    // Detail page of child should show breadcrumb containing parent name.
    await expect(page.getByRole("heading", { name: childName })).toBeVisible();
    // There should be a chip link pointing at the parent tag's detail page.
    await expect(page.locator(`a[href*='/tag/']`, { hasText: parentName }).first()).toBeVisible();

    // Navigate to /tags tree view.
    await page.goto("/tags");
    await expect(page).toHaveURL(/\/tags/);

    // Parent node exists at root level.
    const parentLink = treeNodeLink(page, parentName);
    await expect(parentLink).toBeVisible();

    // Child should not be visible yet (node collapsed by default).
    await expect(treeNodeLink(page, childName)).toHaveCount(0);

    // Expand the parent and the child should appear.
    await toggleTreeNode(page, parentName);
    await expect(treeNodeLink(page, childName)).toBeVisible();

    // Collapse again — child disappears.
    await toggleTreeNode(page, parentName);
    await expect(treeNodeLink(page, childName)).toHaveCount(0);
  });

  test("expand-all and collapse-all controls toggle the whole tree", async ({ page }) => {
    const parentName = `Root-${faker.string.alphanumeric(6)}`;
    const childName = `Leaf-${faker.string.alphanumeric(6)}`;

    await createTag(page, parentName);
    await createTag(page, childName, parentName);

    await page.goto("/tags");
    await expect(treeNodeLink(page, parentName)).toBeVisible();

    const expandButton = page.getByTestId("tag-tree-expand-all");
    const collapseButton = page.getByTestId("tag-tree-collapse-all");

    // Make sure allTags has finished loading and the tree has rendered the parent
    // before firing expand-all — if openAll runs while `tree.value` is still [],
    // it walks no nodes and sets no state.
    await page.waitForLoadState("networkidle");
    await expect(treeNodeLink(page, parentName)).toBeVisible();
    await expect(treeNodeLink(page, childName)).toHaveCount(0);

    await expandButton.click();
    await expect(treeNodeLink(page, childName)).toBeVisible({ timeout: 15000 });

    await collapseButton.click();
    await expect(treeNodeLink(page, childName)).toHaveCount(0, { timeout: 15000 });
  });

  test("clicking tree node link navigates to the tag detail page", async ({ page }) => {
    const tagName = `Nav-${faker.string.alphanumeric(6)}`;

    await createTag(page, tagName);
    await page.goto("/tags");

    const link = treeNodeLink(page, tagName);
    await expect(link).toBeVisible();
    await link.click();

    await expect(page).toHaveURL(/\/tag\/[0-9a-f-]+/i);
    await expect(page.getByRole("heading", { name: tagName })).toBeVisible();
  });

  test("tag detail page renders icon and hierarchy breadcrumb for nested tags", async ({ page }) => {
    const parentName = `P-${faker.string.alphanumeric(6)}`;
    const childName = `C-${faker.string.alphanumeric(6)}`;

    await createTag(page, parentName);
    await createTag(page, childName, parentName);

    // We should now be on child's detail page.
    await expect(page).toHaveURL(/\/tag\/[0-9a-f-]+/i);

    // Heading shows the child's name.
    await expect(page.getByRole("heading", { name: childName })).toBeVisible();

    // Breadcrumb shows a chip linking to the parent tag.
    const parentChip = page.locator(`a[href*='/tag/']`, { hasText: parentName }).first();
    await expect(parentChip).toBeVisible();

    // Icon container (rounded-full element with svg) is rendered near the title.
    await expect(page.locator("svg").first()).toBeVisible();

    // "Created" metadata is present on the detail page.
    await expect(page.getByText(/created/i).first()).toBeVisible();
  });

  test("empty tree shows create-tag call-to-action", async ({ page }) => {
    await page.goto("/tags");
    // Brand new user, no tags — the Root component renders an empty-state Create button.
    const emptyCta = page.getByRole("button", { name: /create/i }).first();
    await expect(emptyCta).toBeVisible();
  });
});
