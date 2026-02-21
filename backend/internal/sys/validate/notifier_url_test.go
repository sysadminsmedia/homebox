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
