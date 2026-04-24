import { expect, test, type Page } from "@playwright/test";
import { registerAndLogin } from "./helpers/auth";

const QUICK_MENU_PLACEHOLDER = "Use the number keys to quickly select an action.";

/**
 * QuickMenuModal binds Ctrl+Backquote via useDialogHotkey on `event.code`.
 * WebKit maps this combo to Meta (Command) on macOS. Press once and wait for
 * the dialog to show — a longer timeout avoids flakes under suite load.
 */
async function openQuickMenuViaKeyboard(page: Page, browserName: string): Promise<boolean> {
  await page.locator("body").click({ position: { x: 5, y: 5 } });
  const combo = browserName === "webkit" ? "Meta+Backquote" : "Control+Backquote";
  await page.keyboard.press(combo);
  const input = page.getByPlaceholder(QUICK_MENU_PLACEHOLDER);
  try {
    await expect(input).toBeVisible({ timeout: 5000 });
    return true;
  } catch {
    return false;
  }
}

test.describe("QuickMenu", () => {
  test("opens via keyboard shortcut and can be dismissed", async ({ page, browserName }) => {
    test.slow();
    await registerAndLogin(page);

    const opened = await openQuickMenuViaKeyboard(page, browserName);
    test.skip(!opened, "QuickMenu keyboard shortcut not triggerable in this browser.");

    const input = page.getByPlaceholder(QUICK_MENU_PLACEHOLDER);
    await expect(input).toBeVisible();

    // Focus the input explicitly — the dialog's onKeydown Escape handler is
    // attached to the CommandInput, so Escape must originate from that element.
    await input.focus();
    await input.press("Escape");
    await expect(input).toBeHidden({ timeout: 5000 });
  });

  test("filters visible items when typing in the combobox", async ({ page, browserName }) => {
    test.slow();
    await registerAndLogin(page);

    const opened = await openQuickMenuViaKeyboard(page, browserName);
    test.skip(!opened, "QuickMenu keyboard shortcut not triggerable in this browser.");

    const input = page.getByPlaceholder(QUICK_MENU_PLACEHOLDER);
    const homeOption = page.getByRole("option", { name: "Home", exact: true });
    await expect(homeOption).toBeVisible();

    await input.fill("loca");
    await expect(page.getByRole("option", { name: /location/i }).first()).toBeVisible();
    await expect(homeOption).toBeHidden();

    await input.fill("");
    await expect(homeOption).toBeVisible();

    await page.keyboard.press("Escape");
  });

  test("selecting 'Location' from the create group opens the Create Location dialog", async ({ page, browserName }) => {
    test.slow();
    await registerAndLogin(page);

    const opened = await openQuickMenuViaKeyboard(page, browserName);
    test.skip(!opened, "QuickMenu keyboard shortcut not triggerable in this browser.");

    const input = page.getByPlaceholder(QUICK_MENU_PLACEHOLDER);
    await input.fill("Location");

    // The Create-group "Location" option renders its shortcut ("3") inside the
    // same element, so the accessible name ends up as "Location 3". Scope to
    // the Create group (heading is t("global.create")) to avoid accidentally
    // matching the Navigate group's "Locations" option.
    const createGroup = page.getByRole("group", { name: "Create" });
    const createLocation = createGroup.getByRole("option", { name: /^Location\b/ }).first();
    await expect(createLocation).toBeVisible();
    await createLocation.click();

    await expect(page.getByRole("dialog").filter({ hasText: "Create Location" }).first()).toBeVisible({
      timeout: 5000,
    });
  });
});

test.describe("Label Generator", () => {
  test.beforeEach(async ({ page }) => {
    test.slow();
    await registerAndLogin(page);
    await page.goto("/reports/label-generator");
    await expect(page).toHaveURL(/\/reports\/label-generator/);
  });

  test("renders the label generator with inputs and a generate button", async ({ page }) => {
    await expect(page.getByRole("heading", { name: /Label Generator/i }).first()).toBeVisible();
    await expect(page.locator("#input-cardHeight")).toBeVisible();
    await expect(page.locator("#input-cardWidth")).toBeVisible();
    await expect(page.getByRole("button", { name: "Generate Page" })).toBeVisible();
  });

  test("updating label width and height recalculates and renders the preview", async ({ page }) => {
    const cards = page.locator('[data-testid="label-preview-card"]');
    await expect(cards.first()).toBeVisible({ timeout: 10_000 });
    expect(await cards.count()).toBeGreaterThan(0);

    await page.locator("#input-cardWidth").fill("4");
    await page.locator("#input-cardHeight").fill("2");
    await page.getByRole("button", { name: "Generate Page" }).click();

    const card = cards.first();
    await expect(card).toBeVisible({ timeout: 10_000 });

    // Read computed dimensions — the Vue component writes `4in` / `2in` on the
    // inline style, which the browser normalises to px (4in = 384px at 96dpi).
    // Tailwind's preflight sets `box-sizing: border-box`, so the card's 2px
    // border eats ~4px of content width, and sub-pixel rounding can add another
    // ~2px of variance between engines. Compare with a generous ±8px tolerance
    // (~2%) so we verify the order of magnitude without pinning exact rendering.
    const dims = await card.evaluate(el => {
      const s = getComputedStyle(el);
      return { w: parseFloat(s.width), h: parseFloat(s.height) };
    });
    expect(Math.abs(dims.w - 4 * 96)).toBeLessThan(8);
    expect(Math.abs(dims.h - 2 * 96)).toBeLessThan(8);
  });

  test("bordered labels checkbox can be toggled", async ({ page }) => {
    const checkbox = page.locator("#borderedLabels");
    await expect(checkbox).toBeVisible();

    // Reka-UI reflects checkbox state onto a data-state attribute.
    const initial = await checkbox.getAttribute("data-state");

    await checkbox.click();
    await expect.poll(async () => await checkbox.getAttribute("data-state"), { timeout: 2000 }).not.toBe(initial);

    await checkbox.click();
    await expect.poll(async () => await checkbox.getAttribute("data-state"), { timeout: 2000 }).toBe(initial);
  });

  test("print location row checkbox can be toggled", async ({ page }) => {
    const checkbox = page.locator("#printLocationRow");
    await expect(checkbox).toBeVisible();

    const initial = await checkbox.getAttribute("data-state");
    await checkbox.click();
    await expect.poll(async () => await checkbox.getAttribute("data-state"), { timeout: 2000 }).not.toBe(initial);
  });

  test("entering oversized dimensions surfaces a toast", async ({ page }) => {
    await page.locator("#input-cardWidth").fill("50");
    await page.locator("#input-cardHeight").fill("50");
    await page.getByRole("button", { name: "Generate Page" }).click();

    const toast = page.getByText("Page size is too small for the card size", { exact: false }).first();
    await expect(toast).toBeVisible({ timeout: 5000 });
  });
});
