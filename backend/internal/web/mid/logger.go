package mid

import (
	"bufio"
	"errors"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

type spy struct {
	http.ResponseWriter
	status int
}

func (s *spy) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

// Write implicitly sends a 200 status if the handler hasn't called
// WriteHeader yet, matching how http.ResponseWriter behaves. Without this
// override, a handler that writes a body without an explicit WriteHeader
// call leaves s.status at its zero value.
func (s *spy) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}

	return s.ResponseWriter.Write(b)
}

func (s *spy) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := s.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("response writer does not support hijacking")
	}
	return hj.Hijack()
}

func Logger(l zerolog.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID, _ := r.Context().Value(middleware.RequestIDKey).(string)

			l.Info().Ctx(r.Context()).Str("method", r.Method).Str("path", r.URL.Path).Str("rid", reqID).Msg("request received")

			s := &spy{ResponseWriter: w}
			h.ServeHTTP(s, r)

			l.Info().Ctx(r.Context()).Str("method", r.Method).Str("path", r.URL.Path).Int("status", s.status).Str("rid", reqID).Msg("request finished")
		})
	}
}
