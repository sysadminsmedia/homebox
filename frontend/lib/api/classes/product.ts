import { BaseAPI, route } from "../base";
import type { ItemCreate } from "../types/data-contracts";

export class ProductAPI extends BaseAPI {
  searchFromBarcode(productEAN: string) {
    return this.http.get<ItemCreate>({ url: route(`/products/search-from-barcode`, { productEAN }) });
  }
}