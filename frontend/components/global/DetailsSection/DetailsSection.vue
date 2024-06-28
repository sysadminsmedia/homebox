<template>
  <div class="border-t border-gray-300 px-4 py-5 sm:p-0">
    <dl class="sm:divide-y sm:divide-gray-300">
      <div v-for="(detail, i) in details" :key="i" class="py-4 sm:grid group sm:grid-cols-3 sm:gap-4 sm:px-6">
        <dt class="text-sm font-medium text-base-content">
          {{ detail.name }}
        </dt>
        <dd class="text-sm text-base-content text-start sm:col-span-2">
          <slot :name="detail.slot || detail.name" v-bind="{ detail }">
            <DateTime
              v-if="detail.type == 'date'"
              :date="detail.text"
              :datetime-type="detail.date ? 'date' : 'datetime'"
            />
            <Currency v-else-if="detail.type == 'currency'" :amount="detail.text" />
            <template v-else-if="detail.type === 'link'">
              <div class="tooltip tooltip-primary tooltip-right" :data-tip="detail.href">
                <a class="btn btn-primary btn-xs" :href="detail.href" target="_blank">
                  <MdiOpenInNew class="mr-2 swap-on" />
                  {{ detail.text }}
                </a>
              </div>
            </template>
            <template v-else-if="detail.type === 'markdown'">
              <ClientOnly>
                <Markdown :source="detail.text" />
              </ClientOnly>
            </template>
            <template v-else>
              <span class="flex items-center">
                {{ detail.text }}
                <span
                  v-if="detail.copyable"
                  class="opacity-0 group-hover:opacity-100 ml-4 my-0 duration-75 transition-opacity"
                >
                  <CopyText
                    v-if="detail.text.toString()"
                    :text="detail.text.toString()"
                    :icon-size="16"
                    class="btn btn-xs btn-ghost btn-circle"
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
