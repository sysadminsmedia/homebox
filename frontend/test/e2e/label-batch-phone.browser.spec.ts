import type { Locator, Page } from "@playwright/test";
import { expect, test } from "@playwright/test";

const EMAIL = "demo@example.com";
const PASSWORD = process.env.E2E_PASSWORD || "demodemo";

// Logs in through the API, as the customs inventory spec does: WebKit on Linux drops the
// cookies of a form login.
async function apiToken(page: Page): Promise<string> {
  const res = await page.request.post("/api/v1/users/login", {
    data: { username: EMAIL, password: PASSWORD, stayLoggedIn: false },
  });
  expect(res.ok()).toBeTruthy();
  const { token } = (await res.json()) as { token: string };
  return token.startsWith("Bearer") ? token : `Bearer ${token}`;
}

// A sheet template shows every control of the page, including the sheet options and the
// hints in the "Send to" buttons.
async function createSheetTemplate(page: Page, token: string, name: string): Promise<string> {
  const res = await page.request.post("/api/v1/label-templates", {
    headers: { Authorization: token },
    data: {
      name,
      description: "",
      width: 66.7,
      height: 25.4,
      dpi: 300,
      isShared: false,
      outputFormat: "png",
      preset: "avery_5160",
      canvasData: { version: "6.0.0", objects: [] },
    },
  });
  expect(res.ok()).toBeTruthy();
  const { id } = (await res.json()) as { id: string };
  return id;
}

async function openWithTemplate(page: Page, name: string) {
  await page.goto("/label-templates/batch");
  await expect(page.locator("li[draggable]").first()).toBeVisible();
  // A sheet template that is already selected adds a second combobox, for the page size.
  await page.locator("aside").getByRole("combobox").first().click();
  await page.getByRole("option", { name }).click();
  await expect(page.getByLabel("Skip first")).toBeVisible();
}

async function expectInsideViewport(page: Page, control: Locator) {
  await expect(control).toBeVisible();
  const box = await control.boundingBox();
  // clientWidth leaves out the scroll bar, which takes up space on WebKit.
  const visibleWidth = await page.evaluate(() => document.documentElement.clientWidth);
  expect(box).not.toBeNull();
  expect(box!.x).toBeGreaterThanOrEqual(0);
  expect(box!.x + box!.width).toBeLessThanOrEqual(visibleWidth);
}

test.describe("batch print page layout", () => {
  let token: string;
  let templateId: string;
  let templateName: string;

  test.beforeEach(async ({ page }, testInfo) => {
    token = await apiToken(page);
    templateName = `e2e phone layout ${testInfo.project.name} ${testInfo.testId} ${Date.now()}`;
    templateId = await createSheetTemplate(page, token, templateName);
  });

  test.afterEach(async ({ page }) => {
    await page.request.delete(`/api/v1/label-templates/${templateId}`, { headers: { Authorization: token } });
  });

  for (const width of [375, 420]) {
    test(`fits a ${width}px wide screen`, async ({ page }) => {
      await page.setViewportSize({ width, height: 812 });
      await openWithTemplate(page, templateName);

      // <main> hides its overflow, so content that is cut off never shows up in the scroll
      // width of the document. Both have to be measured.
      const scrollWidths = await page.evaluate(() => {
        const root = document.documentElement;
        const main = document.querySelector("main")!;
        return {
          document: root.scrollWidth - root.clientWidth,
          main: main.scrollWidth - main.clientWidth,
        };
      });
      expect(scrollWidths).toEqual({ document: 0, main: 0 });

      const list = page.locator("main section").first();
      const filters = [
        list.getByPlaceholder("Search"),
        list.getByRole("button", { name: "Tags" }),
        list.getByRole("combobox", { name: "Order By" }),
        list.getByRole("combobox", { name: "Order Direction" }),
        list.getByRole("button", { name: "Items", exact: true }),
      ];
      for (const filter of filters) await expectInsideViewport(page, filter);

      await expectInsideViewport(page, page.getByRole("spinbutton", { name: "Copies" }).first());

      const destinations = page.getByRole("radio");
      expect(await destinations.count()).toBeGreaterThanOrEqual(2);
      for (const destination of await destinations.all()) await expectInsideViewport(page, destination);

      await expectInsideViewport(page, page.getByRole("button", { name: "Clear selection" }));
      await expectInsideViewport(page, page.getByRole("button", { name: /^Download PDF/ }));
    });
  }

  test("keeps the two columns on a wide screen", async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 800 });
    await openWithTemplate(page, templateName);

    const list = await page.locator("main section").first().boundingBox();
    const side = await page.locator("aside").boundingBox();
    expect(side!.width).toBe(400);
    expect(list!.x + list!.width).toBeLessThan(side!.x);
  });
});
