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
  <div class="markdown text-wrap break-words" v-html="raw"></div>
</template>

<style scoped>
  * {
    --y-gap: 0.65rem;
    word-wrap: break-word; /*Fix for long words going out of emelent bounds and issue #407 */
    overflow-wrap: break-word; /*Fix for long words going out of emelent bounds and issue #407 */
    white-space: pre-wrap; /*Fix for long words going out of emelent bounds and issue #407 */
  }
  .markdown {
    max-width: 100%;
    overflow: hidden; /*Fix for long words going out of emelent bounds and issue #407 */
  }
</style>
