import { BaseAPI, route } from "../base";
import { parseDate } from "../base/base-api";
import type {
  EntityPath,
  ItemAttachmentUpdate,
  ItemCreate,
  ItemOut,
  ItemPatch,
  ItemSummary,
  ItemUpdate,
  LocationOut,
  LocationOutCount,
  LocationUpdate,
  MaintenanceEntry,
  MaintenanceEntryCreate,
  MaintenanceEntryWithDetails,
  TreeItem,
} from "../types/data-contracts";
import type { AttachmentTypes, ItemSummaryPaginationResult } from "../types/non-generated";
import type { MaintenanceFilters } from "./maintenance.ts";
import type { Requests } from "~~/lib/requests";

/**
 * Maps the backend `parent` field to `location` for backward compatibility.
 * Many frontend components still reference `item.location` which is the same as `item.parent`.
 */
function mapParentToLocation<T extends { parent?: any; location?: any }>(item: T): T {
  if (item.parent && !item.location) {
    item.location = item.parent;
  }
  return item;
}

function mapSummaryCompat(item: ItemSummary): ItemSummary {
  return mapParentToLocation(item);
}

function mapOutCompat(item: ItemOut): ItemOut {
  mapParentToLocation(item);
  if ("syncChildEntityLocations" in item) {
    item.syncChildItemsLocations = (item as any).syncChildEntityLocations;
  }
  return item;
}

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

    return this.http.post<FormData, ItemOut>({
      url: route(`/entities/${id}/attachments`),
      data: formData,
    });
  }

  delete(id: string, attachmentId: string) {
    return this.http.delete<void>({ url: route(`/entities/${id}/attachments/${attachmentId}`) });
  }

  update(id: string, attachmentId: string, data: ItemAttachmentUpdate) {
    return this.http.put<ItemAttachmentUpdate, ItemOut>({
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
    const payload = await this.http.get<ItemSummaryPaginationResult<ItemSummary>>({ url: route("/entities", q) });
    if (payload.data?.items) {
      payload.data.items = payload.data.items.map(mapSummaryCompat);
    }
    return payload;
  }

  async create(item: ItemCreate) {
    const payload = await this.http.post<ItemCreate, ItemOut>({ url: route("/entities"), body: item });
    if (payload.data) {
      payload.data = mapOutCompat(payload.data);
    }
    return payload;
  }

  async get(id: string) {
    const payload = await this.http.get<ItemOut>({ url: route(`/entities/${id}`) });

    if (!payload.data) {
      return payload;
    }

    // Map parent -> location for backward compat
    payload.data = mapOutCompat(payload.data);

    // Parse Date Types
    payload.data = parseDate(payload.data, ["purchaseTime", "soldTime", "warrantyExpires"]);
    return payload;
  }

  delete(id: string) {
    return this.http.delete<void>({ url: route(`/entities/${id}`) });
  }

  async update(id: string, item: ItemUpdate) {
    const payload = await this.http.put<ItemCreate, ItemOut>({
      url: route(`/entities/${id}`),
      body: this.dropFields(item),
    });
    if (!payload.data) {
      return payload;
    }

    payload.data = mapOutCompat(payload.data);
    payload.data = parseDate(payload.data, ["purchaseTime", "soldTime", "warrantyExpires"]);
    return payload;
  }

  async patch(id: string, item: ItemPatch) {
    const resp = await this.http.patch<ItemPatch, ItemOut>({
      url: route(`/entities/${id}`),
      body: this.dropFields(item),
    });

    if (!resp.data) {
      return resp;
    }

    resp.data = mapOutCompat(resp.data);
    resp.data = parseDate(resp.data, ["purchaseTime", "soldTime", "warrantyExpires"]);
    return resp;
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
    const payload = await this.http.post<typeof options, ItemOut>({
      url: route(`/entities/${id}/duplicate`),
      body: {
        copyMaintenance: options.copyMaintenance,
        copyAttachments: options.copyAttachments,
        copyCustomFields: options.copyCustomFields,
        copyPrefix: options.copyPrefix,
      },
    });
    if (payload.data) {
      payload.data = mapOutCompat(payload.data);
    }
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

  getLocations(q: LocationsQuery = { filterChildren: false }) {
    return this.http.get<LocationOutCount[]>({ url: route("/entities", { ...q, isLocation: true }) });
  }

  getTree(tq: TreeQuery = { withItems: false }) {
    return this.http.get<TreeItem[]>({ url: route("/entities/tree", tq) });
  }

  createLocation(body: ItemCreate) {
    return this.http.post<ItemCreate, LocationOut>({ url: route("/entities"), body });
  }

  getLocation(id: string) {
    return this.http.get<LocationOut>({ url: route(`/entities/${id}`) });
  }

  deleteLocation(id: string) {
    return this.http.delete<void>({ url: route(`/entities/${id}`) });
  }

  updateLocation(id: string, body: LocationUpdate) {
    return this.http.put<LocationUpdate, LocationOut>({ url: route(`/entities/${id}`), body });
  }
}
