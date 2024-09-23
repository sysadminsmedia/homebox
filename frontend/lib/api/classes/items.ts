import { BaseAPI, route } from "../base";
import { parseDate } from "../base/base-api";
import type {
  ItemAttachmentUpdate,
  ItemCreate,
  ItemOut,
  ItemPatch,
  ItemPath,
  ItemSummary,
  ItemUpdate,
  MaintenanceEntry,
  MaintenanceEntryCreate,
  MaintenanceLog,
} from "../types/data-contracts";
import type { AttachmentTypes, PaginationResult } from "../types/non-generated";
import type { Requests } from "~~/lib/requests";

export type ItemsQuery = {
  orderBy?: string;
  includeArchived?: boolean;
  page?: number;
  pageSize?: number;
  locations?: string[];
  labels?: string[];
  negateLabels?: boolean;
  parentIds?: string[];
  q?: string;
  fields?: string[];
};

export class AttachmentsAPI extends BaseAPI {
  add(id: string, file: File | Blob, filename: string, type: AttachmentTypes | null = null) {
    const formData = new FormData();
    formData.append("file", file);
    if (type) {
      formData.append("type", type);
    }
    formData.append("name", filename);

    return this.http.post<FormData, ItemOut>({
      url: route(`/items/${id}/attachments`),
      data: formData,
    });
  }

  delete(id: string, attachmentId: string) {
    return this.http.delete<void>({ url: route(`/items/${id}/attachments/${attachmentId}`) });
  }

  update(id: string, attachmentId: string, data: ItemAttachmentUpdate) {
    return this.http.put<ItemAttachmentUpdate, ItemOut>({
      url: route(`/items/${id}/attachments/${attachmentId}`),
      body: data,
    });
  }
}

export class FieldsAPI extends BaseAPI {
  getAll() {
    return this.http.get<string[]>({ url: route("/items/fields") });
  }

  getAllValues(field: string) {
    return this.http.get<string[]>({ url: route(`/items/fields/values`, { field }) });
  }
}

type MaintenanceEntryQuery = {
  scheduled?: boolean;
  completed?: boolean;
};

export class ItemMaintenanceAPI extends BaseAPI {
  getLog(itemId: string, q: MaintenanceEntryQuery = {}) {
    return this.http.get<MaintenanceLog>({ url: route(`/items/${itemId}/maintenance`, q) });
  }

  create(itemId: string, data: MaintenanceEntryCreate) {
    return this.http.post<MaintenanceEntryCreate, MaintenanceEntry>({
      url: route(`/items/${itemId}/maintenance`),
      body: data,
    });
  }
}

export class ItemsApi extends BaseAPI {
  attachments: AttachmentsAPI;
  maintenance: ItemMaintenanceAPI;
  fields: FieldsAPI;

  constructor(http: Requests, token: string) {
    super(http, token);
    this.fields = new FieldsAPI(http);
    this.attachments = new AttachmentsAPI(http);
    this.maintenance = new ItemMaintenanceAPI(http);
  }

  fullpath(id: string) {
    return this.http.get<ItemPath[]>({ url: route(`/items/${id}/path`) });
  }

  getAll(q: ItemsQuery = {}) {
    return this.http.get<PaginationResult<ItemSummary>>({ url: route("/items", q) });
  }

  create(item: ItemCreate) {
    return this.http.post<ItemCreate, ItemOut>({ url: route("/items"), body: item });
  }

  async get(id: string) {
    const payload = await this.http.get<ItemOut>({ url: route(`/items/${id}`) });

    if (!payload.data) {
      return payload;
    }

    // Parse Date Types
    payload.data = parseDate(payload.data, ["purchaseTime", "soldTime", "warrantyExpires"]);
    return payload;
  }

  delete(id: string) {
    return this.http.delete<void>({ url: route(`/items/${id}`) });
  }

  async update(id: string, item: ItemUpdate) {
    const payload = await this.http.put<ItemCreate, ItemOut>({
      url: route(`/items/${id}`),
      body: this.dropFields(item),
    });
    if (!payload.data) {
      return payload;
    }

    payload.data = parseDate(payload.data, ["purchaseTime", "soldTime", "warrantyExpires"]);
    return payload;
  }

  async patch(id: string, item: ItemPatch) {
    const resp = await this.http.patch<ItemPatch, ItemOut>({
      url: route(`/items/${id}`),
      body: this.dropFields(item),
    });

    if (!resp.data) {
      return resp;
    }

    resp.data = parseDate(resp.data, ["purchaseTime", "soldTime", "warrantyExpires"]);
    return resp;
  }

  import(file: File | Blob) {
    const formData = new FormData();
    formData.append("csv", file);

    return this.http.post<FormData, void>({
      url: route("/items/import"),
      data: formData,
    });
  }

  exportURL() {
    return route("/items/export");
  }
}
