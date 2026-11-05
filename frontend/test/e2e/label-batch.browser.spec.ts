import type { Locator, Page } from "@playwright/test";
import { expect, test } from "@playwright/test";

const EMAIL = "demo@example.com";
const PASSWORD = process.env.E2E_PASSWORD || "demodemo";

type Direction = "Ascending" | "Descending";

// Logs in through the API, as the customs inventory spec does: WebKit on Linux drops the
// cookies of a form login.
async function login(page: Page) {
  const res = await page.request.post("/api/v1/users/login", {
    data: { username: EMAIL, password: PASSWORD, stayLoggedIn: false },
  });
  expect(res.ok()).toBeTruthy();
}

async function openBatchPage(page: Page) {
  await login(page);
  await page.goto("/label-templates/batch");
  await expect(rows(page).first()).toBeVisible();
}

function rows(page: Page): Locator {
  return page.locator("li[draggable]");
}

function isEntitiesRequest(url: URL, match: (params: URLSearchParams) => boolean): boolean {
  return url.pathname === "/api/v1/entities" && match(url.searchParams);
}

function entitiesResponse(page: Page, match: (params: URLSearchParams) => boolean) {
  return page.waitForResponse(res => isEntitiesRequest(new URL(res.url()), match));
}

const hasDirection = (direction: Direction) => (params: URLSearchParams) =>
  params.get("orderDirection") === (direction === "Ascending" ? "asc" : "desc");

async function pickDirection(page: Page, direction: Direction) {
  await page.getByRole("combobox", { name: "Order Direction" }).click();
  await page.getByRole("option", { name: direction }).click();
}

async function chooseDirection(page: Page, direction: Direction) {
  const response = entitiesResponse(page, hasDirection(direction));
  await pickDirection(page, direction);
  await response;
}

// The asset IDs shown in the rows, top to bottom, as numbers.
async function assetIds(page: Page): Promise<number[]> {
  const texts = await rows(page).allInnerTexts();
  return texts.flatMap(text => {
    const id = text.match(/#(\d+-\d+)/)?.[1];
    return id ? [Number(id.replace("-", ""))] : [];
  });
}

async function expectAssetIdOrder(page: Page, direction: Direction) {
  await expect
    .poll(async () => {
      const ids = await assetIds(page);
      const ordered = ids.every((id, i) => {
        if (i === 0) return true;
        const previous = ids[i - 1]!;
        return direction === "Ascending" ? previous < id : previous > id;
      });
      return ids.length > 2 && ordered;
    })
    .toBe(true);
}

async function rowName(page: Page, index: number): Promise<string> {
  return (await rows(page).nth(index).getByRole("checkbox").getAttribute("aria-label"))!;
}

async function tick(page: Page, index: number) {
  await rows(page).nth(index).getByRole("checkbox").click();
}

// Asserts the print order badge of the row showing the named item.
async function expectBadge(page: Page, name: string, badge: number) {
  const row = rows(page).filter({ has: page.getByRole("checkbox", { name, exact: true }) });
  await expect(row.getByText(String(badge), { exact: true })).toBeVisible();
}

test.describe("batch print list", () => {
  test("sorts the list and re-orders the selection on a sort change", async ({ page }) => {
    await openBatchPage(page);

    await expectAssetIdOrder(page, "Ascending");
    await expect(page.getByText(/^Showing the first/)).toHaveCount(0);

    await chooseDirection(page, "Descending");
    await expectAssetIdOrder(page, "Descending");

    await chooseDirection(page, "Ascending");
    await expectAssetIdOrder(page, "Ascending");

    // Tick against the list order, so the selection has an order of its own.
    const first = await rowName(page, 0);
    const second = await rowName(page, 1);
    await tick(page, 1);
    await tick(page, 0);
    await expectBadge(page, second, 1);
    await expectBadge(page, first, 2);

    // A search change leaves the selection order alone.
    const search = page.getByRole("textbox", { name: "Search" });
    const filtered = entitiesResponse(page, params => params.get("q") === first);
    await search.fill(first);
    await filtered;
    const cleared = entitiesResponse(page, params => !params.get("q"));
    await search.fill("");
    await cleared;
    await expectBadge(page, second, 1);
    await expectBadge(page, first, 2);

    // Descending puts the second row above the first, so this order already fits the list.
    await chooseDirection(page, "Descending");
    await expectBadge(page, second, 1);
    await expectBadge(page, first, 2);

    // Ascending doesn't fit it, so here the selection has to follow the list.
    await chooseDirection(page, "Ascending");
    await expectBadge(page, first, 1);
    await expectBadge(page, second, 2);

    await page.getByRole("group").getByRole("button", { name: "Locations" }).click();
    await expect(page.getByRole("combobox", { name: "Order By" })).toHaveCount(0);
    await expect(page.getByRole("combobox", { name: "Order Direction" })).toHaveCount(0);
  });

  test("ignores a stale response", async ({ page }) => {
    await openBatchPage(page);
    await expectAssetIdOrder(page, "Ascending");

    const first = await rowName(page, 0);
    const second = await rowName(page, 1);
    await tick(page, 1);
    await tick(page, 0);

    let release = () => {};
    const held = new Promise<void>(resolve => {
      release = resolve;
    });
    await page.route(
      url => isEntitiesRequest(url, hasDirection("Descending")),
      async route => {
        await held;
        await route.continue();
      }
    );

    await pickDirection(page, "Descending");
    await chooseDirection(page, "Ascending");
    await expectAssetIdOrder(page, "Ascending");
    await expectBadge(page, first, 1);

    const stale = entitiesResponse(page, hasDirection("Descending"));
    release();
    await stale;
    // Nothing should change, so there is no state to wait for: give a wrong update time to land.
    await page.waitForTimeout(250);

    await expectAssetIdOrder(page, "Ascending");
    await expectBadge(page, first, 1);
    await expectBadge(page, second, 2);
  });

  test("says when more items match than it shows", async ({ page }) => {
    await login(page);
    await page.route(
      url => isEntitiesRequest(url, () => true),
      async route => {
        const response = await route.fetch();
        const body = await response.json();
        await route.fulfill({ response, json: { ...body, total: body.items.length + 50 } });
      }
    );
    await page.goto("/label-templates/batch");

    await expect(page.getByText(/^Showing the first \d+ of \d+ items\./)).toBeVisible();
  });
});
