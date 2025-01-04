// Code generated by ent, DO NOT EDIT.

package user

import (
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the user type in the database.
	Label = "user"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldEmail holds the string denoting the email field in the database.
	FieldEmail = "email"
	// FieldPassword holds the string denoting the password field in the database.
	FieldPassword = "password"
	// FieldIsSuperuser holds the string denoting the is_superuser field in the database.
	FieldIsSuperuser = "is_superuser"
	// FieldSuperuser holds the string denoting the superuser field in the database.
	FieldSuperuser = "superuser"
	// FieldRole holds the string denoting the role field in the database.
	FieldRole = "role"
	// FieldActivatedOn holds the string denoting the activated_on field in the database.
	FieldActivatedOn = "activated_on"
	// EdgeGroup holds the string denoting the group edge name in mutations.
	EdgeGroup = "group"
	// EdgeAuthTokens holds the string denoting the auth_tokens edge name in mutations.
	EdgeAuthTokens = "auth_tokens"
	// EdgeNotifiers holds the string denoting the notifiers edge name in mutations.
	EdgeNotifiers = "notifiers"
	// EdgeOauth holds the string denoting the oauth edge name in mutations.
	EdgeOauth = "oauth"
	// Table holds the table name of the user in the database.
	Table = "users"
	// GroupTable is the table that holds the group relation/edge.
	GroupTable = "users"
	// GroupInverseTable is the table name for the Group entity.
	// It exists in this package in order to avoid circular dependency with the "group" package.
	GroupInverseTable = "groups"
	// GroupColumn is the table column denoting the group relation/edge.
	GroupColumn = "group_users"
	// AuthTokensTable is the table that holds the auth_tokens relation/edge.
	AuthTokensTable = "auth_tokens"
	// AuthTokensInverseTable is the table name for the AuthTokens entity.
	// It exists in this package in order to avoid circular dependency with the "authtokens" package.
	AuthTokensInverseTable = "auth_tokens"
	// AuthTokensColumn is the table column denoting the auth_tokens relation/edge.
	AuthTokensColumn = "user_auth_tokens"
	// NotifiersTable is the table that holds the notifiers relation/edge.
	NotifiersTable = "notifiers"
	// NotifiersInverseTable is the table name for the Notifier entity.
	// It exists in this package in order to avoid circular dependency with the "notifier" package.
	NotifiersInverseTable = "notifiers"
	// NotifiersColumn is the table column denoting the notifiers relation/edge.
	NotifiersColumn = "user_id"
	// OauthTable is the table that holds the oauth relation/edge.
	OauthTable = "oauths"
	// OauthInverseTable is the table name for the OAuth entity.
	// It exists in this package in order to avoid circular dependency with the "oauth" package.
	OauthInverseTable = "oauths"
	// OauthColumn is the table column denoting the oauth relation/edge.
	OauthColumn = "user_oauth"
)

// Columns holds all SQL columns for user fields.
var Columns = []string{
	FieldID,
	FieldCreatedAt,
	FieldUpdatedAt,
	FieldName,
	FieldEmail,
	FieldPassword,
	FieldIsSuperuser,
	FieldSuperuser,
	FieldRole,
	FieldActivatedOn,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "users"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"group_users",
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	for i := range ForeignKeys {
		if column == ForeignKeys[i] {
			return true
		}
	}
	return false
}

var (
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
	// NameValidator is a validator for the "name" field. It is called by the builders before save.
	NameValidator func(string) error
	// EmailValidator is a validator for the "email" field. It is called by the builders before save.
	EmailValidator func(string) error
	// PasswordValidator is a validator for the "password" field. It is called by the builders before save.
	PasswordValidator func(string) error
	// DefaultIsSuperuser holds the default value on creation for the "is_superuser" field.
	DefaultIsSuperuser bool
	// DefaultSuperuser holds the default value on creation for the "superuser" field.
	DefaultSuperuser bool
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// Role defines the type for the "role" enum field.
type Role string

// RoleUser is the default value of the Role enum.
const DefaultRole = RoleUser

// Role values.
const (
	RoleUser  Role = "user"
	RoleOwner Role = "owner"
)

func (r Role) String() string {
	return string(r)
}

// RoleValidator is a validator for the "role" field enum values. It is called by the builders before save.
func RoleValidator(r Role) error {
	switch r {
	case RoleUser, RoleOwner:
		return nil
	default:
		return fmt.Errorf("user: invalid enum value for role field: %q", r)
	}
}

// OrderOption defines the ordering options for the User queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByUpdatedAt orders the results by the updated_at field.
func ByUpdatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedAt, opts...).ToFunc()
}

// ByName orders the results by the name field.
func ByName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldName, opts...).ToFunc()
}

// ByEmail orders the results by the email field.
func ByEmail(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldEmail, opts...).ToFunc()
}

// ByPassword orders the results by the password field.
func ByPassword(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPassword, opts...).ToFunc()
}

// ByIsSuperuser orders the results by the is_superuser field.
func ByIsSuperuser(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIsSuperuser, opts...).ToFunc()
}

// BySuperuser orders the results by the superuser field.
func BySuperuser(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSuperuser, opts...).ToFunc()
}

// ByRole orders the results by the role field.
func ByRole(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldRole, opts...).ToFunc()
}

// ByActivatedOn orders the results by the activated_on field.
func ByActivatedOn(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldActivatedOn, opts...).ToFunc()
}

// ByGroupField orders the results by group field.
func ByGroupField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newGroupStep(), sql.OrderByField(field, opts...))
	}
}

// ByAuthTokensCount orders the results by auth_tokens count.
func ByAuthTokensCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newAuthTokensStep(), opts...)
	}
}

// ByAuthTokens orders the results by auth_tokens terms.
func ByAuthTokens(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newAuthTokensStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByNotifiersCount orders the results by notifiers count.
func ByNotifiersCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newNotifiersStep(), opts...)
	}
}

// ByNotifiers orders the results by notifiers terms.
func ByNotifiers(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newNotifiersStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByOauthCount orders the results by oauth count.
func ByOauthCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newOauthStep(), opts...)
	}
}

// ByOauth orders the results by oauth terms.
func ByOauth(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newOauthStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newGroupStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(GroupInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, GroupTable, GroupColumn),
	)
}
func newAuthTokensStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(AuthTokensInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, AuthTokensTable, AuthTokensColumn),
	)
}
func newNotifiersStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(NotifiersInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, NotifiersTable, NotifiersColumn),
	)
}
func newOauthStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(OauthInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, OauthTable, OauthColumn),
	)
}
