import type { Locator, Page } from "@playwright/test";
import { expect, test } from "@playwright/test";

const EMAIL = "demo@example.com";
const PASSWORD = process.env.E2E_PASSWORD || "demodemo";
const FILE_NAME = /^Inventarliste_[a-z0-9-]+_\d{4}-\d{2}-\d{2}\.(xlsx|zip)$/;

// Logs in through the API rather than the login form. The server sets its auth cookies
// with Domain=localhost, which WebKit on Linux drops when the page receives them, so a
// form login leaves the WebKit run signed out. Playwright's request context keeps the
// cookies and shares them with the page on every browser.
async function login(page: Page) {
  const res = await page.request.post("/api/v1/users/login", {
    data: { username: EMAIL, password: PASSWORD, stayLoggedIn: false },
  });
  expect(res.ok()).toBeTruthy();
}

// Selects every row of the results table and returns how many are selected.
async function selectAllRows(page: Page): Promise<number> {
  const header = page.locator("thead");
  const selected = page.getByText(/^(\d+) of \1 rows selected\.$/).first();
  // The search can still be re-rendering the table; retry until the whole page is selected.
  await expect(async () => {
    if ((await header.getByRole("checkbox").getAttribute("data-state")) !== "checked") {
      await header.getByRole("checkbox").click();
    }
    await expect(selected).toBeVisible({ timeout: 1000 });
  }).toPass({ timeout: 15000 });
  return Number((await selected.textContent())!.match(/^(\d+)/)![1]);
}

// Opens the export dialog with the button in the search results header.
async function openExportDialog(page: Page): Promise<Locator> {
  await page.getByRole("button", { name: "Export inventory list …" }).click();
  const dialog = page.getByRole("dialog", { name: "Export customs inventory list" });
  await expect(dialog).toBeVisible();
  return dialog;
}

// The dialog restores the last used options, so set the checkbox explicitly.
async function setIncludeInvoices(dialog: Locator, on: boolean) {
  const checkbox = dialog.locator("#customs-receipts");
  if (((await checkbox.getAttribute("data-state")) === "checked") !== on) {
    await dialog.getByText("Include invoices (download as ZIP)").click();
  }
  await expect(checkbox).toHaveAttribute("data-state", on ? "checked" : "unchecked");
}

async function exportFile(page: Page, dialog: Locator): Promise<string> {
  const exportButton = dialog.getByRole("button", { name: /^Export( anyway)?$/ });
  await expect(exportButton).toBeEnabled();
  const [download] = await Promise.all([page.waitForEvent("download"), exportButton.click()]);
  await expect(dialog).toBeHidden();
  return download.suggestedFilename();
}

test.describe("Customs inventory export", () => {
  test("exports the customs inventory list from search", async ({ page }) => {
    await login(page);
    await page.goto("/search");
    await page.getByRole("button", { name: "Table" }).click();
    await page.waitForLoadState("networkidle");
    await expect(page.getByText(/rows selected/).first()).toBeVisible();

    const count = await selectAllRows(page);

    // The item table's actions menu no longer offers the export.
    await page.locator("thead").getByRole("button", { name: "Open menu" }).click();
    await expect(page.getByRole("menuitem", { name: /Customs Inventory/ })).toHaveCount(0);
    await page.keyboard.press("Escape");

    let dialog = await openExportDialog(page);
    await expect(dialog.getByRole("radio", { name: `Only selected items (${count})` })).toBeChecked();

    // Pick the 15th of the month the picker opens on as the reference date.
    await dialog.locator(".dp__input").click();
    await page.locator(".dp__menu .dp__cell_inner:not(.dp__cell_offset)", { hasText: /^15$/ }).click();
    await page.locator(".dp__action_select").click();

    const total = dialog.getByText("Total value").locator("..");
    await expect(total).toContainText(/\d/);

    // Without invoices: only the workbook.
    await setIncludeInvoices(dialog, false);
    const xlsx = await exportFile(page, dialog);
    expect(xlsx).toMatch(FILE_NAME);
    expect(xlsx.endsWith(".xlsx")).toBe(true);

    // With invoices on, the answer is a ZIP.
    dialog = await openExportDialog(page);
    await expect(total).toContainText(/\d/);
    await setIncludeInvoices(dialog, true);
    const zip = await exportFile(page, dialog);
    expect(zip).toMatch(FILE_NAME);
    expect(zip.endsWith(".zip")).toBe(true);

    // A new search clears the selection, so no stale rows are exported.
    const searched = page.waitForResponse(
      r => r.url().includes("/api/v1/entities") && r.url().includes("zzz-no-match")
    );
    await page.getByRole("textbox", { name: "Search" }).fill("zzz-no-match");
    await page.keyboard.press("Enter");
    await searched;
    dialog = await openExportDialog(page);
    await expect(dialog.getByRole("radio", { name: "All items in the current search (0)" })).toBeVisible();
    await expect(dialog.getByRole("radio", { name: "Only selected items (0)" })).toBeDisabled();
  });
});
