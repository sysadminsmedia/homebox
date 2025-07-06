import { BaseAPI, route } from "../base";
import type { ActionAmountResult, ItemCreate } from "../types/data-contracts";

export class ActionsAPI extends BaseAPI {
  ensureAssetIDs() {
    return this.http.post<void, ActionAmountResult>({
      url: route("/actions/ensure-asset-ids"),
    });
  }

  resetItemDateTimes() {
    return this.http.post<void, ActionAmountResult>({
      url: route("/actions/zero-item-time-fields"),
    });
  }

  ensureImportRefs() {
    return this.http.post<void, ActionAmountResult>({
      url: route("/actions/ensure-import-refs"),
    });
  }

  setPrimaryPhotos() {
    return this.http.post<void, ActionAmountResult>({
      url: route("/actions/set-primary-photos"),
    });
  }

  createMissingThumbnails() {
    return this.http.post<void, ActionAmountResult>({
      url: route("/actions/create-missing-thumbnails"),
    });
  }
  
  getEAN(productEAN: string) {
    return this.http.get<ItemCreate>({ url: route(`/getproductfromean`, { productEAN }) });
  }
}
