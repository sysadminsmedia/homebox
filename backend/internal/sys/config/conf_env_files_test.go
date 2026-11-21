package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ardanlabs/conf/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeSecret(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "secret")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

// unsetOnCleanup restores a variable that loadEnvFiles sets outside of t.Setenv.
func unsetOnCleanup(t *testing.T, key string) {
	t.Helper()
	t.Cleanup(func() { _ = os.Unsetenv(key) })
}

func Test_LoadEnvFiles_ReadsFileIntoVariable(t *testing.T) {
	t.Setenv("HBOXTEST_AUTH_API_KEY_PEPPER_FILE", writeSecret(t, "pepper-from-file\n"))
	unsetOnCleanup(t, "HBOXTEST_AUTH_API_KEY_PEPPER")

	require.NoError(t, loadEnvFiles("HBOXTEST"))

	var cfg Config
	_, err := conf.Parse("HBOXTEST", &cfg)
	require.NoError(t, err)
	assert.Equal(t, "pepper-from-file", cfg.Auth.APIKeyPepper)
}

func Test_LoadEnvFiles_FileTakesPrecedence(t *testing.T) {
	t.Setenv("HBOXTEST_OIDC_CLIENT_SECRET", "from-env")
	t.Setenv("HBOXTEST_OIDC_CLIENT_SECRET_FILE", writeSecret(t, "from-file\r\n"))

	require.NoError(t, loadEnvFiles("HBOXTEST"))

	assert.Equal(t, "from-file", os.Getenv("HBOXTEST_OIDC_CLIENT_SECRET"))
}

func Test_LoadEnvFiles_MissingFile(t *testing.T) {
	t.Setenv("HBOXTEST_MAILER_PASSWORD_FILE", filepath.Join(t.TempDir(), "missing"))
	unsetOnCleanup(t, "HBOXTEST_MAILER_PASSWORD")

	err := loadEnvFiles("HBOXTEST")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HBOXTEST_MAILER_PASSWORD_FILE")
}

func Test_LoadEnvFiles_IgnoresOtherVariables(t *testing.T) {
	path := writeSecret(t, "ignored")
	t.Setenv("HBOXTEST_FILE", path)
	t.Setenv("OTHER_SECRET_FILE", path)

	require.NoError(t, loadEnvFiles("HBOXTEST"))

	_, set := os.LookupEnv("HBOXTEST")
	assert.False(t, set)
	_, set = os.LookupEnv("OTHER_SECRET")
	assert.False(t, set)
}
