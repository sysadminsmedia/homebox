import { BaseAPI, route } from "../base";
import type {
  MaintenanceEntry,
  MaintenanceEntryWithDetails,
  MaintenanceEntryUpdate,
  MaintenancesFilterStatus,
} from "../types/data-contracts";

export interface MaintenancesFilters {
  status?: MaintenancesFilterStatus;
}

export class MaintenanceAPI extends BaseAPI {
  getAll(filters: MaintenancesFilters) {
    return this.http.get<MaintenanceEntryWithDetails[]>({
      url: route(`/maintenances`, { status: filters.status?.toString() }),
    });
  }

  delete(id: string) {
    return this.http.delete<void>({ url: route(`/maintenances/${id}`) });
  }

  update(id: string, data: MaintenanceEntryUpdate) {
    return this.http.put<MaintenanceEntryUpdate, MaintenanceEntry>({
      url: route(`/maintenances/${id}`),
      body: data,
    });
  }
}
