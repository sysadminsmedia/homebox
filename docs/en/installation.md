# Installation

There are two main ways to run the application.

1. As a [Docker](https://www.docker.com/) container.
2. Using the correct executable for your platform by downloading it from the [Releases](https://github.com/sysadminsmedia/homebox/releases).

::: info Configuration Options
The application can be configured using environment variables. You can find a list of all available options in the [configuration section](./configure/index).
:::

## Docker

The following instructions assume Docker is already installed on your system. See [(Docker's official installation guide)](https://docs.docker.com/engine/install/)

The official image is `ghcr.io/sysadminsmedia/homebox:latest`. For each image there are two tags, respectively the regular tag and $TAG-rootless, which uses a non-root image.

### Docker Run

Great for testing out the application, but not recommended for stable use. Checkout the docker-compose below for the recommended deployment.


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
  --env HBOX_OPTIONS_ALLOW_ANALYTICS=false \
  --volume /path/to/data/folder/:/data \
  ghcr.io/sysadminsmedia/homebox:latest
# ghcr.io/sysadminsmedia/homebox:latest-rootless
```

### Docker Compose

1. Create a `docker-compose.yml` file.

<ConfigEditor />

::: info
If you use the `rootless` image, and instead of using named volumes you would prefer using a hostMount directly (e.g., `volumes: [ /path/to/data/folder:/data ]`) you need to `chown` the chosen directory in advance to the `65532` user (as shown in the Docker example above).
:::

::: warning
If you have previously set up docker compose with the `HBOX_WEB_READ_TIMEOUT`, `HBOX_WEB_WRITE_TIMEOUT`, or `HBOX_IDLE_TIMEOUT` options, and you were previously using the hay-kot image, please note that you will have to add an `s` for seconds or `m` for minutes to the end of the integers. A dependency update removed the defaultation to seconds and it now requires an explicit duration time.
:::

2. While in the same folder as docker-compose.yml, start the container by running:

```bash
docker compose up --detach
```

3. Navigate to `http://server.local.ip.address:3100/` to access the web interface. (replace with the right IP address).

You can learn more about Docker by [reading the official Docker documentation.](https://docs.docker.com/)

## Windows

1. Download the appropriate release for your CPU architecture from the [releases page on GitHub](https://github.com/sysadminsmedia/homebox/releases).
2. Extract the archive.
3. Run `homebox.exe`. This will start the server on port 7745.
4. You can test it by accessing http://localhost:7745.

## Linux

1. Download the appropriate release for your CPU architecture from the [releases page on GitHub](https://github.com/sysadminsmedia/homebox/releases).
2. Extract the archive.
3. Run the `homebox` executable.
4. The web interface will be accessible on port 7745 by default. Access the page by navigating to `http://server.local.ip.address:7745/` (replace with the right ip address)

## macOS

1. Download the appropriate release for your CPU architecture from the [releases page on GitHub](https://github.com/sysadminsmedia/homebox/releases). (Use `homebox_Darwin_x86_64.tar.gz` for Intel-based macs and `homebox_Darwin_arm64.tar.gz` for Apple Silicon)
2. Extract the archive.
3. Run the `homebox` executable.
4. The web interface will be accessible on port 7745 by default. Access the page by navigating to `http://local.ip.address:7745/` (replace with the right ip address)

<script setup>
  import ConfigEditor from '../.vitepress/components/ConfigEditor.vue'
</script>
