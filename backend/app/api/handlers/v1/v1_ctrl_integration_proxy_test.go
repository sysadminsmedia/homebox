package v1

import (
	"net/http"
	"testing"

	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

func TestIntegrationProxyHTTPClientValidatesRedirectTargets(t *testing.T) {
	ctrl := &V1Controller{
		config: &config.Config{
			Notifier: config.NotifierConf{
				BlockCloudMetadata: true,
				Dns64Nets:          []string{"64:ff9b::/96", "64:ff9b:1::/48"},
			},
		},
	}

	client := ctrl.integrationProxyHTTPClient()
	req, err := http.NewRequest(http.MethodGet, "http://169.254.169.254/latest/meta-data", nil)
	if err != nil {
		t.Fatalf("failed to build redirect request: %v", err)
	}

	if err := client.CheckRedirect(req, nil); err == nil {
		t.Fatal("expected redirect to cloud metadata endpoint to be blocked")
	}
}
