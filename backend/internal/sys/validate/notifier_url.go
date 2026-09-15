package validate

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

// genericScheme is the shoutrrr service name for the generic webhook service.
const genericScheme = "generic"

// notifierHostSource describes where a notifier scheme's network destination
// comes from, which determines whether the SSRF policy has a host to check.
type notifierHostSource int

const (
	// hostIsIdentifier means the URL's host component is an account, token,
	// channel or webhook identifier rather than a hostname. The service always
	// connects to a fixed vendor endpoint, so no user-controlled destination
	// reaches the network layer and there is nothing for the policy to check.
	hostIsIdentifier notifierHostSource = iota
	// hostFromURL means the URL's host component is the server the service
	// connects to, chosen freely by whoever supplied the URL.
	hostFromURL
	// hostFromWrappedURL means the scheme wraps a full HTTP(S) URL whose host
	// is the destination (generic://).
	hostFromWrappedURL
)

// notifierSchemes classifies every service in shoutrrr's registry by where its
// network destination comes from. Anything not listed here is rejected: failing
// closed means a service added by a future shoutrrr upgrade cannot silently
// escape the SSRF policy, at the cost of needing an entry added here first.
//
// The policy previously applied only to generic://, on the assumption that it
// was the only scheme able to reach an arbitrary host. That was wrong — gotify,
// ntfy, mattermost, matrix and others all take a fully user-supplied host, and
// several accept a disabletls/scheme parameter that downgrades the connection to
// plaintext HTTP against that host. Those URLs bypassed the allow/block lists
// entirely, including the block_bogon_nets and block_cloud_metadata rules that
// are enabled by default.
var notifierSchemes = map[string]notifierHostSource{
	genericScheme: hostFromWrappedURL,

	// The URL host is the server the notifier connects to.
	"bark":       hostFromURL,
	"googlechat": hostFromURL,
	"gotify":     hostFromURL,
	"hangouts":   hostFromURL,
	"lark":       hostFromURL,
	"matrix":     hostFromURL,
	"mattermost": hostFromURL,
	"mqtt":       hostFromURL,
	"mqtts":      hostFromURL,
	"ntfy":       hostFromURL,
	"opsgenie":   hostFromURL,
	"pagerduty":  hostFromURL,
	"rocketchat": hostFromURL,
	"signal":     hostFromURL,
	"smtp":       hostFromURL,
	"zulip":      hostFromURL,

	// The URL host is an identifier; the service talks to a fixed vendor
	// endpoint (teams takes its webhook URL from a query parameter that
	// shoutrrr itself restricts to Microsoft workflow domains).
	"discord":    hostIsIdentifier,
	"ifttt":      hostIsIdentifier,
	"join":       hostIsIdentifier,
	"logger":     hostIsIdentifier,
	"notifiarr":  hostIsIdentifier,
	"pushbullet": hostIsIdentifier,
	"pushover":   hostIsIdentifier,
	"slack":      hostIsIdentifier,
	"teams":      hostIsIdentifier,
	"telegram":   hostIsIdentifier,
	"twilio":     hostIsIdentifier,
	"wecom":      hostIsIdentifier,
}

// ValidateNotifierURL validates a notifier URL against the configured block/allow
// lists. Every scheme that carries a user-supplied network destination is checked,
// not just generic://, and unknown schemes are rejected outright.
func ValidateNotifierURL(notifierURL string, cfg *config.NotifierConf) error {
	// Defensively guard against nil cfg
	if cfg == nil {
		return fmt.Errorf("notifier configuration is nil, cannot validate URL")
	}

	scheme, _, ok := splitScheme(notifierURL)
	if !ok {
		return fmt.Errorf("notifier URL is missing a scheme")
	}

	// shoutrrr routes on the part before the "+" (generic+http -> generic).
	service, _, _ := strings.Cut(scheme, "+")

	source, known := notifierSchemes[service]
	if !known {
		return fmt.Errorf("unsupported notifier scheme %q", service)
	}

	switch source {
	case hostFromWrappedURL:
		// Extract the actual URL from the generic:// wrapper
		actualURL, err := extractGenericURL(notifierURL)
		if err != nil {
			return fmt.Errorf("invalid generic notifier URL: %w", err)
		}

		// Parse the URL to extract the hostname
		parsedURL, err := url.Parse(actualURL)
		if err != nil {
			return fmt.Errorf("invalid URL in generic notifier: %w", err)
		}

		return validateHostAgainstPolicy(parsedURL.Hostname(), cfg)

	case hostFromURL:
		parsedURL, err := url.Parse(notifierURL)
		if err != nil {
			return fmt.Errorf("invalid notifier URL: %w", err)
		}

		host := parsedURL.Hostname()
		if host == "" {
			// No host means the service falls back to its own vendor default
			// endpoint, so there is no user-controlled destination to check.
			return nil
		}

		return validateHostAgainstPolicy(host, cfg)

	case hostIsIdentifier:
		return nil
	}

	return nil
}

// splitScheme splits a notifier URL into its lower-cased scheme and the remainder
// after "://".
//
// URL schemes are case-insensitive (RFC 3986 section 3.1), and both net/url and
// shoutrrr's router lower-case the scheme before dispatching, so the policy has to
// match on the same normalized form. Matching the raw string meant "generic+HTTP://"
// routed to the generic webhook service while every prefix test here returned false,
// skipping validation entirely.
func splitScheme(notifierURL string) (scheme, rest string, ok bool) {
	i := strings.Index(notifierURL, "://")
	if i < 0 {
		return "", "", false
	}

	return strings.ToLower(notifierURL[:i]), notifierURL[i+len("://"):], true
}

// validateHostAgainstPolicy resolves host and checks every resolved IP (plus any
// DNS64-embedded IPv4) against the configured allow/block policy. It is shared by
// ValidateNotifierURL (the initial notifier URL) and NotifierRedirectGuard (each
// redirect hop at delivery time) so both enforce identical rules — the redirect
// path previously escaped these checks entirely, which was the SSRF bypass.
func validateHostAgainstPolicy(host string, cfg *config.NotifierConf) error {
	_, err := resolveHostForPolicy(context.Background(), host, cfg)

	return err
}

// resolveHostForPolicy resolves host, enforces the allow/block policy against every
// resolved address, and returns the addresses that were checked so a caller can dial
// exactly what it validated. Dialing the returned addresses rather than re-resolving
// the name is what closes the DNS-rebinding window between validation and delivery.
func resolveHostForPolicy(ctx context.Context, host string, cfg *config.NotifierConf) ([]net.IP, error) {
	if cfg == nil {
		return nil, fmt.Errorf("notifier configuration is nil, cannot validate URL")
	}

	if host == "" {
		return nil, fmt.Errorf("no hostname found in URL")
	}

	// Resolve the hostname to an IP address
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve hostname: %w", err)
	}

	if len(addrs) == 0 {
		return nil, fmt.Errorf("hostname did not resolve to any IP addresses")
	}

	resolved := make([]net.IP, 0, len(addrs))
	for _, addr := range addrs {
		resolved = append(resolved, addr.IP)
	}

	// Expand DNS64-synthesized IPv6 addresses (RFC 6052) into their embedded
	// IPv4 addresses so the allow/block rules below are applied to the IPv4
	// destination the NAT64 gateway will actually reach. The original IPv6
	// address stays in the list so IPv6 rules still apply to it.
	checkIPs := make([]net.IP, 0, len(resolved))
	for _, ip := range resolved {
		checkIPs = append(checkIPs, ip)
		embedded, inDNS64Range := dns64EmbeddedIPv4s(ip, cfg.Dns64Nets)
		if inDNS64Range && len(embedded) == 0 {
			return nil, fmt.Errorf("IP %s is in a DNS64 range but no valid embedded IPv4 address could be extracted", ip.String())
		}
		checkIPs = append(checkIPs, embedded...)
	}

	// If AllowNets is configured it acts as an allowlist: every IP must match,
	// and passing skips the remaining block checks.
	if len(cfg.AllowNets) > 0 {
		if err := checkAllowNets(checkIPs, cfg.AllowNets); err != nil {
			return nil, err
		}

		return resolved, nil
	}

	// Check BlockNets - block specific networks if configured
	if len(cfg.BlockNets) > 0 {
		if err := checkBlockNets(checkIPs, cfg.BlockNets); err != nil {
			return nil, err
		}
	}

	if err := checkBlockedCategories(checkIPs, cfg); err != nil {
		return nil, err
	}

	return resolved, nil
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
		// Block localhost if configured
		if cfg.BlockLocalhost && isLocalhost(ip) {
			return fmt.Errorf("localhost addresses are blocked")
		}

		// Block RFC1918 private networks if configured
		if cfg.BlockLocalNets && isPrivateNetwork(ip) {
			return fmt.Errorf("private network addresses (RFC1918) are blocked")
		}

		// Block bogon networks (reserved IPs) if configured
		if cfg.BlockBogonNets && isBogonNetwork(ip) {
			return fmt.Errorf("bogon/reserved network addresses are blocked")
		}

		// Block cloud metadata endpoints if configured
		if cfg.BlockCloudMetadata && isCloudMetadata(ip) {
			return fmt.Errorf("cloud metadata endpoints are blocked")
		}
	}

	return nil
}

// isGenericNotifier checks if the URL is a generic notifier that needs validation.
// The scheme is matched case-insensitively, matching how shoutrrr routes it.
func isGenericNotifier(notifierURL string) bool {
	scheme, _, ok := splitScheme(notifierURL)
	if !ok {
		return false
	}

	service, transport, hasTransport := strings.Cut(scheme, "+")
	if service != genericScheme {
		return false
	}

	return !hasTransport || transport == "http" || transport == "https"
}

// extractGenericURL extracts the actual HTTP(S) URL from a generic notifier URL
func extractGenericURL(notifierURL string) (string, error) {
	scheme, rest, ok := splitScheme(notifierURL)
	if !ok {
		return "", fmt.Errorf("not a generic notifier URL")
	}

	service, transport, hasTransport := strings.Cut(scheme, "+")
	if service != genericScheme {
		return "", fmt.Errorf("not a generic notifier URL")
	}

	// generic+http:// and generic+https:// carry the transport in the scheme.
	if hasTransport {
		if transport != "http" && transport != "https" {
			return "", fmt.Errorf("unsupported generic notifier transport %q", transport)
		}

		return transport + "://" + rest, nil
	}

	if rest == "" {
		return "", fmt.Errorf("generic notifier URL is empty")
	}

	// Support the nested generic://http://host/path form.
	if lower := strings.ToLower(rest); strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return rest, nil
	}

	// Support shorthand generic://host/path by defaulting to HTTPS.
	return "https://" + rest, nil
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
