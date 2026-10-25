package repo

import (
	"context"

	"github.com/sysadminsmedia/homebox/backend/internal/data/authz"

	// Stitch schema policies, interceptors, and hooks into the ent client
	// used by the test suite.
	_ "github.com/sysadminsmedia/homebox/backend/internal/data/ent/runtime"
)

// testCtx returns a privacy-bypassing system context. The repo test suite
// validates business logic; ORM-level authorization has its own enforcement
// suite that builds real viewer contexts.
func testCtx() context.Context {
	return authz.NewSystemContext(context.Background())
}
