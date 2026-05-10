package tools

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/mcp"
)

const (
	getMaintenanceScheduleDefaultLimit = 50
	getMaintenanceScheduleMaxLimit     = 200
)

type getMaintenanceScheduleInput struct {
	Status  string    `json:"status,omitempty"   jsonschema:"scheduled | completed | both; defaults to scheduled"`
	Limit   int       `json:"limit,omitempty"    jsonschema:"max number of entries to return; defaults to 50; capped at 200"`
	GroupID uuid.UUID `json:"group_id,omitempty" jsonschema:"optional group to query; omit to use your default group; call list_my_groups to see available groups"`
}

type getMaintenanceScheduleItem struct {
	ID            uuid.UUID `json:"id"`
	ItemID        uuid.UUID `json:"item_id"`
	ItemName      string    `json:"item_name,omitempty"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	ScheduledDate string    `json:"scheduled_date,omitempty"`
	CompletedDate string    `json:"completed_date,omitempty"`
	Cost          float64   `json:"cost,omitempty"`
}

type getMaintenanceScheduleOutput struct {
	Status  string                       `json:"status"`
	Total   int                          `json:"total"`
	Results []getMaintenanceScheduleItem `json:"results"`
}

func init() {
	mcp.Register(func(s *mcpsdk.Server, d mcp.Deps) {
		mcpsdk.AddTool(s,
			&mcpsdk.Tool{
				Name:        "get_maintenance_schedule",
				Description: "List maintenance entries across the calling user's group. Filter by status (scheduled, completed, or both) to find upcoming work or review history.",
			},
			func(ctx context.Context, _ *mcpsdk.CallToolRequest, in getMaintenanceScheduleInput) (*mcpsdk.CallToolResult, getMaintenanceScheduleOutput, error) {
				sctx, err := mcp.ResolveGroup(ctx, in.GroupID)
				if err != nil {
					return nil, getMaintenanceScheduleOutput{}, err
				}

				var status repo.MaintenanceFilterStatus
				switch in.Status {
				case "", string(repo.MaintenanceFilterStatusScheduled):
					status = repo.MaintenanceFilterStatusScheduled
				case string(repo.MaintenanceFilterStatusCompleted):
					status = repo.MaintenanceFilterStatusCompleted
				case string(repo.MaintenanceFilterStatusBoth):
					status = repo.MaintenanceFilterStatusBoth
				default:
					return nil, getMaintenanceScheduleOutput{}, fmt.Errorf("invalid status %q: must be one of scheduled, completed, both", in.Status)
				}

				limit := in.Limit
				if limit <= 0 {
					limit = getMaintenanceScheduleDefaultLimit
				}
				if limit > getMaintenanceScheduleMaxLimit {
					limit = getMaintenanceScheduleMaxLimit
				}

				entries, err := d.Repos.MaintEntry.GetAllMaintenance(ctx, sctx.GID, repo.MaintenanceFilters{
					Status: status,
				})
				if err != nil {
					return nil, getMaintenanceScheduleOutput{}, fmt.Errorf("get maintenance schedule: %w", err)
				}

				if len(entries) > limit {
					entries = entries[:limit]
				}

				out := getMaintenanceScheduleOutput{
					Status:  string(status),
					Total:   len(entries),
					Results: make([]getMaintenanceScheduleItem, 0, len(entries)),
				}
				for _, e := range entries {
					item := getMaintenanceScheduleItem{
						ID:            e.ID,
						ItemID:        e.ItemID,
						ItemName:      e.ItemName,
						Name:          e.Name,
						Description:   e.Description,
						ScheduledDate: e.ScheduledDate.String(),
						CompletedDate: e.CompletedDate.String(),
						Cost:          e.Cost,
					}
					out.Results = append(out.Results, item)
				}
				return nil, out, nil
			})
	})
}
