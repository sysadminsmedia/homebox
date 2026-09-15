import { BaseAPI, route } from "../base";
import type { ExportOut, ResultsRepoExportOut } from "../types/data-contracts";

/**
 * Re-export so consumers only need to import from this module. The shape is
 * generated from the Go `repo.ExportOut` struct via swagger.
 */
export type CollectionExport = ExportOut;

/**
 * Client for the collection backup/restore endpoints. Always group-scoped:
 * the server reads the tenant from the auth token and refuses to act on
 * anything that doesn't belong to it.
 */
export class BackupsAPI extends BaseAPI {
  /** Kick off a new export. Returns the pending job row. */
  startExport() {
    return this.http.post<null, ExportOut>({
      url: route("/group/exports"),
    });
  }

  /** List every export job for the current group, newest first. */
  list() {
    return this.http.get<ResultsRepoExportOut>({
      url: route("/group/exports"),
    });
  }

  /** Fetch a single export job. */
  get(id: string) {
    return this.http.get<ExportOut>({
      url: route(`/group/exports/${id}`),
    });
  }

  /** Delete a job row and its blob artifact. */
  delete(id: string) {
    return this.http.delete<void>({
      url: route(`/group/exports/${id}`),
    });
  }

  /** Returns the URL to download the artifact directly. */
  downloadURL(id: string) {
    return route(`/group/exports/${id}/download`);
  }

  /**
   * Upload a previously-produced export zip and enqueue an import job. The
   * destination group must be empty; the server returns 409 otherwise.
   */
  importZip(file: File | Blob) {
    const formData = new FormData();
    formData.append("file", file);
    return this.http.post<FormData, void>({
      url: route("/group/import"),
      data: formData,
    });
  }
}
