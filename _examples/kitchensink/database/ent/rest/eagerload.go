// Code generated by ent, DO NOT EDIT.

package rest

import "github.com/lrstanley/entrest/_examples/kitchensink/database/ent"

// EagerLoadCategory eager-loads the edges of a Category entity, if any edges
// were requested to be eager-loaded, based off associated annotations.
func EagerLoadCategory(query *ent.CategoryQuery) *ent.CategoryQuery {
	return query
}

// EagerLoadFollow eager-loads the edges of a Follow entity, if any edges
// were requested to be eager-loaded, based off associated annotations.
func EagerLoadFollow(query *ent.FollowsQuery) *ent.FollowsQuery {
	return query.WithUser().WithPet()
}

// EagerLoadFriendship eager-loads the edges of a Friendship entity, if any edges
// were requested to be eager-loaded, based off associated annotations.
func EagerLoadFriendship(query *ent.FriendshipQuery) *ent.FriendshipQuery {
	return query
}

// EagerLoadPet eager-loads the edges of a Pet entity, if any edges
// were requested to be eager-loaded, based off associated annotations.
func EagerLoadPet(query *ent.PetQuery) *ent.PetQuery {
	return query.WithCategories().WithOwner()
}

// EagerLoadSetting eager-loads the edges of a Setting entity, if any edges
// were requested to be eager-loaded, based off associated annotations.
func EagerLoadSetting(query *ent.SettingsQuery) *ent.SettingsQuery {
	return query.WithAdmins()
}

// EagerLoadUser eager-loads the edges of a User entity, if any edges
// were requested to be eager-loaded, based off associated annotations.
func EagerLoadUser(query *ent.UserQuery) *ent.UserQuery {
	return query.WithPets()
}
