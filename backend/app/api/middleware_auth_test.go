package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	_ "github.com/sysadminsmedia/homebox/backend/pkgs/cgofreesqlite"
	"github.com/sysadminsmedia/homebox/backend/pkgs/hasher"
)

func newAuthTestApp(t *testing.T) (*app, services.UserAuthTokenDetail) {
	t.Helper()
	ctx := context.Background()

	client, err := ent.Open("sqlite3", "file:"+uuid.NewString()+"?mode=memory&cache=shared&_fk=1&_time_format=sqlite")
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.Schema.Create(ctx))

	repos := repo.New(client, eventbus.New(), config.Storage{
		PrefixPath: "/",
		ConnString: "file://" + os.TempDir(),
	}, "mem://{{ .Topic }}", config.Thumbnail{}, nil)

	group, err := repos.Groups.GroupCreate(ctx, "auth-test", uuid.Nil)
	require.NoError(t, err)

	const email, password = "auth-test@example.com", "correct horse battery staple"
	hash, err := hasher.HashPassword(password)
	require.NoError(t, err)
	_, err = repos.Users.Create(ctx, repo.UserCreate{
		Name:           "Auth Test",
		Email:          email,
		Password:       &hash,
		DefaultGroupID: group.ID,
	})
	require.NoError(t, err)

	svc := services.New(repos)
	tokens, err := svc.User.Login(ctx, email, password, false)
	require.NoError(t, err)
	require.NotEmpty(t, tokens.Raw)
	require.NotEmpty(t, tokens.AttachmentToken)

	return &app{conf: &config.Config{}, repos: repos, services: svc}, tokens
}

func TestAuthTokenQueryStringOnlyAcceptsAttachmentTokens(t *testing.T) {
	a, tokens := newAuthTestApp(t)

	h := a.mwAuthToken(errchain.HandlerFunc(func(http.ResponseWriter, *http.Request) error {
		return nil
	}))

	serve := func(setup func(*http.Request)) error {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/entities/x/attachments/y", nil)
		setup(req)
		return h.ServeHTTP(httptest.NewRecorder(), req)
	}
	withQuery := func(token string) func(*http.Request) {
		return func(r *http.Request) { r.URL.RawQuery = "access_token=" + url.QueryEscape(token) }
	}

	t.Run("SessionTokenInQueryRejected", func(t *testing.T) {
		err := serve(withQuery(tokens.Raw))
		var reqErr *validate.RequestError
		require.ErrorAs(t, err, &reqErr)
		assert.Equal(t, http.StatusUnauthorized, reqErr.Status)
	})

	t.Run("AttachmentTokenInQueryAccepted", func(t *testing.T) {
		assert.NoError(t, serve(withQuery(tokens.AttachmentToken)))
	})

	t.Run("SessionTokenInQueryAcceptedOnWebSocketRoute", func(t *testing.T) {
		ws := a.mwAllowQuerySessionToken(h)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/ws/events?access_token="+url.QueryEscape(tokens.Raw), nil)
		assert.NoError(t, ws.ServeHTTP(httptest.NewRecorder(), req))
	})

	t.Run("SessionTokenInHeaderAccepted", func(t *testing.T) {
		assert.NoError(t, serve(func(r *http.Request) {
			r.Header.Set("Authorization", "Bearer "+tokens.Raw)
		}))
	})
}
