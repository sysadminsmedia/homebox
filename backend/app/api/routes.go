package main

import (
	"embed"
	"errors"
	"io"
	"mime"
	"net/http"
	"path"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/hay-kot/httpkit/errchain"
	httpSwagger "github.com/swaggo/http-swagger/v2" // http-swagger middleware
	"github.com/sysadminsmedia/homebox/backend/app/api/handlers/debughandlers"
	v1 "github.com/sysadminsmedia/homebox/backend/app/api/handlers/v1"
	"github.com/sysadminsmedia/homebox/backend/app/api/providers"
	_ "github.com/sysadminsmedia/homebox/backend/app/api/static/docs"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/authroles"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
)

const prefix = "/api"

var (
	ErrDir = errors.New("path is dir")

	//go:embed all:static/public/*
	public embed.FS
)

func (a *app) debugRouter() *http.ServeMux {
	dbg := http.NewServeMux()
	debughandlers.New(dbg)

	return dbg
}

// registerRoutes registers all the routes for the API
func (a *app) mountRoutes(r *chi.Mux, chain *errchain.ErrChain, repos *repo.AllRepos) {
	registerMimes()

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// =========================================================================
	// API Version 1

	v1Base := v1.BaseURLFunc(prefix)

	v1Ctrl := v1.NewControllerV1(
		a.services,
		a.repos,
		a.bus,
		v1.WithMaxUploadSize(a.conf.Web.MaxUploadSize),
		v1.WithRegistration(a.conf.Options.AllowRegistration),
		v1.WithDemoStatus(a.conf.Demo), // Disable Password Change in Demo Mode
	)

	r.Get(v1Base("/status"), chain.ToHandlerFunc(v1Ctrl.HandleBase(func() bool { return true }, v1.Build{
		Version:   version,
		Commit:    commit,
		BuildTime: buildTime,
	})))

	r.Get(v1Base("/currencies"), chain.ToHandlerFunc(v1Ctrl.HandleCurrency()))

	providers := []v1.AuthProvider{
		providers.NewLocalProvider(a.services.User),
	}

	r.Post(v1Base("/users/register"), chain.ToHandlerFunc(v1Ctrl.HandleUserRegistration()))
	r.Post(v1Base("/users/login"), chain.ToHandlerFunc(v1Ctrl.HandleAuthLogin(providers...)))

	userMW := []errchain.Middleware{
		a.mwAuthToken,
		a.mwRoles(RoleModeOr, authroles.RoleUser.String()),
	}

	r.Get(v1Base("/ws/events"), chain.ToHandlerFunc(v1Ctrl.HandleCacheWS(), userMW...))
	r.Get(v1Base("/users/self"), chain.ToHandlerFunc(v1Ctrl.HandleUserSelf(), userMW...))
	r.Put(v1Base("/users/self"), chain.ToHandlerFunc(v1Ctrl.HandleUserSelfUpdate(), userMW...))
	r.Delete(v1Base("/users/self"), chain.ToHandlerFunc(v1Ctrl.HandleUserSelfDelete(), userMW...))
	r.Post(v1Base("/users/logout"), chain.ToHandlerFunc(v1Ctrl.HandleAuthLogout(), userMW...))
	r.Get(v1Base("/users/refresh"), chain.ToHandlerFunc(v1Ctrl.HandleAuthRefresh(), userMW...))
	r.Put(v1Base("/users/self/change-password"), chain.ToHandlerFunc(v1Ctrl.HandleUserSelfChangePassword(), userMW...))

	r.Post(v1Base("/groups/invitations"), chain.ToHandlerFunc(v1Ctrl.HandleGroupInvitationsCreate(), userMW...))
	r.Get(v1Base("/groups/statistics"), chain.ToHandlerFunc(v1Ctrl.HandleGroupStatistics(), userMW...))
	r.Get(v1Base("/groups/statistics/purchase-price"), chain.ToHandlerFunc(v1Ctrl.HandleGroupStatisticsPriceOverTime(), userMW...))
	r.Get(v1Base("/groups/statistics/locations"), chain.ToHandlerFunc(v1Ctrl.HandleGroupStatisticsLocations(), userMW...))
	r.Get(v1Base("/groups/statistics/labels"), chain.ToHandlerFunc(v1Ctrl.HandleGroupStatisticsLabels(), userMW...))

	// TODO: I don't like /groups being the URL for users
	r.Get(v1Base("/groups"), chain.ToHandlerFunc(v1Ctrl.HandleGroupGet(), userMW...))
	r.Put(v1Base("/groups"), chain.ToHandlerFunc(v1Ctrl.HandleGroupUpdate(), userMW...))

	r.Post(v1Base("/actions/ensure-asset-ids"), chain.ToHandlerFunc(v1Ctrl.HandleEnsureAssetID(), userMW...))
	r.Post(v1Base("/actions/zero-item-time-fields"), chain.ToHandlerFunc(v1Ctrl.HandleItemDateZeroOut(), userMW...))
	r.Post(v1Base("/actions/ensure-import-refs"), chain.ToHandlerFunc(v1Ctrl.HandleEnsureImportRefs(), userMW...))
	r.Post(v1Base("/actions/set-primary-photos"), chain.ToHandlerFunc(v1Ctrl.HandleSetPrimaryPhotos(), userMW...))

	r.Get(v1Base("/locations"), chain.ToHandlerFunc(v1Ctrl.HandleLocationGetAll(), userMW...))
	r.Post(v1Base("/locations"), chain.ToHandlerFunc(v1Ctrl.HandleLocationCreate(), userMW...))
	r.Get(v1Base("/locations/tree"), chain.ToHandlerFunc(v1Ctrl.HandleLocationTreeQuery(), userMW...))
	r.Get(v1Base("/locations/{id}"), chain.ToHandlerFunc(v1Ctrl.HandleLocationGet(), userMW...))
	r.Put(v1Base("/locations/{id}"), chain.ToHandlerFunc(v1Ctrl.HandleLocationUpdate(), userMW...))
	r.Delete(v1Base("/locations/{id}"), chain.ToHandlerFunc(v1Ctrl.HandleLocationDelete(), userMW...))

	r.Get(v1Base("/labels"), chain.ToHandlerFunc(v1Ctrl.HandleLabelsGetAll(), userMW...))
	r.Post(v1Base("/labels"), chain.ToHandlerFunc(v1Ctrl.HandleLabelsCreate(), userMW...))
	r.Get(v1Base("/labels/{id}"), chain.ToHandlerFunc(v1Ctrl.HandleLabelGet(), userMW...))
	r.Put(v1Base("/labels/{id}"), chain.ToHandlerFunc(v1Ctrl.HandleLabelUpdate(), userMW...))
	r.Delete(v1Base("/labels/{id}"), chain.ToHandlerFunc(v1Ctrl.HandleLabelDelete(), userMW...))

	r.Get(v1Base("/items"), chain.ToHandlerFunc(v1Ctrl.HandleItemsGetAll(), userMW...))
	r.Post(v1Base("/items"), chain.ToHandlerFunc(v1Ctrl.HandleItemsCreate(), userMW...))
	r.Post(v1Base("/items/import"), chain.ToHandlerFunc(v1Ctrl.HandleItemsImport(), userMW...))
	r.Get(v1Base("/items/export"), chain.ToHandlerFunc(v1Ctrl.HandleItemsExport(), userMW...))
	r.Get(v1Base("/items/fields"), chain.ToHandlerFunc(v1Ctrl.HandleGetAllCustomFieldNames(), userMW...))
	r.Get(v1Base("/items/fields/values"), chain.ToHandlerFunc(v1Ctrl.HandleGetAllCustomFieldValues(), userMW...))

	r.Get(v1Base("/items/{id}"), chain.ToHandlerFunc(v1Ctrl.HandleItemGet(), userMW...))
	r.Get(v1Base("/items/{id}/path"), chain.ToHandlerFunc(v1Ctrl.HandleItemFullPath(), userMW...))
	r.Put(v1Base("/items/{id}"), chain.ToHandlerFunc(v1Ctrl.HandleItemUpdate(), userMW...))
	r.Patch(v1Base("/items/{id}"), chain.ToHandlerFunc(v1Ctrl.HandleItemPatch(), userMW...))
	r.Delete(v1Base("/items/{id}"), chain.ToHandlerFunc(v1Ctrl.HandleItemDelete(), userMW...))

	r.Post(v1Base("/items/{id}/attachments"), chain.ToHandlerFunc(v1Ctrl.HandleItemAttachmentCreate(), userMW...))
	r.Put(v1Base("/items/{id}/attachments/{attachment_id}"), chain.ToHandlerFunc(v1Ctrl.HandleItemAttachmentUpdate(), userMW...))
	r.Delete(v1Base("/items/{id}/attachments/{attachment_id}"), chain.ToHandlerFunc(v1Ctrl.HandleItemAttachmentDelete(), userMW...))

	r.Get(v1Base("/items/{id}/maintenance"), chain.ToHandlerFunc(v1Ctrl.HandleMaintenanceLogGet(), userMW...))
	r.Post(v1Base("/items/{id}/maintenance"), chain.ToHandlerFunc(v1Ctrl.HandleMaintenanceEntryCreate(), userMW...))
	r.Put(v1Base("/items/{id}/maintenance/{entry_id}"), chain.ToHandlerFunc(v1Ctrl.HandleMaintenanceEntryUpdate(), userMW...))
	r.Delete(v1Base("/items/{id}/maintenance/{entry_id}"), chain.ToHandlerFunc(v1Ctrl.HandleMaintenanceEntryDelete(), userMW...))

	r.Get(v1Base("/assets/{id}"), chain.ToHandlerFunc(v1Ctrl.HandleAssetGet(), userMW...))

	// Notifiers
	r.Get(v1Base("/notifiers"), chain.ToHandlerFunc(v1Ctrl.HandleGetUserNotifiers(), userMW...))
	r.Post(v1Base("/notifiers"), chain.ToHandlerFunc(v1Ctrl.HandleCreateNotifier(), userMW...))
	r.Put(v1Base("/notifiers/{id}"), chain.ToHandlerFunc(v1Ctrl.HandleUpdateNotifier(), userMW...))
	r.Delete(v1Base("/notifiers/{id}"), chain.ToHandlerFunc(v1Ctrl.HandleDeleteNotifier(), userMW...))
	r.Post(v1Base("/notifiers/test"), chain.ToHandlerFunc(v1Ctrl.HandlerNotifierTest(), userMW...))

	// Asset-Like endpoints
	assetMW := []errchain.Middleware{
		a.mwAuthToken,
		a.mwRoles(RoleModeOr, authroles.RoleUser.String(), authroles.RoleAttachments.String()),
	}

	r.Get(
		v1Base("/qrcode"),
		chain.ToHandlerFunc(v1Ctrl.HandleGenerateQRCode(), assetMW...),
	)
	r.Get(
		v1Base("/items/{id}/attachments/{attachment_id}"),
		chain.ToHandlerFunc(v1Ctrl.HandleItemAttachmentGet(), assetMW...),
	)

	// Reporting Services
	r.Get(v1Base("/reporting/bill-of-materials"), chain.ToHandlerFunc(v1Ctrl.HandleBillOfMaterialsExport(), userMW...))

	r.NotFound(chain.ToHandlerFunc(notFoundHandler()))
}

func registerMimes() {
	err := mime.AddExtensionType(".js", "application/javascript")
	if err != nil {
		panic(err)
	}

	err = mime.AddExtensionType(".mjs", "application/javascript")
	if err != nil {
		panic(err)
	}
}

// notFoundHandler perform the main logic around handling the internal SPA embed and ensuring that
// the client side routing is handled correctly.
func notFoundHandler() errchain.HandlerFunc {
	tryRead := func(fs embed.FS, prefix, requestedPath string, w http.ResponseWriter) error {
		f, err := fs.Open(path.Join(prefix, requestedPath))
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()

		stat, _ := f.Stat()
		if stat.IsDir() {
			return ErrDir
		}

		contentType := mime.TypeByExtension(filepath.Ext(requestedPath))
		w.Header().Set("Content-Type", contentType)
		_, err = io.Copy(w, f)
		return err
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		err := tryRead(public, "static/public", r.URL.Path, w)
		if err != nil {
			// Fallback to the index.html file.
			// should succeed in all cases.
			err = tryRead(public, "static/public", "index.html", w)
			if err != nil {
				return err
			}
		}
		return nil
	}
}
