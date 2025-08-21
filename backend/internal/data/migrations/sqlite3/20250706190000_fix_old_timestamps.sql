-- +goose Up
-- GENERATED with 20250706190000_generate_migration.py
-- Migrating auth_tokens/created_at
update auth_tokens set created_at = substr(created_at,1, instr(created_at, ' +')-1) || substr(created_at, instr(created_at, ' +')+1,3) || ':' || substr(created_at, instr(created_at, ' +')+4,2) where created_at like '% +%';
update auth_tokens set created_at = substr(created_at,1, instr(created_at, ' -')-1) || substr(created_at, instr(created_at, ' -')+1,3) || ':' || substr(created_at, instr(created_at, ' -')+4,2) where created_at like '% -%';

-- Migrating auth_tokens/updated_at
update auth_tokens set updated_at = substr(updated_at,1, instr(updated_at, ' +')-1) || substr(updated_at, instr(updated_at, ' +')+1,3) || ':' || substr(updated_at, instr(updated_at, ' +')+4,2) where updated_at like '% +%';
update auth_tokens set updated_at = substr(updated_at,1, instr(updated_at, ' -')-1) || substr(updated_at, instr(updated_at, ' -')+1,3) || ':' || substr(updated_at, instr(updated_at, ' -')+4,2) where updated_at like '% -%';

-- Migrating auth_tokens/expires_at
update auth_tokens set expires_at = substr(expires_at,1, instr(expires_at, ' +')-1) || substr(expires_at, instr(expires_at, ' +')+1,3) || ':' || substr(expires_at, instr(expires_at, ' +')+4,2) where expires_at like '% +%';
update auth_tokens set expires_at = substr(expires_at,1, instr(expires_at, ' -')-1) || substr(expires_at, instr(expires_at, ' -')+1,3) || ':' || substr(expires_at, instr(expires_at, ' -')+4,2) where expires_at like '% -%';

-- Migrating groups/created_at
update groups set created_at = substr(created_at,1, instr(created_at, ' +')-1) || substr(created_at, instr(created_at, ' +')+1,3) || ':' || substr(created_at, instr(created_at, ' +')+4,2) where created_at like '% +%';
update groups set created_at = substr(created_at,1, instr(created_at, ' -')-1) || substr(created_at, instr(created_at, ' -')+1,3) || ':' || substr(created_at, instr(created_at, ' -')+4,2) where created_at like '% -%';

-- Migrating groups/updated_at
update groups set updated_at = substr(updated_at,1, instr(updated_at, ' +')-1) || substr(updated_at, instr(updated_at, ' +')+1,3) || ':' || substr(updated_at, instr(updated_at, ' +')+4,2) where updated_at like '% +%';
update groups set updated_at = substr(updated_at,1, instr(updated_at, ' -')-1) || substr(updated_at, instr(updated_at, ' -')+1,3) || ':' || substr(updated_at, instr(updated_at, ' -')+4,2) where updated_at like '% -%';

-- Migrating group_invitation_tokens/created_at
update group_invitation_tokens set created_at = substr(created_at,1, instr(created_at, ' +')-1) || substr(created_at, instr(created_at, ' +')+1,3) || ':' || substr(created_at, instr(created_at, ' +')+4,2) where created_at like '% +%';
update group_invitation_tokens set created_at = substr(created_at,1, instr(created_at, ' -')-1) || substr(created_at, instr(created_at, ' -')+1,3) || ':' || substr(created_at, instr(created_at, ' -')+4,2) where created_at like '% -%';

-- Migrating group_invitation_tokens/updated_at
update group_invitation_tokens set updated_at = substr(updated_at,1, instr(updated_at, ' +')-1) || substr(updated_at, instr(updated_at, ' +')+1,3) || ':' || substr(updated_at, instr(updated_at, ' +')+4,2) where updated_at like '% +%';
update group_invitation_tokens set updated_at = substr(updated_at,1, instr(updated_at, ' -')-1) || substr(updated_at, instr(updated_at, ' -')+1,3) || ':' || substr(updated_at, instr(updated_at, ' -')+4,2) where updated_at like '% -%';

-- Migrating group_invitation_tokens/expires_at
update group_invitation_tokens set expires_at = substr(expires_at,1, instr(expires_at, ' +')-1) || substr(expires_at, instr(expires_at, ' +')+1,3) || ':' || substr(expires_at, instr(expires_at, ' +')+4,2) where expires_at like '% +%';
update group_invitation_tokens set expires_at = substr(expires_at,1, instr(expires_at, ' -')-1) || substr(expires_at, instr(expires_at, ' -')+1,3) || ':' || substr(expires_at, instr(expires_at, ' -')+4,2) where expires_at like '% -%';

-- Migrating item_fields/created_at
update item_fields set created_at = substr(created_at,1, instr(created_at, ' +')-1) || substr(created_at, instr(created_at, ' +')+1,3) || ':' || substr(created_at, instr(created_at, ' +')+4,2) where created_at like '% +%';
update item_fields set created_at = substr(created_at,1, instr(created_at, ' -')-1) || substr(created_at, instr(created_at, ' -')+1,3) || ':' || substr(created_at, instr(created_at, ' -')+4,2) where created_at like '% -%';

-- Migrating item_fields/updated_at
update item_fields set updated_at = substr(updated_at,1, instr(updated_at, ' +')-1) || substr(updated_at, instr(updated_at, ' +')+1,3) || ':' || substr(updated_at, instr(updated_at, ' +')+4,2) where updated_at like '% +%';
update item_fields set updated_at = substr(updated_at,1, instr(updated_at, ' -')-1) || substr(updated_at, instr(updated_at, ' -')+1,3) || ':' || substr(updated_at, instr(updated_at, ' -')+4,2) where updated_at like '% -%';

-- Migrating item_fields/time_value
update item_fields set time_value = substr(time_value,1, instr(time_value, ' +')-1) || substr(time_value, instr(time_value, ' +')+1,3) || ':' || substr(time_value, instr(time_value, ' +')+4,2) where time_value like '% +%';
update item_fields set time_value = substr(time_value,1, instr(time_value, ' -')-1) || substr(time_value, instr(time_value, ' -')+1,3) || ':' || substr(time_value, instr(time_value, ' -')+4,2) where time_value like '% -%';

-- Migrating labels/created_at
update labels set created_at = substr(created_at,1, instr(created_at, ' +')-1) || substr(created_at, instr(created_at, ' +')+1,3) || ':' || substr(created_at, instr(created_at, ' +')+4,2) where created_at like '% +%';
update labels set created_at = substr(created_at,1, instr(created_at, ' -')-1) || substr(created_at, instr(created_at, ' -')+1,3) || ':' || substr(created_at, instr(created_at, ' -')+4,2) where created_at like '% -%';

-- Migrating labels/updated_at
update labels set updated_at = substr(updated_at,1, instr(updated_at, ' +')-1) || substr(updated_at, instr(updated_at, ' +')+1,3) || ':' || substr(updated_at, instr(updated_at, ' +')+4,2) where updated_at like '% +%';
update labels set updated_at = substr(updated_at,1, instr(updated_at, ' -')-1) || substr(updated_at, instr(updated_at, ' -')+1,3) || ':' || substr(updated_at, instr(updated_at, ' -')+4,2) where updated_at like '% -%';

-- Migrating locations/created_at
update locations set created_at = substr(created_at,1, instr(created_at, ' +')-1) || substr(created_at, instr(created_at, ' +')+1,3) || ':' || substr(created_at, instr(created_at, ' +')+4,2) where created_at like '% +%';
update locations set created_at = substr(created_at,1, instr(created_at, ' -')-1) || substr(created_at, instr(created_at, ' -')+1,3) || ':' || substr(created_at, instr(created_at, ' -')+4,2) where created_at like '% -%';

-- Migrating locations/updated_at
update locations set updated_at = substr(updated_at,1, instr(updated_at, ' +')-1) || substr(updated_at, instr(updated_at, ' +')+1,3) || ':' || substr(updated_at, instr(updated_at, ' +')+4,2) where updated_at like '% +%';
update locations set updated_at = substr(updated_at,1, instr(updated_at, ' -')-1) || substr(updated_at, instr(updated_at, ' -')+1,3) || ':' || substr(updated_at, instr(updated_at, ' -')+4,2) where updated_at like '% -%';

-- Migrating maintenance_entries/created_at
update maintenance_entries set created_at = substr(created_at,1, instr(created_at, ' +')-1) || substr(created_at, instr(created_at, ' +')+1,3) || ':' || substr(created_at, instr(created_at, ' +')+4,2) where created_at like '% +%';
update maintenance_entries set created_at = substr(created_at,1, instr(created_at, ' -')-1) || substr(created_at, instr(created_at, ' -')+1,3) || ':' || substr(created_at, instr(created_at, ' -')+4,2) where created_at like '% -%';

-- Migrating maintenance_entries/updated_at
update maintenance_entries set updated_at = substr(updated_at,1, instr(updated_at, ' +')-1) || substr(updated_at, instr(updated_at, ' +')+1,3) || ':' || substr(updated_at, instr(updated_at, ' +')+4,2) where updated_at like '% +%';
update maintenance_entries set updated_at = substr(updated_at,1, instr(updated_at, ' -')-1) || substr(updated_at, instr(updated_at, ' -')+1,3) || ':' || substr(updated_at, instr(updated_at, ' -')+4,2) where updated_at like '% -%';

-- Migrating maintenance_entries/date
update maintenance_entries set date = substr(date,1, instr(date, ' +')-1) || substr(date, instr(date, ' +')+1,3) || ':' || substr(date, instr(date, ' +')+4,2) where date like '% +%';
update maintenance_entries set date = substr(date,1, instr(date, ' -')-1) || substr(date, instr(date, ' -')+1,3) || ':' || substr(date, instr(date, ' -')+4,2) where date like '% -%';

-- Migrating maintenance_entries/scheduled_date
update maintenance_entries set scheduled_date = substr(scheduled_date,1, instr(scheduled_date, ' +')-1) || substr(scheduled_date, instr(scheduled_date, ' +')+1,3) || ':' || substr(scheduled_date, instr(scheduled_date, ' +')+4,2) where scheduled_date like '% +%';
update maintenance_entries set scheduled_date = substr(scheduled_date,1, instr(scheduled_date, ' -')-1) || substr(scheduled_date, instr(scheduled_date, ' -')+1,3) || ':' || substr(scheduled_date, instr(scheduled_date, ' -')+4,2) where scheduled_date like '% -%';

-- Migrating notifiers/created_at
update notifiers set created_at = substr(created_at,1, instr(created_at, ' +')-1) || substr(created_at, instr(created_at, ' +')+1,3) || ':' || substr(created_at, instr(created_at, ' +')+4,2) where created_at like '% +%';
update notifiers set created_at = substr(created_at,1, instr(created_at, ' -')-1) || substr(created_at, instr(created_at, ' -')+1,3) || ':' || substr(created_at, instr(created_at, ' -')+4,2) where created_at like '% -%';

-- Migrating notifiers/updated_at
update notifiers set updated_at = substr(updated_at,1, instr(updated_at, ' +')-1) || substr(updated_at, instr(updated_at, ' +')+1,3) || ':' || substr(updated_at, instr(updated_at, ' +')+4,2) where updated_at like '% +%';
update notifiers set updated_at = substr(updated_at,1, instr(updated_at, ' -')-1) || substr(updated_at, instr(updated_at, ' -')+1,3) || ':' || substr(updated_at, instr(updated_at, ' -')+4,2) where updated_at like '% -%';

-- Migrating users/created_at
update users set created_at = substr(created_at,1, instr(created_at, ' +')-1) || substr(created_at, instr(created_at, ' +')+1,3) || ':' || substr(created_at, instr(created_at, ' +')+4,2) where created_at like '% +%';
update users set created_at = substr(created_at,1, instr(created_at, ' -')-1) || substr(created_at, instr(created_at, ' -')+1,3) || ':' || substr(created_at, instr(created_at, ' -')+4,2) where created_at like '% -%';

-- Migrating users/updated_at
update users set updated_at = substr(updated_at,1, instr(updated_at, ' +')-1) || substr(updated_at, instr(updated_at, ' +')+1,3) || ':' || substr(updated_at, instr(updated_at, ' +')+4,2) where updated_at like '% +%';
update users set updated_at = substr(updated_at,1, instr(updated_at, ' -')-1) || substr(updated_at, instr(updated_at, ' -')+1,3) || ':' || substr(updated_at, instr(updated_at, ' -')+4,2) where updated_at like '% -%';

-- Migrating users/activated_on
update users set activated_on = substr(activated_on,1, instr(activated_on, ' +')-1) || substr(activated_on, instr(activated_on, ' +')+1,3) || ':' || substr(activated_on, instr(activated_on, ' +')+4,2) where activated_on like '% +%';
update users set activated_on = substr(activated_on,1, instr(activated_on, ' -')-1) || substr(activated_on, instr(activated_on, ' -')+1,3) || ':' || substr(activated_on, instr(activated_on, ' -')+4,2) where activated_on like '% -%';

-- Migrating items/created_at
update items set created_at = substr(created_at,1, instr(created_at, ' +')-1) || substr(created_at, instr(created_at, ' +')+1,3) || ':' || substr(created_at, instr(created_at, ' +')+4,2) where created_at like '% +%';
update items set created_at = substr(created_at,1, instr(created_at, ' -')-1) || substr(created_at, instr(created_at, ' -')+1,3) || ':' || substr(created_at, instr(created_at, ' -')+4,2) where created_at like '% -%';

-- Migrating items/updated_at
update items set updated_at = substr(updated_at,1, instr(updated_at, ' +')-1) || substr(updated_at, instr(updated_at, ' +')+1,3) || ':' || substr(updated_at, instr(updated_at, ' +')+4,2) where updated_at like '% +%';
update items set updated_at = substr(updated_at,1, instr(updated_at, ' -')-1) || substr(updated_at, instr(updated_at, ' -')+1,3) || ':' || substr(updated_at, instr(updated_at, ' -')+4,2) where updated_at like '% -%';

-- Migrating items/warranty_expires
update items set warranty_expires = substr(warranty_expires,1, instr(warranty_expires, ' +')-1) || substr(warranty_expires, instr(warranty_expires, ' +')+1,3) || ':' || substr(warranty_expires, instr(warranty_expires, ' +')+4,2) where warranty_expires like '% +%';
update items set warranty_expires = substr(warranty_expires,1, instr(warranty_expires, ' -')-1) || substr(warranty_expires, instr(warranty_expires, ' -')+1,3) || ':' || substr(warranty_expires, instr(warranty_expires, ' -')+4,2) where warranty_expires like '% -%';

-- Migrating items/purchase_time
update items set purchase_time = substr(purchase_time,1, instr(purchase_time, ' +')-1) || substr(purchase_time, instr(purchase_time, ' +')+1,3) || ':' || substr(purchase_time, instr(purchase_time, ' +')+4,2) where purchase_time like '% +%';
update items set purchase_time = substr(purchase_time,1, instr(purchase_time, ' -')-1) || substr(purchase_time, instr(purchase_time, ' -')+1,3) || ':' || substr(purchase_time, instr(purchase_time, ' -')+4,2) where purchase_time like '% -%';

-- Migrating items/sold_time
update items set sold_time = substr(sold_time,1, instr(sold_time, ' +')-1) || substr(sold_time, instr(sold_time, ' +')+1,3) || ':' || substr(sold_time, instr(sold_time, ' +')+4,2) where sold_time like '% +%';
update items set sold_time = substr(sold_time,1, instr(sold_time, ' -')-1) || substr(sold_time, instr(sold_time, ' -')+1,3) || ':' || substr(sold_time, instr(sold_time, ' -')+4,2) where sold_time like '% -%';

-- Migrating attachments/created_at
update attachments set created_at = substr(created_at,1, instr(created_at, ' +')-1) || substr(created_at, instr(created_at, ' +')+1,3) || ':' || substr(created_at, instr(created_at, ' +')+4,2) where created_at like '% +%';
update attachments set created_at = substr(created_at,1, instr(created_at, ' -')-1) || substr(created_at, instr(created_at, ' -')+1,3) || ':' || substr(created_at, instr(created_at, ' -')+4,2) where created_at like '% -%';

-- Migrating attachments/updated_at
update attachments set updated_at = substr(updated_at,1, instr(updated_at, ' +')-1) || substr(updated_at, instr(updated_at, ' +')+1,3) || ':' || substr(updated_at, instr(updated_at, ' +')+4,2) where updated_at like '% +%';
update attachments set updated_at = substr(updated_at,1, instr(updated_at, ' -')-1) || substr(updated_at, instr(updated_at, ' -')+1,3) || ':' || substr(updated_at, instr(updated_at, ' -')+4,2) where updated_at like '% -%';

