package tools

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sysadminsmedia/homebox/backend/internal/mcp"
)

type getEntityPathInput struct {
	ID      uuid.UUID `json:"id"                 jsonschema:"the entity ID to resolve"`
	GroupID uuid.UUID `json:"group_id,omitempty" jsonschema:"optional group to query; omit to use your default group; call list_my_groups to see available groups"`
}

type getEntityPathSegment struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Type string    `json:"type"`
}

type getEntityPathOutput struct {
	Path []getEntityPathSegment `json:"path"`
}

func init() {
	mcp.Register(func(s *mcpsdk.Server, d mcp.Deps) {
		mcpsdk.AddTool(s,
			&mcpsdk.Tool{
				Name:         "get_entity_path",
				Description:  "Return the full breadcrumb path from the root location down to the given entity. Useful for answering 'where is X?' — the last element of the path is the entity itself; everything before it is its containing chain.",
				InputSchema:  mcp.MustSchema[getEntityPathInput](),
				OutputSchema: mcp.MustSchema[getEntityPathOutput](),
			},
			func(ctx context.Context, _ *mcpsdk.CallToolRequest, in getEntityPathInput) (*mcpsdk.CallToolResult, getEntityPathOutput, error) {
				sctx, err := mcp.ResolveGroup(ctx, in.GroupID)
				if err != nil {
					return nil, getEntityPathOutput{}, err
				}

				path, err := d.Repos.Entities.PathForEntity(ctx, sctx.GID, in.ID)
				if err != nil {
					return nil, getEntityPathOutput{}, fmt.Errorf("get entity path: %w", err)
				}

				out := getEntityPathOutput{
					Path: make([]getEntityPathSegment, 0, len(path)),
				}
				for _, p := range path {
					out.Path = append(out.Path, getEntityPathSegment{
						ID:   p.ID,
						Name: p.Name,
						Type: string(p.Type),
					})
				}
				return nil, out, nil
			})
	})
}
