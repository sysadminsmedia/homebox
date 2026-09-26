-- +goose Up
CREATE TABLE maintenance_plans (
  id                uuid                  NOT NULL PRIMARY KEY,
  created_at        datetime              NOT NULL,
  updated_at        datetime              NOT NULL,
  entity_id         uuid                  NOT NULL,
  name              text                  NOT NULL,
  description       text,
  interval_value    integer               NOT NULL,
  interval_unit     text                  NOT NULL,
  active            boolean DEFAULT true  NOT NULL,
  last_completed_at datetime,
  next_due_at       datetime,
  CONSTRAINT maintenance_plans_entities_maintenance_plans
    FOREIGN KEY (entity_id) REFERENCES entities (id) ON DELETE CASCADE
);

CREATE INDEX idx_maintenance_plans_entity_id ON maintenance_plans (entity_id);

CREATE TABLE maintenance_entries_new (
  id             uuid           not null primary key,
  created_at     datetime       not null,
  updated_at     datetime       not null,
  date           datetime,
  scheduled_date datetime,
  name           text           not null,
  description    text,
  cost           real default 0 not null,
  entity_id      uuid           not null
    constraint maintenance_entries_entities_maintenance_entries
      references entities (id) on delete cascade,
  plan_id        uuid
    constraint maintenance_entries_plan_id_fkey
      references maintenance_plans (id) on delete set null
);

INSERT INTO maintenance_entries_new (
  id, created_at, updated_at, date, scheduled_date, name, description, cost, entity_id, plan_id
)
SELECT
  id, created_at, updated_at, date, scheduled_date, name, description, cost, entity_id, NULL
FROM maintenance_entries;

DROP TABLE maintenance_entries;

ALTER TABLE maintenance_entries_new RENAME TO maintenance_entries;

CREATE INDEX idx_maintenance_entries_plan_id ON maintenance_entries (plan_id);

-- +goose Down
DROP INDEX IF EXISTS idx_maintenance_entries_plan_id;

CREATE TABLE maintenance_entries_old (
  id             uuid           not null primary key,
  created_at     datetime       not null,
  updated_at     datetime       not null,
  date           datetime,
  scheduled_date datetime,
  name           text           not null,
  description    text,
  cost           real default 0 not null,
  entity_id      uuid           not null
    constraint maintenance_entries_entities_maintenance_entries_old
      references entities (id) on delete cascade
);

INSERT INTO maintenance_entries_old (
  id, created_at, updated_at, date, scheduled_date, name, description, cost, entity_id
)
SELECT id, created_at, updated_at, date, scheduled_date, name, description, cost, entity_id
FROM maintenance_entries;

DROP TABLE maintenance_entries;

ALTER TABLE maintenance_entries_old RENAME TO maintenance_entries;

DROP INDEX IF EXISTS idx_maintenance_plans_entity_id;
DROP TABLE IF EXISTS maintenance_plans;
