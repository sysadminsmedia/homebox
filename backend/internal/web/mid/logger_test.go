package mid

import (
	"io"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// serveSlowBody starts a server whose WriteTimeout is deliberately shorter than
// the time the handler takes to stream its body, and returns how many bytes the
// client managed to read. clearDeadline mirrors what a streaming download
// handler does before writing.
func serveSlowBody(t *testing.T, clearDeadline bool) (int64, error) {
	t.Helper()

	const (
		writeTimeout = 200 * time.Millisecond
		chunks       = 20
		chunkSize    = 32 * 1024
		chunkDelay   = 30 * time.Millisecond // 20 * 30ms = 600ms > writeTimeout
	)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if clearDeadline {
			// The fix under test: reaching the underlying connection requires
			// every wrapper in the chain to expose Unwrap.
			assert.NoError(t, http.NewResponseController(w).SetWriteDeadline(time.Time{}))
		}
		w.Header().Set("Content-Length", strconv.Itoa(chunks*chunkSize))
		w.WriteHeader(http.StatusOK)
		chunk := make([]byte, chunkSize)
		for i := 0; i < chunks; i++ {
			if _, err := w.Write(chunk); err != nil {
				return
			}
			_ = http.NewResponseController(w).Flush()
			time.Sleep(chunkDelay)
		}
	})

	srv := &http.Server{
		WriteTimeout: writeTimeout,
		// Mirror the production order: RequestID then Logger, which is the
		// middleware that wraps the ResponseWriter.
		Handler: middleware.RequestID(Logger(zerolog.Nop())(handler)),
	}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() { _ = srv.Close() })

	resp, err := http.Get("http://" + ln.Addr().String() + "/download")
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	return io.Copy(io.Discard, resp.Body)
}

// http.Server.WriteTimeout is an absolute deadline armed when the request is
// read, not an idle timeout: it does not extend as bytes are written. Streaming
// handlers therefore have to clear it, or long downloads (backups, large
// attachments) are severed mid-body and the user gets a truncated file. This is
// issue #1653.
func TestLoggerAllowsHandlerToClearWriteDeadline(t *testing.T) {
	const want = 20 * 32 * 1024

	n, err := serveSlowBody(t, true)
	require.NoError(t, err, "body should stream to completion once the deadline is cleared")
	assert.Equal(t, int64(want), n, "whole body should arrive")
}

// Control: without clearing the deadline the same response is truncated. If
// this ever stops failing, the test above has stopped proving anything.
func TestWriteTimeoutTruncatesWithoutClearing(t *testing.T) {
	const want = 20 * 32 * 1024

	n, err := serveSlowBody(t, false)
	if err == nil && n == int64(want) {
		t.Fatal("expected the response to be truncated by WriteTimeout, but it completed")
	}
}
