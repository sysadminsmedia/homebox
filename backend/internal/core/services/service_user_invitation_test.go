package services

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/usergroup"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/pkgs/hasher"
)

// newInvitation creates a fresh group with an invitation of the given uses and
// returns the group ID, invitation ID and raw token.
func newInvitation(t *testing.T, uses int) (uuid.UUID, uuid.UUID, string) {
	t.Helper()
	ctx := context.Background()

	g, err := tRepos.Groups.GroupCreate(ctx, "invite-"+fk.Str(8), uuid.Nil)
	require.NoError(t, err)

	raw := hasher.GenerateToken()
	inv, err := tRepos.Groups.InvitationCreate(ctx, g.ID, repo.GroupInvitationCreate{
		Token:     raw.Hash,
		ExpiresAt: time.Now().Add(time.Hour),
		Uses:      uses,
	})
	require.NoError(t, err)

	return g.ID, inv.ID, raw.Raw
}

func memberCount(t *testing.T, gid uuid.UUID) int {
	t.Helper()
	n, err := tClient.UserGroup.Query().Where(usergroup.GroupID(gid)).Count(context.Background())
	require.NoError(t, err)
	return n
}

func TestRegisterUser_InvitationNotOversubscribedByConcurrentRegistrations(t *testing.T) {
	gid, invID, token := newInvitation(t, 1)

	const racers = 6
	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		successes int
	)
	start := make(chan struct{})
	for i := 0; i < racers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := tSvc.User.RegisterUser(context.Background(), UserRegistration{
				GroupToken: token,
				Name:       fk.Str(10),
				Email:      fk.Email(),
				Password:   fk.Str(16),
			})
			if err == nil {
				mu.Lock()
				successes++
				mu.Unlock()
			}
		}()
	}
	close(start)
	wg.Wait()

	assert.Equal(t, 1, successes, "a uses=1 invitation must admit exactly one registration")
	assert.Equal(t, 1, memberCount(t, gid), "only one membership may be created")

	invs, err := tRepos.Groups.InvitationGetAll(context.Background(), gid)
	require.NoError(t, err)
	for _, inv := range invs {
		if inv.ID == invID {
			assert.Equal(t, 0, inv.Uses)
		}
	}
}

func TestRegisterUser_FailedRegistrationRefundsInvitationUse(t *testing.T) {
	gid, invID, token := newInvitation(t, 1)

	// tUser's email already exists, so user creation fails after the use is claimed.
	_, err := tSvc.User.RegisterUser(context.Background(), UserRegistration{
		GroupToken: token,
		Name:       fk.Str(10),
		Email:      tUser.Email,
		Password:   fk.Str(16),
	})
	require.Error(t, err)

	invs, err := tRepos.Groups.InvitationGetAll(context.Background(), gid)
	require.NoError(t, err)
	found := false
	for _, inv := range invs {
		if inv.ID == invID {
			found = true
			assert.Equal(t, 1, inv.Uses, "the claimed use must be refunded when registration fails")
		}
	}
	require.True(t, found)

	// The refunded use is still redeemable.
	_, err = tSvc.User.RegisterUser(context.Background(), UserRegistration{
		GroupToken: token,
		Name:       fk.Str(10),
		Email:      fk.Email(),
		Password:   fk.Str(16),
	})
	require.NoError(t, err)
	assert.Equal(t, 1, memberCount(t, gid))
}
