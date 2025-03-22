---
layout: page
sidebar: false
---

<script setup lang="ts">
import { ref, onMounted, watch, onBeforeUnmount } from 'vue';
import { useData } from 'vitepress';

const apiSpec = ref(null);
const demoBaseUrl = "https://demo.homebox.software/api";

// Fetch and patch the spec on mount
onMounted(async () => {
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
});

// For navigation refresh
const componentKey = ref(0);
const handleHashChange = () => {
  componentKey.value++;
};

onMounted(() => {
  window.addEventListener('hashchange', handleHashChange);
});
onBeforeUnmount(() => {
  window.removeEventListener('hashchange', handleHashChange);
});

// Handle theme changes to re-render the component
const { isDark } = useData();
const theme = ref(isDark.value ? 'dark' : 'light');
watch(isDark, (newVal) => {
  theme.value = newVal ? 'dark' : 'light';
  componentKey.value++;
});

// Append Stoplight Elements script and stylesheet
const elementScript = document.createElement('script');
elementScript.src = 'https://unpkg.com/@stoplight/elements/web-components.min.js';
document.head.appendChild(elementScript);

const elementStyle = document.createElement('link');
elementStyle.rel = 'stylesheet';
elementStyle.href = 'https://unpkg.com/@stoplight/elements/styles.min.css';
document.head.appendChild(elementStyle);
</script>

<client-only>
  <!-- Render only after the spec is loaded -->
  <elements-api
    v-if="apiSpec"
    :key="componentKey"
    :apiDescription="apiSpec"
    router="hash"
    layout="responsive"
    hideSchemas="true"
    :data-theme="theme"
    tryItBaseUrl="https://demo.homebox.software/api"
  />
</client-only>
