import { BaseAPI, route } from "../base";

export class ReportsAPI extends BaseAPI {
  billOfMaterialsURL(tenant?: string): string {
    if (tenant) {
      return route("/reporting/bill-of-materials", { tenant });
    }

    return route("/reporting/bill-of-materials");
  }
}
