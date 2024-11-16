import { BaseAPI } from "./base";
import { ItemsApi } from "./classes/items";
import { LabelsApi } from "./classes/labels";
import { LocationsApi } from "./classes/locations";
import { GroupApi } from "./classes/group";
import { UserApi } from "./classes/users";
import { ActionsAPI } from "./classes/actions";
import { StatsAPI } from "./classes/stats";
import { AssetsApi } from "./classes/assets";
import { ReportsAPI } from "./classes/reports";
import { NotifiersAPI } from "./classes/notifiers";
import { MaintenanceAPI } from "./classes/maintenance";
import type { Requests } from "~~/lib/requests";

export class UserClient extends BaseAPI {
  locations: LocationsApi;
  labels: LabelsApi;
  items: ItemsApi;
  maintenance: MaintenanceAPI;
  group: GroupApi;
  user: UserApi;
  actions: ActionsAPI;
  stats: StatsAPI;
  assets: AssetsApi;
  reports: ReportsAPI;
  notifiers: NotifiersAPI;

  constructor(requests: Requests, attachmentToken: string) {
    super(requests, attachmentToken);

    this.locations = new LocationsApi(requests);
    this.labels = new LabelsApi(requests);
    this.items = new ItemsApi(requests, attachmentToken);
    this.maintenance = new MaintenanceAPI(requests);
    this.group = new GroupApi(requests);
    this.user = new UserApi(requests);
    this.actions = new ActionsAPI(requests);
    this.stats = new StatsAPI(requests);
    this.assets = new AssetsApi(requests);
    this.reports = new ReportsAPI(requests);
    this.notifiers = new NotifiersAPI(requests);

    Object.freeze(this);
  }
}
