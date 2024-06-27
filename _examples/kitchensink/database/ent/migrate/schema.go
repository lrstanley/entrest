// Code generated by ent, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// CategoriesColumns holds the columns for the "categories" table.
	CategoriesColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "name", Type: field.TypeString},
		{Name: "readonly", Type: field.TypeString},
		{Name: "skip_in_spec", Type: field.TypeString},
	}
	// CategoriesTable holds the schema information for the "categories" table.
	CategoriesTable = &schema.Table{
		Name:       "categories",
		Columns:    CategoriesColumns,
		PrimaryKey: []*schema.Column{CategoriesColumns[0]},
	}
	// FollowsColumns holds the columns for the "follows" table.
	FollowsColumns = []*schema.Column{
		{Name: "followed_at", Type: field.TypeTime},
		{Name: "user_id", Type: field.TypeInt},
		{Name: "pet_id", Type: field.TypeInt},
	}
	// FollowsTable holds the schema information for the "follows" table.
	FollowsTable = &schema.Table{
		Name:       "follows",
		Columns:    FollowsColumns,
		PrimaryKey: []*schema.Column{FollowsColumns[1], FollowsColumns[2]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "follows_users_user",
				Columns:    []*schema.Column{FollowsColumns[1]},
				RefColumns: []*schema.Column{UsersColumns[0]},
				OnDelete:   schema.Cascade,
			},
			{
				Symbol:     "follows_pets_pet",
				Columns:    []*schema.Column{FollowsColumns[2]},
				RefColumns: []*schema.Column{PetsColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
	}
	// FriendshipsColumns holds the columns for the "friendships" table.
	FriendshipsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "user_id", Type: field.TypeInt},
		{Name: "friend_id", Type: field.TypeInt},
	}
	// FriendshipsTable holds the schema information for the "friendships" table.
	FriendshipsTable = &schema.Table{
		Name:       "friendships",
		Columns:    FriendshipsColumns,
		PrimaryKey: []*schema.Column{FriendshipsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "friendships_users_user",
				Columns:    []*schema.Column{FriendshipsColumns[2]},
				RefColumns: []*schema.Column{UsersColumns[0]},
				OnDelete:   schema.Cascade,
			},
			{
				Symbol:     "friendships_users_friend",
				Columns:    []*schema.Column{FriendshipsColumns[3]},
				RefColumns: []*schema.Column{UsersColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
		Indexes: []*schema.Index{
			{
				Name:    "friendship_user_id_friend_id",
				Unique:  true,
				Columns: []*schema.Column{FriendshipsColumns[2], FriendshipsColumns[3]},
			},
		},
	}
	// PetsColumns holds the columns for the "pets" table.
	PetsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "name", Type: field.TypeString},
		{Name: "nicknames", Type: field.TypeJSON, Nullable: true},
		{Name: "age", Type: field.TypeInt, Nullable: true},
		{Name: "user_pets", Type: field.TypeInt, Nullable: true},
	}
	// PetsTable holds the schema information for the "pets" table.
	PetsTable = &schema.Table{
		Name:       "pets",
		Columns:    PetsColumns,
		PrimaryKey: []*schema.Column{PetsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "pets_users_pets",
				Columns:    []*schema.Column{PetsColumns[4]},
				RefColumns: []*schema.Column{UsersColumns[0]},
				OnDelete:   schema.SetNull,
			},
		},
	}
	// SettingsColumns holds the columns for the "settings" table.
	SettingsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "global_banner", Type: field.TypeString, Nullable: true, Size: 1000},
	}
	// SettingsTable holds the schema information for the "settings" table.
	SettingsTable = &schema.Table{
		Name:       "settings",
		Columns:    SettingsColumns,
		PrimaryKey: []*schema.Column{SettingsColumns[0]},
	}
	// SkippedsColumns holds the columns for the "skippeds" table.
	SkippedsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "name", Type: field.TypeString},
	}
	// SkippedsTable holds the schema information for the "skippeds" table.
	SkippedsTable = &schema.Table{
		Name:       "skippeds",
		Columns:    SkippedsColumns,
		PrimaryKey: []*schema.Column{SkippedsColumns[0]},
	}
	// UsersColumns holds the columns for the "users" table.
	UsersColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "name", Type: field.TypeString},
		{Name: "type", Type: field.TypeEnum, Enums: []string{"SYSTEM", "USER"}, Default: "USER"},
		{Name: "description", Type: field.TypeString, Nullable: true, Size: 1000},
		{Name: "enabled", Type: field.TypeBool, Default: true},
		{Name: "email", Type: field.TypeString, Nullable: true, Size: 320},
		{Name: "avatar", Type: field.TypeBytes, Nullable: true, Size: 1048576},
		{Name: "settings_admins", Type: field.TypeInt, Nullable: true},
	}
	// UsersTable holds the schema information for the "users" table.
	UsersTable = &schema.Table{
		Name:       "users",
		Columns:    UsersColumns,
		PrimaryKey: []*schema.Column{UsersColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "users_settings_admins",
				Columns:    []*schema.Column{UsersColumns[9]},
				RefColumns: []*schema.Column{SettingsColumns[0]},
				OnDelete:   schema.SetNull,
			},
		},
	}
	// CategoryPetsColumns holds the columns for the "category_pets" table.
	CategoryPetsColumns = []*schema.Column{
		{Name: "category_id", Type: field.TypeInt},
		{Name: "pet_id", Type: field.TypeInt},
	}
	// CategoryPetsTable holds the schema information for the "category_pets" table.
	CategoryPetsTable = &schema.Table{
		Name:       "category_pets",
		Columns:    CategoryPetsColumns,
		PrimaryKey: []*schema.Column{CategoryPetsColumns[0], CategoryPetsColumns[1]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "category_pets_category_id",
				Columns:    []*schema.Column{CategoryPetsColumns[0]},
				RefColumns: []*schema.Column{CategoriesColumns[0]},
				OnDelete:   schema.Cascade,
			},
			{
				Symbol:     "category_pets_pet_id",
				Columns:    []*schema.Column{CategoryPetsColumns[1]},
				RefColumns: []*schema.Column{PetsColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
	}
	// PetFriendsColumns holds the columns for the "pet_friends" table.
	PetFriendsColumns = []*schema.Column{
		{Name: "pet_id", Type: field.TypeInt},
		{Name: "friend_id", Type: field.TypeInt},
	}
	// PetFriendsTable holds the schema information for the "pet_friends" table.
	PetFriendsTable = &schema.Table{
		Name:       "pet_friends",
		Columns:    PetFriendsColumns,
		PrimaryKey: []*schema.Column{PetFriendsColumns[0], PetFriendsColumns[1]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "pet_friends_pet_id",
				Columns:    []*schema.Column{PetFriendsColumns[0]},
				RefColumns: []*schema.Column{PetsColumns[0]},
				OnDelete:   schema.Cascade,
			},
			{
				Symbol:     "pet_friends_friend_id",
				Columns:    []*schema.Column{PetFriendsColumns[1]},
				RefColumns: []*schema.Column{PetsColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		CategoriesTable,
		FollowsTable,
		FriendshipsTable,
		PetsTable,
		SettingsTable,
		SkippedsTable,
		UsersTable,
		CategoryPetsTable,
		PetFriendsTable,
	}
)

func init() {
	FollowsTable.ForeignKeys[0].RefTable = UsersTable
	FollowsTable.ForeignKeys[1].RefTable = PetsTable
	FriendshipsTable.ForeignKeys[0].RefTable = UsersTable
	FriendshipsTable.ForeignKeys[1].RefTable = UsersTable
	PetsTable.ForeignKeys[0].RefTable = UsersTable
	UsersTable.ForeignKeys[0].RefTable = SettingsTable
	CategoryPetsTable.ForeignKeys[0].RefTable = CategoriesTable
	CategoryPetsTable.ForeignKeys[1].RefTable = PetsTable
	PetFriendsTable.ForeignKeys[0].RefTable = PetsTable
	PetFriendsTable.ForeignKeys[1].RefTable = PetsTable
}
