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
  <div class="markdown text-wrap" v-html="raw"></div>
</template>

<style scoped>
  * {
    --y-gap: 0.65rem;
  }

  .markdown > :first-child {
    margin-top: 0px !important;
  }

  .markdown :where(p, ul, ol, dl, blockquote, h1, h2, h3, h4, h5, h6) {
    margin-top: var(--y-gap);
    margin-bottom: var(--y-gap);
  }

  .markdown :where(ul) {
    list-style: disc;
    margin-left: 2rem;
  }

  .markdown :where(ol) {
    list-style: decimal;
    margin-left: 2rem;
  }
  /* Heading Styles */
  .markdown :where(h1) {
    font-size: 2rem;
    font-weight: 700;
  }

  .markdown :where(h2) {
    font-size: 1.5rem;
    font-weight: 700;
  }

  .markdown :where(h3) {
    font-size: 1.25rem;
    font-weight: 700;
  }

  .markdown :where(h4) {
    font-size: 1rem;
    font-weight: 700;
  }

  .markdown :where(h5) {
    font-size: 0.875rem;
    font-weight: 700;
  }

  .markdown :where(h6) {
    font-size: 0.75rem;
    font-weight: 700;
  }
</style>
