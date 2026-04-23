import { expect, test, type Locator, type Page } from "@playwright/test";
import { faker } from "@faker-js/faker";
import { registerAndLogin } from "./helpers/auth";

async function createLocation(page: Page, name: string) {
  await expect(page.getByTestId("logout-button")).toBeVisible();
  await page.keyboard.press("Escape");
  await page.keyboard.press("Shift+Digit3");
  const dialog = page.getByRole("dialog").filter({ hasText: "Create Location" });
  await expect(dialog).toBeVisible();
  await dialog.getByLabel("Location Name", { exact: false }).first().fill(name);
  await dialog.getByRole("button", { name: "Create", exact: true }).click();
  await expect(page).toHaveURL(/\/location\/[0-9a-f-]+/i);
}

async function createItem(page: Page, itemName: string, locationName: string): Promise<string> {
  // Navigate to /home so the CreateItem modal doesn't auto-populate parent from
  // the current /location/<id> route — we want to exercise the manual select.
  await page.goto("/home");
  await expect(page.getByTestId("logout-button")).toBeVisible();
  await page.keyboard.press("Escape");
  await page.keyboard.press("Shift+Digit1");
  const dialog = page.getByRole("dialog").filter({ hasText: "Create Item" });
  await expect(dialog).toBeVisible();

  await dialog.getByRole("combobox", { name: "Parent Location" }).click();
  const search = page.getByPlaceholder("Search Locations");
  await expect(search).toBeVisible();
  await search.fill(locationName);
  const option = page.getByRole("option", { name: locationName, exact: true }).first();
  await expect(option).toBeVisible();
  await search.press("Enter");
  await expect(dialog.getByRole("combobox", { name: "Parent Location" })).toContainText(locationName);

  await dialog.getByLabel("Item Name", { exact: false }).first().fill(itemName);
  await dialog.getByRole("button", { name: "Create", exact: true }).click();

  await page.waitForURL(/\/item\/[a-f0-9-]+$/);
  const match = page.url().match(/\/item\/([a-f0-9-]+)/);
  if (!match) {
    throw new Error(`Could not determine item id from URL: ${page.url()}`);
  }
  return match[1]!;
}

async function showAllMaintenance(page: Page) {
  // Default filter is "Scheduled"; switch to "Both" so completed and unscheduled entries show.
  await page.getByRole("button", { name: "Both", exact: true }).click();
}

/**
 * Returns a locator for the maintenance entry card on either the global
 * /maintenance view or an item's /item/<id>/maintenance view.  The cards
 * themselves are rendered as `Card` (`div`) — we anchor to the Delete
 * button (which only lives inside a card) and walk up to the card root.
 */
function maintenanceCard(page: Page, entryName: string): Locator {
  // BaseCard renders as a div with shadow-xl overflow-hidden. Target the
  // card root directly via its shadow class so we don't accidentally grab
  // an outer wrapper.
  return page.locator("div.shadow-xl").filter({ hasText: entryName });
}

async function createMaintenanceEntry(
  page: Page,
  itemId: string,
  opts: { name: string; cost: string; notes?: string }
) {
  // The backend requires either completedDate or scheduledDate to be set. The UI's
  // DatePicker is a third-party calendar widget that's brittle to drive from Playwright,
  // so we create the record via the REST API directly. The flow under test is the
  // display/edit/delete UI, not the create dialog itself (covered elsewhere).
  const today = new Date().toISOString().slice(0, 10);
  const res = await page.request.post(`/api/v1/entities/${itemId}/maintenance`, {
    data: {
      name: opts.name,
      description: opts.notes ?? "",
      cost: opts.cost,
      scheduledDate: today,
      completedDate: "",
    },
  });
  expect(res.status(), `maintenance create: ${await res.text()}`).toBeLessThan(400);
}

test.describe("Maintenance records", () => {
  test.slow();

  test("create on item page, edit and delete from /maintenance, currency rendering", async ({ page }) => {
    await registerAndLogin(page);

    const locationName = `Loc-${faker.string.alphanumeric(6)}`;
    const itemName = `Item-${faker.string.alphanumeric(6)}`;
    await createLocation(page, locationName);
    const itemId = await createItem(page, itemName, locationName);

    const entryName = `Oil Change ${faker.string.alphanumeric(4)}`;
    const entryCost = "12.34";
    const entryNotes = "Initial service";
    await createMaintenanceEntry(page, itemId, { name: entryName, cost: entryCost, notes: entryNotes });

    // Entry appears on the item's maintenance page with USD-formatted cost
    await page.goto(`/item/${itemId}/maintenance`);
    await showAllMaintenance(page);
    const itemCard = maintenanceCard(page, entryName);
    await expect(itemCard).toBeVisible();
    await expect(itemCard).toContainText("$12.34");

    // Global /maintenance page shows the entry with a link back to the item
    await page.goto("/maintenance");
    await showAllMaintenance(page);
    const globalCard = maintenanceCard(page, entryName);
    await expect(globalCard).toBeVisible();
    await expect(globalCard.getByRole("link", { name: itemName })).toBeVisible();
    await expect(globalCard).toContainText("$12.34");

    const updatedName = `${entryName} (updated)`;
    const updatedCost = "99.99";
    await globalCard.getByRole("button", { name: "Edit", exact: true }).click();

    const editDialog = page.getByRole("dialog").filter({ hasText: "Edit Entry" });
    await expect(editDialog).toBeVisible();
    await editDialog.getByLabel("Entry Name", { exact: false }).first().fill(updatedName);
    await editDialog.getByLabel("Cost", { exact: false }).first().fill(updatedCost);
    await editDialog.getByRole("button", { name: "Update", exact: true }).click();
    await expect(editDialog).toBeHidden();

    const updatedCard = maintenanceCard(page, updatedName);
    await expect(updatedCard).toBeVisible();
    await expect(updatedCard).toContainText("$99.99");

    await page.goto(`/item/${itemId}/maintenance`);
    await showAllMaintenance(page);
    const itemUpdatedCard = maintenanceCard(page, updatedName);
    await expect(itemUpdatedCard).toBeVisible();
    await expect(itemUpdatedCard).toContainText("$99.99");

    await page.goto("/maintenance");
    await showAllMaintenance(page);
    const cardForDelete = maintenanceCard(page, updatedName);
    await expect(cardForDelete).toBeVisible();
    await cardForDelete.getByRole("button", { name: "Delete", exact: true }).click();

    const confirmDialog = page.getByRole("alertdialog");
    await expect(confirmDialog).toBeVisible();
    await confirmDialog.getByRole("button", { name: /Confirm/i }).click();

    // After deletion, the card should no longer be present on either view.
    await expect(maintenanceCard(page, updatedName)).toHaveCount(0);
    await page.goto(`/item/${itemId}/maintenance`);
    await showAllMaintenance(page);
    await expect(maintenanceCard(page, updatedName)).toHaveCount(0);
  });
});
