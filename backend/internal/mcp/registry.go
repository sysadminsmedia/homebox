package mcp

import (
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Registrar binds a single tool to a server using the supplied dependencies.
// It exists so tool files can stay strongly typed (each tool can declare its
// own input/output structs and call mcpsdk.AddTool[In,Out] directly) while
// still being collected into a uniform registry.
type Registrar func(s *mcpsdk.Server, d Deps)

var registry []Registrar

// Register adds a tool to the global registry. Call this from a tool file's
// init() so the tool is wired in for every server instance built by NewHandler.
func Register(r Registrar) {
	registry = append(registry, r)
}

func registerAll(s *mcpsdk.Server, d Deps) {
	for _, r := range registry {
		r(s, d)
	}
}
