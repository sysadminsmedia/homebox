import { BaseAPI, route } from "../base";
import type { CustomsInventoryPreview, CustomsInventoryRequest } from "../types/data-contracts";
import { filenameFromDisposition } from "~~/lib/customs-inventory";

export class ReportsAPI extends BaseAPI {
  billOfMaterialsURL(tenant?: string): string {
    if (tenant) {
      return route("/reporting/bill-of-materials", { tenant });
    }

    return route("/reporting/bill-of-materials");
  }

  customsInventoryPreview(body: CustomsInventoryRequest) {
    return this.http.post<CustomsInventoryRequest, CustomsInventoryPreview>({
      url: route("/reporting/customs-inventory/preview"),
      body,
    });
  }

  async customsInventoryExport(body: CustomsInventoryRequest): Promise<{ blob: Blob; filename: string }> {
    const resp = await this.http.post<CustomsInventoryRequest, unknown>({
      url: route("/reporting/customs-inventory/export"),
      body,
    });
    if (resp.error) {
      throw new Error(`Failed to export customs inventory: ${resp.response.statusText || resp.status}`);
    }
    const filename = filenameFromDisposition(resp.response.headers.get("Content-Disposition"), "Inventarliste.zip");
    return { blob: await resp.response.blob(), filename };
  }
}
