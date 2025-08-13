<template>
  <div class="config-preview">
    <div class="card">
      <div class="card-header">
        <div class="card-title-with-actions">
          <h2 class="card-title">Docker Compose Configuration</h2>
          <div class="card-actions">
            <button class="icon-button" @click="$emit('copy')" title="Copy to clipboard">
              Copy
            </button>
            <button class="icon-button" @click="$emit('download')" title="Download as file">
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
          :value="config"
        ></textarea>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  config: {
    type: String,
    required: true
  },
})

defineEmits(['copy', 'download'])
</script>

<style scoped>
@import './common.css';

.config-preview {
  height: 100%;
}

.card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.card-content {
  flex: 1;
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