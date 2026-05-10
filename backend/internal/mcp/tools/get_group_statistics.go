package tools

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sysadminsmedia/homebox/backend/internal/mcp"
)

type getGroupStatisticsInput struct {
	GroupID uuid.UUID `json:"group_id,omitempty" jsonschema:"optional group to query; omit to use your default group; call list_my_groups to see available groups"`
}

type getGroupStatisticsOutput struct {
	TotalUsers        int     `json:"total_users"`
	TotalItems        int     `json:"total_items"`
	TotalLocations    int     `json:"total_locations"`
	TotalTags         int     `json:"total_tags"`
	TotalItemPrice    float64 `json:"total_item_price"`
	TotalWithWarranty int     `json:"total_with_warranty"`
}

func init() {
	mcp.Register(func(s *mcpsdk.Server, d mcp.Deps) {
		mcpsdk.AddTool(s,
			&mcpsdk.Tool{
				Name:         "get_group_statistics",
				Description:  "Return high-level inventory statistics for the calling user's group: counts of items, locations, labels, total purchase value, etc.",
				InputSchema:  mcp.MustSchema[getGroupStatisticsInput](),
				OutputSchema: mcp.MustSchema[getGroupStatisticsOutput](),
			},
			func(ctx context.Context, _ *mcpsdk.CallToolRequest, in getGroupStatisticsInput) (*mcpsdk.CallToolResult, getGroupStatisticsOutput, error) {
				sctx, err := mcp.ResolveGroup(ctx, in.GroupID)
				if err != nil {
					return nil, getGroupStatisticsOutput{}, err
				}

				stats, err := d.Repos.Groups.StatsGroup(ctx, sctx.GID)
				if err != nil {
					return nil, getGroupStatisticsOutput{}, fmt.Errorf("get group statistics: %w", err)
				}

				out := getGroupStatisticsOutput{
					TotalUsers:        stats.TotalUsers,
					TotalItems:        stats.TotalItems,
					TotalLocations:    stats.TotalLocations,
					TotalTags:         stats.TotalTags,
					TotalItemPrice:    stats.TotalItemPrice,
					TotalWithWarranty: stats.TotalWithWarranty,
				}
				return nil, out, nil
			})
	})
}
