package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteBackupCopyResetsPermissionsAndSymlinks(t *testing.T) {
	dir := t.TempDir()
	bak := filepath.Join(dir, "env.bak")
	require.NoError(t, os.WriteFile(bak, []byte("old"), 0o644))
	require.NoError(t, writeBackupCopy(bak, []byte("new")))
	st, err := os.Stat(bak)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), st.Mode().Perm())

	target := filepath.Join(dir, "target")
	require.NoError(t, os.WriteFile(target, []byte("keep"), 0o644))
	require.NoError(t, os.Remove(bak))
	require.NoError(t, os.Symlink(target, bak))
	require.NoError(t, writeBackupCopy(bak, []byte("secret")))
	got, _ := os.ReadFile(target)
	require.Equal(t, "keep", string(got), "the symlink target is untouched")
}
