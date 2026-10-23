package printer

import (
	"context"
	"fmt"
	"net"
	"time"
)

// ClientOption configures a printer client.
type ClientOption func(*clientOptions)

type clientOptions struct {
	allowPublic bool
}

// WithAllowPublicAddresses lets a client connect to hosts outside private
// networks. Off by default, matching HBOX_PRINTER_ALLOW_PUBLIC_ADDRESSES.
func WithAllowPublicAddresses(allow bool) ClientOption {
	return func(o *clientOptions) { o.allowPublic = allow }
}

func applyOptions(opts []ClientOption) clientOptions {
	var o clientOptions
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

const dialTimeout = 10 * time.Second

// safeDialContext returns a dialer that resolves the host itself and connects
// only to an address that passes the private-network check. Validating the
// stored address when the printer is saved is not enough on its own: a DNS
// record can change between validation and the connection, so the check is
// bound to the IP that is actually dialled.
func safeDialContext(allowPublic bool) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("cannot resolve printer address %q: %w", host, err)
		}

		dialer := &net.Dialer{Timeout: dialTimeout}
		var lastErr error
		for _, ip := range ips {
			if !allowPublic && !isPrivateIP(ip.IP) {
				lastErr = fmt.Errorf("printer address %q resolves to %s, which is not on a private network", host, ip.IP)
				continue
			}
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.IP.String(), port))
			if err == nil {
				return conn, nil
			}
			lastErr = err
		}
		if lastErr == nil {
			lastErr = fmt.Errorf("no addresses found for printer host %q", host)
		}
		return nil, lastErr
	}
}
