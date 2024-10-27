// Package migrations provides a way to embed the migrations into the binary.
package migrations

import (
	"embed"
	"os"
	"path"
)

//go:embed all:sqlite3
var Files embed.FS

// Write writes the embedded migrations to a temporary directory.
// It returns an error and a cleanup function. The cleanup function
// should be called when the migrations are no longer needed.
func Write(temp string, dialect string) error {
	err := os.MkdirAll(temp, 0o755)
	if err != nil {
		return err
	}

	fsDir, err := Files.ReadDir(dialect)
	if err != nil {
		return err
	}

	for _, f := range fsDir {
		if f.IsDir() {
			continue
		}

		b, err := Files.ReadFile(path.Join(dialect, f.Name()))
		if err != nil {
			return err
		}

		err = os.WriteFile(path.Join(temp, f.Name()), b, 0o644)
		if err != nil {
			return err
		}
	}

	return nil
}
