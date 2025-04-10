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

          <!-- Basic Tab -->
          <div v-show="activeTab === 'basic'" class="tab-content">
            <div class="card">
              <div class="card-header">
                <h2 class="card-title">Basic Configuration</h2>
                <p class="card-description">Configure the basic settings for your Homebox instance.</p>
              </div>
              <div class="card-content">
                <div class="form-row">
                  <label for="rootless">Use Rootless Image</label>
                  <div class="toggle-switch">
                    <input 
                      type="checkbox" 
                      id="rootless" 
                      v-model="config.rootless"
                    />
                    <label for="rootless"></label>
                  </div>
                </div>

                <div class="form-group">
                  <label for="port">External Port</label>
                  <input 
                    type="text" 
                    id="port" 
                    v-model="config.port"
                  />
                  <p class="help-text">Only used if HTTPS with Traefik is not enabled</p>
                </div>

                <div class="form-group">
                  <label for="maxFileUpload">Max File Upload (MB)</label>
                  <input 
                    type="text" 
                    id="maxFileUpload" 
                    v-model="config.maxFileUpload"
                  />
                </div>

                <div class="form-row">
                  <label for="allowAnalytics">Allow Analytics</label>
                  <div class="toggle-switch">
                    <input 
                      type="checkbox" 
                      id="allowAnalytics" 
                      v-model="config.allowAnalytics"
                    />
                    <label for="allowAnalytics"></label>
                  </div>
                </div>

                <div class="separator"></div>

                <div class="form-row">
                  <label for="allowRegistration">Allow Registration</label>
                  <div class="toggle-switch">
                    <input 
                      type="checkbox" 
                      id="allowRegistration" 
                      v-model="config.allowRegistration"
                    />
                    <label for="allowRegistration"></label>
                  </div>
                </div>

                <div class="form-row">
                  <label for="autoIncrementAssetId">Auto Increment Asset ID</label>
                  <div class="toggle-switch">
                    <input 
                      type="checkbox" 
                      id="autoIncrementAssetId" 
                      v-model="config.autoIncrementAssetId"
                    />
                    <label for="autoIncrementAssetId"></label>
                  </div>
                </div>

                <div class="form-row">
                  <label for="checkGithubRelease">Check GitHub Release</label>
                  <div class="toggle-switch">
                    <input 
                      type="checkbox" 
                      id="checkGithubRelease" 
                      v-model="config.checkGithubRelease"
                    />
                    <label for="checkGithubRelease"></label>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Database Tab -->
          <div v-show="activeTab === 'database'" class="tab-content">
            <div class="card">
              <div class="card-header">
                <h2 class="card-title">Database Configuration</h2>
                <p class="card-description">Configure the database for your Homebox instance.</p>
              </div>
              <div class="card-content">
                <div class="form-group">
                  <label for="databaseType">Database Type</label>
                  <select id="databaseType" v-model="config.databaseType">
                    <option value="sqlite">SQLite (Default)</option>
                    <option value="postgres">PostgreSQL</option>
                  </select>
                </div>

                <div v-if="config.databaseType === 'postgres'" class="nested-form">
                  <div class="form-group">
                    <label for="postgresHost">PostgreSQL Host</label>
                    <input 
                      type="text" 
                      id="postgresHost" 
                      v-model="config.postgresConfig.host"
                    />
                  </div>

                  <div class="form-group">
                    <label for="postgresPort">PostgreSQL Port</label>
                    <input 
                      type="text" 
                      id="postgresPort" 
                      v-model="config.postgresConfig.port"
                    />
                  </div>

                  <div class="form-group">
                    <label for="postgresUsername">PostgreSQL Username</label>
                    <input 
                      type="text" 
                      id="postgresUsername" 
                      v-model="config.postgresConfig.username"
                    />
                  </div>

                  <div class="form-group">
                    <label for="postgresPassword">PostgreSQL Password</label>
                    <div class="password-input">
                      <input 
                        :type="showPassword ? 'text' : 'password'" 
                        id="postgresPassword" 
                        v-model="config.postgresConfig.password"
                      />
                      <button 
                        class="icon-button" 
                        @click="showPassword = !showPassword"
                        type="button"
                      >
                        <span v-if="showPassword">Hide</span>
                        <span v-else>Show</span>
                      </button>
                      <button 
                        class="icon-button" 
                        @click="regeneratePassword"
                        type="button"
                        title="Generate new random password"
                      >
                        Regenerate
                      </button>
                    </div>
                  </div>

                  <div class="form-group">
                    <label for="postgresDatabase">PostgreSQL Database</label>
                    <input 
                      type="text" 
                      id="postgresDatabase" 
                      v-model="config.postgresConfig.database"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Proxy Tab -->
          <div v-show="activeTab === 'proxy'" class="tab-content">
            <div class="card">
              <div class="card-header">
                <h2 class="card-title">Proxy Configuration with Traefik</h2>
                <p class="card-description">Configure Traefik as a reverse proxy with automatic HTTPS for your Homebox instance.</p>
              </div>
              <div class="card-content">
                <div class="form-row">
                  <label for="useTraefik">Enable HTTPS with Traefik</label>
                  <div class="toggle-switch">
                    <input 
                      type="checkbox" 
                      id="useTraefik" 
                      v-model="config.useTraefik"
                    />
                    <label for="useTraefik"></label>
                  </div>
                </div>

                <div v-if="config.useTraefik" class="nested-form">
                  <div class="form-group">
                    <label for="traefikDomain">Domain Name</label>
                    <input 
                      type="text" 
                      id="traefikDomain" 
                      v-model="config.traefikDomain"
                      placeholder="homebox.example.com"
                    />
                    <p class="help-text">The domain name must be pointed to your server's IP address</p>
                  </div>

                  <div class="form-group">
                    <label for="traefikEmail">Email Address (for Let's Encrypt)</label>
                    <input 
                      type="email" 
                      id="traefikEmail" 
                      v-model="config.traefikEmail"
                      placeholder="your-email@example.com"
                    />
                    <p class="help-text">Required for Let's Encrypt certificate notifications</p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Storage Tab -->
          <div v-show="activeTab === 'storage'" class="tab-content">
            <div class="card">
              <div class="card-header">
                <h2 class="card-title">Storage Configuration</h2>
                <p class="card-description">Configure storage options for your Homebox instance and related services.</p>
              </div>
              <div class="card-content">
                <!-- Homebox Storage -->
                <div class="storage-selector">
                  <h3>Homebox Data Storage</h3>
                  <p class="help-text">Store Homebox data in a Docker volume or host directory</p>
                  
                  <div class="radio-group">
                    <div class="radio-option">
                      <input 
                        type="radio" 
                        id="homeboxStorage-volume" 
                        value="volume" 
                        v-model="config.storageConfig.homeboxStorage.type"
                      />
                      <label for="homeboxStorage-volume">Docker Volume</label>
                    </div>
                    <div class="radio-option">
                      <input 
                        type="radio" 
                        id="homeboxStorage-directory" 
                        value="directory" 
                        v-model="config.storageConfig.homeboxStorage.type"
                      />
                      <label for="homeboxStorage-directory">Host Directory</label>
                    </div>
                  </div>

                  <div v-if="config.storageConfig.homeboxStorage.type === 'volume'" class="form-group">
                    <label for="homeboxStorage-volume-name">Volume Name</label>
                    <input 
                      type="text" 
                      id="homeboxStorage-volume-name" 
                      v-model="config.storageConfig.homeboxStorage.volumeName"
                    />
                  </div>
                  <div v-else class="form-group">
                    <label for="homeboxStorage-directory-path">Directory Path</label>
                    <input 
                      type="text" 
                      id="homeboxStorage-directory-path" 
                      v-model="config.storageConfig.homeboxStorage.directory"
                    />
                    <p class="help-text">Absolute path recommended (e.g., /home/user/homebox-data)</p>
                  </div>
                </div>

                <!-- PostgreSQL Storage -->
                <div v-if="config.databaseType === 'postgres'" class="storage-selector">
                  <h3>PostgreSQL Data Storage</h3>
                  <p class="help-text">Store PostgreSQL data in a Docker volume or host directory</p>
                  
                  <div class="radio-group">
                    <div class="radio-option">
                      <input 
                        type="radio" 
                        id="postgresStorage-volume" 
                        value="volume" 
                        v-model="config.storageConfig.postgresStorage.type"
                      />
                      <label for="postgresStorage-volume">Docker Volume</label>
                    </div>
                    <div class="radio-option">
                      <input 
                        type="radio" 
                        id="postgresStorage-directory" 
                        value="directory" 
                        v-model="config.storageConfig.postgresStorage.type"
                      />
                      <label for="postgresStorage-directory">Host Directory</label>
                    </div>
                  </div>

                  <div v-if="config.storageConfig.postgresStorage.type === 'volume'" class="form-group">
                    <label for="postgresStorage-volume-name">Volume Name</label>
                    <input 
                      type="text" 
                      id="postgresStorage-volume-name" 
                      v-model="config.storageConfig.postgresStorage.volumeName"
                    />
                  </div>
                  <div v-else class="form-group">
                    <label for="postgresStorage-directory-path">Directory Path</label>
                    <input 
                      type="text" 
                      id="postgresStorage-directory-path" 
                      v-model="config.storageConfig.postgresStorage.directory"
                    />
                    <p class="help-text">Absolute path recommended (e.g., /home/user/postgres-data)</p>
                  </div>
                </div>

                <!-- Traefik Storage -->
                <div v-if="config.useTraefik" class="storage-selector">
                  <h3>Traefik Data Storage</h3>
                  <p class="help-text">Store Traefik certificates in a Docker volume or host directory</p>
                  
                  <div class="radio-group">
                    <div class="radio-option">
                      <input 
                        type="radio" 
                        id="traefikStorage-volume" 
                        value="volume" 
                        v-model="config.storageConfig.traefikStorage.type"
                      />
                      <label for="traefikStorage-volume">Docker Volume</label>
                    </div>
                    <div class="radio-option">
                      <input 
                        type="radio" 
                        id="traefikStorage-directory" 
                        value="directory" 
                        v-model="config.storageConfig.traefikStorage.type"
                      />
                      <label for="traefikStorage-directory">Host Directory</label>
                    </div>
                  </div>

                  <div v-if="config.storageConfig.traefikStorage.type === 'volume'" class="form-group">
                    <label for="traefikStorage-volume-name">Volume Name</label>
                    <input 
                      type="text" 
                      id="traefikStorage-volume-name" 
                      v-model="config.storageConfig.traefikStorage.volumeName"
                    />
                  </div>
                  <div v-else class="form-group">
                    <label for="traefikStorage-directory-path">Directory Path</label>
                    <input 
                      type="text" 
                      id="traefikStorage-directory-path" 
                      v-model="config.storageConfig.traefikStorage.directory"
                    />
                    <p class="help-text">Absolute path recommended (e.g., /home/user/traefik-data)</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="config-preview">
        <div class="card">
          <div class="card-header">
            <div class="card-title-with-actions">
              <h2 class="card-title">Docker Compose Configuration</h2>
              <div class="card-actions">
                <button class="icon-button" @click="copyToClipboard" title="Copy to clipboard">
                  Copy
                </button>
                <button class="icon-button" @click="downloadConfig" title="Download as file">
                  Download
                </button>
              </div>
            </div>
            <p class="card-description">This configuration will be saved as docker-compose.yml</p>
          </div>
          <div class="card-content">
            <textarea 
              class="code-preview" 
              readonly 
              :value="generateDockerCompose()"
            ></textarea>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'

const showPassword = ref(false)
const activeTab = ref('basic')

const tabs = [
  { label: 'Basic', value: 'basic' },
  { label: 'Database', value: 'database' },
  { label: 'Proxy', value: 'proxy' },
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
  useTraefik: false,
  traefikDomain: "homebox.example.com",
  traefikEmail: "",
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
  },
})

function regeneratePassword() {
  config.postgresConfig.password = generateRandomPassword()
  alert('A new random password has been generated for the database.')
}

function generateDockerCompose() {
  const services = {
    homebox: {
      image: config.rootless
        ? "ghcr.io/sysadminsmedia/homebox:latest-rootless"
        : "ghcr.io/sysadminsmedia/homebox:latest",
      container_name: "homebox",
      restart: "always",
      environment: [
        `HBOX_LOG_LEVEL=${config.logLevel}`,
        `HBOX_LOG_FORMAT=${config.logFormat}`,
        `HBOX_WEB_MAX_FILE_UPLOAD=${config.maxFileUpload}`,
        `HBOX_OPTIONS_ALLOW_ANALYTICS=${config.allowAnalytics}`,
        `HBOX_OPTIONS_ALLOW_REGISTRATION=${config.allowRegistration}`,
        `HBOX_OPTIONS_AUTO_INCREMENT_ASSET_ID=${config.autoIncrementAssetId}`,
        `HBOX_OPTIONS_CHECK_GITHUB_RELEASE=${config.checkGithubRelease}`,
      ],
      volumes: [],
    },
  }

  // Configure homebox volumes based on storage type
  if (config.storageConfig.homeboxStorage.type === "volume") {
    services.homebox.volumes.push(`${config.storageConfig.homeboxStorage.volumeName}:/data/`)
  } else {
    services.homebox.volumes.push(`${config.storageConfig.homeboxStorage.directory}:/data/`)
  }

  // Add ports or labels based on whether Traefik is used
  if (config.useTraefik) {
    services.homebox.labels = [
      "traefik.enable=true",
      `traefik.http.routers.homebox.rule=Host(\`${config.traefikDomain}\`)`,
      "traefik.http.routers.homebox.entrypoints=websecure",
      "traefik.http.routers.homebox.tls.certresolver=letsencrypt",
      "traefik.http.services.homebox.loadbalancer.server.port=7745",
    ]
    // No need to expose ports when using Traefik
  } else {
    services.homebox.ports = [`${config.port}:7745`]
  }

  // Add database configuration if PostgreSQL is selected
  if (config.databaseType === "postgres") {
    services.homebox.environment.push(
      "HBOX_DATABASE_DRIVER=postgres",
      `HBOX_DATABASE_HOST=${config.postgresConfig.host}`,
      `HBOX_DATABASE_PORT=${config.postgresConfig.port}`,
      `HBOX_DATABASE_USERNAME=${config.postgresConfig.username}`,
      `HBOX_DATABASE_PASSWORD=${config.postgresConfig.password}`,
      `HBOX_DATABASE_DATABASE=${config.postgresConfig.database}`,
    )

    // Add PostgreSQL service
    services["postgres"] = {
      image: "postgres:14",
      container_name: "homebox-postgres",
      restart: "always",
      environment: [
        `POSTGRES_USER=${config.postgresConfig.username}`,
        `POSTGRES_PASSWORD=${config.postgresConfig.password}`,
        `POSTGRES_DB=${config.postgresConfig.database}`,
      ],
      volumes: [],
    }

    // Configure postgres volumes based on storage type
    if (config.storageConfig.postgresStorage.type === "volume") {
      services.postgres.volumes.push(`${config.storageConfig.postgresStorage.volumeName}:/var/lib/postgresql/data`)
    } else {
      services.postgres.volumes.push(`${config.storageConfig.postgresStorage.directory}:/var/lib/postgresql/data`)
    }
  }

  // Add Traefik if selected
  if (config.useTraefik) {
    services["traefik"] = {
      image: "traefik:v2.10",
      container_name: "homebox-traefik",
      restart: "always",
      ports: ["80:80", "443:443"],
      command: [
        "--api.insecure=false",
        "--providers.docker=true",
        "--providers.docker.exposedbydefault=false",
        "--entrypoints.web.address=:80",
        "--entrypoints.web.http.redirections.entrypoint.to=websecure",
        "--entrypoints.web.http.redirections.entrypoint.scheme=https",
        "--entrypoints.websecure.address=:443",
        "--certificatesresolvers.letsencrypt.acme.tlschallenge=true",
        `--certificatesresolvers.letsencrypt.acme.email=${config.traefikEmail}`,
        "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json",
      ],
      volumes: ["/var/run/docker.sock:/var/run/docker.sock:ro"],
    }

    // Configure traefik volumes based on storage type
    if (config.storageConfig.traefikStorage.type === "volume") {
      services.traefik.volumes.push(`${config.storageConfig.traefikStorage.volumeName}:/letsencrypt`)
    } else {
      services.traefik.volumes.push(`${config.storageConfig.traefikStorage.directory}:/letsencrypt`)
    }
  }

  // Format the Docker Compose YAML
  let dockerCompose = "# generated by homebox config generator v0.0.1\n\nservices:\n"

  // Add services
  Object.entries(services).forEach(([serviceName, serviceConfig]) => {
    dockerCompose += `  ${serviceName}:\n`
    Object.entries(serviceConfig).forEach(([key, value]) => {
      if (Array.isArray(value)) {
        dockerCompose += `    ${key}:\n`
        value.forEach((item) => {
          dockerCompose += `      - ${item}\n`
        })
      } else {
        dockerCompose += `    ${key}: ${value}\n`
      }
    })
  })

  // Add volumes section if needed
  const volumeNames = []

  // Only add volumes that are configured as Docker volumes, not directories
  if (config.storageConfig.homeboxStorage.type === "volume") {
    volumeNames.push(config.storageConfig.homeboxStorage.volumeName)
  }

  if (config.databaseType === "postgres" && config.storageConfig.postgresStorage.type === "volume") {
    volumeNames.push(config.storageConfig.postgresStorage.volumeName)
  }

  if (config.useTraefik && config.storageConfig.traefikStorage.type === "volume") {
    volumeNames.push(config.storageConfig.traefikStorage.volumeName)
  }

  if (volumeNames.length > 0) {
    dockerCompose += "\nvolumes:\n"
    volumeNames.forEach((volumeName) => {
      dockerCompose += `  ${volumeName}:\n    driver: local\n`
    })
  }

  return dockerCompose
}

function copyToClipboard() {
  navigator.clipboard.writeText(generateDockerCompose())
  alert('Docker Compose configuration has been copied to your clipboard.')
}

function downloadConfig() {
  const element = document.createElement("a")
  const file = new Blob([generateDockerCompose()], { type: "text/plain" })
  element.href = URL.createObjectURL(file)
  element.download = "docker-compose.yml"
  document.body.appendChild(element)
  element.click()
  document.body.removeChild(element)
}
</script>

<style scoped>
.config-generator {
  font-family: var(--vp-font-family-base);
  color: var(--vp-c-text-1);
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

.config-layout {
  display: grid;
  grid-template-columns: 1fr;
  gap: 2rem;
}

.card {
  background-color: var(--vp-c-bg-soft);
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  margin-bottom: 1.5rem;
}

.card-header {
  padding: 1.5rem;
  border-bottom: 1px solid var(--vp-c-divider);
}

.card-title {
  font-size: 1.25rem;
  font-weight: 600;
  margin: 0 0 0.5rem;
  
  border-top: 0px;
  padding-top: 0px;
}

.card-title-with-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-actions {
  display: flex;
  gap: 0.5rem;
}

.card-description {
  color: var(--vp-c-text-2);
  font-size: 0.875rem;
  margin: 0;
}

.card-content {
  padding: 1.5rem;
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

.form-group {
  margin-bottom: 1.25rem;
}

.form-group label {
  display: block;
  font-weight: 500;
  margin-bottom: 0.5rem;
}

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid var(--vp-c-divider);
  border-radius: 4px;
  background-color: var(--vp-c-bg);
  color: var(--vp-c-text-1);
  font-size: 0.875rem;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
  outline: none;
  border-color: var(--vp-c-brand);
  box-shadow: 0 0 0 2px rgba(var(--vp-c-brand-rgb), 0.1);
}

.form-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.25rem;
}

.toggle-switch {
  position: relative;
  display: inline-block;
  width: 40px;
  height: 20px;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-switch label {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--vp-c-divider);
  transition: .4s;
  border-radius: 20px;
}

.toggle-switch label:before {
  position: absolute;
  content: "";
  height: 16px;
  width: 16px;
  left: 2px;
  bottom: 2px;
  background-color: white;
  transition: .4s;
  border-radius: 50%;
}

.toggle-switch input:checked + label {
  background-color: var(--vp-c-brand);
}

.toggle-switch input:checked + label:before {
  transform: translateX(20px);
}

.help-text {
  font-size: 0.75rem;
  color: var(--vp-c-text-2);
  margin-top: 0.25rem;
}

.separator {
  height: 1px;
  background-color: var(--vp-c-divider);
  margin: 1.5rem 0;
}

.nested-form {
  margin-top: 1rem;
  padding: 1rem;
  border: 1px solid var(--vp-c-divider);
  border-radius: 4px;
}

.password-input {
  display: flex;
  gap: 0.5rem;
}

.password-input input {
  flex: 1;
}

.icon-button {
  padding: 0.5rem;
  background-color: var(--vp-c-bg-mute);
  border: 1px solid var(--vp-c-divider);
  border-radius: 4px;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.icon-button:hover {
  background-color: var(--vp-c-bg-alt);
}

.storage-selector {
  margin-bottom: 2rem;
  padding: 1rem;
  border: 1px solid var(--vp-c-divider);
  border-radius: 4px;
}

.storage-selector h3 {
  font-size: 1rem;
  font-weight: 600;
  margin: 0 0 0.5rem;
}

.radio-group {
  display: flex;
  gap: 1.5rem;
  margin: 1rem 0;
}

.radio-option {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.code-preview {
  width: 100%;
  height: 600px;
  font-family: monospace;
  padding: 1rem;
  border: 1px solid var(--vp-c-divider);
  border-radius: 4px;
  background-color: var(--vp-c-bg);
  color: var(--vp-c-text-1);
  resize: none;
  white-space: pre;
  overflow: auto;
}
</style>