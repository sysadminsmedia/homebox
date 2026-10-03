package v1

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"

	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

// BackupOptions tells the UI which destination types this server allows.
type BackupOptions struct {
	Enabled              bool `json:"enabled"`
	LocalEnabled         bool `json:"localEnabled"`
	AllowCustomEndpoints bool `json:"allowCustomEndpoints"`
}

// backupError maps service errors onto HTTP statuses. Settings problems the
// user can fix are 400s; everything else is a 500 with the cause logged.
func backupError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, services.ErrBackupInvalid):
		return validate.NewRequestError(err, http.StatusBadRequest)
	case errors.Is(err, services.ErrBackupDisabled):
		return validate.NewRequestError(err, http.StatusForbidden)
	case ent.IsNotFound(err):
		return validate.NewRequestError(err, http.StatusNotFound)
	default:
		log.Err(err).Msg("backup destinations")
		return validate.NewRequestError(err, http.StatusInternalServerError)
	}
}

func (ctrl *V1Controller) denyDemoBackupChange() error {
	if ctrl.isDemo {
		return validate.NewRequestError(errors.New("backup destinations are not available in demo mode"), http.StatusForbidden)
	}
	return nil
}

// HandleBackupOptions godoc
//
//	@Summary	Get Backup Options
//	@Tags		Backups
//	@Produce	json
//	@Success	200	{object}	BackupOptions
//	@Router		/v1/group/backup-options [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleBackupOptions() errchain.HandlerFunc {
	fn := func(r *http.Request) (BackupOptions, error) {
		o := ctrl.svc.Backups.Options()
		return BackupOptions{Enabled: o.Enabled, LocalEnabled: o.LocalEnabled, AllowCustomEndpoints: o.AllowCustomEndpoints}, nil
	}
	return adapters.Command(fn, http.StatusOK)
}

// HandleBackupDestinationsList godoc
//
//	@Summary		List Backup Destinations
//	@Description	Returns the group's backup destinations with schedule, retention and health status. Group owners only.
//	@Tags			Backups
//	@Produce		json
//	@Success		200	{object}	Results[repo.BackupDestinationOut]
//	@Router			/v1/group/backup-destinations [GET]
//	@Security		Bearer
func (ctrl *V1Controller) HandleBackupDestinationsList() errchain.HandlerFunc {
	fn := func(r *http.Request) (Results[repo.BackupDestinationOut], error) {
		ctx := services.NewContext(r.Context())
		rows, err := ctrl.svc.Backups.ListDestinations(ctx, ctx.GID)
		if err != nil {
			return Results[repo.BackupDestinationOut]{}, backupError(err)
		}
		return WrapResults(rows), nil
	}
	return adapters.Command(fn, http.StatusOK)
}

// HandleBackupDestinationCreate godoc
//
//	@Summary		Create a Backup Destination
//	@Description	Credentials are never stored: cloud destinations read them from the server environment.
//	@Tags			Backups
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		repo.BackupSettings	true	"Destination settings"
//	@Success		201		{object}	repo.BackupDestinationOut
//	@Router			/v1/group/backup-destinations [POST]
//	@Security		Bearer
func (ctrl *V1Controller) HandleBackupDestinationCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, in repo.BackupSettings) (repo.BackupDestinationOut, error) {
		if err := ctrl.denyDemoBackupChange(); err != nil {
			return repo.BackupDestinationOut{}, err
		}
		ctx := services.NewContext(r.Context())
		out, err := ctrl.svc.Backups.CreateDestination(ctx, ctx.GID, in)
		return out, backupError(err)
	}
	return adapters.Action(fn, http.StatusCreated)
}

// HandleBackupDestinationGet godoc
//
//	@Summary	Get a Backup Destination
//	@Tags		Backups
//	@Produce	json
//	@Param		id	path		string	true	"Destination ID"
//	@Success	200	{object}	repo.BackupDestinationOut
//	@Router		/v1/group/backup-destinations/{id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleBackupDestinationGet() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID) (repo.BackupDestinationOut, error) {
		ctx := services.NewContext(r.Context())
		out, err := ctrl.svc.Backups.GetDestination(ctx, ctx.GID, id)
		return out, backupError(err)
	}
	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandleBackupDestinationUpdate godoc
//
//	@Summary	Update a Backup Destination
//	@Tags		Backups
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string				true	"Destination ID"
//	@Param		payload	body		repo.BackupSettings	true	"Destination settings"
//	@Success	200		{object}	repo.BackupDestinationOut
//	@Router		/v1/group/backup-destinations/{id} [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleBackupDestinationUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID, in repo.BackupSettings) (repo.BackupDestinationOut, error) {
		if err := ctrl.denyDemoBackupChange(); err != nil {
			return repo.BackupDestinationOut{}, err
		}
		ctx := services.NewContext(r.Context())
		out, err := ctrl.svc.Backups.UpdateDestination(ctx, ctx.GID, id, in)
		return out, backupError(err)
	}
	return adapters.ActionID("id", fn, http.StatusOK)
}

// HandleBackupDestinationDelete godoc
//
//	@Summary		Delete a Backup Destination
//	@Description	Removes the destination and its backup history. Files already written to the destination are not deleted.
//	@Tags			Backups
//	@Param			id	path	string	true	"Destination ID"
//	@Success		204
//	@Router			/v1/group/backup-destinations/{id} [DELETE]
//	@Security		Bearer
func (ctrl *V1Controller) HandleBackupDestinationDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID) (any, error) {
		if err := ctrl.denyDemoBackupChange(); err != nil {
			return nil, err
		}
		ctx := services.NewContext(r.Context())
		return nil, backupError(ctrl.svc.Backups.DeleteDestination(ctx, ctx.GID, id))
	}
	return adapters.CommandID("id", fn, http.StatusNoContent)
}

// HandleBackupDestinationTest godoc
//
//	@Summary		Test a Saved Backup Destination
//	@Description	Writes and deletes a small test object and records the result as the destination's health.
//	@Tags			Backups
//	@Produce		json
//	@Param			id	path		string	true	"Destination ID"
//	@Success		200	{object}	services.TestResult
//	@Router			/v1/group/backup-destinations/{id}/test [POST]
//	@Security		Bearer
func (ctrl *V1Controller) HandleBackupDestinationTest() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID) (services.TestResult, error) {
		if err := ctrl.denyDemoBackupChange(); err != nil {
			return services.TestResult{}, err
		}
		ctx := services.NewContext(r.Context())
		out, err := ctrl.svc.Backups.TestDestination(ctx, ctx.GID, id)
		return out, backupError(err)
	}
	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandleBackupSettingsTest godoc
//
//	@Summary	Test Unsaved Backup Destination Settings
//	@Tags		Backups
//	@Accept		json
//	@Produce	json
//	@Param		payload	body		repo.BackupSettings	true	"Destination settings"
//	@Success	200		{object}	services.TestResult
//	@Router		/v1/group/backup-destinations/test [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleBackupSettingsTest() errchain.HandlerFunc {
	fn := func(r *http.Request, in repo.BackupSettings) (services.TestResult, error) {
		if err := ctrl.denyDemoBackupChange(); err != nil {
			return services.TestResult{}, err
		}
		ctx := services.NewContext(r.Context())
		out, err := ctrl.svc.Backups.TestSettings(ctx, ctx.GID, in)
		return out, backupError(err)
	}
	return adapters.Action(fn, http.StatusOK)
}

// HandleBackupDestinationRun godoc
//
//	@Summary		Back Up Now
//	@Description	Starts an on-demand backup to this destination.
//	@Tags			Backups
//	@Produce		json
//	@Param			id	path		string	true	"Destination ID"
//	@Success		202	{object}	repo.ExportOut
//	@Router			/v1/group/backup-destinations/{id}/run [POST]
//	@Security		Bearer
func (ctrl *V1Controller) HandleBackupDestinationRun() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if err := ctrl.denyDemoBackupChange(); err != nil {
			return err
		}
		ctx := services.NewContext(r.Context())
		id, err := ctrl.routeID(r)
		if err != nil {
			return err
		}
		out, err := ctrl.svc.Backups.RunNow(ctx, ctx.GID, id)
		if err != nil {
			return backupError(err)
		}
		return server.JSON(w, http.StatusAccepted, out)
	}
}

// HandleBackupDestinationVersions godoc
//
//	@Summary		List Backup Versions
//	@Description	Returns the backups written to a destination, newest first.
//	@Tags			Backups
//	@Produce		json
//	@Param			id	path		string	true	"Destination ID"
//	@Success		200	{object}	Results[repo.ExportOut]
//	@Router			/v1/group/backup-destinations/{id}/versions [GET]
//	@Security		Bearer
func (ctrl *V1Controller) HandleBackupDestinationVersions() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID) (Results[repo.ExportOut], error) {
		ctx := services.NewContext(r.Context())
		rows, err := ctrl.svc.Backups.ListVersions(ctx, ctx.GID, id)
		if err != nil {
			return Results[repo.ExportOut]{}, backupError(err)
		}
		return WrapResults(rows), nil
	}
	return adapters.CommandID("id", fn, http.StatusOK)
}
