---
layout: page
sidebar: false
---

<script setup lang="ts">
import { useData } from 'vitepress';
import { onBeforeRouteUpdate } from 'vue-router';
import { ref } from 'vue';

// Create a key ref to force re-render of the elements-api component
const componentKey = ref(0);

onBeforeRouteUpdate((to, from, next) => {
  // Increment the key to trigger re-initialization on route change
  componentKey.value++;
  next();
});

const elementScript = document.createElement('script');
elementScript.src = 'https://unpkg.com/@stoplight/elements/web-components.min.js';
document.head.appendChild(elementScript);

const elementStyle = document.createElement('link');
elementStyle.rel = 'stylesheet';
elementStyle.href = 'https://unpkg.com/@stoplight/elements/styles.min.css';
document.head.appendChild(elementStyle);

const { isDark } = useData();
let theme = 'light';
if (isDark.value) {
  theme = 'dark';
}
</script>

<style>
.TryItPanel {
  display: none;
}
</style>

<client-only>
  <elements-api
    :key="componentKey"
    apiDescriptionUrl="https://cdn.jsdelivr.net/gh/sysadminsmedia/homebox@main/docs/docs/api/openapi-2.0.json"
    layout="responsive"
    hideSchemas="true"
    :data-theme="theme"
  />
</client-only>
