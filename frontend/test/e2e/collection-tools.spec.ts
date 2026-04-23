import { expect, test, type Page } from "@playwright/test";
import { registerAndLogin } from "./helpers/auth";

const TOOLS_ROUTE = "/collection/tools";
const IMPORT_API = "**/api/v1/entities/import**";
const ENSURE_IDS_API = "**/api/v1/actions/ensure-asset-ids";

const ENSURE_IDS_CONFIRM_TEXT =
  "Are you sure you want to ensure all assets have an ID? This can take a while and cannot be undone.";

async function openTools(page: Page) {
  await page.goto(TOOLS_ROUTE);
  await page.waitForLoadState("networkidle");
  await expect(page.getByRole("heading", { name: "Import Inventory" })).toBeVisible();
}

async function clickEnsureAssetIDs(page: Page) {
  await page.getByRole("button", { name: "Ensure Asset IDs", exact: true }).first().click();
  await expect(page.getByText(ENSURE_IDS_CONFIRM_TEXT)).toBeVisible();
}

test.describe("Collection Tools Page", () => {
  test.beforeEach(async ({ page }) => {
    test.slow();
    await registerAndLogin(page);
  });

  test("CSV export triggers a file download event", async ({ page }) => {
    await openTools(page);

    const downloadPromise = page.waitForEvent("download", { timeout: 15_000 });
    await page.getByRole("button", { name: "Export Inventory", exact: true }).click();

    const download = await downloadPromise;
    const filename = download.suggestedFilename();
    expect(filename.length).toBeGreaterThan(0);
    expect(filename.toLowerCase()).toContain(".csv");
  });

  test("CSV import dialog opens and keeps submit disabled without a file", async ({ page }) => {
    await openTools(page);
    await page.getByRole("button", { name: "Import Inventory", exact: true }).click();

    await expect(page.getByRole("heading", { name: "Import CSV File" })).toBeVisible();
    await expect(page.locator("input[type='file']")).toBeVisible();
    await expect(page.getByRole("button", { name: "Submit" })).toBeDisabled();
  });

  test("CSV import rejects malformed input via API error response", async ({ page }) => {
    await page.route(IMPORT_API, async route => {
      await route.fulfill({
        status: 422,
        contentType: "application/json",
        body: JSON.stringify({ message: "invalid columns" }),
      });
    });

    await openTools(page);
    await page.getByRole("button", { name: "Import Inventory", exact: true }).click();
    await expect(page.getByRole("heading", { name: "Import CSV File" })).toBeVisible();

    await page.locator("input[type='file']").setInputFiles({
      name: "malformed.csv",
      mimeType: "text/csv",
      buffer: Buffer.from("not,the,right,columns\n1,2,3,4\n"),
    });

    const submitBtn = page.getByRole("button", { name: "Submit" });
    await expect(submitBtn).toBeEnabled();

    const requestPromise = page.waitForRequest(IMPORT_API);
    await submitBtn.click();
    await requestPromise;

    await expect(page.getByText("Import failed. Please try again later.")).toBeVisible();
  });

  test("Ensure Asset IDs confirmation can be canceled", async ({ page }) => {
    await openTools(page);
    await clickEnsureAssetIDs(page);

    await page.getByRole("button", { name: "Cancel", exact: true }).click();
    await expect(page.getByText(ENSURE_IDS_CONFIRM_TEXT)).toBeHidden();
  });

  test("Ensure Asset IDs completes successfully and reports count", async ({ page }) => {
    await page.route(ENSURE_IDS_API, async route => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ completed: 7 }),
      });
    });

    await openTools(page);
    await clickEnsureAssetIDs(page);

    const requestPromise = page.waitForRequest(ENSURE_IDS_API);
    await page.getByRole("button", { name: "Confirm", exact: true }).last().click();
    await requestPromise;

    await expect(page.getByText("7 assets have been updated.")).toBeVisible();
  });

  test("Ensure Asset IDs shows error toast on API failure", async ({ page }) => {
    await page.route(ENSURE_IDS_API, async route => {
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ message: "boom" }),
      });
    });

    await openTools(page);
    await clickEnsureAssetIDs(page);

    const requestPromise = page.waitForRequest(ENSURE_IDS_API);
    await page.getByRole("button", { name: "Confirm", exact: true }).last().click();
    await requestPromise;

    await expect(page.getByText("Failed to ensure asset IDs.")).toBeVisible();
  });
});
