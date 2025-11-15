#!/usr/bin/env python
import os

# Extract fields with
""" WITH tables AS (
  SELECT name AS table_name
  FROM sqlite_master
  WHERE type = 'table'
    AND name NOT LIKE 'sqlite_%'
)

SELECT
  '["' || t.table_name || '", "' || c.name || '"],' AS table_column
FROM tables t
JOIN pragma_table_info(t.table_name) c
WHERE c.name like'%date%'; """

fields = [["auth_tokens", "created_at"],
          ["auth_tokens", "updated_at"],
          ["auth_tokens", "expires_at"],
          ["groups", "created_at"],
          ["groups", "updated_at"],
          ["group_invitation_tokens", "created_at"],
          ["group_invitation_tokens", "updated_at"],
          ["group_invitation_tokens", "expires_at"],
          ["item_fields", "created_at"],
          ["item_fields", "updated_at"],
          ["item_fields", "time_value"],
          ["labels", "created_at"],
          ["labels", "updated_at"],
          ["locations", "created_at"],
          ["locations", "updated_at"],
          ["maintenance_entries", "created_at"],
          ["maintenance_entries", "updated_at"],
          ["maintenance_entries", "date"],
          ["maintenance_entries", "scheduled_date"],
          ["notifiers", "created_at"],
          ["notifiers", "updated_at"],
          ["users", "created_at"],
          ["users", "updated_at"],
          ["users", "activated_on"],
          ["items", "created_at"],
          ["items", "updated_at"],
          ["items", "warranty_expires"],
          ["items", "purchase_time"],
          ["items", "sold_time"],
          ["attachments", "created_at"],
          ["attachments", "updated_at"]]


def generate_migration(table_name, field_name):
    return f"""update {table_name} set {field_name} = substr({field_name},1, instr({field_name}, ' +')-1) || substr({field_name}, instr({field_name}, ' +')+1,3) || ':' || substr({field_name}, instr({field_name}, ' +')+4,2) where {field_name} like '% +%';\n""" + \
           f"""update {table_name} set {field_name} = substr({field_name},1, instr({field_name}, ' -')-1) || substr({field_name}, instr({field_name}, ' -')+1,3) || ':' || substr({field_name}, instr({field_name}, ' -')+4,2) where {field_name} like '% -%';"""


print("-- +goose Up")
print(f"-- GENERATED with {os.path.basename(__file__)}")
for table, column in fields:
    print(f"-- Migrating {table}/{column}")
    print(generate_migration(table, column))
    print()
