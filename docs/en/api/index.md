---
layout: page
sidebar: false
---

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount } from 'vue';
import { useData } from 'vitepress';

// Reactive key for re-rendering the elements-api component
const componentKey = ref(0);

// Set BaseURL
const BaseURL = "https://demo.homebox.software/api";

// Access dark mode setting from VitePress
const { isDark } = useData();
const theme = ref(isDark.value ? 'dark' : 'light');

// Watch for changes to the dark mode value and force a re-render when it changes
watch(isDark, (newVal) => {
  theme.value = newVal ? 'dark' : 'light';
  // Increment key to force a refresh of the Stoplight component and its CSS
  componentKey.value++;
});

// Use a native hashchange listener (as before) to refresh on navigation changes
const handleHashChange = () => {
  componentKey.value++;
};

onMounted(() => {
  window.addEventListener('hashchange', handleHashChange);
});
onBeforeUnmount(() => {
  window.removeEventListener('hashchange', handleHashChange);
});

// Append the Stoplight Elements script and stylesheet
const elementScript = document.createElement('script');
elementScript.src = 'https://unpkg.com/@stoplight/elements/web-components.min.js';
document.head.appendChild(elementScript);

const elementStyle = document.createElement('link');
elementStyle.rel = 'stylesheet';
elementStyle.href = 'https://unpkg.com/@stoplight/elements/styles.min.css';
document.head.appendChild(elementStyle);
</script>

<client-only>
  <elements-api
    :key="componentKey"
    apiDescriptionUrl="https://raw.githubusercontent.com/sysadminsmedia/homebox/refs/heads/main/docs/en/api/openapi-2.0.json"
    router="hash"
    layout="responsive"
    hideSchemas="true"
    hideTryIt="true"
    :data-theme="theme"
    :tryItBaseUrl="BaseURL"
  />
</client-only>
