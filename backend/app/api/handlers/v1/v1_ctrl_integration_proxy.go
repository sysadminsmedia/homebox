package v1

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

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

// blockedCIDRs lists address ranges the proxy must never reach.
// Prevents SSRF attacks against cloud metadata services (e.g. AWS IMDS at
// 169.254.169.254), loopback services, and internal infrastructure.
// Public hostnames are unrestricted; only private/reserved IPs are rejected.
var blockedCIDRs = func() []*net.IPNet {
	blocks := []string{
		"127.0.0.0/8",    // IPv4 loopback
		"::1/128",        // IPv6 loopback
		"169.254.0.0/16", // IPv4 link-local (AWS/GCP/Azure IMDS)
		"fe80::/10",      // IPv6 link-local
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"0.0.0.0/8",      // Unspecified
		"::/128",         // IPv6 unspecified
		"100.64.0.0/10",  // Shared address space (RFC6598)
		"fc00::/7",       // IPv6 unique-local (ULA)
	}
	nets := make([]*net.IPNet, 0, len(blocks))
	for _, cidr := range blocks {
		_, ipNet, _ := net.ParseCIDR(cidr)
		nets = append(nets, ipNet)
	}
	return nets
}()

// checkBlockedIP returns an error if ip falls within any of the blocked ranges
// (loopback, link-local, RFC1918, cloud metadata, etc.).
func checkBlockedIP(ip net.IP) error {
	for _, block := range blockedCIDRs {
		if block.Contains(ip) {
			return fmt.Errorf("integration proxy: address %s is in a blocked range", ip)
		}
	}
	return nil
}

// ssrfSafeDialContext is a DialContext for proxyHTTPClient that rejects
// connections to private, loopback, link-local and other reserved ranges.
// Both literal-IP hosts and DNS-resolved hostnames are validated before dialing.
func ssrfSafeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("integration proxy: invalid address %q: %w", addr, err)
	}
	// Fast path: literal IP — validate directly, no DNS lookup or rebinding window.
	if ip := net.ParseIP(host); ip != nil {
		if err := checkBlockedIP(ip); err != nil {
			return nil, err
		}
		d := &net.Dialer{}
		return d.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
	}
	// Hostname: resolve all addresses and validate each before dialing.
	ips, lookupErr := net.DefaultResolver.LookupIPAddr(ctx, host)
	if lookupErr != nil {
		return nil, fmt.Errorf("integration proxy: DNS lookup failed: %w", lookupErr)
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("integration proxy: no addresses resolved for %q", host)
	}
	var lastErr error
	for _, ia := range ips {
		if err := checkBlockedIP(ia.IP); err != nil {
			lastErr = err
			continue
		}
		d := &net.Dialer{}
		conn, dialErr := d.DialContext(ctx, network, net.JoinHostPort(ia.IP.String(), port))
		if dialErr == nil {
			return conn, nil
		}
		lastErr = dialErr
	}
	return nil, lastErr
}

// proxyHTTPClient is a shared client with a hard timeout and bounded pool.
// Using a dedicated client (not http.DefaultClient) prevents upstream services
// from hanging the server indefinitely.
var proxyHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 90 * time.Second,
		DialContext:     ssrfSafeDialContext,
	},
}

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
		if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
			return validate.NewRequestError(
				fmt.Errorf("%s_url must use http:// or https:// scheme", name),
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

		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, upstream, nil)
		if err != nil {
			return validate.NewRequestError(err, http.StatusBadRequest)
		}
		req.Header.Set("Authorization", "Token "+token)

		resp, err := proxyHTTPClient.Do(req)
		if err != nil {
			// Log only host+path to avoid leaking query strings or embedded credentials.
			var safeURL string
			if u, parseErr := url.Parse(upstream); parseErr == nil {
				safeURL = u.Host + u.Path
			} else {
				safeURL = "(unparseable)"
			}
			log.Err(err).Str("integration", name).Str("upstream", safeURL).Msg("integration proxy: upstream request failed")
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

		const maxResponseSize int64 = 10 * 1024 * 1024 // 10 MB

		// Reject known-oversized responses before writing any bytes to the client.
		if resp.ContentLength > maxResponseSize {
			return validate.NewRequestError(
				fmt.Errorf("upstream response too large (%d bytes)", resp.ContentLength),
				http.StatusBadGateway,
			)
		}

		// Buffer up to maxResponseSize+1 bytes so we can detect true truncation
		// and return a clean 502 rather than a partial 200 with invalid JSON.
		buf, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize+1))
		if readErr != nil {
			log.Err(readErr).Str("integration", name).Msg("integration proxy: failed to read response")
			return validate.NewRequestError(fmt.Errorf("failed to read upstream response"), http.StatusBadGateway)
		}
		if int64(len(buf)) > maxResponseSize {
			log.Warn().Str("integration", name).Msg("integration proxy: upstream response exceeded 10 MB limit")
			return validate.NewRequestError(
				fmt.Errorf("upstream response exceeds 10 MB limit"),
				http.StatusBadGateway,
			)
		}

		if ct := resp.Header.Get("Content-Type"); ct != "" {
			w.Header().Set("Content-Type", ct)
		}
		w.WriteHeader(http.StatusOK)
		if _, writeErr := w.Write(buf); writeErr != nil {
			log.Err(writeErr).Str("integration", name).Msg("integration proxy: failed to write response")
		}
		return nil
	}
}
