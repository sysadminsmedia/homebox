-- +goose Up
-- The location-override column is already indexed. Index the fallback
-- parent so the descendant CTE can use both indexes instead of scanning all
-- entities in the group for each recursion step.
CREATE INDEX IF NOT EXISTS idx_entities_parent_group
    ON entities (entity_children, group_entities);

-- +goose Down
DROP INDEX IF EXISTS idx_entities_parent_group;
