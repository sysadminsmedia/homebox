import { test, expect } from "@playwright/test";

test("valid login", async ({ page }) => {
  await page.goto("http://localhost:3000/home");
  await expect(page).toHaveURL("/");
  await page.fill("input[type='text']", "demo@example.com");
  await page.fill("input[placeholder='Password']", "demo");
  await page.click("button[type='submit']");
  await expect(page).toHaveURL("/home");
});

test("invalid login", async ({ page }) => {
  await page.goto("http://localhost:3000/home");
  await expect(page).toHaveURL("/");
  await page.fill("input[type='text']", "dummy@example.com");
  await page.fill("input[placeholder='Password']", "dummy");
  await page.click("button[type='submit']");
  await expect(page.locator("div[class*='top-2']")).toHaveText("Invalid email or password");
  await expect(page).toHaveURL("/");
});
