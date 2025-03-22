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

// Function to fetch and patch the OpenAPI spec
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

// Function to load the Stoplight Elements script
function loadStoplightScript() {
  return new Promise<void>((resolve, reject) => {
    // Check if the script is already loaded
    if (window.customElements && window.customElements.get('elements-api')) {
      resolve();
      return;
    }
    const script = document.createElement('script');
    script.src = 'https://unpkg.com/@stoplight/elements/web-components.min.js';
    script.onload = () => {
      resolve();
    };
    script.onerror = (err) => {
      reject(err);
    };
    document.head.appendChild(script);

    // Also load the stylesheet
    const link = document.createElement('link');
    link.rel = 'stylesheet';
    link.href = 'https://unpkg.com/@stoplight/elements/styles.min.css';
    document.head.appendChild(link);
  });
}

// Listen for hash changes to force re-render
const handleHashChange = () => {
  componentKey.value++;
};

onMounted(() => {
  window.addEventListener('hashchange', handleHashChange);
  fetchSpec();
  loadStoplightScript()
    .then(() => {
      stoplightLoaded.value = true;
    })
    .catch((err) => {
      console.error("Error loading Stoplight script:", err);
    });
});

onBeforeUnmount(() => {
  window.removeEventListener('hashchange', handleHashChange);
});

// Watch for dark mode changes to force re-render
const { isDark } = useData();
const theme = ref(isDark.value ? 'dark' : 'light');
watch(isDark, (newVal) => {
  theme.value = newVal ? 'dark' : 'light';
  componentKey.value++;
});
</script>

<template>
  <client-only>
    <div v-if="apiSpec && stoplightLoaded">
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
    <div v-else>
      Loading API Spec and Stoplight...
    </div>
  </client-only>
</template>