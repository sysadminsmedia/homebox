import { expect, test } from "@playwright/test";

// the PWA service worker would let WebKit requests bypass page.route
test.use({ serviceWorkers: "block" });

test("failed CSV import only shows the error toast", async ({ page }) => {
  await page.route("**/api/v1/entities/import", route =>
    route.fulfill({ status: 500, contentType: "application/json", body: JSON.stringify({ error: "import failed" }) })
  );

  await page.goto("/home");
  await expect(page).toHaveURL("/");
  await page.fill("input[type='text']", "demo@example.com");
  await page.fill("input[type='password']", "demodemo");
  await page.click("button[type='submit']");
  await expect(page).toHaveURL("/home");

  await page.goto("/collection/tools");
  await page.getByRole("button", { name: "Import Inventory" }).click();
  await page.locator("input[type='file'][accept='.csv,.tsv']").setInputFiles({
    name: "items.csv",
    mimeType: "text/csv",
    buffer: Buffer.from("HB.location,HB.name\nGarage,Drill\n"),
  });
  await page.getByRole("button", { name: "Submit" }).click();

  await expect(page.getByText("Import failed. Please try again later.")).toBeVisible();
  // count once instead of toHaveCount(0), which would pass as soon as a stray success toast times out
  await page.waitForTimeout(500);
  expect(await page.getByText("Import successful!").count()).toBe(0);
});
