import { test, expect } from "@playwright/test";

test("valid login", async ({ page }) => {
  await page.goto("/home");
  await expect(page).toHaveURL("/");
  await page.fill("input[type='text']", "demo@example.com");
  await page.fill("input[placeholder='Password']", "demo");
  await page.click("button[type='submit']");
  await expect(page).toHaveURL("/home");
});

test("invalid login", async ({ page }) => {
  await page.goto("/home");
  await expect(page).toHaveURL("/");
  await page.fill("input[type='text']", "dummy@example.com");
  await page.fill("input[placeholder='Password']", "dummy");
  await page.click("button[type='submit']");
  await expect(page.locator("div[class*='top-2']")).toHaveText("Invalid email or password");
  await expect(page).toHaveURL("/");
});

test("registration", async ({ page }) => {
  test.slow();
  // Register a new user
  await page.goto("/home");
  await expect(page).toHaveURL("/");
  await page.click("button[class$='btn-wide']");
  await page.fill(
    "html > body > div:nth-of-type(1) > div > div:nth-of-type(2) > div:nth-of-type(2) > div > div > form > div > div > div:nth-of-type(1) > input",
    "test@example.com"
  );
  await page.fill(
    "html > body > div:nth-of-type(1) > div > div:nth-of-type(2) > div:nth-of-type(2) > div > div > form > div > div > div:nth-of-type(2) > input",
    "Test User"
  );
  await page.fill("input[placeholder='Password']", "ThisIsAStrongDemoPass");
  await page.click("button[class$='mt-2']");
  await expect(page).toHaveURL("/");

  // Try to register the same user again (it should fail)
  await page.goto("/home");
  await expect(page).toHaveURL("/");
  await page.click("button[class$='btn-wide']");
  await page.fill(
    "html > body > div:nth-of-type(1) > div > div:nth-of-type(2) > div:nth-of-type(2) > div > div > form > div > div > div:nth-of-type(1) > input",
    "test@example.com"
  );
  await page.fill(
    "html > body > div:nth-of-type(1) > div > div:nth-of-type(2) > div:nth-of-type(2) > div > div > form > div > div > div:nth-of-type(2) > input",
    "Test User"
  );
  await page.fill("input[placeholder='Password']", "ThisIsAStrongDemoPass");
  await page.click("button[class$='mt-2']");
  await expect(page).toHaveURL("/");
  await expect(page.locator("div[class*='top-2']")).toHaveText("Problem registering user");
});
