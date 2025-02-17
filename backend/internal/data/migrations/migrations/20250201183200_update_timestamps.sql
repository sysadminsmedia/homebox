-- This migrates some default timestamps to the new ISO8601 format.
UPDATE maintenance_entries
SET
    date = strftime('%Y-%m-%d %H:%M:%S', substr(date, 1, 20))
        || substr(date, 21, 3)
        || ':'
        || substr(date, 24, 2),
    scheduled_date = strftime('%Y-%m-%d %H:%M:%S', substr(date, 1, 20))
        || substr(date, 21, 3)
        || ':'
        || substr(date, 24, 2)
WHERE date OR scheduled_date LIKE '____-__-__ __:__:__ +____ %';

UPDATE items
SET
    sold_time = strftime('%Y-%m-%d %H:%M:%S', substr(sold_time, 1, 20))
        || substr(sold_time, 21, 3)
        || ':'
        || substr(sold_time, 24, 2),
    purchase_time = strftime('%Y-%m-%d %H:%M:%S', substr(purchase_time, 1, 20))
        || substr(purchase_time, 21, 3)
        || ':'
        || substr(purchase_time, 24, 2),
    warranty_expires = strftime('%Y-%m-%d %H:%M:%S', substr(warranty_expires, 1, 20))
        || substr(warranty_expires, 21, 3)
        || ':'
        || substr(warranty_expires, 24, 2)
WHERE sold_time OR purchase_time OR warranty_expires LIKE '____-__-__ __:__:__ +____ %';


-- This migration updates all of the old golang style timestamps to the new ISO8601 format.
UPDATE attachments
SET
    created_at = strftime('%Y-%m-%d %H:%M:%S', substr(created_at, 1, 29))
        || '.' || substr(created_at, 21, 9)
        || substr(created_at, 31, 3)
        || ':'
        || substr(created_at, 34, 2),
    updated_at = strftime('%Y-%m-%d %H:%M:%S', substr(updated_at, 1, 29))
        || '.' || substr(updated_at, 21, 9)
        || substr(updated_at, 31, 3)
        || ':'
        || substr(updated_at, 34, 2)
WHERE (created_at LIKE '____-__-__ __:__:__.% -__:__%' OR created_at LIKE '____-__-__ __:__:__.% +__:__%')
   OR (updated_at LIKE '____-__-__ __:__:__.% -__:__%' OR updated_at LIKE '____-__-__ __:__:__.% +__:__%');

UPDATE auth_tokens
SET
    created_at = strftime('%Y-%m-%d %H:%M:%S', substr(created_at, 1, 29))
        || '.' || substr(created_at, 21, 9)
        || substr(created_at, 31, 3)
        || ':'
        || substr(created_at, 34, 2),
    updated_at = strftime('%Y-%m-%d %H:%M:%S', substr(updated_at, 1, 29))
        || '.' || substr(updated_at, 21, 9)
        || substr(updated_at, 31, 3)
        || ':'
        || substr(updated_at, 34, 2),
    expires_at = strftime('%Y-%m-%d %H:%M:%S', substr(expires_at, 1, 29))
        || '.' || substr(expires_at, 21, 9)
        || substr(expires_at, 31, 3)
        || ':'
        || substr(expires_at, 34, 2)
WHERE (created_at LIKE '____-__-__ __:__:__.% -__:__%' OR created_at LIKE '____-__-__ __:__:__.% +__:__%')
   OR (updated_at LIKE '____-__-__ __:__:__.% -__:__%' OR updated_at LIKE '____-__-__ __:__:__.% +__:__%')
   OR (expires_at LIKE '____-__-__ __:__:__.% -__:__%' OR expires_at LIKE '____-__-__ __:__:__.% +__:__%');

UPDATE documents
SET
    created_at = strftime('%Y-%m-%d %H:%M:%S', substr(created_at, 1, 29))
        || '.' || substr(created_at, 21, 9)
        || substr(created_at, 31, 3)
        || ':'
        || substr(created_at, 34, 2),
    updated_at = strftime('%Y-%m-%d %H:%M:%S', substr(updated_at, 1, 29))
        || '.' || substr(updated_at, 21, 9)
        || substr(updated_at, 31, 3)
        || ':'
        || substr(updated_at, 34, 2)
WHERE (created_at LIKE '____-__-__ __:__:__.% -__:__%' OR created_at LIKE '____-__-__ __:__:__.% +__:__%')
   OR (updated_at LIKE '____-__-__ __:__:__.% -__:__%' OR updated_at LIKE '____-__-__ __:__:__.% +__:__%');

UPDATE group_invitation_tokens
SET
    created_at = strftime('%Y-%m-%d %H:%M:%S', substr(created_at, 1, 29))
        || '.' || substr(created_at, 21, 9)
        || substr(created_at, 31, 3)
        || ':'
        || substr(created_at, 34, 2),
    updated_at = strftime('%Y-%m-%d %H:%M:%S', substr(updated_at, 1, 29))
        || '.' || substr(updated_at, 21, 9)
        || substr(updated_at, 31, 3)
        || ':'
        || substr(updated_at, 34, 2),
    expires_at = strftime('%Y-%m-%d %H:%M:%S', substr(expires_at, 1, 29))
        || '.' || substr(expires_at, 21, 9)
        || substr(expires_at, 31, 3)
        || ':'
        || substr(expires_at, 34, 2)
WHERE (created_at LIKE '____-__-__ __:__:__.% -__:__%' OR created_at LIKE '____-__-__ __:__:__.% +__:__%')
   OR (updated_at LIKE '____-__-__ __:__:__.% -__:__%' OR updated_at LIKE '____-__-__ __:__:__.% +__:__%')
   OR (expires_at LIKE '____-__-__ __:__:__.% -__:__%' OR expires_at LIKE '____-__-__ __:__:__.% +__:__%');

UPDATE groups
SET
    created_at = strftime('%Y-%m-%d %H:%M:%S', substr(created_at, 1, 29))
        || '.' || substr(created_at, 21, 9)
        || substr(created_at, 31, 3)
        || ':'
        || substr(created_at, 34, 2),
    updated_at = strftime('%Y-%m-%d %H:%M:%S', substr(updated_at, 1, 29))
        || '.' || substr(updated_at, 21, 9)
        || substr(updated_at, 31, 3)
        || ':'
        || substr(updated_at, 34, 2)
WHERE (created_at LIKE '____-__-__ __:__:__.% -__:__%' OR created_at LIKE '____-__-__ __:__:__.% +__:__%')
   OR (updated_at LIKE '____-__-__ __:__:__.% -__:__%' OR updated_at LIKE '____-__-__ __:__:__.% +__:__%');

UPDATE item_fields
SET
    created_at = strftime('%Y-%m-%d %H:%M:%S', substr(created_at, 1, 29))
        || '.' || substr(created_at, 21, 9)
        || substr(created_at, 31, 3)
        || ':'
        || substr(created_at, 34, 2),
    updated_at = strftime('%Y-%m-%d %H:%M:%S', substr(updated_at, 1, 29))
        || '.' || substr(updated_at, 21, 9)
        || substr(updated_at, 31, 3)
        || ':'
        || substr(updated_at, 34, 2),
    time_value = strftime('%Y-%m-%d %H:%M:%S', substr(time_value, 1, 29))
        || '.' || substr(time_value, 21, 9)
        || substr(time_value, 31, 3)
        || ':'
        || substr(time_value, 34, 2)
WHERE (created_at LIKE '____-__-__ __:__:__.% -__:__%' OR created_at LIKE '____-__-__ __:__:__.% +__:__%')
   OR (updated_at LIKE '____-__-__ __:__:__.% -__:__%' OR updated_at LIKE '____-__-__ __:__:__.% +__:__%')
   OR (time_value LIKE '____-__-__ __:__:__.% -__:__%' OR time_value LIKE '____-__-__ __:__:__.% +__:__%');

UPDATE items
SET
    created_at = strftime('%Y-%m-%d %H:%M:%S', substr(created_at, 1, 29))
        || '.' || substr(created_at, 21, 9)
        || substr(created_at, 31, 3)
        || ':'
        || substr(created_at, 34, 2),
    updated_at = strftime('%Y-%m-%d %H:%M:%S', substr(updated_at, 1, 29))
        || '.' || substr(updated_at, 21, 9)
        || substr(updated_at, 31, 3)
        || ':'
        || substr(updated_at, 34, 2),
    sold_time = strftime('%Y-%m-%d %H:%M:%S', substr(created_at, 1, 29))
        || '.' || substr(created_at, 21, 9)
        || substr(created_at, 31, 3)
        || ':'
        || substr(created_at, 34, 2),
    purchase_time = strftime('%Y-%m-%d %H:%M:%S', substr(updated_at, 1, 29))
        || '.' || substr(updated_at, 21, 9)
        || substr(updated_at, 31, 3)
        || ':'
        || substr(updated_at, 34, 2),
    warranty_expires = strftime('%Y-%m-%d %H:%M:%S', substr(warranty_expires, 1, 29))
        || '.' || substr(warranty_expires, 21, 9)
        || substr(warranty_expires, 31, 3)
        || ':'
        || substr(warranty_expires, 34, 2)
WHERE (created_at LIKE '____-__-__ __:__:__.% -__:__%' OR created_at LIKE '____-__-__ __:__:__.% +__:__%')
   OR (updated_at LIKE '____-__-__ __:__:__.% -__:__%' OR updated_at LIKE '____-__-__ __:__:__.% +__:__%')
   OR (sold_time LIKE '____-__-__ __:__:__.% -__:__%' OR sold_time LIKE '____-__-__ __:__:__.% +__:__%')
   OR (purchase_time LIKE '____-__-__ __:__:__.% -__:__%' OR purchase_time LIKE '____-__-__ __:__:__.% +__:__%')
   OR (warranty_expires LIKE '____-__-__ __:__:__.% -__:__%' OR warranty_expires LIKE '____-__-__ __:__:__.% +__:__%');

UPDATE labels
SET
    created_at = strftime('%Y-%m-%d %H:%M:%S', substr(created_at, 1, 29))
        || '.' || substr(created_at, 21, 9)
        || substr(created_at, 31, 3)
        || ':'
        || substr(created_at, 34, 2),
    updated_at = strftime('%Y-%m-%d %H:%M:%S', substr(updated_at, 1, 29))
        || '.' || substr(updated_at, 21, 9)
        || substr(updated_at, 31, 3)
        || ':'
        || substr(updated_at, 34, 2)
WHERE (created_at LIKE '____-__-__ __:__:__.% -__:__%' OR created_at LIKE '____-__-__ __:__:__.% +__:__%')
   OR (updated_at LIKE '____-__-__ __:__:__.% -__:__%' OR updated_at LIKE '____-__-__ __:__:__.% +__:__%');

UPDATE locations
SET
    created_at = strftime('%Y-%m-%d %H:%M:%S', substr(created_at, 1, 29))
        || '.' || substr(created_at, 21, 9)
        || substr(created_at, 31, 3)
        || ':'
        || substr(created_at, 34, 2),
    updated_at = strftime('%Y-%m-%d %H:%M:%S', substr(updated_at, 1, 29))
        || '.' || substr(updated_at, 21, 9)
        || substr(updated_at, 31, 3)
        || ':'
        || substr(updated_at, 34, 2)
WHERE (created_at LIKE '____-__-__ __:__:__.% -__:__%' OR created_at LIKE '____-__-__ __:__:__.% +__:__%')
   OR (updated_at LIKE '____-__-__ __:__:__.% -__:__%' OR updated_at LIKE '____-__-__ __:__:__.% +__:__%');

UPDATE maintenance_entries
SET
    created_at = strftime('%Y-%m-%d %H:%M:%S', substr(created_at, 1, 29))
        || '.' || substr(created_at, 21, 9)
        || substr(created_at, 31, 3)
        || ':'
        || substr(created_at, 34, 2),
    updated_at = strftime('%Y-%m-%d %H:%M:%S', substr(updated_at, 1, 29))
        || '.' || substr(updated_at, 21, 9)
        || substr(updated_at, 31, 3)
        || ':'
        || substr(updated_at, 34, 2),
    date = strftime('%Y-%m-%d %H:%M:%S', substr(date, 1, 29))
        || '.' || substr(date, 21, 9)
        || substr(date, 31, 3)
        || ':'
        || substr(date, 34, 2),
    scheduled_date = strftime('%Y-%m-%d %H:%M:%S', substr(scheduled_date, 1, 29))
        || '.' || substr(scheduled_date, 21, 9)
        || substr(scheduled_date, 31, 3)
        || ':'
        || substr(scheduled_date, 34, 2)
WHERE (created_at LIKE '____-__-__ __:__:__.% -__:__%' OR created_at LIKE '____-__-__ __:__:__.% +__:__%')
   OR (updated_at LIKE '____-__-__ __:__:__.% -__:__%' OR updated_at LIKE '____-__-__ __:__:__.% +__:__%');

UPDATE notifiers
SET
    created_at = strftime('%Y-%m-%d %H:%M:%S', substr(created_at, 1, 29))
        || '.' || substr(created_at, 21, 9)
        || substr(created_at, 31, 3)
        || ':'
        || substr(created_at, 34, 2),
    updated_at = strftime('%Y-%m-%d %H:%M:%S', substr(updated_at, 1, 29))
        || '.' || substr(updated_at, 21, 9)
        || substr(updated_at, 31, 3)
        || ':'
        || substr(updated_at, 34, 2)
WHERE (created_at LIKE '____-__-__ __:__:__.% -__:__%' OR created_at LIKE '____-__-__ __:__:__.% +__:__%')
   OR (updated_at LIKE '____-__-__ __:__:__.% -__:__%' OR updated_at LIKE '____-__-__ __:__:__.% +__:__%');

UPDATE users
SET
    created_at = strftime('%Y-%m-%d %H:%M:%S', substr(created_at, 1, 29))
        || '.' || substr(created_at, 21, 9)
        || substr(created_at, 31, 3)
        || ':'
        || substr(created_at, 34, 2),
    updated_at = strftime('%Y-%m-%d %H:%M:%S', substr(updated_at, 1, 29))
        || '.' || substr(updated_at, 21, 9)
        || substr(updated_at, 31, 3)
        || ':'
        || substr(updated_at, 34, 2),
    activated_on = strftime('%Y-%m-%d %H:%M:%S', substr(activated_on, 1, 29))
        || '.' || substr(activated_on, 21, 9)
        || substr(activated_on, 31, 3)
        || ':'
        || substr(activated_on, 34, 2)
WHERE (created_at LIKE '____-__-__ __:__:__.% -__:__%' OR created_at LIKE '____-__-__ __:__:__.% +__:__%')
   OR (updated_at LIKE '____-__-__ __:__:__.% -__:__%' OR updated_at LIKE '____-__-__ __:__:__.% +__:__%')
   OR (activated_on LIKE '____-__-__ __:__:__.% -__:__%' OR activated_on LIKE '____-__-__ __:__:__.% +__:__%');