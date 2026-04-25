import { BaseAPI, route } from "../base";
import type {
  EntityCreate,
  EntityListResult,
  EntityOut,
  EntityPatch,
  EntityPath,
  EntitySummary,
  EntityUpdate,
  ItemAttachmentUpdate,
  MaintenanceEntry,
  MaintenanceEntryCreate,
  MaintenanceEntryWithDetails,
  TreeItem,
} from "../types/data-contracts";
import type { AttachmentTypes } from "../types/non-generated";
import type { MaintenanceFilters } from "./maintenance.ts";
import type { Requests } from "~~/lib/requests";

export type ItemsQuery = {
  orderBy?: string;
  includeArchived?: boolean;
  page?: number;
  pageSize?: number;
  parentIds?: string[];
  tags?: string[];
  negateTags?: boolean;
  onlyWithoutPhoto?: boolean;
  onlyWithPhoto?: boolean;
  q?: string;
  fields?: string[];
};

export type LocationsQuery = {
  filterChildren: boolean;
};

export type TreeQuery = {
  withItems: boolean;
};

export class AttachmentsAPI extends BaseAPI {
  add(id: string, file: File | Blob, filename: string, type: AttachmentTypes | null = null, primary?: boolean) {
    const formData = new FormData();
    formData.append("file", file);
    if (type) {
      formData.append("type", type);
    }
    formData.append("name", filename);
    if (primary !== undefined) {
      formData.append("primary", primary.toString());
    }

    return this.http.post<FormData, EntityOut>({
      url: route(`/entities/${id}/attachments`),
      data: formData,
    });
  }

  delete(id: string, attachmentId: string) {
    return this.http.delete<void>({ url: route(`/entities/${id}/attachments/${attachmentId}`) });
  }

  update(id: string, attachmentId: string, data: ItemAttachmentUpdate) {
    return this.http.put<ItemAttachmentUpdate, EntityOut>({
      url: route(`/entities/${id}/attachments/${attachmentId}`),
      body: data,
    });
  }
}

export class FieldsAPI extends BaseAPI {
  getAll() {
    return this.http.get<string[]>({ url: route("/entities/fields") });
  }

  getAllValues(field: string) {
    return this.http.get<string[]>({ url: route(`/entities/fields/values`, { field }) });
  }
}

export class ItemMaintenanceAPI extends BaseAPI {
  getLog(itemId: string, filters: MaintenanceFilters = {}) {
    return this.http.get<MaintenanceEntryWithDetails[]>({
      url: route(`/entities/${itemId}/maintenance`, { status: filters.status?.toString() }),
    });
  }

  create(itemId: string, data: MaintenanceEntryCreate) {
    return this.http.post<MaintenanceEntryCreate, MaintenanceEntry>({
      url: route(`/entities/${itemId}/maintenance`),
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
    return this.http.get<EntityPath[]>({ url: route(`/entities/${id}/path`) });
  }

  async getAll(q: ItemsQuery = {}) {
    const payload = await this.http.get<EntityListResult>({ url: route("/entities", q) });
    return payload;
  }

  async create(item: EntityCreate) {
    const payload = await this.http.post<EntityCreate, EntityOut>({ url: route("/entities"), body: item });
    return payload;
  }

  async get(id: string) {
    return this.http.get<EntityOut>({ url: route(`/entities/${id}`) });
  }

  delete(id: string) {
    return this.http.delete<void>({ url: route(`/entities/${id}`) });
  }

  async update(id: string, item: EntityUpdate) {
    return this.http.put<EntityCreate, EntityOut>({
      url: route(`/entities/${id}`),
      body: this.dropFields(item),
    });
  }

  async patch(id: string, item: EntityPatch) {
    return this.http.patch<EntityPatch, EntityOut>({
      url: route(`/entities/${id}`),
      body: this.dropFields(item),
    });
  }

  async duplicate(
    id: string,
    options: {
      copyMaintenance?: boolean;
      copyAttachments?: boolean;
      copyCustomFields?: boolean;
      copyPrefix?: string;
    } = {}
  ) {
    const payload = await this.http.post<typeof options, EntityOut>({
      url: route(`/entities/${id}/duplicate`),
      body: {
        copyMaintenance: options.copyMaintenance,
        copyAttachments: options.copyAttachments,
        copyCustomFields: options.copyCustomFields,
        copyPrefix: options.copyPrefix,
      },
    });
    return payload;
  }

  import(file: File | Blob) {
    const formData = new FormData();
    formData.append("csv", file);

    return this.http.post<FormData, void>({
      url: route("/entities/import"),
      data: formData,
    });
  }

  exportURL(tenant?: string) {
    if (tenant) {
      return route("/entities/export", { tenant });
    }

    return route("/entities/export");
  }

  // =========================================================================
  // Location / Container methods (formerly in LocationsApi)
  // =========================================================================

  async getLocations(q: LocationsQuery = { filterChildren: false }) {
    const resp = await this.http.get<{ items: EntitySummary[] }>({
      url: route("/entities", { ...q, isLocation: true }),
    });
    // Unwrap paginated response to flat array for backward compat
    return {
      ...resp,
      data: resp.data?.items ?? [],
    } as { data: EntitySummary[]; error: any; status: number };
  }

  getTree(tq: TreeQuery = { withItems: false }) {
    return this.http.get<TreeItem[]>({ url: route("/entities/tree", tq) });
  }

  createLocation(body: EntityCreate) {
    return this.http.post<EntityCreate, EntityOut>({ url: route("/entities"), body });
  }

  getLocation(id: string) {
    return this.http.get<EntityOut>({ url: route(`/entities/${id}`) });
  }

  deleteLocation(id: string) {
    return this.http.delete<void>({ url: route(`/entities/${id}`) });
  }

  updateLocation(id: string, body: EntityUpdate) {
    return this.http.put<EntityUpdate, EntityOut>({ url: route(`/entities/${id}`), body });
  }
}
