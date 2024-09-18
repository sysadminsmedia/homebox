import { BaseAPI, route } from "../base";
import type { MaintenanceEntry, MaintenanceEntryUpdate } from "../types/data-contracts";

export enum MaintenancesFilter {
  Scheduled = "scheduled",
  Completed = "completed",
  Both = "both",
}

export class MaintenanceAPI extends BaseAPI {
  getAll(filter: MaintenancesFilter) {
    return this.http.get<MaintenanceEntry[]>({ url: route(`/maintenances`, { filter }) });
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
