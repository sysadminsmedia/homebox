import { expect, test, type Page } from "@playwright/test";
import { faker } from "@faker-js/faker";
import { registerAndLogin, STRONG_PASSWORD } from "./helpers/auth";

const PRIMARY_LINKS = ["/home", "/locations", "/tags", "/items", "/templates", "/maintenance", "/profile"] as const;

const COLLECTION_LINKS = [
  "/collection/members",
  "/collection/invites",
  "/collection/notifiers",
  "/collection/settings",
  "/collection/entity-types",
  "/collection/tools",
] as const;

test.describe("navigation", () => {
  test.beforeEach(async ({ page }) => {
    test.slow();
    await registerAndLogin(page);
  });

  test("sidebar can be collapsed and expanded", async ({ page }) => {
    // The desktop sidebar wrapper carries data-state; the mobile Sheet variant does not.
    const sidebarWrapper = page.locator('div[data-state][data-variant="sidebar"]').first();
    const trigger = page.locator('[data-sidebar="trigger"]').first();

    await expect(sidebarWrapper).toHaveAttribute("data-state", "expanded");
    await trigger.click();
    await expect(sidebarWrapper).toHaveAttribute("data-state", "collapsed");
    await trigger.click();
    await expect(sidebarWrapper).toHaveAttribute("data-state", "expanded");
  });

  for (const href of PRIMARY_LINKS) {
    test(`primary nav link ${href} navigates correctly`, async ({ page }) => {
      // Scope to the sidebar so in-page anchors sharing the same href don't match.
      const nav = page.locator('[data-sidebar="sidebar"]').first();
      await nav.locator(`a[href="${href}"]`).first().click();
      await expect(page).toHaveURL(new RegExp(`${href}(\\?.*)?$`));
    });
  }

  for (const href of COLLECTION_LINKS) {
    test(`collection sub-link ${href} navigates correctly`, async ({ page }) => {
      const nav = page.locator('[data-sidebar="sidebar"]').first();
      // `.last()` targets the collapsible sub-menu entry rather than the parent
      // "Collection" link — /collection/members in particular appears twice in
      // the sidebar (once as the parent group's href, once as the Members
      // sub-link). Don't replace with `.first()` without re-verifying.
      await nav.locator(`a[href="${href}"]`).last().click();
      await expect(page).toHaveURL(new RegExp(`${href}(\\?.*)?$`));
    });
  }

  test("collection selector dropdown opens", async ({ page }) => {
    // The Selector.vue Button uses role="combobox" with title="Select Collection". Its
    // accessible name is recomputed from its text content after the collection loads
    // (becoming "Test Users' Home"), so we match the stable title attribute instead.
    const selector = page.locator('[data-sidebar="sidebar"] [role="combobox"][title="Select Collection"]');
    await expect(selector).toBeVisible();
    await expect(selector).toHaveAttribute("aria-expanded", "false");
    await selector.click();
    await expect(selector).toHaveAttribute("aria-expanded", "true");
  });

  test("logout returns to /", async ({ page }) => {
    await page.getByTestId("logout-button").click();
    await expect(page).toHaveURL("/");
  });
});
