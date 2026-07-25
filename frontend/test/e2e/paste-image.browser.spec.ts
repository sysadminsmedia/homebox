import type { APIRequestContext, Page } from "@playwright/test";
import { expect, test } from "@playwright/test";

// 1x1 transparent PNG.
const PNG_BASE64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==";

const PASSWORD = "ThisIsAStrongDemoPass";

type Fixture = {
  email: string;
  password: string;
  token: string;
  attachmentToken: string;
  locationId: string;
  itemId: string;
};

async function api(request: APIRequestContext, token: string, method: "get" | "post", path: string, body?: unknown) {
  const response = await request[method](`/api/v1${path}`, {
    headers: token ? { Authorization: token } : {},
    ...(body === undefined ? {} : { data: body }),
  });

  expect(response.ok(), `${method.toUpperCase()} ${path} -> ${response.status()} ${await response.text()}`).toBe(true);

  return response.status() === 204 ? null : response.json();
}

/**
 * Registers a throwaway user and seeds it with a location and an item. The user lands in
 * its own group, so nothing already in the database is touched.
 */
async function seed(request: APIRequestContext): Promise<Fixture> {
  const email = `paste-${Date.now()}-${Math.floor(Math.random() * 1e6)}@example.com`;

  await api(request, "", "post", "/users/register", { email, name: "Paste Test", password: PASSWORD });

  const { token, attachmentToken } = await api(request, "", "post", "/users/login", {
    username: email,
    password: PASSWORD,
    stayLoggedIn: false,
  });

  const entityTypes = await api(request, token, "get", "/entity-types");
  const typeId = (isLocation: boolean) =>
    (entityTypes.items ?? entityTypes).find((type: { isLocation: boolean }) => type.isLocation === isLocation).id;

  const location = await api(request, token, "post", "/entities", {
    name: "Paste Test Location",
    description: "",
    quantity: 1,
    tagIds: [],
    entityTypeId: typeId(true),
  });

  const item = await api(request, token, "post", "/entities", {
    name: "Paste Test Item",
    description: "",
    quantity: 1,
    tagIds: [],
    parentId: location.id,
    entityTypeId: typeId(false),
  });

  return { email, password: PASSWORD, token, attachmentToken, locationId: location.id, itemId: item.id };
}

async function login(page: Page, fixture: Fixture) {
  await page.goto("/home");
  await expect(page).toHaveURL("/");
  await page.fill("input[type='text']", fixture.email);
  await page.fill("input[type='password']", fixture.password);
  await page.click("button[type='submit']");
  await expect(page).toHaveURL("/home");

  // The backend sets its auth cookies with Domain=localhost when served from
  // localhost, and WebKit silently rejects any cookie whose Domain attribute
  // is a single-label host (https://bugs.webkit.org/show_bug.cgi?id=160368) -
  // so on webkit the login above leaves no usable session and every API call
  // 401s. Playwright's addCookies() writes straight into the jar, bypassing
  // Set-Cookie parsing, so re-add the session here (values from seed()'s API
  // login) and re-enter the app with it. Harmless on the other browsers.
  //
  // Tear the page down first: a straggler 401 from the dying page would run
  // invalidateSession(), which deletes the client-writable cookies - including
  // the ones injected below, if they were already in the jar.
  const origin = new URL(page.url()).origin;
  await page.goto("about:blank");
  const cookie = (name: string, value: string, httpOnly: boolean) => ({
    name,
    value,
    url: origin,
    httpOnly,
    sameSite: "Lax" as const,
  });
  await page
    .context()
    .addCookies([
      cookie("hb.auth.token", fixture.token.replace(/^Bearer /, ""), true),
      cookie("hb.auth.remember", "false", true),
      cookie("hb.auth.session", "true", false),
      cookie("hb.auth.attachment_token", fixture.attachmentToken, false),
    ]);
  await page.goto("/home");
  await expect(page).toHaveURL("/home");
}

/**
 * Firefox builds a constructed ClipboardEvent with an *empty* copy of the DataTransfer it
 * is handed, so a synthetic paste cannot carry a file there. Real pastes are unaffected -
 * this is purely a limit on what the test harness can fake.
 */
async function supportsSyntheticPaste(page: Page): Promise<boolean> {
  return page.evaluate(() => {
    const transfer = new DataTransfer();
    transfer.items.add(new File(["x"], "probe.png", { type: "image/png" }));
    return new ClipboardEvent("paste", { clipboardData: transfer }).clipboardData?.files.length === 1;
  });
}

/** Dispatches the paste event a browser fires when the clipboard holds an image. */
async function pasteImage(page: Page) {
  await page.evaluate(async base64 => {
    const blob = await (await fetch(`data:image/png;base64,${base64}`)).blob();
    const transfer = new DataTransfer();
    transfer.items.add(new File([blob], "image.png", { type: "image/png" }));

    document.dispatchEvent(new ClipboardEvent("paste", { clipboardData: transfer, bubbles: true, cancelable: true }));
  }, PNG_BASE64);
}

async function attachmentCount(request: APIRequestContext, fixture: Fixture, id: string): Promise<number> {
  const entity = await api(request, fixture.token, "get", `/entities/${id}`);
  return entity.attachments.length;
}

// Every test here registers a user and writes an attachment. Run them in order: against
// a sqlite backend, concurrent attachment writes hit SQLITE_BUSY, the same reason
// `pnpm test:ci` runs vitest with --no-file-parallelism.
test.describe.configure({ mode: "serial" });

test.describe("paste image from clipboard", () => {
  test("pastes into the entity create modal", async ({ page, request }) => {
    const fixture = await seed(request);
    await login(page, fixture);
    test.skip(!(await supportsSyntheticPaste(page)), "browser cannot carry a file in a synthetic paste");

    await page.getByRole("button", { name: "Create", exact: true }).first().click();
    await page.getByRole("menuitem", { name: "Item / Asset" }).click();
    // The upload button is aria-hidden - the file input stacked on top of it is the
    // real control - so anchor on the input itself.
    await expect(page.locator("input#photo-uploader")).toBeAttached();

    await pasteImage(page);

    // The pasted image shows up as a preview thumbnail before the item is created.
    await expect(page.locator("img[src^='data:image/png']").first()).toBeVisible();

    await page.getByLabel(/^Item Name/).fill("Pasted Photo Item");
    await page.getByRole("combobox", { name: /Parent Location/ }).click();
    await page.getByPlaceholder("Search Locations").fill("Paste Test Location");
    await expect(page.getByRole("option", { name: "Paste Test Location" })).toBeVisible();
    // The option list re-renders as results settle, so drive it from the keyboard
    // rather than racing a click against a moving target.
    await page.keyboard.press("ArrowDown");
    await page.keyboard.press("Enter");
    await expect(page.getByRole("combobox", { name: /Parent Location/ })).toContainText("Paste Test Location");

    const upload = page.waitForResponse(r => /\/attachments$/.test(r.url()) && r.request().method() === "POST");
    await page.getByRole("button", { name: "Create", exact: true }).last().click();
    await upload;

    // ...and is attached to the item that gets created.
    const created = await api(request, fixture.token, "get", "/entities?q=Pasted Photo Item");
    const detail = await api(request, fixture.token, "get", `/entities/${created.items[0].id}`);
    expect(detail.attachments).toHaveLength(1);
    expect(detail.attachments[0].title).toMatch(/^pasted-image-\d{8}-\d{6}\.png$/);
    expect(detail.attachments[0].mimeType).toBe("image/png");
  });

  test("pastes an attachment onto the item edit page", async ({ page, request }) => {
    const fixture = await seed(request);
    await login(page, fixture);

    test.skip(!(await supportsSyntheticPaste(page)), "browser cannot carry a file in a synthetic paste");

    expect(await attachmentCount(request, fixture, fixture.itemId)).toBe(0);

    await page.goto(`/item/${fixture.itemId}/edit`);
    await expect(page.getByText("Attachments").first()).toBeVisible();

    await pasteImage(page);

    await expect(page.getByText("Attachment uploaded")).toBeVisible();
    await expect.poll(() => attachmentCount(request, fixture, fixture.itemId)).toBe(1);
  });

  test("pastes an attachment onto the location edit page", async ({ page, request }) => {
    const fixture = await seed(request);
    await login(page, fixture);

    test.skip(!(await supportsSyntheticPaste(page)), "browser cannot carry a file in a synthetic paste");

    expect(await attachmentCount(request, fixture, fixture.locationId)).toBe(0);

    await page.goto(`/location/${fixture.locationId}/edit`);
    await expect(page.getByText("Attachments").first()).toBeVisible();

    await pasteImage(page);

    await expect(page.getByText("Attachment uploaded")).toBeVisible();
    await expect.poll(() => attachmentCount(request, fixture, fixture.locationId)).toBe(1);
  });

  test("uploads via the Paste Image button", async ({ page, context, request, browserName }) => {
    // Only Chromium lets Playwright pre-grant the clipboard permissions this needs.
    test.skip(browserName !== "chromium", "clipboard permissions are chromium-only in playwright");

    await context.grantPermissions(["clipboard-read", "clipboard-write"]);

    const fixture = await seed(request);
    await login(page, fixture);

    await page.goto(`/item/${fixture.itemId}/edit`);
    await expect(page.getByTestId("paste-image-button")).toBeVisible();

    await page.evaluate(async base64 => {
      const blob = await (await fetch(`data:image/png;base64,${base64}`)).blob();
      await navigator.clipboard.write([new ClipboardItem({ "image/png": blob })]);
    }, PNG_BASE64);

    await page.getByTestId("paste-image-button").click();

    await expect(page.getByText("Attachment uploaded")).toBeVisible();
    await expect.poll(() => attachmentCount(request, fixture, fixture.itemId)).toBe(1);
  });

  test("leaves ordinary text pastes alone", async ({ page, context, request, browserName }) => {
    test.skip(browserName !== "chromium", "needs clipboard permissions to paste for real");

    await context.grantPermissions(["clipboard-read", "clipboard-write"]);

    const fixture = await seed(request);
    await login(page, fixture);

    await page.goto(`/item/${fixture.itemId}/edit`);

    // The label carries a trailing character counter, hence the prefix match.
    const name = page.getByLabel(/^Name/);
    await expect(name).toHaveValue("Paste Test Item");
    await name.fill("");

    await page.evaluate(() => navigator.clipboard.writeText("Pasted Name"));
    await name.focus();
    await page.keyboard.press("ControlOrMeta+v");

    // The paste listener must not swallow text pastes into form fields.
    await expect(name).toHaveValue("Pasted Name");
  });
});
