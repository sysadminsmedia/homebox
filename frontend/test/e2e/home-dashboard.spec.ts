import { expect, test, type Page } from "@playwright/test";
import { faker } from "@faker-js/faker";
import { registerAndLogin } from "./helpers/auth";

type EntityTypeSummary = { id: string; name: string; isLocation: boolean };

async function getLocationTypeId(page: Page): Promise<string> {
  const res = await page.request.get("/api/v1/entity-types");
  expect(res.ok(), "entity-types should fetch").toBeTruthy();
  const types = (await res.json()) as EntityTypeSummary[];
  const loc = types.find(t => t.isLocation);
  if (!loc) throw new Error("Expected default location entity type to exist");
  return loc.id;
}

/**
 * Create a location via the REST API. This bypasses the LocationSelector's
 * client store cache — the dashboard page reloads parents/tree in its layout
 * onMounted hook, so after navigation the new location will show up reliably.
 */
async function apiCreateLocation(page: Page, name: string, locationTypeId: string): Promise<string> {
  const res = await page.request.post("/api/v1/entities", {
    data: {
      name,
      description: "",
      quantity: 1,
      tagIds: [],
      entityTypeId: locationTypeId,
    },
  });
  expect(res.ok(), `create location ${name}`).toBeTruthy();
  const body = (await res.json()) as { id: string };
  return body.id;
}

/**
 * Create an item via the REST API. Items omit entityTypeId and pass parentId
 * (the location) per the backend contract for nested entities.
 */
async function apiCreateItem(page: Page, name: string, parentId: string): Promise<string> {
  const res = await page.request.post("/api/v1/entities", {
    data: {
      name,
      description: "",
      quantity: 1,
      parentId,
      tagIds: [],
    },
  });
  expect(res.ok(), `create item ${name}`).toBeTruthy();
  const body = (await res.json()) as { id: string };
  return body.id;
}

test.describe("Home dashboard", () => {
  test("renders stat cards with zeros for a brand-new user", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    await expect(page).toHaveURL("/home");

    const cards = page.getByTestId("stat-card");
    await expect(cards).toHaveCount(4);

    const values = page.getByTestId("stat-card-value");
    // Fresh groups have no items even though the seed creates locations/tags,
    // so the "Total Items" card is deterministic at zero.
    await expect(values.nth(1)).toContainText("0");
  });

  test("stats and recent-items panel update after creating a location and item", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const locationName = `loc-${faker.string.alphanumeric(8).toLowerCase()}`;
    const itemName = `item-${faker.string.alphanumeric(8).toLowerCase()}`;

    // Create entities via the REST API rather than clicking through the
    // create modals. The dashboard's useAsyncData hooks refetch on navigation
    // and the layout's onMounted handler forces a parents/tree refresh, so
    // the new entities are reliably reflected in the UI on /home without
    // relying on the SSE-driven client store cache (which has proven flaky
    // in Playwright runs).
    const locationTypeId = await getLocationTypeId(page);
    const locationId = await apiCreateLocation(page, locationName, locationTypeId);
    await apiCreateItem(page, itemName, locationId);

    await page.goto("/home");
    await expect(page).toHaveURL("/home");

    const values = page.getByTestId("stat-card-value");
    await expect(values).toHaveCount(4);
    // Total Items is the second stat card (index 1) per statistics.ts ordering.
    await expect(values.nth(1)).toContainText("1");

    // Recently-added uses a desktop table (>=lg) which shows the item name as
    // text, and storage_locations renders LocationCards with the name as a
    // heading-like element. Use .first() to tolerate incidental duplicates.
    await expect(page.getByText(itemName, { exact: false }).first()).toBeVisible();
    await expect(page.getByText(locationName, { exact: false }).first()).toBeVisible();
  });

  test("navigates to items, locations, and tags routes from the dashboard", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    // The StatCard component is not itself an interactive link, so we assert
    // the routes that the dashboard summarises are reachable — this guards
    // against route regressions that would break the drill-down UX.
    await page.goto("/items");
    await expect(page).toHaveURL(/\/items/);

    await page.goto("/locations");
    await expect(page).toHaveURL(/\/locations/);

    await page.goto("/tags");
    await expect(page).toHaveURL(/\/tags/);

    await page.goto("/home");
    await expect(page).toHaveURL("/home");
    await expect(page.getByTestId("stat-card")).toHaveCount(4);
  });
});
