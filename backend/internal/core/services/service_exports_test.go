package services

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gocloud.dev/blob"

	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entity"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/predicate"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/tag"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
)

func tagInGroup(gid uuid.UUID) predicate.Tag {
	return tag.HasGroupWith(group.ID(gid))
}

// TestExportRoundTrip writes some entities into a fresh source group, runs
// the export to produce a zip artifact, and then replays that artifact into
// a separate empty destination group. Counts and selected fields are
// asserted on the destination side.
//
// This is the load-bearing integration test for the raw-SQL dump/restore
// path: anything that doesn't round-trip cleanly (timestamps, UUIDs, JSON
// columns, self-referential FKs) shows up here.
func TestExportRoundTrip(t *testing.T) {
	ctx := context.Background()

	// --- Source group with data ----------------------------------------
	src, err := tRepos.Groups.GroupCreate(ctx, "export-src-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)

	containerET, err := tRepos.EntityTypes.GetDefault(ctx, src.ID, true)
	require.NoError(t, err)
	itemET, err := tRepos.EntityTypes.GetDefault(ctx, src.ID, false)
	require.NoError(t, err)

	// One location, one item nested in it.
	loc, err := tRepos.Entities.Create(ctx, src.ID, repo.EntityCreate{
		Name:         "Garage",
		Description:  "primary",
		EntityTypeID: containerET.ID,
	})
	require.NoError(t, err)

	item, err := tRepos.Entities.Create(ctx, src.ID, repo.EntityCreate{
		Name:         "Drill",
		Description:  "cordless",
		ParentID:     loc.ID,
		EntityTypeID: itemET.ID,
	})
	require.NoError(t, err)

	// Tag and link to the item (exercises the tag_entities junction).
	tg, err := tRepos.Tags.Create(ctx, src.ID, repo.TagCreate{
		Name:        "tools",
		Description: "stuff that hits other stuff",
	})
	require.NoError(t, err)
	_, err = tClient.Entity.UpdateOneID(item.ID).AddTagIDs(tg.ID).Save(ctx)
	require.NoError(t, err)

	// Real attachment + a fabricated thumbnail row pointing at it.
	// This is the scenario that broke before: the thumbnail row has
	// entity_attachments=NULL and is reachable only via the parent's
	// attachment_thumbnail FK, so the original entity-only scope missed it.
	parentAtt, err := tRepos.Attachments.Create(ctx, item.ID,
		repo.ItemCreateAttachment{
			Title:   "manual.pdf",
			Content: bytes.NewReader([]byte("dummy pdf body")),
		},
		attachment.TypeManual, false)
	require.NoError(t, err)

	srcGroup, err := tClient.Group.Get(ctx, src.ID)
	require.NoError(t, err)
	thumbUpload, err := tRepos.Attachments.UploadFile(ctx, srcGroup,
		repo.ItemCreateAttachment{
			Title:   "manual-thumb",
			Content: bytes.NewReader([]byte("dummy thumbnail body")),
		})
	require.NoError(t, err)
	thumbAtt, err := tClient.Attachment.Create().
		SetType(attachment.TypeThumbnail).
		SetTitle("manual-thumb").
		SetPath(thumbUpload.Path).
		SetMimeType("image/webp").
		Save(ctx)
	require.NoError(t, err)
	_, err = tClient.Attachment.UpdateOneID(parentAtt.ID).SetThumbnailID(thumbAtt.ID).Save(ctx)
	require.NoError(t, err)

	// --- Export --------------------------------------------------------
	expRow, err := tRepos.Exports.Create(ctx, src.ID)
	require.NoError(t, err)

	artifactPath, sizeBytes, err := tSvc.Exports.buildArtifact(ctx, expRow.ID, src.ID)
	require.NoError(t, err)
	require.NotEmpty(t, artifactPath)
	require.Greater(t, sizeBytes, int64(0))

	// Artifact must live under the source group's prefix.
	assert.True(t, strings.HasPrefix(artifactPath, src.ID.String()+"/exports/"),
		"artifact path %q must be scoped to source group", artifactPath)

	// --- Destination: fresh group with seeded defaults -----------------
	// Mirror what real registration does: a new group has default locations
	// and tags but no items. The import must tolerate this and wipe them.
	dst, err := tRepos.Groups.GroupCreate(ctx, "export-dst-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)

	dstContainerET, err := tRepos.EntityTypes.GetDefault(ctx, dst.ID, true)
	require.NoError(t, err)
	for _, name := range []string{"Living Room", "Garage", "Kitchen"} {
		_, err := tRepos.Entities.Create(ctx, dst.ID, repo.EntityCreate{
			Name:         name,
			EntityTypeID: dstContainerET.ID,
		})
		require.NoError(t, err)
	}
	for _, name := range []string{"Appliances", "Electronics"} {
		_, err := tRepos.Tags.Create(ctx, dst.ID, repo.TagCreate{Name: name})
		require.NoError(t, err)
	}

	ready, err := tSvc.Exports.IsGroupReadyForImport(ctx, dst.ID)
	require.NoError(t, err)
	require.True(t, ready, "dst group with only seeded defaults must be importable")

	// Stage the just-built artifact as if it had been uploaded for import.
	// We re-publish it under the destination's import prefix to satisfy the
	// worker's scope check.
	importKey := dst.ID.String() + "/imports/" + uuid.New().String() + ".zip"
	require.NoError(t, copyBlobUnderTest(ctx, tSvc.Exports, artifactPath, importKey))

	// Create the tracked import row the worker reads to find the upload key
	// and to report status/progress against.
	impRow, err := tRepos.Exports.CreateImport(ctx, dst.ID, importKey, sizeBytes)
	require.NoError(t, err)
	tSvc.Exports.RunImport(ctx, dst.ID, tUser.ID, impRow.ID)

	// --- Assertions ----------------------------------------------------
	dstEntities, err := tClient.Entity.Query().Where(entity.HasGroupWith(group.ID(dst.ID))).All(ctx)
	require.NoError(t, err)
	require.Len(t, dstEntities, 2, "exactly the location and the item should remain — seeded defaults wiped, source data restored")

	gotItem, err := tClient.Entity.Query().
		Where(entity.HasGroupWith(group.ID(dst.ID)), entity.Name("Drill")).
		Only(ctx)
	require.NoError(t, err)

	parent, err := gotItem.QueryParent().Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Garage", parent.Name, "parent FK must be restored on second pass")

	tags, err := gotItem.QueryTag().All(ctx)
	require.NoError(t, err)
	require.Len(t, tags, 1, "tag_entities junction must round-trip")
	assert.Equal(t, "tools", tags[0].Name)

	// Seeded tags must be gone — only the imported "tools" tag should remain.
	allTags, err := tClient.Tag.Query().Where(tagInGroup(dst.ID)).All(ctx)
	require.NoError(t, err)
	require.Len(t, allTags, 1, "seeded tags should have been wiped")
	assert.Equal(t, "tools", allTags[0].Name)

	// IDs are intentionally regenerated on import (so re-importing the same
	// archive into a server that already has the data doesn't conflict on
	// PK). Names + relationship structure are what matters.
	assert.NotEqual(t, item.ID, gotItem.ID, "import should remap PKs")
	assert.NotEqual(t, tg.ID, tags[0].ID, "import should remap PKs")

	// Attachment + thumbnail must both round-trip with the parent→thumbnail
	// link intact and both blobs present at their new on-disk paths.
	gotAtts, err := gotItem.QueryAttachments().All(ctx)
	require.NoError(t, err)
	require.Len(t, gotAtts, 1, "parent attachment row must round-trip")

	gotThumb, err := gotAtts[0].QueryThumbnail().Only(ctx)
	require.NoError(t, err, "parent attachment must have its thumbnail edge restored")
	assert.Equal(t, "image/webp", gotThumb.MimeType)

	// Imported paths must be rewritten to the destination group's prefix —
	// otherwise the DB would point at the source group and on-delete cascade
	// would leak blobs.
	dstPrefix := dst.ID.String() + "/"
	assert.True(t, strings.HasPrefix(gotAtts[0].Path, dstPrefix),
		"parent attachment path must point at dst group (got %q)", gotAtts[0].Path)
	assert.True(t, strings.HasPrefix(gotThumb.Path, dstPrefix),
		"thumbnail path must point at dst group (got %q)", gotThumb.Path)
	assert.NotContains(t, gotAtts[0].Path, src.ID.String(),
		"source gid must not appear anywhere in the imported path")

	bk, err := blob.OpenBucket(ctx, tRepos.Attachments.GetConnString())
	require.NoError(t, err)
	defer func() { _ = bk.Close() }()

	parentBlob, err := bk.ReadAll(ctx, tRepos.Attachments.GetFullPath(gotAtts[0].Path))
	require.NoError(t, err, "parent attachment blob must be present at the rewritten path")
	assert.Equal(t, "dummy pdf body", string(parentBlob))

	thumbBlob, err := bk.ReadAll(ctx, tRepos.Attachments.GetFullPath(gotThumb.Path))
	require.NoError(t, err, "thumbnail blob must be present at the rewritten path")
	assert.Equal(t, "dummy thumbnail body", string(thumbBlob))
}

// copyBlobUnderTest reuses the export service's bucket plumbing to copy a
// blob from one key to another in the same backing store. Used to "stage"
// the just-produced export under the destination group's import prefix.
func copyBlobUnderTest(ctx context.Context, svc *ExportService, srcKey, dstKey string) error {
	att := svc.repos.Attachments
	bk, err := blob.OpenBucket(ctx, att.GetConnString())
	if err != nil {
		return err
	}
	defer func() { _ = bk.Close() }()

	r, err := bk.NewReader(ctx, att.GetFullPath(srcKey), nil)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	w, err := bk.NewWriter(ctx, att.GetFullPath(dstKey), nil)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, r); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}
