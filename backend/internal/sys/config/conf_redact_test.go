package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sentinel = redactedValue

func Test_AuthConfig_RedactsAPIKeyPepper(t *testing.T) {
	c := AuthConfig{APIKeyPepper: "super-secret-pepper"}

	out, err := json.Marshal(c)
	require.NoError(t, err)

	assert.NotContains(t, string(out), "super-secret-pepper")
	assert.Contains(t, string(out), sentinel)
}

func Test_AuthConfig_EmptyAPIKeyPepperStaysEmpty(t *testing.T) {
	c := AuthConfig{}

	out, err := json.Marshal(c)
	require.NoError(t, err)

	assert.NotContains(t, string(out), sentinel)
}

func Test_OIDCConf_RedactsClientSecret(t *testing.T) {
	c := OIDCConf{ClientID: "public-client-id", ClientSecret: "shh"}

	out, err := json.Marshal(c)
	require.NoError(t, err)

	assert.Contains(t, string(out), "public-client-id")
	assert.NotContains(t, string(out), `"shh"`)
	assert.Contains(t, string(out), sentinel)
}

func Test_MailerConf_RedactsPassword(t *testing.T) {
	c := MailerConf{Host: "smtp.example.com", Username: "u", Password: "pw", From: "f"}

	out, err := json.Marshal(c)
	require.NoError(t, err)

	assert.NotContains(t, string(out), `"pw"`)
	assert.Contains(t, string(out), `"u"`)
	assert.Contains(t, string(out), sentinel)
}

func Test_Database_RedactsPasswordAndPubSubCreds(t *testing.T) {
	c := Database{
		Driver:           "postgres",
		Username:         "homebox",
		Password:         "dbpass",
		Host:             "db",
		Port:             "5432",
		Database:         "homebox",
		PubSubConnString: "postgres://pubuser:pubpass@db:5432/homebox?sslmode=disable",
	}

	out, err := json.Marshal(c)
	require.NoError(t, err)

	s := string(out)
	assert.NotContains(t, s, "dbpass")
	assert.NotContains(t, s, "pubpass")
	assert.Contains(t, s, "pubuser", "username portion should remain visible")
	assert.Contains(t, s, sentinel)
}

func Test_Database_LeavesUncredentialedPubSubAlone(t *testing.T) {
	c := Database{PubSubConnString: "mem://{{ .Topic }}"}

	out, err := json.Marshal(c)
	require.NoError(t, err)

	assert.Contains(t, string(out), "mem://")
}

func Test_Storage_RedactsConnStringUserinfo(t *testing.T) {
	c := Storage{ConnString: "s3://AKIA:secret@bucket.example.com/path"}

	out, err := json.Marshal(c)
	require.NoError(t, err)

	assert.NotContains(t, string(out), "secret@bucket")
	assert.Contains(t, string(out), "REDACTED")
}

func Test_BarcodeAPIConf_RedactsToken(t *testing.T) {
	c := BarcodeAPIConf{TokenBarcodespider: "token-xyz", OpenFoodFactsContact: "contact@example.com"}

	out, err := json.Marshal(c)
	require.NoError(t, err)

	assert.NotContains(t, string(out), "token-xyz")
	assert.Contains(t, string(out), "contact@example.com")
	assert.Contains(t, string(out), sentinel)
}

func Test_OTelConfig_RedactsHeaders(t *testing.T) {
	c := OTelConfig{Headers: "Authorization=Bearer hunter2,X-Other=val"}

	out, err := json.Marshal(c)
	require.NoError(t, err)

	assert.NotContains(t, string(out), "hunter2")
	assert.Contains(t, string(out), sentinel)
}

func Test_Config_FullMarshalRedactsAllSecrets(t *testing.T) {
	c := &Config{
		Auth:    AuthConfig{APIKeyPepper: "pepper-secret"},
		OIDC:    OIDCConf{ClientSecret: "oidc-secret"},
		Mailer:  MailerConf{Password: "mailer-secret"},
		Storage: Storage{ConnString: "s3://k:s3secret@b/p"},
		Database: Database{
			Password:         "db-secret",
			PubSubConnString: "postgres://u:pubsecret@h/d",
		},
		Barcode: BarcodeAPIConf{TokenBarcodespider: "bs-secret"},
		Otel:    OTelConfig{Headers: "Authorization=Bearer otel-secret"},
	}

	out, err := json.MarshalIndent(c, "", "  ")
	require.NoError(t, err)

	for _, secret := range []string{
		"pepper-secret",
		"oidc-secret",
		"mailer-secret",
		"s3secret",
		"db-secret",
		"pubsecret",
		"bs-secret",
		"otel-secret",
	} {
		assert.NotContainsf(t, string(out), secret, "expected %q to be redacted in output", secret)
	}
}
