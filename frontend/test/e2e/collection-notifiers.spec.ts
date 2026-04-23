import { expect, test, type Locator, type Page } from "@playwright/test";
import { faker } from "@faker-js/faker";
import { registerAndLogin } from "./helpers/auth";

const VALID_URL = "generic://example.com/webhook?template=json";

async function gotoNotifiers(page: Page) {
  await page.goto("/collection/notifiers");
  await expect(page).toHaveURL(/\/collection\/notifiers/);
  await expect(page.getByRole("heading", { name: "Notifiers", exact: true })).toBeVisible({ timeout: 10000 });
}

async function openCreateDialog(page: Page): Promise<Locator> {
  await page.getByRole("main").getByRole("button", { name: "Create", exact: true }).click();
  const dialog = page.getByRole("dialog");
  await expect(dialog).toBeVisible();
  // Note: the component renders "Edit Notifier" in both create and edit modes due to an
  // existing i18n condition bug, so we assert on the form fields instead of the title.
  await expect(dialog.getByLabel("Name", { exact: true })).toBeVisible();
  await expect(dialog.getByLabel("URL", { exact: true })).toBeVisible();
  return dialog;
}

async function openEditDialog(article: Locator, page: Page): Promise<Locator> {
  // Edit is an icon-only button with no accessible name; it is the second (outline) action
  // inside the article's action group. Use the inner `data-button` element (tooltip wraps
  // the real button) and pick index 1 for the edit (pencil) action.
  await article.locator("[data-button]").nth(1).click();
  const dialog = page.getByRole("dialog");
  await expect(dialog).toBeVisible();
  await expect(dialog.getByLabel("Name", { exact: true })).toBeVisible();
  await expect(dialog.getByLabel("URL", { exact: true })).toBeVisible();
  return dialog;
}

async function fillNotifierForm(dialog: Locator, name: string, url: string) {
  await dialog.getByLabel("Name", { exact: true }).fill(name);
  await dialog.getByLabel("URL", { exact: true }).fill(url);
}

async function submitAndWaitClose(dialog: Locator) {
  await dialog.getByRole("button", { name: "Submit", exact: true }).click();
  await expect(dialog).toBeHidden({ timeout: 5000 });
}

async function createNotifier(page: Page, name: string, url = VALID_URL) {
  const dialog = await openCreateDialog(page);
  await fillNotifierForm(dialog, name, url);
  await submitAndWaitClose(dialog);
}

test.describe("Collection notifiers", () => {
  test.beforeEach(async ({ page }) => {
    test.slow();
    await registerAndLogin(page);
    await gotoNotifiers(page);
  });

  test("shows empty state when no notifiers exist", async ({ page }) => {
    await expect(page.getByText("No notifiers configured")).toBeVisible();
  });

  test("creates a notifier with a valid URL", async ({ page }) => {
    const notifierName = `Notifier ${faker.string.alphanumeric(8)}`;
    await createNotifier(page, notifierName);

    const article = page.locator("article").filter({ hasText: notifierName });
    await expect(article).toBeVisible({ timeout: 5000 });
    await expect(article.getByText("Active", { exact: true })).toBeVisible();
  });

  test("rejects an invalid URL", async ({ page }) => {
    const notifierName = `Invalid ${faker.string.alphanumeric(6)}`;
    const dialog = await openCreateDialog(page);
    await fillNotifierForm(dialog, notifierName, "not-a-valid-url");
    await dialog.getByRole("button", { name: "Submit", exact: true }).click();

    await expect(page.getByText("Failed to create notifier.")).toBeVisible({ timeout: 5000 });
    await expect(page.locator("article").filter({ hasText: notifierName })).toHaveCount(0);
  });

  test("edits an existing notifier", async ({ page }) => {
    const originalName = `Original ${faker.string.alphanumeric(6)}`;
    const updatedName = `Updated ${faker.string.alphanumeric(6)}`;

    await createNotifier(page, originalName);

    const article = page.locator("article").filter({ hasText: originalName });
    const dialog = await openEditDialog(article, page);
    await dialog.getByLabel("Name", { exact: true }).fill(updatedName);
    await submitAndWaitClose(dialog);

    await expect(page.locator("article").filter({ hasText: updatedName })).toBeVisible();
    await expect(page.locator("article").filter({ hasText: originalName })).toHaveCount(0);
  });

  test("toggles the active flag", async ({ page }) => {
    const notifierName = `Toggle ${faker.string.alphanumeric(6)}`;
    await createNotifier(page, notifierName);

    const article = page.locator("article").filter({ hasText: notifierName });
    await expect(article.getByText("Active", { exact: true })).toBeVisible();

    let dialog = await openEditDialog(article, page);
    await dialog.getByRole("checkbox").click();
    await submitAndWaitClose(dialog);
    await expect(article.getByText("Inactive", { exact: true })).toBeVisible({ timeout: 5000 });

    dialog = await openEditDialog(article, page);
    await dialog.getByRole("checkbox").click();
    await submitAndWaitClose(dialog);
    await expect(article.getByText("Active", { exact: true })).toBeVisible({ timeout: 5000 });
  });

  test("deletes a notifier", async ({ page }) => {
    const notifierName = `Doomed ${faker.string.alphanumeric(6)}`;
    await createNotifier(page, notifierName);

    const article = page.locator("article").filter({ hasText: notifierName });
    await expect(article).toBeVisible();

    // Delete is the first (destructive) icon-only button within the article.
    await article.locator("[data-button]").nth(0).click();
    await page.getByRole("alertdialog").getByRole("button", { name: "Confirm" }).click();

    await expect(page.locator("article").filter({ hasText: notifierName })).toHaveCount(0, { timeout: 5000 });
  });
});
