package mid

import (
	"net"
	"net/http"
	"slices"
	"strings"

	"github.com/rs/zerolog"
)


func TrustedIps(logger zerolog.Logger, trustedHosts []string) func(http.Handler) http.Handler {
	var trustedIps []string
	var trusting_all = false
	if slices.Contains(trustedHosts, "0.0.0.0") {
		trusting_all = true
	}
	if !trusting_all {
		for _, host := range trustedHosts {
			addrs, err := net.LookupIP(host)
			if err != nil {
				logger.Err(err)
				continue
			}
			for _, addr := range addrs {
				trustedIps = append(trustedIps, addr.String())
			}
		}
		logger.Info().Msgf("Trusted ips: %q", trustedIps)
	} else {
		logger.Info().Msgf("Trusted ips: ALL (0.0.0.0)")
	}
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				port_colon_idx := strings.LastIndex(request.RemoteAddr, ":")
				addr := request.RemoteAddr[:port_colon_idx]
				if trusting_all || slices.Contains(trustedIps, addr) {
					handler.ServeHTTP(writer, request)
				} else {
					writer.WriteHeader(http.StatusUnauthorized)
					_, err := writer.Write([]byte{})
					if err != nil {
						logger.Err(err)
					}
				}
			},
		)
	}
}