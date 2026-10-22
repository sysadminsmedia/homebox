import { BaseAPI, route } from "../base";
import type {
  PrinterCreate,
  PrinterOut,
  PrinterSummary,
  PrinterUpdate,
  PrinterStatusResponse,
  PrinterTestResponse,
  BrotherMediaInfo,
} from "../types/data-contracts";

export class PrintersApi extends BaseAPI {
  getAll() {
    return this.http.get<PrinterSummary[]>({ url: route("/printers") });
  }

  create(body: PrinterCreate) {
    return this.http.post<PrinterCreate, PrinterOut>({ url: route("/printers"), body });
  }

  get(id: string) {
    return this.http.get<PrinterOut>({ url: route(`/printers/${id}`) });
  }

  delete(id: string) {
    return this.http.delete<void>({ url: route(`/printers/${id}`) });
  }

  update(id: string, body: PrinterUpdate) {
    return this.http.put<PrinterUpdate, PrinterOut>({ url: route(`/printers/${id}`), body });
  }

  setDefault(id: string) {
    return this.http.post<void, void>({ url: route(`/printers/${id}/set-default`), body: undefined });
  }

  getStatus(id: string) {
    return this.http.get<PrinterStatusResponse>({ url: route(`/printers/${id}/status`) });
  }

  testPrint(id: string) {
    return this.http.post<void, PrinterTestResponse>({ url: route(`/printers/${id}/test`), body: undefined });
  }

  getMediaTypes() {
    return this.http.get<BrotherMediaInfo[]>({ url: route("/printers/media-types") });
  }
}
