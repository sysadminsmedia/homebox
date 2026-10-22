import { BaseAPI, route } from "../base";
import type {
  LabelTemplateCreate,
  LabelTemplateOut,
  LabelTemplateSummary,
  LabelTemplateUpdate,
  LabelTemplateRenderRequest,
  LabelTemplateRenderLocationsRequest,
  LabelmakerLabelPreset,
  LabelmakerBarcodeFormatInfo,
  LabelTemplatePrintRequest,
  LabelTemplatePrintResponse,
  LabelTemplatePrintLocationsRequest,
} from "../types/data-contracts";

export class LabelTemplatesApi extends BaseAPI {
  getAll() {
    return this.http.get<LabelTemplateSummary[]>({ url: route("/label-templates") });
  }

  create(body: LabelTemplateCreate) {
    return this.http.post<LabelTemplateCreate, LabelTemplateOut>({ url: route("/label-templates"), body });
  }

  get(id: string) {
    return this.http.get<LabelTemplateOut>({ url: route(`/label-templates/${id}`) });
  }

  delete(id: string) {
    return this.http.delete<void>({ url: route(`/label-templates/${id}`) });
  }

  update(id: string, body: LabelTemplateUpdate) {
    return this.http.put<LabelTemplateUpdate, LabelTemplateOut>({ url: route(`/label-templates/${id}`), body });
  }

  duplicate(id: string) {
    return this.http.post<void, LabelTemplateOut>({ url: route(`/label-templates/${id}/duplicate`), body: undefined });
  }

  getPresets() {
    return this.http.get<LabelmakerLabelPreset[]>({ url: route("/label-templates/presets") });
  }

  getBarcodeFormats() {
    return this.http.get<LabelmakerBarcodeFormatInfo[]>({ url: route("/label-templates/barcode-formats") });
  }

  getPreviewUrl(id: string) {
    // authURL applies route() itself and appends the attachment token + tenant
    return this.authURL(`/label-templates/${id}/preview`);
  }

  // Rendering returns binary (PNG/PDF). Go through the shared Requests client so the
  // Authorization header and X-Tenant header are sent like every other API call.
  private async renderBlob<T>(url: string, body: T): Promise<Blob> {
    const resp = await this.http.post<T, unknown>({ url, body });
    if (resp.error) {
      throw new Error(`Failed to render label: ${resp.response.statusText || resp.status}`);
    }
    return resp.response.blob();
  }

  render(id: string, body: LabelTemplateRenderRequest): Promise<Blob> {
    return this.renderBlob(route(`/label-templates/${id}/render`), body);
  }

  print(id: string, body: Omit<LabelTemplatePrintRequest, "printerId"> & { printerId?: string }) {
    return this.http.post<LabelTemplatePrintRequest, LabelTemplatePrintResponse>({
      url: route(`/label-templates/${id}/print`),
      body: body as LabelTemplatePrintRequest,
    });
  }

  renderLocations(id: string, body: LabelTemplateRenderLocationsRequest): Promise<Blob> {
    return this.renderBlob(route(`/label-templates/${id}/render-locations`), body);
  }

  printLocations(id: string, body: Omit<LabelTemplatePrintLocationsRequest, "printerId"> & { printerId?: string }) {
    return this.http.post<LabelTemplatePrintLocationsRequest, LabelTemplatePrintResponse>({
      url: route(`/label-templates/${id}/print-locations`),
      body: body as LabelTemplatePrintLocationsRequest,
    });
  }
}
