package services

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"

	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/customsinventory"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
)

var customsNow = time.Date(2026, 10, 31, 10, 0, 0, 0, time.UTC)

const customsPDF = "%PDF-1.7 grandfather clock invoice"

// customsFixture is a fresh group with two locations, four items and two tags.
type customsFixture struct {
	ctx      Context
	shipment uuid.UUID // tag "Shipment A"
	owned    uuid.UUID // tag "owned > 2 years"
	clock    uuid.UUID // 1 × 1200.00, 2019-03-12, Shipment A, receipt
	chairs   uuid.UUID // 3 × 89.50, 2020-05-01, Shipment A
	espresso uuid.UUID // 1 × 649.90, 2026-09-02, Shipment A + owned (conflict)
	lamp     uuid.UUID // 1 × 20.00, no date, owned
	itemType uuid.UUID
	living   uuid.UUID
}

func newCustomsFixture(t testing.TB) customsFixture {
	t.Helper()
	bg := context.Background()

	g, err := tRepos.Groups.GroupCreate(bg, "customs-"+fk.Str(6), uuid.Nil)
	require.NoError(t, err)
	f := customsFixture{ctx: Context{Context: bg, GID: g.ID, UID: tUser.ID}}

	locType, err := tRepos.EntityTypes.GetDefault(bg, g.ID, true)
	require.NoError(t, err)
	itemType, err := tRepos.EntityTypes.GetDefault(bg, g.ID, false)
	require.NoError(t, err)

	newEntity := func(name string, parent, typ uuid.UUID) uuid.UUID {
		e, err := tRepos.Entities.Create(bg, g.ID, repo.EntityCreate{Name: name, ParentID: parent, EntityTypeID: typ})
		require.NoError(t, err)
		return e.ID
	}
	home := newEntity("Home", uuid.Nil, locType.ID)
	f.living = newEntity("Living room", home, locType.ID)
	office := newEntity("Office", home, locType.ID)
	f.itemType = itemType.ID

	newTag := func(name string) uuid.UUID {
		tg, err := tRepos.Tags.Create(bg, g.ID, repo.TagCreate{Name: name})
		require.NoError(t, err)
		return tg.ID
	}
	f.shipment = newTag("Shipment A")
	f.owned = newTag("owned > 2 years")

	f.clock = f.newItem(t, "Grandfather clock", f.living, 1, 1, 1200.00, "2019-03-12", f.shipment)
	f.chairs = f.newItem(t, "Dining chair", f.living, 2, 3, 89.50, "2020-05-01", f.shipment)
	f.espresso = f.newItem(t, "Espresso machine", office, 3, 1, 649.90, "2026-09-02", f.shipment, f.owned)
	f.lamp = f.newItem(t, "Old lamp", office, 4, 1, 20.00, "", f.owned)

	_, err = tRepos.Attachments.Create(bg, f.clock,
		repo.ItemCreateAttachment{Title: "rechnung.pdf", Content: bytes.NewReader([]byte(customsPDF))},
		attachment.TypeReceipt, false)
	require.NoError(t, err)

	return f
}

// newItem creates an item under parent (a location or another item).
func (f customsFixture) newItem(t testing.TB, name string, parent uuid.UUID, asset int64, qty, price float64, bought string, tags ...uuid.UUID) uuid.UUID {
	t.Helper()
	e, err := tRepos.Entities.Create(f.ctx, f.ctx.GID, repo.EntityCreate{Name: name, ParentID: parent, EntityTypeID: f.itemType})
	require.NoError(t, err)
	upd := tClient.Entity.UpdateOneID(e.ID).SetQuantity(qty).SetPurchasePrice(price).AddTagIDs(tags...)
	if bought != "" {
		d, err := time.Parse(time.DateOnly, bought)
		require.NoError(t, err)
		upd.SetPurchaseDate(d)
	}
	_, err = upd.Save(f.ctx)
	require.NoError(t, err)
	require.NoError(t, tRepos.Entities.SetAssetID(f.ctx, f.ctx.GID, e.ID, repo.AssetID(asset)))
	return e.ID
}

type customsBox struct{ box, book, pouch, pen uuid.UUID }

// withBox adds a moving box in the living room holding books and a pouch with
// pens. Contents value: 3 × 10.00 + 2.00 + 2 × 1.25 = 34.50.
func withBox(t testing.TB, f customsFixture) customsBox {
	t.Helper()
	var b customsBox
	b.box = f.newItem(t, "Moving box", f.living, 5, 1, 0, "2019-01-01", f.shipment)
	b.book = f.newItem(t, "Book", b.box, 6, 3, 10.00, "2019-01-01")
	b.pouch = f.newItem(t, "Pouch", b.box, 7, 1, 2.00, "2019-01-01")
	b.pen = f.newItem(t, "Pen", b.pouch, 8, 2, 1.25, "2019-01-01")
	return b
}

func (f customsFixture) request(ids ...uuid.UUID) customsinventory.Request {
	return customsinventory.Request{
		Scope: customsinventory.Scope{ItemIDs: ids},
		Options: customsinventory.Options{
			GroupBy:         customsinventory.GroupByLocation,
			DutyFreeTagIDs:  []uuid.UUID{f.owned},
			ReferenceDate:   "2026-11-15",
			MergeIdentical:  true,
			IncludeReceipts: true,
		},
		Header: customsinventory.Header{Title: "Shipment A"},
	}
}

func customsService() *CustomsInventoryService {
	return &CustomsInventoryService{repos: tRepos, now: func() time.Time { return customsNow }}
}

func issueByCode(p customsinventory.Preview, code customsinventory.IssueCode) *customsinventory.Issue {
	for i := range p.Issues {
		if p.Issues[i].Code == code {
			return &p.Issues[i]
		}
	}
	return nil
}

func TestCustomsInventory_PreviewByIDs(t *testing.T) {
	f := newCustomsFixture(t)
	p, err := customsService().Preview(f.ctx, f.request(f.clock, f.chairs, f.espresso, f.lamp))
	require.NoError(t, err)

	assert.Equal(t, 4, p.ItemCount)
	assert.Equal(t, 4, p.LineCount)
	assert.InDelta(t, 1.0, p.DutiablePieceCount, 0)
	assert.Equal(t, 1, p.ReceiptCount)
	// duty-free: 1200.00 + 3 × 89.50 + 20.00; dutiable: 649.90
	assert.Equal(t, int64(120000+26850+2000), p.DutyFreeTotalCents)
	assert.Equal(t, int64(64990), p.DutiableTotalCents)
	assert.Equal(t, int64(120000+26850+2000+64990), p.GrandTotalCents)
	assert.Equal(t, "2026-05-15", p.Cutoff)
	assert.NotNil(t, issueByCode(p, customsinventory.IssueConflictDateTag))
}

func TestCustomsInventory_ScopeByQueryMatchesSearch(t *testing.T) {
	f := newCustomsFixture(t)
	svc := customsService()

	req := f.request()
	req.Scope = customsinventory.Scope{Query: &customsinventory.ScopeQuery{TagIDs: []uuid.UUID{f.shipment}}}
	p, err := svc.Preview(f.ctx, req)
	require.NoError(t, err)

	search, err := tRepos.Entities.QueryByGroup(f.ctx, f.ctx.GID, repo.EntityQuery{Page: -1, PageSize: -1, TagIDs: []uuid.UUID{f.shipment}})
	require.NoError(t, err)
	assert.Equal(t, search.Total, p.ItemCount)
	assert.Equal(t, 3, p.ItemCount)

	req.Scope.Query = &customsinventory.ScopeQuery{Q: "#000-002"}
	p, err = svc.Preview(f.ctx, req)
	require.NoError(t, err)
	assert.Equal(t, 1, p.ItemCount)
	assert.Equal(t, int64(26850), p.GrandTotalCents)
}

func TestCustomsInventory_ForeignIDsExcluded(t *testing.T) {
	f := newCustomsFixture(t)
	other := newCustomsFixture(t)

	p, err := customsService().Preview(f.ctx, f.request(f.clock, other.clock))
	require.NoError(t, err)

	assert.Equal(t, 1, p.ItemCount)
	assert.Equal(t, 1, p.ExcludedCount)
	assert.Equal(t, int64(120000), p.GrandTotalCents)
	is := issueByCode(p, customsinventory.IssueExcludedNotFound)
	require.NotNil(t, is)
	require.Len(t, is.Items, 1)
	assert.Equal(t, other.clock, is.Items[0].ID)
	assert.Empty(t, is.Items[0].Name)
	assert.Empty(t, is.Items[0].AssetID)
}

func TestCustomsInventory_ForeignTagIgnored(t *testing.T) {
	f := newCustomsFixture(t)
	other := newCustomsFixture(t)

	req := f.request(f.lamp)
	req.Options.DutyFreeTagIDs = []uuid.UUID{other.owned}
	p, err := customsService().Preview(f.ctx, req)
	require.NoError(t, err)

	assert.InDelta(t, 1.0, p.DutiablePieceCount, 0, "a tag from another group never makes an item duty-free")
	assert.NotNil(t, issueByCode(p, customsinventory.IssueMissingPurchaseDate))
}

func zipEntries(t *testing.T, b []byte) map[string][]byte {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	require.NoError(t, err)
	out := map[string][]byte{}
	for _, zf := range zr.File {
		rc, err := zf.Open()
		require.NoError(t, err)
		data, err := io.ReadAll(rc)
		require.NoError(t, err)
		_ = rc.Close()
		out[zf.Name] = data
	}
	return out
}

func TestCustomsInventory_ExportZip(t *testing.T) {
	f := newCustomsFixture(t)
	file, err := customsService().Export(f.ctx, f.request(f.clock, f.chairs, f.espresso, f.lamp))
	require.NoError(t, err)

	assert.Equal(t, "Inventarliste_shipment-a_2026-10-31.zip", file.Name)
	assert.Equal(t, "application/zip", file.ContentType)
	var buf bytes.Buffer
	require.NoError(t, file.Write(&buf))

	entries := zipEntries(t, buf.Bytes())
	assert.Contains(t, entries, "Inventarliste_shipment-a_2026-10-31.xlsx")
	assert.Equal(t, customsPDF, string(entries["Belege/000-001_Rechnung.pdf"]))

	// Two items sharing an asset ID: the second one gets "_b".
	require.NoError(t, tRepos.Entities.SetAssetID(f.ctx, f.ctx.GID, f.chairs, repo.AssetID(1)))
	_, err = tRepos.Attachments.Create(f.ctx, f.chairs,
		repo.ItemCreateAttachment{Title: "chairs.pdf", Content: bytes.NewReader([]byte("chairs"))},
		attachment.TypeReceipt, false)
	require.NoError(t, err)

	file, err = customsService().Export(f.ctx, f.request(f.clock, f.chairs))
	require.NoError(t, err)
	buf.Reset()
	require.NoError(t, file.Write(&buf))
	entries = zipEntries(t, buf.Bytes())
	var names []string
	for n := range entries {
		names = append(names, n)
	}
	slices.Sort(names)
	assert.Equal(t, []string{"Belege/000-001_Rechnung.pdf", "Belege/000-001_b_Rechnung.pdf", "Inventarliste_shipment-a_2026-10-31.xlsx"}, names)
}

func TestCustomsInventory_ExportXLSXOnly(t *testing.T) {
	f := newCustomsFixture(t)
	req := f.request(f.clock, f.chairs)
	req.Options.IncludeReceipts = false
	file, err := customsService().Export(f.ctx, req)
	require.NoError(t, err)

	assert.True(t, strings.HasSuffix(file.Name, ".xlsx"), file.Name)
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", file.ContentType)
	var buf bytes.Buffer
	require.NoError(t, file.Write(&buf))
	x, err := excelize.OpenReader(&buf)
	require.NoError(t, err)
	defer func() { _ = x.Close() }()
	title, err := x.GetCellValue(customsinventory.SheetName, "A1")
	require.NoError(t, err)
	assert.Equal(t, "Shipment A", title)
}

func TestCustomsInventory_ScopeTooLarge(t *testing.T) {
	f := newCustomsFixture(t)
	svc := customsService()
	svc.maxScope = 2

	_, err := svc.Preview(f.ctx, f.request(f.clock, f.chairs, f.espresso))
	require.ErrorIs(t, err, customsinventory.ErrScopeTooLarge)

	req := f.request()
	req.Scope = customsinventory.Scope{Query: &customsinventory.ScopeQuery{TagIDs: []uuid.UUID{f.shipment}}}
	_, err = svc.Export(f.ctx, req)
	require.ErrorIs(t, err, customsinventory.ErrScopeTooLarge)
}

func TestToEntityQuery(t *testing.T) {
	q := toEntityQuery(customsinventory.ScopeQuery{Q: "lamp", Fields: []string{"color=red", "bad"}, IncludeArchived: true})
	assert.Equal(t, -1, q.Page)
	assert.Equal(t, -1, q.PageSize)
	assert.Equal(t, "lamp", q.Search)
	assert.Equal(t, []repo.FieldQuery{{Name: "color", Value: "red"}}, q.Fields)
	assert.True(t, q.IncludeArchived)

	q = toEntityQuery(customsinventory.ScopeQuery{Q: "#000-123"})
	assert.Empty(t, q.Search)
	assert.Equal(t, repo.AssetID(123), q.AssetID)
}

func BenchmarkCustomsInventoryPreview1000(b *testing.B) {
	b.StopTimer()
	f := newCustomsFixture(b)
	itemType, err := tRepos.EntityTypes.GetDefault(f.ctx, f.ctx.GID, false)
	require.NoError(b, err)
	ids := []uuid.UUID{f.clock, f.chairs, f.espresso, f.lamp}
	for i := len(ids); i < 1000; i++ {
		e, err := tRepos.Entities.Create(f.ctx, f.ctx.GID, repo.EntityCreate{Name: fk.Str(8), EntityTypeID: itemType.ID})
		require.NoError(b, err)
		ids = append(ids, e.ID)
	}
	svc := customsService()
	req := f.request(ids...)
	b.StartTimer()

	for range b.N {
		if _, err := svc.Preview(f.ctx, req); err != nil {
			b.Fatal(err)
		}
	}
}

func TestCustomsInventory_RollUpContents(t *testing.T) {
	f := newCustomsFixture(t)
	box := withBox(t, f).box
	req := f.request(box)
	req.Options.RollUpContents = true
	p, err := customsService().Preview(f.ctx, req)
	require.NoError(t, err)
	assert.Equal(t, 1, p.LineCount)
	assert.Equal(t, 3, p.ContentCount)
	assert.Equal(t, 4, p.ItemCount)
	assert.Equal(t, int64(3450), p.GrandTotalCents)
}

func TestCustomsInventory_RollUpByQuery(t *testing.T) {
	f := newCustomsFixture(t)
	withBox(t, f)
	req := f.request()
	req.Scope = customsinventory.Scope{Query: &customsinventory.ScopeQuery{TagIDs: []uuid.UUID{f.shipment}}}
	req.Options.RollUpContents = true
	p, err := customsService().Preview(f.ctx, req)
	require.NoError(t, err)
	assert.Equal(t, 3, p.ContentCount)
	assert.Equal(t, 4, p.LineCount, "clock, chairs, espresso, box")
}

func TestCustomsInventory_RollUpCountsTowardLimit(t *testing.T) {
	f := newCustomsFixture(t)
	box := withBox(t, f).box
	req := f.request(box)
	req.Options.RollUpContents = true
	svc := customsService()
	svc.maxScope = 3
	_, err := svc.Preview(f.ctx, req)
	require.ErrorIs(t, err, customsinventory.ErrScopeTooLarge)
}

func TestCustomsInventory_RollUpOffByDefault(t *testing.T) {
	f := newCustomsFixture(t)
	box := withBox(t, f).box
	p, err := customsService().Preview(f.ctx, f.request(box))
	require.NoError(t, err)
	assert.Equal(t, 0, p.ContentCount)
	assert.Equal(t, int64(0), p.GrandTotalCents)
	missing := issueByCode(p, customsinventory.IssueMissingPrice)
	require.NotNil(t, missing)
	assert.Equal(t, "Moving box", missing.Items[0].Name)
}
