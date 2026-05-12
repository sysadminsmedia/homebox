package repo

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
)

func TestEntityRepository_GetFoundEntityContact(t *testing.T) {
	ctx := context.Background()

	g, err := tRepos.Groups.GroupCreate(ctx, "found-item-contact", uuid.Nil)
	require.NoError(t, err)

	password := "password"
	owner, err := tRepos.Users.Create(ctx, UserCreate{
		Name:           "Owner",
		Email:          "owner@example.com",
		Password:       &password,
		DefaultGroupID: g.ID,
		IsOwner:        true,
	})
	require.NoError(t, err)

	member, err := tRepos.Users.Create(ctx, UserCreate{
		Name:           "Member",
		Email:          "member@example.com",
		Password:       &password,
		DefaultGroupID: g.ID,
	})
	require.NoError(t, err)

	itemType, err := tRepos.EntityTypes.GetDefault(ctx, g.ID, false)
	require.NoError(t, err)

	item, err := tRepos.Entities.Create(ctx, g.ID, EntityCreate{
		Name:         "Found umbrella",
		Description:  "Private description",
		EntityTypeID: itemType.ID,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = tRepos.Entities.Delete(ctx, item.ID)
		_ = tRepos.Users.Delete(ctx, member.ID)
		_ = tRepos.Users.Delete(ctx, owner.ID)
		_ = tRepos.Groups.GroupDelete(ctx, g.ID)
	})

	contact, err := tRepos.Entities.GetFoundEntityContact(ctx, item.ID)
	require.NoError(t, err)
	require.Equal(t, item.ID, contact.ItemID)
	require.Equal(t, "owner@example.com", contact.OwnerEmail)
}

func TestEntityRepository_GetFoundEntityContactByAssetID(t *testing.T) {
	ctx := context.Background()

	g, err := tRepos.Groups.GroupCreate(ctx, "found-asset-contact", uuid.Nil)
	require.NoError(t, err)

	password := "password"
	owner, err := tRepos.Users.Create(ctx, UserCreate{
		Name:           "Asset Owner",
		Email:          "asset-owner@example.com",
		Password:       &password,
		DefaultGroupID: g.ID,
		IsOwner:        true,
	})
	require.NoError(t, err)

	item, err := tRepos.Entities.Create(ctx, g.ID, EntityCreate{
		Name:        "Found backpack",
		Description: "Private description",
		AssetID:     AssetID(4242),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = tRepos.Entities.Delete(ctx, item.ID)
		_ = tRepos.Users.Delete(ctx, owner.ID)
		_ = tRepos.Groups.GroupDelete(ctx, g.ID)
	})

	contact, err := tRepos.Entities.GetFoundEntityContactByAssetID(ctx, AssetID(4242))
	require.NoError(t, err)
	require.Equal(t, item.ID, contact.ItemID)
	require.Equal(t, "asset-owner@example.com", contact.OwnerEmail)
}

func TestEntityRepository_GetFoundEntityContactByAssetID_Ambiguous(t *testing.T) {
	ctx := context.Background()
	password := "password"
	assetID := AssetID(4343)

	firstGroup, err := tRepos.Groups.GroupCreate(ctx, "found-asset-ambiguous-1", uuid.Nil)
	require.NoError(t, err)
	firstOwner, err := tRepos.Users.Create(ctx, UserCreate{
		Name:           "First Owner",
		Email:          "asset-owner-1@example.com",
		Password:       &password,
		DefaultGroupID: firstGroup.ID,
		IsOwner:        true,
	})
	require.NoError(t, err)
	firstItem, err := tRepos.Entities.Create(ctx, firstGroup.ID, EntityCreate{
		Name:    "First item",
		AssetID: assetID,
	})
	require.NoError(t, err)

	secondGroup, err := tRepos.Groups.GroupCreate(ctx, "found-asset-ambiguous-2", uuid.Nil)
	require.NoError(t, err)
	secondOwner, err := tRepos.Users.Create(ctx, UserCreate{
		Name:           "Second Owner",
		Email:          "asset-owner-2@example.com",
		Password:       &password,
		DefaultGroupID: secondGroup.ID,
		IsOwner:        true,
	})
	require.NoError(t, err)
	secondItem, err := tRepos.Entities.Create(ctx, secondGroup.ID, EntityCreate{
		Name:    "Second item",
		AssetID: assetID,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = tRepos.Entities.Delete(ctx, secondItem.ID)
		_ = tRepos.Entities.Delete(ctx, firstItem.ID)
		_ = tRepos.Users.Delete(ctx, secondOwner.ID)
		_ = tRepos.Users.Delete(ctx, firstOwner.ID)
		_ = tRepos.Groups.GroupDelete(ctx, secondGroup.ID)
		_ = tRepos.Groups.GroupDelete(ctx, firstGroup.ID)
	})

	_, err = tRepos.Entities.GetFoundEntityContactByAssetID(ctx, assetID)
	require.Error(t, err)
	require.True(t, ent.IsNotFound(err))
}

func TestEntityRepository_GetFoundEntityContact_NotFound(t *testing.T) {
	_, err := tRepos.Entities.GetFoundEntityContact(context.Background(), uuid.New())
	require.Error(t, err)
	require.True(t, ent.IsNotFound(err))
}
