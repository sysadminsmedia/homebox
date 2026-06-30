package validate

import (
	"context"
	"net"
	"testing"

	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

// Repeated fixture values used across many test cases.
const (
	cidrPrivate24       = "192.168.1.0/24"
	urlGenericIPv6Local = "generic://http://[fd00::1]/webhook"
	urlGenericIPv4Local = "generic://http://192.168.1.100/webhook"
)

// dns64DefaultNets mirrors the conf tag default for Dns64Nets.
var dns64DefaultNets = []string{"64:ff9b::/96", "64:ff9b:1::/48"}

func TestValidateNotifierURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		config      config.NotifierConf
		expectError bool
	}{
		{
			name: "non-generic notifier passes validation",
			url:  "discord://token@id",
			config: config.NotifierConf{
				BlockLocalhost: true,
			},
			expectError: false,
		},
		{
			name: "generic notifier with public IP passes",
			url:  "generic://https://8.8.8.8/webhook",
			config: config.NotifierConf{
				BlockLocalhost:     true,
				BlockLocalNets:     true,
				BlockBogonNets:     true,
				BlockCloudMetadata: true,
			},
			expectError: false,
		},
		{
			name: "generic notifier shorthand host/path passes",
			url:  "generic://8.8.8.8/webhook",
			config: config.NotifierConf{
				BlockLocalhost:     true,
				BlockLocalNets:     true,
				BlockBogonNets:     true,
				BlockCloudMetadata: true,
			},
			expectError: false,
		},
		{
			name: "generic notifier with localhost blocked",
			url:  "generic://http://localhost:8080/webhook",
			config: config.NotifierConf{
				BlockLocalhost: true,
			},
			expectError: true,
		},
		{
			name: "generic notifier with 127.0.0.1 blocked",
			url:  "generic://http://127.0.0.1:8080/webhook",
			config: config.NotifierConf{
				BlockLocalhost: true,
			},
			expectError: true,
		},
		{
			name: "generic notifier with private network blocked",
			url:  "generic://http://192.168.1.1/webhook",
			config: config.NotifierConf{
				BlockLocalNets: true,
			},
			expectError: true,
		},
		{
			name: "generic notifier with 10.0.0.0/8 blocked",
			url:  "generic://http://10.0.0.1/webhook",
			config: config.NotifierConf{
				BlockLocalNets: true,
			},
			expectError: true,
		},
		{
			name: "generic notifier with 172.16.0.0/12 blocked",
			url:  "generic://http://172.16.0.1/webhook",
			config: config.NotifierConf{
				BlockLocalNets: true,
			},
			expectError: true,
		},
		{
			name: "generic notifier with cloud metadata blocked",
			url:  "generic://http://169.254.169.254/latest/meta-data",
			config: config.NotifierConf{
				BlockCloudMetadata: true,
			},
			expectError: true,
		},
		{
			name: "generic notifier with link-local blocked by bogon",
			url:  "generic://http://169.254.1.1/webhook",
			config: config.NotifierConf{
				BlockBogonNets: true,
			},
			expectError: true,
		},
		{
			name: "generic+https notifier with localhost blocked",
			url:  "generic+https://127.0.0.1/webhook",
			config: config.NotifierConf{
				BlockLocalhost: true,
			},
			expectError: true,
		},
		{
			name: "generic+http notifier with private net blocked",
			url:  "generic+http://192.168.1.100/webhook",
			config: config.NotifierConf{
				BlockLocalNets: true,
			},
			expectError: true,
		},
		{
			name: "allow list permits private IP",
			url:  "generic://http://192.168.1.1/webhook",
			config: config.NotifierConf{
				AllowNets:      []string{cidrPrivate24},
				BlockLocalNets: true,
			},
			expectError: false,
		},
		{
			name: "allow list blocks non-matching IP",
			url:  "generic://http://10.0.0.1/webhook",
			config: config.NotifierConf{
				AllowNets: []string{cidrPrivate24},
			},
			expectError: true,
		},
		{
			name: "localhost_allowed_when_not_blocked",
			url:  "generic://http://localhost:8080/webhook",
			config: config.NotifierConf{
				BlockLocalhost: false,
			},
			expectError: false,
		},
		{
			name: "block_nets blocks specific network",
			url:  "generic://http://192.168.1.1/webhook",
			config: config.NotifierConf{
				BlockNets: []string{cidrPrivate24},
			},
			expectError: true,
		},
		{
			name: "block_nets allows non-matching network",
			url:  "generic://http://10.0.0.1/webhook",
			config: config.NotifierConf{
				BlockNets: []string{cidrPrivate24},
			},
			expectError: false,
		},
		{
			name: "block_nets with multiple networks",
			url:  "generic://http://172.16.0.1/webhook",
			config: config.NotifierConf{
				BlockNets: []string{"192.168.0.0/16", "172.16.0.0/12", "10.0.0.0/8"},
			},
			expectError: true,
		},
		{
			name: "block_nets does not affect non-generic notifiers",
			url:  "discord://token@id",
			config: config.NotifierConf{
				BlockNets: []string{"0.0.0.0/0"},
			},
			expectError: false,
		},
		{
			name: "allow_nets takes precedence over block_nets",
			url:  "generic://http://192.168.1.1/webhook",
			config: config.NotifierConf{
				AllowNets: []string{cidrPrivate24},
				BlockNets: []string{cidrPrivate24},
			},
			expectError: false,
		},
		// IPv6 test cases
		{
			name: "ipv6_loopback_blocked_when_localhost_blocked",
			url:  "generic://http://[::1]:8080/webhook",
			config: config.NotifierConf{
				BlockLocalhost: true,
			},
			expectError: true,
		},
		{
			name: "ipv6_loopback_allowed_when_localhost_not_blocked",
			url:  "generic://http://[::1]:8080/webhook",
			config: config.NotifierConf{
				BlockLocalhost: false,
			},
			expectError: false,
		},
		{
			name: "ipv6_ula_blocked_by_bogon_nets",
			url:  urlGenericIPv6Local,
			config: config.NotifierConf{
				BlockBogonNets: true,
			},
			expectError: true,
		},
		{
			name: "ipv6_ula_allowed_when_bogon_nets_not_blocked",
			url:  urlGenericIPv6Local,
			config: config.NotifierConf{
				BlockBogonNets: false,
			},
			expectError: false,
		},
		{
			name: "ipv6_ula_allowed_via_allow_nets",
			url:  urlGenericIPv6Local,
			config: config.NotifierConf{
				AllowNets:      []string{"fd00::/8"},
				BlockBogonNets: true,
			},
			expectError: false,
		},
		{
			name: "ipv6_link_local_blocked_by_bogon_nets",
			url:  "generic://http://[fe80::1]/webhook",
			config: config.NotifierConf{
				BlockBogonNets: true,
			},
			expectError: true,
		},
		{
			name: "ipv6_link_local_allowed_when_bogon_nets_not_blocked",
			url:  "generic://http://[fe80::1]/webhook",
			config: config.NotifierConf{
				BlockBogonNets: false,
			},
			expectError: false,
		},
		{
			name: "ipv6_aws_metadata_blocked_when_cloud_metadata_blocked",
			url:  "generic://http://[fd00:ec2::254]/webhook",
			config: config.NotifierConf{
				BlockCloudMetadata: true,
			},
			expectError: true,
		},
		{
			name: "ipv6_aws_metadata_allowed_when_cloud_metadata_not_blocked",
			url:  "generic://http://[fd00:ec2::254]/webhook",
			config: config.NotifierConf{
				BlockCloudMetadata: false,
			},
			expectError: false,
		},
		{
			name: "ipv6_block_nets_blocks_specific_ipv6_network",
			url:  "generic://http://[2001:db8::1]/webhook",
			config: config.NotifierConf{
				BlockNets: []string{"2001:db8::/32"},
			},
			expectError: true,
		},
		{
			name: "ipv6_block_nets_allows_non_matching_ipv6",
			url:  "generic://http://[2001:db8::1]/webhook",
			config: config.NotifierConf{
				BlockNets: []string{"2001:db9::/32"},
			},
			expectError: false,
		},
		{
			name: "ipv6_ula_blocked_by_local_nets",
			url:  urlGenericIPv6Local,
			config: config.NotifierConf{
				BlockLocalNets: true,
			},
			expectError: true,
		},
		{
			name: "ipv6_ula_allowed_when_local_nets_not_blocked",
			url:  urlGenericIPv6Local,
			config: config.NotifierConf{
				BlockLocalNets: false,
			},
			expectError: false,
		},
		// DNS64 test cases
		{
			name: "dns64_embedded_cloud_metadata_blocked",
			url:  "generic://http://[64:ff9b::a9fe:a9fe]/webhook",
			config: config.NotifierConf{
				BlockCloudMetadata: true,
				Dns64Nets:          dns64DefaultNets,
			},
			expectError: true,
		},
		{
			name: "dns64_embedded_cloud_metadata_blocked_by_bogon",
			url:  "generic://http://[64:ff9b::a9fe:a9fe]/webhook",
			config: config.NotifierConf{
				BlockBogonNets: true,
				Dns64Nets:      dns64DefaultNets,
			},
			expectError: true,
		},
		{
			name: "dns64_embedded_private_ip_blocked",
			url:  "generic://http://[64:ff9b::c0a8:101]/webhook",
			config: config.NotifierConf{
				BlockLocalNets: true,
				Dns64Nets:      dns64DefaultNets,
			},
			expectError: true,
		},
		{
			name: "dns64_embedded_localhost_blocked",
			url:  "generic://http://[64:ff9b::7f00:1]/webhook",
			config: config.NotifierConf{
				BlockLocalhost: true,
				Dns64Nets:      dns64DefaultNets,
			},
			expectError: true,
		},
		{
			name: "dns64_embedded_public_ip_passes",
			url:  "generic://http://[64:ff9b::808:808]/webhook",
			config: config.NotifierConf{
				BlockLocalhost:     true,
				BlockLocalNets:     true,
				BlockBogonNets:     true,
				BlockCloudMetadata: true,
				Dns64Nets:          dns64DefaultNets,
			},
			expectError: false,
		},
		{
			name: "dns64_local_use_prefix_embedded_metadata_blocked",
			url:  "generic://http://[64:ff9b:1::a9fe:a9fe]/webhook",
			config: config.NotifierConf{
				BlockCloudMetadata: true,
				Dns64Nets:          dns64DefaultNets,
			},
			expectError: true,
		},
		{
			name: "dns64_local_use_deployment_specific_prefix_blocked",
			url:  "generic://http://[64:ff9b:1:abcd::a9fe:a9fe]/webhook",
			config: config.NotifierConf{
				BlockCloudMetadata: true,
				Dns64Nets:          dns64DefaultNets,
			},
			expectError: true,
		},
		{
			name: "dns64_custom_prefix_embedded_metadata_blocked",
			url:  "generic://http://[2001:db8:64::a9fe:a9fe]/webhook",
			config: config.NotifierConf{
				BlockCloudMetadata: true,
				Dns64Nets:          []string{"2001:db8:64::/96"},
			},
			expectError: true,
		},
		{
			name: "dns64_embedded_ip_checked_against_block_nets",
			url:  "generic://http://[64:ff9b::c0a8:101]/webhook",
			config: config.NotifierConf{
				BlockNets: []string{cidrPrivate24},
				Dns64Nets: dns64DefaultNets,
			},
			expectError: true,
		},
		{
			name: "dns64_allow_nets_requires_embedded_ip_allowed",
			url:  "generic://http://[64:ff9b::c0a8:101]/webhook",
			config: config.NotifierConf{
				AllowNets: []string{"64:ff9b::/96"},
				Dns64Nets: dns64DefaultNets,
			},
			expectError: true,
		},
		{
			name: "dns64_allow_nets_passes_when_prefix_and_embedded_ip_allowed",
			url:  "generic://http://[64:ff9b::c0a8:101]/webhook",
			config: config.NotifierConf{
				AllowNets: []string{"64:ff9b::/96", cidrPrivate24},
				Dns64Nets: dns64DefaultNets,
			},
			expectError: false,
		},
		{
			name: "dns64_range_without_extractable_ipv4_fails_closed",
			url:  "generic://http://[64:ff9b::a9fe:a9fe]/webhook",
			config: config.NotifierConf{
				Dns64Nets: []string{"64:ff9b::a9fe:a9fe/128"},
			},
			expectError: true,
		},
		{
			name:        "dns64_checks_disabled_when_no_nets_configured",
			url:         "generic://http://[64:ff9b::a9fe:a9fe]/webhook",
			config:      config.NotifierConf{BlockCloudMetadata: true},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotifierURL(tt.url, &tt.config)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

func TestValidateOutboundHTTPURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		config      config.NotifierConf
		expectError bool
	}{
		{
			name: "plain http public URL passes",
			url:  "http://8.8.8.8/webhook",
			config: config.NotifierConf{
				BlockLocalhost:     true,
				BlockLocalNets:     true,
				BlockBogonNets:     true,
				BlockCloudMetadata: true,
				Dns64Nets:          dns64DefaultNets,
			},
			expectError: false,
		},
		{
			name: "plain http private URL obeys shared policy",
			url:  "http://192.168.1.1/webhook",
			config: config.NotifierConf{
				BlockLocalNets: true,
			},
			expectError: true,
		},
		{
			name:        "non-http URL is rejected",
			url:         "ftp://example.com/file",
			config:      config.NotifierConf{},
			expectError: true,
		},
		{
			name: "plain http DNS64 metadata URL is blocked",
			url:  "http://[64:ff9b::a9fe:a9fe]/webhook",
			config: config.NotifierConf{
				BlockCloudMetadata: true,
				Dns64Nets:          dns64DefaultNets,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOutboundHTTPURL(tt.url, &tt.config)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

func TestOutboundHTTPTransportBlocksDialTarget(t *testing.T) {
	transport := NewOutboundHTTPTransport(&config.NotifierConf{
		BlockLocalhost: true,
	})

	conn, err := transport.DialContext(context.Background(), "tcp", "127.0.0.1:80")
	if err == nil {
		_ = conn.Close()
		t.Fatal("expected localhost dial target to be blocked")
	}
}

func TestFilterOutboundIPsByNetworkBeforePolicyValidation(t *testing.T) {
	ips := []net.IP{net.ParseIP("::1"), net.ParseIP("8.8.8.8")}
	filtered := filterOutboundIPsByNetwork("tcp4", ips)

	if len(filtered) != 1 || !filtered[0].Equal(net.ParseIP("8.8.8.8")) {
		t.Fatalf("expected only the IPv4 address after filtering, got %#v", filtered)
	}

	err := validateOutboundIPs(filtered, &config.NotifierConf{
		BlockLocalhost: true,
	})
	if err != nil {
		t.Fatalf("expected filtered IPv4 target to pass localhost policy: %v", err)
	}
}

func TestIsGenericNotifier(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"generic://", "generic://http://example.com", true},
		{"generic+https://", "generic+https://example.com", true},
		{"generic+http://", "generic+http://example.com", true},
		{"discord://", "discord://token@id", false},
		{"slack://", "slack://token@channel", false},
		{"smtp://", "smtp://user:pass@host:587", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGenericNotifier(tt.url)
			if result != tt.expected {
				t.Errorf("expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestExtractGenericURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expected    string
		expectError bool
	}{
		{
			name:        "generic://",
			url:         "generic://http://example.com/webhook",
			expected:    "http://example.com/webhook",
			expectError: false,
		},
		{
			name:        "generic:// shorthand defaults to https",
			url:         "generic://example.com/webhook",
			expected:    "https://example.com/webhook",
			expectError: false,
		},
		{
			name:        "generic+https://",
			url:         "generic+https://example.com/webhook",
			expected:    "https://example.com/webhook",
			expectError: false,
		},
		{
			name:        "generic+http://",
			url:         "generic+http://example.com/webhook",
			expected:    "http://example.com/webhook",
			expectError: false,
		},
		{
			name:        "not generic",
			url:         "discord://token@id",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractGenericURL(tt.url)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestEmbeddedIPv4(t *testing.T) {
	// Layout examples are taken from RFC 6052 section 2.4, all embedding 192.0.2.33.
	const rfc6052Embedded = "192.0.2.33"
	tests := []struct {
		name      string
		ip        string
		prefixLen int
		expected  string // empty means extraction must fail
	}{
		{"rfc6052_example_32", "2001:db8:c000:221::", 32, rfc6052Embedded},
		{"rfc6052_example_40", "2001:db8:1c0:2:21::", 40, rfc6052Embedded},
		{"rfc6052_example_48", "2001:db8:122:c000:2:2100::", 48, rfc6052Embedded},
		{"rfc6052_example_56", "2001:db8:122:3c0:0:221::", 56, rfc6052Embedded},
		{"rfc6052_example_64", "2001:db8:122:344:c0:2:2100::", 64, rfc6052Embedded},
		{"rfc6052_example_96", "2001:db8:122:344::c000:221", 96, rfc6052Embedded},
		{"well_known_prefix_96", "64:ff9b::a9fe:a9fe", 96, "169.254.169.254"},
		{"nonzero_u_octet_rejected", "2001:db8:122:344:ffc0:2:2100::", 64, ""},
		{"nonzero_suffix_rejected", "2001:db8:122:344:c0:2:2100:1", 64, ""},
		{"unsupported_prefix_length", "64:ff9b::a9fe:a9fe", 128, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := embeddedIPv4(net.ParseIP(tt.ip), tt.prefixLen)
			if tt.expected == "" {
				if result != nil {
					t.Errorf("expected no extraction but got %v", result)
				}
				return
			}
			if result == nil || result.String() != tt.expected {
				t.Errorf("expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestValidateNotifierURL_InvalidCIDR_AllowNets(t *testing.T) {
	// Test with invalid CIDR in AllowNets - should skip invalid and check valid ones
	tests := []struct {
		name        string
		url         string
		config      config.NotifierConf
		expectError bool
		description string
	}{
		{
			name: "invalid CIDR in AllowNets is skipped",
			url:  urlGenericIPv4Local,
			config: config.NotifierConf{
				AllowNets: []string{
					"invalid-cidr", // Invalid - should be logged and skipped
					cidrPrivate24,  // Valid - should match
				},
			},
			expectError: false,
			description: "IP should be allowed by valid CIDR despite invalid CIDR present",
		},
		{
			name: "all CIDRs invalid in AllowNets",
			url:  urlGenericIPv4Local,
			config: config.NotifierConf{
				AllowNets: []string{
					"invalid-cidr-1",
					"not-a-cidr",
					"999.999.999.999/32",
				},
			},
			expectError: true,
			description: "IP should be blocked when all AllowNets CIDRs are invalid",
		},
		{
			name: "mixed valid and invalid CIDRs",
			url:  "generic://http://10.0.0.100/webhook",
			config: config.NotifierConf{
				AllowNets: []string{
					"invalid",
					"10.0.0.0/24",
					"also-invalid",
				},
			},
			expectError: false,
			description: "Valid CIDR should work despite surrounding invalid CIDRs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotifierURL(tt.url, &tt.config)
			if tt.expectError && err == nil {
				t.Errorf("%s: expected error but got none", tt.description)
			}
			if !tt.expectError && err != nil {
				t.Errorf("%s: expected no error but got: %v", tt.description, err)
			}
		})
	}
}

func TestValidateNotifierURL_InvalidCIDR_BlockNets(t *testing.T) {
	// Test with invalid CIDR in BlockNets - should skip invalid and check valid ones
	tests := []struct {
		name        string
		url         string
		config      config.NotifierConf
		expectError bool
		description string
	}{
		{
			name: "invalid CIDR in BlockNets is skipped",
			url:  urlGenericIPv4Local,
			config: config.NotifierConf{
				BlockNets: []string{
					"invalid-cidr", // Invalid - should be logged and skipped
					"10.0.0.0/8",   // Valid but doesn't match
				},
			},
			expectError: false,
			description: "IP should pass when invalid CIDR is skipped and valid one doesn't match",
		},
		{
			name: "invalid CIDR doesn't prevent valid blocking",
			url:  urlGenericIPv4Local,
			config: config.NotifierConf{
				BlockNets: []string{
					"not-a-cidr",
					cidrPrivate24, // Valid - should block
					"also-invalid",
				},
			},
			expectError: true,
			description: "Valid blocking CIDR should work despite invalid CIDRs",
		},
		{
			name: "all CIDRs invalid in BlockNets",
			url:  urlGenericIPv4Local,
			config: config.NotifierConf{
				BlockNets: []string{
					"invalid-1",
					"invalid-2",
					"999.999.999.999/32",
				},
			},
			expectError: false,
			description: "IP should pass when all BlockNets CIDRs are invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotifierURL(tt.url, &tt.config)
			if tt.expectError && err == nil {
				t.Errorf("%s: expected error but got none", tt.description)
			}
			if !tt.expectError && err != nil {
				t.Errorf("%s: expected no error but got: %v", tt.description, err)
			}
		})
	}
}
