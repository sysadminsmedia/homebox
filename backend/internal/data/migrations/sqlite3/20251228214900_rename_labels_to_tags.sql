-- +goose Up
-- +goose no transaction
PRAGMA foreign_keys=off;

CREATE TABLE IF NOT EXISTS tags (
	id           uuid     not null
		primary key,
	created_at   datetime not null,
	updated_at   datetime not null,
	name         text     not null,
	description  text,
	color        text,
	group_tags   uuid     not null
		constraint tags_groups_tags
			references groups
			on delete cascade
);

INSERT INTO tags_temp(id, created_at, updated_at, name, description, color, group_tags)
SELECT id, created_at, updated_at, name, description, color, group_labels FROM labels;

DROP TABLE labels;


CREATE TABLE IF NOT EXISTS tag_items (
	tag_id uuid not null
		constraint tag_items_tag_id
			references tags
			on delete cascade,
	item_id uuid not null
		constraint tag_items_item_id
			references items
			on delete cascade,
	primary key (tag_id, item_id)
);

INSERT INTO tag_items(tag_id, item_id)
SELECT label_id, item_id FROM label_items;

DROP TABLE IF EXISTS label_items;


CREATE TABLE IF NOT EXISTS item_templates_temp
(
	id                           uuid                  not null
		primary key,
	created_at                   datetime              not null,
	updated_at                   datetime              not null,
	name                         text                  not null,
	description                  text,
	notes                        text,
	default_quantity             integer default 1     not null,
	default_insured              bool    default false not null,
	default_name                 text,
	default_description          text,
	default_manufacturer         text,
	default_model_number         text,
	default_lifetime_warranty    bool    default false not null,
	default_warranty_details     text,
	include_warranty_fields      bool    default false not null,
	include_purchase_fields      bool    default false not null,
	include_sold_fields          bool    default false not null,
	default_tag_ids              json,
	item_template_location       uuid
		references locations(id)
			on delete set null,
	group_item_templates         uuid                  not null
		constraint item_templates_groups_item_templates
			references groups
			on delete cascade
);

CREATE TABLE IF NOT EXISTS template_fields_temp
(
	id                   uuid                  not null
		primary key,
	created_at           datetime              not null,
	updated_at           datetime              not null,
	name                 text                  not null,
	description          text,
	type                 text                  not null,
	text_value           text,
	item_template_fields uuid
		constraint template_fields_item_templates_fields
			references item_templates_temp
			on delete cascade
);

INSERT INTO item_templates_temp(id, created_at, updated_at, name, description, notes, default_quantity, default_insured, default_name, default_description, default_manufacturer, default_model_number, default_lifetime_warranty, default_warranty_details, include_warranty_fields, include_purchase_fields, include_sold_fields, default_tag_ids, item_template_location, group_item_templates)
SELECT id, created_at, updated_at, name, description, notes, default_quantity, default_insured, default_name, default_description, default_manufacturer, default_model_number, default_lifetime_warranty, default_warranty_details, include_warranty_fields, include_purchase_fields, include_sold_fields, default_label_ids, item_template_location, group_item_templates FROM item_templates;

INSERT INTO template_fields_temp(id, created_at, updated_at, name, description, type, text_value, item_template_fields)
SELECT id, created_at, updated_at, name, description, type, text_value, item_template_fields FROM template_fields;

DROP TABLE IF EXISTS template_fields;
DROP TABLE IF EXISTS item_templates;
ALTER TABLE item_templates_temp RENAME TO item_templates;
ALTER TABLE template_fields_temp RENAME TO template_fields;

PRAGMA foreign_keys=on;