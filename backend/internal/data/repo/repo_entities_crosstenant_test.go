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
