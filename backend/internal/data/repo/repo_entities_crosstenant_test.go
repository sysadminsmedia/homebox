package repo

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEntityRepository_UpdateByGroup_CrossTenantLeak guards against a cross-tenant
// IDOR in the entity update path. PUT /entities/{id} calls UpdateByGroup; the update
// itself is group-scoped (a no-op across tenants), but the record returned in the
// response must also be group-scoped. Returning the entity via an unscoped lookup
// leaks another tenant's full inventory record (name, notes, purchase/sale details)
// even though nothing is written.
func TestEntityRepository_UpdateByGroup_CrossTenantLeak(t *testing.T) {
	ctx := context.Background()

	// Victim tenant: a separate group with an entity holding sensitive data.
	victimGroup, err := tRepos.Groups.GroupCreate(ctx, "victim-group", uuid.Nil)
	require.NoError(t, err)

	victimET, err := tRepos.EntityTypes.GetDefault(ctx, victimGroup.ID, false)
	require.NoError(t, err)

	const secretName = "victim-secret-serial-and-notes"
	victim, err := tRepos.Entities.Create(ctx, victimGroup.ID, EntityCreate{
		Name:         secretName,
		Description:  "confidential cross-tenant data",
		EntityTypeID: victimET.ID,
	})
	require.NoError(t, err)

	// Attacker (tGroup) attempts a PUT against the victim's entity UUID with a
	// minimal body. gid is the attacker's group, ID is the victim's entity.
	out, err := tRepos.Entities.UpdateByGroup(ctx, tGroup.ID, EntityUpdate{
		ID:   victim.ID,
		Name: "poc-test",
	})

	// The attacker must NOT receive the victim's record back.
	require.Error(t, err, "cross-tenant update must be rejected, not leak the entity")
	assert.NotEqual(t, secretName, out.Name, "victim's data must not leak through the response")
	assert.Equal(t, uuid.Nil, out.ID, "no entity should be returned to the attacker")

	// And the victim's entity must be untouched.
	stillThere, err := tRepos.Entities.GetOneByGroup(ctx, victimGroup.ID, victim.ID)
	require.NoError(t, err)
	assert.Equal(t, secretName, stillThere.Name, "victim entity must be unmodified")
}

// TestEntityRepository_UpdateByGroup_CrossTenantFieldWrite guards the nested
// custom-field sync inside UpdateByGroup. The top-level record update is
// group-scoped and silently affects zero rows for a foreign entity ID, so
// execution used to fall through to the field block, which scoped its
// create/update/delete only by entity ID. An attacker could therefore wipe or
// inject fields on another tenant's entity — blindly, since the response is a
// 404 from the group-scoped read at the end.
func TestEntityRepository_UpdateByGroup_CrossTenantFieldWrite(t *testing.T) {
	ctx := context.Background()

	victimGroup, err := tRepos.Groups.GroupCreate(ctx, "victim-group-fields", uuid.Nil)
	require.NoError(t, err)

	victimET, err := tRepos.EntityTypes.GetDefault(ctx, victimGroup.ID, false)
	require.NoError(t, err)

	victim, err := tRepos.Entities.Create(ctx, victimGroup.ID, EntityCreate{
		Name:         "victim-entity-with-fields",
		EntityTypeID: victimET.ID,
	})
	require.NoError(t, err)

	// Seed the victim entity with a custom field, as its own tenant would.
	const (
		fieldName  = "license-key"
		fieldValue = "victim-secret-value"
	)
	seeded, err := tRepos.Entities.UpdateByGroup(ctx, victimGroup.ID, EntityUpdate{
		ID:           victim.ID,
		Name:         victim.Name,
		EntityTypeID: victimET.ID,
		Fields: []EntityFieldData{
			{Name: fieldName, Type: "text", TextValue: fieldValue},
		},
	})
	require.NoError(t, err)
	require.Len(t, seeded.Fields, 1)

	// Attacker (tGroup) submits an empty fields array against the victim's
	// entity UUID: the delete branch would remove every existing field.
	_, err = tRepos.Entities.UpdateByGroup(ctx, tGroup.ID, EntityUpdate{
		ID:     victim.ID,
		Name:   "poc-test",
		Fields: []EntityFieldData{},
	})
	require.Error(t, err, "cross-tenant update must be rejected")

	// Attacker attempts to inject a new field onto the victim's entity.
	_, err = tRepos.Entities.UpdateByGroup(ctx, tGroup.ID, EntityUpdate{
		ID:   victim.ID,
		Name: "poc-test",
		Fields: []EntityFieldData{
			{Name: "attacker-injected", Type: "text", TextValue: "pwned"},
		},
	})
	require.Error(t, err, "cross-tenant update must be rejected")

	// Attacker attempts to overwrite the known field by ID.
	_, err = tRepos.Entities.UpdateByGroup(ctx, tGroup.ID, EntityUpdate{
		ID:   victim.ID,
		Name: "poc-test",
		Fields: []EntityFieldData{
			{ID: seeded.Fields[0].ID, Name: fieldName, Type: "text", TextValue: "tampered"},
		},
	})
	require.Error(t, err, "cross-tenant update must be rejected")

	// The victim's fields must be exactly as seeded: not deleted, not added
	// to, and not modified.
	after, err := tRepos.Entities.GetOneByGroup(ctx, victimGroup.ID, victim.ID)
	require.NoError(t, err)
	require.Len(t, after.Fields, 1, "victim's custom fields must not be deleted or added to")
	assert.Equal(t, fieldName, after.Fields[0].Name)
	assert.Equal(t, fieldValue, after.Fields[0].TextValue, "victim's field value must not be tampered with")
}
