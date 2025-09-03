<template>
  <div class="tab-content">
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
                @click="$emit('togglePassword')"
                type="button"
              >
                <span v-if="showPassword">Hide</span>
                <span v-else>Show</span>
              </button>
              <button 
                class="icon-button" 
                @click="$emit('regeneratePassword')"
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
</template>

<script setup>
defineProps({
  config: {
    type: Object,
    required: true
  },
  showPassword: {
    type: Boolean,
    default: false
  }
})

defineEmits(['togglePassword', 'regeneratePassword'])
</script>

<style scoped>
@import './common.css';

.password-input {
  display: flex;
  gap: 0.5rem;
}

.password-input input {
  flex: 1;
}
</style>