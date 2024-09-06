<template>
  <NuxtLink
    ref="card"
    :to="`/location/${location.id}`"
    class="card rounded-md bg-base-100 text-base-content shadow-md transition duration-300"
  >
    <div
      class="card-body"
      :class="{
        'p-4': !dense,
        'px-3 py-2': dense,
      }"
    >
      <h2 class="flex items-center justify-between gap-2">
        <label class="swap swap-rotate" :class="isActive ? 'swap-active' : ''">
          <MdiArrowRight class="swap-on size-6" />
          <MdiMapMarkerOutline class="swap-off size-6" />
        </label>
        <span class="mx-auto">
          {{ location.name }}
        </span>
        <span class="badge badge-primary badge-lg h-6" :class="{ 'opacity-0': !hasCount }">
          {{ count }}
        </span>
      </h2>
    </div>
  </NuxtLink>
</template>

<script lang="ts" setup>
  import type { LocationOut, LocationOutCount, LocationSummary } from "~~/lib/api/types/data-contracts";
  import MdiArrowRight from "~icons/mdi/arrow-right";
  import MdiMapMarkerOutline from "~icons/mdi/map-marker-outline";

  const props = defineProps({
    location: {
      type: Object as () => LocationOutCount | LocationOut | LocationSummary,
      required: true,
    },
    dense: {
      type: Boolean,
      default: false,
    },
  });

  const hasCount = computed(() => {
    return !!(props.location as LocationOutCount).itemCount;
  });

  const count = computed(() => {
    if (hasCount.value) {
      return (props.location as LocationOutCount).itemCount;
    }
  });

  const card = ref(null);
  const isHover = useElementHover(card);
  const { focused } = useFocus(card);

  const isActive = computed(() => isHover.value || focused.value);
</script>
