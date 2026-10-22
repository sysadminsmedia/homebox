package printer

import (
	"net"
	"testing"
)

func TestValidatePrinterAddress(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		allowPublic bool
		wantErr     bool
	}{
		// Private addresses - should always pass
		{
			name:        "localhost",
			address:     "127.0.0.1:631",
			allowPublic: false,
			wantErr:     false,
		},
		{
			name:        "localhost hostname",
			address:     "localhost:631",
			allowPublic: false,
			wantErr:     false,
		},
		{
			name:        "private 192.168.x.x",
			address:     "192.168.1.100:9100",
			allowPublic: false,
			wantErr:     false,
		},
		{
			name:        "private 10.x.x.x",
			address:     "10.0.0.50:631",
			allowPublic: false,
			wantErr:     false,
		},
		{
			name:        "private 172.16.x.x",
			address:     "172.16.0.1:631",
			allowPublic: false,
			wantErr:     false,
		},
		{
			name:        "IPP URL private",
			address:     "ipp://192.168.1.100:631/ipp/print",
			allowPublic: false,
			wantErr:     false,
		},

		// Cloud metadata - should always fail
		{
			name:        "AWS metadata IP blocked",
			address:     "169.254.169.254:80",
			allowPublic: false,
			wantErr:     true,
		},
		{
			name:        "AWS metadata even with allowPublic",
			address:     "169.254.169.254:80",
			allowPublic: true,
			wantErr:     true,
		},
		{
			name:        "GCP metadata hostname blocked",
			address:     "metadata.google.internal:80",
			allowPublic: true,
			wantErr:     true,
		},

		// Public addresses - blocked by default
		{
			name:        "public IP blocked by default",
			address:     "8.8.8.8:631",
			allowPublic: false,
			wantErr:     true,
		},
		{
			name:        "public IP allowed when enabled",
			address:     "8.8.8.8:631",
			allowPublic: true,
			wantErr:     false,
		},

		// Invalid addresses
		{
			name:        "empty address",
			address:     "",
			allowPublic: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePrinterAddress(tt.address, tt.allowPublic)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePrinterAddress(%q, %v) error = %v, wantErr %v",
					tt.address, tt.allowPublic, err, tt.wantErr)
			}
		})
	}
}

func TestExtractHost(t *testing.T) {
	tests := []struct {
		address  string
		wantHost string
		wantErr  bool
	}{
		{"192.168.1.1:631", "192.168.1.1", false},
		{"192.168.1.1", "192.168.1.1", false},
		{"ipp://192.168.1.1:631/ipp/print", "192.168.1.1", false},
		{"http://printer.local:631", "printer.local", false},
		{"printer.local:9100", "printer.local", false},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.address, func(t *testing.T) {
			got, err := extractHost(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractHost(%q) error = %v, wantErr %v", tt.address, err, tt.wantErr)
				return
			}
			if got != tt.wantHost {
				t.Errorf("extractHost(%q) = %q, want %q", tt.address, got, tt.wantHost)
			}
		})
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"127.0.0.1", true},
		{"10.0.0.1", true},
		{"10.255.255.255", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"192.168.0.1", true},
		{"192.168.255.255", true},
		{"169.254.1.1", true},
		{"169.254.169.254", false}, // Cloud metadata - explicitly blocked
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"::1", true},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := parseIP(tt.ip)
			if ip == nil {
				t.Fatalf("failed to parse IP: %s", tt.ip)
			}
			if got := isPrivateIP(ip); got != tt.want {
				t.Errorf("isPrivateIP(%s) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func parseIP(s string) net.IP {
	return net.ParseIP(s)
}
