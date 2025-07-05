<template>
  <div class="tab-content">
    <div class="card">
      <div class="card-header">
        <h2 class="card-title">HTTPS Configuration</h2>
        <p class="card-description">Configure HTTPS for your Homebox instance.</p>
      </div>
      <div class="card-content">
        <div class="form-group">
          <label for="httpsOption">HTTPS Option</label>
          <select id="httpsOption" v-model="config.httpsOption">
            <option value="none">None (HTTP only)</option>
            <option value="traefik">Traefik (Automatic HTTPS with Let's Encrypt)</option>
            <option value="nginx">Nginx (Custom SSL certificates)</option>
            <option value="caddy">Caddy (Automatic HTTPS with Let's Encrypt)</option>
            <option value="cloudflared">Cloudflare Tunnel</option>
          </select>
        </div>

        <!-- Traefik Configuration -->
        <div v-if="config.httpsOption === 'traefik'" class="nested-form">
          <h3>Traefik Configuration</h3>
          <p class="help-text">Traefik automatically handles HTTPS certificates via Let's Encrypt</p>
          
          <div class="form-group">
            <label for="traefikDomain">Domain Name</label>
            <input 
              type="text" 
              id="traefikDomain" 
              v-model="config.traefikConfig.domain"
              placeholder="homebox.example.com"
            />
            <p class="help-text">The domain name must be pointed to your server's IP address</p>
          </div>

          <div class="form-group">
            <label for="traefikEmail">Email Address (for Let's Encrypt)</label>
            <input 
              type="email" 
              id="traefikEmail" 
              v-model="config.traefikConfig.email"
              placeholder="your-email@example.com"
            />
            <p class="help-text">Required for Let's Encrypt certificate notifications</p>
          </div>
        </div>

        <!-- Nginx Configuration -->
        <div v-if="config.httpsOption === 'nginx'" class="nested-form">
          <h3>Nginx Configuration</h3>
          <p class="help-text">Nginx requires you to provide SSL certificates</p>
          
          <div class="form-group">
            <label for="nginxDomain">Domain Name</label>
            <input 
              type="text" 
              id="nginxDomain" 
              v-model="config.nginxConfig.domain"
              placeholder="homebox.example.com"
            />
          </div>

          <div class="form-group">
            <label for="nginxPort">HTTPS Port</label>
            <input 
              type="text" 
              id="nginxPort" 
              v-model="config.nginxConfig.port"
            />
          </div>

          <div class="form-group">
            <label for="nginxSslCert">SSL Certificate Path</label>
            <input 
              type="text" 
              id="nginxSslCert" 
              v-model="config.nginxConfig.sslCertPath"
            />
            <p class="help-text">Path to SSL certificate file inside the Nginx container</p>
          </div>

          <div class="form-group">
            <label for="nginxSslKey">SSL Key Path</label>
            <input 
              type="text" 
              id="nginxSslKey" 
              v-model="config.nginxConfig.sslKeyPath"
            />
            <p class="help-text">Path to SSL key file inside the Nginx container</p>
          </div>
        </div>

        <!-- Caddy Configuration -->
        <div v-if="config.httpsOption === 'caddy'" class="nested-form">
          <h3>Caddy Configuration</h3>
          <p class="help-text">Caddy automatically handles HTTPS certificates via Let's Encrypt</p>
          
          <div class="form-group">
            <label for="caddyDomain">Domain Name</label>
            <input 
              type="text" 
              id="caddyDomain" 
              v-model="config.caddyConfig.domain"
              placeholder="homebox.example.com"
            />
            <p class="help-text">The domain name must be pointed to your server's IP address</p>
          </div>

          <div class="form-group">
            <label for="caddyEmail">Email Address (for Let's Encrypt)</label>
            <input 
              type="email" 
              id="caddyEmail" 
              v-model="config.caddyConfig.email"
              placeholder="your-email@example.com"
            />
            <p class="help-text">Optional: Used for Let's Encrypt certificate notifications</p>
          </div>
        </div>

        <!-- Cloudflared Configuration -->
        <div v-if="config.httpsOption === 'cloudflared'" class="nested-form">
          <h3>Cloudflare Tunnel Configuration</h3>
          <p class="help-text">Cloudflare Tunnel provides secure access without exposing ports</p>
          
          <div class="form-group">
            <label for="cloudflaredTunnel">Tunnel Name</label>
            <input 
              type="text" 
              id="cloudflaredTunnel" 
              v-model="config.cloudflaredConfig.tunnel"
            />
          </div>

          <div class="form-group">
            <label for="cloudflaredDomain">Domain Name</label>
            <input 
              type="text" 
              id="cloudflaredDomain" 
              v-model="config.cloudflaredConfig.domain"
              placeholder="homebox.example.com"
            />
            <p class="help-text">The domain must be managed by Cloudflare</p>
          </div>

          <div class="form-group">
            <label for="cloudflaredToken">Tunnel Token</label>
            <input 
              type="password" 
              id="cloudflaredToken" 
              v-model="config.cloudflaredConfig.token"
              placeholder="Your Cloudflare Tunnel token"
            />
            <p class="help-text">Create a tunnel in the Cloudflare Zero Trust dashboard to get a token</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  config: {
    type: Object,
    required: true
  }
})
</script>

<style scoped>
@import './common.css';

h3 {
  font-size: 1rem;
  font-weight: 600;
  margin: 0 0 0.5rem;
}
</style>