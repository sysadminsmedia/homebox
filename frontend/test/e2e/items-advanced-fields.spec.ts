import { expect, test, type Page, type APIRequestContext, type Locator } from "@playwright/test";
import { faker } from "@faker-js/faker";
import { registerAndLogin, STRONG_PASSWORD } from "./helpers/auth";

async function apiCreateEntity(
  request: APIRequestContext,
  name: string,
  opts: { parentId?: string; entityTypeId?: string } = {}
): Promise<string> {
  const resp = await request.post("/api/v1/entities", {
    data: {
      name,
      description: "",
      quantity: 1,
      tagIds: [],
      ...opts,
    },
  });
  expect(resp.ok(), `entity create failed: ${resp.status()} ${await resp.text()}`).toBeTruthy();
  const body = (await resp.json()) as { id: string };
  return body.id;
}

async function getLocationTypeId(request: APIRequestContext): Promise<string> {
  const resp = await request.get("/api/v1/entity-types");
  expect(resp.ok(), `entity-types GET failed: ${resp.status()}`).toBeTruthy();
  const data = (await resp.json()) as Array<{ id: string; isLocation: boolean }>;
  const loc = data.find(d => d.isLocation);
  if (!loc) throw new Error("no location entity-type found");
  return loc.id;
}

async function setupItemAndOpenEdit(page: Page): Promise<{ itemId: string }> {
  await registerAndLogin(page);
  const locationName = `loc-${faker.string.alphanumeric(8).toLowerCase()}`;
  const itemName = `item-${faker.string.alphanumeric(8).toLowerCase()}`;
  const entityTypeId = await getLocationTypeId(page.request);
  const locationId = await apiCreateEntity(page.request, locationName, { entityTypeId });
  const itemId = await apiCreateEntity(page.request, itemName, { parentId: locationId });

  await page.goto(`/item/${itemId}/edit`);
  await expect(page).toHaveURL(new RegExp(`/item/${itemId}/edit$`));
  await expect(page.getByRole("heading", { name: "Edit Details" })).toBeVisible();
  return { itemId };
}

async function ensureSwitchOn(page: Page, name: string) {
  const sw = page.getByRole("switch", { name });
  await expect(sw).toBeVisible();
  if ((await sw.getAttribute("data-state")) !== "checked") {
    await sw.click();
    await expect(sw).toHaveAttribute("data-state", "checked");
  }
}

async function saveAndReturn(page: Page, itemId: string) {
  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(page).toHaveURL(new RegExp(`/item/${itemId}$`));
}

/** Scope to the shadcn `<Card>` that hosts the given `<h3>` section title. */
function cardBySectionTitle(page: Page, title: string): Locator {
  return page
    .getByRole("heading", { name: title, level: 3 })
    .locator("xpath=ancestor::div[contains(@class,'rounded-lg')][1]");
}

/** The detail page renders each field as a `<dt>`; return the enclosing row. */
function detailRow(page: Page, label: string): Locator {
  return page
    .locator("dt")
    .filter({ hasText: new RegExp(`^${label}$`) })
    .first()
    .locator("..");
}

async function fillDatePicker(scope: Locator, page: Page, value: string, options: { allowMissing?: boolean } = {}) {
  const input = scope.locator("input[aria-label='Select Date']").first();
  if ((await input.count()) === 0) {
    if (options.allowMissing) return;
    throw new Error(
      `fillDatePicker: no 'input[aria-label="Select Date"]' found in the given scope — ` +
        `the test is about to write '${value}' to a field that doesn't exist. ` +
        `If a caller legitimately expects the picker to be absent, pass { allowMissing: true }.`
    );
  }
  await input.fill(value);
  await page.keyboard.press("Enter");
}

test.describe("Item advanced fields", () => {
  test("purchase details (price, purchased from, date) persist and render", async ({ page }) => {
    test.slow();
    const { itemId } = await setupItemAndOpenEdit(page);
    await ensureSwitchOn(page, "Advanced");

    const purchaseFrom = `seller-${faker.string.alphanumeric(6).toLowerCase()}`;
    const purchaseCard = cardBySectionTitle(page, "Purchase Details");
    await purchaseCard.getByLabel("Purchased From").first().fill(purchaseFrom);
    await purchaseCard.getByLabel("Purchase Price").first().fill("123.45");
    await fillDatePicker(purchaseCard, page, "01/15/2024");

    await saveAndReturn(page, itemId);

    await ensureSwitchOn(page, "Show Empty");
    await expect(page.getByText("Purchase Details", { exact: true }).first()).toBeVisible();
    await expect(page.getByText(purchaseFrom, { exact: false }).first()).toBeVisible();
    await expect(page.getByText(/123\.45/).first()).toBeVisible();
  });

  test("warranty: lifetime checkbox persists and shows Yes on detail", async ({ page }) => {
    test.slow();
    const { itemId } = await setupItemAndOpenEdit(page);
    await ensureSwitchOn(page, "Advanced");

    const warrantyCard = cardBySectionTitle(page, "Warranty Details");
    const lifetime = warrantyCard.getByRole("checkbox", { name: "Lifetime Warranty" }).first();
    await lifetime.click();
    await expect(lifetime).toHaveAttribute("data-state", "checked");

    await warrantyCard.locator("textarea").first().fill("Covers all **manufacturing** defects.");

    await saveAndReturn(page, itemId);

    await expect(page.getByText("Warranty Details", { exact: true }).first()).toBeVisible();
    await expect(detailRow(page, "Lifetime Warranty")).toContainText("Yes");
  });

  test("warranty: expires-on date without lifetime checkbox", async ({ page }) => {
    test.slow();
    const { itemId } = await setupItemAndOpenEdit(page);
    await ensureSwitchOn(page, "Advanced");

    const warrantyCard = cardBySectionTitle(page, "Warranty Details");
    await fillDatePicker(warrantyCard, page, "06/20/2026");

    await saveAndReturn(page, itemId);

    await expect(page.getByText("Warranty Details", { exact: true }).first()).toBeVisible();
    await expect(detailRow(page, "Lifetime Warranty")).toContainText("No");
    // DetailsSection renders dates via <DateTime format="relative"> which
    // fmtDate expands to "{{relative}} (M/d/yyyy)" at en-US — accept both
    // leading-zero and non-leading-zero month renderings.
    await expect(detailRow(page, "Warranty Expires")).toContainText(/0?6\/20\/2026/);
  });

  test("insurance checkbox toggles to Yes on detail view", async ({ page }) => {
    test.slow();
    const { itemId } = await setupItemAndOpenEdit(page);

    const insured = page.getByRole("checkbox", { name: "Insured" }).first();
    await expect(insured).toHaveAttribute("data-state", "unchecked");
    await insured.click();
    await expect(insured).toHaveAttribute("data-state", "checked");

    await saveAndReturn(page, itemId);

    await expect(detailRow(page, "Insured")).toContainText("Yes");
  });

  test("archive checkbox persists, then unarchive restores No", async ({ page }) => {
    test.slow();
    const { itemId } = await setupItemAndOpenEdit(page);

    const archived = page.getByRole("checkbox", { name: "Archived" }).first();
    await expect(archived).toHaveAttribute("data-state", "unchecked");
    await archived.click();
    await expect(archived).toHaveAttribute("data-state", "checked");

    await saveAndReturn(page, itemId);

    // "Show Empty" must be on so the Archived=No row renders after we unarchive.
    await ensureSwitchOn(page, "Show Empty");
    await expect(detailRow(page, "Archived")).toContainText("Yes");

    await page.goto(`/item/${itemId}/edit`);
    await expect(page).toHaveURL(new RegExp(`/item/${itemId}/edit$`));
    const archivedAgain = page.getByRole("checkbox", { name: "Archived" }).first();
    await expect(archivedAgain).toHaveAttribute("data-state", "checked");
    await archivedAgain.click();
    await expect(archivedAgain).toHaveAttribute("data-state", "unchecked");
    await saveAndReturn(page, itemId);

    await ensureSwitchOn(page, "Show Empty");
    await expect(detailRow(page, "Archived")).toContainText("No");
  });

  test("notes with markdown render as formatted HTML on detail page", async ({ page }) => {
    test.slow();
    const { itemId } = await setupItemAndOpenEdit(page);

    // The Notes MarkdownEditor's <Label> renders its text inside a
    // <span class="truncate">Notes</span>, alongside a separate length
    // indicator span ("N/1000"). Anchor on that inner span and walk up to
    // the MarkdownEditor's root (w-full div), then grab its textarea.
    const notesTextarea = page
      .locator("span.truncate")
      .filter({ hasText: /^Notes$/ })
      .first()
      .locator("xpath=ancestor::div[contains(@class,'w-full')][1]//textarea")
      .first();
    await expect(notesTextarea).toBeVisible();
    await notesTextarea.fill("## Heading\n\n**bold** and *italic* text.");

    await saveAndReturn(page, itemId);

    const notesRow = detailRow(page, "Notes");
    await expect(notesRow).toBeVisible();
    await expect(notesRow.locator("h2").first()).toHaveText(/Heading/);
    await expect(notesRow.locator("strong").first()).toHaveText("bold");
    await expect(notesRow.locator("em").first()).toHaveText("italic");
  });

  test("combined advanced fields round-trip through save", async ({ page }) => {
    test.slow();
    const { itemId } = await setupItemAndOpenEdit(page);
    await ensureSwitchOn(page, "Advanced");

    const insured = page.getByRole("checkbox", { name: "Insured" }).first();
    await insured.click();
    await expect(insured).toHaveAttribute("data-state", "checked");

    const seller = `store-${faker.string.alphanumeric(5).toLowerCase()}`;
    const purchaseCard = cardBySectionTitle(page, "Purchase Details");
    await purchaseCard.getByLabel("Purchased From").first().fill(seller);
    await purchaseCard.getByLabel("Purchase Price").first().fill("99.99");

    const warrantyCard = cardBySectionTitle(page, "Warranty Details");
    const lifetime = warrantyCard.getByRole("checkbox", { name: "Lifetime Warranty" }).first();
    await lifetime.click();
    await expect(lifetime).toHaveAttribute("data-state", "checked");

    await saveAndReturn(page, itemId);

    await page.goto(`/item/${itemId}/edit`);
    await expect(page).toHaveURL(new RegExp(`/item/${itemId}/edit$`));
    await ensureSwitchOn(page, "Advanced");

    await expect(page.getByRole("checkbox", { name: "Insured" }).first()).toHaveAttribute("data-state", "checked");

    const purchaseCardAfter = cardBySectionTitle(page, "Purchase Details");
    await expect(purchaseCardAfter.getByLabel("Purchased From").first()).toHaveValue(seller);
    await expect(purchaseCardAfter.getByLabel("Purchase Price").first()).toHaveValue("99.99");

    const warrantyCardAfter = cardBySectionTitle(page, "Warranty Details");
    await expect(warrantyCardAfter.getByRole("checkbox", { name: "Lifetime Warranty" }).first()).toHaveAttribute(
      "data-state",
      "checked"
    );
  });
});
