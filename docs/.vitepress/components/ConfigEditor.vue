<template>
  <div class="config-generator">
    <div class="config-layout">
      <div class="config-form">
        <div class="tabs">
          <div class="tab-list">
            <button 
              v-for="tab in tabs" 
              :key="tab.value" 
              class="tab-button" 
              :class="{ active: activeTab === tab.value }"
              @click="activeTab = tab.value"
            >
              {{ tab.label }}
            </button>
          </div>

          <BasicConfig 
            v-show="activeTab === 'basic'" 
            :config="config" 
          />
          
          <DatabaseConfig 
            v-show="activeTab === 'database'" 
            :config="config" 
            :show-password="showPassword"
            @toggle-password="showPassword = !showPassword"
            @regenerate-password="regeneratePassword"
          />
          
          <HttpsConfig 
            v-show="activeTab === 'https'" 
            :config="config" 
          />
          
          <StorageConfig 
            v-show="activeTab === 'storage'" 
            :config="config" 
          />
        </div>
      </div>

      <ConfigPreview 
        :config="generateDockerCompose(config)"
        @copy="copyToClipboard"
        @download="downloadConfig"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import BasicConfig from './BasicConfig.vue'
import DatabaseConfig from './DatabaseConfig.vue'
import HttpsConfig from './HttpsConfig.vue'
import StorageConfig from './StorageConfig.vue'
import ConfigPreview from './ConfigPreview.vue'
import { generateDockerCompose } from './dockerComposeGenerator'

const showPassword = ref(false)
const activeTab = ref('basic')

const tabs = [
  { label: 'Basic', value: 'basic' },
  { label: 'Database', value: 'database' },
  { label: 'HTTPS', value: 'https' },
  { label: 'Storage', value: 'storage' }
]

function generateRandomPassword(length = 16) {
  const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+"
  let password = ""
  for (let i = 0; i < length; i++) {
    const randomIndex = Math.floor(Math.random() * charset.length)
    password += charset[randomIndex]
  }
  return password
}

const config = reactive({
  image: "ghcr.io/sysadminsmedia/homebox:latest",
  rootless: false,
  port: "3100",
  logLevel: "info",
  logFormat: "text",
  maxFileUpload: "10",
  allowAnalytics: false,
  
  // HTTPS options
  httpsOption: "none", // none, traefik, nginx, caddy, cloudflared
  
  // Traefik config
  traefikConfig: {
    domain: "homebox.example.com",
    email: "",
  },
  
  // Nginx config
  nginxConfig: {
    domain: "homebox.example.com",
    port: "443",
    sslCertPath: "/etc/nginx/ssl/cert.pem",
    sslKeyPath: "/etc/nginx/ssl/key.pem",
  },
  
  // Caddy config
  caddyConfig: {
    domain: "homebox.example.com",
    email: "",
  },
  
  // Cloudflared config
  cloudflaredConfig: {
    tunnel: "homebox-tunnel",
    domain: "homebox.example.com",
    token: "",
  },
  
  databaseType: "sqlite",
  postgresConfig: {
    host: "postgres",
    port: "5432",
    username: "homebox",
    password: generateRandomPassword(),
    database: "homebox",
  },
  allowRegistration: true,
  autoIncrementAssetId: true,
  checkGithubRelease: true,
  storageConfig: {
    homeboxStorage: {
      type: "volume", // "volume" or "directory"
      directory: "./homebox-data",
      volumeName: "homebox-data",
    },
    postgresStorage: {
      type: "volume",
      directory: "./postgres-data",
      volumeName: "postgres-data",
    },
    traefikStorage: {
      type: "volume",
      directory: "./traefik-data",
      volumeName: "traefik-data",
    },
    nginxStorage: {
      type: "volume",
      directory: "./nginx-data",
      volumeName: "nginx-data",
    },
    caddyStorage: {
      type: "volume",
      directory: "./caddy-data",
      volumeName: "caddy-data",
    },
    cloudflaredStorage: {
      type: "volume",
      directory: "./cloudflared-data",
      volumeName: "cloudflared-data",
    },
  },
})

function regeneratePassword() {
  config.postgresConfig.password = generateRandomPassword()
  alert('A new random password has been generated for the database.')
}

function copyToClipboard() {
  navigator.clipboard.writeText(generateDockerCompose(config))
  alert('Docker Compose configuration has been copied to your clipboard.')
}

function downloadConfig() {
  const element = document.createElement("a")
  const file = new Blob([generateDockerCompose(config)], { type: "text/plain" })
  element.href = URL.createObjectURL(file)
  element.download = "docker-compose.yml"
  document.body.appendChild(element)
  element.click()
  document.body.removeChild(element)
}
</script>

<style>
.config-generator {
  font-family: var(--vp-font-family-base);
  color: var(--vp-c-text-1);
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

.title {
  font-size: 2rem;
  font-weight: 600;
  margin-bottom: 2rem;
  color: var(--vp-c-brand);
}

.config-layout {
  display: grid;
  grid-template-columns: 1fr;
  gap: 2rem;
}

.tabs {
  width: 100%;
}

.tab-list {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.tab-button {
  padding: 0.5rem;
  background-color: var(--vp-c-bg-mute);
  border: 1px solid var(--vp-c-divider);
  border-radius: 4px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.tab-button.active {
  background-color: var(--vp-c-brand);
  color: white;
  border-color: var(--vp-c-brand);
}

.tab-button:hover:not(.active) {
  background-color: var(--vp-c-bg-alt);
}
</style>