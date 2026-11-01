import { beforeAll, expect } from "vitest";
import { faker } from "@faker-js/faker";
import type { UserClient } from "../user";
import { factories } from "./factories";

const cache = {
  token: "",
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

/*
 * Returns the id of one of the group's location or item entity types. Entities
 * created without an entityTypeId default to the item type, so locations need
 * this to actually be locations.
 */
export async function entityTypeId(api: UserClient, isLocation: boolean): Promise<string> {
  const { response, data } = await api.entityTypes.getAll();
  expect(response.status).toBe(200);
  const entityType = data.find(t => t.isLocation === isLocation);
  expect(entityType).toBeTruthy();
  return entityType!.id;
}

beforeAll(async () => {
  await sharedUserClient();
});
