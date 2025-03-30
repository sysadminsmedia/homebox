<template>
  <div class="border-t border-gray-300 px-4 py-5 sm:p-0">
    <dl class="sm:divide-y sm:divide-gray-300">
      <div v-for="(detail, i) in details" :key="i" class="group py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
        <dt class="text-base-content text-sm font-medium">
          {{ $t(detail.name) }}
        </dt>
        <dd class="text-base-content text-start text-sm sm:col-span-2">
          <slot :name="detail.slot || detail.name" v-bind="{ detail }">
            <DateTime
              v-if="detail.type == 'date'"
              :date="detail.text"
              :datetime-type="detail.date ? 'date' : 'datetime'"
            />
            <Currency v-else-if="detail.type == 'currency'" :amount="detail.text" />
            <template v-else-if="detail.type === 'link'">
              <div class="tooltip tooltip-top tooltip-primary" :data-tip="detail.href">
                <a class="btn btn-primary btn-xs" :href="detail.href" target="_blank">
                  <MdiOpenInNew class="swap-on mr-2" />
                  {{ detail.text }}
                </a>
              </div>
            </template>
            <template v-else-if="detail.type === 'markdown'">
              <ClientOnly>
                <div class="markdown-container w-full overflow-hidden break-words">
                  <Markdown :source="detail.text" />
                </div>
              </ClientOnly>
            </template>
            <template v-else>
              <!-- Fixed version with improved overflow handling -->
              <span class="flex w-full items-center break-words">
                <span class="overflow-hidden break-all">{{ detail.text }}</span>
                <span
                  v-if="detail.copyable"
                  class="my-0 ml-4 shrink-0 opacity-0 transition-opacity duration-75 group-hover:opacity-100"
                >
                  <CopyText
                    v-if="detail.text.toString()"
                    :text="detail.text.toString()"
                    :icon-size="16"
                    class="btn btn-circle btn-ghost btn-xs"
                  />
                </span>
              </span>
            </template>
          </slot>
        </dd>
      </div>
    </dl>
  </div>
</template>

<script setup lang="ts">
  import type { AnyDetail, Detail } from "./types";
  import MdiOpenInNew from "~icons/mdi/open-in-new";

  defineProps({
    details: {
      type: Object as () => (Detail | AnyDetail)[],
      required: true,
    },
  });
</script>

<style>
  /* Use :deep to target elements inside the markdown container */
  :deep(.markdown-container) {
    /* General container styles */
    width: 100%;
    overflow-wrap: break-word;
    word-wrap: break-word;

    /* Handle tables */
    table {
      table-layout: fixed;
      width: 100%;
    }

    td,
    th {
      word-break: break-word;
      overflow-wrap: break-word;
      max-width: 100%;
    }

    /* Handle code blocks */
    pre,
    code {
      white-space: pre-wrap;
      word-break: break-all;
      overflow-x: auto;
    }

    /* Handle images */
    img {
      max-width: 100%;
      height: auto;
    }

    /* Handle headings */
    h1,
    h2,
    h3,
    h4,
    h5,
    h6 {
      word-break: break-word;
      overflow-wrap: break-word;
    }

    /* Handle blockquotes */
    blockquote {
      overflow-wrap: break-word;
      word-break: break-word;
    }

    /* Handle inline elements */
    a,
    strong,
    em {
      word-break: break-all;
      overflow-wrap: break-word;
    }
  }

  /* Non-scoped styles for regular text */
  .break-all {
    word-break: break-all;
    max-width: 100%;
  }

  /* Handle very long words */
  pre,
  code,
  a,
  p,
  span,
  div,
  td,
  th,
  li,
  blockquote,
  h1,
  h2,
  h3,
  h4,
  h5,
  h6 {
    overflow-wrap: break-word;
    word-wrap: break-word;
    -ms-word-break: break-all;
    word-break: break-all;
    word-break: break-word;
    -ms-hyphens: auto;
    -moz-hyphens: auto;
    -webkit-hyphens: auto;
    hyphens: auto;
  }
</style>
