import { expect, test } from "@playwright/test";

test.describe("Wipe Inventory E2E Test", () => {
  test.beforeEach(async ({ page }) => {
    // Login as demo user (owner with permissions)
    await page.goto("/");
    await page.fill("input[type='text']", "demo@example.com");
    await page.fill("input[type='password']", "demo");
    await page.click("button[type='submit']");
    await expect(page).toHaveURL("/home");
  });

  test("should open wipe inventory dialog with all options", async ({ page }) => {
    // Navigate to Tools page
    await page.goto("/tools");
    await page.waitForLoadState("networkidle");

    // Scroll to the bottom where wipe inventory is located
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
    await page.waitForTimeout(500);

    // Find and click the Wipe Inventory button
    const wipeButton = page.locator("button", { hasText: "Wipe Inventory" }).last();
    await expect(wipeButton).toBeVisible();
    await wipeButton.click();

    // Wait for dialog to appear
    await page.waitForTimeout(1000);

    // Verify dialog title is visible
    await expect(page.locator("text=Wipe Inventory").first()).toBeVisible();

    // Verify all checkboxes are present
    await expect(page.locator("input#wipe-labels-checkbox")).toBeVisible();
    await expect(page.locator("input#wipe-locations-checkbox")).toBeVisible();
    await expect(page.locator("input#wipe-maintenance-checkbox")).toBeVisible();

    // Verify labels for checkboxes
    await expect(page.locator("label[for='wipe-labels-checkbox']")).toBeVisible();
    await expect(page.locator("label[for='wipe-locations-checkbox']")).toBeVisible();
    await expect(page.locator("label[for='wipe-maintenance-checkbox']")).toBeVisible();

    // Verify both Cancel and Confirm buttons are present
    await expect(page.locator("button", { hasText: "Cancel" })).toBeVisible();
    const confirmButton = page.locator("button", { hasText: "Confirm" });
    await expect(confirmButton).toBeVisible();

    // Take screenshot of the modal
    await page.screenshot({
      path: "/tmp/playwright-logs/wipe-inventory-modal-initial.png",
    });
    console.log("✅ Screenshot saved: wipe-inventory-modal-initial.png");

    // Check all three options
    await page.check("input#wipe-labels-checkbox");
    await page.check("input#wipe-locations-checkbox");
    await page.check("input#wipe-maintenance-checkbox");
    await page.waitForTimeout(500);

    // Verify checkboxes are checked
    await expect(page.locator("input#wipe-labels-checkbox")).toBeChecked();
    await expect(page.locator("input#wipe-locations-checkbox")).toBeChecked();
    await expect(page.locator("input#wipe-maintenance-checkbox")).toBeChecked();

    // Take screenshot with all options checked
    await page.screenshot({
      path: "/tmp/playwright-logs/wipe-inventory-modal-options-checked.png",
    });
    console.log("✅ Screenshot saved: wipe-inventory-modal-options-checked.png");

    // Click Confirm button
    await confirmButton.click();
    await page.waitForTimeout(2000);

    // Wait for the dialog to close (verify button is no longer visible)
    await expect(confirmButton).not.toBeVisible({ timeout: 5000 });

    // Check for success toast notification
    // The toast should contain text about items being deleted
    const toastLocator = page.locator("[role='status'], [class*='toast'], [class*='sonner']");
    await expect(toastLocator.first()).toBeVisible({ timeout: 10000 });

    // Take screenshot of the page after confirmation
    await page.screenshot({
      path: "/tmp/playwright-logs/after-wipe-confirmation.png",
      fullPage: true,
    });
    console.log("✅ Screenshot saved: after-wipe-confirmation.png");

    console.log("✅ Test completed successfully!");
    console.log("✅ Wipe Inventory dialog opened correctly");
    console.log("✅ All three options (labels, locations, maintenance) are available");
    console.log("✅ Confirm button triggers the action");
    console.log("✅ Dialog closes after confirmation");
  });

  test("should cancel wipe inventory operation", async ({ page }) => {
    // Navigate to Tools page
    await page.goto("/tools");
    await page.waitForLoadState("networkidle");

    // Scroll to wipe inventory section
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
    await page.waitForTimeout(500);

    // Click Wipe Inventory button
    const wipeButton = page.locator("button", { hasText: "Wipe Inventory" }).last();
    await wipeButton.click();
    await page.waitForTimeout(1000);

    // Verify dialog is open
    await expect(page.locator("text=Wipe Inventory").first()).toBeVisible();

    // Click Cancel button
    const cancelButton = page.locator("button", { hasText: "Cancel" });
    await cancelButton.click();
    await page.waitForTimeout(1000);

    // Verify dialog is closed
    await expect(page.locator("text=Wipe Inventory").first()).not.toBeVisible({ timeout: 5000 });

    // Take screenshot after cancel
    await page.screenshot({
      path: "/tmp/playwright-logs/after-cancel.png",
    });
    console.log("✅ Screenshot saved: after-cancel.png");
    console.log("✅ Cancel button works correctly");
  });
});
