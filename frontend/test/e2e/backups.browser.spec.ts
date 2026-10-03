import type { Page, Response, Route } from "@playwright/test";
import { expect, test } from "@playwright/test";

// The e2e server runs in demo mode, where backup destinations cannot be
// changed, so these tests mock the backup API in the browser and exercise the
// real UI against it.

// Each test signs in, and sign-in hashes the password. Running them one after
// another keeps the demo server responsive instead of racing eight logins.
test.describe.configure({ mode: "serial" });

// Pages are loaded cold and the first render of the destinations section waits
// on three API calls, so give assertions more room than the 5s default.
const slow = expect.configure({ timeout: 15_000 });

type Dest = Record<string, unknown> & { id: string; name: string; type: string };

interface MockOptions {
  enabled?: boolean;
  oauthProviders?: string[];
  oidcSuggestion?: { provider: string; destType: string; email: string } | null;
  remoteEnabled?: boolean;
  destinations?: Dest[];
  testResult?: Record<string, unknown>;
}

function json(route: Route, body: unknown, status = 200) {
  return route.fulfill({ status, contentType: "application/json", body: JSON.stringify(body) });
}

function destination(overrides: Partial<Dest>): Dest {
  return {
    id: crypto.randomUUID(),
    groupId: crypto.randomUUID(),
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    name: "Destination",
    description: "",
    type: "primary",
    connString: "",
    username: "",
    hostKey: "",
    prefix: "homebox-backups",
    enabled: true,
    scheduleEnabled: true,
    frequency: "daily",
    cronExpr: "",
    intervalHours: 1,
    atHour: 3,
    atMinute: 0,
    weekday: 0,
    dayOfMonth: 1,
    skipIfUnchanged: true,
    keepDaily: 7,
    keepWeekly: 4,
    keepMonthly: 6,
    healthIntervalMinutes: 15,
    alertsEnabled: true,
    alertFailureThreshold: 2,
    alertStaleHours: 48,
    healthStatus: "unknown",
    healthFailures: 0,
    hasSecret: false,
    ...overrides,
  };
}

async function mockBackups(page: Page, opts: MockOptions = {}) {
  const destinations = [...(opts.destinations ?? [])];
  const posted: Record<string, unknown>[] = [];

  await page.route("**/api/v1/group/backup-options", route =>
    json(route, {
      enabled: opts.enabled ?? true,
      localEnabled: false,
      allowCustomEndpoints: true,
      remoteEnabled: opts.remoteEnabled ?? false,
      oauthProviders: opts.oauthProviders ?? [],
      ...(opts.oidcSuggestion ? { oidcSuggestion: opts.oidcSuggestion } : {}),
    })
  );
  await page.route("**/api/v1/group/backup-destinations", async route => {
    if (route.request().method() === "POST") {
      const body = route.request().postDataJSON() as Record<string, unknown>;
      posted.push(body);
      const created = destination({ ...body, id: crypto.randomUUID() } as Partial<Dest>);
      destinations.push(created);
      return json(route, created, 201);
    }
    return json(route, { items: destinations });
  });
  await page.route("**/api/v1/group/backup-destinations/test", route =>
    json(route, opts.testResult ?? { ok: true, latencyMs: 12, message: "Wrote and deleted a 1 KB test file" })
  );
  return { posted };
}

// The API sets its session cookies with `Domain=localhost`, which WebKit
// refuses to store. Without them the next full page load is unauthenticated and
// lands back on the login form, so replay them as host-only cookies. Other
// browsers already hold them and are unaffected.
async function keepSessionCookies(page: Page, login: Response) {
  const origin = new URL(page.url()).origin;
  const cookies = (await login.headersArray())
    .filter(h => h.name.toLowerCase() === "set-cookie")
    .map(h => {
      const [pair = "", ...attrs] = h.value.split(";").map(part => part.trim());
      const eq = pair.indexOf("=");
      const expires = attrs.find(a => a.toLowerCase().startsWith("expires="));
      return {
        name: pair.slice(0, eq),
        value: pair.slice(eq + 1),
        url: origin,
        httpOnly: attrs.some(a => a.toLowerCase() === "httponly"),
        ...(expires ? { expires: new Date(expires.slice("expires=".length)).getTime() / 1000 } : {}),
      };
    });
  await page.context().addCookies(cookies);
}

async function login(page: Page) {
  await page.goto("/home");
  await page.fill("input[type='text']", "demo@example.com");
  await page.fill("input[type='password']", "demodemo");
  const loggedIn = page.waitForResponse(r => r.url().endsWith("/api/v1/users/login") && r.ok());
  await page.click("button[type='submit']");
  await keepSessionCookies(page, await loggedIn);
  await slow(page).toHaveURL("/home");
  // Let the session settle before navigating away, or the auth guard can bounce
  // the next page load back to the login form.
  await page.waitForLoadState("networkidle");
}

async function openTools(page: Page) {
  await page.goto("/collection/tools");
  await slow(page.getByText("Backup & Restore").first()).toBeVisible();
}

async function pickType(page: Page, label: string | RegExp) {
  await page.locator('[role="dialog"] [role="combobox"]').first().click();
  await page.getByRole("option", { name: label }).click();
}

test("the backups section is hidden when scheduled backups are off", async ({ page }) => {
  await mockBackups(page, { enabled: false });
  await login(page);
  await openTools(page);
  await slow(page.getByText("Automatic backups & destinations")).toHaveCount(0);
  // Manual backups are unaffected.
  await slow(page.getByRole("button", { name: "Start Backup" })).toBeVisible();
});

test("a primary-storage destination can be added and listed", async ({ page }) => {
  const { posted } = await mockBackups(page);
  await login(page);
  await openTools(page);

  await slow(page.getByText("No extra destinations yet")).toBeVisible();
  await page.getByRole("button", { name: "Add destination" }).click();
  await page.getByLabel("Name", { exact: true }).fill("Nightly");
  await page.getByRole("button", { name: "Save" }).click();

  await slow(page.getByText("Nightly", { exact: true })).toBeVisible();
  await slow(page.getByText("Daily at 03:00")).toBeVisible();
  expect(posted).toHaveLength(1);
  expect(posted[0]).toMatchObject({ name: "Nightly", type: "primary", frequency: "daily", skipIfUnchanged: true });
});

test("a custom cron schedule is sent and summarised", async ({ page }) => {
  const { posted } = await mockBackups(page);
  await login(page);
  await openTools(page);

  await page.getByRole("button", { name: "Add destination" }).click();
  await page.getByLabel("Name", { exact: true }).fill("Cron");
  await page.locator('[role="dialog"] [role="combobox"]').nth(1).click();
  await page.getByRole("option", { name: "Custom (cron)" }).click();
  await page.getByLabel("Cron expression").fill("0 3 * * 1-5");
  await page.getByRole("button", { name: "Save" }).click();

  await slow(page.getByText("Cron 0 3 * * 1-5")).toBeVisible();
  expect(posted[0]).toMatchObject({ frequency: "cron", cronExpr: "0 3 * * 1-5" });
});

test("a cloud drive cannot be saved until an account is connected", async ({ page }) => {
  await mockBackups(page, { oauthProviders: ["google"] });
  await login(page);
  await openTools(page);

  await page.getByRole("button", { name: "Add destination" }).click();
  await page.getByLabel("Name", { exact: true }).fill("Drive");
  await pickType(page, "Google Drive");
  await slow(page.getByText("No account connected yet.")).toBeVisible();
  await slow(page.getByRole("button", { name: "Connect Google Drive" })).toBeVisible();
  await slow(page.getByRole("button", { name: "Save" })).toBeDisabled();
});

test("providers without a configured app are not offered", async ({ page }) => {
  await mockBackups(page, { oauthProviders: ["dropbox"], remoteEnabled: true });
  await login(page);
  await openTools(page);

  await page.getByRole("button", { name: "Add destination" }).click();
  await page.locator('[role="dialog"] [role="combobox"]').first().click();
  await slow(page.getByRole("option", { name: "Dropbox" })).toBeVisible();
  await slow(page.getByRole("option", { name: "Google Drive" })).toHaveCount(0);
  await slow(page.getByRole("option", { name: "OneDrive" })).toHaveCount(0);
  await slow(page.getByRole("option", { name: /SFTP/ })).toBeVisible();
});

test("an SFTP host key is trusted from the test result", async ({ page }) => {
  const fingerprint = "SHA256:p5K5KtC8t8/urilLu1nFRZSxa2ChT0j/sL9AYfw9jlc";
  await mockBackups(page, {
    remoteEnabled: true,
    testResult: {
      ok: false,
      latencyMs: 4,
      message: "the server's SSH host key has not been verified yet",
      hostKey: fingerprint,
    },
  });
  await login(page);
  await openTools(page);

  await page.getByRole("button", { name: "Add destination" }).click();
  await page.getByLabel("Name", { exact: true }).fill("NAS");
  await pickType(page, /SFTP/);
  await page.getByLabel("Address").fill("sftp://nas.lan:22/mnt/tank/backups");
  await page.getByLabel("Username").fill("homebox");
  await page.getByLabel("Password", { exact: true }).fill("secret");

  await page.getByRole("button", { name: "Test connection" }).click();
  await slow(page.getByText(fingerprint).first()).toBeVisible();
  await page.getByRole("button", { name: "Trust this host key" }).click();
  await slow(page.getByLabel("SSH host key fingerprint")).toHaveValue(fingerprint);
});

test("the OIDC offer is shown once and can be dismissed", async ({ page }) => {
  await mockBackups(page, {
    oauthProviders: ["google"],
    oidcSuggestion: { provider: "google", destType: "gdrive", email: "me@example.com" },
  });
  await login(page);
  await openTools(page);

  const title = page.getByText("Back up to your Google Drive?");
  await slow(title).toBeVisible();
  await slow(page.getByText("me@example.com")).toBeVisible();

  await page.getByRole("button", { name: "Not now" }).click();
  await slow(title).toHaveCount(0);

  await page.reload();
  await slow(page.getByText("Automatic backups & destinations")).toBeVisible();
  await slow(title).toHaveCount(0);
});

test("a destination that is unreachable shows a banner", async ({ page }) => {
  await mockBackups(page, {
    destinations: [
      destination({
        name: "Offsite NAS",
        type: "sftp",
        healthStatus: "unreachable",
        healthError: "connection refused",
      }),
    ],
  });
  await login(page);
  await openTools(page);

  await slow(page.getByRole("alert").getByText('"Offsite NAS" is unreachable')).toBeVisible();
  await slow(page.getByRole("alert").getByText("connection refused")).toBeVisible();
});
