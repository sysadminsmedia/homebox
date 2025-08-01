package repo

import (
	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
	"time"
)

type EntitiesRepository struct {
	db  *ent.Client
	bus *eventbus.EventBus
}

type (
	EntityField struct {
		ID    uuid.UUID `json:"id"`
		Name  string    `json:"name"`
		Type  string    `json:"type"`
		Value string    `json:"value"`
	}

	EntitySummary struct {
		// Basics
		ID          uuid.UUID `json:"id"`
		ImportRef   *string   `json:"import_ref,omitempty"`
		AssetID     AssetID   `json:"assetId,string"`
		Name        string    `json:"name"`
		Description *string   `json:"description"`
		Quantity    int       `json:"quantity"`
		Insured     bool      `json:"insured"`
		Archived    bool      `json:"archived"`

		// Some edges
		Location *EntitySummary `json:"location"`
		Labels   []LabelSummary `json:"labels"`
		ImageID  *uuid.UUID     `json:"imageId,omitempty"`

		// Additional Data
		PurchasePrice float64   `json:"purchasePrice"`
		SoldTime      time.Time `json:"soldTime"`
	}

	EntityOut struct {
		Parent *EntitySummary `json:"parent,omitempty" extensions:"x-nullable,x-omitempty"`
		EntitySummary

		SyncChildItemsLocations bool `json:"syncChildItemsLocations"`

		// Specific detail information
		SerialNumber string `json:"serialNumber"`
		ModelNumber  string `json:"modelNumber"`
		Manufacturer string `json:"manufacturer"`

		// Warranty
		LifetimeWarranty bool       `json:"lifetimeWarranty"`
		WarrantyExpires  types.Date `json:"warrantyExpires"`
		WarrantyDetails  string     `json:"warrantyDetails"`

		// Purchase
		PurchaseTime types.Date `json:"purchaseTime"`
		PurchaseFrom string     `json:"purchaseFrom"`

		// Sold
		SoldTime  types.Date `json:"soldTime"`
		SoldTo    string     `json:"soldTo"`
		SoldPrice float64    `json:"soldPrice"`
		SoldNotes string     `json:"soldNotes"`

		// Extras
		Notes string `json:"notes"`

		// Edges
		Attachments []ItemAttachment `json:"attachments,omitempty" extensions:"x-nullable,x-omitempty"`
		Fields      []EntityField    `json:"fields,omitempty" extensions:"x-nullable,x-omitempty"`
	}
)
