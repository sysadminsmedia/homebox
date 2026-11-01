package customsinventory

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Names used in several tests.
const (
	nameFiles         = "Files"
	nameEspresso      = "Espresso machine"
	typeNameFurniture = "Furniture"
	tagNameDutiable   = "dutiable"
)

// Tags of the golden fixture. tagP, tagP2 and tagZ are shared with classify_test.go.
var (
	tagA  = id(911) // "Group A"
	tagA1 = id(912) // "Group A child", child of A
	tagB  = id(913) // "Group B"
)

// Entity types and locations of the golden fixture.
var (
	typeFurniture = id(801)
	typeBox       = id(802)
	typeAppliance = id(803)
	typeTool      = id(804)
	typeLocation  = id(805)
	typeVehicle   = id(806)

	locLiving  = id(701)
	locOffice  = id(702)
	locKitchen = id(703)
	locGarage  = id(704)

	missingID = id(999)
)

var (
	pathLiving  = []string{"Home", "Living room"}
	pathOffice  = []string{"Home", "Office"}
	pathKitchen = []string{"Home", "Kitchen"}
	pathGarage  = []string{"Home", "Garage"}
)

func at(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func furniture(itemID int, asset int64, name string, qty, price float64, bought string, loc uuid.UUID, path []string) Item {
	return Item{
		ID: id(itemID), AssetID: asset, Name: name,
		EntityTypeID: typeFurniture, EntityTypeName: typeNameFurniture,
		Quantity: qty, PurchasePrice: price, PurchaseDate: dateOrZero(bought),
		LocationID: loc, LocationPath: path,
	}
}

func receipt(attID int, title, mime, created string) Attachment {
	return Attachment{ID: id(attID), Type: "receipt", Title: title, MimeType: mime, CreatedAt: at(created)}
}

// goldenInput is the fixture behind the golden report. Prices are per piece.
func goldenInput() Input {
	clock := furniture(101, 101, "Grandfather clock", 1, 1200.00, "2019-03-12", locLiving, pathLiving)
	clock.Attachments = []Attachment{receipt(5101, "invoice-clock.pdf", "application/pdf", "2019-03-12T12:00:00Z")}

	sofa := furniture(102, 102, "Sofa", 1, 890.00, "2021-06-01", locLiving, pathLiving)
	sofa.Attachments = []Attachment{
		{ID: id(5102), Type: "receipt", Title: "https://shop.example/order/1", IsLink: true, CreatedAt: at("2021-06-01T12:00:00Z")},
		{ID: id(5103), Type: "photo", Title: "sofa.jpg", MimeType: "image/jpeg", CreatedAt: at("2021-06-02T12:00:00Z")},
	}

	chair := furniture(103, 103, "Dining chair", 4, 89.50, "2020-05-01", locLiving, pathLiving)
	shelf := furniture(110, 110, "Shelf", 2.5, 0.33, "2020-01-01", locOffice, pathOffice)

	items := []Item{clock, sofa, chair, shelf}

	for n := 131; n <= 144; n++ {
		name := " files "
		if n%2 == 1 {
			name = nameFiles
		}
		items = append(items, Item{
			ID: id(n), AssetID: int64(n), Name: name, EntityTypeID: typeBox, EntityTypeName: "Box",
			Quantity: 1, TagIDs: []uuid.UUID{tagP2}, LocationID: locOffice, LocationPath: pathOffice,
		})
	}
	items = append(items, Item{
		ID: id(145), AssetID: 145, Name: nameFiles, EntityTypeID: typeBox, EntityTypeName: "Box",
		Quantity: 1, PurchasePrice: 5.00, TagIDs: []uuid.UUID{tagP2}, LocationID: locOffice, LocationPath: pathOffice,
	})

	espresso := Item{
		ID: id(146), AssetID: 146, Name: nameEspresso, EntityTypeID: typeAppliance, EntityTypeName: "Appliance",
		Quantity: 1, PurchasePrice: 649.90, PurchaseDate: d("2026-09-02"), TagIDs: []uuid.UUID{tagP2},
		LocationID: locKitchen, LocationPath: pathKitchen,
		Attachments: []Attachment{
			{ID: id(5146), Type: "attachment", Title: "scan.jpg", MimeType: "image/jpeg", CreatedAt: at("2026-09-03T10:00:00Z")},
			receipt(5145, "invoice-espresso.pdf", "application/pdf", "2026-09-02T10:00:00Z"),
		},
	}
	officeChair := furniture(150, 150, "Office chair", 1, 0, "", locOffice, pathOffice)
	toolbox := Item{
		ID: id(160), Name: "Toolbox", EntityTypeID: typeTool, EntityTypeName: "Tool",
		Quantity: 1, PurchasePrice: 30.00, PurchaseDate: d("2018-01-01"),
	}
	floorLamp := furniture(1631, 163, "Floor lamp", 1, 79.00, "2026-10-01", locLiving, pathLiving)
	floorLamp.Attachments = []Attachment{receipt(5163, "floor-lamp.pdf", "application/pdf", "2026-10-01T10:00:00Z")}
	deskLamp := furniture(1632, 163, "Desk lamp", 1, 45.50, "2026-10-01", locLiving, pathLiving)
	deskLamp.Attachments = []Attachment{receipt(5164, "desk-lamp.pdf", "application/pdf", "2026-10-01T11:00:00Z")}

	bicycle := Item{
		ID: id(170), AssetID: 170, Name: "Bicycle", EntityTypeID: typeVehicle, EntityTypeName: "Vehicle",
		Quantity: 1, TagIDs: []uuid.UUID{tagZ, tagP2}, LocationID: locGarage, LocationPath: pathGarage,
	}
	tent := Item{
		ID: id(171), AssetID: 171, Name: "Tent", EntityTypeID: typeTool, EntityTypeName: "Tool",
		Quantity: 1, PurchaseDate: d("2026-11-20"), LocationID: locGarage, LocationPath: pathGarage,
	}
	basement := Item{ID: id(180), AssetID: 180, Name: "Basement", EntityTypeID: typeLocation, EntityTypeName: "Location", IsLocation: true}
	radio := furniture(181, 181, "Old radio", 1, 25.00, "2010-01-01", locOffice, pathOffice)
	radio.Archived = true
	bookshelf := furniture(182, 182, "Bookshelf", 1, 120.00, "2015-01-01", locOffice, pathOffice)
	bookshelf.SoldDate = d("2026-01-10")

	items = append(items, espresso, officeChair, toolbox, floorLamp, deskLamp, bicycle, tent, basement, radio, bookshelf)

	scope := make([]uuid.UUID, 0, len(items)+1)
	for _, it := range items {
		scope = append(scope, it.ID)
	}
	scope = append(scope, missingID)

	return Input{
		Options: Options{
			GroupBy:         GroupByLocation,
			DutyFreeTagIDs:  []uuid.UUID{tagP},
			DutiableTagIDs:  []uuid.UUID{tagZ},
			ReferenceDate:   "2026-11-15",
			MergeIdentical:  true,
			IncludeReceipts: true,
			ReceiptTypes:    []string{"receipt", "attachment"},
		},
		Header: Header{
			Title:   "Inventory list – Shipment A",
			Name:    "Jane Doe",
			Address: "1 Example Street\n12345 Sample Town",
			Phone:   "+1 555 0100",
			Email:   "jane@example.com",
		},
		Currency:    "eur",
		GeneratedAt: at("2026-10-31T10:00:00Z"),
		Tags: []Tag{
			{ID: tagP, Name: "owned"},
			{ID: tagP2, ParentID: tagP, Name: "owned > 2 years"},
			{ID: tagZ, Name: tagNameDutiable},
			{ID: tagA, Name: "Group A"},
			{ID: tagA1, ParentID: tagA, Name: "Group A child"},
			{ID: tagB, Name: "Group B"},
		},
		Items:    items,
		ScopeIDs: scope,
	}
}

// smallInput wraps items in a minimal input with every item in scope.
func smallInput(items ...Item) Input {
	in := Input{
		Options:     Options{GroupBy: GroupByLocation, ReferenceDate: "2026-11-15", MergeIdentical: true},
		Currency:    "EUR",
		GeneratedAt: at("2026-10-31T10:00:00Z"),
		Items:       items,
	}
	for _, it := range items {
		in.ScopeIDs = append(in.ScopeIDs, it.ID)
	}
	return in
}

func mustBuild(t *testing.T, in Input) Report {
	t.Helper()
	r, err := Build(in)
	require.NoError(t, err)
	return r
}

func setItem(in *Input, itemID uuid.UUID, mutate func(*Item)) {
	for i := range in.Items {
		if in.Items[i].ID == itemID {
			mutate(&in.Items[i])
			return
		}
	}
	panic("no item " + itemID.String())
}

// allLines returns every line of the report in position order.
func allLines(r Report) []Line {
	var out []Line
	for _, g := range r.Groups {
		out = append(out, g.Lines...)
	}
	return append(out, r.Dutiable...)
}

func lineByName(t *testing.T, r Report, name string) Line {
	t.Helper()
	for _, l := range allLines(r) {
		if l.Name == name {
			return l
		}
	}
	t.Fatalf("no line %q", name)
	return Line{}
}

func issueNames(r Report, code IssueCode) []string {
	for _, is := range r.Issues {
		if is.Code == code {
			names := make([]string, len(is.Items))
			for i, it := range is.Items {
				names[i] = it.Name
			}
			return names
		}
	}
	return nil
}
