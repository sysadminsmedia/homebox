package validate

import (
	"testing"

	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

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
			url:  "generic://https://example.com/webhook",
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
				AllowNets:      []string{"192.168.1.0/24"},
				BlockLocalNets: true,
			},
			expectError: false,
		},
		{
			name: "allow list blocks non-matching IP",
			url:  "generic://http://10.0.0.1/webhook",
			config: config.NotifierConf{
				AllowNets: []string{"192.168.1.0/24"},
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
				BlockNets: []string{"192.168.1.0/24"},
			},
			expectError: true,
		},
		{
			name: "block_nets allows non-matching network",
			url:  "generic://http://10.0.0.1/webhook",
			config: config.NotifierConf{
				BlockNets: []string{"192.168.1.0/24"},
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
				AllowNets: []string{"192.168.1.0/24"},
				BlockNets: []string{"192.168.1.0/24"},
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
			url:  "generic://http://[fd00::1]/webhook",
			config: config.NotifierConf{
				BlockBogonNets: true,
			},
			expectError: true,
		},
		{
			name: "ipv6_ula_allowed_when_bogon_nets_not_blocked",
			url:  "generic://http://[fd00::1]/webhook",
			config: config.NotifierConf{
				BlockBogonNets: false,
			},
			expectError: false,
		},
		{
			name: "ipv6_ula_allowed_via_allow_nets",
			url:  "generic://http://[fd00::1]/webhook",
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
			url:  "generic://http://[fd00::1]/webhook",
			config: config.NotifierConf{
				BlockLocalNets: true,
			},
			expectError: true,
		},
		{
			name: "ipv6_ula_allowed_when_local_nets_not_blocked",
			url:  "generic://http://[fd00::1]/webhook",
			config: config.NotifierConf{
				BlockLocalNets: false,
			},
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
			url:  "generic://http://192.168.1.100/webhook",
			config: config.NotifierConf{
				AllowNets: []string{
					"invalid-cidr",   // Invalid - should be logged and skipped
					"192.168.1.0/24", // Valid - should match
				},
			},
			expectError: false,
			description: "IP should be allowed by valid CIDR despite invalid CIDR present",
		},
		{
			name: "all CIDRs invalid in AllowNets",
			url:  "generic://http://192.168.1.100/webhook",
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
			url:  "generic://http://192.168.1.100/webhook",
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
			url:  "generic://http://192.168.1.100/webhook",
			config: config.NotifierConf{
				BlockNets: []string{
					"not-a-cidr",
					"192.168.1.0/24", // Valid - should block
					"also-invalid",
				},
			},
			expectError: true,
			description: "Valid blocking CIDR should work despite invalid CIDRs",
		},
		{
			name: "all CIDRs invalid in BlockNets",
			url:  "generic://http://192.168.1.100/webhook",
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
