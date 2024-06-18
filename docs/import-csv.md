# CSV Imports

## Quick Start

Using the CSV import is the recommended way for adding items to the database. It is always going to be the fastest way to import any large number of items and provides the most flexibility when it comes to adding items.

**Current Limitations**

 - Imports only support importing items, locations, and labels
 - Imports and Exports do not support attachments. Attachments must be uploaded after import
 - CSV Exports do not support nested path exports (e.g. `Home / Office / Desk`) and will only export the Items direct parent, (though imports _do_ support nested paths)
 - Cannot specify item-to-item relationships (e.g. `Item A` is a child of `Item B`)

!!! tip "File Formats"
    The CSV import supports both CSV and TSV files. The only difference is the delimiter used. CSV files use a comma `,` as the delimiter and TSV files use a tab `\t` as the delimiter. The file extension does not matter.

## CSV Reference

Below are the supported columns. They are case-sensitive, can be in any ordered or can be omitted unless otherwise specified.

### Special Syntax Columns

`HB.import_ref`

:   Import Refs are unique strings that can be used to deduplicate imports. Before an item is imported, we check the database for a matching ref. If the ref exists, we skip the creation of that item.

    * String Type
    * Max 100 Characters

    Import Refs are used to de-duplicate imports. It is HIGHLY recommended that you use them to manage your items if you intend to manage your inventory via CSV import/export. If you do not use import refs, you will end up with duplicate items in your database on subsequent imports.

    !!! tip

        Specifying import refs also allows you to update existing items via the CSV import. If you specify an import ref that already exists in the database, we will update the existing item instead of creating a new one.

`HB.location`

:   This is the location of the item that will be created. These are de-duplicated and won't create another instance when reused.

    * Supports Path Separators for nested locations (e.g. `Home / Office / Desk`)

`HB.labels`

:   List of labels to apply to the item separated by a `;` can be existing or new labels.

`HB.field.{field_name}` (e.g. `HB.field.Serial Number`)

:  This is a special column that allows you to add custom fields to the item. The column name must start with `HB.field.` followed by the name of the field. The value of the column will be the value of the field.

    - If the cell value is empty, it will be ignored.

### Standard Columns

| Column               | Type          | Description                                   |
|----------------------|---------------|-----------------------------------------------|
| HB.quantity          | Integer       | The quantity of items to create               |
| HB.name              | String        | Name of the item                              |
| HB.asset_id          | AssetID       | Asset ID for the item                         |
| HB.description       | String        | Description of the item                       |
| HB.insured           | Boolean       | Whether or not the item is insured            |
| HB.serial_number     | String        | Serial number of the item                     |
| HB.model_number      | String        | Model of the item                             |
| HB.manufacturer      | String        | Manufacturer of the item                      |
| HB.notes             | String (1000) | General notes about the product               |
| HB.purchase_from     | String        | Name of the place the item was purchased from |
| HB.purchase_price    | Float64       |                                               |
| HB.purchase_time     | Date          | Date the item was purchased                   |
| HB.lifetime_warranty | Boolean       | true or false - case insensitive              |
| HB.warranty_expires  | Date          | Date in the format                            |
| HB.warranty_details  | String        | Details about the warranty                    |
| HB.sold_to           | String        | Name of the person the item was sold to       |
| HB.sold_time         | Date          | Date the item was sold                        |
| HB.sold_price        | Float64       |                                               |
| HB.sold_notes        | String (1000) |                                               |

**Type Key**

| Type    | Format                                              |
|---------|-----------------------------------------------------|
| String  | Max 255 Characters unless otherwise specified       |
| Date    | YYYY-MM-DD                                          |
| Boolean | true or false, yes or no, 1 or 0 - case insensitive |
| AssetID | 000-000                                             |
