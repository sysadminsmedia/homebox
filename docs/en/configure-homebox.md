# Configure Homebox

## Env Variables & Configuration

| Variable                             | Default                                                                    | Description                                                                            |
|--------------------------------------|----------------------------------------------------------------------------|----------------------------------------------------------------------------------------|
| HBOX_MODE                            | `production`                                                               | application mode used for runtime behavior  can be one of: `development`, `production` |
| HBOX_WEB_PORT                        | 7745                                                                       | port to run the web server on, if you're using docker do not change this               |
| HBOX_WEB_HOST                        |                                                                            | host to run the web server on, if you're using docker do not change this               |
| HBOX_OPTIONS_ALLOW_REGISTRATION      | true                                                                       | allow users to register themselves                                                     |
| HBOX_OPTIONS_AUTO_INCREMENT_ASSET_ID | true                                                                       | auto-increments the asset_id field for new items                                       |
| HBOX_OPTIONS_CURRENCY_CONFIG         |                                                                            | json configuration file containing additional currencie                                |
| HBOX_WEB_MAX_UPLOAD_SIZE             | 10                                                                         | maximum file upload size supported in MB                                               |
| HBOX_WEB_READ_TIMEOUT                | 10s                                                                        | Read timeout of HTTP sever                                                             |
| HBOX_WEB_WRITE_TIMEOUT               | 10s                                                                        | Write timeout of HTTP server                                                           |
| HBOX_WEB_IDLE_TIMEOUT                | 30s                                                                        | Idle timeout of HTTP server                                                            |
| HBOX_STORAGE_DATA                    | /data/                                                                     | path to the data directory, do not change this if you're using docker                  |
| HBOX_LOG_LEVEL                       | `info`                                                                     | log level to use, can be one of `trace`, `debug`, `info`, `warn`, `error`, `critical`  |
| HBOX_LOG_FORMAT                      | `text`                                                                     | log format to use, can be one of: `text`, `json`                                       |
| HBOX_MAILER_HOST                     |                                                                            | email host to use, if not set no email provider will be used                           |
| HBOX_MAILER_PORT                     | 587                                                                        | email port to use                                                                      |
| HBOX_MAILER_USERNAME                 |                                                                            | email user to use                                                                      |
| HBOX_MAILER_PASSWORD                 |                                                                            | email password to use                                                                  |
| HBOX_MAILER_FROM                     |                                                                            | email from address to use                                                              |
| HBOX_SWAGGER_HOST                    | 7745                                                                       | swagger host to use, if not set swagger will be disabled                               |
| HBOX_SWAGGER_SCHEMA                  | `http`                                                                     | swagger schema to use, can be one of: `http`, `https`                                  |
| HBOX_DATABASE_TYPE                   | sqlite3                                                                    | sets the correct database type (`sqlite3` or `postgres`)                               |
| HBOX_DATABASE_SQLITE_PATH            | ./.data/homebox.db?_pragma=busy_timeout=999&_pragma=journal_mode=WAL&_fk=1 | sets the directory path for Sqlite                                                     |
| HBOX_DATABASE_HOST                   |                                                                            | sets the hostname for a postgres database                                              |
| HBOX_DATABASE_PORT                   |                                                                            | sets the port for a postgres database                                                  |
| HBOX_DATABASE_USERNAME               |                                                                            | sets the username for a postgres connection                                            |
| HBOX_DATABASE_PASSWORD               |                                                                            | sets the password for a postgres connection                                            |
| HBOX_DATABASE_DATABASE               |                                                                            | sets the database for a postgres connection                                            |
| HBOX_DATABASE_SSL_MODE               |                                                                            | sets the sslmode for a postgres connection                                             |
| HBOX_ALLOW_ANALYTICS                 | `false`                                                                    | Opt-In basic non-identifiable analytics to assist with optimizing Homebox.             |

::: tip "CLI Arguments"
If you're deploying without docker you can use command line arguments to configure the application. Run `homebox --help` for more information.

```sh
Usage: api [options] [arguments]

OPTIONS
--mode/$HBOX_MODE                                                        <string>  (default: development)
--web-port/$HBOX_WEB_PORT                                                <string>  (default: 7745)
--web-host/$HBOX_WEB_HOST                                                <string>
--web-max-upload-size/$HBOX_WEB_MAX_UPLOAD_SIZE                          <int>     (default: 10)
--storage-data/$HBOX_STORAGE_DATA                                        <string>  (default: ./.data)
--storage-sqlite-url/$HBOX_STORAGE_SQLITE_URL                            <string>  (default: ./.data/homebox.db?_pragma=busy_timeout=999&_pragma=journal_mode=WAL&_fk=1)
--log-level/$HBOX_LOG_LEVEL                                              <string>  (default: info)
--log-format/$HBOX_LOG_FORMAT                                            <string>  (default: text)
--mailer-host/$HBOX_MAILER_HOST                                          <string>
--mailer-port/$HBOX_MAILER_PORT                                          <int>
--mailer-username/$HBOX_MAILER_USERNAME                                  <string>
--mailer-password/$HBOX_MAILER_PASSWORD                                  <string>
--mailer-from/$HBOX_MAILER_FROM                                          <string>
--swagger-host/$HBOX_SWAGGER_HOST                                        <string>  (default: localhost:7745)
--swagger-scheme/$HBOX_SWAGGER_SCHEME                                    <string>  (default: http)
--demo/$HBOX_DEMO                                                        <bool>
--debug-enabled/$HBOX_DEBUG_ENABLED                                      <bool>    (default: false)
--debug-port/$HBOX_DEBUG_PORT                                            <string>  (default: 4000)
--options-allow-registration/$HBOX_OPTIONS_ALLOW_REGISTRATION            <bool>    (default: true)
--options-auto-increment-asset-id/$HBOX_OPTIONS_AUTO_INCREMENT_ASSET_ID  <bool>    (default: true)
--options-currency-config/$HBOX_OPTIONS_CURRENCY_CONFIG                  <string>
--help/-h    display this help message
```
:::
