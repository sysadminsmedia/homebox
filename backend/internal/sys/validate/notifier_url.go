package validate

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

// ValidateNotifierURL validates a notifier URL against the configured block/allow lists.
// This only applies to generic:// notifier URLs which can make arbitrary HTTP requests.
func ValidateNotifierURL(notifierURL string, cfg *config.NotifierConf) error {
	// Only validate generic notifiers
	if !isGenericNotifier(notifierURL) {
		return nil
	}

	// Defensively guard against nil cfg
	if cfg == nil {
		return fmt.Errorf("notifier configuration is nil, cannot validate URL")
	}

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

	host := parsedURL.Hostname()
	if host == "" {
		return fmt.Errorf("no hostname found in URL")
	}

	// Resolve the hostname to an IP address
	// NOTE: DNS responses can change after validation; consider re-validating at request time.
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("failed to resolve hostname: %w", err)
	}

	if len(ips) == 0 {
		return fmt.Errorf("hostname did not resolve to any IP addresses")
	}

	// If AllowNets is configured, only allow IPs in those networks
	if len(cfg.AllowNets) > 0 {
		for _, ip := range ips {
			allowed := false
			for _, allowNet := range cfg.AllowNets {
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
		// If explicitly allowed, skip other checks
		return nil
	}

	// Check BlockNets - block specific networks if configured
	if len(cfg.BlockNets) > 0 {
		for _, ip := range ips {
			for _, blockNet := range cfg.BlockNets {
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
	}

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

// isGenericNotifier checks if the URL is a generic notifier that needs validation
func isGenericNotifier(notifierURL string) bool {
	return strings.HasPrefix(notifierURL, "generic://") ||
		strings.HasPrefix(notifierURL, "generic+https://") ||
		strings.HasPrefix(notifierURL, "generic+http://")
}

// extractGenericURL extracts the actual HTTP(S) URL from a generic notifier URL
func extractGenericURL(notifierURL string) (string, error) {
	if strings.HasPrefix(notifierURL, "generic://") {
		return strings.TrimPrefix(notifierURL, "generic://"), nil
	}
	if strings.HasPrefix(notifierURL, "generic+https://") {
		return strings.TrimPrefix(notifierURL, "generic+"), nil
	}
	if strings.HasPrefix(notifierURL, "generic+http://") {
		return strings.TrimPrefix(notifierURL, "generic+"), nil
	}
	return "", fmt.Errorf("not a generic notifier URL")
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
