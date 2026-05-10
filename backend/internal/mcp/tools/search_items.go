package tools

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/mcp"
)

const searchItemsDefaultPageSize = 25

type searchItemsInput struct {
	Query      string    `json:"query,omitempty"        jsonschema:"text matched against item names and descriptions; omit for an unfiltered list"`
	Page       int       `json:"page,omitempty"         jsonschema:"1-indexed page number; defaults to 1"`
	PageSize   int       `json:"page_size,omitempty"    jsonschema:"items per page; defaults to 25; max useful value is around 100"`
	IsLocation *bool     `json:"is_location,omitempty"  jsonschema:"true returns only locations, false returns only items, omit returns both"`
	GroupID    uuid.UUID `json:"group_id,omitempty"     jsonschema:"optional group to query; omit to use your default group; call list_my_groups to see available groups"`
}

type searchItemHit struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	Quantity      float64   `json:"quantity,omitempty"`
	Archived      bool      `json:"archived,omitempty"`
	PurchasePrice float64   `json:"purchase_price,omitempty"`
	ParentName    string    `json:"parent_name,omitempty"`
	EntityType    string    `json:"entity_type,omitempty"`
}

type searchItemsOutput struct {
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	Results  []searchItemHit `json:"results"`
}

func init() {
	mcp.Register(func(s *mcpsdk.Server, d mcp.Deps) {
		mcpsdk.AddTool(s,
			&mcpsdk.Tool{
				Name:        "search_items",
				Description: "Search the inventory for items (and optionally locations) belonging to the calling user's group. Returns a paginated list of matches with parent location and entity type included.",
			},
			func(ctx context.Context, _ *mcpsdk.CallToolRequest, in searchItemsInput) (*mcpsdk.CallToolResult, searchItemsOutput, error) {
				sctx, err := mcp.ResolveGroup(ctx, in.GroupID)
				if err != nil {
					return nil, searchItemsOutput{}, err
				}

				page := in.Page
				if page <= 0 {
					page = 1
				}
				pageSize := in.PageSize
				if pageSize <= 0 {
					pageSize = searchItemsDefaultPageSize
				}

				res, err := d.Repos.Entities.QueryByGroup(ctx, sctx.GID, repo.EntityQuery{
					Search:     in.Query,
					Page:       page,
					PageSize:   pageSize,
					IsLocation: in.IsLocation,
				})
				if err != nil {
					return nil, searchItemsOutput{}, fmt.Errorf("search items: %w", err)
				}

				out := searchItemsOutput{
					Total:    res.Total,
					Page:     res.Page,
					PageSize: res.PageSize,
					Results:  make([]searchItemHit, 0, len(res.Items)),
				}
				for _, it := range res.Items {
					hit := searchItemHit{
						ID:            it.ID,
						Name:          it.Name,
						Description:   it.Description,
						Quantity:      it.Quantity,
						Archived:      it.Archived,
						PurchasePrice: it.PurchasePrice,
					}
					if it.Parent != nil {
						hit.ParentName = it.Parent.Name
					}
					if it.EntityType != nil {
						hit.EntityType = it.EntityType.Name
					}
					out.Results = append(out.Results, hit)
				}
				return nil, out, nil
			})
	})
}
