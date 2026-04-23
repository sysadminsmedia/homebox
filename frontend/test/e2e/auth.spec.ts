import { expect, test, type Page } from "@playwright/test";
import { faker } from "@faker-js/faker";
import { registerAndLogin, STRONG_PASSWORD } from "./helpers/auth";

test.describe("Login validation", () => {
  test("submitting empty form does not navigate away from /", async ({ page }) => {
    await page.goto("/");
    await expect(page).toHaveURL("/");
    await page.click("button[type='submit']");
    await expect(page).toHaveURL("/");
  });

  test("whitespace-only credentials are rejected", async ({ page }) => {
    await page.goto("/");
    await expect(page).toHaveURL("/");
    await page.fill("input[type='text']", "   ");
    await page.fill("input[type='password']", "   ");
    await page.click("button[type='submit']");
    await page.waitForTimeout(500);
    await expect(page).not.toHaveURL("/home");
  });

  test("invalid credentials surface an error toast", async ({ page }) => {
    await page.goto("/");
    await expect(page).toHaveURL("/");
    await page.fill("input[type='text']", `nobody-${Date.now()}@example.com`);
    await page.fill("input[type='password']", "definitely-not-right");
    await page.click("button[type='submit']");
    await page.waitForTimeout(500);
    await expect(page.locator("div[class*='login-error']").first()).toHaveText("Invalid email or password");
    await expect(page).toHaveURL("/");
  });
});

test.describe("Registration", () => {
  test("disables submit until password meets strength requirements", async ({ page }) => {
    test.slow();
    await page.goto("/");
    await page.getByTestId("register-button").click();

    const email = faker.internet.email().toLowerCase();
    await page.getByTestId("email-input").locator("input").fill(email);
    await page.getByTestId("name-input").locator("input").fill("Weak Pass User");
    await page.getByTestId("password-input").locator("input").fill("short");

    await expect(page.getByText("Password Strength", { exact: false })).toBeVisible();
    await expect(page.getByTestId("confirm-register-button")).toBeDisabled();

    await page.getByTestId("password-input").locator("input").fill(STRONG_PASSWORD);
    await expect(page.getByTestId("confirm-register-button")).toBeEnabled();
  });

  test("duplicate email registration shows error toast", async ({ page }) => {
    test.slow();
    const email = faker.internet.email().toLowerCase();

    const firstRegister = (url: string) =>
      page.waitForResponse(r => r.url().includes("/api/v1/users/register") && r.request().method() === "POST");

    await page.goto("/");
    await page.getByTestId("register-button").click();
    await page.getByTestId("email-input").locator("input").fill(email);
    await page.getByTestId("name-input").locator("input").fill("Duplicate User");
    await page.getByTestId("password-input").locator("input").fill(STRONG_PASSWORD);
    const firstResp = firstRegister("first");
    await page.getByTestId("confirm-register-button").click();
    expect((await firstResp).status()).toBe(204);

    await page.goto("/");
    await page.getByTestId("register-button").click();
    await page.getByTestId("email-input").locator("input").fill(email);
    await page.getByTestId("name-input").locator("input").fill("Duplicate User");
    await page.getByTestId("password-input").locator("input").fill(STRONG_PASSWORD);
    const secondResp = firstRegister("second");
    await page.getByTestId("confirm-register-button").click();
    expect((await secondResp).status()).toBeGreaterThanOrEqual(400);
    await expect(page.getByText("Problem registering user").first()).toBeVisible();
  });

  test("rejects malformed email address", async ({ page }) => {
    test.slow();
    await page.goto("/");
    await page.getByTestId("register-button").click();
    await page.getByTestId("email-input").locator("input").fill("not-a-valid-email");
    await page.getByTestId("name-input").locator("input").fill("Bad Email User");
    await page.getByTestId("password-input").locator("input").fill(STRONG_PASSWORD);
    await page.getByTestId("confirm-register-button").click();
    await expect(page).toHaveURL("/");
  });
});

test.describe("Session lifecycle", () => {
  test("logout from /home returns user to login page", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);
    await expect(page).toHaveURL("/home");

    const logoutButton = page.getByTestId("logout-button");
    await expect(logoutButton).toBeVisible();
    await logoutButton.click();

    await expect(page).toHaveURL("/");
  });

  test("after logout, visiting a protected route redirects to login", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);
    await page.getByTestId("logout-button").click();
    await expect(page).toHaveURL("/");

    await page.goto("/home");
    await expect(page).toHaveURL("/");
  });

  test("visiting / while authenticated redirects to /home", async ({ page }) => {
    test.slow();
    await registerAndLogin(page);
    await expect(page).toHaveURL("/home");

    await page.goto("/");
    await expect(page).toHaveURL("/home");
  });
});

test.describe("Login page chrome", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await expect(page).toHaveURL("/");
  });

  test("tagline and HomeBox heading render", async ({ page }) => {
    await expect(page.getByText("Track, Organize, and Manage your Things.", { exact: false })).toBeVisible();
    await expect(page.getByRole("heading").first()).toBeVisible();
  });

  test("language selector is visible on the login page", async ({ page }) => {
    await expect(page.getByRole("combobox").filter({ hasText: /English/ })).toBeVisible();
  });

  test("social / external links render with expected href targets", async ({ page }) => {
    const hrefSubstrings = ["sysadminsmedia/homebox", "sysadminszone", "discord.gg", "homebox.software"];

    for (const substring of hrefSubstrings) {
      const anchor = page.locator(`a[href*="${substring}"]`).first();
      await expect(anchor).toBeVisible();
      await expect(anchor).toHaveAttribute("target", "_blank");
      await expect(anchor).toHaveAttribute("rel", /noopener/);
    }
  });

  test("register button toggles to the registration form and back", async ({ page }) => {
    const registerButton = page.getByTestId("register-button");
    await expect(registerButton).toBeVisible();

    await registerButton.click();
    await expect(page.getByTestId("email-input")).toBeVisible();
    await expect(page.getByTestId("name-input")).toBeVisible();
    await expect(page.getByTestId("password-input")).toBeVisible();
    await expect(page.getByTestId("confirm-register-button")).toBeVisible();

    await registerButton.click();
    await expect(page.getByTestId("name-input")).toHaveCount(0);
  });
});
