package customsinventory

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteBundle(t *testing.T) {
	modified := time.Date(2026, 10, 31, 10, 0, 0, 0, time.UTC)
	receipts := []ReceiptFile{
		{Name: "Belege/000-101_Rechnung.pdf", AttachmentID: id(1)},
		{Name: "Belege/000-146_Rechnung_2.jpg", AttachmentID: id(2)},
	}
	contents := map[uuid.UUID]string{id(1): "%PDF-1.7 clock", id(2): "jpeg bytes"}
	open := func(rf ReceiptFile) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(contents[rf.AttachmentID])), nil
	}

	var buf bytes.Buffer
	require.NoError(t, WriteBundle(&buf, "Inventarliste_x_2026-10-31", modified, []byte("xlsx bytes"), receipts, open))

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)
	var names []string
	got := map[string]string{}
	for _, f := range zr.File {
		names = append(names, f.Name)
		assert.NotContains(t, f.Name, "..")
		assert.False(t, strings.HasPrefix(f.Name, "/"))
		assert.True(t, f.Modified.Equal(modified), "%s modified %v", f.Name, f.Modified)
		rc, err := f.Open()
		require.NoError(t, err)
		b, err := io.ReadAll(rc)
		require.NoError(t, err)
		_ = rc.Close()
		got[f.Name] = string(b)
	}
	assert.Equal(t, []string{"Inventarliste_x_2026-10-31.xlsx", "Belege/000-101_Rechnung.pdf", "Belege/000-146_Rechnung_2.jpg"}, names)
	assert.Equal(t, "xlsx bytes", got["Inventarliste_x_2026-10-31.xlsx"])
	assert.Equal(t, "%PDF-1.7 clock", got["Belege/000-101_Rechnung.pdf"])
	assert.Equal(t, "jpeg bytes", got["Belege/000-146_Rechnung_2.jpg"])
}

func TestWriteBundle_OpenError(t *testing.T) {
	boom := errors.New("boom")
	open := func(ReceiptFile) (io.ReadCloser, error) { return nil, boom }
	err := WriteBundle(io.Discard, "x", time.Now(), nil, []ReceiptFile{{Name: "Belege/a_Rechnung.pdf"}}, open)
	require.ErrorIs(t, err, boom)
	assert.ErrorContains(t, err, "Belege/a_Rechnung.pdf")
}

// A receipt that fails partway through must name itself in the error.
func TestWriteBundle_ReadError(t *testing.T) {
	boom := errors.New("disk read failed")
	open := func(ReceiptFile) (io.ReadCloser, error) {
		return io.NopCloser(io.MultiReader(strings.NewReader("first bytes"), iotest.ErrReader(boom))), nil
	}
	err := WriteBundle(io.Discard, "x", time.Now(), nil, []ReceiptFile{{Name: "Belege/b_Rechnung.pdf"}}, open)
	require.ErrorIs(t, err, boom)
	assert.ErrorContains(t, err, "Belege/b_Rechnung.pdf")
}

func TestWriteBundle_RejectsUnsafeNames(t *testing.T) {
	open := func(ReceiptFile) (io.ReadCloser, error) { return io.NopCloser(strings.NewReader("")), nil }
	for _, name := range []string{"../evil.pdf", "/abs.pdf", `Belege\x.pdf`} {
		err := WriteBundle(io.Discard, "x", time.Now(), nil, []ReceiptFile{{Name: name}}, open)
		assert.Error(t, err, name)
	}
}
