package tools

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sysadminsmedia/homebox/backend/internal/mcp"
)

type getItemInput struct {
	ID      uuid.UUID `json:"id"                 jsonschema:"UUID of the inventory entity (item or location) to fetch"`
	GroupID uuid.UUID `json:"group_id,omitempty" jsonschema:"optional group to query; omit to use your default group; call list_my_groups to see available groups"`
}

type getItemField struct {
	Name         string `json:"name"`
	Type         string `json:"type,omitempty"`
	TextValue    string `json:"text_value,omitempty"`
	NumberValue  int    `json:"number_value,omitempty"`
	BooleanValue bool   `json:"boolean_value,omitempty"`
}

type getItemAttachment struct {
	ID    uuid.UUID `json:"id"`
	Title string    `json:"title,omitempty"`
}

type getItemOutput struct {
	ID               uuid.UUID           `json:"id"`
	Name             string              `json:"name"`
	Description      string              `json:"description,omitempty"`
	Quantity         float64             `json:"quantity,omitempty"`
	Archived         bool                `json:"archived,omitempty"`
	AssetID          string              `json:"asset_id,omitempty"`
	ParentName       string              `json:"parent_name,omitempty"`
	ParentID         uuid.UUID           `json:"parent_id,omitempty"`
	EntityType       string              `json:"entity_type,omitempty"`
	SerialNumber     string              `json:"serial_number,omitempty"`
	ModelNumber      string              `json:"model_number,omitempty"`
	Manufacturer     string              `json:"manufacturer,omitempty"`
	LifetimeWarranty bool                `json:"lifetime_warranty,omitempty"`
	WarrantyExpires  string              `json:"warranty_expires,omitempty"`
	WarrantyDetails  string              `json:"warranty_details,omitempty"`
	PurchaseDate     string              `json:"purchase_date,omitempty"`
	PurchaseFrom     string              `json:"purchase_from,omitempty"`
	PurchasePrice    float64             `json:"purchase_price,omitempty"`
	SoldDate         string              `json:"sold_date,omitempty"`
	SoldTo           string              `json:"sold_to,omitempty"`
	SoldPrice        float64             `json:"sold_price,omitempty"`
	SoldNotes        string              `json:"sold_notes,omitempty"`
	Notes            string              `json:"notes,omitempty"`
	Tags             []string            `json:"tags,omitempty"`
	Fields           []getItemField      `json:"fields,omitempty"`
	Attachments      []getItemAttachment `json:"attachments,omitempty"`
	ChildrenCount    int                 `json:"children_count,omitempty"`
	TotalPrice       float64             `json:"total_price,omitempty"`
}

func init() {
	mcp.Register(func(s *mcpsdk.Server, d mcp.Deps) {
		mcpsdk.AddTool(s,
			&mcpsdk.Tool{
				Name:         "get_item",
				Description:  "Fetch full details of a single inventory entity (item or location) by its UUID. Returns parent location, entity type, tags, custom fields, attachments, warranty, purchase, and sold info.",
				InputSchema:  mcp.MustSchema[getItemInput](),
				OutputSchema: mcp.MustSchema[getItemOutput](),
			},
			func(ctx context.Context, _ *mcpsdk.CallToolRequest, in getItemInput) (*mcpsdk.CallToolResult, getItemOutput, error) {
				sctx, err := mcp.ResolveGroup(ctx, in.GroupID)
				if err != nil {
					return nil, getItemOutput{}, err
				}

				e, err := d.Repos.Entities.GetOneByGroup(ctx, sctx.GID, in.ID)
				if err != nil {
					return nil, getItemOutput{}, fmt.Errorf("get item: %w", err)
				}

				out := getItemOutput{
					ID:               e.ID,
					Name:             e.Name,
					Description:      e.Description,
					Quantity:         e.Quantity,
					Archived:         e.Archived,
					SerialNumber:     e.SerialNumber,
					ModelNumber:      e.ModelNumber,
					Manufacturer:     e.Manufacturer,
					LifetimeWarranty: e.LifetimeWarranty,
					WarrantyDetails:  e.WarrantyDetails,
					PurchaseFrom:     e.PurchaseFrom,
					PurchasePrice:    e.PurchasePrice,
					SoldTo:           e.SoldTo,
					SoldPrice:        e.SoldPrice,
					SoldNotes:        e.SoldNotes,
					Notes:            e.Notes,
					TotalPrice:       e.TotalPrice,
					ChildrenCount:    len(e.Children),
				}

				if !e.AssetID.Nil() {
					out.AssetID = e.AssetID.String()
				}

				if e.Parent != nil {
					out.ParentName = e.Parent.Name
					out.ParentID = e.Parent.ID
				}

				if e.EntityType != nil {
					out.EntityType = e.EntityType.Name
				}

				if s := e.WarrantyExpires.String(); s != "" {
					out.WarrantyExpires = s
				}
				if s := e.PurchaseDate.String(); s != "" {
					out.PurchaseDate = s
				}
				if s := e.SoldDate.String(); s != "" {
					out.SoldDate = s
				}

				if len(e.Tags) > 0 {
					tags := make([]string, 0, len(e.Tags))
					for _, t := range e.Tags {
						tags = append(tags, t.Name)
					}
					out.Tags = tags
				}

				if len(e.Fields) > 0 {
					fields := make([]getItemField, 0, len(e.Fields))
					for _, f := range e.Fields {
						fields = append(fields, getItemField{
							Name:         f.Name,
							Type:         f.Type,
							TextValue:    f.TextValue,
							NumberValue:  f.NumberValue,
							BooleanValue: f.BooleanValue,
						})
					}
					out.Fields = fields
				}

				if len(e.Attachments) > 0 {
					attachments := make([]getItemAttachment, 0, len(e.Attachments))
					for _, a := range e.Attachments {
						attachments = append(attachments, getItemAttachment{
							ID:    a.ID,
							Title: a.Title,
						})
					}
					out.Attachments = attachments
				}

				return nil, out, nil
			})
	})
}
