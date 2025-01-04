import { faker } from "@faker-js/faker";
import { describe, test, expect } from "vitest";
import type { ItemField, ItemUpdate, LocationOut } from "../../types/data-contracts";
import { AttachmentTypes } from "../../types/non-generated";
import type { UserClient } from "../../user";
import { factories } from "../factories";
import { sharedUserClient } from "../test-utils";

describe("user should be able to create an item and add an attachment", () => {
  let increment = 0;
  /**
   * useLocation sets up a location resource for testing, and returns a function
   * that can be used to delete the location from the backend server.
   */
  async function useLocation(api: UserClient): Promise<[LocationOut, () => Promise<void>]> {
    const { response, data } = await api.locations.create({
      parentId: null,
      name: `__test__.location.name_${increment}`,
      description: `__test__.location.description_${increment}`,
    });
    expect(response.status).toBe(201);
    increment++;

    const cleanup = async () => {
      const { response } = await api.locations.delete(data.id);
      expect(response.status).toBe(204);
    };

    return [data, cleanup];
  }

  test("user should be able to create an item and add an attachment", async () => {
    const api = await sharedUserClient();
    const [location, cleanup] = await useLocation(api);

    const { response, data: item } = await api.items.create({
      parentId: null,
      name: "test-item",
      labelIds: [],
      description: "test-description",
      locationId: location.id,
    });
    expect(response.status).toBe(201);

    // Add attachment
    {
      const testFile = new Blob(["test"], { type: "text/plain" });
      const { response } = await api.items.attachments.add(item.id, testFile, "test.txt", AttachmentTypes.Attachment);
      expect(response.status).toBe(201);
    }

    // Get Attachment
    const { response: itmResp, data } = await api.items.get(item.id);
    expect(itmResp.status).toBe(200);

    expect(data.attachments).toHaveLength(1);
    expect(data.attachments[0].document.title).toBe("test.txt");

    const resp = await api.items.attachments.delete(data.id, data.attachments[0].id);
    expect(resp.response.status).toBe(204);

    api.items.delete(item.id);
    await cleanup();
  });

  test("user should be able to create and delete fields on an item", async () => {
    const api = await sharedUserClient();
    const [location, cleanup] = await useLocation(api);

    const { response, data: item } = await api.items.create({
      parentId: null,
      name: faker.vehicle.model(),
      labelIds: [],
      description: faker.lorem.paragraph(1),
      locationId: location.id,
    });
    expect(response.status).toBe(201);

    const fields: ItemField[] = [
      factories.itemField(),
      factories.itemField(),
      factories.itemField(),
      factories.itemField(),
    ];

    // Add fields
    const itemUpdate = {
      parentId: null,
      ...item,
      locationId: item.location?.id || null,
      labelIds: item.labels.map(l => l.id),
      fields,
    };

    const { response: updateResponse, data: item2 } = await api.items.update(item.id, itemUpdate as ItemUpdate);
    expect(updateResponse.status).toBe(200);

    expect(item2.fields).toHaveLength(fields.length);

    for (let i = 0; i < fields.length; i++) {
      expect(item2.fields[i].name).toBe(fields[i].name);
      expect(item2.fields[i].textValue).toBe(fields[i].textValue);
      expect(item2.fields[i].numberValue).toBe(fields[i].numberValue);
    }

    itemUpdate.fields = [fields[0], fields[1]];

    const { response: updateResponse2, data: item3 } = await api.items.update(item.id, itemUpdate as ItemUpdate);
    expect(updateResponse2.status).toBe(200);

    expect(item3.fields).toHaveLength(2);
    for (let i = 0; i < item3.fields.length; i++) {
      expect(item3.fields[i].name).toBe(itemUpdate.fields[i].name);
      expect(item3.fields[i].textValue).toBe(itemUpdate.fields[i].textValue);
      expect(item3.fields[i].numberValue).toBe(itemUpdate.fields[i].numberValue);
    }

    cleanup();
  });

  test("users should be able to create and few maintenance logs for an item", async () => {
    const api = await sharedUserClient();
    const [location, cleanup] = await useLocation(api);
    const { response, data: item } = await api.items.create({
      parentId: null,
      name: faker.vehicle.model(),
      labelIds: [],
      description: faker.lorem.paragraph(1),
      locationId: location.id,
    });
    expect(response.status).toBe(201);

    const maintenanceEntries = [];
    for (let i = 0; i < 5; i++) {
      const { response, data } = await api.items.maintenance.create(item.id, {
        name: faker.vehicle.model(),
        description: faker.lorem.paragraph(1),
        completedDate: faker.date.past(),
        scheduledDate: "null",
        cost: faker.number.int(100).toString(),
      });

      expect(response.status).toBe(201);
      maintenanceEntries.push(data);
    }

    // Log
    {
      const { response, data } = await api.items.maintenance.getLog(item.id);
      expect(response.status).toBe(200);
      expect(data).toHaveLength(maintenanceEntries.length);
    }

    cleanup();
  });

  test("full path of item should be retrievable", async () => {
    const api = await sharedUserClient();
    const [location, cleanup] = await useLocation(api);

    const locations = [location.name, faker.animal.dog(), faker.animal.cat(), faker.animal.cow(), faker.animal.bear()];

    let lastLocationId = location.id;
    for (let i = 1; i < locations.length; i++) {
      // Skip first one
      const { response, data: loc } = await api.locations.create({
        parentId: lastLocationId,
        name: locations[i],
        description: "",
      });
      expect(response.status).toBe(201);

      lastLocationId = loc.id;
    }

    const { response, data: item } = await api.items.create({
      name: faker.vehicle.model(),
      labelIds: [],
      description: faker.lorem.paragraph(1),
      locationId: lastLocationId,
    });
    expect(response.status).toBe(201);

    const { response: pathResponse, data: fullpath } = await api.items.fullpath(item.id);
    expect(pathResponse.status).toBe(200);

    const names = fullpath.map(p => p.name);

    expect(names).toHaveLength(locations.length + 1);
    expect(names).toEqual([...locations, item.name]);

    cleanup();
  });

  test("child items sync their location to their parent", async () => {
    const api = await sharedUserClient();
    const [parentLocation, parentCleanup] = await useLocation(api);
    const [childsLocation, childsCleanup] = await useLocation(api);

    const { response: parentResponse, data: parent } = await api.items.create({
      name: "parent-item",
      labelIds: [],
      description: "test-description",
      locationId: parentLocation.id,
    });
    expect(parentResponse.status).toBe(201);
    expect(parent.id).toBeTruthy();

    const { response: child1Response, data: child1Item } = await api.items.create({
      name: "child1-item",
      labelIds: [],
      description: "test-description",
      locationId: childsLocation.id,
    });
    expect(child1Response.status).toBe(201);
    const child1ItemUpdate = {
      parentId: parent.id,
      ...child1Item,
      locationId: child1Item.location?.id,
      labelIds: [],
    };
    const { response: child1UpdatedResponse } = await api.items.update(child1Item.id, child1ItemUpdate as ItemUpdate);
    expect(child1UpdatedResponse.status).toBe(200);

    const { response: child2Response, data: child2Item } = await api.items.create({
      name: "child2-item",
      labelIds: [],
      description: "test-description",
      locationId: childsLocation.id,
    });
    expect(child2Response.status).toBe(201);
    const child2ItemUpdate = {
      parentId: parent.id,
      ...child2Item,
      locationId: child2Item.location?.id,
      labelIds: [],
    };
    const { response: child2UpdatedResponse } = await api.items.update(child2Item.id, child2ItemUpdate as ItemUpdate);
    expect(child2UpdatedResponse.status).toBe(200);

    const itemUpdate = {
      parentId: null,
      ...parent,
      locationId: parentLocation.id,
      labelIds: [],
      syncChildItemsLocations: true,
    };
    const { response: updateResponse } = await api.items.update(parent.id, itemUpdate);
    expect(updateResponse.status).toBe(200);

    const { response: child1FinalResponse, data: child1FinalData } = await api.items.get(child1Item.id);
    expect(child1FinalResponse.status).toBe(200);
    expect(child1FinalData.location?.id).toBe(parentLocation.id);

    const { response: child2FinalResponse, data: child2FinalData } = await api.items.get(child2Item.id);
    expect(child2FinalResponse.status).toBe(200);
    expect(child2FinalData.location?.id).toBe(parentLocation.id);

    parentCleanup();
    childsCleanup();
  });
});
