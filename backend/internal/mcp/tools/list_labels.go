package tools

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sysadminsmedia/homebox/backend/internal/mcp"
)

type listLabelsInput struct {
	GroupID uuid.UUID `json:"group_id,omitempty" jsonschema:"optional group to query; omit to use your default group; call list_my_groups to see available groups"`
}

type listLabelsItem struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Color       string    `json:"color,omitempty"`
}

type listLabelsOutput struct {
	Total   int              `json:"total"`
	Results []listLabelsItem `json:"results"`
}

func init() {
	mcp.Register(func(s *mcpsdk.Server, d mcp.Deps) {
		mcpsdk.AddTool(s,
			&mcpsdk.Tool{
				Name:        "list_labels",
				Description: "List all labels (tags) defined in the calling user's group. Use the returned label IDs to filter `search_items` results.",
			},
			func(ctx context.Context, _ *mcpsdk.CallToolRequest, in listLabelsInput) (*mcpsdk.CallToolResult, listLabelsOutput, error) {
				sctx, err := mcp.ResolveGroup(ctx, in.GroupID)
				if err != nil {
					return nil, listLabelsOutput{}, err
				}

				tags, err := d.Repos.Tags.GetAll(ctx, sctx.GID)
				if err != nil {
					return nil, listLabelsOutput{}, fmt.Errorf("list labels: %w", err)
				}

				out := listLabelsOutput{
					Total:   len(tags),
					Results: make([]listLabelsItem, 0, len(tags)),
				}
				for _, t := range tags {
					out.Results = append(out.Results, listLabelsItem{
						ID:          t.ID,
						Name:        t.Name,
						Description: t.Description,
						Color:       t.Color,
					})
				}
				return nil, out, nil
			})
	})
}
