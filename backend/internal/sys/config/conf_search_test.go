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

func Test_MeilisearchConf_RedactsAPIKey(t *testing.T) {
	c := MeilisearchConf{APIKey: "super-secret-meili-key"}

	out, err := json.Marshal(c)
	require.NoError(t, err)

	assert.NotContains(t, string(out), "super-secret-meili-key")
	assert.Contains(t, string(out), sentinel)
}
