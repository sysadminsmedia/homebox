package config

import (
	"testing"

	"github.com/ardanlabs/conf/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_BackupConf_Defaults(t *testing.T) {
	var cfg Config
	_, err := conf.Parse("HBOXTEST", &cfg)
	require.NoError(t, err)

	assert.True(t, cfg.Backup.Enabled)
	assert.Empty(t, cfg.Backup.LocalRoot, "local destinations are off until a root is set")
	assert.False(t, cfg.Backup.AllowCustomEndpoints,
		"custom endpoints make the server connect to addresses a collection owner enters, so an operator must opt in")
}

func Test_BackupConf_CustomEndpointsOptIn(t *testing.T) {
	t.Setenv("HBOXTEST_BACKUP_ALLOW_CUSTOM_ENDPOINTS", "true")
	var cfg Config
	_, err := conf.Parse("HBOXTEST", &cfg)
	require.NoError(t, err)
	assert.True(t, cfg.Backup.AllowCustomEndpoints)
}
