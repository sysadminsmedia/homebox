package customsinventory

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
)

// entry is one included item with everything derived from it.
type entry struct {
	it           Item
	name         string // trimmed, inner spaces collapsed
	firstWord    string // FirstWord(name)
	class        Classification
	unitCents    int64
	hasPrice     bool
	location     string
	assetLabel   string
	receipts     []Attachment // receipt attachments of the selected types, links excluded, by CreatedAt
	receiptNames []string     // ZIP names, parallel to receipts; empty when receipts are off
	links        []Attachment
	comment      string
}

type groupAcc struct {
	key, title string
	order      int // tag mode: position in GroupTagIDs; other modes: unused
	last       bool
	entries    []*entry
}

func normalizeName(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// FirstWord returns the first run of letters and digits in the normalized
// name, or the whole normalized name when it has none (FR-24).
func FirstWord(name string) string {
	n := normalizeName(name)
	isWord := func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }
	start := strings.IndexFunc(n, isWord)
	if start < 0 {
		return n
	}
	rest := n[start:]
	if end := strings.IndexFunc(rest, func(r rune) bool { return !isWord(r) }); end >= 0 {
		return rest[:end]
	}
	return rest
}

// compareFold compares two strings ignoring case.
func compareFold(a, b string) int {
	return strings.Compare(strings.ToLower(a), strings.ToLower(b))
}

func compareEntries(a, b *entry) int {
	return cmp.Or(
		compareFold(a.name, b.name),
		cmp.Compare(a.it.AssetID, b.it.AssetID),
		strings.Compare(a.it.ID.String(), b.it.ID.String()),
	)
}

func assetLabel(it Item) string {
	if it.AssetID > 0 {
		return FormatAssetID(it.AssetID)
	}
	return "id-" + strings.ReplaceAll(it.ID.String(), "-", "")[:8]
}

// Build turns the input into a verified report. It's the only calculation:
// the preview and the export both call it.
func Build(in Input) (Report, error) {
	o := in.Options
	ref, err := ParseDate(o.ReferenceDate)
	if err != nil {
		return Report{}, fmt.Errorf("%w: reference date must be YYYY-MM-DD", ErrInvalidRequest)
	}
	minMonths := o.MinMonths
	if minMonths == 0 {
		minMonths = DefaultMinMonths
	}
	receiptTypes := o.ReceiptTypes
	if len(receiptTypes) == 0 {
		receiptTypes = DefaultReceiptTypes
	}

	tags := NewTagIndex(in.Tags)
	rules := Rules{
		Reference: ref,
		Cutoff:    Cutoff(ref, minMonths),
		MinMonths: minMonths,
		DutyFree:  tags.Expand(o.DutyFreeTagIDs),
		Dutiable:  tags.Expand(o.DutiableTagIDs),
		Tags:      tags,
	}

	r := Report{
		Header:        in.Header,
		Currency:      strings.ToUpper(in.Currency),
		ReferenceDate: ref,
		Cutoff:        rules.Cutoff,
		MinMonths:     minMonths,
		GeneratedAt:   in.GeneratedAt,
	}
	issues := newIssueSet()

	byID := make(map[uuid.UUID]Item, len(in.Items))
	for _, it := range in.Items {
		byID[it.ID] = it
	}

	var entries []*entry
	seen := make(map[uuid.UUID]bool, len(in.ScopeIDs))
	for _, sid := range in.ScopeIDs {
		if seen[sid] {
			continue
		}
		seen[sid] = true
		r.ScopeIDs = append(r.ScopeIDs, sid)

		it, ok := byID[sid]
		if code, excluded := exclusion(it, ok); excluded {
			ex := Excluded{ID: sid, Code: code}
			if ok {
				ex.AssetID, ex.Name = FormatAssetID(it.AssetID), normalizeName(it.Name)
			}
			r.Excluded = append(r.Excluded, ex)
			issues.add(code, IssueItem{ID: ex.ID, AssetID: ex.AssetID, Name: ex.Name})
			continue
		}
		entries = append(entries, newEntry(it, rules, receiptTypes, minMonths))
	}
	r.ItemCount = len(entries)
	slices.SortFunc(r.Excluded, func(a, b Excluded) int {
		return cmp.Or(compareFold(a.Name, b.Name), strings.Compare(a.AssetID, b.AssetID), strings.Compare(a.ID.String(), b.ID.String()))
	})

	assignReceiptNames(entries, o.IncludeReceipts)
	itemIssues(entries, issues, o.IncludeReceipts)

	groups := groupEntries(entries, o, tags)
	pos := 0
	for _, g := range groups {
		firstWord := o.MergeIdentical && o.MergeFirstWord
		if firstWord {
			slices.SortFunc(g.entries, func(a, b *entry) int {
				return cmp.Or(compareFold(a.firstWord, b.firstWord), compareEntries(a, b))
			})
		} else {
			slices.SortFunc(g.entries, compareEntries)
		}
		var dutyFree, dutiable []*entry
		for _, e := range g.entries {
			if e.class.Status == StatusDutyFree {
				dutyFree = append(dutyFree, e)
			} else {
				dutiable = append(dutiable, e)
			}
		}
		if len(dutyFree) > 0 {
			grp := Group{Title: g.title, Lines: makeLines(dutyFree, o.MergeIdentical, firstWord)}
			for i := range grp.Lines {
				grp.SubtotalCents += grp.Lines[i].LineCents
				grp.SubtotalQty += grp.Lines[i].Quantity
			}
			grp.SubtotalQty = RoundQty(grp.SubtotalQty)
			r.Groups = append(r.Groups, grp)
			r.DutyFreeTotalCents += grp.SubtotalCents
			r.DutyFreeQty += grp.SubtotalQty
		}
		for _, l := range makeLines(dutiable, o.MergeIdentical, firstWord) {
			r.Dutiable = append(r.Dutiable, l)
			r.DutiableTotalCents += l.LineCents
			r.DutiableQty += l.Quantity
		}
	}
	r.GrandTotalCents = r.DutyFreeTotalCents + r.DutiableTotalCents
	r.DutyFreeQty = RoundQty(r.DutyFreeQty)
	r.DutiableQty = RoundQty(r.DutiableQty)
	r.TotalQty = RoundQty(r.DutyFreeQty + r.DutiableQty)

	for gi := range r.Groups {
		for li := range r.Groups[gi].Lines {
			pos++
			r.Groups[gi].Lines[li].Pos = pos
		}
	}
	for li := range r.Dutiable {
		pos++
		r.Dutiable[li].Pos = pos
	}

	if o.IncludeReceipts {
		r.Receipts = collectReceipts(r, entries)
	}
	r.Issues = issues.list()

	return r, Verify(r)
}

func exclusion(it Item, found bool) (IssueCode, bool) {
	switch {
	case !found:
		return IssueExcludedNotFound, true
	case it.IsLocation:
		return IssueExcludedLocation, true
	case it.Archived:
		return IssueExcludedArchived, true
	case !it.SoldDate.Time().IsZero():
		return IssueExcludedSold, true
	default:
		return "", false
	}
}

func newEntry(it Item, rules Rules, receiptTypes []string, minMonths int) *entry {
	e := &entry{
		it:         it,
		name:       normalizeName(it.Name),
		firstWord:  FirstWord(it.Name),
		class:      Classify(it, rules),
		unitCents:  ToCents(it.PurchasePrice),
		location:   strings.Join(it.LocationPath, " › "),
		assetLabel: assetLabel(it),
	}
	e.hasPrice = e.unitCents != 0

	e.comment = ReasonComment(e.class, minMonths)
	if !e.hasPrice {
		e.comment = strings.TrimPrefix(e.comment+"; Wert fehlt", "; ")
	}

	for _, a := range it.Attachments {
		if !slices.Contains(receiptTypes, a.Type) {
			continue
		}
		if a.IsLink {
			e.links = append(e.links, a)
		} else {
			e.receipts = append(e.receipts, a)
		}
	}
	slices.SortFunc(e.receipts, func(a, b Attachment) int {
		return cmp.Or(a.CreatedAt.Compare(b.CreatedAt), strings.Compare(a.ID.String(), b.ID.String()))
	})
	return e
}

// assignReceiptNames names every bundled receipt (FR-19). Items that share an
// asset label get "_b", "_c"… by internal ID, so names never collide.
func assignReceiptNames(entries []*entry, include bool) {
	if !include {
		return
	}
	byLabel := map[string][]*entry{}
	for _, e := range entries {
		byLabel[e.assetLabel] = append(byLabel[e.assetLabel], e)
	}
	for _, list := range byLabel {
		slices.SortFunc(list, func(a, b *entry) int { return strings.Compare(a.it.ID.String(), b.it.ID.String()) })
		for dup, e := range list {
			e.receiptNames = make([]string, len(e.receipts))
			for i, a := range e.receipts {
				e.receiptNames[i] = ReceiptName(e.assetLabel, dup, i, a.Title, a.MimeType)
			}
		}
	}
}

// itemIssues adds the per-item warnings (FR-11). Excluded entities never get here.
func itemIssues(entries []*entry, issues *issueSet, include bool) {
	assetCount := map[int64]int{}
	for _, e := range entries {
		if e.it.AssetID > 0 {
			assetCount[e.it.AssetID]++
		}
	}
	for _, e := range entries {
		ref := IssueItem{ID: e.it.ID, AssetID: FormatAssetID(e.it.AssetID), Name: e.name}
		for _, code := range e.class.Issues {
			item := ref
			if code == IssuePurchasedAfterReference || code == IssueConflictDateTag {
				item.Detail = e.it.PurchaseDate.String()
			}
			issues.add(code, item)
		}
		if !e.hasPrice {
			issues.add(IssueMissingPrice, ref)
		}
		if e.unitCents < 0 || e.it.Quantity < 0 {
			issues.add(IssueNegativeValue, ref)
		}
		if e.it.Quantity == 0 {
			issues.add(IssueZeroQuantity, ref)
		}
		if e.it.AssetID <= 0 {
			issues.add(IssueMissingAssetID, ref)
		} else if assetCount[e.it.AssetID] > 1 {
			issues.add(IssueDuplicateAssetID, ref)
		}
		if e.class.Status == StatusDutiable && len(e.receipts) == 0 {
			issues.add(IssueDutiableWithoutReceipt, ref)
		}
		if include && len(e.links) > 0 {
			item := ref
			item.Detail = e.links[0].Title
			issues.add(IssueReceiptIsLink, item)
		}
	}
}

// groupEntries puts every entry into its group (FR-4) and orders the groups (FR-5).
func groupEntries(entries []*entry, o Options, tags TagIndex) []*groupAcc {
	byKey := map[string]*groupAcc{}
	var groups []*groupAcc
	for _, e := range entries {
		key, title, order, last := groupOf(e, o, tags)
		g, ok := byKey[key]
		if !ok {
			g = &groupAcc{key: key, title: title, order: order, last: last}
			byKey[key] = g
			groups = append(groups, g)
		}
		g.entries = append(g.entries, e)
	}
	slices.SortFunc(groups, func(a, b *groupAcc) int {
		if a.last != b.last {
			if a.last {
				return 1
			}
			return -1
		}
		if o.GroupBy == GroupByTag {
			return cmp.Compare(a.order, b.order)
		}
		return cmp.Or(compareFold(a.title, b.title), strings.Compare(a.key, b.key))
	})
	return groups
}

func groupOf(e *entry, o Options, tags TagIndex) (key, title string, order int, last bool) {
	switch o.GroupBy {
	case GroupByEntityType:
		if e.it.EntityTypeID == uuid.Nil {
			return "other", LabelOther, 0, true
		}
		return e.it.EntityTypeID.String(), e.it.EntityTypeName, 0, false
	case GroupByTag:
		g, ok := tags.FirstGroup(e.it.TagIDs, o.GroupTagIDs)
		if !ok {
			return "other", LabelOther, 0, true
		}
		return g.String(), tags.Name(g), slices.Index(o.GroupTagIDs, g), false
	default:
		if len(e.it.LocationPath) == 0 {
			return "none", LabelNoLocation, 0, true
		}
		return e.it.LocationID.String(), e.location, 0, false
	}
}

type mergeKey struct {
	entityType uuid.UUID
	name       string
	status     Status
	unitCents  int64
}

// makeLines turns sorted entries of one group and status into lines,
// merging identical items when asked (FR-7), by the first word of the name
// in first-word mode (FR-25).
func makeLines(entries []*entry, merge, firstWord bool) []Line {
	var buckets [][]*entry
	index := map[mergeKey]int{}
	for _, e := range entries {
		if !merge {
			buckets = append(buckets, []*entry{e})
			continue
		}
		name := e.name
		if firstWord {
			name = e.firstWord
		}
		k := mergeKey{e.it.EntityTypeID, strings.ToLower(name), e.class.Status, e.unitCents}
		if i, ok := index[k]; ok {
			buckets[i] = append(buckets[i], e)
			continue
		}
		index[k] = len(buckets)
		buckets = append(buckets, []*entry{e})
	}

	lines := make([]Line, len(buckets))
	for i, b := range buckets {
		lines[i] = lineFrom(b, firstWord)
	}
	return lines
}

func appendUnique(list []string, s string) []string {
	if s == "" || slices.Contains(list, s) {
		return list
	}
	return append(list, s)
}

// lineFrom builds one line. In first-word mode, members with different full
// names share the first member's first word as the line name (FR-26).
func lineFrom(members []*entry, firstWord bool) Line {
	first := members[0]
	name := first.name
	if firstWord && slices.ContainsFunc(members, func(m *entry) bool { return !strings.EqualFold(m.name, first.name) }) {
		name = first.firstWord
	}
	l := Line{
		Name:           name,
		EntityTypeName: first.it.EntityTypeName,
		UnitCents:      first.unitCents,
		HasPrice:       first.hasPrice,
		Status:         first.class.Status,
		Reason:         first.class.Reason,
	}

	var qty float64
	var assetIDs []int64
	var dates []types.Date
	for _, m := range members {
		qty += m.it.Quantity
		l.MemberIDs = append(l.MemberIDs, m.it.ID)
		l.MemberQuantities = append(l.MemberQuantities, m.it.Quantity)
		l.Locations = appendUnique(l.Locations, m.location)
		l.Comments = appendUnique(l.Comments, m.comment)
		l.ReceiptNames = append(l.ReceiptNames, m.receiptNames...)
		if m.it.AssetID > 0 {
			assetIDs = append(assetIDs, m.it.AssetID)
		}
		if d := m.it.PurchaseDate; !d.Time().IsZero() && !slices.ContainsFunc(dates, func(x types.Date) bool { return x.Time().Equal(d.Time()) }) {
			dates = append(dates, d)
		}
	}
	l.Quantity = RoundQty(qty)
	l.LineCents = LineCents(l.Quantity, l.UnitCents)

	slices.Sort(assetIDs)
	for _, a := range assetIDs {
		l.AssetIDs = append(l.AssetIDs, FormatAssetID(a))
	}

	slices.SortFunc(dates, func(a, b types.Date) int { return a.Time().Compare(b.Time()) })
	switch len(dates) {
	case 0:
	case 1:
		l.PurchaseDate = dates[0]
	default:
		l.PurchaseDateText = dates[0].String() + " – " + dates[len(dates)-1].String()
	}
	return l
}

// collectReceipts lists the bundled receipts in line order.
func collectReceipts(r Report, entries []*entry) []ReceiptFile {
	byID := make(map[uuid.UUID]*entry, len(entries))
	for _, e := range entries {
		byID[e.it.ID] = e
	}
	var out []ReceiptFile
	for _, l := range allReportLines(r) {
		for _, m := range l.MemberIDs {
			e := byID[m]
			for i, a := range e.receipts {
				out = append(out, ReceiptFile{Name: e.receiptNames[i], AttachmentID: a.ID, ItemID: e.it.ID})
			}
		}
	}
	return out
}

func allReportLines(r Report) []Line {
	var out []Line
	for _, g := range r.Groups {
		out = append(out, g.Lines...)
	}
	return append(out, r.Dutiable...)
}

// issueSet collects issue items per code, one entry per item.
type issueSet struct {
	items map[IssueCode][]IssueItem
	seen  map[IssueCode]map[uuid.UUID]bool
}

func newIssueSet() *issueSet {
	return &issueSet{items: map[IssueCode][]IssueItem{}, seen: map[IssueCode]map[uuid.UUID]bool{}}
}

func (s *issueSet) add(code IssueCode, item IssueItem) {
	if s.seen[code] == nil {
		s.seen[code] = map[uuid.UUID]bool{}
	}
	if s.seen[code][item.ID] {
		return
	}
	s.seen[code][item.ID] = true
	s.items[code] = append(s.items[code], item)
}

// list returns the issues in IssueOrder, items sorted by name, then asset ID.
func (s *issueSet) list() []Issue {
	out := []Issue{}
	for _, code := range IssueOrder {
		items := s.items[code]
		if len(items) == 0 {
			continue
		}
		slices.SortFunc(items, func(a, b IssueItem) int {
			return cmp.Or(compareFold(a.Name, b.Name), strings.Compare(a.AssetID, b.AssetID), strings.Compare(a.ID.String(), b.ID.String()))
		})
		out = append(out, Issue{Code: code, Severity: SeverityOf(code), Items: items})
	}
	return out
}

// Preview summarizes the report for the dialog.
func (r Report) Preview() Preview {
	p := Preview{
		Currency:           r.Currency,
		ReferenceDate:      r.ReferenceDate.String(),
		MinMonths:          r.MinMonths,
		Cutoff:             r.Cutoff.String(),
		ItemCount:          r.ItemCount,
		ExcludedCount:      len(r.Excluded),
		ReceiptCount:       len(r.Receipts),
		DutyFreeTotalCents: r.DutyFreeTotalCents,
		DutiableTotalCents: r.DutiableTotalCents,
		GrandTotalCents:    r.GrandTotalCents,
		PieceCount:         r.TotalQty,
		DutiablePieceCount: r.DutiableQty,
		Issues:             r.Issues,
	}
	for _, g := range r.Groups {
		p.LineCount += len(g.Lines)
	}
	p.LineCount += len(r.Dutiable)
	for _, l := range r.Dutiable {
		p.DutiableItemCount += len(l.MemberIDs)
	}
	return p
}
