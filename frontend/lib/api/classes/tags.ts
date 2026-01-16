import { BaseAPI, route } from "../base";
import type { TagCreate, TagOut } from "../types/data-contracts";

export class TagsApi extends BaseAPI {
  getAll() {
    return this.http.get<TagOut[]>({ url: route("/tags") });
  }

  create(body: TagCreate) {
    return this.http.post<TagCreate, TagOut>({ url: route("/tags"), body });
  }

  get(id: string) {
    return this.http.get<TagOut>({ url: route(`/tags/${id}`) });
  }

  delete(id: string) {
    return this.http.delete<void>({ url: route(`/tags/${id}`) });
  }

  update(id: string, body: TagCreate) {
    return this.http.put<TagCreate, TagOut>({ url: route(`/tags/${id}`), body });
  }
}
