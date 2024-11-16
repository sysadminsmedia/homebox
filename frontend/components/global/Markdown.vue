<script setup lang="ts">
  import MarkdownIt from "markdown-it";
  import DOMPurify from "dompurify";

  type Props = {
    source: string | null | undefined;
  };

  const props = withDefaults(defineProps<Props>(), {
    source: null,
  });

  const md = new MarkdownIt({
    breaks: true,
    html: true,
    linkify: true,
    typographer: true,
  });

  const raw = computed(() => {
    const html = md.render(props.source || "");
    return DOMPurify.sanitize(html);
  });
</script>

<template>
  <!-- eslint-disable-next-line vue/no-v-html -->
  <div class="markdown text-wrap" v-html="raw"></div>
</template>

<style scoped>
  * {
    --y-gap: 0.65rem;
  }
</style>
