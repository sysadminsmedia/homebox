import { BaseAPI, route } from "../base";
import type {
  MaintenanceEntry,
  MaintenancePlan,
  MaintenancePlanUpdate,
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

  updatePlan(id: string, data: MaintenancePlanUpdate) {
    return this.http.put<MaintenancePlanUpdate, MaintenancePlan>({
      url: route(`/maintenance/plans/${id}`),
      body: data,
    });
  }

  deletePlan(id: string) {
    return this.http.delete<void>({ url: route(`/maintenance/plans/${id}`) });
  }
}
