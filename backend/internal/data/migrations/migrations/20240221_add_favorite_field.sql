-- Add favorite column to items table
ALTER TABLE `items` ADD COLUMN `favorite` bool NOT NULL DEFAULT false;
-- create index "item_favorite" to table: "items"
CREATE INDEX `item_favorite` ON `items` (`favorite`);
