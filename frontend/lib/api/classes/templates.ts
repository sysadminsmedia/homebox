import { BaseAPI, route } from "../base";
import type {
  EntityTemplateCreate,
  EntityTemplateOut,
  EntityTemplateSummary,
  EntityTemplateUpdate,
  EntityTemplateCreateItemRequest,
  EntityOut,
} from "../types/data-contracts";

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
