import { expect, test, type Page } from "@playwright/test";
import { faker } from "@faker-js/faker";
import { registerAndLogin, STRONG_PASSWORD } from "./helpers/auth";

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

function createLocationDialog(page: Page) {
  return page.getByRole("dialog").filter({ has: page.getByText("Create Location", { exact: true }) });
}

async function fillLocationName(page: Page, name: string) {
  const dialog = createLocationDialog(page);
  await dialog.getByLabel("Location Name", { exact: false }).first().fill(name);
}

async function fillLocationDescription(page: Page, description: string) {
  const dialog = createLocationDialog(page);
  await dialog.getByLabel("Location Description", { exact: false }).first().fill(description);
}

async function expandAdvanced(page: Page) {
  const dialog = createLocationDialog(page);
  await dialog.getByRole("button", { name: "Show more", exact: false }).click();
}

async function fillLocationNotes(page: Page, notes: string) {
  const dialog = createLocationDialog(page);
  await dialog.getByLabel("Notes", { exact: true }).first().fill(notes);
}

async function submitCreate(page: Page) {
  const dialog = createLocationDialog(page);
  await dialog.getByRole("button", { name: "Create", exact: true }).click();
  // After a successful create we navigate to /location/<id>.
  await expect(page).toHaveURL(/\/location\/[0-9a-f-]+/i);
}

test.describe("Location CRUD", () => {
  test("create root location with name only", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const name = `loc-${faker.string.alphanumeric(8).toLowerCase()}`;
    await openCreateLocationDialog(page);
    await fillLocationName(page, name);
    await submitCreate(page);

    await expect(page.getByRole("heading", { name, level: 1 })).toBeVisible();
  });

  test("create location with description and notes", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const name = `full-${faker.string.alphanumeric(8).toLowerCase()}`;
    const description = `desc-${faker.string.alphanumeric(16).toLowerCase()}`;
    const notes = `notes-${faker.string.alphanumeric(16).toLowerCase()}`;

    // The Create Location modal persists name, description, and tags via
    // the EntityCreate API; notes are not part of the create payload and
    // must be set through the edit form. So we create with name +
    // description here, then edit to add notes below.
    await openCreateLocationDialog(page);
    await fillLocationName(page, name);
    await fillLocationDescription(page, description);
    await submitCreate(page);

    await expect(page.getByTestId("location-detail-name")).toHaveText(new RegExp(name));
    await expect(page.getByText(description, { exact: false }).first()).toBeVisible();

    // Add notes via the edit page, since the create modal does not
    // persist notes on the backend.
    await page.getByRole("button", { name: "Edit", exact: true }).click();
    await expect(page).toHaveURL(/\/location\/[0-9a-f-]+\/edit/i);

    // The Notes MarkdownEditor renders its label text in an inner
    // <span class="truncate"> and the <Textarea autosize> id falls
    // through to a wrapper <div>, so <label for=id> is not associated
    // with the textarea. Anchor on the label's inner span and walk to
    // the nearest sibling textarea.
    const notesTextarea = page
      .locator("span.truncate")
      .filter({ hasText: /^Notes$/ })
      .first()
      .locator("xpath=ancestor::div[contains(@class,'w-full')][1]//textarea")
      .first();
    await expect(notesTextarea).toBeVisible();
    await notesTextarea.fill(notes);

    await page.getByRole("button", { name: "Save", exact: true }).click();

    await expect(page).toHaveURL(/\/location\/[0-9a-f-]+$/i);
    await expect(page.getByTestId("location-detail-name")).toHaveText(new RegExp(name));
    await expect(page.getByText(description, { exact: false }).first()).toBeVisible();
    // Notes are rendered inside the Details section, not the header card.
    await expect(page.getByText(notes, { exact: false }).first()).toBeVisible();
  });

  test("edit a location's name and description", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const originalName = `orig-${faker.string.alphanumeric(8).toLowerCase()}`;
    const renamedName = `renamed-${faker.string.alphanumeric(8).toLowerCase()}`;
    const newDescription = `desc-${faker.string.alphanumeric(16).toLowerCase()}`;

    await openCreateLocationDialog(page);
    await fillLocationName(page, originalName);
    await submitCreate(page);
    await expect(page.getByTestId("location-detail-name")).toHaveText(new RegExp(originalName));

    await page.getByRole("button", { name: "Edit", exact: true }).click();
    await expect(page).toHaveURL(/\/location\/[0-9a-f-]+\/edit/i);

    // Name is a FormTextField (inline) — the <input id=id> receives the
    // label's `for`, so getByLabel works. The label's accessible name
    // includes a length indicator ("Name 0/255"), so match loosely.
    // "Name" can also appear in the hidden-by-default custom fields
    // section, so .first() consistently picks the main Name input.
    const nameInput = page.getByLabel("Name", { exact: false }).first();
    await expect(nameInput).toBeVisible();
    await nameInput.fill(renamedName);

    // Description uses MarkdownEditor. Its <Label for=id> points to the
    // <Textarea autosize> wrapper <div>, not the textarea itself, so
    // getByLabel does not resolve to the form control. Use the same
    // anchor pattern as items-advanced-fields.spec.ts: locate the label's
    // inner span, then walk up to the MarkdownEditor root and find the
    // textarea.
    const descriptionTextarea = page
      .locator("span.truncate")
      .filter({ hasText: /^Description$/ })
      .first()
      .locator("xpath=ancestor::div[contains(@class,'w-full')][1]//textarea")
      .first();
    await expect(descriptionTextarea).toBeVisible();
    await descriptionTextarea.fill(newDescription);

    await page.getByRole("button", { name: "Save", exact: true }).click();

    await expect(page).toHaveURL(/\/location\/[0-9a-f-]+$/i);
    await expect(page.getByText("Location updated", { exact: false }).first()).toBeVisible();
    await expect(page.getByTestId("location-detail-name")).toHaveText(new RegExp(renamedName));
    await expect(page.getByText(newDescription, { exact: false }).first()).toBeVisible();
  });

  test("delete a location with confirmation", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const name = `del-${faker.string.alphanumeric(8).toLowerCase()}`;
    await openCreateLocationDialog(page);
    await fillLocationName(page, name);
    await submitCreate(page);
    await expect(page.getByRole("heading", { name, level: 1 })).toBeVisible();

    await page.getByRole("button", { name: "Delete", exact: true }).click();

    // ModalConfirm uses an alertdialog with a "Confirm" action.
    const alert = page.getByRole("alertdialog");
    await expect(alert).toBeVisible();
    await alert.getByRole("button", { name: "Confirm", exact: true }).click();

    // On successful delete we redirect to /locations.
    await expect(page).toHaveURL("/locations");
    await expect(page.getByText("Location deleted", { exact: false }).first()).toBeVisible();
  });

  test("empty-name create is rejected", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    await openCreateLocationDialog(page);
    const dialog = createLocationDialog(page);
    await dialog.getByRole("button", { name: "Create", exact: true }).click();

    // The Location name input is HTML-required, so the form should not submit:
    // the dialog stays open and the URL does not change to /location/<id>.
    await expect(dialog).toBeVisible();
    await expect(page).not.toHaveURL(/\/location\/[0-9a-f-]+/i);

    const nameInput = dialog.getByLabel("Location Name", { exact: false }).first();
    const isInvalid = await nameInput.evaluate(
      (el: HTMLInputElement) => !el.validity.valid && el.validity.valueMissing
    );
    expect(isInvalid).toBe(true);
  });
});
