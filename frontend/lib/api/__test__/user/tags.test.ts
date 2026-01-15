import { describe, expect, test } from "vitest";
import type { TagOut } from "../../types/data-contracts";
import type { UserClient } from "../../user";
import { factories } from "../factories";
import { sharedUserClient } from "../test-utils";

describe("tags lifecycle (create, update, delete)", () => {
  /**
   * useTag sets up a tag resource for testing, and returns a function
   * that can be used to delete the tag from the backend server.
   */
  async function useTag(api: UserClient): Promise<[TagOut, () => Promise<void>]> {
    const { response, data } = await api.tags.create(factories.tag());
    expect(response.status).toBe(201);

    const cleanup = async () => {
      const { response } = await api.tags.delete(data.id);
      expect(response.status).toBe(204);
    };
    return [data, cleanup];
  }

  test("user should be able to create a tag", async () => {
    const api = await sharedUserClient();

    const tagData = factories.tag();

    const { response, data } = await api.tags.create(tagData);

    expect(response.status).toBe(201);
    expect(data.id).toBeTruthy();

    // Ensure we can get the tag
    const { response: getResponse, data: getData } = await api.tags.get(data.id);

    expect(getResponse.status).toBe(200);
    expect(getData.id).toBe(data.id);
    expect(getData.name).toBe(tagData.name);
    expect(getData.description).toBe(tagData.description);

    // Cleanup
    const { response: deleteResponse } = await api.tags.delete(data.id);
    expect(deleteResponse.status).toBe(204);
  });

  test("user should be able to update a tag", async () => {
    const api = await sharedUserClient();
    const [tag, cleanup] = await useTag(api);

    const tagData = {
      name: "test-tag",
      description: "test-description",
      color: "",
    };

    const { response, data } = await api.tags.update(tag.id, tagData);
    expect(response.status).toBe(200);
    expect(data.id).toBe(tag.id);

    // Ensure we can get the tag
    const { response: getResponse, data: getData } = await api.tags.get(data.id);
    expect(getResponse.status).toBe(200);
    expect(getData.id).toBe(data.id);
    expect(getData.name).toBe(tagData.name);
    expect(getData.description).toBe(tagData.description);

    // Cleanup
    await cleanup();
  });

  test("user should be able to delete a tag", async () => {
    const api = await sharedUserClient();
    const [tag, _] = await useTag(api);

    const { response } = await api.tags.delete(tag.id);
    expect(response.status).toBe(204);

    // Ensure we can't get the tag
    const { response: getResponse } = await api.tags.get(tag.id);
    expect(getResponse.status).toBe(404);
  });
});
