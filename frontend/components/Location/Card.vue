<template>
  <Card>
    <NuxtLink :to="`/location/${location.id}`" class="group/location-card transition duration-300">
      <div
        :class="{
          'p-4': !dense,
          'px-3 py-2': dense,
        }"
      >
        <h2 class="flex items-center justify-between gap-2">
          <div class="relative size-6">
            <div
              class="absolute inset-0 flex items-center justify-center transition-transform duration-300 group-hover/location-card:-rotate-90"
            >
              <MdiMapMarkerOutline class="size-6 group-hover/location-card:hidden" />
              <MdiArrowUp class="hidden size-6 group-hover/location-card:block" />
            </div>
          </div>
          <span class="mx-auto">
            {{ location.name }}
          </span>
          <Badge :class="{ 'opacity-0': !hasCount }">
            {{ count }}
          </Badge>
        </h2>
      </div>
    </NuxtLink>
  </Card>
</template>

<script lang="ts" setup>
  import type { LocationOut, LocationOutCount, LocationSummary } from "~~/lib/api/types/data-contracts";
  import MdiArrowUp from "~icons/mdi/arrow-down";
  import MdiMapMarkerOutline from "~icons/mdi/map-marker-outline";
  import { Card } from "@/components/ui/card";
  import { Badge } from "@/components/ui/badge";

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
</script>
