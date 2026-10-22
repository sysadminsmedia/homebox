package printer

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
)

// privateIPBlocks contains the CIDR ranges for private/local networks
var (
	privateIPBlocks     []*net.IPNet
	privateIPBlocksOnce sync.Once
)

func getPrivateIPBlocks() []*net.IPNet {
	privateIPBlocksOnce.Do(func() {
		// RFC 1918 private address ranges
		// RFC 5735/5737 special-use addresses
		// RFC 4193 unique local addresses (IPv6)
		privateCIDRs := []string{
			"127.0.0.0/8",    // IPv4 loopback
			"10.0.0.0/8",     // RFC 1918 Class A
			"172.16.0.0/12",  // RFC 1918 Class B
			"192.168.0.0/16", // RFC 1918 Class C
			"169.254.0.0/16", // Link-local (excluding cloud metadata)
			"::1/128",        // IPv6 loopback
			"fc00::/7",       // IPv6 unique local
			"fe80::/10",      // IPv6 link-local
		}

		for _, cidr := range privateCIDRs {
			_, block, _ := net.ParseCIDR(cidr)
			privateIPBlocks = append(privateIPBlocks, block)
		}
	})
	return privateIPBlocks
}

// isPrivateIP checks if an IP address is in a private/local range
func isPrivateIP(ip net.IP) bool {
	// Check against cloud metadata IP (should be blocked even in link-local range)
	if ip.String() == "169.254.169.254" {
		return false
	}

	for _, block := range getPrivateIPBlocks() {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

// ValidatePrinterAddress validates that a printer address is safe to connect to.
// If allowPublic is false, only private/local IP addresses are allowed.
// Returns an error if the address is invalid or not allowed.
func ValidatePrinterAddress(address string, allowPublic bool) error {
	if address == "" {
		return fmt.Errorf("printer address cannot be empty")
	}

	// If public addresses are allowed, just do basic validation
	if allowPublic {
		return validateAddressFormat(address)
	}

	// Parse the address to extract the host
	host, err := extractHost(address)
	if err != nil {
		return err
	}

	// Resolve the hostname to IP addresses
	ips, err := net.LookupIP(host)
	if err != nil {
		// If it's already an IP address, try parsing it directly
		ip := net.ParseIP(host)
		if ip == nil {
			return fmt.Errorf("cannot resolve printer address: %w", err)
		}
		ips = []net.IP{ip}
	}

	// Check that all resolved IPs are private
	for _, ip := range ips {
		if !isPrivateIP(ip) {
			return fmt.Errorf("printer address must be on a private network (got %s)", ip.String())
		}
	}

	return nil
}

// extractHost extracts the hostname/IP from a printer address.
// Handles various formats: hostname, hostname:port, scheme://hostname, etc.
func extractHost(address string) (string, error) {
	// If it looks like a URL, parse it
	if strings.Contains(address, "://") {
		u, err := url.Parse(address)
		if err != nil {
			return "", fmt.Errorf("invalid printer address URL: %w", err)
		}
		host := u.Hostname()
		if host == "" {
			return "", fmt.Errorf("printer address has no host")
		}
		return host, nil
	}

	// Otherwise, it might be host:port or just host
	host := address
	if strings.Contains(address, ":") {
		var err error
		host, _, err = net.SplitHostPort(address)
		if err != nil {
			// Maybe it's just an IPv6 address without port
			if ip := net.ParseIP(address); ip != nil {
				return address, nil
			}
			return "", fmt.Errorf("invalid printer address format: %w", err)
		}
	}

	if host == "" {
		return "", fmt.Errorf("printer address has no host")
	}

	return host, nil
}

// validateAddressFormat does basic format validation without IP restrictions
func validateAddressFormat(address string) error {
	host, err := extractHost(address)
	if err != nil {
		return err
	}

	// Block cloud metadata endpoints even when public is allowed
	if host == "169.254.169.254" ||
		host == "metadata.google.internal" ||
		host == "metadata.goog" {
		return fmt.Errorf("cloud metadata endpoints are not allowed as printer addresses")
	}

	return nil
}
