import { expect, test, type Locator, type Page } from "@playwright/test";
import { faker } from "@faker-js/faker";
import { registerAndLogin } from "./helpers/auth";

function getDialog(page: Page, titleText: string): Locator {
  return page.getByRole("dialog").filter({ has: page.getByText(titleText, { exact: true }) });
}

async function openCreateTagDialog(page: Page) {
  await expect(page.getByTestId("logout-button")).toBeVisible();
  await page.keyboard.press("Escape");
  await page.keyboard.press("Shift+Digit2");
  await expect(getDialog(page, "Create Tag").first()).toBeVisible();
}

async function fillTagName(dialog: Locator, name: string) {
  await dialog.getByLabel("Tag Name", { exact: false }).first().fill(name);
}

async function selectIcon(dialog: Locator, iconName: string) {
  await dialog.getByRole("button", { name: `Select ${iconName} icon` }).click();
}

async function randomizeColor(dialog: Locator): Promise<string> {
  await dialog.getByRole("button", { name: "Randomize color" }).click();
  const hexLocator = dialog.locator("span.font-mono").first();
  await expect(hexLocator).toHaveText(/^#[0-9a-fA-F]{6}$/);
  return (await hexLocator.textContent()) || "";
}

async function submitCreateAndExpectNavigation(dialog: Locator, page: Page) {
  await dialog.getByRole("button", { name: "Create", exact: true }).click();
  await expect(page).toHaveURL(/\/tag\/[0-9a-f-]+/i);
}

async function submitUpdate(dialog: Locator) {
  await dialog.getByRole("button", { name: "Update", exact: true }).click();
}

async function createTagWithName(page: Page, name: string) {
  await openCreateTagDialog(page);
  const dialog = getDialog(page, "Create Tag");
  await fillTagName(dialog, name);
  await submitCreateAndExpectNavigation(dialog, page);
  await expect(page.getByRole("heading", { name })).toBeVisible();
}

test.describe("Tag CRUD", () => {
  test("create tag with name only", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);
    await createTagWithName(page, `tag-${faker.string.alphanumeric(8).toLowerCase()}`);
  });

  test("create tag with a color", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const tagName = `color-${faker.string.alphanumeric(8).toLowerCase()}`;
    await openCreateTagDialog(page);
    const dialog = getDialog(page, "Create Tag");
    await fillTagName(dialog, tagName);
    const hex = await randomizeColor(dialog);
    expect(hex).toMatch(/^#[0-9a-fA-F]{6}$/);
    await submitCreateAndExpectNavigation(dialog, page);
    await expect(page.getByRole("heading", { name: tagName })).toBeVisible();
  });

  test("create tag with an icon", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const tagName = `icon-${faker.string.alphanumeric(8).toLowerCase()}`;
    await openCreateTagDialog(page);
    const dialog = getDialog(page, "Create Tag");
    await fillTagName(dialog, tagName);
    await selectIcon(dialog, "laptop");
    await submitCreateAndExpectNavigation(dialog, page);
    await expect(page.getByRole("heading", { name: tagName })).toBeVisible();
  });

  test("edit a tag: rename, change color, change icon", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const originalName = `orig-${faker.string.alphanumeric(8).toLowerCase()}`;
    const renamedName = `renamed-${faker.string.alphanumeric(8).toLowerCase()}`;
    await createTagWithName(page, originalName);

    await page.getByRole("button", { name: "Edit", exact: true }).click();
    const updateDialog = getDialog(page, "Update Tag");
    await expect(updateDialog.first()).toBeVisible();

    await fillTagName(updateDialog, renamedName);
    const newHex = await randomizeColor(updateDialog);
    expect(newHex).toMatch(/^#[0-9a-fA-F]{6}$/);
    await selectIcon(updateDialog, "wrench-outline");
    await submitUpdate(updateDialog);

    await expect(page.getByText("Tag updated", { exact: false }).first()).toBeVisible();
    await expect(page.getByRole("heading", { name: renamedName })).toBeVisible();
  });

  test("delete a tag", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const tagName = `del-${faker.string.alphanumeric(8).toLowerCase()}`;
    await createTagWithName(page, tagName);

    await page.getByRole("button", { name: "Delete", exact: true }).click();

    const alert = page.getByRole("alertdialog");
    await expect(alert).toBeVisible();
    await alert.getByRole("button", { name: "Confirm", exact: true }).click();

    await expect(page).toHaveURL("/home");
    await expect(page.getByText("Tag deleted", { exact: false }).first()).toBeVisible();
  });

  test("empty-name update is rejected", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const tagName = `empty-${faker.string.alphanumeric(8).toLowerCase()}`;
    await createTagWithName(page, tagName);

    await page.getByRole("button", { name: "Edit", exact: true }).click();
    const updateDialog = getDialog(page, "Update Tag");
    await expect(updateDialog.first()).toBeVisible();

    await fillTagName(updateDialog, "");
    await submitUpdate(updateDialog);

    // Dialog stays open (submit rejected); the page URL and the existing tag have not changed.
    await expect(updateDialog.first()).toBeVisible();
    await expect(page).toHaveURL(/\/tag\/[0-9a-f-]+/i);

    // A "Tag updated" toast would only appear if the backend silently accepted
    // the empty name — assert no such toast is present.
    await expect(page.getByText(/Tag updated/i)).toHaveCount(0);

    // Close the dialog so the detail page is no longer aria-hidden behind it,
    // then confirm the original tag name is still rendered (i.e. a no-op client
    // handler didn't blank it out).
    await updateDialog.getByRole("button", { name: /Close/i }).click();
    await expect(updateDialog).toBeHidden();
    await expect(page.getByRole("heading", { name: tagName })).toBeVisible();
  });
});
