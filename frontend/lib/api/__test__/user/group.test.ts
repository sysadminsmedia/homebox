import { faker } from "@faker-js/faker";
import { describe, expect, test } from "vitest";
import { factories } from "../factories";
import { sharedUserClient } from "../test-utils";

describe("first time user workflow (register, login, join group)", () => {
  test("user should be able to update group", async () => {
    const { client } = await factories.client.singleUse();

    const name = faker.person.firstName();

    const { response, data: group } = await client.group.update({
      name,
      currency: "eur",
    });

    expect(response.status).toBe(200);
    expect(group.name).toBe(name);
  });

  test("user should be able to get own group", async () => {
    const { client } = await factories.client.singleUse();

    const { response, data: group } = await client.group.get();

    expect(response.status).toBe(200);
    expect(group.name).toBeTruthy();
    expect(group.currency).toBe("USD");
  });

  test("user should be able to join create join token and have user signup", async () => {
    const api = factories.client.public();

    // Setup User 1 Token
    const client = await sharedUserClient();
    const { data: user1 } = await client.user.self();

    const { response, data } = await client.group.createInvitation({
      expiresAt: new Date(Date.now() + 1000 * 60 * 60 * 24 * 7),
      uses: 1,
    });

    expect(response.status).toBe(201);
    expect(data.token).toBeTruthy();

    // Create User 2 with token
    const duplicateUser = factories.user();
    duplicateUser.token = data.token;

    const { response: registerResp } = await api.register(duplicateUser);
    expect(registerResp.status).toBe(204);

    const { response: loginResp, data: loginData } = await api.login(duplicateUser.email, duplicateUser.password);
    expect(loginResp.status).toBe(200);

    // Get Self and Assert
    const client2 = factories.client.user(loginData.token);
    const { data: user2 } = await client2.user.self();

    expect(user2.item.defaultGroupId).toBe(user1.item.defaultGroupId);

    // Cleanup User 2
    const { response: deleteResp } = await client2.user.delete();
    expect(deleteResp.status).toBe(204);
  });
});
