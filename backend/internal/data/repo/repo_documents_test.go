package repo

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
)

func useDocs(t *testing.T, num int) []DocumentOut {
	t.Helper()

	results := make([]DocumentOut, 0, num)
	ids := make([]uuid.UUID, 0, num)

	for i := 0; i < num; i++ {
		doc, err := tRepos.Docs.Create(context.Background(), tGroup.ID, DocumentCreate{
			Title:   fk.Str(10) + ".md",
			Content: bytes.NewReader([]byte(fk.Str(10))),
		})

		require.NoError(t, err)
		assert.NotNil(t, doc)
		results = append(results, doc)
		ids = append(ids, doc.ID)
	}

	t.Cleanup(func() {
		for _, id := range ids {
			err := tRepos.Docs.Delete(context.Background(), id)
			if err != nil {
				assert.True(t, ent.IsNotFound(err))
			}
		}
	})

	return results
}

func TestDocumentRepository_CreateUpdateDelete(t *testing.T) {
	temp := t.TempDir()
	r := DocumentRepository{
		db:         tClient,
		storePath:  temp,
		connString: "file:///",
	}

	type args struct {
		ctx context.Context
		gid uuid.UUID
		doc DocumentCreate
	}
	tests := []struct {
		name    string
		content string
		args    args
		title   string
		wantErr bool
	}{
		{
			name:    "basic create",
			title:   "test.md",
			content: "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
			args: args{
				ctx: context.Background(),
				gid: tGroup.ID,
				doc: DocumentCreate{
					Title:   "test.md",
					Content: bytes.NewReader([]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit.")),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Document
			got, err := r.Create(tt.args.ctx, tt.args.gid, tt.args.doc)
			require.NoError(t, err)
			assert.Equal(t, tt.title, got.Title)
			assert.Equal(t, fmt.Sprintf("%s/%s/documents", temp, tt.args.gid), filepath.Dir(got.Path))

			ensureRead := func() {
				// Read Document
				bts, err := os.ReadFile(got.Path)
				require.NoError(t, err)
				assert.Equal(t, tt.content, string(bts))
			}
			ensureRead()

			// Update Document
			got, err = r.Rename(tt.args.ctx, got.ID, "__"+tt.title+"__")
			require.NoError(t, err)
			assert.Equal(t, "__"+tt.title+"__", got.Title)

			ensureRead()

			// Delete Document
			err = r.Delete(tt.args.ctx, got.ID)
			require.NoError(t, err)

			_, err = os.Stat(got.Path)
			require.Error(t, err)
		})
	}
}
