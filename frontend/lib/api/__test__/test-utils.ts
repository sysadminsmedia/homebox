import { beforeAll, expect } from "vitest";
import { faker } from "@faker-js/faker";
import type { UserClient } from "../user";
import { factories } from "./factories";

const cache = {
  token: "",
  entityTypeIds: null as { itemTypeId: string; locationTypeId: string } | null,
};

/*
 * Shared UserApi token for tests where the creation of a user is _not_ import
 * to the test. This is useful for tests that are testing the user API itself.
 */
export async function sharedUserClient(): Promise<UserClient> {
  if (cache.token) {
    return factories.client.user(cache.token);
  }
  const testUser = {
    email: faker.internet.email(),
    name: faker.person.fullName(),
    password: faker.internet.password(),
    token: "",
  };

  const api = factories.client.public();
  const { response: tryLoginResp, data } = await api.login(testUser.email, testUser.password);

  if (tryLoginResp.status === 200) {
    cache.token = data.token;
    return factories.client.user(cache.token);
  }

  const { response: registerResp } = await api.register(testUser);
  expect(registerResp.status).toBe(204);

  const { response: loginResp, data: loginData } = await api.login(testUser.email, testUser.password);
  expect(loginResp.status).toBe(200);

  cache.token = loginData.token;
  return factories.client.user(loginData.token);
}

export async function sharedEntityTypeIds(api?: UserClient): Promise<{ itemTypeId: string; locationTypeId: string }> {
  if (cache.entityTypeIds) {
    return cache.entityTypeIds;
  }

  const client = api ?? (await sharedUserClient());
  const { response, data } = await client.entityTypes.getAll();
  expect(response.status).toBe(200);

  const itemType = data.find(t => !t.isLocation);
  const locationType = data.find(t => t.isLocation);
  expect(itemType).toBeTruthy();
  expect(locationType).toBeTruthy();

  cache.entityTypeIds = {
    itemTypeId: itemType!.id,
    locationTypeId: locationType!.id,
  };
  return cache.entityTypeIds;
}

beforeAll(async () => {
  await sharedUserClient();
});
