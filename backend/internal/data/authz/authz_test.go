package authz

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestValid(t *testing.T) {
	for _, p := range All() {
		if !Valid(p) {
			t.Errorf("catalog permission %q reported invalid", p)
		}
	}
	if Valid("entity:fly") {
		t.Error("unknown permission reported valid")
	}
}

func TestValidateStrings(t *testing.T) {
	invalid := ValidateStrings([]string{"entity:read", "bogus", "tag:manage", "also-bogus"})
	if len(invalid) != 2 || invalid[0] != "bogus" || invalid[1] != "also-bogus" {
		t.Errorf("unexpected invalid set: %v", invalid)
	}
	if got := ValidateStrings(AllStrings()); len(got) != 0 {
		t.Errorf("full catalog reported invalid entries: %v", got)
	}
}

func TestViewerHas(t *testing.T) {
	v := NewViewer(uuid.New(), uuid.New(), false, []string{"entity:read", "nonsense"}, nil)
	if !v.Has(PermEntityRead) {
		t.Error("expected entity:read")
	}
	if v.Has(PermEntityDelete) {
		t.Error("unexpected entity:delete")
	}
	if v.Has("nonsense") {
		t.Error("invalid permission should not be stored")
	}

	var nilViewer *Viewer
	if nilViewer.Has(PermEntityRead) {
		t.Error("nil viewer must hold nothing")
	}
}

func TestViewerSuperuser(t *testing.T) {
	v := NewViewer(uuid.New(), uuid.New(), true, nil, nil)
	for _, p := range All() {
		if !v.Has(p) {
			t.Errorf("superuser missing %q", p)
		}
	}
	if got := v.PermStrings(); len(got) != len(All()) {
		t.Errorf("superuser PermStrings = %v", got)
	}
}

func TestViewerContext(t *testing.T) {
	ctx := context.Background()
	if FromContext(ctx) != nil {
		t.Error("expected nil viewer on empty context")
	}
	v := NewViewer(uuid.New(), uuid.New(), false, []string{"entity:read"}, nil)
	if got := FromContext(NewContext(ctx, v)); got != v {
		t.Error("viewer roundtrip failed")
	}
}

func TestSystemContext(t *testing.T) {
	ctx := context.Background()
	if IsSystem(ctx) {
		t.Error("plain context must not be system")
	}
	if !IsSystem(NewSystemContext(ctx)) {
		t.Error("system context not detected")
	}
}

func TestGrantActions(t *testing.T) {
	ga, invalid := GrantActionsFromStrings([]string{"update", "bogus"})
	if len(invalid) != 1 || invalid[0] != "bogus" {
		t.Errorf("unexpected invalid actions: %v", invalid)
	}
	if !ga.Read {
		t.Error("update must imply read")
	}
	if !ga.Update || ga.Delete || ga.Attachments {
		t.Errorf("unexpected actions: %+v", ga)
	}
	if got := ga.Strings(); len(got) != 2 || got[0] != "read" || got[1] != "update" {
		t.Errorf("unexpected Strings(): %v", got)
	}

	empty, _ := GrantActionsFromStrings(nil)
	if empty.Any() {
		t.Error("empty actions must not report Any")
	}
}

func TestWildcards(t *testing.T) {
	if !Valid(Wildcard) {
		t.Error("* must be valid")
	}
	if !Valid("entity:*") {
		t.Error("entity:* must be valid")
	}
	if Valid("bogus:*") {
		t.Error("wildcard for unknown resource must be invalid")
	}

	if !SetHas([]string{"*"}, PermSettingsManage) {
		t.Error("* must cover settings:manage")
	}
	if !SetHas([]string{"entity:*"}, PermEntityDelete) {
		t.Error("entity:* must cover entity:delete")
	}
	if SetHas([]string{"entity:*"}, PermTagManage) {
		t.Error("entity:* must not cover tag:manage")
	}

	if got := Expand([]string{"*"}); len(got) != len(All()) {
		t.Errorf("Expand(*) = %d perms, want full catalog (%d)", len(got), len(All()))
	}
	if got := Expand([]string{"entity:*", "tag:manage"}); len(got) != 5 {
		t.Errorf("Expand(entity:*, tag:manage) = %v, want 4 entity perms + tag:manage", got)
	}
}

func TestViewerWildcardExpansion(t *testing.T) {
	v := NewViewer(uuid.New(), uuid.New(), false, []string{"*"}, nil)
	for _, p := range All() {
		if !v.Has(p) {
			t.Errorf("wildcard viewer missing %q", p)
		}
	}

	v = NewViewer(uuid.New(), uuid.New(), false, []string{"entity:*"}, nil)
	if !v.Has(PermEntityUpdate) || v.Has(PermSettingsManage) {
		t.Error("resource wildcard expansion incorrect")
	}
}

func TestFullAccessIsWildcard(t *testing.T) {
	fa := FullAccess()
	if len(fa) != 1 || fa[0] != Wildcard {
		t.Errorf("FullAccess() = %v, want [\"*\"]: stored full access must be the wildcard, not an enumerated snapshot", fa)
	}
}
