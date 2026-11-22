import type { Page } from "@playwright/test";
import { expect, test } from "@playwright/test";

const EMAIL = "demo@example.com";
const PASSWORD = process.env.E2E_PASSWORD || "demodemo";

async function login(page: Page) {
  await page.goto("/home");
  await expect(page).toHaveURL("/");
  await page.fill("input[type='text']", EMAIL);
  await page.fill("input[type='password']", PASSWORD);
  await page.click("button[type='submit']");
  await expect(page).toHaveURL("/home");
}

async function apiToken(page: Page): Promise<string> {
  const res = await page.request.post("/api/v1/users/login", {
    data: { username: EMAIL, password: PASSWORD, stayLoggedIn: false },
  });
  expect(res.ok()).toBeTruthy();
  const { token } = (await res.json()) as { token: string };
  return token.startsWith("Bearer") ? token : `Bearer ${token}`;
}

function detailValue(page: Page, label: string) {
  return page.locator("dl > div", { has: page.locator("dt", { hasText: label }) }).locator("dd");
}

test.describe("template default boolean flags", () => {
  let token: string;
  let templateId: string | undefined;

  test.beforeEach(async ({ page }) => {
    await login(page);
    token = await apiToken(page);
    templateId = undefined;
  });

  test.afterEach(async ({ page }) => {
    if (templateId) {
      await page.request.delete(`/api/v1/templates/${templateId}`, {
        headers: { Authorization: token },
      });
    }
  });

  test("insured and lifetime warranty switches persist on create and edit", async ({ page }) => {
    const templateName = `e2e-flags-${Date.now()}`;

    await page.goto("/templates");
    await page.getByRole("main").getByRole("button", { name: "Create", exact: true }).click();

    const createDialog = page.getByRole("dialog");
    await expect(createDialog).toBeVisible();
    await createDialog.getByLabel("Template Name").fill(templateName);

    const createInsured = createDialog.locator("#defaultInsured");
    const createWarranty = createDialog.locator("#defaultLifetimeWarranty");
    await createInsured.click();
    await createWarranty.click();
    await expect(createInsured).toHaveAttribute("aria-checked", "true");
    await expect(createWarranty).toHaveAttribute("aria-checked", "true");

    await createDialog.getByRole("button", { name: "Create", exact: true }).click();
    await expect(createDialog).toBeHidden();
    await expect(page.getByText(templateName)).toBeVisible();

    const listRes = await page.request.get("/api/v1/templates", {
      headers: { Authorization: token },
    });
    expect(listRes.ok()).toBeTruthy();
    const templates = (await listRes.json()) as Array<{
      id: string;
      name: string;
    }>;
    templateId = templates.find(item => item.name === templateName)?.id;
    expect(templateId).toBeTruthy();

    await page.goto(`/template/${templateId}`);
    await expect(detailValue(page, "Insured")).toHaveText("Yes");
    await expect(detailValue(page, "Lifetime Warranty")).toHaveText("Yes");

    await page.reload();
    await expect(detailValue(page, "Insured")).toHaveText("Yes");
    await expect(detailValue(page, "Lifetime Warranty")).toHaveText("Yes");

    await page.getByRole("button", { name: "Edit", exact: true }).click();
    const editDialog = page.getByRole("dialog");
    await expect(editDialog).toBeVisible();

    const editInsured = editDialog.locator("#editInsured");
    const editWarranty = editDialog.locator("#editWarranty");
    await expect(editInsured).toHaveAttribute("aria-checked", "true");
    await expect(editWarranty).toHaveAttribute("aria-checked", "true");

    await editInsured.click();
    await editWarranty.click();
    await expect(editInsured).toHaveAttribute("aria-checked", "false");
    await expect(editWarranty).toHaveAttribute("aria-checked", "false");

    await editDialog.getByRole("button", { name: "Update", exact: true }).click();
    await expect(editDialog).toBeHidden();
    await expect(detailValue(page, "Insured")).toHaveText("No");
    await expect(detailValue(page, "Lifetime Warranty")).toHaveText("No");

    await page.reload();
    await expect(detailValue(page, "Insured")).toHaveText("No");
    await expect(detailValue(page, "Lifetime Warranty")).toHaveText("No");
  });
});
