package customsinventory

import (
	"archive/zip"
	"fmt"
	"io"
	"strings"
	"time"
)

// OpenFunc opens the content of one receipt.
type OpenFunc func(ReceiptFile) (io.ReadCloser, error)

// WriteBundle streams a ZIP with the workbook at the top and every receipt
// under its generated name (FR-20). Receipts are copied, never read into memory.
func WriteBundle(w io.Writer, base string, modified time.Time, xlsx []byte, receipts []ReceiptFile, open OpenFunc) error {
	zw := zip.NewWriter(w)

	entry, err := zw.CreateHeader(&zip.FileHeader{Name: base + ".xlsx", Method: zip.Deflate, Modified: modified})
	if err != nil {
		return err
	}
	if _, err := entry.Write(xlsx); err != nil {
		return err
	}

	for _, rf := range receipts {
		if err := writeReceipt(zw, rf, modified, open); err != nil {
			return err
		}
	}
	return zw.Close()
}

func writeReceipt(zw *zip.Writer, rf ReceiptFile, modified time.Time, open OpenFunc) error {
	if strings.Contains(rf.Name, "..") || strings.HasPrefix(rf.Name, "/") || strings.Contains(rf.Name, `\`) {
		return fmt.Errorf("customs inventory: unsafe receipt name %q", rf.Name)
	}
	rc, err := open(rf)
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()

	entry, err := zw.CreateHeader(&zip.FileHeader{Name: rf.Name, Method: zip.Deflate, Modified: modified})
	if err != nil {
		return err
	}
	_, err = io.Copy(entry, rc)
	return err
}
