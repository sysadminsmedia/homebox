<script setup lang="ts">
  import { computed } from "vue";
  import type { ItemSummary } from "~/lib/api/types/data-contracts";
  import LabelChip from "@/components/Label/Chip.vue";
  import Badge from "~/components/ui/badge/Badge.vue";

  const props = defineProps<{
    item: ItemSummary;
  }>();

  const api = useUserApi();

  const imageUrl = computed(() => {
    if (!props.item.imageId) {
      return "/no-image.jpg";
    }
    if (props.item.thumbnailId) {
      return api.authURL(`/items/${props.item.id}/attachments/${props.item.thumbnailId}`);
    } else {
      return api.authURL(`/items/${props.item.id}/attachments/${props.item.imageId}`);
    }
  });
</script>

<template>
  <div class="flex items-start gap-3">
    <div class="shrink-0">
      <img :src="imageUrl" class="size-32 rounded-lg bg-muted object-cover" />
    </div>
    <div class="flex min-w-0 flex-1 flex-col gap-2">
      <h2 class="truncate text-xl font-bold">{{ item.name }}</h2>
      <Badge class="w-min text-nowrap bg-secondary text-secondary-foreground hover:bg-secondary/70 hover:underline">
        <NuxtLink v-if="item.location" :to="`/location/${item.location.id}`">
          {{ item.location.name }}
        </NuxtLink>
      </Badge>
      <div class="flex flex-wrap gap-2">
        <LabelChip v-for="tag in item.tags" :key="tag.id" :tag="tag" size="sm" />
      </div>
      <p class="whitespace-pre-line break-words text-sm text-muted-foreground">
        {{ item.description || $t("components.item.no_description") }}
      </p>
    </div>
  </div>
</template>
