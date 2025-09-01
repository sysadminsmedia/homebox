import { expect, test } from "@playwright/test";

test("valid login", async ({ page }) => {
  await page.goto("/home");
  await page.waitForTimeout(1000); // Wait for vue to load
  await expect(page).toHaveURL("/");
  await page.fill("input[type='text']", "demo@example.com");
  await page.fill("input[type='password']", "demo");
  await page.click("button[type='submit']");
  await expect(page).toHaveURL("/home");
});

test("invalid login", async ({ page }) => {
  await page.goto("/home");
  await expect(page).toHaveURL("/");
  await page.fill("input[type='text']", "dummy@example.com");
  await page.fill("input[type='password']", "dummy");
  await page.click("button[type='submit']");
  await page.waitForTimeout(500);
  await expect(page.locator("div[class*='login-error']").first()).toHaveText("Invalid email or password");
  await expect(page).toHaveURL("/");
});

test("registration", async ({ page }) => {
  test.slow();
  // Register a new user
  await page.goto("/home");
  await expect(page).toHaveURL("/");
  await page.getByTestId("register-button").click();

  await page.waitForTimeout(1000);

  await page.getByTestId("email-input").locator("input").fill("test@example.com");
  await page.getByTestId("name-input").locator("input").fill("Test User");
  await page.getByTestId("password-input").locator("input").fill("ThisIsAStrongDemoPass");
  await page.getByTestId("confirm-register-button").click();
  await expect(page).toHaveURL("/");

  // Try to register the same user again (it should fail)
  await page.goto("/home");
  await expect(page).toHaveURL("/");
  await page.getByTestId("register-button").click();
  await page.getByTestId("email-input").locator("input").fill("test@example.com");
  await page.getByTestId("name-input").locator("input").fill("Test User");
  await page.getByTestId("password-input").locator("input").fill("ThisIsAStrongDemoPass");
  await page.getByTestId("confirm-register-button").click();
  await expect(page).toHaveURL("/");
  await page.waitForTimeout(500);
  await expect(page.locator("div[class*='login-error']").first()).toHaveText("Problem registering user");
});
