import { BaseAPI, route } from "../base";
import type { CustomsInventoryPreview, CustomsInventoryRequest } from "../types/data-contracts";
import { fallbackExportFilename, filenameFromDisposition, serverErrorMessage } from "~~/lib/customs-inventory";

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
      // Rejected requests carry the reason, e.g. a scope that is too large.
      throw new Error(serverErrorMessage(resp.data) || resp.response.statusText || String(resp.status));
    }
    const headers = resp.response.headers;
    const filename = filenameFromDisposition(
      headers.get("Content-Disposition"),
      fallbackExportFilename(headers.get("Content-Type"))
    );
    return { blob: await resp.response.blob(), filename };
  }
}
