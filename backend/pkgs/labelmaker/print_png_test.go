package labelmaker

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

func printCfg(command string) *config.Config {
	return &config.Config{LabelMaker: config.LabelMakerConf{PrintCommand: &command}}
}

func TestPrintPNG_CopiesPlaceholderRunsOnce(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses cp")
	}
	dir := t.TempDir()
	png := []byte("fake-png-bytes")

	cfg := printCfg("cp {{.FileName}} " + filepath.Join(dir, "copy-{{.Copies}}.png"))
	require.NoError(t, PrintPNG(cfg, png, map[string]string{"TitleText": "x"}, 3))

	got, err := os.ReadFile(filepath.Join(dir, "copy-3.png"))
	require.NoError(t, err)
	assert.Equal(t, png, got)
}

func TestPrintPNG_RunsPerCopyWithoutPlaceholder(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses sh")
	}
	dir := t.TempDir()
	counter := filepath.Join(dir, "count")
	script := filepath.Join(dir, "run.sh")
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\necho \"$2\" >> \"$1\"\n"), 0o600))

	cfg := printCfg("sh " + script + " " + counter + " {{.TitleText}}")
	require.NoError(t, PrintPNG(cfg, []byte("x"), map[string]string{"TitleText": "run"}, 3))

	got, err := os.ReadFile(counter)
	require.NoError(t, err)
	assert.Equal(t, "run\nrun\nrun\n", string(got))
}

func TestPrintPNG_ErrorsWithoutCommand(t *testing.T) {
	require.Error(t, PrintPNG(&config.Config{}, []byte("x"), nil, 1))
	require.Error(t, PrintPNG(printCfg(""), []byte("x"), nil, 1))
}

func TestPrintPNG_RemovesTempFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses sh")
	}
	dir := t.TempDir()
	record := filepath.Join(dir, "name")
	script := filepath.Join(dir, "run.sh")
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\necho \"$2\" > \"$1\"\n"), 0o600))

	require.NoError(t, PrintPNG(printCfg("sh "+script+" "+record+" {{.FileName}}"), []byte("x"), nil, 1))

	name, err := os.ReadFile(record)
	require.NoError(t, err)
	_, err = os.Stat(strings.TrimSpace(string(name)))
	assert.True(t, os.IsNotExist(err))
}

func TestPrintCommandDefaults_NoMissingKeys(t *testing.T) {
	d := printCommandDefaults()
	for _, k := range []string{"FileName", "Width", "Height", "WidthMm", "HeightMm", "Dpi", "URL", "TitleText", "Copies"} {
		assert.Contains(t, d, k)
	}
	assert.Equal(t, "1", d["Copies"])
}
