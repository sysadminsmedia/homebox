---
outline: false
aside: false
---

# Configure Homebox

## Env Variables & Configuration

| Variable                                | Default                                                                    | Description                                                                                                                                                                               |
|-----------------------------------------|----------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| HBOX_MODE                               | `production`                                                               | application mode used for runtime behavior  can be one of: `development`, `production`                                                                                                    |
| HBOX_WEB_PORT                           | 7745                                                                       | port to run the web server on, if you're using docker do not change this                                                                                                                  |
| HBOX_WEB_HOST                           |                                                                            | host to run the web server on, if you're using docker do not change this. see below for examples                                                                                          |
| HBOX_OPTIONS_ALLOW_REGISTRATION         | true                                                                       | allow users to register themselves                                                                                                                                                        |
| HBOX_OPTIONS_AUTO_INCREMENT_ASSET_ID    | true                                                                       | auto-increments the asset_id field for new items                                                                                                                                          |
| HBOX_OPTIONS_CURRENCY_CONFIG            |                                                                            | json configuration file containing additional currencies                                                                                                                                   |
| HBOX_OPTIONS_ALLOW_ANALYTICS            | false                                                                      | Allows the homebox team to view extremely basic information about the system that your running on. This helps make decisions regarding builds and other general decisions.                |
| HBOX_WEB_MAX_UPLOAD_SIZE                | 10                                                                         | maximum file upload size supported in MB                                                                                                                                                  |
| HBOX_WEB_READ_TIMEOUT                   | 10s                                                                        | Read timeout of HTTP sever                                                                                                                                                                |
| HBOX_WEB_WRITE_TIMEOUT                  | 10s                                                                        | Write timeout of HTTP server                                                                                                                                                              |
| HBOX_WEB_IDLE_TIMEOUT                   | 30s                                                                        | Idle timeout of HTTP server                                                                                                                                                               |
| HBOX_STORAGE_CONN_STRING                | file:///./                                                                 | path to the data directory, do not change this if you're using docker                                                                                                                     |
| HBOX_STORAGE_PREFIX_PATH                | .data                                                                      | prefix path for the storage, if not set the storage will be used as is                                                                                                                    |
| HBOX_LOG_LEVEL                          | `info`                                                                     | log level to use, can be one of `trace`, `debug`, `info`, `warn`, `error`, `critical`                                                                                                     |
| HBOX_LOG_FORMAT                         | `text`                                                                     | log format to use, can be one of: `text`, `json`                                                                                                                                          |
| HBOX_MAILER_HOST                        |                                                                            | email host to use, if not set no email provider will be used                                                                                                                              |
| HBOX_MAILER_PORT                        | 587                                                                        | email port to use                                                                                                                                                                         |
| HBOX_MAILER_USERNAME                    |                                                                            | email user to use                                                                                                                                                                         |
| HBOX_MAILER_PASSWORD                    |                                                                            | email password to use                                                                                                                                                                     |
| HBOX_MAILER_FROM                        |                                                                            | email from address to use                                                                                                                                                                 |
| HBOX_DATABASE_DRIVER                    | sqlite3                                                                    | sets the correct database type (`sqlite3` or `postgres`)                                                                                                                                  |
| HBOX_DATABASE_SQLITE_PATH               | ./.data/homebox.db?_pragma=busy_timeout=999&_pragma=journal_mode=WAL&_fk=1&_time_format=sqlite | sets the directory path for Sqlite                                                                                                                                                        |
| HBOX_DATABASE_HOST                      |                                                                            | sets the hostname for a postgres database                                                                                                                                                 |
| HBOX_DATABASE_PORT                      |                                                                            | sets the port for a postgres database                                                                                                                                                     |
| HBOX_DATABASE_USERNAME                  |                                                                            | sets the username for a postgres connection (optional if using cert auth)                                                                                                                 |
| HBOX_DATABASE_PASSWORD                  |                                                                            | sets the password for a postgres connection (optional if using cert auth)                                                                                                                 |
| HBOX_DATABASE_DATABASE                  |                                                                            | sets the database for a postgres connection                                                                                                                                               |
| HBOX_DATABASE_SSL_MODE                  | prefer                                                                     | sets the sslmode for a postgres connection                                                                                                                                                |
| HBOX_DATABASE_SSL_CERT                  |                                                                            | sets the sslcert for a postgres connection (should be a path)                                                                                                                             |
| HBOX_DATABASE_SSL_KEY                   |                                                                            | sets the sslkey for a postgres connection (should be a path)                                                                                                                              |
| HBOX_DATABASE_SSL_ROOT_CERT             |                                                                            | sets the sslrootcert for a postgres connection (should be a path)                                                                                                                         |
| HBOX_OPTIONS_GITHUB_RELEASE_CHECK       | true                                                                       | check for new github releases                                                                                                                                                             |
| HBOX_OPTIONS_ALLOW_LOCAL_LOGIN          | true                                                                       | allow users to login with username/password when OIDC is enabled                                                                                                                          |
| HBOX_OPTIONS_TRUST_PROXY                | false                                                                      | trust proxy headers for determining request scheme (X-Forwarded-Proto)                                                                                                                    |
| HBOX_OPTIONS_HOSTNAME                   |                                                                            | override hostname used for OIDC redirect URLs and other absolute URLs                                                                                                                     |
| HBOX_OIDC_ENABLED                       | false                                                                      | enable OpenID Connect (OIDC) authentication                                                                                                                                               |
| HBOX_OIDC_ISSUER_URL                    |                                                                            | OIDC provider issuer URL (required when OIDC is enabled)                                                                                                                                  |
| HBOX_OIDC_CLIENT_ID                     |                                                                            | OIDC client ID (required when OIDC is enabled)                                                                                                                                            |
| HBOX_OIDC_CLIENT_SECRET                 |                                                                            | OIDC client secret (required when OIDC is enabled)                                                                                                                                        |
| HBOX_OIDC_SCOPE                         | openid profile email                                                       | OIDC scopes to request from the provider                                                                                                                                                  |
| HBOX_OIDC_ALLOWED_GROUPS                |                                                                            | comma-separated list of groups that are allowed to login (empty means all groups allowed)                                                                                                 |
| HBOX_OIDC_AUTO_REDIRECT                 | false                                                                      | auto redirect to OIDC authentication (automatically redirects to OIDC provider, but does not disable local login. See HBOX_OPTIONS_ALLOW_LOCAL_LOGIN)                                     |
| HBOX_OIDC_VERIFY_EMAIL                  | false                                                                      | require email verification from OIDC provider                                                                                                                                             |
| HBOX_OIDC_GROUP_CLAIM                   | groups                                                                     | name of the claim in the ID token that contains user groups                                                                                                                               |
| HBOX_OIDC_EMAIL_CLAIM                   | email                                                                      | name of the claim in the ID token that contains user email                                                                                                                                |
| HBOX_OIDC_NAME_CLAIM                    | name                                                                       | name of the claim in the ID token that contains user display name                                                                                                                         |
| HBOX_OIDC_EMAIL_VERIFIED_CLAIM          | email_verified                                                             | name of the claim in the ID token that contains user email verification status                                                                                                            |
| HBOX_OIDC_BUTTON_TEXT                   | Sign in with OIDC                                                          | text displayed on the OIDC login button                                                                                                                                                   |
| HBOX_OIDC_STATE_EXPIRY                  | 10m                                                                        | how long OIDC state parameters are valid (for CSRF protection)                                                                                                                            |
| HBOX_OIDC_REQUEST_TIMEOUT               | 30s                                                                        | timeout for OIDC provider requests (token exchange, userinfo, etc.)                                                                                                                       |
| HBOX_LABEL_MAKER_WIDTH                  | 526                                                                        | width for generated labels in pixels                                                                                                                                                      |
| HBOX_LABEL_MAKER_HEIGHT                 | 200                                                                        | height for generated labels in pixels                                                                                                                                                     |
| HBOX_LABEL_MAKER_PADDING                | 32                                                                         | space between elements on label                                                                                                                                                           |
| HBOX_LABEL_MAKER_FONT_SIZE              | 32.0                                                                       | font size for label text                                                                                                                                                                  |
| HBOX_LABEL_MAKER_PRINT_COMMAND          |                                                                            | the command to use for printing labels. if empty, label printing is disabled. <span v-pre>`{{.FileName}}`</span> in the command will be replaced with the png filename of the label       |
| HBOX_LABEL_MAKER_DYNAMIC_LENGTH         | true                                                                       | allow label generation with open length. `HBOX_LABEL_MAKER_HEIGHT` is still used for layout and minimal height. If not used, long text may be cut off, but all labels have the same size. |
| HBOX_LABEL_MAKER_ADDITIONAL_INFORMATION |                                                                            | Additional information added to the label like name or phone number                                                                                                                       |
| HBOX_LABEL_MAKER_REGULAR_FONT_PATH      |                                                                            | path to regular font file for label generation (e.g., `/fonts/NotoSansKR-Regular.ttf`). If not set, uses embedded font. Supports TTF format.                                             |
| HBOX_LABEL_MAKER_BOLD_FONT_PATH         |                                                                            | path to bold font file for label generation (e.g., `/fonts/NotoSansKR-Bold.ttf`). If not set, uses embedded font. Supports TTF format.                                                   |
| HBOX_THUMBNAIL_ENABLED                  | true                                                                       | enable thumbnail generation for images, supports PNG, JPEG, AVIF, WEBP, GIF file types                                                                                                    |
| HBOX_THUMBNAIL_WIDTH                    | 500                                                                        | width for generated thumbnails in pixels                                                                                                                                                  |
| HBOX_THUMBNAIL_HEIGHT                   | 500                                                                        | height for generated thumbnails in pixels                                                                                                                                                 |

### HBOX_WEB_HOST examples

| Value                       | Notes                                                      |
|-----------------------------|------------------------------------------------------------|
| 0.0.0.0                     | Visible all interfaces (default behaviour)                 |
| 127.0.0.1                   | Only visible on same host                                  |
| 100.64.0.1                  | Only visible on a specific interface (e.g., VPN in a VPS). |
| unix?path=/run/homebox.sock | Listen on unix socket at specified path                    |
| sysd?name=homebox.socket    | Listen on systemd socket                                   |

For unix and systemd socket address syntax and available options, see the [anyhttp address-syntax documentation](https://pkg.go.dev/go.balki.me/anyhttp#readme-address-syntax).

#### Private network example

Below example starts homebox in an isolated network. The process cannot make
any external requests (including check for newer release) and thus more secure.

```bash
❯ sudo systemd-run --property=PrivateNetwork=yes --uid $UID --pty --same-dir --wait --collect homebox --web-host "unix?path=/run/user/$UID/homebox.sock"
Running as unit: run-p74482-i74483.service
Press ^] three times within 1s to disconnect TTY.
2025/07/11 22:33:29 goose: no migrations to run. current version: 20250706190000
10:33PM INF ../../../go/src/app/app/api/handlers/v1/v1_ctrl_auth.go:98 > registering auth provider name=local
10:33PM INF ../../../go/src/app/app/api/main.go:275 > Server is running on unix?path=/run/user/1000/homebox.sock
10:33PM ERR ../../../go/src/app/app/api/main.go:403 > failed to get latest github release error="failed to make latest version request: Get \"https://api.github.com/repos/sysadminsmedia/homebox/releases/l
atest\": dial tcp: lookup api.github.com on [::1]:53: read udp [::1]:50951->[::1]:53: read: connection refused"
10:33PM INF ../../../go/src/app/internal/web/mid/logger.go:36 > request received method=GET path=/ rid=hname/PoXyRgt6ol-000001
10:33PM INF ../../../go/src/app/internal/web/mid/logger.go:41 > request finished method=GET path=/ rid=hname/PoXyRgt6ol-000001 status=0
```

#### Systemd socket example

In the example below, Homebox listens on a systemd socket securely so that only
the webserver (Caddy) can access it. Other processes/containers on the host
cannot connect to Homebox directly, bypassing the webserver.

File: homebox.socket
```systemd
# /usr/local/lib/systemd/system/homebox.socket
[Unit]
Description=Homebox socket

[Socket]
ListenStream=/run/homebox.sock
SocketGroup=caddy
SocketMode=0660

[Install]
WantedBy=sockets.target
```

File: homebox.service
```systemd
# /usr/local/lib/systemd/system/homebox.service
[Unit]
Description=Homebox
After=network.target
Documentation=https://homebox.software

[Service]
DynamicUser=yes
StateDirectory=homebox
Environment=HBOX_WEB_HOST=sysd?name=homebox.socket
WorkingDirectory=/var/lib/homebox

ExecStart=/usr/local/bin/homebox

NoNewPrivileges=yes
CapabilityBoundingSet=
RestrictNamespaces=true
SystemCallFilter=@system-service
```
Usage:

```bash
systemctl start homebox.socket
```

::: warning Security Considerations
For postgreSQL in production:

- Do not use the default `postgres` user
- Do not use the default `postgres` database
- Always use a strong unique password
- Always use SSL (`sslmode=require` or `sslmode=verify-full`)
- Consider using a connection pooler like `pgbouncer`

For SQLite in production:

- Secure file permissions for the database file (e.g. `chmod 600`)
- Use a secure directory for the database file
- Use a secure backup strategy
- Monitor the file size and consider using a different database for large installations
  :::

## OIDC Configuration

HomeBox supports OpenID Connect (OIDC) authentication, allowing users to login using external identity providers like Keycloak, Authentik, Google, Microsoft, etc.

### Basic OIDC Setup

1. **Enable OIDC**: Set `HBOX_OIDC_ENABLED=true`
2. **Provider Configuration**: Set the required provider details:
   - `HBOX_OIDC_ISSUER_URL`: Your OIDC provider's issuer URL
   - `HBOX_OIDC_CLIENT_ID`: Client ID from your OIDC provider
   - `HBOX_OIDC_CLIENT_SECRET`: Client secret from your OIDC provider

3. **Configure Redirect URI**: In your OIDC provider, set the redirect URI to:
   `https://your-homebox-domain.com/api/v1/users/login/oidc/callback`

### Advanced OIDC Configuration

- **Group Authorization**: Use `HBOX_OIDC_ALLOWED_GROUPS` to restrict access to specific groups
- **Custom Claims**: Configure `HBOX_OIDC_GROUP_CLAIM`, `HBOX_OIDC_EMAIL_CLAIM`, and `HBOX_OIDC_NAME_CLAIM` if your provider uses different claim names
- **Auto Redirect to OIDC**: Set `HBOX_OIDC_AUTO_REDIRECT=true` to automatically redirect users directly to OIDC
- **Local Login**: Set `HBOX_OPTIONS_ALLOW_LOCAL_LOGIN=false` to completely disable username/password login
- **Email Verification**: Set `HBOX_OIDC_VERIFY_EMAIL=true` to require email verification from the OIDC provider

### Security Considerations

::: warning OIDC Security
- Store `HBOX_OIDC_CLIENT_SECRET` securely (use environment variables, not config files)
- Use HTTPS for production deployments
- Configure proper redirect URIs in your OIDC provider
- Consider setting `HBOX_OIDC_ALLOWED_GROUPS` for group-based access control
:::

::: tip CLI Arguments
If you're deploying without docker you can use command line arguments to configure the application. Run `homebox --help`
for more information.

```sh
Options:
      --barcode-token-barcodespider         <string>                                                                                                               
      --database-database                   <string>                                                                                                               
      --database-driver                     <string>    (default: sqlite3)                                                                                         
      --database-host                       <string>                                                                                                               
      --database-password                   <string>                                                                                                               
      --database-port                       <string>                                                                                                               
      --database-pub-sub-conn-string        <string>    (default: mem://{{ .Topic }})                                                                              
      --database-sqlite-path                <string>    (default: ./.data/homebox.db?_pragma=busy_timeout=999&_pragma=journal_mode=WAL&_fk=1&_time_format=sqlite)  
      --database-ssl-cert                   <string>                                                                                                               
      --database-ssl-key                    <string>                                                                                                               
      --database-ssl-mode                   <string>    (default: require)                                                                                         
      --database-ssl-root-cert              <string>                                                                                                               
      --database-username                   <string>                                                                                                               
      --debug-enabled                       <bool>      (default: false)                                                                                           
      --debug-port                          <string>    (default: 4000)                                                                                            
      --demo                                <bool>                                                                                                                 
  -h, --help                                                                                                                                                       display this help message
      --label-maker-additional-information  <string>                                                                                                               
      --label-maker-bold-font-path          <string>                                                                                                               
      --label-maker-dynamic-length          <bool>      (default: true)                                                                                            
      --label-maker-font-size               <float>     (default: 32.0)                                                                                            
      --label-maker-height                  <int>       (default: 200)                                                                                             
      --label-maker-label-service-timeout   <int>                                                                                                                  
      --label-maker-label-service-url       <string>                                                                                                               
      --label-maker-margin                  <int>       (default: 32)                                                                                              
      --label-maker-padding                 <int>       (default: 32)                                                                                              
      --label-maker-print-command           <string>                                                                                                               
      --label-maker-regular-font-path       <string>                                                                                                               
      --label-maker-width                   <int>       (default: 526)                                                                                             
      --log-format                          <string>    (default: text)                                                                                            
      --log-level                           <string>    (default: info)                                                                                            
      --mailer-from                         <string>                                                                                                               
      --mailer-host                         <string>                                                                                                               
      --mailer-password                     <string>                                                                                                               
      --mailer-port                         <int>                                                                                                                  
      --mailer-username                     <string>                                                                                                               
      --mode                                <string>    (default: development)                                                                                     
      --oidc-allowed-groups                 <string>                                                                                                               
      --oidc-auto-redirect                  <bool>      (default: false)                                                                                           
      --oidc-button-text                    <string>    (default: Sign in with OIDC)                                                                               
      --oidc-client-id                      <string>                                                                                                               
      --oidc-client-secret                  <string>                                                                                                               
      --oidc-email-claim                    <string>    (default: email)                                                                                           
      --oidc-email-verified-claim           <string>    (default: email_verified)                                                                                  
      --oidc-enabled                        <bool>      (default: false)                                                                                           
      --oidc-group-claim                    <string>    (default: groups)                                                                                          
      --oidc-issuer-url                     <string>                                                                                                               
      --oidc-name-claim                     <string>    (default: name)                                                                                            
      --oidc-request-timeout                <duration>  (default: 30s)                                                                                             
      --oidc-scope                          <string>    (default: openid profile email)                                                                            
      --oidc-state-expiry                   <duration>  (default: 10m)                                                                                             
      --oidc-verify-email                   <bool>      (default: false)                                                                                           
      --options-allow-analytics             <bool>      (default: false)                                                                                           
      --options-allow-local-login           <bool>      (default: true)                                                                                            
      --options-allow-registration          <bool>      (default: true)                                                                                            
      --options-auto-increment-asset-id     <bool>      (default: true)                                                                                            
      --options-currency-config             <string>                                                                                                               
      --options-github-release-check        <bool>      (default: true)                                                                                            
      --options-hostname                    <string>                                                                                                               
      --options-trust-proxy                 <bool>      (default: false)                                                                                           
      --storage-conn-string                 <string>    (default: file:///./)                                                                                      
      --storage-prefix-path                 <string>    (default: .data)                                                                                           
      --thumbnail-enabled                   <bool>      (default: true)                                                                                            
      --thumbnail-height                    <int>       (default: 500)                                                                                             
      --thumbnail-width                     <int>       (default: 500)                                                                                             
  -v, --version                                                                                                                                                    display version
      --web-host                            <string>                                                                                                               
      --web-idle-timeout                    <duration>  (default: 30s)                                                                                             
      --web-max-upload-size                 <int>       (default: 10)                                                                                              
      --web-port                            <string>    (default: 7745)                                                                                            
      --web-read-timeout                    <duration>  (default: 10s)                                                                                             
      --web-write-timeout                   <duration>  (default: 10s)                                                                                         
```

:::
