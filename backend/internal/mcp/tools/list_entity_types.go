package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sysadminsmedia/homebox/backend/internal/mcp"
)

type listEntityTypesInput struct {
	GroupID uuid.UUID `json:"group_id,omitempty" jsonschema:"optional group to query; omit to use your default group; call list_my_groups to see available groups"`
}

type listEntityTypesItem struct {
	ID                uuid.UUID  `json:"id"`
	Name              string     `json:"name"`
	Description       string     `json:"description,omitempty"`
	IsLocation        bool       `json:"is_location"`
	Icon              string     `json:"icon,omitempty"`
	DefaultTemplateID *uuid.UUID `json:"default_template_id,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type listEntityTypesOutput struct {
	ItemTypes     []listEntityTypesItem `json:"item_types"`
	LocationTypes []listEntityTypesItem `json:"location_types"`
}

func init() {
	mcp.Register(func(s *mcpsdk.Server, d mcp.Deps) {
		mcpsdk.AddTool(s,
			&mcpsdk.Tool{
				Name:         "list_entity_types",
				Description:  "List all entity types (categories) defined in the calling user's group. Each type indicates whether entities of that type are locations or items.",
				InputSchema:  mcp.MustSchema[listEntityTypesInput](),
				OutputSchema: mcp.MustSchema[listEntityTypesOutput](),
			},
			func(ctx context.Context, _ *mcpsdk.CallToolRequest, in listEntityTypesInput) (*mcpsdk.CallToolResult, listEntityTypesOutput, error) {
				sctx, err := mcp.ResolveGroup(ctx, in.GroupID)
				if err != nil {
					return nil, listEntityTypesOutput{}, err
				}

				types, err := d.Repos.EntityTypes.GetAll(ctx, sctx.GID)
				if err != nil {
					return nil, listEntityTypesOutput{}, fmt.Errorf("list entity types: %w", err)
				}

				out := listEntityTypesOutput{
					ItemTypes:     make([]listEntityTypesItem, 0, len(types)),
					LocationTypes: make([]listEntityTypesItem, 0, len(types)),
				}
				for _, et := range types {
					item := listEntityTypesItem{
						ID:                et.ID,
						Name:              et.Name,
						Description:       et.Description,
						IsLocation:        et.IsLocation,
						Icon:              et.Icon,
						DefaultTemplateID: et.DefaultTemplateID,
						CreatedAt:         et.CreatedAt,
						UpdatedAt:         et.UpdatedAt,
					}
					if et.IsLocation {
						out.LocationTypes = append(out.LocationTypes, item)
					} else {
						out.ItemTypes = append(out.ItemTypes, item)
					}
				}
				return nil, out, nil
			})
	})
}
