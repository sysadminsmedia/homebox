package labelmaker

import (
	"bytes"
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

// skipOnWindows skips shell-based tests on Windows.
func skipOnWindows(t *testing.T, reason string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip(reason)
	}
}

func printCfg(command string) *config.Config {
	return &config.Config{LabelMaker: config.LabelMakerConf{PrintCommand: &command}}
}

func TestPrintPNG_CopiesPlaceholderRunsOnce(t *testing.T) {
	skipOnWindows(t, "uses cp")
	dir := t.TempDir()
	png := []byte("fake-png-bytes")

	cfg := printCfg("cp {{.FileName}} " + filepath.Join(dir, "copy-{{.Copies}}.png"))
	require.NoError(t, PrintPNG(context.Background(), cfg, png, map[string]string{"TitleText": "x"}, 3))

	got, err := os.ReadFile(filepath.Join(dir, "copy-3.png"))
	require.NoError(t, err)
	assert.Equal(t, png, got)
}

func TestPrintPNG_RunsPerCopyWithoutPlaceholder(t *testing.T) {
	skipOnWindows(t, "uses sh")
	dir := t.TempDir()
	counter := filepath.Join(dir, "count")
	script := filepath.Join(dir, "run.sh")
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\necho \"$2\" >> \"$1\"\n"), 0o600))

	cfg := printCfg("sh " + script + " " + counter + " {{.TitleText}}")
	require.NoError(t, PrintPNG(context.Background(), cfg, []byte("x"), map[string]string{"TitleText": "run"}, 3))

	got, err := os.ReadFile(counter)
	require.NoError(t, err)
	assert.Equal(t, "run\nrun\nrun\n", string(got))
}

func TestPrintPNG_ErrorsWithoutCommand(t *testing.T) {
	require.Error(t, PrintPNG(context.Background(), &config.Config{}, []byte("x"), nil, 1))
	require.Error(t, PrintPNG(context.Background(), printCfg(""), []byte("x"), nil, 1))
}

func TestPrintPNG_RemovesTempFile(t *testing.T) {
	skipOnWindows(t, "uses sh")
	dir := t.TempDir()
	record := filepath.Join(dir, "name")
	script := filepath.Join(dir, "run.sh")
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\necho \"$2\" > \"$1\"\n"), 0o600))

	require.NoError(t, PrintPNG(context.Background(), printCfg("sh "+script+" "+record+" {{.FileName}}"), []byte("x"), nil, 1))

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

func TestPrintPNG_StopsWhenContextCancelled(t *testing.T) {
	skipOnWindows(t, "uses sh")
	dir := t.TempDir()
	counter := filepath.Join(dir, "count")
	script := filepath.Join(dir, "run.sh")
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\necho x >> \"$1\"\n"), 0o600))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.Error(t, PrintPNG(ctx, printCfg("sh "+script+" "+counter), []byte("x"), nil, 3))
	_, err := os.Stat(counter)
	assert.True(t, os.IsNotExist(err), "no copy must run once the context is cancelled")
}

func TestRunPrintCommand_TimesOut(t *testing.T) {
	skipOnWindows(t, "uses sleep")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	err := runPrintCommand(ctx, printCfg("sleep 5"), map[string]string{})
	require.Error(t, err)
	// The caller's deadline ended the command, so the error says so instead of
	// claiming the 60 s command timeout fired.
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Contains(t, err.Error(), "aborted")
}

func TestBoundedBuffer_KeepsOnlyTheStart(t *testing.T) {
	b := &boundedBuffer{limit: 8}
	n, err := b.Write([]byte("12345"))
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "12345", string(b.Bytes()))
	assert.Empty(t, b.suffix())

	// Writes always report full success so the child never sees a short write.
	n, err = b.Write([]byte("678910"))
	require.NoError(t, err)
	assert.Equal(t, 6, n)
	assert.Equal(t, "12345678", string(b.Bytes()))
	assert.Equal(t, " (truncated)", b.suffix())

	n, err = b.Write([]byte("more"))
	require.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, "12345678", string(b.Bytes()))
}

func TestRunPrintCommand_FailureLogsBoundedEscapedOutput(t *testing.T) {
	skipOnWindows(t, "uses sh")
	var logged bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&logged)
	t.Cleanup(func() { log.SetOutput(prev) })

	// The print command is split on whitespace without shell quoting, so the
	// shell pipeline lives in a script: 100 kB of "y\n" lines, far past the
	// cap, then a non-zero exit.
	script := filepath.Join(t.TempDir(), "noisy-fail.sh")
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\nyes | head -c 100000\nexit 3\n"), 0o700))
	err := runPrintCommand(context.Background(), printCfg(script), map[string]string{})
	require.Error(t, err)

	out := logged.String()
	assert.Contains(t, out, "Could not print label on server")
	assert.Contains(t, out, "(truncated)")
	// The output is quoted, so its newlines are escaped and the entry stays one line.
	assert.Contains(t, out, `y\ny`)
	assert.Equal(t, 1, strings.Count(out, "\n"), "log entry must be a single line: %q", out)
	assert.Less(t, len(out), maxPrintCommandOutput*2, "log entry must be bounded")
}
