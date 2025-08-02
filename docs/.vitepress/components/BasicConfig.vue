<template>
  <div class="tab-content">
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
          <p class="help-text">Only used if HTTPS is not enabled</p>
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
@import 'common.css';
</style>