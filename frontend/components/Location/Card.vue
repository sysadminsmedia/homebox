<template>
  <Card class="relative overflow-hidden">
    <NuxtLink :to="`/location/${location.id}`" class="group/location-card transition duration-300">
      <div v-if="imageUrl" class="relative h-[140px]">
        <img
          class="absolute h-[140px] w-full object-cover blur-md"
          loading="lazy"
          :src="imageUrl"
          alt=""
        />
        <img
          class="absolute h-[140px] w-full object-cover shadow-md"
          loading="lazy"
          :src="imageUrl"
          :alt="location.name"
        />
      </div>
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
          <span class="mx-auto font-semibold">
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
  import type { EntityOut, EntitySummary } from "~~/lib/api/types/data-contracts";
  import MdiArrowUp from "~icons/mdi/arrow-down";
  import MdiMapMarkerOutline from "~icons/mdi/map-marker-outline";
  import { Card } from "@/components/ui/card";
  import { Badge } from "@/components/ui/badge";

  const api = useUserApi();

  const props = defineProps({
    location: {
      type: Object as () => EntitySummary | EntityOut,
      required: true,
    },
    dense: {
      type: Boolean,
      default: false,
    },
  });

  const imageUrl = computed(() => {
    const loc = props.location as (EntitySummary | EntityOut) & { imageId?: string; thumbnailId?: string };
    if (!loc.imageId && !loc.thumbnailId) {
      return null;
    }
    if (loc.thumbnailId) {
      return api.authURL(`/entities/${loc.id}/attachments/${loc.thumbnailId}`);
    } else {
      return api.authURL(`/entities/${loc.id}/attachments/${loc.imageId}`);
    }
  });

  const hasCount = computed(() => {
    return !!(props.location as EntitySummary).itemCount;
  });

  const count = computed(() => {
    return hasCount.value ? (props.location as EntitySummary).itemCount : undefined;
  });
</script>

