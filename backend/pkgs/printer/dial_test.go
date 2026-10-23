package printer

import (
	"context"
	"strings"
	"testing"
)

func TestSafeDialContextRejectsPublicAddresses(t *testing.T) {
	dial := safeDialContext(false)
	for _, addr := range []string{"93.184.216.34:9100", "169.254.169.254:80", "[2001:db8::1]:631"} {
		_, err := dial(context.Background(), "tcp", addr)
		if err == nil || !strings.Contains(err.Error(), "not on a private network") {
			t.Errorf("%s: expected private-network rejection, got %v", addr, err)
		}
	}
}

func TestSafeDialContextAllowsPublicWhenConfigured(t *testing.T) {
	// Port 1 on a documentation-range address: the dial itself fails, but it
	// must fail on the connection, not on the address check.
	ctx, cancel := context.WithTimeout(context.Background(), 200*1000*1000)
	defer cancel()
	_, err := safeDialContext(true)(ctx, "tcp", "192.0.2.1:1")
	if err != nil && strings.Contains(err.Error(), "not on a private network") {
		t.Fatalf("public address was rejected although allowed: %v", err)
	}
}
