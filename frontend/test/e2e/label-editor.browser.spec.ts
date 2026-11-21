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

async function createTemplate(page: Page, token: string): Promise<string> {
  const res = await page.request.post("/api/v1/label-templates", {
    headers: { Authorization: token },
    data: {
      name: `e2e tool rail ${Date.now()}`,
      description: "",
      width: 62,
      height: 29,
      dpi: 300,
      isShared: false,
      outputFormat: "png",
      preset: "brother_dk1209",
      canvasData: { version: "6.0.0", objects: [] },
    },
  });
  expect(res.ok()).toBeTruthy();
  const { id } = (await res.json()) as { id: string };
  return id;
}

/**
 * Opens a tool-rail menu and checks that it is actually placed next to the button.
 * A menu that exists but was never positioned sits above the viewport, so a plain
 * visibility check is not enough.
 */
async function expectMenuOpens(page: Page, buttonName: string) {
  const button = page.getByRole("button", { name: buttonName, exact: true });
  await button.hover();
  await button.click();
  const menu = page.locator("[role='menu'][data-state='open']");
  await expect(menu).toBeVisible();
  const box = await menu.boundingBox();
  const viewport = page.viewportSize()!;
  expect(box).not.toBeNull();
  expect(box!.x).toBeGreaterThanOrEqual(0);
  expect(box!.y).toBeGreaterThanOrEqual(0);
  expect(box!.x + box!.width).toBeLessThanOrEqual(viewport.width);
  expect(box!.y + box!.height).toBeLessThanOrEqual(viewport.height);
  await page.keyboard.press("Escape");
  await expect(menu).toBeHidden();
}

test.describe("label template editor", () => {
  let token: string;
  let templateId: string;

  test.beforeEach(async ({ page }) => {
    await login(page);
    token = await apiToken(page);
    templateId = await createTemplate(page, token);
    await page.goto(`/label-templates/${templateId}/edit`);
    await expect(page.getByRole("toolbar")).toBeVisible();
  });

  test.afterEach(async ({ page }) => {
    if (templateId) {
      await page.request.delete(`/api/v1/label-templates/${templateId}`, { headers: { Authorization: token } });
    }
  });

  test("tool rail menus open next to their buttons", async ({ page }) => {
    await expectMenuOpens(page, "Add Barcode");
    await expectMenuOpens(page, "Add Shape");

    // Align is only enabled with a selection; adding a text element selects it.
    await page.getByRole("button", { name: "Add Text", exact: true }).click();
    await expect(page.getByRole("button", { name: "Align", exact: true })).toBeEnabled();
    await expectMenuOpens(page, "Align");
  });

  test("a barcode can be added from the tool rail", async ({ page }) => {
    await page.getByRole("button", { name: "Add Barcode", exact: true }).click();
    await page.getByRole("menuitem", { name: /QR Code/ }).hover();
    await page.getByRole("menuitem", { name: "Item URL" }).click();
    await page.getByRole("tab", { name: /Layers/ }).click();
    const layers = page.getByRole("complementary", { name: "Inspector" }).getByRole("listitem");
    await expect(layers).toHaveCount(1);
    await expect(layers.first()).toContainText(/QR/i);
  });
});
