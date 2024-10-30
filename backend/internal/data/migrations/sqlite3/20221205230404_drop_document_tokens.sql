-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
DROP TABLE `document_tokens`;
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;