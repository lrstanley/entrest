// Code generated by ent, DO NOT EDIT.

package pet

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

const (
	// Label holds the string label denoting the pet type in the database.
	Label = "pet"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldNicknames holds the string denoting the nicknames field in the database.
	FieldNicknames = "nicknames"
	// FieldAge holds the string denoting the age field in the database.
	FieldAge = "age"
	// EdgeCategories holds the string denoting the categories edge name in mutations.
	EdgeCategories = "categories"
	// EdgeOwner holds the string denoting the owner edge name in mutations.
	EdgeOwner = "owner"
	// EdgeFriends holds the string denoting the friends edge name in mutations.
	EdgeFriends = "friends"
	// EdgeFollowedBy holds the string denoting the followed_by edge name in mutations.
	EdgeFollowedBy = "followed_by"
	// EdgeFollowing holds the string denoting the following edge name in mutations.
	EdgeFollowing = "following"
	// Table holds the table name of the pet in the database.
	Table = "pets"
	// CategoriesTable is the table that holds the categories relation/edge. The primary key declared below.
	CategoriesTable = "category_pets"
	// CategoriesInverseTable is the table name for the Category entity.
	// It exists in this package in order to avoid circular dependency with the "category" package.
	CategoriesInverseTable = "categories"
	// OwnerTable is the table that holds the owner relation/edge.
	OwnerTable = "pets"
	// OwnerInverseTable is the table name for the User entity.
	// It exists in this package in order to avoid circular dependency with the "user" package.
	OwnerInverseTable = "users"
	// OwnerColumn is the table column denoting the owner relation/edge.
	OwnerColumn = "user_pets"
	// FriendsTable is the table that holds the friends relation/edge. The primary key declared below.
	FriendsTable = "pet_friends"
	// FollowedByTable is the table that holds the followed_by relation/edge. The primary key declared below.
	FollowedByTable = "follows"
	// FollowedByInverseTable is the table name for the User entity.
	// It exists in this package in order to avoid circular dependency with the "user" package.
	FollowedByInverseTable = "users"
	// FollowingTable is the table that holds the following relation/edge.
	FollowingTable = "follows"
	// FollowingInverseTable is the table name for the Follows entity.
	// It exists in this package in order to avoid circular dependency with the "follows" package.
	FollowingInverseTable = "follows"
	// FollowingColumn is the table column denoting the following relation/edge.
	FollowingColumn = "pet_id"
)

// Columns holds all SQL columns for pet fields.
var Columns = []string{
	FieldID,
	FieldName,
	FieldNicknames,
	FieldAge,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "pets"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"user_pets",
}

var (
	// CategoriesPrimaryKey and CategoriesColumn2 are the table columns denoting the
	// primary key for the categories relation (M2M).
	CategoriesPrimaryKey = []string{"category_id", "pet_id"}
	// FriendsPrimaryKey and FriendsColumn2 are the table columns denoting the
	// primary key for the friends relation (M2M).
	FriendsPrimaryKey = []string{"pet_id", "friend_id"}
	// FollowedByPrimaryKey and FollowedByColumn2 are the table columns denoting the
	// primary key for the followed_by relation (M2M).
	FollowedByPrimaryKey = []string{"user_id", "pet_id"}
)

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

// OrderOption defines the ordering options for the Pet queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByName orders the results by the name field.
func ByName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldName, opts...).ToFunc()
}

// ByAge orders the results by the age field.
func ByAge(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAge, opts...).ToFunc()
}

// ByCategoriesCount orders the results by categories count.
func ByCategoriesCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newCategoriesStep(), opts...)
	}
}

// ByCategories orders the results by categories terms.
func ByCategories(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newCategoriesStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByOwnerField orders the results by owner field.
func ByOwnerField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newOwnerStep(), sql.OrderByField(field, opts...))
	}
}

// ByFriendsCount orders the results by friends count.
func ByFriendsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newFriendsStep(), opts...)
	}
}

// ByFriends orders the results by friends terms.
func ByFriends(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newFriendsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByFollowedByCount orders the results by followed_by count.
func ByFollowedByCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newFollowedByStep(), opts...)
	}
}

// ByFollowedBy orders the results by followed_by terms.
func ByFollowedBy(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newFollowedByStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByFollowingCount orders the results by following count.
func ByFollowingCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newFollowingStep(), opts...)
	}
}

// ByFollowing orders the results by following terms.
func ByFollowing(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newFollowingStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newCategoriesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(CategoriesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, true, CategoriesTable, CategoriesPrimaryKey...),
	)
}
func newOwnerStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(OwnerInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, OwnerTable, OwnerColumn),
	)
}
func newFriendsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(Table, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, false, FriendsTable, FriendsPrimaryKey...),
	)
}
func newFollowedByStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(FollowedByInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, true, FollowedByTable, FollowedByPrimaryKey...),
	)
}
func newFollowingStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(FollowingInverseTable, FollowingColumn),
		sqlgraph.Edge(sqlgraph.O2M, true, FollowingTable, FollowingColumn),
	)
}