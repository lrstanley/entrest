// Code generated by ent, DO NOT EDIT.

package settings

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

const (
	// Label holds the string label denoting the settings type in the database.
	Label = "settings"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// FieldGlobalBanner holds the string denoting the global_banner field in the database.
	FieldGlobalBanner = "global_banner"
	// EdgeAdmins holds the string denoting the admins edge name in mutations.
	EdgeAdmins = "admins"
	// Table holds the table name of the settings in the database.
	Table = "settings"
	// AdminsTable is the table that holds the admins relation/edge.
	AdminsTable = "users"
	// AdminsInverseTable is the table name for the User entity.
	// It exists in this package in order to avoid circular dependency with the "user" package.
	AdminsInverseTable = "users"
	// AdminsColumn is the table column denoting the admins relation/edge.
	AdminsColumn = "settings_admins"
)

// Columns holds all SQL columns for settings fields.
var Columns = []string{
	FieldID,
	FieldCreatedAt,
	FieldUpdatedAt,
	FieldGlobalBanner,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
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
	// GlobalBannerValidator is a validator for the "global_banner" field. It is called by the builders before save.
	GlobalBannerValidator func(string) error
)

// OrderOption defines the ordering options for the Settings queries.
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

// ByGlobalBanner orders the results by the global_banner field.
func ByGlobalBanner(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldGlobalBanner, opts...).ToFunc()
}

// ByAdminsCount orders the results by admins count.
func ByAdminsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newAdminsStep(), opts...)
	}
}

// ByAdmins orders the results by admins terms.
func ByAdmins(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newAdminsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newAdminsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(AdminsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, AdminsTable, AdminsColumn),
	)
}