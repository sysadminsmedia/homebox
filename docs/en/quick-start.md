# Quick Start

## Docker Run

Great for testing out the application, but not recommended for stable use. Checkout the docker-compose for the recommended deployment.

For each image there are two tags, respectively the regular tag and $TAG-rootless, which uses a non-root image.

```sh
# If using the rootless image, ensure data
# folder has correct permissions
$ mkdir -p /path/to/data/folder
$ chown 65532:65532 -R /path/to/data/folder
# ---------------------------------------
# Run the image
$ docker run -d \
  --name homebox \
  --restart unless-stopped \
  --publish 3100:7745 \
  --env TZ=Europe/Bucharest \
  --volume /path/to/data/folder/:/data \
  ghcr.io/sysadminsmedia/homebox:latest
# ghcr.io/sysadminsmedia/homebox:latest-rootless

```

## Docker-Compose

```yaml
services:
  homebox:
    image: ghcr.io/sysadminsmedia/homebox:latest
#   image: ghcr.io/sysadminsmedia/homebox:latest-rootless
    container_name: homebox
    restart: always
    environment:
    - HBOX_LOG_LEVEL=info
    - HBOX_LOG_FORMAT=text
    - HBOX_WEB_MAX_UPLOAD_SIZE=10
    volumes:
      - homebox-data:/data/
    ports:
      - 3100:7745

volumes:
   homebox-data:
     driver: local
```
::: info
If you use the `rootless` image, and instead of using named volumes you would prefer using a hostMount directly (e.g., `volumes: [ /path/to/data/folder:/data ]`) you need to `chown` the chosen directory in advance to the `65532` user (as shown in the Docker example above).
:::

::: warning
If you have previously set up docker compose with the `HBOX_WEB_READ_TIMEOUT`, `HBOX_WEB_WRITE_TIMEOUT`, or `HBOX_IDLE_TIMEOUT` options, and you were previously using the hay-kot image, please note that you will have to add an `s` for seconds or `m` for minutes to the end of the integers. A dependency update removed the defaultation to seconds and it now requires an explicit duration time.
:::

## Env Variables & Configuration

| Variable                             | Default                | Description                                                                        |
| ------------------------------------ | ---------------------- | ---------------------------------------------------------------------------------- |
| HBOX_MODE                            | production             | application mode used for runtime behavior  can be one of: development, production |
| HBOX_WEB_PORT                        | 7745                   | port to run the web server on, if you're using docker do not change this           |
| HBOX_WEB_HOST                        |                        | host to run the web server on, if you're using docker do not change this           |
| HBOX_OPTIONS_ALLOW_REGISTRATION      | true                   | allow users to register themselves                                                 |
| HBOX_OPTIONS_AUTO_INCREMENT_ASSET_ID | true                   | auto-increments the asset_id field for new items                                   |
| HBOX_OPTIONS_CURRENCY_CONFIG         |                        | json configuration file containing additional currencie                            |
| HBOX_WEB_MAX_UPLOAD_SIZE             | 10                     | maximum file upload size supported in MB                                           |
| HBOX_WEB_READ_TIMEOUT                | 10s                    | Read timeout of HTTP sever                                                         |
| HBOX_WEB_WRITE_TIMEOUT               | 10s                    | Write timeout of HTTP server                                                       |
| HBOX_WEB_IDLE_TIMEOUT                | 30s                    | Idle timeout of HTTP server                                                        |
| HBOX_STORAGE_DATA                    | /data/                 | path to the data directory, do not change this if you're using docker              |
| HBOX_STORAGE_SQLITE_URL              | /data/homebox.db?_fk=1 | sqlite database url, if you're using docker do not change this                     |
| HBOX_LOG_LEVEL                       | info                   | log level to use, can be one of trace, debug, info, warn, error, critical         |
| HBOX_LOG_FORMAT                      | text                   | log format to use, can be one of: text, json                                       |
| HBOX_MAILER_HOST                     |                        | email host to use, if not set no email provider will be used                       |
| HBOX_MAILER_PORT                     | 587                    | email port to use                                                                  |
| HBOX_MAILER_USERNAME                 |                        | email user to use                                                                  |
| HBOX_MAILER_PASSWORD                 |                        | email password to use                                                              |
| HBOX_MAILER_FROM                     |                        | email from address to use                                                          |
| HBOX_SWAGGER_HOST                    | 7745                   | swagger host to use, if not set swagger will be disabled                           |
| HBOX_SWAGGER_SCHEMA                  | http                   | swagger schema to use, can be one of: http, https                                  |

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
--storage-sqlite-url/$HBOX_STORAGE_SQLITE_URL                            <string>  (default: ./.data/homebox.db?_fk=1)
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
