import { expect, test, type Page, type Locator } from "@playwright/test";
import { faker } from "@faker-js/faker";
import { registerAndLogin } from "./helpers/auth";

async function createLocation(page: Page, name: string) {
  await expect(page.getByTestId("logout-button")).toBeVisible();
  await page.keyboard.press("Escape");
  await page.keyboard.press("Shift+Digit3");
  const dialog = page.getByRole("dialog").filter({ hasText: "Create Location" }).first();
  await expect(dialog).toBeVisible();
  await dialog.getByLabel("Location Name", { exact: false }).first().fill(name);
  await dialog.getByRole("button", { name: "Create", exact: true }).click();
  await expect(page).toHaveURL(/\/location\/[0-9a-fA-F-]+/);
  await expect(page.getByRole("heading", { name, level: 1 })).toBeVisible();
}

async function gotoTemplatesPage(page: Page) {
  await page.goto("/templates");
  await expect(page).toHaveURL("/templates");
  await expect(page.getByRole("heading", { name: "Templates" }).first()).toBeVisible();
}

async function openCreateTemplateModal(page: Page) {
  // The header button is always present on /templates; the empty-state button appears only when there are none.
  await page.getByTestId("create-template-button").click();
  const dialog = page.getByRole("dialog").filter({ hasText: "Create Template" }).first();
  await expect(dialog).toBeVisible();
  return dialog;
}

async function selectLocationInDialog(dialog: Locator, locationName: string) {
  const combo = dialog.getByRole("combobox", { name: "Parent Location" });
  await combo.click();
  // LocationSelector popover renders a CommandInput (search) + CommandItem options.
  // Clicking an option via force: true doesn't fire reka-ui's internal @select handler.
  // Typing into the search + pressing Enter selects the first match reliably.
  const page = dialog.page();
  const searchInput = page.getByPlaceholder(/search/i).last();
  await searchInput.fill(locationName);
  const option = page.getByRole("option", { name: locationName, exact: true }).first();
  await expect(option).toBeVisible();
  await searchInput.press("Enter");
  await expect(combo).toContainText(locationName);
}

test.describe("Templates CRUD", () => {
  test("create, edit, and delete a template", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const locationName = `Tpl Loc ${faker.string.alphanumeric(6)}`;
    await createLocation(page, locationName);

    await gotoTemplatesPage(page);

    // Should start empty - empty state visible.
    await expect(page.getByText("No templates yet.")).toBeVisible();

    const templateName = `Tpl ${faker.string.alphanumeric(6)}`;
    const dialog = await openCreateTemplateModal(page);

    // Fill template name (first field with label "Template Name").
    await dialog.getByLabel("Template Name", { exact: false }).first().fill(templateName);
    await dialog.getByLabel("Template Description", { exact: false }).first().fill("An e2e-created template");
    await dialog.getByLabel("Item Name", { exact: false }).first().fill("Default item name");

    // Select default location.
    await selectLocationInDialog(dialog, locationName);

    // Submit.
    await dialog.getByTestId("template-create-submit").click();

    // Dialog closes, card appears on the templates page.
    await expect(dialog).toBeHidden();
    const card = page.getByTestId(`template-card-${templateName}`);
    await expect(card).toBeVisible();

    // Navigate into the template detail page via the edit link on the card.
    await card.getByTestId("template-card-edit").click();
    await expect(page).toHaveURL(/\/template\/[0-9a-fA-F-]+/);
    await expect(page.getByRole("heading", { name: templateName }).first()).toBeVisible();
    // Default location should appear in the detail view.
    await expect(page.getByText(locationName, { exact: false }).first()).toBeVisible();

    // Edit: change the template name.
    const renamed = `${templateName} edited`;
    await page.getByTestId("template-detail-edit").click();
    const editDialog = page.getByRole("dialog").filter({ hasText: "Edit Template" }).first();
    await expect(editDialog).toBeVisible();
    await editDialog.getByLabel("Template Name", { exact: false }).first().fill(renamed);
    await editDialog.getByTestId("template-update-submit").click();

    // The detail view title should reflect the new name.
    await expect(editDialog).toBeHidden();
    await expect(page.getByRole("heading", { name: renamed }).first()).toBeVisible();

    // Delete the template via the detail page.
    await page.getByTestId("template-detail-delete").click();
    await page.getByRole("alertdialog").getByRole("button", { name: "Confirm" }).click();

    // After deletion we land back on /templates and the card is gone.
    await expect(page).toHaveURL("/templates");
    await expect(page.getByTestId(`template-card-${renamed}`)).toHaveCount(0);
    await expect(page.getByText("No templates yet.")).toBeVisible();
  });

  test("delete a template from the templates page card", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);
    await gotoTemplatesPage(page);

    const templateName = `Tpl ${faker.string.alphanumeric(6)}`;
    const dialog = await openCreateTemplateModal(page);
    await dialog.getByLabel("Template Name", { exact: false }).first().fill(templateName);
    await dialog.getByTestId("template-create-submit").click();
    await expect(dialog).toBeHidden();

    const card = page.getByTestId(`template-card-${templateName}`);
    await expect(card).toBeVisible();

    await card.getByTestId("template-card-delete").click();
    await page.getByRole("alertdialog").getByRole("button", { name: "Confirm" }).click();

    await expect(page.getByTestId(`template-card-${templateName}`)).toHaveCount(0);
  });

  test("apply template in Create Item modal populates defaults", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const locationName = `Tpl Loc ${faker.string.alphanumeric(6)}`;
    await createLocation(page, locationName);

    await gotoTemplatesPage(page);

    const templateName = `Apply Tpl ${faker.string.alphanumeric(6)}`;
    const defaultItemName = `Default item ${faker.string.alphanumeric(6)}`;
    const defaultDesc = "Default description for template";

    const dialog = await openCreateTemplateModal(page);
    await dialog.getByLabel("Template Name", { exact: false }).first().fill(templateName);
    await dialog.getByLabel("Item Name", { exact: false }).first().fill(defaultItemName);
    await dialog.getByLabel("Item Description", { exact: false }).first().fill(defaultDesc);
    await selectLocationInDialog(dialog, locationName);
    await dialog.getByTestId("template-create-submit").click();
    await expect(dialog).toBeHidden();
    await expect(page.getByTestId(`template-card-${templateName}`)).toBeVisible();

    // Go to /home so the CreateItem hotkey is unambiguous.
    await page.goto("/home");
    await expect(page).toHaveURL("/home");
    await expect(page.getByTestId("logout-button")).toBeVisible();

    // Open the Create Item dialog via the Shift+1 shortcut.
    await page.keyboard.press("Escape");
    await page.keyboard.press("Shift+Digit1");
    const itemDialog = page.getByRole("dialog").filter({ hasText: "Create Item" }).first();
    await expect(itemDialog).toBeVisible();

    // Open the compact template selector and pick the template.
    await itemDialog.getByTestId("template-selector-compact").click();
    const templateOption = page.getByRole("option", { name: templateName }).first();
    await expect(templateOption).toBeVisible();
    // The compact template popover animation can keep the option "not stable";
    // force the click to avoid flakiness.
    await templateOption.click({ force: true });

    // Template banner should render the template name.
    await expect(itemDialog.getByText(`Using template: ${templateName}`)).toBeVisible();

    // Verify the form fields were populated from the template defaults.
    await expect(itemDialog.getByLabel("Item Name", { exact: false }).first()).toHaveValue(defaultItemName);
    await expect(itemDialog.getByLabel("Item Description", { exact: false }).first()).toHaveValue(defaultDesc);

    // Close and reopen the Create Item dialog so restoreLastTemplate() runs on
    // the next mount and authoritatively overrides the default location with
    // the template's location. (handleTemplateSelected only overrides when
    // form.location is empty, so if the locations store has already loaded a
    // first seeded location during initial mount we'd miss it.) Use the dialog's
    // explicit Close button rather than Escape — if a future "unsaved changes"
    // guard intercepts Escape, this teardown would hang; the DialogClose X is
    // always present and always dismissable. Update this if the UX around
    // template application or dialog close changes.
    await itemDialog.getByRole("button", { name: /Close/i }).click();
    await expect(itemDialog).toBeHidden();
    await page.keyboard.press("Shift+Digit1");
    await expect(itemDialog).toBeVisible();

    // The restored template banner should still show.
    await expect(itemDialog.getByText(`Using template: ${templateName}`)).toBeVisible();
    // The location combobox should now contain the template's default location.
    await expect(itemDialog.getByRole("combobox", { name: "Parent Location" })).toContainText(locationName);
    // And the other defaults should still be populated.
    await expect(itemDialog.getByLabel("Item Name", { exact: false }).first()).toHaveValue(defaultItemName);
    await expect(itemDialog.getByLabel("Item Description", { exact: false }).first()).toHaveValue(defaultDesc);
  });
});
