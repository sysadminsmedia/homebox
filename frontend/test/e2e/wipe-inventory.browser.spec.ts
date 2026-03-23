import type { Page } from "@playwright/test";
import { expect, test } from "@playwright/test";

const STATUS_ROUTE = "**/api/v1/status";
const WIPE_ROUTE = "**/api/v1/actions/wipe-inventory";

const buildStatusResponse = (demo: boolean) => ({
  allowRegistration: true,
  build: { buildTime: new Date().toISOString(), commit: "test", version: "v0.0.0" },
  demo,
  health: true,
  labelPrinting: false,
  latest: { date: new Date().toISOString(), version: "v0.0.0" },
  message: "",
  oidc: { allowLocal: true, autoRedirect: false, buttonText: "", enabled: false },
  title: "Homebox",
  versions: [],
});

async function mockStatus(page: Page, demo: boolean) {
  await page.route(STATUS_ROUTE, route => {
    route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(buildStatusResponse(demo)),
    });
  });
}

async function login(page: Page, email = "demo@example.com", password = "demo") {
  await page.goto("/home");
  await expect(page).toHaveURL("/");
  await page.fill("input[type='text']", email);
  await page.fill("input[type='password']", password);
  await page.click("button[type='submit']");
  await expect(page).toHaveURL("/home");
}

async function openWipeInventory(page: Page) {
  await page.goto("/tools");
  await page.waitForLoadState("networkidle");
  await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));

  const wipeButton = page.getByRole("button", { name: "Wipe Inventory" }).last();
  await expect(wipeButton).toBeVisible();
  await wipeButton.click();
}

test.describe.skip("Wipe Inventory", () => {
  test("shows demo mode warning without wipe options", async ({ page }) => {
    await mockStatus(page, true);
    await login(page);
    await openWipeInventory(page);

    await expect(
      page.getByText(
        "Inventory, tags, locations and maintenance records cannot be wiped whilst Homebox is in demo mode.",
        { exact: false }
      )
    ).toBeVisible();

    await expect(page.locator("input#wipe-tags-checkbox")).toHaveCount(0);
    await expect(page.locator("input#wipe-locations-checkbox")).toHaveCount(0);
    await expect(page.locator("input#wipe-maintenance-checkbox")).toHaveCount(0);
  });

  test.describe.skip("production mode", () => {
    test.beforeEach(async ({ page }) => {
      await mockStatus(page, false);
      await login(page);
    });

    test.skip("renders wipe options and submits all flags", async ({ page }) => {
      await page.route(WIPE_ROUTE, route => {
        route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ completed: 0 }) });
      });

      await openWipeInventory(page);
      await expect(page.getByText("Wipe Inventory").first()).toBeVisible();

      const tags = page.locator("input#wipe-tags-checkbox");
      const locations = page.locator("input#wipe-locations-checkbox");
      const maintenance = page.locator("input#wipe-maintenance-checkbox");

      await expect(tags).toBeVisible();
      await expect(locations).toBeVisible();
      await expect(maintenance).toBeVisible();

      await tags.check();
      await locations.check();
      await maintenance.check();

      const requestPromise = page.waitForRequest(WIPE_ROUTE);
      await page.getByRole("button", { name: "Confirm" }).last().click();
      const request = await requestPromise;

      expect(request.postDataJSON()).toEqual({
        wipeTags: true,
        wipeLocations: true,
        wipeMaintenance: true,
      });

      await expect(page.locator("[role='status']").first()).toBeVisible();
    });

    test.skip("blocks wipe attempts from non-owners", async ({ page }) => {
      await page.route(WIPE_ROUTE, route => {
        route.fulfill({
          status: 403,
          contentType: "application/json",
          body: JSON.stringify({ message: "forbidden" }),
        });
      });

      await openWipeInventory(page);

      const requestPromise = page.waitForRequest(WIPE_ROUTE);
      await page.getByRole("button", { name: "Confirm" }).last().click();
      await requestPromise;

      await expect(page.getByText("Failed to wipe inventory.")).toBeVisible();
    });

    const checkboxCases = [
      {
        name: "tags only",
        selection: { tags: true, locations: false, maintenance: false },
      },
      {
        name: "locations only",
        selection: { tags: false, locations: true, maintenance: false },
      },
      {
        name: "maintenance only",
        selection: { tags: false, locations: false, maintenance: true },
      },
    ];

    for (const scenario of checkboxCases) {
      test.skip(`submits correct flags when ${scenario.name} is selected`, async ({ page }) => {
        await page.route(WIPE_ROUTE, route => {
          route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ completed: 0 }) });
        });

        await openWipeInventory(page);
        await expect(page.getByText("Wipe Inventory").first()).toBeVisible();

        const tags = page.locator("input#wipe-tags-checkbox");
        const locations = page.locator("input#wipe-locations-checkbox");
        const maintenance = page.locator("input#wipe-maintenance-checkbox");

        if (scenario.selection.tags) {
          await tags.check();
        } else {
          await tags.uncheck();
        }

        if (scenario.selection.locations) {
          await locations.check();
        } else {
          await locations.uncheck();
        }

        if (scenario.selection.maintenance) {
          await maintenance.check();
        } else {
          await maintenance.uncheck();
        }

        const requestPromise = page.waitForRequest(WIPE_ROUTE);
        await page.getByRole("button", { name: "Confirm" }).last().click();
        const request = await requestPromise;

        expect(request.postDataJSON()).toEqual({
          wipeTags: scenario.selection.tags,
          wipeLocations: scenario.selection.locations,
          wipeMaintenance: scenario.selection.maintenance,
        });
      });
    }
  });
});
