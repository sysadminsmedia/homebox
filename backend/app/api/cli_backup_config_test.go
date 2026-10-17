package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sysadminsmedia/homebox/backend/internal/sys/backupsetup"
)

type cliRun struct {
	code        int
	out, errOut string
}

func runCLI(t *testing.T, stdin string, tty bool, secrets []string, args ...string) cliRun {
	t.Helper()
	var out, errOut bytes.Buffer
	next := 0
	code := runBackupConfig(args, backupConfigIO{
		in: bufio.NewReader(strings.NewReader(stdin)), out: &out, errOut: &errOut, tty: tty,
		readPass: func() (string, error) {
			require.Less(t, next, len(secrets), "unexpected secret prompt")
			v := secrets[next]
			next++
			return v, nil
		},
	})
	return cliRun{code, out.String(), errOut.String()}
}

const flagBaseURL = "--base-url"

const googleSecret = "GOCSPX-supersecretvalue_123"

func googleFlags(envFile string) []string {
	return []string{
		"--non-interactive", "--output", envFile,
		flagBaseURL, "https://homebox.example.com", "--behind-proxy",
		"--local-root", "/backups",
		"--google-client-id", "1234-abc.apps.googleusercontent.com", "--google-client-secret", googleSecret,
	}
}

func TestBackupConfigWritesAnEnvFile(t *testing.T) {
	env := filepath.Join(t.TempDir(), ".env")
	r := runCLI(t, "", false, nil, googleFlags(env)...)
	require.Equal(t, 0, r.code, r.errOut)

	b, err := os.ReadFile(env)
	require.NoError(t, err)
	vars := backupsetup.ParseEnv(b)
	assert.Equal(t, "/backups", vars[backupsetup.EnvLocalRoot])
	assert.Equal(t, "1234-abc.apps.googleusercontent.com", vars[backupsetup.EnvGoogleID])
	assert.Equal(t, googleSecret, vars[backupsetup.EnvGoogleSecret])
	assert.Equal(t, "homebox.example.com", vars[backupsetup.EnvHostname])
	assert.Equal(t, "true", vars[backupsetup.EnvTrustProxy])
	assert.GreaterOrEqual(t, len(vars[backupsetup.EnvEncryptionKey]), 32)

	st, err := os.Stat(env)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), st.Mode().Perm(), "secrets are owner-only")

	// Secrets never reach the terminal; the guide and compose snippet do.
	assert.NotContains(t, r.errOut, googleSecret)
	assert.NotContains(t, r.errOut, vars[backupsetup.EnvEncryptionKey])
	assert.Contains(t, r.errOut, "https://homebox.example.com/api/v1/group/backup-oauth/callback")
	assert.Contains(t, r.errOut, "env_file:")
	assert.Contains(t, r.errOut, ":/backups")
	assert.Contains(t, r.errOut, "In production")
	assert.Contains(t, r.errOut, "generated a new encryption key")
}

func TestBackupConfigRefusesInvalidInputAndWritesNothing(t *testing.T) {
	cases := map[string][]string{
		"secret without id":         {flagBaseURL, "https://h.example.com", "--google-client-secret", "abc"},
		"http public host":          {flagBaseURL, "http://h.example.com", "--dropbox-client-id", "a", "--dropbox-client-secret", "b"},
		"missing address":           {"--google-client-id", "a", "--google-client-secret", "b"},
		"relative local root":       {"--local-root", "backups"},
		"short key":                 {"--encryption-key", "short"},
		"https without proxy trust": {flagBaseURL, "https://h.example.com", "--dropbox-client-id", "a", "--dropbox-client-secret", "b"},
		"injection":                 {flagBaseURL, "https://h.example.com", "--dropbox-client-id", "a", "--dropbox-client-secret", "b\nHBOX_DEMO=true"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			env := filepath.Join(t.TempDir(), ".env")
			r := runCLI(t, "", false, nil, append([]string{"--non-interactive", "--output", env}, args...)...)
			assert.Equal(t, 1, r.code, r.errOut)
			assert.Contains(t, r.errOut, "error:")
			assert.Contains(t, r.errOut, "nothing was written")
			_, err := os.Stat(env)
			assert.True(t, os.IsNotExist(err), "no file on a validation failure")
		})
	}
}

func TestBackupConfigMergesIntoAnExistingFileAndKeepsTheKey(t *testing.T) {
	env := filepath.Join(t.TempDir(), ".env")
	original := "# my homebox\nHBOX_MODE=production\nHBOX_BACKUP_ENCRYPTION_KEY=keep-this-key-keep-this-key-1234\nHBOX_BACKUP_LOCAL_ROOT=/old\nOTHER=value\n"
	require.NoError(t, os.WriteFile(env, []byte(original), 0o644))

	r := runCLI(t, "", false, nil, googleFlags(env)...)
	require.Equal(t, 0, r.code, r.errOut)
	assert.Contains(t, r.errOut, "keeping the existing encryption key")

	b, err := os.ReadFile(env)
	require.NoError(t, err)
	got := string(b)
	assert.Contains(t, got, "# my homebox")
	assert.Contains(t, got, "HBOX_MODE=production")
	assert.Contains(t, got, "OTHER=value")
	assert.Contains(t, got, "HBOX_BACKUP_ENCRYPTION_KEY=keep-this-key-keep-this-key-1234")
	assert.Contains(t, got, "HBOX_BACKUP_LOCAL_ROOT=/backups")
	assert.NotContains(t, got, "/old")

	bak, err := os.ReadFile(env + ".bak")
	require.NoError(t, err)
	assert.Equal(t, original, string(bak), "the previous file is saved")
	st, err := os.Stat(env + ".bak")
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), st.Mode().Perm(), "a backup of a file with secrets is owner-only too")
	st, err = os.Stat(env)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), st.Mode().Perm(), "an existing file is tightened to 600")

	// Running it again changes nothing.
	r = runCLI(t, "", false, nil, googleFlags(env)...)
	require.Equal(t, 0, r.code)
	again, _ := os.ReadFile(env)
	assert.Equal(t, got, string(again), "idempotent")
}

func TestBackupConfigRotateKeyIsExplicitAndLoud(t *testing.T) {
	env := filepath.Join(t.TempDir(), ".env")
	require.NoError(t, os.WriteFile(env, []byte("HBOX_BACKUP_ENCRYPTION_KEY=old-key-old-key-old-key-old-key-1\n"), 0o600))

	r := runCLI(t, "", false, nil, "--non-interactive", "--output", env, "--rotate-key")
	require.Equal(t, 0, r.code, r.errOut)
	assert.Contains(t, r.errOut, "WARNING")
	assert.Contains(t, r.errOut, "must be re-entered")
	b, _ := os.ReadFile(env)
	assert.NotContains(t, string(b), "old-key-old-key")
}

func TestBackupConfigStdoutAndDryRun(t *testing.T) {
	dir := t.TempDir()
	env := filepath.Join(dir, ".env")

	r := runCLI(t, "", false, nil, append(googleFlags(env), "--stdout")...)
	require.Equal(t, 0, r.code, r.errOut)
	vars := backupsetup.ParseEnv([]byte(r.out))
	assert.Equal(t, googleSecret, vars[backupsetup.EnvGoogleSecret], "the env content goes to stdout")
	assert.NotContains(t, r.errOut, googleSecret, "and only there")
	_, err := os.Stat(env)
	assert.True(t, os.IsNotExist(err), "--stdout writes no file")

	r = runCLI(t, "", false, nil, append(googleFlags(env), "--dry-run")...)
	require.Equal(t, 0, r.code, r.errOut)
	assert.Contains(t, r.errOut, "dry run")
	assert.Contains(t, r.errOut, "+ HBOX_BACKUP_GOOGLE_CLIENT_SECRET=GOC****", "secrets are masked in the preview")
	assert.Empty(t, r.out)
	_, err = os.Stat(env)
	assert.True(t, os.IsNotExist(err), "--dry-run writes nothing")
}

func TestBackupConfigStdoutMergesAnExplicitOutputFile(t *testing.T) {
	env := filepath.Join(t.TempDir(), "prod.env")
	original := "HBOX_MODE=production\nHBOX_BACKUP_ENCRYPTION_KEY=keep-this-key-keep-this-key-1234\n"
	require.NoError(t, os.WriteFile(env, []byte(original), 0o600))

	r := runCLI(t, "", false, nil, append(googleFlags(env), "--stdout")...)
	require.Equal(t, 0, r.code, r.errOut)
	vars := backupsetup.ParseEnv([]byte(r.out))
	assert.Equal(t, "keep-this-key-keep-this-key-1234", vars[backupsetup.EnvEncryptionKey], "the existing key survives")
	assert.Equal(t, "production", vars["HBOX_MODE"])
	after, _ := os.ReadFile(env)
	assert.Equal(t, original, string(after), "the file itself is not touched")
}

func TestBackupConfigInteractive(t *testing.T) {
	env := filepath.Join(t.TempDir(), ".env")
	answers := strings.Join([]string{
		"/backups",                    // local root
		"n",                           // custom endpoints
		"https://homebox.example.com", // address
		"",                            // trust the proxy: Enter accepts the default (yes)
		"y",                           // Google Drive
		"1234-abc.apps.googleusercontent.com",
		"n", // OneDrive
		"n", // Dropbox
		"y", // confirm write
	}, "\n") + "\n"

	r := runCLI(t, answers, true, []string{googleSecret}, "--output", env)
	require.Equal(t, 0, r.code, r.errOut)
	assert.Contains(t, r.errOut, "Google Cloud console", "the provider guide is shown before asking for the credentials")
	assert.NotContains(t, r.errOut, googleSecret)

	b, err := os.ReadFile(env)
	require.NoError(t, err)
	vars := backupsetup.ParseEnv(b)
	assert.Equal(t, "/backups", vars[backupsetup.EnvLocalRoot])
	assert.Equal(t, googleSecret, vars[backupsetup.EnvGoogleSecret])
	assert.Equal(t, "homebox.example.com", vars[backupsetup.EnvHostname])
	assert.Equal(t, "true", vars[backupsetup.EnvTrustProxy])
	assert.NotContains(t, vars, backupsetup.EnvMicrosoftID)
}

func TestBackupConfigInteractiveDeclineAndEarlyEnd(t *testing.T) {
	env := filepath.Join(t.TempDir(), ".env")

	// Skipping everything still produces a valid file (a key), but declining the
	// confirmation writes nothing.
	r := runCLI(t, "\n\n\nn\n", true, nil, "--output", env)
	assert.Equal(t, 1, r.code)
	assert.Contains(t, r.errOut, "nothing was written")
	_, err := os.Stat(env)
	assert.True(t, os.IsNotExist(err))

	// Input that ends mid-wizard is an error, not a silent default.
	r = runCLI(t, "/backups\n", true, nil, "--output", env)
	assert.Equal(t, 1, r.code)
	assert.Contains(t, r.errOut, "input ended")
	_, err = os.Stat(env)
	assert.True(t, os.IsNotExist(err))
}

func TestBackupConfigPromptsDefaultFromTheExistingFile(t *testing.T) {
	env := filepath.Join(t.TempDir(), ".env")
	require.NoError(t, os.WriteFile(env, []byte("HBOX_BACKUP_LOCAL_ROOT=/mnt/nas\nHBOX_OPTIONS_HOSTNAME=homebox.example.com\n"), 0o600))

	// Enter accepts the defaults: the existing local root and address; no providers.
	r := runCLI(t, "\n\n\ny\nn\nn\nn\ny\n", true, nil, "--output", env)
	require.Equal(t, 0, r.code, r.errOut)
	assert.Contains(t, r.errOut, "[/mnt/nas]")
	assert.Contains(t, r.errOut, "[https://homebox.example.com]")
}

func TestBackupConfigHelpAndDispatch(t *testing.T) {
	r := runCLI(t, "", false, nil, "-h")
	assert.Equal(t, 0, r.code)
	assert.Contains(t, r.errOut, "Usage: homebox backup-config")
	assert.Contains(t, r.errOut, "--rotate-key")

	r = runCLI(t, "", false, nil, "--no-such-flag")
	assert.Equal(t, 2, r.code)

	handled, _ := runBackupConfigCLI([]string{"homebox"})
	assert.False(t, handled, "the server starts when no subcommand is given")
	handled, _ = runBackupConfigCLI([]string{"homebox", "reset-password"})
	assert.False(t, handled)
}

func TestBackupConfigAcceptsProxyTrustAlreadyConfigured(t *testing.T) {
	flags := func(env string) []string {
		return []string{
			"--non-interactive", "--output", env, flagBaseURL, "https://homebox.example.com",
			"--dropbox-client-id", "abc", "--dropbox-client-secret", "def",
		}
	}

	// Already in the target file.
	env := filepath.Join(t.TempDir(), ".env")
	require.NoError(t, os.WriteFile(env, []byte("HBOX_OPTIONS_TRUST_PROXY=true\n"), 0o600))
	r := runCLI(t, "", false, nil, flags(env)...)
	require.Equal(t, 0, r.code, r.errOut)

	// Already in the process environment (e.g. an existing deployment).
	t.Setenv(backupsetup.EnvTrustProxy, "true")
	r = runCLI(t, "", false, nil, flags(filepath.Join(t.TempDir(), ".env"))...)
	require.Equal(t, 0, r.code, r.errOut)
}

func TestBackupConfigCustomEndpointsAreOptIn(t *testing.T) {
	read := func(env string) map[string]string {
		b, err := os.ReadFile(env)
		require.NoError(t, err)
		return backupsetup.ParseEnv(b)
	}

	t.Run("off by default: nothing is written and the server's default applies", func(t *testing.T) {
		env := filepath.Join(t.TempDir(), ".env")
		r := runCLI(t, "", false, nil, "--non-interactive", "--output", env, "--local-root", "/backups")
		require.Equal(t, 0, r.code, r.errOut)
		assert.NotContains(t, read(env), backupsetup.EnvCustomEndpoints)
		assert.NotContains(t, r.errOut, "custom endpoints are on")
	})

	t.Run("the flag turns it on, with a warning", func(t *testing.T) {
		env := filepath.Join(t.TempDir(), ".env")
		r := runCLI(t, "", false, nil, "--non-interactive", "--output", env, "--allow-custom-endpoints")
		require.Equal(t, 0, r.code, r.errOut)
		assert.Equal(t, "true", read(env)[backupsetup.EnvCustomEndpoints])
		assert.Contains(t, r.errOut, "custom endpoints are on")
	})

	t.Run("the wizard asks, and a yes turns it on", func(t *testing.T) {
		env := filepath.Join(t.TempDir(), ".env")
		// local root (blank), custom endpoints (yes), address (blank), confirm.
		r := runCLI(t, "\ny\n\ny\n", true, nil, "--output", env)
		require.Equal(t, 0, r.code, r.errOut)
		assert.Contains(t, r.errOut, "make the Homebox server connect to an address")
		assert.Equal(t, "true", read(env)[backupsetup.EnvCustomEndpoints])
	})

	t.Run("an existing choice in the file is kept and not asked again", func(t *testing.T) {
		env := filepath.Join(t.TempDir(), ".env")
		require.NoError(t, os.WriteFile(env, []byte("HBOX_BACKUP_ALLOW_CUSTOM_ENDPOINTS=true\n"), 0o600))
		r := runCLI(t, "\n\ny\n", true, nil, "--output", env)
		require.Equal(t, 0, r.code, r.errOut)
		assert.NotContains(t, r.errOut, "Allow SFTP, WebDAV, SMB")
		assert.Equal(t, "true", read(env)[backupsetup.EnvCustomEndpoints])
	})
}
