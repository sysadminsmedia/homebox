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

  setImage(id: string, file: File | Blob, filename: string) {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("name", filename);

    return this.http.post<FormData, EntityTemplateOut>({
      url: route(`/templates/${id}/image`),
      data: formData,
    });
  }

  deleteImage(id: string) {
    return this.http.delete<EntityTemplateOut>({ url: route(`/templates/${id}/image`) });
  }

  createItem(templateId: string, body: EntityTemplateCreateItemRequest) {
    return this.http.post<EntityTemplateCreateItemRequest, EntityOut>({
      url: route(`/templates/${templateId}/create-item`),
      body,
    });
  }
}
