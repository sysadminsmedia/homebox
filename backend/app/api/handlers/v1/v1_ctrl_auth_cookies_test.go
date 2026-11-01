package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var allAuthCookies = []string{cookieNameToken, cookieNameRemember, cookieNameSession, cookieNameAttachmentToken}

func isExpired(c *http.Cookie) bool {
	return c.MaxAge < 0 || (!c.Expires.IsZero() && c.Expires.Before(time.Now()))
}

// The auth cookies must be host-only. With Domain=localhost, WebKit on Linux drops
// them and the login form leaves the browser signed out.
func TestSetCookies_HostOnlyAndExpiresLegacyDomainCookies(t *testing.T) {
	ctrl := &V1Controller{}
	rec := httptest.NewRecorder()

	ctrl.setCookies(rec, "localhost", "session-token", time.Now().Add(time.Hour), true, "attachment-token")

	set := map[string]string{}
	var expiredLegacy []string
	for _, c := range rec.Result().Cookies() {
		if isExpired(c) {
			assert.Equal(t, "localhost", c.Domain, "%s: only the old Domain cookie is expired", c.Name)
			assert.NotContains(t, set, c.Name, "%s: the old cookie must be expired before the new one is set", c.Name)
			expiredLegacy = append(expiredLegacy, c.Name)
			continue
		}
		assert.Empty(t, c.Domain, "%s must not carry a Domain attribute", c.Name)
		set[c.Name] = c.Value
	}

	assert.Equal(t, map[string]string{
		cookieNameToken:           "session-token",
		cookieNameRemember:        "true",
		cookieNameSession:         "true",
		cookieNameAttachmentToken: "attachment-token",
	}, set)
	assert.ElementsMatch(t, allAuthCookies, expiredLegacy)
}

func TestUnsetCookies_ExpiresHostOnlyAndLegacyDomainCookies(t *testing.T) {
	ctrl := &V1Controller{}
	rec := httptest.NewRecorder()

	ctrl.unsetCookies(rec, "localhost")

	var hostOnly, legacy []string
	for _, c := range rec.Result().Cookies() {
		assert.True(t, isExpired(c), "%s (Domain=%q) must be expired", c.Name, c.Domain)
		if c.Domain == "" {
			hostOnly = append(hostOnly, c.Name)
		} else {
			assert.Equal(t, "localhost", c.Domain)
			legacy = append(legacy, c.Name)
		}
	}

	assert.ElementsMatch(t, allAuthCookies, hostOnly)
	assert.ElementsMatch(t, allAuthCookies, legacy)
}
