<template>
  <div class="border-t border-gray-300 px-4 py-5 sm:p-0">
    <dl class="sm:divide-y sm:divide-gray-300">
      <div v-for="(detail, i) in details" :key="i" class="group py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
        <dt class="text-sm font-medium text-base-content">
          {{ detail.name }}
        </dt>
        <dd class="text-start text-sm text-base-content sm:col-span-2">
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
                <Markdown :source="detail.text" />
              </ClientOnly>
            </template>
            <template v-else>
              <span class="flex items-center text-wrap">
                {{ detail.text }}
                <span
                  v-if="detail.copyable"
                  class="my-0 ml-4 opacity-0 transition-opacity duration-75 group-hover:opacity-100"
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
