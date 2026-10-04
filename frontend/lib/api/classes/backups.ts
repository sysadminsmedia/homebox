import { BaseAPI, route } from "../base";
import type {
  BackupDestinationOut,
  BackupInput,
  BackupOAuthStartOut,
  BackupOptions,
  ExportOut,
  ResultsRepoBackupDestinationOut,
  ResultsRepoExportOut,
  TestResult,
} from "../types/data-contracts";

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

  // ---------------------------------------------------------------------------
  // Scheduled backups and their destinations (group owners only)
  // ---------------------------------------------------------------------------

  /** What this server allows (feature switch, local destinations, custom endpoints). */
  options() {
    return this.http.get<BackupOptions>({ url: route("/group/backup-options") });
  }

  listDestinations() {
    return this.http.get<ResultsRepoBackupDestinationOut>({ url: route("/group/backup-destinations") });
  }

  createDestination(data: BackupInput) {
    return this.http.post<BackupInput, BackupDestinationOut>({
      url: route("/group/backup-destinations"),
      body: data,
    });
  }

  updateDestination(id: string, data: BackupInput) {
    return this.http.put<BackupInput, BackupDestinationOut>({
      url: route(`/group/backup-destinations/${id}`),
      body: data,
    });
  }

  deleteDestination(id: string) {
    return this.http.delete<void>({ url: route(`/group/backup-destinations/${id}`) });
  }

  /** Test a saved destination; the result is also recorded as its health. */
  testDestination(id: string) {
    return this.http.post<null, TestResult>({ url: route(`/group/backup-destinations/${id}/test`) });
  }

  /** Test settings that have not been saved yet. */
  testSettings(data: BackupInput) {
    return this.http.post<BackupInput, TestResult>({ url: route("/group/backup-destinations/test"), body: data });
  }

  /** Start an on-demand backup to a destination. */
  runDestination(id: string) {
    return this.http.post<null, ExportOut>({ url: route(`/group/backup-destinations/${id}/run`) });
  }

  destinationVersions(id: string) {
    return this.http.get<ResultsRepoExportOut>({ url: route(`/group/backup-destinations/${id}/versions`) });
  }

  /** Begin connecting a cloud drive. Open the returned URL in a popup. */
  startOAuth(provider: "google" | "microsoft" | "dropbox", useLoginAccount = false) {
    return this.http.post<{ provider: string; useLoginAccount: boolean }, BackupOAuthStartOut>({
      url: route("/group/backup-oauth/start"),
      body: { provider, useLoginAccount },
    });
  }
}
