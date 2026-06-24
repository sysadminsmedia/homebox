package v1

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"gocloud.dev/blob"

	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

// HandleExportsList godoc
//
//	@Summary		List Collection Exports
//	@Description	Returns export job rows for the caller's group, newest first.
//	@Tags			Group
//	@Produce		json
//	@Success		200	{object}	Results[repo.ExportOut]
//	@Router			/v1/group/exports [GET]
//	@Security		Bearer
func (ctrl *V1Controller) HandleExportsList() errchain.HandlerFunc {
	fn := func(r *http.Request) (Results[repo.ExportOut], error) {
		ctx := services.NewContext(r.Context())
		rows, err := ctrl.repo.Exports.ListByGroup(ctx, ctx.GID)
		if err != nil {
			return Results[repo.ExportOut]{}, err
		}
		return WrapResults(rows), nil
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandleExportsCreate godoc
//
//	@Summary		Start a Collection Export
//	@Description	Creates a pending export row and enqueues the build job. Poll the listing endpoint or watch the WebSocket for completion.
//	@Tags			Group
//	@Produce		json
//	@Success		202	{object}	repo.ExportOut
//	@Router			/v1/group/exports [POST]
//	@Security		Bearer
func (ctrl *V1Controller) HandleExportsCreate() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := services.NewContext(r.Context())
		out, err := ctrl.svc.Exports.Enqueue(ctx, ctx.GID)
		if err != nil {
			log.Err(err).Msg("failed to enqueue export")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		return server.JSON(w, http.StatusAccepted, out)
	}
}

// HandleExportGet godoc
//
//	@Summary	Get an Export
//	@Tags		Group
//	@Produce	json
//	@Param		id	path		string	true	"Export ID"
//	@Success	200	{object}	repo.ExportOut
//	@Router		/v1/group/exports/{id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleExportGet() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID) (repo.ExportOut, error) {
		ctx := services.NewContext(r.Context())
		return ctrl.repo.Exports.Get(ctx, ctx.GID, id)
	}

	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandleExportDownload godoc
//
//	@Summary	Download an Export Artifact
//	@Tags		Group
//	@Produce	application/zip
//	@Param		id	path	string	true	"Export ID"
//	@Success	200	{file}	file
//	@Router		/v1/group/exports/{id}/download [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleExportDownload() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := services.NewContext(r.Context())
		id, err := ctrl.routeID(r)
		if err != nil {
			return err
		}
		out, err := ctrl.repo.Exports.Get(ctx, ctx.GID, id)
		if err != nil {
			if ent.IsNotFound(err) {
				return validate.NewRequestError(err, http.StatusNotFound)
			}
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		if out.Status != "completed" || out.ArtifactPath == "" {
			return validate.NewRequestError(errors.New("export not ready"), http.StatusConflict)
		}
		// Defence in depth: refuse to stream anything that doesn't live under the
		// caller's group prefix. The repo Get above already enforces ownership;
		// this catches a stale row whose artifact_path was tampered with.
		expectedPrefix := ctx.GID.String() + "/exports/"
		if !strings.HasPrefix(out.ArtifactPath, expectedPrefix) {
			return validate.NewRequestError(errors.New("artifact outside group prefix"), http.StatusForbidden)
		}

		bucket, err := blob.OpenBucket(r.Context(), ctrl.repo.Attachments.GetConnString())
		if err != nil {
			log.Err(err).Msg("export download: open bucket")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		defer func() { _ = bucket.Close() }()

		reader, err := bucket.NewReader(r.Context(), ctrl.repo.Attachments.GetFullPath(out.ArtifactPath), nil)
		if err != nil {
			log.Err(err).Str("artifact_path", out.ArtifactPath).Msg("export download: open reader")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		defer func() { _ = reader.Close() }()

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition",
			fmt.Sprintf(`attachment; filename="homebox-export-%s.zip"`, out.ID.String()))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", out.SizeBytes))
		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, reader)
		return err
	}
}

// HandleExportDelete godoc
//
//	@Summary		Delete an Export
//	@Description	Deletes the export row and its blob artifact.
//	@Tags			Group
//	@Param			id	path	string	true	"Export ID"
//	@Success		204
//	@Router			/v1/group/exports/{id} [DELETE]
//	@Security		Bearer
func (ctrl *V1Controller) HandleExportDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID) (any, error) {
		ctx := services.NewContext(r.Context())
		out, err := ctrl.repo.Exports.Get(ctx, ctx.GID, id)
		if err != nil {
			// Idempotent: a missing row is already in the desired state.
			// Swallow here so the adapter writes 204 instead of letting the
			// error middleware translate ent.IsNotFound into a 404.
			if ent.IsNotFound(err) {
				return nil, nil
			}
			return nil, err
		}
		if out.ArtifactPath != "" {
			// Defence in depth: only touch blobs that live under the caller's
			// group prefix. The repo Get above already enforces ownership; this
			// catches a stale row whose artifact_path was tampered with, and
			// path.Clean collapses any traversal segments before the prefix
			// check so "<gid>/exports/../../other" can't slip through.
			cleanPath := path.Clean(out.ArtifactPath)
			expectedPrefix := ctx.GID.String() + "/exports/"
			if strings.HasPrefix(cleanPath, expectedPrefix) {
				bucket, err := blob.OpenBucket(r.Context(), ctrl.repo.Attachments.GetConnString())
				if err == nil {
					_ = bucket.Delete(r.Context(), ctrl.repo.Attachments.GetFullPath(cleanPath))
					_ = bucket.Close()
				}
			}
		}
		if _, err := ctrl.repo.Exports.Delete(ctx, ctx.GID, id); err != nil {
			return nil, err
		}
		return nil, nil
	}

	return adapters.CommandID("id", fn, http.StatusNoContent)
}

// HandleCollectionImport godoc
//
//	@Summary		Import a Collection Zip
//	@Description	Uploads a collection-export zip and enqueues the import job. The destination group must be empty. Returns the tracked import row so clients can poll for progress.
//	@Tags			Group
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"Export zip"
//	@Success		202		{object}	repo.ExportOut
//	@Router			/v1/group/import [POST]
//	@Security		Bearer
func (ctrl *V1Controller) HandleCollectionImport() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if ctrl.isDemo {
			return validate.NewRequestError(errors.New("import is not allowed in demo mode"), http.StatusForbidden)
		}

		ctx := services.NewContext(r.Context())

		isOwner, err := ctrl.repo.Groups.IsOwnerOf(ctx, ctx.UID, ctx.GID)
		if err != nil {
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		if !isOwner {
			return validate.NewRequestError(errors.New("only group owners can import"), http.StatusForbidden)
		}

		// Precondition: no items yet. Default seeded locations/tags are fine —
		// the worker wipes them as part of the restore. Front-loading the
		// check here gives instant 409 feedback for clearly-bad attempts.
		ready, err := ctrl.svc.Exports.IsGroupReadyForImport(r.Context(), ctx.GID)
		if err != nil {
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		if !ready {
			return validate.NewRequestError(
				errors.New("import requires a collection with no user-created items, tags, templates, notifiers, or custom types"),
				http.StatusConflict)
		}

		// maxImportSize is in MB and applies to the whole request body via the
		// path-aware middleware; here we pass `maxParseMemory` to ParseMultipartForm
		// as the memory-vs-disk threshold so larger archives spool gracefully.
		if err := r.ParseMultipartForm(ctrl.maxParseMemory << 20); err != nil {
			log.Err(err).Msg("import: parse multipart")
			return validate.NewRequestError(err, http.StatusBadRequest)
		}
		// Remove any spooled temp files the multipart parser may have created.
		// Registered before file.Close so the close (LIFO) runs first — on
		// Windows os.Remove fails while the handle is still open.
		defer func() {
			if r.MultipartForm != nil {
				_ = r.MultipartForm.RemoveAll()
			}
		}()
		file, _, err := r.FormFile("file")
		if err != nil {
			return validate.NewRequestError(err, http.StatusBadRequest)
		}
		defer func() { _ = file.Close() }()

		// Stage to {gid}/imports/{uuid}.zip in blob storage. Using the gid
		// prefix makes it impossible for one tenant's uploads to collide with
		// another's, and the worker enforces the same prefix as a safety net.
		uploadID := uuid.New()
		uploadKey := fmt.Sprintf("%s/imports/%s.zip", ctx.GID.String(), uploadID.String())

		bucket, err := blob.OpenBucket(r.Context(), ctrl.repo.Attachments.GetConnString())
		if err != nil {
			log.Err(err).Msg("import: open bucket")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		defer func() { _ = bucket.Close() }()

		bw, err := bucket.NewWriter(r.Context(), ctrl.repo.Attachments.GetFullPath(uploadKey),
			&blob.WriterOptions{ContentType: "application/zip"})
		if err != nil {
			log.Err(err).Msg("import: open writer")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		// io.Copy returns the staged byte count; we record it on the import
		// row so the UI can render "X MB queued" before the worker starts.
		uploadSize, err := io.Copy(bw, file)
		if err != nil {
			_ = bw.Close()
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		if err := bw.Close(); err != nil {
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		row, err := ctrl.svc.Exports.EnqueueImport(r.Context(), ctx.GID, ctx.UID, uploadKey, uploadSize)
		if err != nil {
			// Best-effort cleanup of the staged upload if we couldn't enqueue.
			_ = bucket.Delete(r.Context(), ctrl.repo.Attachments.GetFullPath(uploadKey))
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		return server.JSON(w, http.StatusAccepted, row)
	}
}
