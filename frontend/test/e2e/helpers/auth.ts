import { expect, type APIRequestContext, type Page } from "@playwright/test";
import { faker } from "@faker-js/faker";

export const STRONG_PASSWORD = "ThisIsAStrongDemoPass";

export type ApiEntity = { id: string; name: string };

/**
 * Create a location via REST. Fetches entity-types to find the default
 * isLocation type so the resulting entity is a location rather than an item.
 */
export async function apiCreateLocation(
  request: APIRequestContext,
  name: string,
  parentId?: string
): Promise<ApiEntity> {
  const etRes = await request.get("/api/v1/entity-types");
  if (!etRes.ok()) throw new Error(`entity-types fetch failed: ${etRes.status()}`);
  const entityTypes = (await etRes.json()) as Array<{ id: string; isLocation: boolean }>;
  const locationType = entityTypes.find(et => et.isLocation);
  const res = await request.post("/api/v1/entities", {
    data: {
      name,
      description: "",
      quantity: 1,
      tagIds: [],
      ...(parentId ? { parentId } : {}),
      ...(locationType ? { entityTypeId: locationType.id } : {}),
    },
  });
  if (!res.ok()) throw new Error(`create location failed: ${res.status()} ${await res.text()}`);
  return (await res.json()) as ApiEntity;
}

async function fillLogin(page: Page, email: string, password: string) {
  const loginEmail = page.getByRole("textbox", { name: "Email" });
  await loginEmail.fill("");
  await loginEmail.fill(email);
  const pw = page.getByRole("textbox", { name: "Password" });
  await pw.fill("");
  await pw.fill(password);
  await page.getByRole("button", { name: "Login", exact: true }).click();
  await expect(page).toHaveURL("/home");
}

export async function registerAndLogin(page: Page) {
  const email = faker.internet.email().toLowerCase();
  const password = STRONG_PASSWORD;
  const name = "Test User";

  const res = await page.request.post("/api/v1/users/register", {
    data: { name, email, password, token: "" },
  });
  expect(res.status(), "register should succeed").toBeLessThan(400);

  await page.goto("/");
  await fillLogin(page, email, password);
  return { email, password, name };
}
