// Package customsinventory turns plain item data into a verified customs
// inventory report and writes it as an .xlsx workbook, optionally zipped
// together with the item receipts.
//
// The package does not touch the database. The caller loads the items of one
// group, maps them to Item, and passes them to Build.
package customsinventory

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
)

// MaxScope is the largest number of entities a single export may cover.
const MaxScope = 5000

// DefaultMinMonths is the minimum ownership period used when Options.MinMonths is 0.
const DefaultMinMonths = 6

// MaxMinMonths is the largest allowed minimum ownership period.
const MaxMinMonths = 60

var (
	// DefaultReceiptTypes are the attachment types counted as receipts when none are given.
	DefaultReceiptTypes = []string{"receipt"}
	// AllowedReceiptTypes are the attachment types that may be counted as receipts.
	AllowedReceiptTypes = []string{"receipt", "attachment", "warranty", "manual"}
)

var (
	ErrInvalidRequest = errors.New("customs inventory: invalid request")
	ErrScopeTooLarge  = errors.New("customs inventory: scope too large")
)

type GroupBy string

const (
	GroupByLocation   GroupBy = "location"
	GroupByEntityType GroupBy = "entityType"
	GroupByTag        GroupBy = "tag"
)

type Status string

const (
	StatusDutyFree Status = "dutyFree"
	StatusDutiable Status = "dutiable"
)

type Reason string

const (
	ReasonDate                    Reason = "date"
	ReasonTagDutyFree             Reason = "tag_duty_free"
	ReasonOwnedTooShort           Reason = "owned_too_short"
	ReasonPurchasedAfterReference Reason = "purchased_after_reference"
	ReasonMissingPurchaseDate     Reason = "missing_purchase_date"
	ReasonTagDutiable             Reason = "tag_dutiable"
)

type IssueCode string

const (
	IssueMissingPurchaseDate     IssueCode = "missing_purchase_date"
	IssueConflictDateTag         IssueCode = "conflict_date_tag"
	IssueConflictTags            IssueCode = "conflict_tags"
	IssuePurchasedAfterReference IssueCode = "purchased_after_reference"
	IssueMissingPrice            IssueCode = "missing_price"
	IssueNegativeValue           IssueCode = "negative_value"
	IssueZeroQuantity            IssueCode = "zero_quantity"
	IssueMissingAssetID          IssueCode = "missing_asset_id"
	IssueDuplicateAssetID        IssueCode = "duplicate_asset_id"
	IssueDutiableWithoutReceipt  IssueCode = "dutiable_without_receipt"
	IssueReceiptIsLink           IssueCode = "receipt_is_link"
	IssueContentSplit            IssueCode = "content_split"
	IssueContainerQuantity       IssueCode = "container_quantity"
	IssueExcludedLocation        IssueCode = "excluded_location"
	IssueExcludedArchived        IssueCode = "excluded_archived"
	IssueExcludedSold            IssueCode = "excluded_sold"
	IssueExcludedNotFound        IssueCode = "excluded_not_found"
)

// IssueOrder is the order in which issues are reported.
var IssueOrder = []IssueCode{
	IssueMissingPurchaseDate,
	IssueConflictDateTag,
	IssueConflictTags,
	IssuePurchasedAfterReference,
	IssueMissingPrice,
	IssueNegativeValue,
	IssueZeroQuantity,
	IssueMissingAssetID,
	IssueDuplicateAssetID,
	IssueDutiableWithoutReceipt,
	IssueReceiptIsLink,
	IssueContentSplit,
	IssueContainerQuantity,
	IssueExcludedLocation,
	IssueExcludedArchived,
	IssueExcludedSold,
	IssueExcludedNotFound,
}

type Severity string

const (
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// SeverityOf returns the severity of an issue code: info for excluded
// entities and for containers whose contents are listed separately, warning
// for everything else.
func SeverityOf(c IssueCode) Severity {
	switch c {
	case IssueContainerQuantity, IssueExcludedLocation, IssueExcludedArchived, IssueExcludedSold, IssueExcludedNotFound:
		return SeverityInfo
	default:
		return SeverityWarning
	}
}

// --- API contract (JSON) ---

type Request struct {
	Scope   Scope   `json:"scope"`
	Options Options `json:"options"`
	Header  Header  `json:"header"`
}

// Scope is either a list of item IDs or a search query, never both.
type Scope struct {
	ItemIDs []uuid.UUID `json:"itemIds"`
	Query   *ScopeQuery `json:"query,omitempty" extensions:"x-nullable"`
}

// ScopeQuery mirrors the filters of GET /v1/entities, without paging.
type ScopeQuery struct {
	Q                string      `json:"q"`
	ParentIDs        []uuid.UUID `json:"parentIds"`
	EntityTypeIDs    []uuid.UUID `json:"entityTypeIds"`
	TagIDs           []uuid.UUID `json:"tags"`
	NegateTags       bool        `json:"negateTags"`
	MatchAllTags     bool        `json:"matchAllTags"`
	OnlyWithPhoto    bool        `json:"onlyWithPhoto"`
	OnlyWithoutPhoto bool        `json:"onlyWithoutPhoto"`
	IncludeArchived  bool        `json:"includeArchived"`
	Fields           []string    `json:"fields"` // "name=value"
}

type Options struct {
	GroupBy         GroupBy     `json:"groupBy"`
	GroupTagIDs     []uuid.UUID `json:"groupTagIds"`
	DutyFreeTagIDs  []uuid.UUID `json:"dutyFreeTagIds"`
	DutiableTagIDs  []uuid.UUID `json:"dutiableTagIds"`
	ReferenceDate   string      `json:"referenceDate"` // YYYY-MM-DD
	MinMonths       int         `json:"minMonths"`     // 0 → DefaultMinMonths; allowed 1..MaxMinMonths
	MergeIdentical  bool        `json:"mergeIdentical"`
	MergeFirstWord  bool        `json:"mergeFirstWord"` // only with MergeIdentical: match on the first word of the name
	RollUpContents  bool        `json:"rollUpContents"` // load the contents of the scope items and roll them into their containers
	IncludeReceipts bool        `json:"includeReceipts"`
	ReceiptTypes    []string    `json:"receiptTypes"`
}

type Header struct {
	Title   string `json:"title"   validate:"max=200"`
	Name    string `json:"name"    validate:"max=200"`
	Address string `json:"address" validate:"max=1000"` // newline-separated
	Phone   string `json:"phone"   validate:"max=100"`
	Email   string `json:"email"   validate:"max=200"`
}

// Validate checks the request. Every error wraps ErrInvalidRequest.
func (r Request) Validate() error {
	invalid := func(format string, args ...any) error {
		return fmt.Errorf("%w: %s", ErrInvalidRequest, fmt.Sprintf(format, args...))
	}

	hasIDs, hasQuery := len(r.Scope.ItemIDs) > 0, r.Scope.Query != nil
	switch {
	case hasIDs && hasQuery:
		return invalid("scope must have either item IDs or a query, not both")
	case !hasIDs && !hasQuery:
		return invalid("scope must have item IDs or a query")
	}

	o := r.Options
	if _, err := ParseDate(o.ReferenceDate); err != nil {
		return invalid("reference date must be YYYY-MM-DD")
	}
	if o.MinMonths < 0 || o.MinMonths > MaxMinMonths {
		return invalid("minimum months must be between 1 and %d", MaxMinMonths)
	}

	switch o.GroupBy {
	case GroupByLocation, GroupByEntityType:
	case GroupByTag:
		if len(o.GroupTagIDs) == 0 {
			return invalid("tag grouping needs at least one group tag")
		}
	default:
		return invalid("unknown grouping %q", o.GroupBy)
	}

	for _, t := range o.ReceiptTypes {
		if !slices.Contains(AllowedReceiptTypes, t) {
			return invalid("attachment type %q can't be a receipt", t)
		}
	}

	return nil
}

type Preview struct {
	Currency           string  `json:"currency"`
	ReferenceDate      string  `json:"referenceDate"`
	MinMonths          int     `json:"minMonths"` // after applying the default
	Cutoff             string  `json:"cutoff"`
	ItemCount          int     `json:"itemCount"`
	LineCount          int     `json:"lineCount"`
	DutiableItemCount  int     `json:"dutiableItemCount"`
	PieceCount         float64 `json:"pieceCount"`         // sum of all line quantities
	ContentCount       int     `json:"contentCount"`       // entities absorbed into a container line
	DutiablePieceCount float64 `json:"dutiablePieceCount"` // sum of the dutiable line quantities
	ExcludedCount      int     `json:"excludedCount"`
	ReceiptCount       int     `json:"receiptCount"`
	DutyFreeTotalCents int64   `json:"dutyFreeTotalCents"`
	DutiableTotalCents int64   `json:"dutiableTotalCents"`
	GrandTotalCents    int64   `json:"grandTotalCents"`
	Issues             []Issue `json:"issues"`
}

type Issue struct {
	Code     IssueCode   `json:"code"`
	Severity Severity    `json:"severity"`
	Items    []IssueItem `json:"items"`
}

type IssueItem struct {
	ID      uuid.UUID `json:"id"`
	AssetID string    `json:"assetId"`
	Name    string    `json:"name"`
	Detail  string    `json:"detail"`
}

// --- domain input (built by the service) ---

type Tag struct {
	ID, ParentID uuid.UUID
	Name         string
}

type Attachment struct {
	ID                    uuid.UUID
	Type, Title, MimeType string
	CreatedAt             time.Time
	IsLink                bool
}

type Item struct {
	ID             uuid.UUID
	ParentID       uuid.UUID // the parent entity (an item or a location), or uuid.Nil
	AssetID        int64
	Name           string
	EntityTypeID   uuid.UUID
	EntityTypeName string
	IsLocation     bool
	Archived       bool
	SoldDate       types.Date
	Quantity       float64
	PurchasePrice  float64 // per piece
	PurchaseDate   types.Date
	LocationID     uuid.UUID
	LocationPath   []string // root first; empty = no location
	TagIDs         []uuid.UUID
	Attachments    []Attachment
}

type Input struct {
	Options     Options
	Header      Header
	Currency    string
	GeneratedAt time.Time
	Tags        []Tag
	Items       []Item
	ScopeIDs    []uuid.UUID // every requested or resolved ID, found or not
}

// --- output ---

type Report struct {
	Header                Header
	Currency              string
	ReferenceDate, Cutoff types.Date
	MinMonths             int
	GeneratedAt           time.Time
	ItemCount             int
	Groups                []Group
	Dutiable              []Line
	DutyFreeTotalCents    int64
	DutiableTotalCents    int64
	GrandTotalCents       int64
	DutyFreeQty           float64
	DutiableQty           float64
	TotalQty              float64
	Excluded              []Excluded
	Issues                []Issue
	Receipts              []ReceiptFile
	ScopeIDs              []uuid.UUID
}

type Group struct {
	Title         string
	Lines         []Line
	SubtotalCents int64
	SubtotalQty   float64
}

type Line struct {
	Pos              int
	Quantity         float64
	Name             string
	EntityTypeName   string
	Locations        []string
	AssetIDs         []string
	PurchaseDate     types.Date
	PurchaseDateText string // set only for merged date ranges
	UnitCents        int64  // price per piece
	HasPrice         bool
	LineCents        int64 // LineCents(Quantity, UnitCents)
	ReceiptNames     []string
	Comments         []string
	Status           Status
	Reason           Reason
	MemberIDs        []uuid.UUID
	MemberQuantities []float64
	MemberOwnCents   []int64   // each member's own price per piece, parallel to MemberIDs
	Contents         []Content // entities absorbed into the members, by member
}

// Content is an entity absorbed into a line member (FR-30, FR-31).
type Content struct {
	ID, OwnerID uuid.UUID
	Quantity    float64
	UnitCents   int64 // its own price per piece
}

type Excluded struct {
	ID      uuid.UUID
	AssetID string
	Name    string
	Code    IssueCode
}

type ReceiptFile struct {
	Name         string
	AttachmentID uuid.UUID
	ItemID       uuid.UUID
}
