import { BaseAPI } from "./base";
import { ItemsApi } from "./classes/items";
import { TagsApi } from "./classes/tags";
import { EntityTypesApi } from "./classes/entity-types";
import { GroupApi } from "./classes/group";
import { UserApi } from "./classes/users";
import { ActionsAPI } from "./classes/actions";
import { StatsAPI } from "./classes/stats";
import { AssetsApi } from "./classes/assets";
import { ReportsAPI } from "./classes/reports";
import { NotifiersAPI } from "./classes/notifiers";
import { MaintenanceAPI } from "./classes/maintenance";
import { ProductAPI } from "./classes/product";
import { TemplatesApi } from "./classes/templates";
import type { Requests } from "~~/lib/requests";

export class UserClient extends BaseAPI {
  tags: TagsApi;
  items: ItemsApi;
  templates: TemplatesApi;
  entityTypes: EntityTypesApi;
  maintenance: MaintenanceAPI;
  group: GroupApi;
  user: UserApi;
  actions: ActionsAPI;
  stats: StatsAPI;
  assets: AssetsApi;
  reports: ReportsAPI;
  notifiers: NotifiersAPI;
  products: ProductAPI;

  /** Backward-compat shim that delegates to the entities (items) API. */
  locations: {
    getAll: InstanceType<typeof ItemsApi>["getLocations"];
    getTree: InstanceType<typeof ItemsApi>["getTree"];
    create: InstanceType<typeof ItemsApi>["createLocation"];
    get: InstanceType<typeof ItemsApi>["getLocation"];
    delete: InstanceType<typeof ItemsApi>["deleteLocation"];
    update: InstanceType<typeof ItemsApi>["updateLocation"];
  };

  constructor(requests: Requests, attachmentToken: string) {
    super(requests, attachmentToken);

    this.tags = new TagsApi(requests);
    this.items = new ItemsApi(requests, attachmentToken);
    this.templates = new TemplatesApi(requests);
    this.entityTypes = new EntityTypesApi(requests);
    this.maintenance = new MaintenanceAPI(requests);
    this.group = new GroupApi(requests);
    this.user = new UserApi(requests);
    this.actions = new ActionsAPI(requests);
    this.stats = new StatsAPI(requests);
    this.assets = new AssetsApi(requests);
    this.reports = new ReportsAPI(requests);
    this.notifiers = new NotifiersAPI(requests);
    this.products = new ProductAPI(requests);

    // Backward-compat shim: api.locations.* delegates to api.items.*
    this.locations = {
      getAll: this.items.getLocations.bind(this.items),
      getTree: this.items.getTree.bind(this.items),
      create: this.items.createLocation.bind(this.items),
      get: this.items.getLocation.bind(this.items),
      delete: this.items.deleteLocation.bind(this.items),
      update: this.items.updateLocation.bind(this.items),
    };

    Object.freeze(this);
  }
}
