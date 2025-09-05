package main

import (
	"embed"
	"errors"
	"fmt"
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

	v1Ctrl := v1.NewControllerV1(
		a.services,
		a.repos,
		a.bus,
		a.conf,
		v1.WithMaxUploadSize(a.conf.Web.MaxUploadSize),
		v1.WithRegistration(a.conf.Options.AllowRegistration),
		v1.WithDemoStatus(a.conf.Demo), // Disable Password Change in Demo Mode
		v1.WithURL(fmt.Sprintf("%s:%s", a.conf.Web.Host, a.conf.Web.Port)),
	)

	r.Route(prefix+"/v1", func(r chi.Router) {
		r.Get("/status", chain.ToHandlerFunc(v1Ctrl.HandleBase(func() bool { return true }, v1.Build{
			Version:   version,
			Commit:    commit,
			BuildTime: buildTime,
		})))

		r.Get("/currencies", chain.ToHandlerFunc(v1Ctrl.HandleCurrency()))

		providers := []v1.AuthProvider{
			providers.NewLocalProvider(a.services.User),
		}

		r.Post("/users/register", chain.ToHandlerFunc(v1Ctrl.HandleUserRegistration()))
		r.Post("/users/login", chain.ToHandlerFunc(v1Ctrl.HandleAuthLogin(providers...)))

		userMW := []errchain.Middleware{
			a.mwAuthToken,
			a.mwRoles(RoleModeOr, authroles.RoleUser.String()),
		}

		r.Get("/ws/events", chain.ToHandlerFunc(v1Ctrl.HandleCacheWS(), userMW...))
		r.Get("/users/self", chain.ToHandlerFunc(v1Ctrl.HandleUserSelf(), userMW...))
		r.Put("/users/self", chain.ToHandlerFunc(v1Ctrl.HandleUserSelfUpdate(), userMW...))
		r.Delete("/users/self", chain.ToHandlerFunc(v1Ctrl.HandleUserSelfDelete(), userMW...))
		r.Get("/users/self/settings", chain.ToHandlerFunc(v1Ctrl.HandleUserSelfSettingsGet(), userMW...))
		r.Put("/users/self/settings", chain.ToHandlerFunc(v1Ctrl.HandleUserSelfSettingsUpdate(), userMW...))
		r.Post("/users/logout", chain.ToHandlerFunc(v1Ctrl.HandleAuthLogout(), userMW...))
		r.Get("/users/refresh", chain.ToHandlerFunc(v1Ctrl.HandleAuthRefresh(), userMW...))
		r.Put("/users/self/change-password", chain.ToHandlerFunc(v1Ctrl.HandleUserSelfChangePassword(), userMW...))

		r.Post("/groups/invitations", chain.ToHandlerFunc(v1Ctrl.HandleGroupInvitationsCreate(), userMW...))
		r.Get("/groups/statistics", chain.ToHandlerFunc(v1Ctrl.HandleGroupStatistics(), userMW...))
		r.Get("/groups/statistics/purchase-price", chain.ToHandlerFunc(v1Ctrl.HandleGroupStatisticsPriceOverTime(), userMW...))
		r.Get("/groups/statistics/locations", chain.ToHandlerFunc(v1Ctrl.HandleGroupStatisticsLocations(), userMW...))
		r.Get("/groups/statistics/labels", chain.ToHandlerFunc(v1Ctrl.HandleGroupStatisticsLabels(), userMW...))

		// TODO: I don't like /groups being the URL for users
		r.Get("/groups", chain.ToHandlerFunc(v1Ctrl.HandleGroupGet(), userMW...))
		r.Put("/groups", chain.ToHandlerFunc(v1Ctrl.HandleGroupUpdate(), userMW...))

		r.Post("/actions/ensure-asset-ids", chain.ToHandlerFunc(v1Ctrl.HandleEnsureAssetID(), userMW...))
		r.Post("/actions/zero-item-time-fields", chain.ToHandlerFunc(v1Ctrl.HandleItemDateZeroOut(), userMW...))
		r.Post("/actions/ensure-import-refs", chain.ToHandlerFunc(v1Ctrl.HandleEnsureImportRefs(), userMW...))
		r.Post("/actions/set-primary-photos", chain.ToHandlerFunc(v1Ctrl.HandleSetPrimaryPhotos(), userMW...))
		r.Post("/actions/create-missing-thumbnails", chain.ToHandlerFunc(v1Ctrl.HandleCreateMissingThumbnails(), userMW...))

		r.Get("/locations", chain.ToHandlerFunc(v1Ctrl.HandleLocationGetAll(), userMW...))
		r.Post("/locations", chain.ToHandlerFunc(v1Ctrl.HandleLocationCreate(), userMW...))
		r.Get("/locations/tree", chain.ToHandlerFunc(v1Ctrl.HandleLocationTreeQuery(), userMW...))
		r.Get("/locations/{id}", chain.ToHandlerFunc(v1Ctrl.HandleLocationGet(), userMW...))
		r.Put("/locations/{id}", chain.ToHandlerFunc(v1Ctrl.HandleLocationUpdate(), userMW...))
		r.Delete("/locations/{id}", chain.ToHandlerFunc(v1Ctrl.HandleLocationDelete(), userMW...))

		r.Get("/labels", chain.ToHandlerFunc(v1Ctrl.HandleLabelsGetAll(), userMW...))
		r.Post("/labels", chain.ToHandlerFunc(v1Ctrl.HandleLabelsCreate(), userMW...))
		r.Get("/labels/{id}", chain.ToHandlerFunc(v1Ctrl.HandleLabelGet(), userMW...))
		r.Put("/labels/{id}", chain.ToHandlerFunc(v1Ctrl.HandleLabelUpdate(), userMW...))
		r.Delete("/labels/{id}", chain.ToHandlerFunc(v1Ctrl.HandleLabelDelete(), userMW...))

		r.Get("/items", chain.ToHandlerFunc(v1Ctrl.HandleItemsGetAll(), userMW...))
		r.Post("/items", chain.ToHandlerFunc(v1Ctrl.HandleItemsCreate(), userMW...))
		r.Post("/items/import", chain.ToHandlerFunc(v1Ctrl.HandleItemsImport(), userMW...))
		r.Get("/items/export", chain.ToHandlerFunc(v1Ctrl.HandleItemsExport(), userMW...))
		r.Get("/items/fields", chain.ToHandlerFunc(v1Ctrl.HandleGetAllCustomFieldNames(), userMW...))
		r.Get("/items/fields/values", chain.ToHandlerFunc(v1Ctrl.HandleGetAllCustomFieldValues(), userMW...))

		r.Get("/items/{id}", chain.ToHandlerFunc(v1Ctrl.HandleItemGet(), userMW...))
		r.Get("/items/{id}/path", chain.ToHandlerFunc(v1Ctrl.HandleItemFullPath(), userMW...))
		r.Put("/items/{id}", chain.ToHandlerFunc(v1Ctrl.HandleItemUpdate(), userMW...))
		r.Patch("/items/{id}", chain.ToHandlerFunc(v1Ctrl.HandleItemPatch(), userMW...))
		r.Delete("/items/{id}", chain.ToHandlerFunc(v1Ctrl.HandleItemDelete(), userMW...))
		r.Post("/items/{id}/duplicate", chain.ToHandlerFunc(v1Ctrl.HandleItemDuplicate(), userMW...))

		r.Post("/items/{id}/attachments", chain.ToHandlerFunc(v1Ctrl.HandleItemAttachmentCreate(), userMW...))
		r.Put("/items/{id}/attachments/{attachment_id}", chain.ToHandlerFunc(v1Ctrl.HandleItemAttachmentUpdate(), userMW...))
		r.Delete("/items/{id}/attachments/{attachment_id}", chain.ToHandlerFunc(v1Ctrl.HandleItemAttachmentDelete(), userMW...))

		r.Get("/items/{id}/maintenance", chain.ToHandlerFunc(v1Ctrl.HandleMaintenanceLogGet(), userMW...))
		r.Post("/items/{id}/maintenance", chain.ToHandlerFunc(v1Ctrl.HandleMaintenanceEntryCreate(), userMW...))

		r.Get("/assets/{id}", chain.ToHandlerFunc(v1Ctrl.HandleAssetGet(), userMW...))

		// Maintenance
		r.Get("/maintenance", chain.ToHandlerFunc(v1Ctrl.HandleMaintenanceGetAll(), userMW...))
		r.Put("/maintenance/{id}", chain.ToHandlerFunc(v1Ctrl.HandleMaintenanceEntryUpdate(), userMW...))
		r.Delete("/maintenance/{id}", chain.ToHandlerFunc(v1Ctrl.HandleMaintenanceEntryDelete(), userMW...))

		// Notifiers
		r.Get("/notifiers", chain.ToHandlerFunc(v1Ctrl.HandleGetUserNotifiers(), userMW...))
		r.Post("/notifiers", chain.ToHandlerFunc(v1Ctrl.HandleCreateNotifier(), userMW...))
		r.Put("/notifiers/{id}", chain.ToHandlerFunc(v1Ctrl.HandleUpdateNotifier(), userMW...))
		r.Delete("/notifiers/{id}", chain.ToHandlerFunc(v1Ctrl.HandleDeleteNotifier(), userMW...))
		r.Post("/notifiers/test", chain.ToHandlerFunc(v1Ctrl.HandlerNotifierTest(), userMW...))

		// Asset-Like endpoints
		assetMW := []errchain.Middleware{
			a.mwAuthToken,
			a.mwRoles(RoleModeOr, authroles.RoleUser.String(), authroles.RoleAttachments.String()),
		}

		r.Get("/products/search-from-barcode", chain.ToHandlerFunc(v1Ctrl.HandleProductSearchFromBarcode(a.conf.Barcode), userMW...))

		r.Get("/qrcode", chain.ToHandlerFunc(v1Ctrl.HandleGenerateQRCode(), assetMW...))
		r.Get(
			"/items/{id}/attachments/{attachment_id}",
			chain.ToHandlerFunc(v1Ctrl.HandleItemAttachmentGet(), assetMW...),
		)

		// Labelmaker
		r.Get("/labelmaker/location/{id}", chain.ToHandlerFunc(v1Ctrl.HandleGetLocationLabel(), userMW...))
		r.Get("/labelmaker/item/{id}", chain.ToHandlerFunc(v1Ctrl.HandleGetItemLabel(), userMW...))
		r.Get("/labelmaker/asset/{id}", chain.ToHandlerFunc(v1Ctrl.HandleGetAssetLabel(), userMW...))

		// Reporting Services
		r.Get("/reporting/bill-of-materials", chain.ToHandlerFunc(v1Ctrl.HandleBillOfMaterialsExport(), userMW...))

		r.NotFound(http.NotFound)
	})

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
