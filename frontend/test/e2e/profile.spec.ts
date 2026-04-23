import { expect, test, type Page } from "@playwright/test";
import { faker } from "@faker-js/faker";
import { registerAndLogin, STRONG_PASSWORD } from "./helpers/auth";

const ANOTHER_STRONG_PASSWORD = "AnotherVeryStrongPass123!";

async function gotoProfile(page: Page) {
  await page.goto("/profile");
  await expect(page).toHaveURL("/profile");
  await expect(page.getByRole("heading", { name: "User Profile", exact: true })).toBeVisible();
}

test.describe("profile page", () => {
  test.beforeEach(async ({ page }) => {
    test.slow();
    await registerAndLogin(page);
  });

  test("displays user profile details", async ({ page }) => {
    await gotoProfile(page);

    await expect(page.getByText("Test User").first()).toBeVisible();
    await expect(page.locator("dd").first()).toBeVisible();
  });

  test("change password: wrong current password shows error toast", async ({ page }) => {
    await gotoProfile(page);

    await page.getByRole("button", { name: "Change Password" }).first().click();

    const dialog = page.getByRole("dialog");
    await expect(dialog).toBeVisible();

    const passwordInputs = dialog.locator("input[type='password']");
    await passwordInputs.nth(0).fill("this-is-not-my-password");
    await passwordInputs.nth(1).fill(ANOTHER_STRONG_PASSWORD);

    const submit = dialog.getByRole("button", { name: "Submit" });
    await expect(submit).toBeEnabled();
    await submit.click();

    await expect(page.getByText("Failed to change password.")).toBeVisible();
  });

  test.skip("change password: matching current password succeeds", async ({ page }) => {
    await gotoProfile(page);

    await page.getByRole("button", { name: "Change Password" }).first().click();

    const dialog = page.getByRole("dialog");
    await expect(dialog).toBeVisible();

    const passwordInputs = dialog.locator("input[type='password']");
    await passwordInputs.nth(0).fill(STRONG_PASSWORD);
    await passwordInputs.nth(1).fill(ANOTHER_STRONG_PASSWORD);

    const submit = dialog.getByRole("button", { name: "Submit" });
    await expect(submit).toBeEnabled();
    const responsePromise = page.waitForResponse(r => r.url().includes("/users/self/change-password"));
    await submit.click();
    const resp = await responsePromise;
    expect(resp.status()).toBeLessThan(400);
    await expect(dialog).toBeHidden();
  });

  test("theme switcher: picking a theme persists across reload", async ({ page }) => {
    await gotoProfile(page);

    await page.locator("[data-set-theme='night']").first().click();
    await expect(page.locator("html")).toHaveAttribute("data-theme", "night");

    await page.reload();
    await expect(page).toHaveURL("/profile");
    await expect(page.locator("html")).toHaveAttribute("data-theme", "night");

    await page.locator("[data-set-theme='cupcake']").first().click();
    await expect(page.locator("html")).toHaveAttribute("data-theme", "cupcake");

    await page.reload();
    await expect(page).toHaveURL("/profile");
    await expect(page.locator("html")).toHaveAttribute("data-theme", "cupcake");

    await page.locator("[data-set-theme='homebox']").first().click();
    await expect(page.locator("html")).toHaveAttribute("data-theme", "homebox");
  });

  test("language switcher changes the active locale", async ({ page }) => {
    await gotoProfile(page);

    await page.getByRole("combobox").filter({ hasText: /English/ }).first().click();
    const listbox = page.getByRole("listbox");
    await expect(listbox).toBeVisible();
    await listbox
      .getByRole("option", { name: /Deutsch/i })
      .first()
      .click();

    await expect(page.locator("html")).not.toHaveAttribute("lang", "en");
  });

  test("duplicate-item settings dialog toggles persist", async ({ page }) => {
    await gotoProfile(page);

    await page.getByRole("button", { name: "Duplicate Settings" }).first().click();
    const dialog = page.getByRole("dialog");
    await expect(dialog).toBeVisible();

    const switchIds = ["#copy-maintenance", "#copy-attachments", "#copy-custom-fields"] as const;
    // reka-ui's Switch uses `data-state="checked"` or `"unchecked"`. Wait for
    // that attribute to settle into a real value before capturing — a raw
    // getAttribute() doesn't auto-retry, so under hydration load it can return
    // null/"" and the later not.toHaveAttribute(…, "") check would pass
    // vacuously even if the click was a no-op.
    const switchesWithInitial = await Promise.all(
      switchIds.map(async id => {
        const locator = dialog.locator(id);
        await expect(locator).toHaveAttribute("data-state", /^(checked|unchecked)$/);
        const initial = (await locator.getAttribute("data-state")) ?? "";
        expect(initial).toMatch(/^(checked|unchecked)$/);
        return { id, locator, initial };
      })
    );

    for (const { locator } of switchesWithInitial) {
      await locator.click();
    }
    for (const { locator, initial } of switchesWithInitial) {
      await expect(locator).not.toHaveAttribute("data-state", initial);
    }

    await page.keyboard.press("Escape");
    await expect(dialog).toBeHidden();

    await page.getByRole("button", { name: "Duplicate Settings" }).first().click();
    const reopened = page.getByRole("dialog");
    await expect(reopened).toBeVisible();

    for (const { id, initial } of switchesWithInitial) {
      await expect(reopened.locator(id)).not.toHaveAttribute("data-state", initial);
    }
  });

  test("delete account dialog: cancel path does not delete", async ({ page }) => {
    await gotoProfile(page);

    // "Delete Account" also matches a heading, so scope to the button role.
    await page.getByRole("button", { name: "Delete Account" }).click();

    const alert = page.getByRole("alertdialog");
    await expect(alert).toBeVisible();
    await expect(alert).toContainText(/are you sure/i);

    await alert.getByRole("button", { name: "Cancel" }).click();
    await expect(alert).toBeHidden();

    await expect(page).toHaveURL("/profile");
    await page.goto("/home");
    await expect(page).toHaveURL("/home");
  });

  test("legacy header toggle button updates its label", async ({ page }) => {
    await gotoProfile(page);

    const legacyBtn = page.getByRole("button", { name: /Legacy Header/i });
    // Make sure the button is hydrated with a non-empty label before capturing
    // it — otherwise textContent() can return null / "" and the later
    // not.toHaveText(initialLabel) would pass vacuously.
    await expect(legacyBtn).toHaveText(/.+/);
    const initialLabel = (await legacyBtn.textContent())?.trim() ?? "";
    expect(initialLabel.length).toBeGreaterThan(0);

    await legacyBtn.click();
    await expect(legacyBtn).not.toHaveText(initialLabel);

    // Toggle back to avoid polluting shared state.
    await legacyBtn.click();
    await expect(legacyBtn).toHaveText(initialLabel);
  });
});
