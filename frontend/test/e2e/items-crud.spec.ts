import { expect, test, type Page } from "@playwright/test";
import { faker } from "@faker-js/faker";
import { registerAndLogin } from "./helpers/auth";

/**
 * Open the "Create Location" dialog via the Shift+3 hotkey
 * (see Location/CreateModal.vue -> useDialogHotkey).
 */
async function openCreateLocationDialog(page: Page) {
  await expect(page.getByTestId("logout-button")).toBeVisible();
  await page.keyboard.press("Escape");
  await page.keyboard.press("Shift+Digit3");
  await expect(page.getByRole("dialog").getByText("Create Location", { exact: true }).first()).toBeVisible();
}

/**
 * Open the "Create Item" dialog via the Shift+1 hotkey
 * (see Item/CreateModal.vue -> useDialogHotkey).
 *
 * We navigate to /home first because pressing Shift+1 on /location/<id> can
 * occasionally land on a different/stale dialog if there is leftover focus
 * from the previously-closed Create Location modal.
 */
async function openCreateItemDialog(page: Page) {
  // Always start from /home: pressing Shift+1 on /location/<id> can
  // occasionally land on a different/stale dialog if there is leftover focus
  // from the previously-closed Create Location modal.
  await page.goto("/home");
  await expect(page.getByTestId("logout-button")).toBeVisible();
  await page.keyboard.press("Escape");
  await page.keyboard.press("Shift+Digit1");
  await expect(page.getByRole("dialog").getByText("Create Item", { exact: true }).first()).toBeVisible();
}

function createLocationDialog(page: Page) {
  return page.getByRole("dialog").filter({ has: page.getByText("Create Location", { exact: true }) });
}

function createItemDialog(page: Page) {
  return page.getByRole("dialog").filter({ has: page.getByText("Create Item", { exact: true }) });
}

/**
 * Create a location and return its name. Navigates to /location/<id> on success.
 */
async function createLocation(page: Page) {
  const locationName = `loc-${faker.string.alphanumeric(8).toLowerCase()}`;
  await openCreateLocationDialog(page);
  const dialog = createLocationDialog(page);
  await dialog.getByLabel("Location Name", { exact: false }).first().fill(locationName);
  await dialog.getByRole("button", { name: "Create", exact: true }).click();
  await expect(page).toHaveURL(/\/location\/[0-9a-f-]+/i);
  await expect(page.getByRole("heading", { name: locationName, level: 1 })).toBeVisible();
  return locationName;
}

/**
 * Select a location inside the currently-open Create Item dialog via the
 * LocationSelector combobox popover.
 *
 * Note: the Create Item dialog also contains a TemplateSelector rendered as
 * a role="combobox" icon button in the header, so the location selector is
 * NOT the first combobox in the dialog. Scope by accessible name via the
 * associated "Parent Location" label.
 */
async function selectLocationInItemDialog(page: Page, locationName: string) {
  const dialog = createItemDialog(page);
  const locationCombobox = dialog.getByRole("combobox", { name: "Parent Location" }).first();
  await expect(locationCombobox).toBeVisible();
  await locationCombobox.click();
  // The popover renders outside the dialog in a portal, so query the page directly.
  await page.getByRole("option", { name: locationName, exact: false }).first().click();
  await expect(locationCombobox).toContainText(locationName);
}

/**
 * Submit the Create Item dialog and wait until we land on /item/<id>.
 */
async function submitCreateItem(page: Page) {
  const dialog = createItemDialog(page);
  await dialog.getByRole("button", { name: "Create", exact: true }).click();
  await expect(page).toHaveURL(/\/item\/[0-9a-f-]+$/i);
}

test.describe("Item CRUD", () => {
  test("create item with name and location only", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const locationName = await createLocation(page);

    const itemName = `item-${faker.string.alphanumeric(8).toLowerCase()}`;
    await openCreateItemDialog(page);
    await selectLocationInItemDialog(page, locationName);
    const dialog = createItemDialog(page);
    await dialog.getByLabel("Item Name", { exact: false }).first().fill(itemName);
    await submitCreateItem(page);

    await expect(page.getByRole("heading", { name: itemName, level: 1 })).toBeVisible();
    await expect(page.getByText("Item created", { exact: false }).first()).toBeVisible();
  });

  test("create item with description and quantity from the modal", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const locationName = await createLocation(page);

    const itemName = `item-${faker.string.alphanumeric(8).toLowerCase()}`;
    const description = faker.lorem.sentence();
    const quantity = 7;

    await openCreateItemDialog(page);
    await selectLocationInItemDialog(page, locationName);

    const dialog = createItemDialog(page);
    await dialog.getByLabel("Item Name", { exact: false }).first().fill(itemName);
    await dialog.getByLabel("Item Quantity", { exact: false }).first().fill(String(quantity));
    await dialog.getByLabel("Item Description", { exact: false }).first().fill(description);
    await submitCreateItem(page);

    await expect(page.getByRole("heading", { name: itemName, level: 1 })).toBeVisible();
    await expect(page.getByText(description, { exact: false }).first()).toBeVisible();
    // Quantity is surfaced from the item API. Hit the backend directly to
    // avoid brittle text assertions against a bare number that can appear in
    // many places on the detail page (asset IDs, dates, etc.).
    const itemUrl = new URL(page.url());
    const itemIdMatch = itemUrl.pathname.match(/\/item\/([0-9a-f-]+)/i);
    expect(itemIdMatch, "item id should be present in URL").not.toBeNull();
    const getRes = await page.request.get(`/api/v1/entities/${itemIdMatch![1]}`);
    expect(getRes.ok(), "GET /entities/<id> should succeed").toBe(true);
    const body = await getRes.json();
    expect(body.quantity, "quantity should round-trip via the API").toBe(quantity);
  });

  test("edit an item to add manufacturer, model, and serial", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const locationName = await createLocation(page);

    const itemName = `item-${faker.string.alphanumeric(8).toLowerCase()}`;
    await openCreateItemDialog(page);
    await selectLocationInItemDialog(page, locationName);
    const createDialog = createItemDialog(page);
    await createDialog.getByLabel("Item Name", { exact: false }).first().fill(itemName);
    await submitCreateItem(page);
    await expect(page.getByRole("heading", { name: itemName, level: 1 })).toBeVisible();

    await page.getByRole("link", { name: "Edit", exact: true }).first().click();
    await expect(page).toHaveURL(/\/item\/[0-9a-f-]+\/edit$/i);

    const manufacturer = faker.company.name();
    const modelNumber = faker.string.alphanumeric(10);
    const serialNumber = faker.string.alphanumeric(12);
    const renamedName = `renamed-${faker.string.alphanumeric(8).toLowerCase()}`;

    // The inline FormTextField renders the char-count alongside the label text
    // (e.g. "Name 13/255"), so exact label matching does not work. The labels
    // themselves are unique among the visible edit-form inputs.
    await page.getByRole("textbox", { name: /^Name/ }).first().fill(renamedName);
    await page.getByRole("textbox", { name: /^Manufacturer/ }).first().fill(manufacturer);
    await page.getByRole("textbox", { name: /^Model Number/ }).first().fill(modelNumber);
    await page.getByRole("textbox", { name: /^Serial Number/ }).first().fill(serialNumber);

    await page.getByRole("button", { name: "Save", exact: true }).click();

    await expect(page).toHaveURL(/\/item\/[0-9a-f-]+$/i);
    await expect(page.getByRole("heading", { name: renamedName, level: 1 })).toBeVisible();
    await expect(page.getByText(manufacturer, { exact: false }).first()).toBeVisible();
    await expect(page.getByText(modelNumber, { exact: false }).first()).toBeVisible();
    await expect(page.getByText(serialNumber, { exact: false }).first()).toBeVisible();
  });

  test("delete an item via the confirmation dialog", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const locationName = await createLocation(page);

    const itemName = `del-${faker.string.alphanumeric(8).toLowerCase()}`;
    await openCreateItemDialog(page);
    await selectLocationInItemDialog(page, locationName);
    const dialog = createItemDialog(page);
    await dialog.getByLabel("Item Name", { exact: false }).first().fill(itemName);
    await submitCreateItem(page);
    await expect(page.getByRole("heading", { name: itemName, level: 1 })).toBeVisible();

    await page.getByRole("button", { name: "More actions" }).click();
    // The menu item contains an icon + text; use non-exact matching and scope
    // to the visible open menu to avoid matching the sidebar "Delete" text.
    await page.getByRole("menu").getByRole("menuitem", { name: "Delete", exact: false }).click();

    const confirmDialog = page.getByRole("alertdialog");
    await expect(confirmDialog).toBeVisible();
    await confirmDialog.getByRole("button", { name: "Confirm", exact: true }).click();

    await expect(page).toHaveURL("/home");
    await expect(page.getByText("Item deleted", { exact: false }).first()).toBeVisible();
  });

  test("create without a location shows a required-field error", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    // Skip creating a location so form.location is empty on submit — the create
    // handler should surface the "Please select a location." toast and keep the
    // modal open instead of navigating.
    const itemName = `no-loc-${faker.string.alphanumeric(8).toLowerCase()}`;
    await openCreateItemDialog(page);
    const dialog = createItemDialog(page);
    await dialog.getByLabel("Item Name", { exact: false }).first().fill(itemName);
    await dialog.getByRole("button", { name: "Create", exact: true }).click();

    await expect(dialog).toBeVisible();
    await expect(page).not.toHaveURL(/\/item\/[0-9a-f-]+/i);
    await expect(page.getByText("Please select a location", { exact: false }).first()).toBeVisible();
  });
});
