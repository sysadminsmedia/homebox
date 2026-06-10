package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
)

func newExternalLinkEntity(t *testing.T) repo.EntityOut {
	t.Helper()

	loc, err := tRepos.Entities.CreateContainer(testCtx(), tGroup.ID, repo.EntityCreate{Name: fk.Str(10)})
	require.NoError(t, err)

	entity, err := tRepos.Entities.Create(testCtx(), tGroup.ID, repo.EntityCreate{
		Name:     fk.Str(10),
		ParentID: loc.ID,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = tRepos.Entities.Delete(testCtx(), entity.ID)
	})

	return entity
}

func TestEntityService_AttachmentAddExternalLink_DefaultType(t *testing.T) {
	svc := &EntityService{repo: tRepos}
	entity := newExternalLinkEntity(t)

	out, err := svc.AttachmentAddExternalLink(tCtx, entity.ID, "link", "https://example.com/doc/42", "Example Doc", "")
	require.NoError(t, err)
	require.NotEmpty(t, out.Attachments)

	var found bool
	for _, att := range out.Attachments {
		if att.Path == "https://example.com/doc/42" {
			found = true
			assert.Equal(t, repo.MimeTypeLinkURL, att.MimeType)
			assert.Equal(t, "Example Doc", att.Title)
			assert.Equal(t, string(attachment.TypeAttachment), att.Type)
		}
	}
	assert.True(t, found)
}

func TestEntityService_AttachmentAddExternalLink_ManualType(t *testing.T) {
	svc := &EntityService{repo: tRepos}
	entity := newExternalLinkEntity(t)

	out, err := svc.AttachmentAddExternalLink(tCtx, entity.ID, "link", "https://example.com/manual", "Manual", attachment.TypeManual)
	require.NoError(t, err)

	var found bool
	for _, att := range out.Attachments {
		if att.Path == "https://example.com/manual" {
			found = true
			assert.Equal(t, string(attachment.TypeManual), att.Type)
		}
	}
	assert.True(t, found)
}

func TestEntityService_AttachmentAddExternalLink_WarrantyType(t *testing.T) {
	svc := &EntityService{repo: tRepos}
	entity := newExternalLinkEntity(t)

	out, err := svc.AttachmentAddExternalLink(tCtx, entity.ID, "link", "https://example.com/warranty", "Warranty", attachment.TypeWarranty)
	require.NoError(t, err)

	var found bool
	for _, att := range out.Attachments {
		if att.Path == "https://example.com/warranty" {
			found = true
			assert.Equal(t, string(attachment.TypeWarranty), att.Type)
		}
	}
	assert.True(t, found)
}

func TestEntityService_AttachmentAddExternalLink_ReceiptType(t *testing.T) {
	svc := &EntityService{repo: tRepos}
	entity := newExternalLinkEntity(t)

	out, err := svc.AttachmentAddExternalLink(tCtx, entity.ID, "link", "https://example.com/receipt", "Receipt", attachment.TypeReceipt)
	require.NoError(t, err)

	var found bool
	for _, att := range out.Attachments {
		if att.Path == "https://example.com/receipt" {
			found = true
			assert.Equal(t, string(attachment.TypeReceipt), att.Type)
		}
	}
	assert.True(t, found)
}

func TestEntityService_AttachmentAddExternalLink_InvalidEntity(t *testing.T) {
	svc := &EntityService{repo: tRepos}

	_, err := svc.AttachmentAddExternalLink(tCtx, uuid.New(), "link", "https://example.com/missing", "Missing", attachment.TypeAttachment)
	assert.Error(t, err)
}

func TestEntityService_AttachmentAddExternalLink_UnknownSourceType(t *testing.T) {
	svc := &EntityService{repo: tRepos}
	entity := newExternalLinkEntity(t)

	_, err := svc.AttachmentAddExternalLink(tCtx, entity.ID, "paperless", "42", "Paperless", attachment.TypeAttachment)
	assert.Error(t, err)
}

func TestEntityService_AttachmentDelete_ExternalLink(t *testing.T) {
	svc := &EntityService{repo: tRepos}
	entity := newExternalLinkEntity(t)

	out, err := svc.AttachmentAddExternalLink(tCtx, entity.ID, "link", "https://example.com/delete", "Delete Me", attachment.TypeAttachment)
	require.NoError(t, err)
	require.NotEmpty(t, out.Attachments)

	var createdID uuid.UUID
	for _, att := range out.Attachments {
		if att.Path == "https://example.com/delete" {
			createdID = att.ID
			break
		}
	}
	require.NotEqual(t, uuid.Nil, createdID)

	err = svc.AttachmentDelete(tCtx, tCtx.GID, createdID)
	require.NoError(t, err)

	_, err = tRepos.Attachments.Get(testCtx(), tCtx.GID, createdID)
	assert.Error(t, err)
}
