import { expect, test, type Page } from "@playwright/test";
import { faker } from "@faker-js/faker";
import { registerAndLogin, STRONG_PASSWORD } from "./helpers/auth";

const PASSWORD = STRONG_PASSWORD;

async function loginOnly(page: Page, email: string, password: string) {
  // Wait for the /status response so the `whenever(status)` effect that
  // auto-fills demo creds has already fired — otherwise it can clobber our fill.
  const statusPromise = page.waitForResponse(r => r.url().includes("/api/v1/status"));
  await page.goto("/");
  await statusPromise;
  const loginEmail = page.getByRole("textbox", { name: "Email" });
  await expect(loginEmail).toBeVisible({ timeout: 10000 });
  await loginEmail.fill("");
  await loginEmail.fill(email);
  const pw = page.getByRole("textbox", { name: "Password" });
  await pw.fill("");
  await pw.fill(password);
  await page.getByRole("button", { name: "Login", exact: true }).click();
  await expect(page).toHaveURL("/home", { timeout: 15000 });
}

async function confirmAlert(page: Page) {
  await page.getByRole("alertdialog").getByRole("button", { name: "Confirm" }).click();
}

async function createInvite(page: Page): Promise<string> {
  await page.goto("/collection/invites");
  await expect(page).toHaveURL(/\/collection\/invites/);
  // Wait for the invites page to finish loading (loading placeholder disappears).
  await expect(page.getByText("Loading", { exact: false }).first()).toBeHidden({ timeout: 15000 });

  await page.getByRole("button", { name: "Create Invite" }).click();

  const dialog = page.getByRole("dialog");
  await expect(dialog).toBeVisible();
  // Modal pre-fills uses=1 with a default 7-day expiry, so submit directly.
  await dialog.getByRole("button", { name: "Create", exact: true }).click();
  await expect(dialog).toBeHidden({ timeout: 5000 });

  const tokenCell = page.locator("span.font-mono").first();
  await expect(tokenCell).toBeVisible({ timeout: 10000 });
  const token = (await tokenCell.textContent())?.trim() ?? "";
  expect(token).toMatch(/^[A-Z0-9]{26}$/);
  return token;
}

test.describe("Collection members & invites", () => {
  test("lists the current user on the members page", async ({ page }) => {
    test.slow();
    const { email, name } = await registerAndLogin(page);

    await page.goto("/collection/members");
    await expect(page).toHaveURL(/\/collection\/members/);

    const row = page.getByRole("row").filter({ hasText: email });
    await expect(row).toBeVisible({ timeout: 10000 });
    await expect(row).toContainText(name);
  });

  test("creates, displays, copies, and deletes an invite", async ({ page, context, browserName }) => {
    test.slow();
    await registerAndLogin(page);

    // Clipboard permissions are only a valid `grantPermissions` name on Chromium.
    // Firefox / WebKit will throw "Unknown permission" if we pass them — skip there.
    // The backend's Permissions-Policy header sets `clipboard-read=(self)` only in
    // demo mode (see security.go), which the E2E task stack always enables via
    // `HBOX_DEMO: true`, so readText is allowed under test.
    if (browserName === "chromium") {
      await context.grantPermissions(["clipboard-read", "clipboard-write"]);
    }

    const token = await createInvite(page);

    // Locate the invite row by its token cell. The row contains the CopyText button (no accessible name)
    // and a destructive Delete button.
    const row = page.getByRole("row").filter({ has: page.locator("span.font-mono", { hasText: token }) });
    await expect(row).toBeVisible();

    // The CopyText button is the first button inside the row (it sits before the Delete icon button).
    // It has no accessible name on its own — just an icon + tooltip — so use a positional locator.
    const copyButton = row.getByRole("button").first();
    await expect(copyButton).toBeVisible();
    await copyButton.click();

    // Verify the clipboard actually received the token. Chromium is the only
    // browser Playwright exposes `clipboard-read` permission to (see the
    // grantPermissions guard above); Firefox/WebKit skip this check.
    if (browserName === "chromium") {
      const clipboardText = await page.evaluate(() => navigator.clipboard.readText());
      expect(clipboardText).toContain(token);
    }

    // The Delete button has aria-label="Delete" (from $t('global.delete')).
    await row.getByRole("button", { name: "Delete", exact: true }).click();
    await confirmAlert(page);

    await expect(row).toBeHidden({ timeout: 5000 });
  });

  test("second user joins via the invite token URL", async ({ page, browser }) => {
    test.slow();
    const inviter = await registerAndLogin(page);
    const token = await createInvite(page);

    const secondContext = await browser.newContext();
    try {
      const secondPage = await secondContext.newPage();

      // Append a unique suffix so parallel/retried runs can't collide on the
      // email and trip a 409 from the UNIQUE constraint.
      const inviteeEmail = `${faker.internet.username().toLowerCase()}-${crypto.randomUUID()}@example.com`;
      const inviteeName = "Second User";

      // Register via API with the invite token. The UI register flow + shared
      // `whenever(status)` effect can race with our field fills; the API path
      // is deterministic and still exercises the backend join-via-token path.
      const regRes = await secondContext.request.post("/api/v1/users/register", {
        data: { name: inviteeName, email: inviteeEmail, password: PASSWORD, token },
      });
      expect(regRes.ok(), `register (token) status=${regRes.status()}`).toBeTruthy();

      await loginOnly(secondPage, inviteeEmail, PASSWORD);

      await secondPage.goto("/collection/members");
      await expect(secondPage.getByRole("row").filter({ hasText: inviteeEmail })).toBeVisible({ timeout: 15000 });
      await expect(secondPage.getByRole("row").filter({ hasText: inviter.email })).toBeVisible({ timeout: 15000 });
    } finally {
      await secondContext.close();
    }
  });
});
