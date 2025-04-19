// Code generated by ent, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// AttachmentsColumns holds the columns for the "attachments" table.
	AttachmentsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "type", Type: field.TypeEnum, Enums: []string{"photo", "manual", "warranty", "attachment", "receipt"}, Default: "attachment"},
		{Name: "primary", Type: field.TypeBool, Default: false},
		{Name: "title", Type: field.TypeString, Default: ""},
		{Name: "path", Type: field.TypeString, Default: ""},
		{Name: "item_attachments", Type: field.TypeUUID},
	}
	// AttachmentsTable holds the schema information for the "attachments" table.
	AttachmentsTable = &schema.Table{
		Name:       "attachments",
		Columns:    AttachmentsColumns,
		PrimaryKey: []*schema.Column{AttachmentsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "attachments_items_attachments",
				Columns:    []*schema.Column{AttachmentsColumns[7]},
				RefColumns: []*schema.Column{ItemsColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
	}
	// AuthRolesColumns holds the columns for the "auth_roles" table.
	AuthRolesColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "role", Type: field.TypeEnum, Enums: []string{"admin", "user", "attachments"}, Default: "user"},
		{Name: "auth_tokens_roles", Type: field.TypeUUID, Unique: true, Nullable: true},
	}
	// AuthRolesTable holds the schema information for the "auth_roles" table.
	AuthRolesTable = &schema.Table{
		Name:       "auth_roles",
		Columns:    AuthRolesColumns,
		PrimaryKey: []*schema.Column{AuthRolesColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "auth_roles_auth_tokens_roles",
				Columns:    []*schema.Column{AuthRolesColumns[2]},
				RefColumns: []*schema.Column{AuthTokensColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
	}
	// AuthTokensColumns holds the columns for the "auth_tokens" table.
	AuthTokensColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "token", Type: field.TypeBytes, Unique: true},
		{Name: "expires_at", Type: field.TypeTime},
		{Name: "user_auth_tokens", Type: field.TypeUUID, Nullable: true},
	}
	// AuthTokensTable holds the schema information for the "auth_tokens" table.
	AuthTokensTable = &schema.Table{
		Name:       "auth_tokens",
		Columns:    AuthTokensColumns,
		PrimaryKey: []*schema.Column{AuthTokensColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "auth_tokens_users_auth_tokens",
				Columns:    []*schema.Column{AuthTokensColumns[5]},
				RefColumns: []*schema.Column{UsersColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
		Indexes: []*schema.Index{
			{
				Name:    "authtokens_token",
				Unique:  false,
				Columns: []*schema.Column{AuthTokensColumns[3]},
			},
		},
	}
	// GroupsColumns holds the columns for the "groups" table.
	GroupsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "name", Type: field.TypeString, Size: 255},
		{Name: "currency", Type: field.TypeString, Default: "usd"},
	}
	// GroupsTable holds the schema information for the "groups" table.
	GroupsTable = &schema.Table{
		Name:       "groups",
		Columns:    GroupsColumns,
		PrimaryKey: []*schema.Column{GroupsColumns[0]},
	}
	// GroupInvitationTokensColumns holds the columns for the "group_invitation_tokens" table.
	GroupInvitationTokensColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "token", Type: field.TypeBytes, Unique: true},
		{Name: "expires_at", Type: field.TypeTime},
		{Name: "uses", Type: field.TypeInt, Default: 0},
		{Name: "group_invitation_tokens", Type: field.TypeUUID, Nullable: true},
	}
	// GroupInvitationTokensTable holds the schema information for the "group_invitation_tokens" table.
	GroupInvitationTokensTable = &schema.Table{
		Name:       "group_invitation_tokens",
		Columns:    GroupInvitationTokensColumns,
		PrimaryKey: []*schema.Column{GroupInvitationTokensColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "group_invitation_tokens_groups_invitation_tokens",
				Columns:    []*schema.Column{GroupInvitationTokensColumns[6]},
				RefColumns: []*schema.Column{GroupsColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
	}
	// ItemsColumns holds the columns for the "items" table.
	ItemsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "name", Type: field.TypeString, Size: 255},
		{Name: "description", Type: field.TypeString, Nullable: true, Size: 1000},
		{Name: "import_ref", Type: field.TypeString, Nullable: true, Size: 100},
		{Name: "notes", Type: field.TypeString, Nullable: true, Size: 1000},
		{Name: "quantity", Type: field.TypeInt, Default: 1},
		{Name: "insured", Type: field.TypeBool, Default: false},
		{Name: "archived", Type: field.TypeBool, Default: false},
		{Name: "asset_id", Type: field.TypeInt, Default: 0},
		{Name: "sync_child_items_locations", Type: field.TypeBool, Default: false},
		{Name: "serial_number", Type: field.TypeString, Nullable: true, Size: 255},
		{Name: "model_number", Type: field.TypeString, Nullable: true, Size: 255},
		{Name: "manufacturer", Type: field.TypeString, Nullable: true, Size: 255},
		{Name: "lifetime_warranty", Type: field.TypeBool, Default: false},
		{Name: "warranty_expires", Type: field.TypeTime, Nullable: true},
		{Name: "warranty_details", Type: field.TypeString, Nullable: true, Size: 1000},
		{Name: "purchase_time", Type: field.TypeTime, Nullable: true},
		{Name: "purchase_from", Type: field.TypeString, Nullable: true},
		{Name: "purchase_price", Type: field.TypeFloat64, Default: 0},
		{Name: "sold_time", Type: field.TypeTime, Nullable: true},
		{Name: "sold_to", Type: field.TypeString, Nullable: true},
		{Name: "sold_price", Type: field.TypeFloat64, Default: 0},
		{Name: "sold_notes", Type: field.TypeString, Nullable: true, Size: 1000},
		{Name: "group_items", Type: field.TypeUUID},
		{Name: "item_children", Type: field.TypeUUID, Nullable: true},
		{Name: "location_items", Type: field.TypeUUID, Nullable: true},
	}
	// ItemsTable holds the schema information for the "items" table.
	ItemsTable = &schema.Table{
		Name:       "items",
		Columns:    ItemsColumns,
		PrimaryKey: []*schema.Column{ItemsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "items_groups_items",
				Columns:    []*schema.Column{ItemsColumns[25]},
				RefColumns: []*schema.Column{GroupsColumns[0]},
				OnDelete:   schema.Cascade,
			},
			{
				Symbol:     "items_items_children",
				Columns:    []*schema.Column{ItemsColumns[26]},
				RefColumns: []*schema.Column{ItemsColumns[0]},
				OnDelete:   schema.SetNull,
			},
			{
				Symbol:     "items_locations_items",
				Columns:    []*schema.Column{ItemsColumns[27]},
				RefColumns: []*schema.Column{LocationsColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
		Indexes: []*schema.Index{
			{
				Name:    "item_name",
				Unique:  false,
				Columns: []*schema.Column{ItemsColumns[3]},
			},
			{
				Name:    "item_manufacturer",
				Unique:  false,
				Columns: []*schema.Column{ItemsColumns[14]},
			},
			{
				Name:    "item_model_number",
				Unique:  false,
				Columns: []*schema.Column{ItemsColumns[13]},
			},
			{
				Name:    "item_serial_number",
				Unique:  false,
				Columns: []*schema.Column{ItemsColumns[12]},
			},
			{
				Name:    "item_archived",
				Unique:  false,
				Columns: []*schema.Column{ItemsColumns[9]},
			},
			{
				Name:    "item_asset_id",
				Unique:  false,
				Columns: []*schema.Column{ItemsColumns[10]},
			},
		},
	}
	// ItemFieldsColumns holds the columns for the "item_fields" table.
	ItemFieldsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "name", Type: field.TypeString, Size: 255},
		{Name: "description", Type: field.TypeString, Nullable: true, Size: 1000},
		{Name: "type", Type: field.TypeEnum, Enums: []string{"text", "number", "boolean", "time"}},
		{Name: "text_value", Type: field.TypeString, Nullable: true, Size: 500},
		{Name: "number_value", Type: field.TypeInt, Nullable: true},
		{Name: "boolean_value", Type: field.TypeBool, Default: false},
		{Name: "time_value", Type: field.TypeTime},
		{Name: "item_fields", Type: field.TypeUUID, Nullable: true},
	}
	// ItemFieldsTable holds the schema information for the "item_fields" table.
	ItemFieldsTable = &schema.Table{
		Name:       "item_fields",
		Columns:    ItemFieldsColumns,
		PrimaryKey: []*schema.Column{ItemFieldsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "item_fields_items_fields",
				Columns:    []*schema.Column{ItemFieldsColumns[10]},
				RefColumns: []*schema.Column{ItemsColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
	}
	// LabelsColumns holds the columns for the "labels" table.
	LabelsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "name", Type: field.TypeString, Size: 255},
		{Name: "description", Type: field.TypeString, Nullable: true, Size: 1000},
		{Name: "color", Type: field.TypeString, Nullable: true, Size: 255},
		{Name: "group_labels", Type: field.TypeUUID},
	}
	// LabelsTable holds the schema information for the "labels" table.
	LabelsTable = &schema.Table{
		Name:       "labels",
		Columns:    LabelsColumns,
		PrimaryKey: []*schema.Column{LabelsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "labels_groups_labels",
				Columns:    []*schema.Column{LabelsColumns[6]},
				RefColumns: []*schema.Column{GroupsColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
	}
	// LocationsColumns holds the columns for the "locations" table.
	LocationsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "name", Type: field.TypeString, Size: 255},
		{Name: "description", Type: field.TypeString, Nullable: true, Size: 1000},
		{Name: "group_locations", Type: field.TypeUUID},
		{Name: "location_children", Type: field.TypeUUID, Nullable: true},
	}
	// LocationsTable holds the schema information for the "locations" table.
	LocationsTable = &schema.Table{
		Name:       "locations",
		Columns:    LocationsColumns,
		PrimaryKey: []*schema.Column{LocationsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "locations_groups_locations",
				Columns:    []*schema.Column{LocationsColumns[5]},
				RefColumns: []*schema.Column{GroupsColumns[0]},
				OnDelete:   schema.Cascade,
			},
			{
				Symbol:     "locations_locations_children",
				Columns:    []*schema.Column{LocationsColumns[6]},
				RefColumns: []*schema.Column{LocationsColumns[0]},
				OnDelete:   schema.SetNull,
			},
		},
	}
	// MaintenanceEntriesColumns holds the columns for the "maintenance_entries" table.
	MaintenanceEntriesColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "date", Type: field.TypeTime, Nullable: true},
		{Name: "scheduled_date", Type: field.TypeTime, Nullable: true},
		{Name: "name", Type: field.TypeString, Size: 255},
		{Name: "description", Type: field.TypeString, Nullable: true, Size: 2500},
		{Name: "cost", Type: field.TypeFloat64, Default: 0},
		{Name: "item_id", Type: field.TypeUUID},
	}
	// MaintenanceEntriesTable holds the schema information for the "maintenance_entries" table.
	MaintenanceEntriesTable = &schema.Table{
		Name:       "maintenance_entries",
		Columns:    MaintenanceEntriesColumns,
		PrimaryKey: []*schema.Column{MaintenanceEntriesColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "maintenance_entries_items_maintenance_entries",
				Columns:    []*schema.Column{MaintenanceEntriesColumns[8]},
				RefColumns: []*schema.Column{ItemsColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
	}
	// NotifiersColumns holds the columns for the "notifiers" table.
	NotifiersColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "name", Type: field.TypeString, Size: 255},
		{Name: "url", Type: field.TypeString, Size: 2083},
		{Name: "is_active", Type: field.TypeBool, Default: true},
		{Name: "group_id", Type: field.TypeUUID},
		{Name: "user_id", Type: field.TypeUUID},
	}
	// NotifiersTable holds the schema information for the "notifiers" table.
	NotifiersTable = &schema.Table{
		Name:       "notifiers",
		Columns:    NotifiersColumns,
		PrimaryKey: []*schema.Column{NotifiersColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "notifiers_groups_notifiers",
				Columns:    []*schema.Column{NotifiersColumns[6]},
				RefColumns: []*schema.Column{GroupsColumns[0]},
				OnDelete:   schema.Cascade,
			},
			{
				Symbol:     "notifiers_users_notifiers",
				Columns:    []*schema.Column{NotifiersColumns[7]},
				RefColumns: []*schema.Column{UsersColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
		Indexes: []*schema.Index{
			{
				Name:    "notifier_user_id",
				Unique:  false,
				Columns: []*schema.Column{NotifiersColumns[7]},
			},
			{
				Name:    "notifier_user_id_is_active",
				Unique:  false,
				Columns: []*schema.Column{NotifiersColumns[7], NotifiersColumns[5]},
			},
			{
				Name:    "notifier_group_id",
				Unique:  false,
				Columns: []*schema.Column{NotifiersColumns[6]},
			},
			{
				Name:    "notifier_group_id_is_active",
				Unique:  false,
				Columns: []*schema.Column{NotifiersColumns[6], NotifiersColumns[5]},
			},
		},
	}
	// UsersColumns holds the columns for the "users" table.
	UsersColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "name", Type: field.TypeString, Size: 255},
		{Name: "email", Type: field.TypeString, Unique: true, Size: 255},
		{Name: "password", Type: field.TypeString, Size: 255},
		{Name: "is_superuser", Type: field.TypeBool, Default: false},
		{Name: "superuser", Type: field.TypeBool, Default: false},
		{Name: "role", Type: field.TypeEnum, Enums: []string{"user", "owner"}, Default: "user"},
		{Name: "activated_on", Type: field.TypeTime, Nullable: true},
		{Name: "group_users", Type: field.TypeUUID},
	}
	// UsersTable holds the schema information for the "users" table.
	UsersTable = &schema.Table{
		Name:       "users",
		Columns:    UsersColumns,
		PrimaryKey: []*schema.Column{UsersColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "users_groups_users",
				Columns:    []*schema.Column{UsersColumns[10]},
				RefColumns: []*schema.Column{GroupsColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
	}
	// LabelItemsColumns holds the columns for the "label_items" table.
	LabelItemsColumns = []*schema.Column{
		{Name: "label_id", Type: field.TypeUUID},
		{Name: "item_id", Type: field.TypeUUID},
	}
	// LabelItemsTable holds the schema information for the "label_items" table.
	LabelItemsTable = &schema.Table{
		Name:       "label_items",
		Columns:    LabelItemsColumns,
		PrimaryKey: []*schema.Column{LabelItemsColumns[0], LabelItemsColumns[1]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "label_items_label_id",
				Columns:    []*schema.Column{LabelItemsColumns[0]},
				RefColumns: []*schema.Column{LabelsColumns[0]},
				OnDelete:   schema.Cascade,
			},
			{
				Symbol:     "label_items_item_id",
				Columns:    []*schema.Column{LabelItemsColumns[1]},
				RefColumns: []*schema.Column{ItemsColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		AttachmentsTable,
		AuthRolesTable,
		AuthTokensTable,
		GroupsTable,
		GroupInvitationTokensTable,
		ItemsTable,
		ItemFieldsTable,
		LabelsTable,
		LocationsTable,
		MaintenanceEntriesTable,
		NotifiersTable,
		UsersTable,
		LabelItemsTable,
	}
)

func init() {
	AttachmentsTable.ForeignKeys[0].RefTable = ItemsTable
	AuthRolesTable.ForeignKeys[0].RefTable = AuthTokensTable
	AuthTokensTable.ForeignKeys[0].RefTable = UsersTable
	GroupInvitationTokensTable.ForeignKeys[0].RefTable = GroupsTable
	ItemsTable.ForeignKeys[0].RefTable = GroupsTable
	ItemsTable.ForeignKeys[1].RefTable = ItemsTable
	ItemsTable.ForeignKeys[2].RefTable = LocationsTable
	ItemFieldsTable.ForeignKeys[0].RefTable = ItemsTable
	LabelsTable.ForeignKeys[0].RefTable = GroupsTable
	LocationsTable.ForeignKeys[0].RefTable = GroupsTable
	LocationsTable.ForeignKeys[1].RefTable = LocationsTable
	MaintenanceEntriesTable.ForeignKeys[0].RefTable = ItemsTable
	NotifiersTable.ForeignKeys[0].RefTable = GroupsTable
	NotifiersTable.ForeignKeys[1].RefTable = UsersTable
	UsersTable.ForeignKeys[0].RefTable = GroupsTable
	LabelItemsTable.ForeignKeys[0].RefTable = LabelsTable
	LabelItemsTable.ForeignKeys[1].RefTable = ItemsTable
}
