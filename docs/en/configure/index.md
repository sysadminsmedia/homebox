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
| HBOX_OPTIONS_CURRENCY_CONFIG            |                                                                            | json configuration file containing additional currencie                                                                                                                                   |
| HBOX_OPTIONS_ALLOW_ANALYTICS            | false                                                                      | Allows the homebox team to view extremely basic information about the system that your running on. This helps make decisions regarding builds and other general decisions.                |
| HBOX_WEB_MAX_UPLOAD                     | 10                                                                         | maximum file upload size supported in MB                                                                                                                                                  |
| HBOX_WEB_READ_TIMEOUT                   | 10s                                                                        | Read timeout of HTTP sever                                                                                                                                                                |
| HBOX_WEB_WRITE_TIMEOUT                  | 10s                                                                        | Write timeout of HTTP server                                                                                                                                                              |
| HBOX_WEB_IDLE_TIMEOUT                   | 30s                                                                        | Idle timeout of HTTP server                                                                                                                                                               |
| HBOX_STORAGE_CONN_STRING                | file://./                                                                  | path to the data directory, do not change this if you're using docker                                                                                                                     |
| HBOX_STORAGE_PREFIX_PATH                | .data                                                                      | prefix path for the storage, if not set the storage will be used as is                                                                                                                    |
| HBOX_LOG_LEVEL                          | `info`                                                                     | log level to use, can be one of `trace`, `debug`, `info`, `warn`, `error`, `critical`                                                                                                     |
| HBOX_LOG_FORMAT                         | `text`                                                                     | log format to use, can be one of: `text`, `json`                                                                                                                                          |
| HBOX_MAILER_HOST                        |                                                                            | email host to use, if not set no email provider will be used                                                                                                                              |
| HBOX_MAILER_PORT                        | 587                                                                        | email port to use                                                                                                                                                                         |
| HBOX_MAILER_USERNAME                    |                                                                            | email user to use                                                                                                                                                                         |
| HBOX_MAILER_PASSWORD                    |                                                                            | email password to use                                                                                                                                                                     |
| HBOX_MAILER_FROM                        |                                                                            | email from address to use                                                                                                                                                                 |
| HBOX_SWAGGER_HOST                       | 7745                                                                       | swagger host to use, if not set swagger will be disabled                                                                                                                                  |
| HBOX_SWAGGER_SCHEMA                     | `http`                                                                     | swagger schema to use, can be one of: `http`, `https`                                                                                                                                     |
| HBOX_DATABASE_DRIVER                    | sqlite3                                                                    | sets the correct database type (`sqlite3` or `postgres`)                                                                                                                                  |
| HBOX_DATABASE_SQLITE_PATH               | ./.data/homebox.db?_pragma=busy_timeout=999&_pragma=journal_mode=WAL&_fk=1 | sets the directory path for Sqlite                                                                                                                                                        |
| HBOX_DATABASE_HOST                      |                                                                            | sets the hostname for a postgres database                                                                                                                                                 |
| HBOX_DATABASE_PORT                      |                                                                            | sets the port for a postgres database                                                                                                                                                     |
| HBOX_DATABASE_USERNAME                  |                                                                            | sets the username for a postgres connection                                                                                                                                               |
| HBOX_DATABASE_PASSWORD                  |                                                                            | sets the password for a postgres connection                                                                                                                                               |
| HBOX_DATABASE_DATABASE                  |                                                                            | sets the database for a postgres connection                                                                                                                                               |
| HBOX_DATABASE_SSL_MODE                  |                                                                            | sets the sslmode for a postgres connection                                                                                                                                                |
| HBOX_OPTIONS_CHECK_GITHUB_RELEASE       | true                                                                       | check for new github releases                                                                                                                                                             |
| HBOX_LABEL_MAKER_WIDTH                  | 526                                                                        | width for generated labels in pixels                                                                                                                                                      |
| HBOX_LABEL_MAKER_HEIGHT                 | 200                                                                        | height for generated labels in pixels                                                                                                                                                     |
| HBOX_LABEL_MAKER_PADDING                | 32                                                                         | space between elements on label                                                                                                                                                           |
| HBOX_LABEL_MAKER_FONT_SIZE              | 32.0                                                                       | font size for label text                                                                                                                                                                  |
| HBOX_LABEL_MAKER_PRINT_COMMAND          |                                                                            | the command to use for printing labels. if empty, label printing is disabled. <span v-pre>`{{.FileName}}`</span> in the command will be replaced with the png filename of the label       |
| HBOX_LABEL_MAKER_DYNAMIC_LENGTH         | true                                                                       | allow label generation with open length. `HBOX_LABEL_MAKER_HEIGHT` is still used for layout and minimal height. If not used, long text may be cut off, but all labels have the same size. |
| HBOX_LABEL_MAKER_ADDITIONAL_INFORMATION |                                                                            | Additional information added to the label like name or phone number                                                                                                                       |
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

::: tip CLI Arguments
If you're deploying without docker you can use command line arguments to configure the application. Run `homebox --help`
for more information.

```sh
Usage: api [options] [arguments]

OPTIONS
--mode/$HBOX_MODE                                                             <string>  (default: development)
--web-port/$HBOX_WEB_PORT                                                     <string>  (default: 7745)
--web-host/$HBOX_WEB_HOST                                                     <string>
--web-max-file-upload/$HBOX_WEB_MAX_FILE_UPLOAD                               <int>     (default: 10)
--storage-conn-string/$HBOX_STORAGE_CONN_STRING                               <string>  (default: file://./)
--storage-prefix-path/$HBOX_STORAGE_PREFIX_PATH                               <string>  (default: .data)
--log-level/$HBOX_LOG_LEVEL                                                   <string>  (default: info)
--log-format/$HBOX_LOG_FORMAT                                                 <string>  (default: text)
--mailer-host/$HBOX_MAILER_HOST                                               <string>
--mailer-port/$HBOX_MAILER_PORT                                               <int>
--mailer-username/$HBOX_MAILER_USERNAME                                       <string>
--mailer-password/$HBOX_MAILER_PASSWORD                                       <string>
--mailer-from/$HBOX_MAILER_FROM                                               <string>
--swagger-host/$HBOX_SWAGGER_HOST                                             <string>  (default: localhost:7745)
--swagger-scheme/$HBOX_SWAGGER_SCHEME                                         <string>  (default: http)
--demo/$HBOX_DEMO                                                             <bool>
--debug-enabled/$HBOX_DEBUG_ENABLED                                           <bool>    (default: false)
--debug-port/$HBOX_DEBUG_PORT                                                 <string>  (default: 4000)
--database-driver/$HBOX_DATABASE_DRIVER                                       <string>  (default: sqlite3)
--database-sqlite-path/$HBOX_DATABASE_SQLITE_PATH                             <string>  (default: ./.data/homebox.db?_pragma=busy_timeout=999&_pragma=journal_mode=WAL&_fk=1)
--database-host/$HBOX_DATABASE_HOST                                           <string>
--database-port/$HBOX_DATABASE_PORT                                           <string>
--database-username/$HBOX_DATABASE_USERNAME                                   <string>
--database-password/$HBOX_DATABASE_PASSWORD                                   <string>
--database-database/$HBOX_DATABASE_DATABASE                                   <string>
--database-ssl-mode/$HBOX_DATABASE_SSL_MODE                                   <string>
--options-allow-registration/$HBOX_OPTIONS_ALLOW_REGISTRATION                 <bool>    (default: true)
--options-auto-increment-asset-id/$HBOX_OPTIONS_AUTO_INCREMENT_ASSET_ID       <bool>    (default: true)
--options-currency-config/$HBOX_OPTIONS_CURRENCY_CONFIG                       <string>
--options-check-github-release/$HBOX_OPTIONS_CHECK_GITHUB_RELEASE             <bool>    (default: true)
--options-allow-analytics/$HBOX_OPTIONS_ALLOW_ANALYTICS                       <bool>    (default: false)
--label-maker-width/$HBOX_LABEL_MAKER_WIDTH                                   <int>     (default: 526)
--label-maker-height/$HBOX_LABEL_MAKER_HEIGHT                                 <int>     (default: 200)
--label-maker-padding/$HBOX_LABEL_MAKER_PADDING                               <int>     (default: 32)
--label-maker-margin/$HBOX_LABEL_MAKER_MARGIN                                 <int>       (default: 32)
--label-maker-font-size/$HBOX_LABEL_MAKER_FONT_SIZE                           <float>   (default: 32.0)
--label-maker-print-command/$HBOX_LABEL_MAKER_PRINT_COMMAND                   <string>
--label-maker-additional-information/$HBOX_LABEL_MAKER_DYNAMIC_LENGTH         <string>  (default: true) 
--label-maker-additional-information/$HBOX_LABEL_MAKER_ADDITIONAL_INFORMATION <string>
--thumbnail-enabled/$HBOX_THUMBNAIL_ENABLED                                   <bool>    (default: true)
--thumbnail-width/$HBOX_THUMBNAIL_WIDTH                                       <int>     (default: 500)
--thumbnail-height/$HBOX_THUMBNAIL_HEIGHT                                     <int>     (default: 500)
--help/-h    display this help message
```

:::
