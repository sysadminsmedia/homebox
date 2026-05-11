package v1

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"go.opentelemetry.io/otel/attribute"
)

// validIntegrationName restricts integration names to safe lower-case identifiers,
// preventing settings-key injection (e.g. "../../evil").
var validIntegrationName = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)

// HandleIntegrationProxy godoc
//
//	@Summary	Integration Reverse Proxy
//	@Description	Proxies a single GET request to the configured external integration.
//				The integration's credentials (base URL + API token) are read from
//				user settings ({name}_url / {name}_token) and never exposed to the
//				frontend.  This single generic endpoint replaces all per-integration
//				proxy handlers: adding a new integration only requires a Vue component
//				and a settings entry — no new Go code.
//	@Tags		Integrations
//	@Produce	*/*
//	@Param		name	path	string	true	"Integration name, e.g. paperless"
//	@Param		path	query	string	true	"Relative API path on the upstream service, must start with /"
//	@Success	200
//	@Failure	400	{object}	validate.ErrorResponse
//	@Failure	502	{object}	validate.ErrorResponse
//	@Router		/v1/integrations/{name}/proxy [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleIntegrationProxy() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleIntegrationProxy")
		defer span.End()

		name := chi.URLParam(r, "name")
		if !validIntegrationName.MatchString(name) {
			return validate.NewRequestError(fmt.Errorf("invalid integration name"), http.StatusBadRequest)
		}

		rawPath := r.URL.Query().Get("path")
		if rawPath == "" {
			return validate.NewRequestError(fmt.Errorf("path query parameter is required"), http.StatusBadRequest)
		}
		if !strings.HasPrefix(rawPath, "/") || strings.Contains(rawPath, "://") {
			return validate.NewRequestError(fmt.Errorf("path must be a relative path starting with /"), http.StatusBadRequest)
		}

		// Normalise to prevent directory traversal while preserving trailing slash
		// (many REST APIs treat /foo/1/ and /foo/1 differently).
		cleanPath := path.Clean(rawPath)
		if !strings.HasPrefix(cleanPath, "/") {
			return validate.NewRequestError(fmt.Errorf("invalid path after normalisation"), http.StatusBadRequest)
		}
		if strings.HasSuffix(rawPath, "/") && !strings.HasSuffix(cleanPath, "/") {
			cleanPath += "/"
		}

		span.SetAttributes(
			attribute.String("integration.name", name),
			attribute.String("integration.path", cleanPath),
		)

		ctx := services.NewContext(spanCtx)
		settings, svcErr := ctrl.svc.User.GetSettings(ctx.Context, services.UseUserCtx(ctx.Context).ID)
		if svcErr != nil {
			return validate.NewRequestError(svcErr, http.StatusInternalServerError)
		}

		baseURL, _ := settings[name+"_url"].(string)
		if baseURL == "" {
			return validate.NewRequestError(
				fmt.Errorf("%s_url not configured – add it in Settings", name),
				http.StatusBadRequest,
			)
		}

		token, _ := settings[name+"_token"].(string)
		if token == "" {
			return validate.NewRequestError(
				fmt.Errorf("%s_token not configured – add it in Settings", name),
				http.StatusBadRequest,
			)
		}

		upstream := strings.TrimRight(baseURL, "/") + cleanPath

		req, err := http.NewRequest(http.MethodGet, upstream, nil)
		if err != nil {
			return validate.NewRequestError(err, http.StatusBadRequest)
		}
		req.Header.Set("Authorization", "Token "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Err(err).Str("integration", name).Str("upstream", upstream).Msg("integration proxy: upstream request failed")
			return validate.NewRequestError(err, http.StatusBadGateway)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode == http.StatusNotFound {
			return validate.NewRequestError(fmt.Errorf("resource not found at upstream"), http.StatusNotFound)
		}
		if resp.StatusCode >= 400 {
			return validate.NewRequestError(
				fmt.Errorf("upstream returned %d", resp.StatusCode),
				http.StatusBadGateway,
			)
		}

		if ct := resp.Header.Get("Content-Type"); ct != "" {
			w.Header().Set("Content-Type", ct)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(w, resp.Body)
		return nil
	}
}
