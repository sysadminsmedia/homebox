import { expect, test, type Locator, type Page } from "@playwright/test";
import { faker } from "@faker-js/faker";
import { registerAndLogin } from "./helpers/auth";

type EntityTypeSummary = { id: string; name: string; isLocation: boolean };

async function getEntityTypes(page: Page): Promise<EntityTypeSummary[]> {
  const res = await page.request.get("/api/v1/entity-types");
  expect(res.ok()).toBeTruthy();
  return (await res.json()) as EntityTypeSummary[];
}

async function createLocation(page: Page, name: string, locationTypeId: string): Promise<{ id: string; name: string }> {
  const res = await page.request.post("/api/v1/entities", {
    data: {
      name,
      description: "",
      quantity: 1,
      tagIds: [],
      entityTypeId: locationTypeId,
    },
  });
  expect(res.ok()).toBeTruthy();
  const body = (await res.json()) as { id: string; name: string };
  return { id: body.id, name: body.name };
}

async function createItem(
  page: Page,
  params: { name: string; parentId: string; itemTypeId?: string; tagIds?: string[] }
): Promise<{ id: string; name: string }> {
  const res = await page.request.post("/api/v1/entities", {
    data: {
      name: params.name,
      description: "",
      quantity: 1,
      parentId: params.parentId,
      tagIds: params.tagIds ?? [],
    },
  });
  expect(res.ok()).toBeTruthy();
  const body = (await res.json()) as { id: string; name: string };
  return { id: body.id, name: body.name };
}

async function createTag(page: Page, name: string): Promise<{ id: string; name: string }> {
  const res = await page.request.post("/api/v1/tags", {
    data: {
      name,
      description: "",
      color: "",
      icon: "",
    },
  });
  expect(res.ok()).toBeTruthy();
  const body = (await res.json()) as { id: string; name: string };
  return { id: body.id, name: body.name };
}

async function pickEntityTypeIds(page: Page) {
  const types = await getEntityTypes(page);
  const locType = types.find(t => t.isLocation);
  if (!locType) {
    throw new Error("Expected default location entity type to exist");
  }
  const itemType = types.find(t => !t.isLocation);
  return { locType, itemType };
}

// Filter.vue renders "<label> (n)" when count > 0, or just "<label>" initially.
function filterButton(page: Page, label: "Locations" | "Tags") {
  return page.getByRole("button", { name: new RegExp(`^${label}( \\(\\d+\\))?$`) });
}

function openPopover(page: Page): Locator {
  return page.locator("[role='dialog'], [data-reka-popper-content-wrapper]").last();
}

async function togglePopoverOption(popover: Locator, name: string) {
  await popover.getByText(name, { exact: true }).first().click();
}

async function waitForQueryParam(page: Page, key: string, expected: RegExp) {
  await expect
    .poll(
      () => {
        const url = new URL(page.url());
        return url.searchParams.getAll(key).join(",");
      },
      { timeout: 5000 }
    )
    .toMatch(expected);
}

test.describe("Items search, filter, sort, pagination", () => {
  test("text search q param drives results and persists on reload", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const { locType } = await pickEntityTypeIds(page);
    const locName = `loc-${faker.string.alphanumeric(6).toLowerCase()}`;
    const location = await createLocation(page, locName, locType.id);

    const needle = `needle${faker.string.alphanumeric(8).toLowerCase()}`;
    await createItem(page, { name: `${needle} widget`, parentId: location.id });
    await createItem(page, { name: "unrelated thing", parentId: location.id });
    await createItem(page, { name: "another thing", parentId: location.id });

    await page.goto("/items");
    // Wait for the initial items list to render before interacting with the search
    // input — the Items page defers first search until reactive watchers fire.
    await expect(page.getByText("unrelated thing", { exact: true }).first()).toBeVisible();

    await page.locator("main input:not([type]), main input[type='text']").first().fill(needle);

    await waitForQueryParam(page, "q", new RegExp(`^${needle}$`));
    await expect(page.getByText(`${needle} widget`, { exact: false })).toBeVisible();
    await expect(page.getByText("unrelated thing", { exact: true })).toHaveCount(0);

    await page.reload();
    await expect(page).toHaveURL(new RegExp(`[?&]q=${needle}`));
    await expect(page.locator("main input:not([type]), main input[type='text']").first()).toHaveValue(needle);
    await expect(page.getByText(`${needle} widget`, { exact: false })).toBeVisible();
  });

  test("filter by a single location shows only items in that location", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const { locType } = await pickEntityTypeIds(page);
    const suffix = faker.string.alphanumeric(6).toLowerCase();
    const locA = await createLocation(page, `locA-${suffix}`, locType.id);
    const locB = await createLocation(page, `locB-${suffix}`, locType.id);

    const inA = `inA-${suffix}`;
    const inB = `inB-${suffix}`;
    await createItem(page, { name: inA, parentId: locA.id });
    await createItem(page, { name: inB, parentId: locB.id });

    await page.goto("/items");
    await filterButton(page, "Locations").click();

    const popover = openPopover(page);
    await togglePopoverOption(popover, locA.name);
    await page.keyboard.press("Escape");

    await waitForQueryParam(page, "loc", /.+/);
    await expect(page.getByText(inA, { exact: true }).first()).toBeVisible();
    await expect(page.getByText(inB, { exact: true })).toHaveCount(0);
  });

  test("filter by multiple locations ORs results", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const { locType } = await pickEntityTypeIds(page);
    const suffix = faker.string.alphanumeric(6).toLowerCase();
    const locA = await createLocation(page, `locA-${suffix}`, locType.id);
    const locB = await createLocation(page, `locB-${suffix}`, locType.id);
    const locC = await createLocation(page, `locC-${suffix}`, locType.id);

    const inA = `inA-${suffix}`;
    const inB = `inB-${suffix}`;
    const inC = `inC-${suffix}`;
    await createItem(page, { name: inA, parentId: locA.id });
    await createItem(page, { name: inB, parentId: locB.id });
    await createItem(page, { name: inC, parentId: locC.id });

    await page.goto("/items");
    await filterButton(page, "Locations").click();

    const popover = openPopover(page);
    await togglePopoverOption(popover, locA.name);
    await togglePopoverOption(popover, locB.name);
    await page.keyboard.press("Escape");

    await expect.poll(() => new URL(page.url()).searchParams.getAll("loc").length).toBe(2);

    await expect(page.getByText(inA, { exact: true }).first()).toBeVisible();
    await expect(page.getByText(inB, { exact: true }).first()).toBeVisible();
    await expect(page.getByText(inC, { exact: true })).toHaveCount(0);
  });

  test("filter by tag narrows results", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const { locType } = await pickEntityTypeIds(page);
    const suffix = faker.string.alphanumeric(6).toLowerCase();
    const location = await createLocation(page, `loc-${suffix}`, locType.id);
    const tag = await createTag(page, `tag-${suffix}`);

    const tagged = `tagged-${suffix}`;
    const untagged = `untagged-${suffix}`;
    await createItem(page, {
      name: tagged,
      parentId: location.id,
      itemTypeId: "",
      tagIds: [tag.id],
    });
    await createItem(page, { name: untagged, parentId: location.id });

    await page.goto("/items");
    await filterButton(page, "Tags").click();
    const popover = openPopover(page);
    await togglePopoverOption(popover, tag.name);
    await page.keyboard.press("Escape");

    await waitForQueryParam(page, "tag", /.+/);
    await expect(page.getByText(tagged, { exact: true }).first()).toBeVisible();
    await expect(page.getByText(untagged, { exact: true })).toHaveCount(0);
  });

  test("sort order toggle via Options updates orderBy query param", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const { locType } = await pickEntityTypeIds(page);
    const suffix = faker.string.alphanumeric(6).toLowerCase();
    const location = await createLocation(page, `loc-${suffix}`, locType.id);

    await Promise.all([
      createItem(page, { name: `alpha-${suffix}`, parentId: location.id }),
      createItem(page, { name: `beta-${suffix}`, parentId: location.id }),
    ]);

    await page.goto("/items");
    await page.getByRole("button", { name: "Options", exact: true }).click();

    const popover = openPopover(page);
    await popover.getByRole("combobox").click();
    await page.getByRole("option", { name: "Created At" }).click();

    await waitForQueryParam(page, "orderBy", /^createdAt$/);
    await expect(page).toHaveURL(/[?&]orderBy=createdAt/);

    await page.reload();
    await expect(page).toHaveURL(/[?&]orderBy=createdAt/);
  });

  test("pagination shows page controls once results exceed one page", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const { locType } = await pickEntityTypeIds(page);
    const suffix = faker.string.alphanumeric(6).toLowerCase();
    const location = await createLocation(page, `loc-${suffix}`, locType.id);

    // Default page size is 10; create 12 items so we have 2 pages.
    await Promise.all(
      Array.from({ length: 12 }, (_, i) =>
        createItem(page, {
          name: `p-${suffix}-${i.toString().padStart(2, "0")}`,
          parentId: location.id,
          itemTypeId: "",
        })
      )
    );

    await page.goto("/items");

    // Wait for the results count to render; the Items page only displays the
    // pagination row once an initial search response has populated `total`.
    await expect(page.getByText(/12\s+Results/i).first()).toBeVisible();
    await expect(page.getByText(/Page\s+1\s+of\s+2/i).first()).toBeVisible();

    // Numbered pagination buttons expose their accessible name as "Page N", not "N".
    await page.getByRole("button", { name: "Page 2", exact: true }).first().click();

    await waitForQueryParam(page, "page", /^2$/);
    await expect(page.getByText(/Page\s+2\s+of\s+2/i).first()).toBeVisible();
  });

  test("URL query params persist filter+sort state across reload", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);

    const { locType } = await pickEntityTypeIds(page);
    const suffix = faker.string.alphanumeric(6).toLowerCase();
    const location = await createLocation(page, `loc-${suffix}`, locType.id);
    const tag = await createTag(page, `tag-${suffix}`);

    const q = `persist-${suffix}`;
    await Promise.all([
      createItem(page, {
        name: q,
        parentId: location.id,
        itemTypeId: "",
        tagIds: [tag.id],
      }),
      createItem(page, { name: `other-${suffix}`, parentId: location.id }),
    ]);

    const url = `/items?q=${encodeURIComponent(q)}&loc=${location.id}&tag=${tag.id}&orderBy=createdAt`;
    await page.goto(url);

    // Wait for the input to be populated from the URL ?q= param. The Items page
    // mirrors ?q= into a reactive ref during onMounted, so use an auto-retry
    // assertion rather than a synchronous check.
    await expect(page.locator("main input:not([type]), main input[type='text']").first()).toHaveValue(q);
    // Filter badge count lags until the locations/tags stores finish their
    // async fetch on mount; the assertion itself auto-retries.
    await expect(filterButton(page, "Locations")).toHaveText(/\(1\)/);
    await expect(filterButton(page, "Tags")).toHaveText(/\(1\)/);
    await expect(page.getByText(q, { exact: true }).first()).toBeVisible();

    await page.reload();
    await expect(page).toHaveURL(new RegExp(`[?&]q=${q}`));
    await expect(page).toHaveURL(new RegExp(`[?&]loc=${location.id}`));
    await expect(page).toHaveURL(new RegExp(`[?&]tag=${tag.id}`));
    await expect(page).toHaveURL(/[?&]orderBy=createdAt/);
    await expect(page.locator("main input:not([type]), main input[type='text']").first()).toHaveValue(q);
  });
});
