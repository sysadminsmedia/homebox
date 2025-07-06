# Migration Guide

This guide will help you migrate from the original version of Homebox ([https://github.com/hay-kot/homebox](https://github.com/hay-kot/homebox)) to our actively maintained fork.

## Why Migrate?

Migrating to our fork ensures you benefit from:

- **Active Development**: The original Homebox has been archived and is no longer maintained, while our fork receives regular updates and bug fixes.
- **Community Support**: Get help and advice on our [Discord server](https://discord.homebox.software) or [GitHub](https://git.homebox.software).
- **Improved Features**: Enjoy enhancements and optimizations that make Homebox even better.

## Prerequisites

Before starting the migration, ensure you have:

- A working installation of `hay-kot/homebox`.
- Docker and Docker Compose installed on your server (this guide assumes Docker is being used).

## Migration Steps

### 1. Stop the Original Homebox Instance

To avoid conflicts during migration, shut down your existing `hay-kot/homebox` instance:

```bash
docker compose down
```

### 2. Backup Your Data

**This step is critical!** Before proceeding, create a backup of your data to ensure nothing is lost.

> [!WARNING]  
> **Don't skip this step!** Backing up your data is the most important part of the migration process.

Locate the `data` folder used by your current Homebox installation and copy its contents to a safe location on your server. If you are using a data volume, follow the [instructions on Docker's website](https://docs.docker.com/engine/storage/volumes/#back-up-restore-or-migrate-data-volumes).
### 3. Update the Docker Compose File

Modify your `docker-compose.yml` file to point to the new Homebox fork:

- Replace:  
  `ghcr.io/hay-kot/homebox:latest`  
  **With:**  
  `ghcr.io/sysadminsmedia/homebox:latest`

- If you're using the rootless image, replace:  
  `ghcr.io/hay-kot/homebox:latest-rootless`  
  **With:**  
  `ghcr.io/sysadminsmedia/homebox:latest-rootless`

- Update the environment variable:  
  - If you're using `HBOX_STORAGE_SQLITE_URL`, change it to `HBOX_DATABASE_SQLITE_PATH`.
  - If you're using `HBOX_WEB_READ_TIMEOUT`, `HBOX_WEB_WRITE_TIMEOUT`, or `HBOX_IDLE_TIMEOUT`, add an `s` for seconds or `m` for minutes to the end of the integers.

### 4. Start the New Homebox Instance

Launch the new version of Homebox with the following command:

```bash
docker compose up -d
```

Once the service is running, access the web interface and verify:

- All your data has been successfully migrated.
- The service is functioning as expected.

## Troubleshooting

If you run into any issues during the migration process, don't hesitate to reach out for help:

- **Discord**: [Join our community](https://discord.homebox.software) for real-time support.  
- **GitHub**: [Open an issue or discussion](https://git.homebox.software) for technical assistance.
