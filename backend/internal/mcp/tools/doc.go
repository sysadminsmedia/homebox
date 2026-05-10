// Package tools holds the read tools exposed by the Homebox MCP server.
//
// Each tool lives in its own file and self-registers via init() by calling
// mcp.Register. To add a new tool: create a new file here, define typed
// input/output structs, and register the tool with mcpsdk.AddTool inside a
// Registrar callback. Avoid accepting a user_id / group_id in tool input —
// always derive them from mcp.ServiceCtx(ctx).
//
// Importing this package (typically as a blank import from the route wiring)
// fires every file's init() and populates the registry.
package tools
