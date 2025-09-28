import { describe, expect, test } from "vitest";
import { factories } from "./factories";

describe("[GET] /api/v1/status", () => {
  test("server should respond", async () => {
    const api = factories.client.public();
    const { response, data } = await api.status();
    expect(response.status).toBe(200);
    expect(data.health).toBe(true);
  });
});

describe("first time user workflow (register, login, join group)", () => {
  const api = factories.client.public();
  const userData = factories.user();

  test("user should be able to register", async () => {
    const { response } = await api.register(userData);
    expect(response.status).toBe(204);
  });

  test("user should be able to login", async () => {
    const { response, data } = await api.login(userData.email, userData.password);
    expect(response.status).toBe(200);
    expect(data.token).toBeTruthy();

    // Cleanup
    const userApi = factories.client.user(data.token);
    {
      const { response } = await userApi.user.delete();
      expect(response.status).toBe(204);
    }
  });
});
