package validate

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

const outboundResolveTimeout = 5 * time.Second

// ValidateNotifierURL validates a notifier URL against the configured block/allow lists.
// This only applies to generic:// notifier URLs which can make arbitrary HTTP requests.
func ValidateNotifierURL(notifierURL string, cfg *config.NotifierConf) error {
	// Only validate generic notifiers
	if !isGenericNotifier(notifierURL) {
		return nil
	}

	// Extract the actual URL from the generic:// wrapper
	actualURL, err := extractGenericURL(notifierURL)
	if err != nil {
		return fmt.Errorf("invalid generic notifier URL: %w", err)
	}

	return ValidateOutboundHTTPURL(actualURL, cfg)
}

// ValidateOutboundHTTPURL validates an outbound HTTP(S) URL against the
// configured SSRF allow/block policy.
func ValidateOutboundHTTPURL(rawURL string, cfg *config.NotifierConf) error {
	return ValidateOutboundHTTPURLWithContext(context.Background(), rawURL, cfg)
}

// ValidateOutboundHTTPURLWithContext validates an outbound HTTP(S) URL using the
// caller's cancellation context and a bounded DNS lookup timeout.
func ValidateOutboundHTTPURLWithContext(ctx context.Context, rawURL string, cfg *config.NotifierConf) error {
	if cfg == nil {
		return fmt.Errorf("outbound URL validation configuration is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, outboundResolveTimeout)
	defer cancel()

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid outbound URL: %w", err)
	}
	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("outbound URL must use http:// or https:// scheme")
	}

	host := parsedURL.Hostname()
	if host == "" {
		return fmt.Errorf("no hostname found in URL")
	}

	ips, err := resolveOutboundHost(ctx, host)
	if err != nil {
		return err
	}

	return validateOutboundIPs(ips, cfg)
}

// NewOutboundHTTPTransport returns an HTTP transport that enforces the shared
// outbound URL policy when the transport dials the final connection target.
// Callers should still validate request URLs before sending so policy failures
// can be reported as request validation errors instead of transport errors.
// The transport deliberately enforces the supplied NotifierConf as-is so generic
// webhooks and integrations share one operator-controlled outbound policy.
func NewOutboundHTTPTransport(cfg *config.NotifierConf) *http.Transport {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	return &http.Transport{
		DialContext:           outboundDialContext(cfg, dialer),
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func outboundDialContext(cfg *config.NotifierConf, dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	if dialer == nil {
		dialer = &net.Dialer{}
	}

	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, fmt.Errorf("invalid outbound dial address %q: %w", address, err)
		}

		ips, err := resolveOutboundHost(ctx, host)
		if err != nil {
			return nil, err
		}

		filteredIPs := filterOutboundIPsByNetwork(network, ips)
		if len(filteredIPs) == 0 {
			return nil, fmt.Errorf("no resolved IPs for %s match network %s", host, network)
		}
		if err := validateOutboundIPs(filteredIPs, cfg); err != nil {
			return nil, fmt.Errorf("outbound dial blocked: %w", err)
		}

		var lastErr error
		for _, ip := range filteredIPs {
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
			if err == nil {
				return conn, nil
			}
			lastErr = err
		}

		if lastErr != nil {
			return nil, fmt.Errorf("outbound dial failed for %s: %w", host, lastErr)
		}
		return nil, fmt.Errorf("no resolved IPs for %s match network %s", host, network)
	}
}

func filterOutboundIPsByNetwork(network string, ips []net.IP) []net.IP {
	filtered := make([]net.IP, 0, len(ips))
	for _, ip := range ips {
		if ipMatchesNetwork(network, ip) {
			filtered = append(filtered, ip)
		}
	}
	return filtered
}

func resolveOutboundHost(ctx context.Context, host string) ([]net.IP, error) {
	if ip := net.ParseIP(host); ip != nil {
		return []net.IP{ip}, nil
	}

	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve hostname: %w", err)
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("hostname did not resolve to any IP addresses")
	}

	ips := make([]net.IP, 0, len(addrs))
	for _, addr := range addrs {
		ips = append(ips, addr.IP)
	}
	return ips, nil
}

func ipMatchesNetwork(network string, ip net.IP) bool {
	switch network {
	case "tcp4":
		return ip.To4() != nil
	case "tcp6":
		return ip.To4() == nil
	default:
		return true
	}
}

func validateOutboundIPs(ips []net.IP, cfg *config.NotifierConf) error {
	if cfg == nil {
		return fmt.Errorf("outbound URL validation configuration is nil")
	}

	// Expand DNS64-synthesized IPv6 addresses (RFC 6052) into their embedded
	// IPv4 addresses so the allow/block rules below are applied to the IPv4
	// destination the NAT64 gateway will actually reach. The original IPv6
	// address stays in the list so IPv6 rules still apply to it.
	checkIPs := make([]net.IP, 0, len(ips))
	for _, ip := range ips {
		checkIPs = append(checkIPs, ip)
		embedded, inDNS64Range := dns64EmbeddedIPv4s(ip, cfg.Dns64Nets)
		if inDNS64Range && len(embedded) == 0 {
			return fmt.Errorf("IP %s is in a DNS64 range but no valid embedded IPv4 address could be extracted", ip.String())
		}
		checkIPs = append(checkIPs, embedded...)
	}
	ips = checkIPs

	// If AllowNets is configured it acts as an allowlist: every IP must match,
	// and passing skips the remaining block checks.
	if len(cfg.AllowNets) > 0 {
		return checkAllowNets(ips, cfg.AllowNets)
	}

	// Check BlockNets - block specific networks if configured
	if len(cfg.BlockNets) > 0 {
		if err := checkBlockNets(ips, cfg.BlockNets); err != nil {
			return err
		}
	}

	return checkBlockedCategories(ips, cfg)
}

// checkAllowNets verifies every IP falls within one of the allowNets. A nil
// return means validation is complete and no further block checks are needed.
func checkAllowNets(ips []net.IP, allowNets []string) error {
	for _, ip := range ips {
		allowed := false
		for _, allowNet := range allowNets {
			_, ipNet, err := net.ParseCIDR(allowNet)
			if err != nil {
				log.Warn().
					Err(err).
					Str("cidr", allowNet).
					Str("config", "AllowNets").
					Msg("invalid CIDR in notifier AllowNets configuration, skipping")
				continue // Skip invalid CIDR
			}
			if ipNet.Contains(ip) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("IP %s is not in the allowed networks", ip.String())
		}
	}
	return nil
}

// checkBlockNets returns an error if any IP falls within one of the blockNets.
func checkBlockNets(ips []net.IP, blockNets []string) error {
	for _, ip := range ips {
		for _, blockNet := range blockNets {
			_, ipNet, err := net.ParseCIDR(blockNet)
			if err != nil {
				log.Warn().
					Err(err).
					Str("cidr", blockNet).
					Str("config", "BlockNets").
					Msg("invalid CIDR in notifier BlockNets configuration, skipping")
				continue // Skip invalid CIDR
			}
			if ipNet.Contains(ip) {
				return fmt.Errorf("IP %s is in a blocked network (%s)", ip.String(), blockNet)
			}
		}
	}
	return nil
}

// checkBlockedCategories applies the configured category-based blocks
// (localhost, RFC1918, bogon, cloud metadata) to every IP.
func checkBlockedCategories(ips []net.IP, cfg *config.NotifierConf) error {
	for _, ip := range ips {
		if cfg.BlockLocalhost && isLocalhost(ip) {
			return fmt.Errorf("localhost addresses are blocked")
		}
		if cfg.BlockLocalNets && isPrivateNetwork(ip) {
			return fmt.Errorf("private network addresses (RFC1918) are blocked")
		}
		if cfg.BlockBogonNets && isBogonNetwork(ip) {
			return fmt.Errorf("bogon/reserved network addresses are blocked")
		}
		if cfg.BlockCloudMetadata && isCloudMetadata(ip) {
			return fmt.Errorf("cloud metadata endpoints are blocked")
		}
	}
	return nil
}

// isGenericNotifier checks if the URL is a generic notifier that needs validation
func isGenericNotifier(notifierURL string) bool {
	return strings.HasPrefix(notifierURL, "generic://") ||
		strings.HasPrefix(notifierURL, "generic+https://") ||
		strings.HasPrefix(notifierURL, "generic+http://")
}

// extractGenericURL extracts the actual HTTP(S) URL from a generic notifier URL
func extractGenericURL(notifierURL string) (string, error) {
	if strings.HasPrefix(notifierURL, "generic://") {
		rawURL := strings.TrimPrefix(notifierURL, "generic://")

		if rawURL == "" {
			return "", fmt.Errorf("generic notifier URL is empty")
		}

		// Support shorthand generic://host/path by defaulting to HTTPS.
		if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
			return rawURL, nil
		}

		return "https://" + rawURL, nil
	}
	if strings.HasPrefix(notifierURL, "generic+https://") {
		return strings.TrimPrefix(notifierURL, "generic+"), nil
	}
	if strings.HasPrefix(notifierURL, "generic+http://") {
		return strings.TrimPrefix(notifierURL, "generic+"), nil
	}
	return "", fmt.Errorf("not a generic notifier URL")
}

// rfc6052PrefixLens are the prefix lengths at which RFC 6052 permits embedding
// an IPv4 address into an IPv6 address.
var rfc6052PrefixLens = []int{32, 40, 48, 56, 64, 96}

// embeddedIPv4 extracts the IPv4 address embedded in ip using the RFC 6052
// layout for the given prefix length. It returns nil if the address is not
// well-formed at that layout: the "u" octet (bits 64-71) and the suffix bits
// after the embedded address must both be zero.
func embeddedIPv4(ip net.IP, prefixLen int) net.IP {
	b := ip.To16()
	if b == nil {
		return nil
	}

	var v4 [4]byte
	var suffix []byte
	switch prefixLen {
	case 32:
		copy(v4[:], b[4:8])
		suffix = b[9:]
	case 40:
		copy(v4[:3], b[5:8])
		v4[3] = b[9]
		suffix = b[10:]
	case 48:
		copy(v4[:2], b[6:8])
		copy(v4[2:], b[9:11])
		suffix = b[11:]
	case 56:
		v4[0] = b[7]
		copy(v4[1:], b[9:12])
		suffix = b[12:]
	case 64:
		copy(v4[:], b[9:13])
		suffix = b[13:]
	case 96:
		copy(v4[:], b[12:16])
	default:
		return nil
	}

	if prefixLen < 96 && b[8] != 0 {
		return nil
	}
	for _, sb := range suffix {
		if sb != 0 {
			return nil
		}
	}

	return net.IPv4(v4[0], v4[1], v4[2], v4[3])
}

// dns64EmbeddedIPv4s returns the IPv4 addresses embedded in ip when it falls
// inside one of the configured DNS64/NAT64 prefixes, along with whether the
// address is inside any such prefix at all. A configured prefix may be shorter
// than the prefix the NAT64 gateway actually translates with (e.g. the RFC 8215
// 64:ff9b:1::/48 space holds deployment prefixes of /48 through /96), so every
// RFC 6052 layout at or beyond the configured length is tried and all
// well-formed extractions are returned for checking.
func dns64EmbeddedIPv4s(ip net.IP, dns64Nets []string) ([]net.IP, bool) {
	if ip.To4() != nil {
		return nil, false
	}

	inRange := false
	var candidates []net.IP
	for _, dns64Net := range dns64Nets {
		_, ipNet, err := net.ParseCIDR(dns64Net)
		if err != nil {
			log.Warn().
				Err(err).
				Str("cidr", dns64Net).
				Str("config", "Dns64Nets").
				Msg("invalid CIDR in notifier Dns64Nets configuration, skipping")
			continue
		}
		if !ipNet.Contains(ip) {
			continue
		}
		inRange = true

		prefixLen, _ := ipNet.Mask.Size()
		for _, layoutLen := range rfc6052PrefixLens {
			if layoutLen < prefixLen {
				continue
			}
			if v4 := embeddedIPv4(ip, layoutLen); v4 != nil {
				candidates = append(candidates, v4)
			}
		}
	}

	return candidates, inRange
}

// isLocalhost checks if an IP is a localhost address
func isLocalhost(ip net.IP) bool {
	return ip.IsLoopback()
}

// isPrivateNetwork checks if an IP is in private address space.
// This uses the standard library's IsPrivate() which covers:
// - IPv4: RFC1918 (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)
// - IPv6: Unique Local Addresses (fc00::/7)
func isPrivateNetwork(ip net.IP) bool {
	return ip.IsPrivate()
}

// isBogonNetwork checks if an IP is in reserved/bogon address space
func isBogonNetwork(ip net.IP) bool {
	// Separate IPv4 and IPv6 bogon ranges to avoid cross-version matching
	ipv4BogonNetworks := []string{
		"0.0.0.0/8",          // Current network
		"100.64.0.0/10",      // Shared Address Space (RFC6598)
		"169.254.0.0/16",     // Link-local
		"192.0.0.0/24",       // IETF Protocol Assignments
		"192.0.2.0/24",       // TEST-NET-1
		"198.18.0.0/15",      // Benchmarking
		"198.51.100.0/24",    // TEST-NET-2
		"203.0.113.0/24",     // TEST-NET-3
		"224.0.0.0/4",        // Multicast
		"240.0.0.0/4",        // Reserved
		"255.255.255.255/32", // Broadcast
	}

	ipv6BogonNetworks := []string{
		"::/128",        // Unspecified
		"::1/128",       // Loopback
		"::ffff:0:0/96", // IPv4-mapped
		"100::/64",      // Discard prefix
		"2001::/32",     // TEREDO
		"2001:10::/28",  // ORCHID
		"2001:db8::/32", // Documentation
		"fc00::/7",      // Unique local
		"fe80::/10",     // Link-local
		"ff00::/8",      // Multicast
	}

	// Determine if IP is IPv4 or IPv6 and check against appropriate list
	ipv4 := ip.To4()
	if ipv4 != nil {
		// This is an IPv4 address, check against IPv4 bogon ranges
		for _, cidr := range ipv4BogonNetworks {
			_, ipNet, err := net.ParseCIDR(cidr)
			if err != nil {
				log.Warn().
					Err(err).
					Str("cidr", cidr).
					Str("check", "IPv4BogonNetworks").
					Msg("invalid CIDR in hardcoded bogon networks")
				continue
			}
			if ipNet.Contains(ip) {
				return true
			}
		}
	} else {
		// This is an IPv6 address, check against IPv6 bogon ranges
		for _, cidr := range ipv6BogonNetworks {
			_, ipNet, err := net.ParseCIDR(cidr)
			if err != nil {
				log.Warn().
					Err(err).
					Str("cidr", cidr).
					Str("check", "IPv6BogonNetworks").
					Msg("invalid CIDR in hardcoded bogon networks")
				continue
			}
			if ipNet.Contains(ip) {
				return true
			}
		}
	}

	return false
}

// isCloudMetadata checks if an IP is a known cloud metadata endpoint
func isCloudMetadata(ip net.IP) bool {
	metadataAddresses := []string{
		"169.254.169.254/32", // AWS, Azure, GCP, Oracle Cloud
		"169.254.169.253/32", // AWS IMDSv2 alternative
		"fd00:ec2::254/128",  // AWS IPv6 metadata
	}

	for _, cidr := range metadataAddresses {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			log.Warn().
				Err(err).
				Str("cidr", cidr).
				Str("check", "CloudMetadata").
				Msg("invalid CIDR in hardcoded cloud metadata addresses")
			continue
		}
		if ipNet.Contains(ip) {
			return true
		}
	}

	return false
}
