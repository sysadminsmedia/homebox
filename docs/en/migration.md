# Migration Guide

This guide will help you migrate from the original version of Homebox ([https://github.com/hay-kot/homebox](https://github.com/hay-kot/homebox)) to our fork.

## Why Migrate?

- The original version of Homebox has been archived and is no longer maintained.
- Our fork receives regular updates and bug fixes.
- We offer helpful support on our [Discord server](https://discord.homebox.software) and [GitHub](https://git.homebox.software).

## Prerequisites

- A working installation of `hay-kot/homebox`.
- Docker and Docker Compose installed on your server (this guide assumes you are using Docker).

## Migration Steps

### 1. Stop the original version of Homebox

Make sure your `hay-kot/homebox` instance is completely shut down to avoid any conflicts during the migration.

```bash
docker compose down
```

### 2. Backup your data!

Before you start, make sure you have a backup of your data.

> [!WARNING]
> Seriously, this is the most important step. We don't want you to lose anything.

Copy the contents of the `data` folder to a safe location on your server.

### 3. Update the Docker Compose file

Open the `docker-compose.yml` file in your text editor and replace `ghcr.io/hay-kot/homebox:latest` with `ghcr.io/sysadminsmedia/homebox:latest`.
If you are using the rootless image, replace `ghcr.io/hay-kot/homebox:latest-rootless` with `ghcr.io/sysadminsmedia/homebox:latest-rootless`.

### 4. Start the new version of Homebox

Start the new version of Homebox by running the following command:

```bash
docker compose up -d
```

You should now be able to access the web interface!

Check that all your data is present and you are good to go!

## Troubleshooting

If you encounter any issues during the migration process, please ask for help on our [Discord server](https://discord.homebox.software) or [GitHub](https://git.homebox.software).
