package backupsetup

import (
	"fmt"
	"strings"
)

// ProviderGuide returns console steps for registering the OAuth app for one
// provider, with the exact redirect URI filled in.
func ProviderGuide(provider, redirectURI string) string {
	var b strings.Builder
	switch provider {
	case "google":
		b.WriteString("Google Drive\n")
		b.WriteString("  1. Google Cloud console > APIs & Services > Library: enable the \"Google Drive API\".\n")
		b.WriteString("  2. OAuth consent screen: add the scope .../auth/drive.file and PUBLISH the app (\"In production\").\n")
		b.WriteString("     While it is in Testing, Google expires refresh tokens after 7 days and backups would stop.\n")
		b.WriteString("  3. Credentials > Create credentials > OAuth client ID > type \"Web application\".\n")
		fmt.Fprintf(&b, "  4. Authorized redirect URI: %s\n", redirectURI)
		b.WriteString("  5. Copy the client ID and client secret.\n")
	case "microsoft":
		b.WriteString("OneDrive\n")
		b.WriteString("  1. Microsoft Entra admin center > App registrations > New registration.\n")
		b.WriteString("  2. Choose the supported account types (personal + work accounts, or just one).\n")
		fmt.Fprintf(&b, "  3. Redirect URI: platform \"Web\", value %s\n", redirectURI)
		b.WriteString("  4. API permissions > Microsoft Graph > Delegated: Files.ReadWrite.AppFolder and offline_access.\n")
		b.WriteString("  5. Certificates & secrets > New client secret (it expires, so note the date to renew it).\n")
		b.WriteString("  6. Copy the Application (client) ID and the secret VALUE.\n")
	case "dropbox":
		b.WriteString("Dropbox\n")
		b.WriteString("  1. Dropbox App Console > Create app > \"Scoped access\" > access type \"App folder\".\n")
		b.WriteString("  2. Permissions: enable files.content.write, files.content.read and account_info.read.\n")
		fmt.Fprintf(&b, "  3. Settings > Redirect URIs: add %s\n", redirectURI)
		b.WriteString("  4. Copy the App key and App secret.\n")
	}
	return b.String()
}

// ComposeSnippet returns a docker compose fragment that loads the env file and,
// when local destinations are enabled, mounts a host folder at the root.
func ComposeSnippet(envFile, localRoot string) string {
	var b strings.Builder
	b.WriteString("services:\n  homebox:\n")
	fmt.Fprintf(&b, "    env_file:\n      - %s\n", envFile)
	if localRoot != "" {
		b.WriteString("    volumes:\n")
		fmt.Fprintf(&b, "      - /path/on/the/host/homebox-backups:%s   # e.g. a mounted NAS share\n", localRoot)
	}
	return b.String()
}
