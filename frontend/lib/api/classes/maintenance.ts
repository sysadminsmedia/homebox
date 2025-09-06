import { BaseAPI, route } from "../base";
import type {
  MaintenanceEntry,
  MaintenanceEntryUpdate,
  MaintenanceEntryWithDetails,
  MaintenanceFilterStatus,
} from "../types/data-contracts";

export interface MaintenanceFilters {
  status?: MaintenanceFilterStatus;
}

export class MaintenanceAPI extends BaseAPI {
  getAll(filters: MaintenanceFilters) {
    return this.http.get<MaintenanceEntryWithDetails[]>({
      url: route(`/maintenance`, { status: filters.status?.toString() }),
    });
  }

  delete(id: string) {
    return this.http.delete<void>({ url: route(`/maintenance/${id}`) });
  }

  update(id: string, data: MaintenanceEntryUpdate) {
    return this.http.put<MaintenanceEntryUpdate, MaintenanceEntry>({
      url: route(`/maintenance/${id}`),
      body: data,
    });
  }
}
