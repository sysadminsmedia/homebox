-- +goose Up
CREATE TABLE maintenance_plans (
  id                uuid                     NOT NULL PRIMARY KEY,
  created_at        timestamp with time zone NOT NULL,
  updated_at        timestamp with time zone NOT NULL,
  entity_id         uuid                     NOT NULL,
  name              text                     NOT NULL,
  description       text,
  interval_value    bigint                   NOT NULL,
  interval_unit     text                     NOT NULL,
  active            boolean DEFAULT true     NOT NULL,
  last_completed_at timestamp with time zone,
  next_due_at       timestamp with time zone,
  CONSTRAINT maintenance_plans_entities_maintenance_plans
    FOREIGN KEY (entity_id) REFERENCES entities (id) ON DELETE CASCADE
);

CREATE INDEX idx_maintenance_plans_entity_id ON maintenance_plans (entity_id);

ALTER TABLE maintenance_entries
  ADD COLUMN plan_id uuid;

CREATE INDEX idx_maintenance_entries_plan_id ON maintenance_entries (plan_id);

-- +goose Down
DROP INDEX IF EXISTS idx_maintenance_entries_plan_id;
ALTER TABLE maintenance_entries DROP COLUMN IF EXISTS plan_id;
DROP INDEX IF EXISTS idx_maintenance_plans_entity_id;
DROP TABLE IF EXISTS maintenance_plans;
