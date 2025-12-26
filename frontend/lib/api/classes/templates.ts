import { BaseAPI, route } from "../base";
import type {
  ItemTemplateCreate,
  ItemTemplateOut,
  ItemTemplateSummary,
  ItemTemplateUpdate,
  ItemTemplateCreateItemRequest,
  ItemOut,
} from "../types/data-contracts";

export class TemplatesApi extends BaseAPI {
  getAll() {
    return this.http.get<ItemTemplateSummary[]>({ url: route("/templates") });
  }

  create(body: ItemTemplateCreate) {
    return this.http.post<ItemTemplateCreate, ItemTemplateOut>({ url: route("/templates"), body });
  }

  get(id: string) {
    return this.http.get<ItemTemplateOut>({ url: route(`/templates/${id}`) });
  }

  delete(id: string) {
    return this.http.delete<void>({ url: route(`/templates/${id}`) });
  }

  update(id: string, body: ItemTemplateUpdate) {
    return this.http.put<ItemTemplateUpdate, ItemTemplateOut>({ url: route(`/templates/${id}`), body });
  }

  createItem(templateId: string, body: ItemTemplateCreateItemRequest) {
    return this.http.post<ItemTemplateCreateItemRequest, ItemOut>({
      url: route(`/templates/${templateId}/create-item`),
      body,
    });
  }
}
