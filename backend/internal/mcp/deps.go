// Package mcp wires the optional Model Context Protocol server that lets
// MCP-aware clients (for example Claude Desktop) query the user's inventory
// over HTTP. It is mounted at /api/v1/mcp behind the standard auth/tenant
// middleware, so tool invocations inherit the calling user's group scope.
//
// Tools live under internal/mcp/tools and self-register at init time via
// Register. Adding a new tool is purely additive: drop a new file in tools/
// that imports this package and calls Register from its init function.
package mcp

import (
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
)

// Deps is the dependency bundle handed to each tool registration when the
// MCP server is constructed. Tool handlers close over this and pull the
// per-request user/group from context — never from tool input.
type Deps struct {
	Services *services.AllServices
	Repos    *repo.AllRepos
}
