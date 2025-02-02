-- This migrates some special fields to the new ISO8601 format.


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
WHERE created_at LIKE '____-__-__ __:__:__._________ %';

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
WHERE created_at LIKE '____-__-__ __:__:__._________ %';

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
WHERE created_at LIKE '____-__-__ __:__:__._________ %';

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
WHERE created_at LIKE '____-__-__ __:__:__._________ %';

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
WHERE created_at LIKE '____-__-__ __:__:__._________ %';

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
WHERE created_at LIKE '____-__-__ __:__:__._________ %';

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
WHERE created_at LIKE '____-__-__ __:__:__._________ %';

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
WHERE created_at LIKE '____-__-__ __:__:__._________ %';

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
WHERE created_at LIKE '____-__-__ __:__:__._________ %';

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
WHERE created_at LIKE '____-__-__ __:__:__._________ %';

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
WHERE created_at LIKE '____-__-__ __:__:__._________ %';

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
WHERE created_at LIKE '____-__-__ __:__:__._________ %';