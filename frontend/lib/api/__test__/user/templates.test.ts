import { describe, expect, test } from "vitest";
import type { ItemTemplateOut } from "../../types/data-contracts";
import type { UserClient } from "../../user";
import { factories } from "../factories";
import { sharedUserClient } from "../test-utils";

describe("templates lifecycle (create, update, delete)", () => {
  /**
   * useTemplate sets up a template resource for testing, and returns a function
   * that can be used to delete the template from the backend server.
   */
  async function useTemplate(api: UserClient): Promise<[ItemTemplateOut, () => Promise<void>]> {
    const { response, data } = await api.templates.create(factories.template());
    expect(response.status).toBe(201);

    const cleanup = async () => {
      const { response } = await api.templates.delete(data.id);
      expect(response.status).toBe(204);
    };

    return [data, cleanup];
  }

  test("user should be able to create a template", async () => {
    const api = await sharedUserClient();

    const templateData = factories.template();

    const { response, data } = await api.templates.create(templateData);

    expect(response.status).toBe(201);
    expect(data.id).toBeTruthy();

    // Ensure we can get the template
    const { response: getResponse, data: getData } = await api.templates.get(data.id);

    expect(getResponse.status).toBe(200);
    expect(getData.id).toBe(data.id);
    expect(getData.name).toBe(templateData.name);
    expect(getData.description).toBe(templateData.description);
    expect(getData.defaultQuantity).toBe(templateData.defaultQuantity);
    expect(getData.defaultInsured).toBe(templateData.defaultInsured);
    expect(getData.defaultName).toBe(templateData.defaultName);
    expect(getData.defaultDescription).toBe(templateData.defaultDescription);
    expect(getData.defaultManufacturer).toBe(templateData.defaultManufacturer);
    expect(getData.defaultModelNumber).toBe(templateData.defaultModelNumber);

    // Cleanup
    const { response: deleteResponse } = await api.templates.delete(data.id);
    expect(deleteResponse.status).toBe(204);
  });

  test("user should be able to get all templates", async () => {
    const api = await sharedUserClient();
    const [_, cleanup] = await useTemplate(api);

    const { response, data } = await api.templates.getAll();

    expect(response.status).toBe(200);
    expect(Array.isArray(data)).toBe(true);
    expect(data.length).toBeGreaterThanOrEqual(1);

    await cleanup();
  });

  test("user should be able to update a template", async () => {
    const api = await sharedUserClient();
    const [template, cleanup] = await useTemplate(api);

    const updateData = {
      id: template.id,
      name: "updated-template-name",
      description: "updated-description",
      notes: "updated-notes",
      defaultQuantity: 5,
      defaultInsured: true,
      defaultName: "Updated Default Name",
      defaultDescription: "Updated Default Description",
      defaultManufacturer: "Updated Manufacturer",
      defaultModelNumber: "MODEL-999",
      defaultLifetimeWarranty: true,
      defaultWarrantyDetails: "Lifetime coverage",
      defaultLocationId: "",
      defaultLabelIds: [],
      includeWarrantyFields: true,
      includePurchaseFields: true,
      includeSoldFields: false,
      fields: [],
    };

    const { response, data } = await api.templates.update(template.id, updateData);
    expect(response.status).toBe(200);
    expect(data.id).toBe(template.id);

    // Ensure the template was updated
    const { response: getResponse, data: getData } = await api.templates.get(template.id);
    expect(getResponse.status).toBe(200);
    expect(getData.name).toBe(updateData.name);
    expect(getData.description).toBe(updateData.description);
    expect(getData.notes).toBe(updateData.notes);
    expect(getData.defaultQuantity).toBe(updateData.defaultQuantity);
    expect(getData.defaultInsured).toBe(updateData.defaultInsured);
    expect(getData.defaultName).toBe(updateData.defaultName);
    expect(getData.defaultDescription).toBe(updateData.defaultDescription);
    expect(getData.defaultManufacturer).toBe(updateData.defaultManufacturer);
    expect(getData.defaultModelNumber).toBe(updateData.defaultModelNumber);
    expect(getData.defaultLifetimeWarranty).toBe(updateData.defaultLifetimeWarranty);
    expect(getData.includeWarrantyFields).toBe(updateData.includeWarrantyFields);
    expect(getData.includePurchaseFields).toBe(updateData.includePurchaseFields);

    await cleanup();
  });

  test("user should be able to delete a template", async () => {
    const api = await sharedUserClient();
    const [template, _] = await useTemplate(api);

    const { response } = await api.templates.delete(template.id);
    expect(response.status).toBe(204);

    // Ensure we can't get the template
    const { response: getResponse } = await api.templates.get(template.id);
    expect(getResponse.status).toBe(404);
  });

  test("user should be able to create a template with custom fields", async () => {
    const api = await sharedUserClient();
    const NIL_UUID = "00000000-0000-0000-0000-000000000000";

    const templateData = factories.template();
    templateData.fields = [
      { id: NIL_UUID, name: "Custom Field 1", type: "text", textValue: "Value 1" },
      { id: NIL_UUID, name: "Custom Field 2", type: "text", textValue: "Value 2" },
    ];

    const { response, data } = await api.templates.create(templateData);

    expect(response.status).toBe(201);
    expect(data.fields).toHaveLength(2);
    expect(data.fields![0]!.name).toBe("Custom Field 1");
    expect(data.fields![0]!.textValue).toBe("Value 1");
    expect(data.fields![1]!.name).toBe("Custom Field 2");
    expect(data.fields![1]!.textValue).toBe("Value 2");

    // Cleanup
    const { response: deleteResponse } = await api.templates.delete(data.id);
    expect(deleteResponse.status).toBe(204);
  });

  test("user should be able to update template custom fields", async () => {
    const api = await sharedUserClient();
    const NIL_UUID = "00000000-0000-0000-0000-000000000000";

    // Create template with a field
    const templateData = factories.template();
    templateData.fields = [{ id: NIL_UUID, name: "Original Field", type: "text", textValue: "Original Value" }];

    const { response: createResponse, data: createdTemplate } = await api.templates.create(templateData);
    expect(createResponse.status).toBe(201);
    expect(createdTemplate.fields).toHaveLength(1);

    // Update with modified and new fields
    const updateData = {
      id: createdTemplate.id,
      name: createdTemplate.name,
      description: createdTemplate.description,
      notes: createdTemplate.notes,
      defaultQuantity: createdTemplate.defaultQuantity,
      defaultInsured: createdTemplate.defaultInsured,
      defaultName: createdTemplate.defaultName,
      defaultDescription: createdTemplate.defaultDescription,
      defaultManufacturer: createdTemplate.defaultManufacturer,
      defaultModelNumber: createdTemplate.defaultModelNumber,
      defaultLifetimeWarranty: createdTemplate.defaultLifetimeWarranty,
      defaultWarrantyDetails: createdTemplate.defaultWarrantyDetails,
      defaultLocationId: "",
      defaultLabelIds: [],
      includeWarrantyFields: createdTemplate.includeWarrantyFields,
      includePurchaseFields: createdTemplate.includePurchaseFields,
      includeSoldFields: createdTemplate.includeSoldFields,
      fields: [
        { id: createdTemplate.fields![0]!.id, name: "Updated Field", type: "text", textValue: "Updated Value" },
        { id: NIL_UUID, name: "New Field", type: "text", textValue: "New Value" },
      ],
    };

    const { response: updateResponse, data: updatedTemplate } = await api.templates.update(
      createdTemplate.id,
      updateData
    );
    expect(updateResponse.status).toBe(200);
    expect(updatedTemplate.fields).toHaveLength(2);

    // Cleanup
    const { response: deleteResponse } = await api.templates.delete(createdTemplate.id);
    expect(deleteResponse.status).toBe(204);
  });
});

describe("templates with location and labels", () => {
  test("user should be able to create a template with a default location", async () => {
    const api = await sharedUserClient();

    // First create a location
    const locationData = factories.location();
    const { response: locResponse, data: location } = await api.locations.create(locationData);
    expect(locResponse.status).toBe(201);

    // Create template with the location
    const templateData = factories.template();
    templateData.defaultLocationId = location.id;

    const { response, data } = await api.templates.create(templateData);

    expect(response.status).toBe(201);
    expect(data.defaultLocation).toBeTruthy();
    expect(data.defaultLocation?.id).toBe(location.id);
    expect(data.defaultLocation?.name).toBe(location.name);

    // Cleanup
    await api.templates.delete(data.id);
    await api.locations.delete(location.id);
  });

  test("user should be able to create a template with default labels", async () => {
    const api = await sharedUserClient();

    // First create some labels
    const { response: label1Response, data: label1 } = await api.labels.create(factories.label());
    expect(label1Response.status).toBe(201);

    const { response: label2Response, data: label2 } = await api.labels.create(factories.label());
    expect(label2Response.status).toBe(201);

    // Create template with labels
    const templateData = factories.template();
    templateData.defaultLabelIds = [label1.id, label2.id];

    const { response, data } = await api.templates.create(templateData);

    expect(response.status).toBe(201);
    expect(data.defaultLabels).toHaveLength(2);
    expect(data.defaultLabels.map(l => l.id)).toContain(label1.id);
    expect(data.defaultLabels.map(l => l.id)).toContain(label2.id);

    // Cleanup
    await api.templates.delete(data.id);
    await api.labels.delete(label1.id);
    await api.labels.delete(label2.id);
  });

  test("user should be able to update template to remove location", async () => {
    const api = await sharedUserClient();

    // Create a location
    const { response: locResponse, data: location } = await api.locations.create(factories.location());
    expect(locResponse.status).toBe(201);

    // Create template with location
    const templateData = factories.template();
    templateData.defaultLocationId = location.id;

    const { response: createResponse, data: template } = await api.templates.create(templateData);
    expect(createResponse.status).toBe(201);
    expect(template.defaultLocation).toBeTruthy();

    // Update to remove location
    const updateData = {
      id: template.id,
      name: template.name,
      description: template.description,
      notes: template.notes,
      defaultQuantity: template.defaultQuantity,
      defaultInsured: template.defaultInsured,
      defaultName: template.defaultName,
      defaultDescription: template.defaultDescription,
      defaultManufacturer: template.defaultManufacturer,
      defaultModelNumber: template.defaultModelNumber,
      defaultLifetimeWarranty: template.defaultLifetimeWarranty,
      defaultWarrantyDetails: template.defaultWarrantyDetails,
      defaultLocationId: "",
      defaultLabelIds: [],
      includeWarrantyFields: template.includeWarrantyFields,
      includePurchaseFields: template.includePurchaseFields,
      includeSoldFields: template.includeSoldFields,
      fields: [],
    };

    const { response: updateResponse, data: updated } = await api.templates.update(template.id, updateData);
    expect(updateResponse.status).toBe(200);
    expect(updated.defaultLocation).toBeNull();

    // Cleanup
    await api.templates.delete(template.id);
    await api.locations.delete(location.id);
  });
});
