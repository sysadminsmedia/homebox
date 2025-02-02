-- This migrates some special fields to the new ISO8601 format.
UPDATE maintenance_entries
SET
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
WHERE date OR scheduled_date LIKE '____-__-__ __:__:__._________ %';

UPDATE items
SET
    sold_time = strftime('%Y-%m-%d %H:%M:%S', substr(created_at, 1, 29))
        || '.' || substr(created_at, 21, 9)
        || substr(created_at, 31, 3)
        || ':'
        || substr(created_at, 34, 2),
    purchase_time = strftime('%Y-%m-%d %H:%M:%S', substr(updated_at, 1, 29))
        || '.' || substr(updated_at, 21, 9)
        || substr(updated_at, 31, 3)
        || ':'
        || substr(updated_at, 34, 2)
WHERE sold_time OR purchase_time LIKE '____-__-__ __:__:__._________ %';


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
WHERE created_at OR updated_at LIKE '____-__-__ __:__:__._________ %';

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
        || substr(updated_at, 34, 2)
WHERE created_at OR updated_at LIKE '____-__-__ __:__:__._________ %';

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
WHERE created_at OR updated_at LIKE '____-__-__ __:__:__._________ %';

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
        || substr(updated_at, 34, 2)
WHERE created_at OR updated_at LIKE '____-__-__ __:__:__._________ %';

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
WHERE created_at OR updated_at LIKE '____-__-__ __:__:__._________ %';

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
        || substr(updated_at, 34, 2)
WHERE created_at OR updated_at LIKE '____-__-__ __:__:__._________ %';

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
        || substr(updated_at, 34, 2)
WHERE created_at OR updated_at LIKE '____-__-__ __:__:__._________ %';

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
WHERE created_at OR updated_at LIKE '____-__-__ __:__:__._________ %';

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
WHERE created_at OR updated_at LIKE '____-__-__ __:__:__._________ %';

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
        || substr(updated_at, 34, 2)
WHERE created_at OR updated_at LIKE '____-__-__ __:__:__._________ %';

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
WHERE created_at OR updated_at LIKE '____-__-__ __:__:__._________ %';

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
        || substr(updated_at, 34, 2)
WHERE created_at OR updated_at LIKE '____-__-__ __:__:__._________ %';