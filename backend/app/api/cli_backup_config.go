package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"

	"github.com/sysadminsmedia/homebox/backend/internal/sys/backupsetup"
)

// runBackupConfigCLI handles `homebox backup-config`: a wizard that generates
// the server settings scheduled backups need (the credential encryption key,
// the local backup root and the cloud-drive OAuth apps) as a .env file, and
// prints a matching docker compose fragment and provider setup steps.
//
// Returns true when it consumed the command (and the caller should exit), so
// `homebox` with no subcommand still falls through to the server.
func runBackupConfigCLI(args []string) (handled bool, exitCode int) {
	if len(args) < 2 || args[1] != "backup-config" {
		return false, 0
	}
	bio := backupConfigIO{
		in:       bufio.NewReader(os.Stdin),
		out:      os.Stdout,
		errOut:   os.Stderr,
		tty:      term.IsTerminal(int(os.Stdin.Fd())),
		readPass: readPasswordFromTerminal,
	}
	return true, runBackupConfig(args[2:], bio)
}

func readPasswordFromTerminal() (string, error) {
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	_, _ = fmt.Fprintln(os.Stderr)
	return string(b), err
}

// backupConfigIO lets tests drive the wizard without a terminal.
type backupConfigIO struct {
	in       *bufio.Reader
	out      io.Writer // the .env content when --stdout is used
	errOut   io.Writer // prompts and messages
	tty      bool
	readPass func() (string, error)
}

type backupConfigFlags struct {
	output, baseURL, localRoot, tenant, key          string
	googleID, googleSecret, msID, msSecret           string
	dropboxID, dropboxSecret                         string
	toStdout, dryRun, yes, nonInteractive, rotateKey bool
	behindProxy, allowCustomEndpoints                bool
}

func runBackupConfig(args []string, bio backupConfigIO) int {
	var f backupConfigFlags
	fs := flag.NewFlagSet("backup-config", flag.ContinueOnError)
	fs.SetOutput(bio.errOut)
	fs.StringVar(&f.output, "output", ".env", "env file to create or update")
	fs.BoolVar(&f.toStdout, "stdout", false, "print the env file to stdout instead of writing it (prompts go to stderr); with --output, the existing file is merged and kept as-is")
	fs.BoolVar(&f.dryRun, "dry-run", false, "show what would change, write nothing")
	fs.BoolVar(&f.yes, "yes", false, "do not ask for confirmation before writing")
	fs.BoolVar(&f.nonInteractive, "non-interactive", false, "never prompt; use only the flags")
	fs.StringVar(&f.baseURL, "base-url", "", "address Homebox is reached at, e.g. https://homebox.example.com (needed for cloud drives)")
	fs.BoolVar(&f.behindProxy, "behind-proxy", false, "Homebox is behind a TLS-terminating reverse proxy that sets X-Forwarded-Host/-Proto")
	fs.StringVar(&f.localRoot, "local-root", "", "folder inside the container for local-directory backups, e.g. /backups")
	fs.BoolVar(&f.allowCustomEndpoints, "allow-custom-endpoints", false, "let collection owners use SFTP, WebDAV and SMB servers and S3-compatible endpoints (the server then connects to addresses they enter)")
	fs.StringVar(&f.key, "encryption-key", "", "use this encryption key instead of generating one")
	fs.BoolVar(&f.rotateKey, "rotate-key", false, "replace an existing encryption key (stored logins must then be re-entered)")
	fs.StringVar(&f.googleID, "google-client-id", "", "Google OAuth client ID")
	fs.StringVar(&f.googleSecret, "google-client-secret", "", "Google OAuth client secret (visible in shell history; prefer the prompt)")
	fs.StringVar(&f.msID, "microsoft-client-id", "", "Microsoft application (client) ID")
	fs.StringVar(&f.msSecret, "microsoft-client-secret", "", "Microsoft client secret (visible in shell history; prefer the prompt)")
	fs.StringVar(&f.tenant, "microsoft-tenant", "", "Microsoft tenant: common (default), consumers, organizations, a tenant ID or a domain")
	fs.StringVar(&f.dropboxID, "dropbox-client-id", "", "Dropbox app key")
	fs.StringVar(&f.dropboxSecret, "dropbox-client-secret", "", "Dropbox app secret (visible in shell history; prefer the prompt)")
	fs.Usage = func() {
		w := bio.errOut
		_, _ = fmt.Fprintln(w, "Usage: homebox backup-config [flags]")
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, "Generates the settings scheduled backups need as an env file: a credential encryption")
		_, _ = fmt.Fprintln(w, "key, the local backup root and, optionally, OAuth apps for Google Drive, OneDrive and")
		_, _ = fmt.Fprintln(w, "Dropbox. An existing file is updated in place: unrelated lines are kept and an existing")
		_, _ = fmt.Fprintln(w, "encryption key is never replaced unless you pass --rotate-key.")
		_, _ = fmt.Fprintln(w)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	say := func(format string, a ...any) { _, _ = fmt.Fprintf(bio.errOut, format+"\n", a...) }
	interactive := bio.tty && !f.nonInteractive

	// What is already in the target file, to merge with and to default from.
	// With --stdout the default ".env" is not read (it may be unrelated), but a
	// file named explicitly with --output is, so the key is not regenerated.
	outputGiven := false
	fs.Visit(func(fl *flag.Flag) {
		if fl.Name == "output" {
			outputGiven = true
		}
	})
	var existingRaw []byte
	existing := map[string]string{}
	if !f.toStdout || outputGiven {
		if b, err := os.ReadFile(f.output); err == nil {
			existingRaw, existing = b, backupsetup.ParseEnv(b)
		} else if !errors.Is(err, os.ErrNotExist) {
			say("cannot read %s: %v", f.output, err)
			return 1
		}
	}

	p := backupsetup.Params{
		BaseURL:              f.baseURL,
		BehindProxy:          f.behindProxy,
		LocalRoot:            f.localRoot,
		AllowCustomEndpoints: f.allowCustomEndpoints || isTrue(existing[backupsetup.EnvCustomEndpoints]),
		EncryptionKey:        f.key,
		RotateKey:            f.rotateKey,
		Google:               backupsetup.App{ClientID: f.googleID, ClientSecret: f.googleSecret},
		Microsoft:            backupsetup.App{ClientID: f.msID, ClientSecret: f.msSecret},
		MicrosoftTenant:      f.tenant,
		Dropbox:              backupsetup.App{ClientID: f.dropboxID, ClientSecret: f.dropboxSecret},
	}
	// Proxy trust already configured in the target file or this process's
	// environment need not be asked for again.
	p.TrustProxyAlready = isTrue(existing[backupsetup.EnvTrustProxy]) || isTrue(os.Getenv(backupsetup.EnvTrustProxy))

	if interactive {
		if err := promptBackupConfig(&p, existing, bio, say); err != nil {
			say("%v", err)
			return 1
		}
	}

	issues := p.Validate()
	for _, is := range issues {
		prefix := "warning: "
		if is.Level == backupsetup.Error {
			prefix = "error: "
		}
		say("%s%s", prefix, is.Message)
	}
	if backupsetup.HasErrors(issues) {
		say("\nnothing was written. Fix the errors above and run the command again.")
		return 1
	}

	plan := backupsetup.Build(p, existing)
	switch {
	case plan.KeyKept:
		say("keeping the existing encryption key (stored logins stay readable)")
	case plan.KeyRotated:
		say("WARNING: the encryption key is being replaced. Every stored SFTP, WebDAV, SMB and cloud-drive login becomes unreadable and must be re-entered.")
	case plan.KeyGenerated:
		say("generated a new encryption key. Keep it safe and unchanged: it seals stored logins.")
	}

	say("\nChanges to %s:", map[bool]string{true: "the output", false: f.output}[f.toStdout])
	for _, l := range backupsetup.Diff(existing, plan) {
		say("  %s", l)
	}

	merged := backupsetup.Merge(existingRaw, plan.Vars)
	envName := f.output
	if f.toStdout {
		envName = ".env"
	}

	switch {
	case f.dryRun:
		say("\n(dry run: nothing written)")
	case f.toStdout:
		_, _ = bio.out.Write(merged)
	default:
		if interactive && !f.yes && !confirm(bio, say, fmt.Sprintf("\nWrite these changes to %s?", f.output)) {
			say("nothing was written.")
			return 1
		}
		if err := writeEnvFile(f.output, existingRaw, merged); err != nil {
			say("error: %v", err)
			return 1
		}
		say("\nwrote %s (permissions 600).", f.output)
		if len(existingRaw) > 0 {
			say("the previous file is saved as %s.bak", f.output)
		}
	}

	printNextSteps(p, envName, say)
	return 0
}

func promptBackupConfig(p *backupsetup.Params, existing map[string]string, bio backupConfigIO, say func(string, ...any)) error {
	ask := func(label, def string) (string, error) {
		if def != "" {
			_, _ = fmt.Fprintf(bio.errOut, "%s [%s]: ", label, def)
		} else {
			_, _ = fmt.Fprintf(bio.errOut, "%s: ", label)
		}
		line, err := bio.in.ReadString('\n')
		if err != nil && line == "" {
			return "", fmt.Errorf("input ended before the questions were answered")
		}
		line = strings.TrimSpace(line)
		if line == "" {
			return def, nil
		}
		return line, nil
	}
	yesNo := func(label string, def bool) (bool, error) {
		d := "y/N"
		if def {
			d = "Y/n"
		}
		v, err := ask(label+" ("+d+")", "")
		if err != nil {
			return false, err
		}
		switch strings.ToLower(v) {
		case "":
			return def, nil
		case "y", "yes":
			return true, nil
		}
		return false, nil
	}
	secret := func(label string) (string, error) {
		_, _ = fmt.Fprintf(bio.errOut, "%s (hidden): ", label)
		return bio.readPass()
	}

	say("Scheduled backups setup. Press Enter to accept a [default] or skip an optional step.\n")

	var err error
	if p.LocalRoot == "" {
		if p.LocalRoot, err = ask("Folder inside the container for local-directory backups, e.g. /backups (blank = no local destinations)", existing[backupsetup.EnvLocalRoot]); err != nil {
			return err
		}
	}

	if !p.AllowCustomEndpoints {
		say("\nSFTP, WebDAV, SMB and S3-compatible (MinIO, NAS) destinations make the Homebox server connect to an address")
		say("a collection owner types in, anywhere the server can reach. They are off unless you allow them.")
		if p.AllowCustomEndpoints, err = yesNo("Allow SFTP, WebDAV, SMB and custom S3 endpoints", false); err != nil {
			return err
		}
	}

	if p.BaseURL == "" {
		def := ""
		if h := existing[backupsetup.EnvHostname]; h != "" {
			def = "https://" + h
		}
		if p.BaseURL, err = ask("Address Homebox is reached at, e.g. https://homebox.example.com (blank = no cloud drives)", def); err != nil {
			return err
		}
	}
	if p.BaseURL == "" {
		return nil
	}
	if strings.HasPrefix(p.BaseURL, "https://") && !p.BehindProxy && !p.TrustProxyAlready {
		say("\nHomebox itself serves plain HTTP, so an https address means a reverse proxy terminates TLS in front of it.")
		say("Homebox must trust that proxy's X-Forwarded-Host and X-Forwarded-Proto headers to build the right redirect URI.")
		if p.BehindProxy, err = yesNo("Trust the reverse proxy's forwarded headers (HBOX_OPTIONS_TRUST_PROXY=true)", true); err != nil {
			return err
		}
	}

	redirect := backupsetup.RedirectURI(p.BaseURL)
	do := func(name, key string, app *backupsetup.App) error {
		if app.Set() {
			return nil
		}
		on, err := yesNo("Set up "+name, false)
		if err != nil || !on {
			return err
		}
		say("\n%s", backupsetup.ProviderGuide(key, redirect))
		if app.ClientID, err = ask(name+" client ID", ""); err != nil {
			return err
		}
		app.ClientSecret, err = secret(name + " client secret")
		return err
	}
	if err := do("Google Drive", "google", &p.Google); err != nil {
		return err
	}
	if err := do("OneDrive", "microsoft", &p.Microsoft); err != nil {
		return err
	}
	if p.Microsoft.Set() && p.MicrosoftTenant == "" {
		if p.MicrosoftTenant, err = ask("Microsoft tenant (common, consumers, organizations, a tenant ID or a domain)", backupsetup.DefaultMicrosoftTen); err != nil {
			return err
		}
	}
	return do("Dropbox", "dropbox", &p.Dropbox)
}

func isTrue(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "t", "true", "yes", "on":
		return true
	}
	return false
}

func confirm(bio backupConfigIO, say func(string, ...any), q string) bool {
	_, _ = fmt.Fprintf(bio.errOut, "%s (y/N): ", q)
	line, _ := bio.in.ReadString('\n')
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	}
	return false
}

// writeEnvFile replaces the file atomically with owner-only permissions, after
// saving a copy of whatever was there.
func writeEnvFile(path string, previous, content []byte) error {
	if len(previous) > 0 {
		if err := os.WriteFile(path+".bak", previous, 0o600); err != nil {
			return fmt.Errorf("could not save %s.bak: %w", path, err)
		}
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".backup-config-*")
	if err != nil {
		return fmt.Errorf("cannot write in %s: %w", dir, err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func printNextSteps(p backupsetup.Params, envFile string, say func(string, ...any)) {
	say("\n--- docker compose ---\n%s", backupsetup.ComposeSnippet(envFile, p.LocalRoot))

	if len(p.Providers()) > 0 && p.BaseURL != "" {
		redirect := backupsetup.RedirectURI(p.BaseURL)
		say("--- register these OAuth apps (redirect URI: %s) ---", redirect)
		for _, prov := range p.Providers() {
			say("\n%s", backupsetup.ProviderGuide(prov, redirect))
		}
	}

	say("--- next steps ---")
	say("  1. Restart Homebox so it reads the new settings.")
	say("  2. Sign in as a collection owner: Collection > Tools > Backup & Restore.")
	say("  3. Add a destination and use \"Test connection\".")
	say("  Keep %s private and back up the encryption key separately.", envFile)
}
