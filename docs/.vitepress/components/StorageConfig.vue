<template>
  <div class="tab-content">
    <div class="card">
      <div class="card-header">
        <h2 class="card-title">Storage Configuration</h2>
        <p class="card-description">Configure storage options for your Homebox instance and related services.</p>
      </div>
      <div class="card-content">
        <!-- Homebox Storage -->
        <StorageTypeSelector
          storage-key="homeboxStorage"
          label="Homebox Data Storage"
          description="Store Homebox data in a Docker volume or host directory"
          :config="config"
        />

        <!-- PostgreSQL Storage -->
        <StorageTypeSelector
          v-if="config.databaseType === 'postgres'"
          storage-key="postgresStorage"
          label="PostgreSQL Data Storage"
          description="Store PostgreSQL data in a Docker volume or host directory"
          :config="config"
        />

        <!-- HTTPS Service Storage -->
        <StorageTypeSelector
          v-if="config.httpsOption === 'traefik'"
          storage-key="traefikStorage"
          label="Traefik Data Storage"
          description="Store Traefik certificates in a Docker volume or host directory"
          :config="config"
        />

        <StorageTypeSelector
          v-if="config.httpsOption === 'nginx'"
          storage-key="nginxStorage"
          label="Nginx Data Storage"
          description="Store Nginx configuration and certificates in a Docker volume or host directory"
          :config="config"
        />

        <StorageTypeSelector
          v-if="config.httpsOption === 'caddy'"
          storage-key="caddyStorage"
          label="Caddy Data Storage"
          description="Store Caddy configuration and certificates in a Docker volume or host directory"
          :config="config"
        />

        <StorageTypeSelector
          v-if="config.httpsOption === 'cloudflared'"
          storage-key="cloudflaredStorage"
          label="Cloudflared Data Storage"
          description="Store Cloudflared configuration in a Docker volume or host directory"
          :config="config"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import StorageTypeSelector from './StorageTypeSelector.vue'

defineProps({
  config: {
    type: Object,
    required: true
  }
})
</script>

<style scoped>
@import './common.css';
</style>