package mcp

import (
	"net/http"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ServerName and ServerVersion identify this MCP server to connecting clients.
// The version is intentionally tied to the MCP integration shape, not the
// Homebox release version — bump it when tools are added or their schemas
// change in a breaking way.
const (
	ServerName    = "homebox"
	ServerVersion = "0.1.0"
)

// NewHandler builds an http.Handler that speaks the MCP Streamable HTTP
// transport. The returned handler should be mounted behind the standard
// auth/tenant middleware so each tool invocation carries the calling user's
// services.Context.
//
// One *mcpsdk.Server is built up-front with every registered tool. The same
// server is reused across requests; per-request data flows through the
// request context, which the handler propagates into each tool call.
func NewHandler(d Deps) http.Handler {
	server := mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:    ServerName,
		Version: ServerVersion,
	}, nil)

	registerAll(server, d)

	return mcpsdk.NewStreamableHTTPHandler(
		func(*http.Request) *mcpsdk.Server { return server },
		nil,
	)
}
