package authz

// GrantAction is a single row-level capability on one entity. Grants are
// stored as boolean columns on access_grants so they can be used in SQL
// predicates; the string form is the API surface.
type GrantAction string

const (
	GrantRead        GrantAction = "read"
	GrantUpdate      GrantAction = "update"
	GrantDelete      GrantAction = "delete"
	GrantAttachments GrantAction = "attachments"
)

var allGrantActions = []GrantAction{GrantRead, GrantUpdate, GrantDelete, GrantAttachments}

// AllGrantActions returns the full grant-action catalog.
func AllGrantActions() []GrantAction {
	out := make([]GrantAction, len(allGrantActions))
	copy(out, allGrantActions)
	return out
}

// ValidGrantAction reports whether a is a known grant action.
func ValidGrantAction(a GrantAction) bool {
	for _, known := range allGrantActions {
		if a == known {
			return true
		}
	}
	return false
}

// GrantActions is the decoded form of an access grant's boolean columns.
// Update, Delete, and Attachments all imply Read; Normalize enforces that.
type GrantActions struct {
	Read        bool
	Update      bool
	Delete      bool
	Attachments bool
}

// GrantActionsFromStrings parses API action strings. The second return value
// lists unknown actions; the result is normalized.
func GrantActionsFromStrings(actions []string) (GrantActions, []string) {
	var ga GrantActions
	var invalid []string
	for _, a := range actions {
		switch GrantAction(a) {
		case GrantRead:
			ga.Read = true
		case GrantUpdate:
			ga.Update = true
		case GrantDelete:
			ga.Delete = true
		case GrantAttachments:
			ga.Attachments = true
		default:
			invalid = append(invalid, a)
		}
	}
	ga.Normalize()
	return ga, invalid
}

// Normalize forces Read on when any higher action is present.
func (g *GrantActions) Normalize() {
	if g.Update || g.Delete || g.Attachments {
		g.Read = true
	}
}

// Any reports whether at least one action is granted.
func (g GrantActions) Any() bool {
	return g.Read || g.Update || g.Delete || g.Attachments
}

// Strings returns the API string form of the granted actions.
func (g GrantActions) Strings() []string {
	out := make([]string, 0, 4)
	if g.Read {
		out = append(out, string(GrantRead))
	}
	if g.Update {
		out = append(out, string(GrantUpdate))
	}
	if g.Delete {
		out = append(out, string(GrantDelete))
	}
	if g.Attachments {
		out = append(out, string(GrantAttachments))
	}
	return out
}
