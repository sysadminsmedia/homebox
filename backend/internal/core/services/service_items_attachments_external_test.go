package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
)

func newExternalLinkEntity(t *testing.T) repo.EntityOut {
	t.Helper()

	loc, err := tRepos.Entities.CreateContainer(context.Background(), tGroup.ID, repo.EntityCreate{Name: fk.Str(10)})
	require.NoError(t, err)

	entity, err := tRepos.Entities.Create(context.Background(), tGroup.ID, repo.EntityCreate{
		Name:     fk.Str(10),
		ParentID: loc.ID,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = tRepos.Entities.Delete(context.Background(), entity.ID)
	})

	return entity
}

var knownSources = []struct {
	sourceType string
	externalID string
}{
	{"link", "https://example.com/doc"},
}

func TestRedactExternalIdentifierForTrace(t *testing.T) {
	assert.Equal(t, "42", redactExternalIdentifierForTrace("paperless", "42"))
	assert.Equal(
		t,
		"https://host:a379a6f6eeafb9a5/path:baf62c0924b5c759",
		redactExternalIdentifierForTrace("link", "https://user:secret@example.com/doc?token=secret#fragment"),
	)
}

func TestEntityService_AttachmentAddExternalLink_SourceTypes(t *testing.T) {
	svc := &EntityService{repo: tRepos}

	for _, src := range knownSources {
		t.Run(src.sourceType, func(t *testing.T) {
			entity := newExternalLinkEntity(t)

			expectedMime, ok := repo.MimeTypeForSourceType(src.sourceType)
			require.True(t, ok, "knownSources entry %q has no registered mime type", src.sourceType)

			out, err := svc.AttachmentAddExternalLink(tCtx, entity.ID, src.sourceType, src.externalID, "Test Doc", attachment.TypeAttachment)
			require.NoError(t, err)
			require.NotEmpty(t, out.Attachments)

			var found bool
			for _, att := range out.Attachments {
				if att.Path == src.externalID {
					found = true
					assert.Equal(t, expectedMime, att.MimeType)
					assert.Equal(t, "Test Doc", att.Title)
					assert.Equal(t, string(attachment.TypeAttachment), att.Type)
				}
			}
			assert.True(t, found, "expected attachment with path %q in entity output", src.externalID)
		})
	}
}

func TestEntityService_AttachmentAddExternalLink_AttachmentTypes(t *testing.T) {
	src := knownSources[0]
	expectedMime, _ := repo.MimeTypeForSourceType(src.sourceType)

	cases := []struct {
		attType attachment.Type
		title   string
	}{
		{attachment.TypeManual, "Manual"},
		{attachment.TypeWarranty, "Warranty"},
		{attachment.TypeReceipt, "Receipt"},
	}

	svc := &EntityService{repo: tRepos}

	for _, tc := range cases {
		t.Run(string(tc.attType), func(t *testing.T) {
			entity := newExternalLinkEntity(t)

			out, err := svc.AttachmentAddExternalLink(tCtx, entity.ID, src.sourceType, src.externalID, tc.title, tc.attType)
			require.NoError(t, err)

			var found bool
			for _, att := range out.Attachments {
				if att.Path == src.externalID {
					found = true
					assert.Equal(t, string(tc.attType), att.Type)
					assert.Equal(t, expectedMime, att.MimeType)
				}
			}
			assert.True(t, found)
		})
	}
}

func TestEntityService_AttachmentAddExternalLink_MultipleAttachments(t *testing.T) {
	svc := &EntityService{repo: tRepos}
	entity := newExternalLinkEntity(t)

	for _, src := range knownSources {
		_, err := svc.AttachmentAddExternalLink(tCtx, entity.ID, src.sourceType, src.externalID, src.sourceType+" doc", attachment.TypeAttachment)
		require.NoError(t, err, "failed for source type %q", src.sourceType)
	}

	latest, err := svc.repo.Entities.GetOneByGroup(tCtx, tCtx.GID, entity.ID)
	require.NoError(t, err)
	assert.Len(t, latest.Attachments, len(knownSources))
}

func TestEntityService_AttachmentAddExternalLink_InvalidEntity(t *testing.T) {
	svc := &EntityService{repo: tRepos}
	src := knownSources[0]

	_, err := svc.AttachmentAddExternalLink(tCtx, uuid.New(), src.sourceType, src.externalID, "Missing", attachment.TypeAttachment)
	assert.Error(t, err)
}

func TestEntityService_AttachmentAddExternalLink_UnknownSourceType(t *testing.T) {
	svc := &EntityService{repo: tRepos}
	entity := newExternalLinkEntity(t)

	_, err := svc.AttachmentAddExternalLink(tCtx, entity.ID, "unknown-source", "42", "Unknown", attachment.TypeAttachment)
	assert.Error(t, err)
}

func TestEntityService_AttachmentDelete_ExternalLink(t *testing.T) {
	svc := &EntityService{repo: tRepos}
	entity := newExternalLinkEntity(t)
	src := knownSources[0]

	out, err := svc.AttachmentAddExternalLink(tCtx, entity.ID, src.sourceType, src.externalID, "Delete Me", attachment.TypeAttachment)
	require.NoError(t, err)
	require.NotEmpty(t, out.Attachments)

	var createdID uuid.UUID
	for _, att := range out.Attachments {
		if att.Path == src.externalID {
			createdID = att.ID
			break
		}
	}
	require.NotEqual(t, uuid.Nil, createdID)

	err = svc.AttachmentDelete(tCtx, tCtx.GID, createdID)
	require.NoError(t, err)

	_, err = tRepos.Attachments.Get(context.Background(), tCtx.GID, createdID)
	assert.Error(t, err)
}
