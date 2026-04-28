package v1

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"gocloud.dev/blob"

	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
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
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := services.NewContext(r.Context())
		rows, err := ctrl.repo.Exports.ListByGroup(ctx, ctx.GID)
		if err != nil {
			log.Err(err).Msg("failed to list exports")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		return server.JSON(w, http.StatusOK, WrapResults(rows))
	}
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
//	@Summary		Get an Export
//	@Tags			Group
//	@Produce		json
//	@Param			id	path	string	true	"Export ID"
//	@Success		200	{object}	repo.ExportOut
//	@Router			/v1/group/exports/{id} [GET]
//	@Security		Bearer
func (ctrl *V1Controller) HandleExportGet() errchain.HandlerFunc {
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
		return server.JSON(w, http.StatusOK, out)
	}
}

// HandleExportDownload godoc
//
//	@Summary		Download an Export Artifact
//	@Tags			Group
//	@Produce		application/zip
//	@Param			id	path		string	true	"Export ID"
//	@Success		200	{file}		file
//	@Router			/v1/group/exports/{id}/download [GET]
//	@Security		Bearer
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
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := services.NewContext(r.Context())
		id, err := ctrl.routeID(r)
		if err != nil {
			return err
		}
		out, err := ctrl.repo.Exports.Get(ctx, ctx.GID, id)
		if err != nil {
			if ent.IsNotFound(err) {
				w.WriteHeader(http.StatusNoContent)
				return nil
			}
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		if out.ArtifactPath != "" {
			bucket, err := blob.OpenBucket(r.Context(), ctrl.repo.Attachments.GetConnString())
			if err == nil {
				_ = bucket.Delete(r.Context(), ctrl.repo.Attachments.GetFullPath(out.ArtifactPath))
				_ = bucket.Close()
			}
		}
		if _, err := ctrl.repo.Exports.Delete(ctx, ctx.GID, id); err != nil {
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}

// HandleCollectionImport godoc
//
//	@Summary		Import a Collection Zip
//	@Description	Uploads a collection-export zip and enqueues the import job. The destination group must be empty.
//	@Tags			Group
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"Export zip"
//	@Success		202
//	@Router			/v1/group/import [POST]
//	@Security		Bearer
func (ctrl *V1Controller) HandleCollectionImport() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if ctrl.isDemo {
			return validate.NewRequestError(errors.New("import is not allowed in demo mode"), http.StatusForbidden)
		}

		ctx := services.NewContext(r.Context())
		if !ctx.User.IsOwner {
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
				errors.New("import requires a collection with no items"),
				http.StatusConflict)
		}

		if err := r.ParseMultipartForm(ctrl.maxUploadSize << 20); err != nil {
			log.Err(err).Msg("import: parse multipart")
			return validate.NewRequestError(err, http.StatusBadRequest)
		}
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
		if _, err := io.Copy(bw, file); err != nil {
			_ = bw.Close()
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		if err := bw.Close(); err != nil {
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		if err := ctrl.svc.Exports.EnqueueImport(r.Context(), ctx.GID, ctx.UID, uploadKey); err != nil {
			// Best-effort cleanup of the staged upload if we couldn't enqueue.
			_ = bucket.Delete(r.Context(), ctrl.repo.Attachments.GetFullPath(uploadKey))
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusAccepted)
		return nil
	}
}
