<template>
  <div class="card rounded-lg bg-base-100 shadow-xl">
    <div v-if="$slots.title" class="px-4 py-5 sm:px-6">
      <component :is="collapsable ? 'button' : 'div'" v-on="collapsable ? { click: toggle } : {}">
        <h3 class="flex items-center text-lg font-medium leading-6">
          <slot name="title"></slot>
          <template v-if="collapsable">
            <span class="swap swap-rotate ml-2" :class="`${collapsed ? 'swap-active' : ''}`">
              <MdiChevronRight class="swap-on size-6" />
              <MdiChevronDown class="swap-off size-6" />
            </span>
          </template>
        </h3>
      </component>
      <div>
        <p v-if="$slots.subtitle" class="mt-1 max-w-2xl text-sm text-gray-500">
          <slot name="subtitle"></slot>
        </p>
        <template v-if="$slots['title-actions']">
          <slot name="title-actions"></slot>
        </template>
      </div>
    </div>
    <div
      :class="{
        'max-h-[9000px]': collapsable && !collapsed,
        'max-h-0 overflow-hidden': collapsed,
      }"
      class="transition-[max-height] duration-200"
    >
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
  import MdiChevronDown from "~icons/mdi/chevron-down";
  import MdiChevronRight from "~icons/mdi/chevron-right";

  defineProps<{
    collapsable?: boolean;
  }>();

  function toggle() {
    collapsed.value = !collapsed.value;
  }

  const collapsed = ref(false);
</script>
