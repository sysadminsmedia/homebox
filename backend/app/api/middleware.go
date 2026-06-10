package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	v1 "github.com/sysadminsmedia/homebox/backend/app/api/handlers/v1"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/authz"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"github.com/sysadminsmedia/homebox/backend/pkgs/hasher"
	"github.com/sysadminsmedia/homebox/backend/pkgs/set"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func mwTracer() trace.Tracer {
	return otel.Tracer("middleware")
}

func recordMwSpanError(span trace.Span, err error) {
	if err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

type tokenHasKey struct {
	key string
}

var hashedToken = tokenHasKey{key: "hashedToken"}

type RoleMode int

const (
	RoleModeOr  RoleMode = 0
	RoleModeAnd RoleMode = 1
)

// mwRoles is a middleware that will validate the required roles are met. All roles
// are required to be met for the request to be allowed. If the user does not have
// the required roles, a 403 Forbidden will be returned.
//
// WARNING: This middleware _MUST_ be called after mwAuthToken or else it will panic
func (a *app) mwRoles(rm RoleMode, required ...string) errchain.Middleware {
	return func(next errchain.Handler) errchain.Handler {
		return errchain.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			spanCtx, span := mwTracer().Start(r.Context(), "middleware.mwRoles",
				trace.WithAttributes(
					attribute.Int("roles.required.count", len(required)),
					attribute.StringSlice("roles.required", required),
					attribute.Int("roles.mode", int(rm)),
				))
			defer span.End()

			maybeToken := spanCtx.Value(hashedToken)
			if maybeToken == nil {
				panic("mwRoles: token not found in context, you must call mwAuthToken before mwRoles")
			}

			token := maybeToken.(string)

			// API keys grant the same access as the owning user. They live
			// outside the auth_roles table, so we synthesize the "user" role
			// here rather than querying the DB.
			var roles *set.Set[string]
			if services.IsAPIKeyAuth(spanCtx) {
				s := set.New("user")
				roles = &s
				span.SetAttributes(attribute.Bool("roles.api_key", true))
			} else {
				r, err := a.repos.AuthTokens.GetRoles(spanCtx, token)
				if err != nil {
					recordMwSpanError(span, err)
					return err
				}
				roles = r
			}
			span.SetAttributes(attribute.Int("roles.actual.count", roles.Len()))

		outer:
			switch rm {
			case RoleModeOr:
				for _, role := range required {
					if roles.Contains(role) {
						span.SetAttributes(attribute.String("roles.outcome", "ok_or"))
						break outer
					}
				}
				span.SetAttributes(attribute.String("roles.outcome", "forbidden_or"))
				return validate.NewRequestError(errors.New("Forbidden"), http.StatusForbidden)
			case RoleModeAnd:
				for _, req := range required {
					if !roles.Contains(req) {
						span.SetAttributes(
							attribute.String("roles.outcome", "forbidden_and"),
							attribute.String("roles.missing", req),
						)
						return validate.NewRequestError(errors.New("Unauthorized"), http.StatusForbidden)
					}
				}
				span.SetAttributes(attribute.String("roles.outcome", "ok_and"))
			}

			return next.ServeHTTP(w, r.WithContext(spanCtx))
		})
	}
}

type KeyFunc func(r *http.Request) (string, error)

func getBearer(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", errors.New("authorization header is required")
	}

	return auth, nil
}

func getQuery(r *http.Request) (string, error) {
	token := r.URL.Query().Get("access_token")
	if token == "" {
		return "", errors.New("access_token query is required")
	}

	token, err := url.QueryUnescape(token)
	if err != nil {
		return "", errors.New("access_token query is required")
	}

	return token, nil
}

func getWebSocketProtocolToken(r *http.Request) (string, error) {
	raw := r.Header.Get("Sec-WebSocket-Protocol")
	if raw == "" {
		return "", errors.New("websocket protocol token is required")
	}

	parts := strings.Split(raw, ",")
	if len(parts) < 2 {
		return "", errors.New("websocket protocol token is required")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("websocket protocol token is required")
	}

	return token, nil
}

// mwAuthToken is a middleware that will check the database for a stateful token
// and attach it's user to the request context, or return an appropriate error.
// Authorization support is by token via Headers or Query Parameter
//
// Example:
//   - header = "Bearer 1234567890"
//   - query = "?access_token=1234567890"
func (a *app) mwAuthToken(next errchain.Handler) errchain.Handler {
	return errchain.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := mwTracer().Start(r.Context(), "middleware.mwAuthToken",
			trace.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.path", r.URL.Path),
			))
		defer span.End()

		var requestToken string
		tokenSource := "none"

		{
			cookies, _ := v1.GetCookies(r)
			if cookies != nil {
				requestToken = cookies.Token
				tokenSource = "cookie"
			}
		}

		if requestToken == "" {
			keyFuncs := [...]struct {
				name string
				fn   KeyFunc
			}{
				{"bearer", getBearer},
				{"ws_protocol", getWebSocketProtocolToken},
				{"query", getQuery},
			}

			for _, kf := range keyFuncs {
				token, err := kf.fn(r)
				if err == nil {
					requestToken = token
					tokenSource = kf.name
					break
				}
			}
		}

		span.SetAttributes(
			attribute.String("auth.token.source", tokenSource),
			attribute.Bool("auth.token.present", requestToken != ""),
		)

		if requestToken == "" {
			span.SetAttributes(attribute.String("auth.outcome", "no_token"))
			return validate.NewRequestError(errors.New("authorization header or query is required"), http.StatusUnauthorized)
		}

		hadBearerPrefix := strings.HasPrefix(requestToken, "Bearer ")
		requestToken = strings.TrimPrefix(requestToken, "Bearer ")
		span.SetAttributes(
			attribute.Bool("auth.token.had_bearer_prefix", hadBearerPrefix),
			attribute.Int("auth.token.length", len(requestToken)),
		)

		r = r.WithContext(context.WithValue(spanCtx, hashedToken, requestToken))

		usr, err := a.services.User.GetSelf(r.Context(), requestToken)
		if err != nil && !ent.IsNotFound(err) {
			recordMwSpanError(span, err)
			span.SetAttributes(attribute.String("auth.outcome", "lookup_error"))
			return err
		}

		isAPIKey := false
		if err != nil {
			// Session-token lookup missed. API keys are only accepted via the
			// Authorization header — never via cookies or query params, since
			// those paths leak credentials into logs, browser history, and
			// referer headers. Reject without consulting the api_keys table.
			if tokenSource != "bearer" {
				span.SetAttributes(attribute.String("auth.outcome", "token_not_found"))
				return validate.NewRequestError(errors.New("valid authorization token is required"), http.StatusUnauthorized)
			}

			tokenHash := hasher.HashAPIKey(requestToken)
			keyUsr, keyID, keyErr := a.repos.APIKeys.GetUserFromToken(r.Context(), tokenHash)
			if keyErr != nil {
				if ent.IsNotFound(keyErr) {
					span.SetAttributes(attribute.String("auth.outcome", "token_not_found"))
					return validate.NewRequestError(errors.New("valid authorization token is required"), http.StatusUnauthorized)
				}
				recordMwSpanError(span, keyErr)
				span.SetAttributes(attribute.String("auth.outcome", "lookup_error"))
				return keyErr
			}
			usr = keyUsr
			isAPIKey = true

			// Best-effort last_used_at update; failure must not break the
			// request, but we want it surfaced in logs.
			if touchErr := a.repos.APIKeys.TouchLastUsed(r.Context(), keyID, time.Now()); touchErr != nil {
				log.Warn().Err(touchErr).Str("api_key.id", keyID.String()).Msg("failed to update api key last_used_at")
			}
		}

		span.SetAttributes(
			attribute.String("auth.outcome", "authenticated"),
			attribute.String("auth.method", map[bool]string{true: "api_key", false: "session"}[isAPIKey]),
			attribute.String("user.id", usr.ID.String()),
			attribute.String("user.default_group_id", usr.DefaultGroupID.String()),
			attribute.Int("user.groups.count", len(usr.GroupIDs)),
			attribute.Bool("user.is_superuser", usr.IsSuperuser),
		)

		ctxOut := services.SetUserCtx(r.Context(), &usr, requestToken)
		if isAPIKey {
			ctxOut = services.SetAPIKeyAuth(ctxOut)
		}
		r = r.WithContext(ctxOut)
		return next.ServeHTTP(w, r)
	})
}

// mwTenant is a middleware that will parse the X-Tenant header and validate the user has access
// to the requested tenant. If no header is provided, the user's default group is used.
//
// WARNING: This middleware _MUST_ be called after mwAuthToken
func (a *app) mwTenant(next errchain.Handler) errchain.Handler {
	return errchain.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := mwTracer().Start(r.Context(), "middleware.mwTenant")
		defer span.End()

		user := services.UseUserCtx(spanCtx)
		if user == nil {
			err := errors.New("user context not found")
			recordMwSpanError(span, err)
			span.SetAttributes(attribute.String("tenant.outcome", "no_user_ctx"))
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		tenantID := user.DefaultGroupID
		tenantSource := "default"

		tenantHeader := r.Header.Get("X-Tenant")
		if tenantHeader == "" {
			tenantHeader = r.URL.Query().Get("tenant")
			if tenantHeader != "" {
				tenantSource = "query"
			}
		} else {
			tenantSource = "header"
		}

		if tenantHeader != "" {
			parsedTenantID, err := uuid.Parse(tenantHeader)
			if err != nil {
				recordMwSpanError(span, err)
				span.SetAttributes(
					attribute.String("tenant.outcome", "parse_failed"),
					attribute.String("tenant.source", tenantSource),
				)
				return validate.NewRequestError(errors.New("invalid X-Tenant header format"), http.StatusBadRequest)
			}

			tenantID = parsedTenantID
		}

		hasAccess := false
		for _, gid := range user.GroupIDs {
			if gid == tenantID {
				hasAccess = true
				break
			}
		}

		span.SetAttributes(
			attribute.String("user.id", user.ID.String()),
			attribute.String("tenant.id", tenantID.String()),
			attribute.String("tenant.source", tenantSource),
			attribute.Int("user.groups.count", len(user.GroupIDs)),
			attribute.Bool("tenant.has_access", hasAccess),
		)

		if !hasAccess {
			span.SetAttributes(attribute.String("tenant.outcome", "forbidden"))
			return validate.NewRequestError(errors.New("user does not have access to the requested tenant"), http.StatusForbidden)
		}

		span.SetAttributes(attribute.String("tenant.outcome", "ok"))
		r = r.WithContext(services.SetTenantCtx(spanCtx, tenantID))
		return next.ServeHTTP(w, r)
	})
}

// mwViewer resolves the caller's effective permission set for the active
// tenant (direct membership permissions ∪ permission-group permissions) and
// attaches it to the request context as an authz.Viewer. The ent privacy
// layer reads the viewer from the context for every query and mutation.
//
// WARNING: This middleware _MUST_ be called after mwAuthToken and mwTenant.
func (a *app) mwViewer(next errchain.Handler) errchain.Handler {
	return errchain.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := mwTracer().Start(r.Context(), "middleware.mwViewer")
		defer span.End()

		user := services.UseUserCtx(spanCtx)
		if user == nil {
			err := errors.New("user context not found")
			recordMwSpanError(span, err)
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		tenantID := services.UseTenantCtx(spanCtx)
		if tenantID == uuid.Nil {
			err := errors.New("tenant context not found")
			recordMwSpanError(span, err)
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		viewer, err := a.repos.Permissions.ResolveViewer(spanCtx, user.ID, tenantID, user.IsSuperuser)
		if err != nil {
			recordMwSpanError(span, err)
			if ent.IsNotFound(err) {
				// Membership disappeared between mwTenant and here.
				return validate.NewRequestError(errors.New("user does not have access to the requested tenant"), http.StatusForbidden)
			}
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		span.SetAttributes(
			attribute.String("user.id", user.ID.String()),
			attribute.String("tenant.id", tenantID.String()),
			attribute.Int("viewer.permissions.count", len(viewer.Perms)),
			attribute.Int("viewer.permission_groups.count", len(viewer.PermGroupIDs)),
			attribute.Bool("viewer.superuser", viewer.Superuser),
		)

		r = r.WithContext(authz.NewContext(spanCtx, viewer))
		return next.ServeHTTP(w, r)
	})
}

// mwPermission fast-fails requests whose viewer lacks a tenant-wide
// permission, returning a clean 403 before the handler runs. It is purely a
// UX nicety on management routes: the ent privacy layer remains the source
// of truth, and this middleware must never be the only protection. Do not
// apply it to entity CRUD routes — row-level grants apply there and the
// tenant-wide permission is not decisive.
//
// WARNING: This middleware _MUST_ be called after mwViewer.
func (a *app) mwPermission(perm authz.Permission) errchain.Middleware {
	return func(next errchain.Handler) errchain.Handler {
		return errchain.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			v := authz.FromContext(r.Context())
			if v == nil {
				return validate.NewRequestError(errors.New("viewer context not found"), http.StatusInternalServerError)
			}
			if !v.Has(perm) {
				return validate.NewRequestError(errors.New("missing required permission: "+string(perm)), http.StatusForbidden)
			}
			return next.ServeHTTP(w, r)
		})
	}
}

// authRateLimiter tracks authentication attempts per client and applies a backoff when limits are exceeded.
type authRateLimiter struct {
	cfg         config.AuthRateLimit
	mu          sync.Mutex
	state       map[string]*authAttempt
	nowFn       func() time.Time
	stopCleanup chan struct{}
	stopOnce    sync.Once
}

// authAttempt struct represents the state of authentication attempts for a client.
type authAttempt struct {
	attempts    int
	lastAttempt time.Time
	lockedUntil time.Time
}

// newAuthRateLimiter creates a new authRateLimiter instance.
func newAuthRateLimiter(cfg config.AuthRateLimit) *authRateLimiter {
	// Sanity defaults to avoid zero values creating odd behavior.
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 5
	}
	if cfg.BaseBackoff <= 0 {
		cfg.BaseBackoff = 2 * time.Second
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = 2 * time.Minute
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}

	limiter := &authRateLimiter{
		cfg:         cfg,
		state:       make(map[string]*authAttempt),
		nowFn:       time.Now,
		stopCleanup: make(chan struct{}),
	}

	// Start background cleanup goroutine
	go limiter.cleanupLoop()

	return limiter
}

// cleanupLoop periodically removes stale entries from the state map.
func (l *authRateLimiter) cleanupLoop() {
	// Run cleanup every window period (or at least every 5 minutes)
	cleanupInterval := l.cfg.Window
	if cleanupInterval > 5*time.Minute {
		cleanupInterval = 5 * time.Minute
	}

	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.cleanup()
		case <-l.stopCleanup:
			return
		}
	}
}

// cleanup removes stale entries that are outside the window.
func (l *authRateLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.nowFn()
	for key, attempt := range l.state {
		// Remove entries that are:
		// 1. Outside the window AND
		// 2. No longer locked (or lock has expired)
		if now.Sub(attempt.lastAttempt) > l.cfg.Window && now.After(attempt.lockedUntil) {
			delete(l.state, key)
		}
	}
}

// Stop gracefully stops the cleanup goroutine.
func (l *authRateLimiter) Stop() {
	l.stopOnce.Do(func() {
		close(l.stopCleanup)
	})
}

// mwAuthRateLimit enforces request throttling for authentication endpoints with an exponential backoff.
func (a *app) mwAuthRateLimit(next errchain.Handler) errchain.Handler {
	return errchain.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		spanCtx, span := mwTracer().Start(r.Context(), "middleware.mwAuthRateLimit")
		defer span.End()

		limiter := a.authLimiter
		if limiter == nil || !limiter.cfg.Enabled {
			span.SetAttributes(attribute.String("rate_limit.outcome", "disabled"))
			return next.ServeHTTP(w, r.WithContext(spanCtx))
		}

		key := limiter.keyForRequest(r, a.conf.Options.TrustProxy)
		now := limiter.nowFn()
		span.SetAttributes(attribute.String("rate_limit.key", key))

		if retryAfter, allowed := limiter.shouldAllow(key, now); !allowed {
			seconds := int(retryAfter.Round(time.Second).Seconds())
			if seconds < 0 {
				seconds = 0
			}
			span.SetAttributes(
				attribute.String("rate_limit.outcome", "blocked"),
				attribute.Int("rate_limit.retry_after_seconds", seconds),
			)
			w.Header().Set("Retry-After", strconv.Itoa(seconds))
			return validate.NewRequestError(errors.New("too many authentication attempts"), http.StatusTooManyRequests)
		}

		err := next.ServeHTTP(w, r.WithContext(spanCtx))
		limiter.record(key, now, err == nil)
		span.SetAttributes(
			attribute.String("rate_limit.outcome", "allowed"),
			attribute.Bool("rate_limit.recorded_success", err == nil),
		)
		return err
	})
}

// shouldAllow checks if the client should be allowed to authenticate based on the configured rate limit.
func (l *authRateLimiter) shouldAllow(key string, now time.Time) (time.Duration, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	attempt, ok := l.state[key]
	if !ok {
		return 0, true
	}

	if now.Sub(attempt.lastAttempt) > l.cfg.Window {
		delete(l.state, key)
		return 0, true
	}

	if now.Before(attempt.lockedUntil) {
		return time.Until(attempt.lockedUntil), false
	}

	return 0, true
}

// record updates the authentication attempt state for the given client.
func (l *authRateLimiter) record(key string, now time.Time, success bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if success {
		delete(l.state, key)
		return
	}

	attempt, ok := l.state[key]
	if !ok {
		l.state[key] = &authAttempt{attempts: 1, lastAttempt: now}
		return
	}

	if now.Sub(attempt.lastAttempt) > l.cfg.Window {
		attempt.attempts = 0
		attempt.lockedUntil = time.Time{}
	}

	attempt.attempts++
	attempt.lastAttempt = now

	if attempt.attempts > l.cfg.MaxAttempts {
		over := attempt.attempts - l.cfg.MaxAttempts
		delay := l.cfg.BaseBackoff
		for i := 1; i < over; i++ {
			delay *= 2
			if delay >= l.cfg.MaxBackoff {
				delay = l.cfg.MaxBackoff
				break
			}
		}
		if delay > l.cfg.MaxBackoff {
			delay = l.cfg.MaxBackoff
		}
		attempt.lockedUntil = now.Add(delay)
	}
}

// extractClientIP extracts the client IP from the request.
// It only uses proxy headers (X-Real-IP, X-Forwarded-For) if trustProxy is enabled.
func extractClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if ip := r.Header.Get("X-Real-IP"); ip != "" {
			return ip
		}

		if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
			parts := strings.Split(ip, ",")
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}

// keyForRequest returns a unique key for the given request.
func (l *authRateLimiter) keyForRequest(r *http.Request, trustProxy bool) string {
	return extractClientIP(r, trustProxy) + "|" + r.URL.Path
}

// simpleRateLimiter provides token bucket rate limiting per client IP.
type simpleRateLimiter struct {
	mu          sync.Mutex
	limiters    map[string]*rateLimiterEntry
	rate        int           // requests allowed
	window      time.Duration // time window
	trustProxy  bool          // whether to trust proxy headers
	stopCleanup chan struct{}
	stopOnce    sync.Once
}

type rateLimiterEntry struct {
	tokens     int
	lastRefill time.Time
}

// newSimpleRateLimiter creates a new rate limiter with the specified rate and window.
func newSimpleRateLimiter(rate int, window time.Duration, trustProxy bool) *simpleRateLimiter {
	rl := &simpleRateLimiter{
		limiters:    make(map[string]*rateLimiterEntry),
		rate:        rate,
		window:      window,
		trustProxy:  trustProxy,
		stopCleanup: make(chan struct{}),
	}

	// Start background cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// cleanupLoop periodically removes stale entries from the limiters map.
func (rl *simpleRateLimiter) cleanupLoop() {
	// Run cleanup every 2x the window period (or at least every 5 minutes)
	cleanupInterval := rl.window * 2
	if cleanupInterval < 5*time.Minute {
		cleanupInterval = 5 * time.Minute
	}

	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopCleanup:
			return
		}
	}
}

// cleanup removes stale entries that haven't been accessed recently.
func (rl *simpleRateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	// Remove entries that haven't been accessed in 2x the window period
	staleThreshold := rl.window * 2

	for key, entry := range rl.limiters {
		if now.Sub(entry.lastRefill) > staleThreshold {
			delete(rl.limiters, key)
		}
	}
}

// Stop gracefully stops the cleanup goroutine.
func (rl *simpleRateLimiter) Stop() {
	rl.stopOnce.Do(func() {
		close(rl.stopCleanup)
	})
}

// allow checks if the request should be allowed based on the rate limit.
func (rl *simpleRateLimiter) allow(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.limiters[clientIP]

	if !exists {
		// First request from this IP
		rl.limiters[clientIP] = &rateLimiterEntry{
			tokens:     rl.rate - 1,
			lastRefill: now,
		}
		return true
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(entry.lastRefill)
	if elapsed >= rl.window {
		// Full refill
		entry.tokens = rl.rate - 1
		entry.lastRefill = now
		return true
	}

	// Check if tokens are available
	if entry.tokens > 0 {
		entry.tokens--
		return true
	}

	return false
}

// getClientIP extracts the client IP from the request.
// It only uses proxy headers (X-Real-IP, X-Forwarded-For) if trustProxy is enabled.
func (rl *simpleRateLimiter) getClientIP(r *http.Request, trustProxy bool) string {
	return extractClientIP(r, trustProxy)
}

// middleware wraps the rate limiter as an errchain middleware.
func (rl *simpleRateLimiter) middleware(next errchain.Handler) errchain.Handler {
	return errchain.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		clientIP := rl.getClientIP(r, rl.trustProxy)

		if !rl.allow(clientIP) {
			w.Header().Set("Retry-After", strconv.Itoa(int(rl.window.Seconds())))
			return validate.NewRequestError(errors.New("rate limit exceeded"), http.StatusTooManyRequests)
		}

		return next.ServeHTTP(w, r)
	})
}
