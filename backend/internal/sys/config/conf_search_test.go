package config

import (
	"encoding/json"
	"testing"

	"github.com/ardanlabs/conf/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SearchConf_Defaults(t *testing.T) {
	var cfg Config
	_, err := conf.Parse("HBOXTEST", &cfg)
	require.NoError(t, err)

	assert.Equal(t, "database", cfg.Search.Driver)
	assert.Equal(t, "http://localhost:7700", cfg.Search.Meilisearch.Host)
	assert.Equal(t, "homebox_entities", cfg.Search.Meilisearch.Index)
	assert.Equal(t, int64(1000), cfg.Search.Meilisearch.MaxHits)
}

func Test_MeilisearchConf_Validate(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		wantErr bool
	}{
		{"https remote", "https://search.example.com", false},
		{"https remote with port", "https://search.example.com:7700", false},
		{"http localhost", "http://localhost:7700", false},
		{"http 127.0.0.1", "http://127.0.0.1:7700", false},
		{"http 127/8 loopback", "http://127.0.0.2:7700", false},
		{"http ipv6 loopback", "http://[::1]:7700", false},
		{"http hostname deferred to dial", "http://meilisearch:7700", false},
		{"http fqdn deferred to dial", "http://search.example.com:7700", false},
		{"http missing hostname", "http://:7700", true},
		{"http rfc1918 10/8", "http://10.0.0.5:7700", false},
		{"http rfc1918 172.16/12", "http://172.16.0.1:7700", false},
		{"http rfc1918 172.31 upper bound", "http://172.31.255.254:7700", false},
		{"http rfc1918 192.168/16", "http://192.168.1.20:7700", false},
		{"http just outside 172.16/12 rejected", "http://172.32.0.1:7700", true},
		{"http public ip rejected", "http://8.8.8.8:7700", true},
		{"http ipv6 ula fd00::/8", "http://[fd00::1]:7700", false},
		{"http ipv6 ula fc00::/8", "http://[fc00::1]:7700", false},
		{"http ipv6 public rejected", "http://[2001:4860:4860::8888]:7700", true},
		{"empty host", "", true},
		{"unsupported scheme", "ftp://localhost", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MeilisearchConf{Host: tt.host}.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_MeilisearchConf_RedactsAPIKey(t *testing.T) {
	c := MeilisearchConf{APIKey: "super-secret-meili-key"}

	out, err := json.Marshal(c)
	require.NoError(t, err)

	assert.NotContains(t, string(out), "super-secret-meili-key")
	assert.Contains(t, string(out), sentinel)
}
