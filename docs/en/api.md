---
layout: page
sidebar: false
---

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue';
import { useData } from 'vitepress';

// Reactive flags
const apiSpec = ref(null);
const stoplightLoaded = ref(false);
const componentKey = ref(0);
const demoBaseUrl = "https://demo.homebox.software/api";

// Fetch and patch the OpenAPI spec
async function fetchSpec() {
  try {
    const res = await fetch('https://cdn.jsdelivr.net/gh/sysadminsmedia/homebox@main/docs/docs/api/openapi-2.0.json');
    const spec = await res.json();
    // Override the host and basePath
    spec.host = "demo.homebox.software";
    spec.basePath = "/api";
    apiSpec.value = spec;
  } catch (error) {
    console.error("Error fetching the OpenAPI spec:", error);
  }
}

// Load the Stoplight Elements script and wait for it to load
function loadStoplightScript() {
  return new Promise<void>((resolve, reject) => {
    // Only load if not already loaded
    if (document.querySelector('script[src="https://unpkg.com/@stoplight/elements/web-components.min.js"]')) {
      resolve();
      return;
    }
    const script = document.createElement('script');
    script.src = 'https://unpkg.com/@stoplight/elements/web-components.min.js';
    script.onload = () => resolve();
    script.onerror = () => reject(new Error('Failed to load Stoplight Elements script.'));
    document.head.appendChild(script);
  });
}

// Load the stylesheet (we can append it without waiting if desired)
function loadStoplightStyles() {
  if (!document.querySelector('link[href="https://unpkg.com/@stoplight/elements/styles.min.css"]')) {
    const link = document.createElement('link');
    link.rel = 'stylesheet';
    link.href = 'https://unpkg.com/@stoplight/elements/styles.min.css';
    document.head.appendChild(link);
  }
}

// Refresh on hash change
const handleHashChange = () => {
  componentKey.value++;
};

onMounted(async () => {
  window.addEventListener('hashchange', handleHashChange);
  loadStoplightStyles();
  try {
    await loadStoplightScript();
    stoplightLoaded.value = true;
  } catch (error) {
    console.error(error);
  }
  await fetchSpec();
});

onBeforeUnmount(() => {
  window.removeEventListener('hashchange', handleHashChange);
});

// Watch for dark mode changes to force a re-render
const { isDark } = useData();
const theme = ref(isDark.value ? 'dark' : 'light');
watch(isDark, (newVal) => {
  theme.value = newVal ? 'dark' : 'light';
  componentKey.value++;
});
</script>

<template>
  <client-only>
    <!-- Wait until both the API spec and Stoplight Elements are loaded -->
    <div v-if="!apiSpec || !stoplightLoaded">
      Loading API Documentation...
    </div>
    <div v-else>
      <elements-api
        :key="componentKey"
        :apiDescription="apiSpec"
        router="hash"
        layout="responsive"
        hideSchemas="true"
        :data-theme="theme"
        tryItBaseUrl="https://demo.homebox.software/api"
      />
    </div>
  </client-only>
</template>