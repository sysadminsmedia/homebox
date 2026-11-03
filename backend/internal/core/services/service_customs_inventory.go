package services

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gocloud.dev/blob"

	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/customsinventory"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
)

const (
	contentTypeZip  = "application/zip"
	contentTypeXLSX = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
)

// CustomsInventoryService loads the group's data for a customs inventory
// export and hands it to the customsinventory package.
type CustomsInventoryService struct {
	repos    *repo.AllRepos
	now      func() time.Time // defaults to time.Now
	maxScope int              // defaults to customsinventory.MaxScope
}

// CustomsInventoryFile is a generated export, written on demand.
type CustomsInventoryFile struct {
	Name        string // "<BaseName>.zip" or ".xlsx"
	ContentType string
	Write       func(w io.Writer) error
}

func (s *CustomsInventoryService) clock() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}

func (s *CustomsInventoryService) limit() int {
	if s.maxScope > 0 {
		return s.maxScope
	}
	return customsinventory.MaxScope
}

// Preview builds the report and returns its summary. It never writes files.
func (s *CustomsInventoryService) Preview(ctx Context, req customsinventory.Request) (customsinventory.Preview, error) {
	spanCtx, span := entityServiceTracer().Start(ctx, "service.CustomsInventoryService.Preview",
		trace.WithAttributes(attribute.String("group.id", ctx.GID.String())))
	defer span.End()

	r, _, err := s.build(spanCtx, ctx.GID, req, span)
	if err != nil {
		return customsinventory.Preview{}, err
	}
	return r.Preview(), nil
}

// Export builds the report and returns the .xlsx, or a ZIP with the receipts.
func (s *CustomsInventoryService) Export(ctx Context, req customsinventory.Request) (CustomsInventoryFile, error) {
	spanCtx, span := entityServiceTracer().Start(ctx, "service.CustomsInventoryService.Export",
		trace.WithAttributes(attribute.String("group.id", ctx.GID.String())))
	defer span.End()

	r, paths, err := s.build(spanCtx, ctx.GID, req, span)
	if err != nil {
		return CustomsInventoryFile{}, err
	}
	xlsx, err := customsinventory.WriteXLSX(r)
	if err != nil {
		recordServiceSpanError(span, err)
		return CustomsInventoryFile{}, err
	}

	now := s.clock()
	base := customsinventory.BaseName(req.Header.Title, now)
	if !req.Options.IncludeReceipts {
		return CustomsInventoryFile{
			Name:        base + ".xlsx",
			ContentType: contentTypeXLSX,
			Write: func(w io.Writer) error {
				_, err := w.Write(xlsx)
				return err
			},
		}, nil
	}

	// Fail before anything is sent if a receipt file is gone.
	if err := s.checkReceipts(spanCtx, r.Receipts, paths); err != nil {
		recordServiceSpanError(span, err)
		return CustomsInventoryFile{}, err
	}

	return CustomsInventoryFile{
		Name:        base + ".zip",
		ContentType: contentTypeZip,
		Write: func(w io.Writer) error {
			bucket, err := blob.OpenBucket(ctx, s.repos.Attachments.GetConnString())
			if err != nil {
				return err
			}
			defer func() { _ = bucket.Close() }()

			open := func(rf customsinventory.ReceiptFile) (io.ReadCloser, error) {
				return bucket.NewReader(ctx, s.repos.Attachments.GetFullPath(paths[rf.AttachmentID]), nil)
			}
			return customsinventory.WriteBundle(w, base, now, xlsx, r.Receipts, open)
		},
	}, nil
}

func (s *CustomsInventoryService) checkReceipts(ctx context.Context, receipts []customsinventory.ReceiptFile, paths map[uuid.UUID]string) error {
	if len(receipts) == 0 {
		return nil
	}
	bucket, err := blob.OpenBucket(ctx, s.repos.Attachments.GetConnString())
	if err != nil {
		return err
	}
	defer func() { _ = bucket.Close() }()

	for _, rf := range receipts {
		ok, err := bucket.Exists(ctx, s.repos.Attachments.GetFullPath(paths[rf.AttachmentID]))
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("customs inventory: file of attachment %s is missing", rf.AttachmentID)
		}
	}
	return nil
}

// build loads the input and runs the shared calculation. It also returns the
// storage path of every attachment, by attachment ID.
func (s *CustomsInventoryService) build(ctx context.Context, gid uuid.UUID, req customsinventory.Request, span trace.Span) (customsinventory.Report, map[uuid.UUID]string, error) {
	in, paths, err := s.buildInput(ctx, gid, req)
	if err != nil {
		recordServiceSpanError(span, err)
		return customsinventory.Report{}, nil, err
	}
	r, err := customsinventory.Build(in)
	if err != nil {
		recordServiceSpanError(span, err)
		return customsinventory.Report{}, nil, err
	}
	span.SetAttributes(
		attribute.Int("scope.count", len(in.ScopeIDs)),
		attribute.Int("items.count", r.ItemCount),
		attribute.Int("excluded.count", len(r.Excluded)),
		attribute.Int("receipts.count", len(r.Receipts)),
	)
	return r, paths, nil
}

func (s *CustomsInventoryService) buildInput(ctx context.Context, gid uuid.UUID, req customsinventory.Request) (customsinventory.Input, map[uuid.UUID]string, error) {
	ids, err := s.scopeIDs(ctx, gid, req.Scope)
	if err != nil {
		return customsinventory.Input{}, nil, err
	}
	if len(ids) > s.limit() {
		return customsinventory.Input{}, nil, fmt.Errorf("%w: %d entities, at most %d", customsinventory.ErrScopeTooLarge, len(ids), s.limit())
	}

	in := customsinventory.Input{
		Options:     req.Options,
		Header:      req.Header,
		GeneratedAt: s.clock(),
		ScopeIDs:    ids,
	}
	paths := map[uuid.UUID]string{}
	locationPaths := map[uuid.UUID][]string{}

	var frontier []uuid.UUID
	for _, id := range ids {
		it, err := s.loadItem(ctx, gid, id, &in, paths, locationPaths)
		if ent.IsNotFound(err) {
			continue // stays in ScopeIDs, becomes excluded_not_found
		}
		if err != nil {
			return customsinventory.Input{}, nil, err
		}
		if !it.IsLocation && descends(it) {
			frontier = append(frontier, it.ID)
		}
	}
	if req.Options.RollUpContents {
		if err := s.loadContents(ctx, gid, frontier, &in, paths, locationPaths, s.limit()); err != nil {
			return customsinventory.Input{}, nil, err
		}
	}

	tags, err := s.repos.Tags.GetAll(ctx, gid)
	if err != nil {
		return customsinventory.Input{}, nil, err
	}
	for _, t := range tags {
		in.Tags = append(in.Tags, customsinventory.Tag{ID: t.ID, ParentID: t.ParentID, Name: t.Name})
	}

	group, err := s.repos.Groups.GroupByID(ctx, gid)
	if err != nil {
		return customsinventory.Input{}, nil, err
	}
	in.Currency = group.Currency

	return in, paths, nil
}

// loadItem loads one entity of the group, appends it to the input and records
// its attachment paths. Location paths are cached by location ID.
func (s *CustomsInventoryService) loadItem(ctx context.Context, gid, id uuid.UUID, in *customsinventory.Input, paths map[uuid.UUID]string, locationPaths map[uuid.UUID][]string) (customsinventory.Item, error) {
	e, err := s.repos.Entities.GetOneByGroup(ctx, gid, id)
	if err != nil {
		return customsinventory.Item{}, err
	}

	var path []string
	if e.Location != nil {
		cached, ok := locationPaths[e.Location.ID]
		if !ok {
			entries, err := s.repos.Entities.PathForEntity(ctx, gid, e.Location.ID)
			if err != nil {
				return customsinventory.Item{}, err
			}
			for _, p := range entries {
				cached = append(cached, p.Name)
			}
			locationPaths[e.Location.ID] = cached
		}
		path = cached
	}

	it := toCustomsItem(e, path)
	in.Items = append(in.Items, it)
	for _, a := range e.Attachments {
		paths[a.ID] = a.Path
	}
	return it, nil
}

// descends reports whether the contents of an item are loaded: never below
// an archived or sold entity (FR-28).
func descends(it customsinventory.Item) bool {
	return !it.Archived && it.SoldDate.Time().IsZero()
}

// loadContents loads the descendants of the frontier, level by level,
// through item children only (FR-28). Archived children aren't loaded, as in
// the search. Scope and contents together may hold at most budget entities.
func (s *CustomsInventoryService) loadContents(ctx context.Context, gid uuid.UUID, frontier []uuid.UUID, in *customsinventory.Input, paths map[uuid.UUID]string, locationPaths map[uuid.UUID][]string, budget int) error {
	loaded := make(map[uuid.UUID]bool, len(in.ScopeIDs))
	for _, id := range in.ScopeIDs {
		loaded[id] = true
	}
	count := len(in.ScopeIDs)
	contents := 0
	for len(frontier) > 0 {
		res, err := s.repos.Entities.QueryByGroup(ctx, gid, repo.EntityQuery{Page: -1, PageSize: -1, ParentItemIDs: frontier})
		if err != nil {
			return err
		}
		var next []uuid.UUID
		for _, child := range res.Items {
			if loaded[child.ID] || (child.EntityType != nil && child.EntityType.IsLocation) {
				continue
			}
			loaded[child.ID] = true
			count++
			if count > budget {
				return fmt.Errorf("%w: %d entities, at most %d", customsinventory.ErrScopeTooLarge, count, budget)
			}
			it, err := s.loadItem(ctx, gid, child.ID, in, paths, locationPaths)
			if err != nil {
				return err
			}
			contents++
			if descends(it) {
				next = append(next, it.ID)
			}
		}
		frontier = next
	}
	trace.SpanFromContext(ctx).SetAttributes(attribute.Int("contents.count", contents))
	return nil
}

// scopeIDs returns the requested IDs, or runs the search again without
// pages. Duplicates are dropped, the order is kept.
func (s *CustomsInventoryService) scopeIDs(ctx context.Context, gid uuid.UUID, scope customsinventory.Scope) ([]uuid.UUID, error) {
	raw := scope.ItemIDs
	if scope.Query != nil {
		res, err := s.repos.Entities.QueryByGroup(ctx, gid, toEntityQuery(*scope.Query))
		if err != nil {
			return nil, err
		}
		raw = make([]uuid.UUID, len(res.Items))
		for i, it := range res.Items {
			raw[i] = it.ID
		}
	}

	seen := make(map[uuid.UUID]bool, len(raw))
	ids := make([]uuid.UUID, 0, len(raw))
	for _, id := range raw {
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// toEntityQuery mirrors HandleEntitiesGetAll's query parsing, including the
// "#000-123" asset ID search, without paging.
func toEntityQuery(q customsinventory.ScopeQuery) repo.EntityQuery {
	v := repo.EntityQuery{
		Page:             -1,
		PageSize:         -1,
		Search:           q.Q,
		ParentIDs:        q.ParentIDs,
		EntityTypeIDs:    q.EntityTypeIDs,
		TagIDs:           q.TagIDs,
		NegateTags:       q.NegateTags,
		MatchAllTags:     q.MatchAllTags,
		OnlyWithoutPhoto: q.OnlyWithoutPhoto,
		OnlyWithPhoto:    q.OnlyWithPhoto,
		IncludeArchived:  q.IncludeArchived,
	}
	for _, f := range q.Fields {
		if name, value, ok := strings.Cut(f, "="); ok {
			v.Fields = append(v.Fields, repo.FieldQuery{Name: name, Value: value})
		}
	}
	if aidStr, ok := strings.CutPrefix(v.Search, "#"); ok {
		if aid, ok := repo.ParseAssetID(aidStr); ok {
			v.Search = ""
			v.AssetID = aid
		}
	}
	return v
}

func toCustomsItem(e repo.EntityOut, locationPath []string) customsinventory.Item {
	it := customsinventory.Item{
		ID:            e.ID,
		AssetID:       int64(e.AssetID),
		Name:          e.Name,
		Archived:      e.Archived,
		SoldDate:      e.SoldDate,
		Quantity:      e.Quantity,
		PurchasePrice: e.PurchasePrice,
		PurchaseDate:  e.PurchaseDate,
		LocationPath:  locationPath,
	}
	if e.EntityType != nil {
		it.EntityTypeID = e.EntityType.ID
		it.EntityTypeName = e.EntityType.Name
		it.IsLocation = e.EntityType.IsLocation
	}
	if e.Location != nil {
		it.LocationID = e.Location.ID
	}
	if e.Parent != nil {
		it.ParentID = e.Parent.ID
	}
	for _, t := range e.Tags {
		it.TagIDs = append(it.TagIDs, t.ID)
	}
	for _, a := range e.Attachments {
		it.Attachments = append(it.Attachments, customsinventory.Attachment{
			ID:        a.ID,
			Type:      a.Type,
			Title:     a.Title,
			MimeType:  a.MimeType,
			CreatedAt: a.CreatedAt,
			IsLink:    a.MimeType == repo.MimeTypeLinkURL,
		})
	}
	return it
}
