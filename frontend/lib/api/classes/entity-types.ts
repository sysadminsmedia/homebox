import { BaseAPI, route } from "../base";
import type { EntityTypeCreate, EntityTypeSummary, EntityTypeUpdate } from "../types/data-contracts";

export class EntityTypesApi extends BaseAPI {
  getAll() {
    return this.http.get<EntityTypeSummary[]>({ url: route("/entity-types") });
  }

  create(body: EntityTypeCreate) {
    return this.http.post<EntityTypeCreate, EntityTypeSummary>({ url: route("/entity-types"), body });
  }

  update(id: string, body: EntityTypeUpdate) {
    return this.http.put<EntityTypeUpdate, EntityTypeSummary>({ url: route(`/entity-types/${id}`), body });
  }

  delete(id: string) {
    return this.http.delete<void>({ url: route(`/entity-types/${id}`) });
  }
}
