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
    return this.authURL(route(`/label-templates/${id}/preview`));
  }

  async render(id: string, body: LabelTemplateRenderRequest): Promise<Blob> {
    const response = await fetch(this.authURL(route(`/label-templates/${id}/render`)), {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
      credentials: "include",
    });

    if (!response.ok) {
      throw new Error(`Failed to render label: ${response.statusText}`);
    }

    return response.blob();
  }

  print(id: string, body: Omit<LabelTemplatePrintRequest, "printerId"> & { printerId?: string }) {
    return this.http.post<LabelTemplatePrintRequest, LabelTemplatePrintResponse>({
      url: route(`/label-templates/${id}/print`),
      body: body as LabelTemplatePrintRequest,
    });
  }

  async renderLocations(id: string, body: LabelTemplateRenderLocationsRequest): Promise<Blob> {
    const response = await fetch(this.authURL(route(`/label-templates/${id}/render-locations`)), {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
      credentials: "include",
    });

    if (!response.ok) {
      throw new Error(`Failed to render location label: ${response.statusText}`);
    }

    return response.blob();
  }

  printLocations(id: string, body: Omit<LabelTemplatePrintLocationsRequest, "printerId"> & { printerId?: string }) {
    return this.http.post<LabelTemplatePrintLocationsRequest, LabelTemplatePrintResponse>({
      url: route(`/label-templates/${id}/print-locations`),
      body: body as LabelTemplatePrintLocationsRequest,
    });
  }
}
