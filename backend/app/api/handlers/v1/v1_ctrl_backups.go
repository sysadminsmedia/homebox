package v1

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"html/template"
	"net/http"
	"strings"

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
	RemoteEnabled        bool `json:"remoteEnabled"`
	// OAuthProviders lists the configured cloud drives: google, microsoft, dropbox.
	OAuthProviders []string `json:"oauthProviders"`
	// OIDCSuggestion offers the cloud drive matching the identity provider the
	// current user signed in with, when there is one.
	OIDCSuggestion *services.OIDCSuggestion `json:"oidcSuggestion,omitempty" extensions:"x-nullable"`
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

// denyDemoBackupChange refuses destination changes, connection tests and OAuth starts in demo mode.
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
		ctx := services.NewContext(r.Context())
		var offer *services.OIDCSuggestion
		if ctx.User != nil {
			offer = ctrl.svc.Backups.OIDCSuggestionFor(ctx.User.OidcIssuer, ctx.User.Email)
		}
		return BackupOptions{OIDCSuggestion: offer, Enabled: o.Enabled, LocalEnabled: o.LocalEnabled, AllowCustomEndpoints: o.AllowCustomEndpoints, RemoteEnabled: o.RemoteEnabled, OAuthProviders: o.OAuthProviders}, nil
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
//	@Param			payload	body		repo.BackupInput	true	"Destination settings"
//	@Success		201		{object}	repo.BackupDestinationOut
//	@Router			/v1/group/backup-destinations [POST]
//	@Security		Bearer
func (ctrl *V1Controller) HandleBackupDestinationCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, in repo.BackupInput) (repo.BackupDestinationOut, error) {
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
//	@Param		payload	body		repo.BackupInput	true	"Destination settings"
//	@Success	200		{object}	repo.BackupDestinationOut
//	@Router		/v1/group/backup-destinations/{id} [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleBackupDestinationUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID, in repo.BackupInput) (repo.BackupDestinationOut, error) {
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
//	@Param		payload	body		repo.BackupInput	true	"Destination settings"
//	@Success	200		{object}	services.TestResult
//	@Router		/v1/group/backup-destinations/test [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleBackupSettingsTest() errchain.HandlerFunc {
	fn := func(r *http.Request, in repo.BackupInput) (services.TestResult, error) {
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

// BackupOAuthStartIn selects the cloud provider to connect.
type BackupOAuthStartIn struct {
	Provider string `json:"provider" validate:"required,oneof=google microsoft dropbox"`
	// UseLoginAccount asks the provider to preselect the account the user
	// signed in to Homebox with. Honoured only when that login came from the
	// same provider.
	UseLoginAccount bool `json:"useLoginAccount"`
}

// BackupOAuthStartOut is where to send the user to authorize access.
type BackupOAuthStartOut struct {
	AuthURL string `json:"authUrl"`
}

const backupOAuthCallbackPath = "/api/v1/group/backup-oauth/callback"

// oauthRedirectURI is where the provider returns the user. It must match the
// redirect URI registered with the OAuth app, so a forged Host header cannot
// make a provider send a code anywhere that was not registered.
func (ctrl *V1Controller) oauthRedirectURI(r *http.Request) string {
	base := SecureBaseURL(r, &ctrl.config.Options)
	if base == "" {
		base = GetHBURL(r, &ctrl.config.Options, ctrl.url)
	}
	if base == "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		base = scheme + "://" + r.Host
	}
	return strings.TrimSuffix(base, "/") + backupOAuthCallbackPath
}

// HandleBackupOAuthStart godoc
//
//	@Summary		Start Connecting a Cloud Drive
//	@Description	Returns the provider's authorization URL. Open it in a popup; the callback page reports the result to the opener window.
//	@Tags			Backups
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		v1.BackupOAuthStartIn	true	"Provider"
//	@Success		200		{object}	v1.BackupOAuthStartOut
//	@Router			/v1/group/backup-oauth/start [POST]
//	@Security		Bearer
func (ctrl *V1Controller) HandleBackupOAuthStart() errchain.HandlerFunc {
	fn := func(r *http.Request, in BackupOAuthStartIn) (BackupOAuthStartOut, error) {
		if err := ctrl.denyDemoBackupChange(); err != nil {
			return BackupOAuthStartOut{}, err
		}
		ctx := services.NewContext(r.Context())
		var hint string
		if in.UseLoginAccount && ctx.User != nil {
			// The hint comes from the server's record of the user, never from the
			// request, and only when their login matches the provider.
			if offer := ctrl.svc.Backups.OIDCSuggestionFor(ctx.User.OidcIssuer, ctx.User.Email); offer != nil && offer.Provider == in.Provider {
				hint = offer.Email
			}
		}
		u, err := ctrl.svc.Backups.OAuthStart(ctx.GID, ctx.UID, in.Provider, ctrl.oauthRedirectURI(r), hint)
		if err != nil {
			return BackupOAuthStartOut{}, backupError(err)
		}
		return BackupOAuthStartOut{AuthURL: u}, nil
	}
	return adapters.Action(fn, http.StatusOK)
}

var backupOAuthPage = template.Must(template.New("oauth").Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Homebox</title></head>
<body style="font-family:system-ui,sans-serif;max-width:32rem;margin:3rem auto;padding:0 1rem">
<h1>{{if .OK}}Account connected{{else}}Connection failed{{end}}</h1>
<p>{{if .OK}}{{.Account}} is connected. You can close this window.{{else}}{{.Error}}{{end}}</p>
<script nonce="{{.Nonce}}">
(function () {
  var m = {{.Msg}};
  // The app is served with Cross-Origin-Opener-Policy: same-origin, which cuts
  // window.opener once the popup has visited the provider. A same-origin
  // BroadcastChannel is unaffected, so it is the primary channel.
  try { var ch = new BroadcastChannel("homebox-backup-oauth"); ch.postMessage(m); ch.close(); } catch (e) {}
  try { if (window.opener) { window.opener.postMessage(m, window.location.origin); } } catch (e) {}
  if (m.ok) { setTimeout(function () { window.close(); }, 800); }
})();
</script></body></html>`))

// HandleBackupOAuthCallback godoc
//
//	@Summary		Cloud Drive Authorization Callback
//	@Description	The redirect target for cloud providers. Public: access is authorized by the single-use state created when the flow started.
//	@Tags			Backups
//	@Produce		html
//	@Param			state	query	string	false	"State"
//	@Param			code	query	string	false	"Authorization code"
//	@Param			error	query	string	false	"Provider error"
//	@Success		200
//	@Router			/v1/group/backup-oauth/callback [GET]
func (ctrl *V1Controller) HandleBackupOAuthCallback() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		q := r.URL.Query()
		out := ctrl.svc.Backups.OAuthCallback(r.Context(), ctrl.oauthRedirectURI(r), q.Get("state"), q.Get("code"), q.Get("error"))

		nb := make([]byte, 16)
		_, _ = rand.Read(nb)
		nonce := base64.RawURLEncoding.EncodeToString(nb)

		h := w.Header()
		h.Set("Content-Type", "text/html; charset=utf-8")
		h.Set("Cache-Control", "no-store")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Content-Security-Policy", "default-src 'none'; script-src 'nonce-"+nonce+"'; style-src 'unsafe-inline'; base-uri 'none'; form-action 'none'")

		msg := struct {
			Type string `json:"type"`
			services.OAuthOutcome
		}{Type: "homebox-backup-oauth", OAuthOutcome: out}
		return backupOAuthPage.Execute(w, map[string]any{
			"OK": out.OK, "Account": out.Account, "Error": out.Error, "Nonce": nonce, "Msg": msg,
		})
	}
}
