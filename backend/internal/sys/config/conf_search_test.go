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
		{"http ipv6 loopback", "http://[::1]:7700", false},
		{"http remote rejected", "http://search.example.com:7700", true},
		{"http remote ip rejected", "http://10.0.0.5:7700", true},
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
