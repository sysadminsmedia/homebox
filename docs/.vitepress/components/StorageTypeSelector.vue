<template>
  <div class="storage-selector">
    <h3>{{ label }}</h3>
    <p class="help-text">{{ description }}</p>
    
    <div class="radio-group">
      <div class="radio-option">
        <input 
          type="radio" 
          :id="`${storageKey}-volume`" 
          value="volume" 
          v-model="config.storageConfig[storageKey].type"
        />
        <label :for="`${storageKey}-volume`">Docker Volume</label>
      </div>
      <div class="radio-option">
        <input 
          type="radio" 
          :id="`${storageKey}-directory`" 
          value="directory" 
          v-model="config.storageConfig[storageKey].type"
        />
        <label :for="`${storageKey}-directory`">Host Directory</label>
      </div>
    </div>

    <div v-if="config.storageConfig[storageKey].type === 'volume'" class="form-group">
      <label :for="`${storageKey}-volume-name`">Volume Name</label>
      <input 
        type="text" 
        :id="`${storageKey}-volume-name`" 
        v-model="config.storageConfig[storageKey].volumeName"
      />
    </div>
    <div v-else class="form-group">
      <label :for="`${storageKey}-directory-path`">Directory Path</label>
      <input 
        type="text" 
        :id="`${storageKey}-directory-path`" 
        v-model="config.storageConfig[storageKey].directory"
      />
      <p class="help-text">Absolute path recommended (e.g., /home/user/data)</p>
    </div>
  </div>
</template>

<script setup>
defineProps({
  storageKey: {
    type: String,
    required: true
  },
  label: {
    type: String,
    required: true
  },
  description: {
    type: String,
    required: true
  },
  config: {
    type: Object,
    required: true
  }
})
</script>

<style scoped>
@import './common.css';

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
</style>