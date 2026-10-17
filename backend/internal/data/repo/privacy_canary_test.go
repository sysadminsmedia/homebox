package repo

import (
	"context"
	"errors"
	"testing"

	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/privacy"
)

// TestPrivacyDeniesWithoutViewer is the wiring canary for the ent privacy
// layer: a context with neither a viewer nor a system marker must be denied
// at the ORM. If this test fails, policy registration broke (most likely a
// missing ent/runtime import) and the whole authorization layer is offline.
func TestPrivacyDeniesWithoutViewer(t *testing.T) {
	bare := context.Background()

	if _, err := tClient.Entity.Query().All(bare); err == nil {
		t.Fatal("entity query without viewer must be denied")
	} else if !errors.Is(err, privacy.Deny) {
		t.Fatalf("expected a privacy deny, got: %v", err)
	}

	if _, err := tClient.Tag.Query().All(bare); err == nil {
		t.Fatal("tag query without viewer must be denied")
	}

	if _, err := tClient.Group.Create().SetName("nope").Save(bare); err == nil {
		t.Fatal("group create without viewer must be denied")
	}

	if err := tClient.Entity.DeleteOneID(tGroup.ID).Exec(bare); err == nil {
		t.Fatal("entity delete without viewer must be denied")
	}
}
