# Upgrade

## From v0.17.x to v0.18+

::: danger Breaking Changes
This upgrade process involves some potentially breaking changes, please review this documentation carefully before beginning the upgrade process, and follow it closely during your upgrade.
:::

### Configuration Changes
#### Database Configuration
- `HBOX_STORAGE_SQLITE_URL` has been replaced by `HBOX_DATABASE_SQLITE_PATH`
- `HBOX_DATABASE_DRIVER` has been added to set the database type, valid options are `sqlite3` and `postgres`
- `HBOX_DATABASE_HOST`, `HBOX_DATABASE_PORT`, `HBOX_DATABASE_USERNAME`, `HBOX_DATABASE_DATABASE`, and `HBOX_DATABASE_SSL_MODE` have been added to configure postgres connection options.

::: tip
If you don't have `HBOX_STORAGE_SQLITE_URL` set, you can ignore this change, as the default value for `HBOX_DATABASE_DRIVER` is `sqlite3`, and the default value for `HBOX_DATABASE_SQLITE_PATH` is the same as the old `HBOX_STORAGE_SQLITE_URL` value.
:::