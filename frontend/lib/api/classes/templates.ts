import { BaseAPI, route } from "../base";
import type {
  EntityTemplateCreate,
  EntityTemplateOut,
  EntityTemplateSummary,
  EntityTemplateUpdate,
  EntityTemplateCreateItemRequest,
  EntityOut,
  TemplateField,
} from "../types/data-contracts";

/**
 * Builds a text TemplateField for create/update requests. The UI only edits text
 * fields; the backend persists type, name and textValue and ignores the other
 * values, but timeValue must still decode as a time, so it gets Go's zero time.
 */
export function textTemplateField(field: Pick<TemplateField, "id" | "name" | "textValue">): TemplateField {
  return {
    id: field.id,
    name: field.name,
    type: "text",
    textValue: field.textValue,
    numberValue: 0,
    booleanValue: false,
    timeValue: "0001-01-01T00:00:00Z",
  };
}

export class TemplatesApi extends BaseAPI {
  getAll() {
    return this.http.get<EntityTemplateSummary[]>({ url: route("/templates") });
  }

  create(body: EntityTemplateCreate) {
    return this.http.post<EntityTemplateCreate, EntityTemplateOut>({ url: route("/templates"), body });
  }

  get(id: string) {
    return this.http.get<EntityTemplateOut>({ url: route(`/templates/${id}`) });
  }

  delete(id: string) {
    return this.http.delete<void>({ url: route(`/templates/${id}`) });
  }

  update(id: string, body: EntityTemplateUpdate) {
    return this.http.put<EntityTemplateUpdate, EntityTemplateOut>({ url: route(`/templates/${id}`), body });
  }

  createItem(templateId: string, body: EntityTemplateCreateItemRequest) {
    return this.http.post<EntityTemplateCreateItemRequest, EntityOut>({
      url: route(`/templates/${templateId}/create-item`),
      body,
    });
  }
}
