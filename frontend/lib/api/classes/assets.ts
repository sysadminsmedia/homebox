import { BaseAPI, route } from "../base";
import type { EntitySummary } from "../types/data-contracts";
import type { PaginationResult } from "../types/non-generated";

export class AssetsApi extends BaseAPI {
  async get(id: string, page = 1, pageSize = 50) {
    return await this.http.get<PaginationResult<EntitySummary>>({
      url: route(`/assets/${id}`, { page, pageSize }),
    });
  }
}
