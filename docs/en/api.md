---
layout: page
sidebar: false
---

<script setup lang="ts">
import { useData } from 'vitepress';

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

<elements-api
apiDescriptionUrl="https://cdn.jsdelivr.net/gh/sysadminsmedia/homebox@main/docs/docs/api/openapi-2.0.json"
router="hash"
layout="responsive"
hideSchemas="true"
:data-theme="theme"
/>