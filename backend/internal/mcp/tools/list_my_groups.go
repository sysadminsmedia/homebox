package tools

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sysadminsmedia/homebox/backend/internal/mcp"
)

type listMyGroupsInput struct{}

type listMyGroupsItem struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Currency  string    `json:"currency,omitempty"`
	IsDefault bool      `json:"is_default,omitempty"`
}

type listMyGroupsOutput struct {
	Groups []listMyGroupsItem `json:"groups"`
}

func init() {
	mcp.Register(func(s *mcpsdk.Server, d mcp.Deps) {
		mcpsdk.AddTool(s,
			&mcpsdk.Tool{
				Name:         "list_my_groups",
				Description:  "List the groups the calling user is a member of. The returned IDs can be passed as the optional `group_id` argument on other tools to scope queries to a specific group; omit `group_id` to use the user's default group.",
				InputSchema:  mcp.MustSchema[listMyGroupsInput](),
				OutputSchema: mcp.MustSchema[listMyGroupsOutput](),
			},
			func(ctx context.Context, _ *mcpsdk.CallToolRequest, _ listMyGroupsInput) (*mcpsdk.CallToolResult, listMyGroupsOutput, error) {
				sctx, err := mcp.ServiceCtx(ctx)
				if err != nil {
					return nil, listMyGroupsOutput{}, err
				}

				groups, err := d.Repos.Groups.GetAllGroups(ctx, sctx.UID)
				if err != nil {
					return nil, listMyGroupsOutput{}, fmt.Errorf("list groups: %w", err)
				}

				out := listMyGroupsOutput{
					Groups: make([]listMyGroupsItem, 0, len(groups)),
				}
				defaultGID := sctx.User.DefaultGroupID
				for _, g := range groups {
					out.Groups = append(out.Groups, listMyGroupsItem{
						ID:        g.ID,
						Name:      g.Name,
						Currency:  g.Currency,
						IsDefault: g.ID == defaultGID,
					})
				}
				return nil, out, nil
			})
	})
}
