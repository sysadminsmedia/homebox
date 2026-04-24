import { expect, test } from "@playwright/test";
import { faker } from "@faker-js/faker";
import { apiCreateLocation, registerAndLogin } from "./helpers/auth";

test.describe("locations tree view", () => {
  test.beforeEach(async ({ page }) => {
    await registerAndLogin(page);
  });

  test("create parent and child location; tree shows both", async ({ page }) => {
    test.slow();

    const parentName = `Parent-${faker.string.alphanumeric(8)}`;
    const childName = `Child-${faker.string.alphanumeric(8)}`;

    const parent = await apiCreateLocation(page.request, parentName);
    await apiCreateLocation(page.request, childName, parent.id);

    await page.goto("/locations");

    const parentNode = page.getByTestId(`location-tree-node-${parentName}`);
    const childNode = page.getByTestId(`location-tree-node-${childName}`);

    await expect(parentNode).toBeVisible();
    await expect(childNode).toHaveCount(0);

    // Click the chevron area of the toggle row (avoids the NuxtLink which uses
    // @click.stop and would navigate instead of toggle).
    const toggleRow = page.getByTestId(`location-tree-toggle-${parentName}`);
    const chevron = toggleRow.locator("[data-swap]").first();
    await chevron.click();
    await expect(childNode).toBeVisible();

    await chevron.click();
    await expect(childNode).toHaveCount(0);
  });

  test("expand-all and collapse-all buttons toggle every node", async ({ page }) => {
    test.slow();

    const parentName = `P-${faker.string.alphanumeric(8)}`;
    const childName = `C-${faker.string.alphanumeric(8)}`;

    const parent = await apiCreateLocation(page.request, parentName);
    await apiCreateLocation(page.request, childName, parent.id);

    await page.goto("/locations");

    const parentNode = page.getByTestId(`location-tree-node-${parentName}`);
    const childNode = page.getByTestId(`location-tree-node-${childName}`);
    await expect(parentNode).toBeVisible();
    await expect(childNode).toHaveCount(0);

    await page.getByTestId("location-tree-expand-all").click();
    await expect(childNode).toBeVisible();

    await page.getByTestId("location-tree-collapse-all").click();
    await expect(childNode).toHaveCount(0);
  });

  test("show-items toggle button can be clicked without breaking the tree", async ({ page }) => {
    const parentName = `Pkg-${faker.string.alphanumeric(8)}`;
    await apiCreateLocation(page.request, parentName);

    await page.goto("/locations");

    await expect(page.getByTestId(`location-tree-node-${parentName}`)).toBeVisible();

    const toggle = page.getByTestId("location-tree-toggle-items");
    await expect(toggle).toBeVisible();

    await toggle.click();
    await expect(page).toHaveURL(/showItems=false/);

    await toggle.click();
    await expect(page).toHaveURL(/showItems=true/);

    await expect(page.getByTestId(`location-tree-node-${parentName}`)).toBeVisible();
  });

  test("clicking a location link navigates to its detail page", async ({ page }) => {
    const parentName = `Nav-${faker.string.alphanumeric(8)}`;
    await apiCreateLocation(page.request, parentName);

    await page.goto("/locations");

    const link = page.getByTestId(`location-tree-link-${parentName}`);
    await expect(link).toBeVisible();
    await link.click();

    await page.waitForURL(/\/location\/[0-9a-f-]+$/);
    await expect(page.getByTestId("location-detail-name")).toHaveText(parentName);
  });

  test("breadcrumb on location detail page shows the parent chain", async ({ page }) => {
    test.slow();

    const parentName = `Bread-${faker.string.alphanumeric(8)}`;
    const childName = `Crumb-${faker.string.alphanumeric(8)}`;

    const parent = await apiCreateLocation(page.request, parentName);
    const child = await apiCreateLocation(page.request, childName, parent.id);

    await page.goto(`/location/${child.id}`);

    await expect(page.getByTestId("location-detail-name")).toHaveText(childName);

    const breadcrumb = page.getByTestId("location-breadcrumb");
    await expect(breadcrumb).toBeVisible();
    await expect(breadcrumb).toContainText(parentName);

    const parentLink = breadcrumb.getByRole("link", { name: parentName });
    await parentLink.click();

    await page.waitForURL(/\/location\/[0-9a-f-]+$/);
    await expect(page.getByTestId("location-detail-name")).toHaveText(parentName);
  });
});
