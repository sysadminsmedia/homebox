import { BaseAPI, route } from "../base";
import type { LocationCreate, LocationOut, LocationOutCount, LocationUpdate, TreeItem } from "../types/data-contracts";

export type LocationsQuery = {
  filterChildren: boolean;
};

export type TreeQuery = {
  withItems: boolean;
};

export class LocationsApi extends BaseAPI {
  getAll(q: LocationsQuery = { filterChildren: false }) {
    return this.http.get<LocationOutCount[]>({ url: route("/locations", q) });
  }

  getTree(tq = { withItems: false }) {
    return this.http.get<TreeItem[]>({ url: route("/locations/tree", tq) });
  }

  create(body: LocationCreate) {
    return this.http.post<LocationCreate, LocationOut>({ url: route("/locations"), body });
  }

  get(id: string) {
    return this.http.get<LocationOut>({ url: route(`/locations/${id}`) });
  }

  delete(id: string) {
    return this.http.delete<void>({ url: route(`/locations/${id}`) });
  }

  update(id: string, body: LocationUpdate) {
    return this.http.put<LocationUpdate, LocationOut>({ url: route(`/locations/${id}`), body });
  }
}
