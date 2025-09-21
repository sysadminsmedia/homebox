<template>
  <Card class="relative overflow-hidden">
    <div v-if="tableRow" class="absolute left-1 top-1 z-10">
      <Checkbox
        class="size-5 bg-accent hover:bg-background-accent"
        :model-value="tableRow.getIsSelected()"
        :aria-label="$t('components.item.view.selectable.select_card')"
        @update:model-value="tableRow.toggleSelected()"
      />
    </div>
    <NuxtLink :to="`/item/${item.id}`">
      <div class="relative h-[200px]">
        <img v-if="imageUrl" class="h-[200px] w-full object-cover shadow-md" loading="lazy" :src="imageUrl" alt="" />
        <div class="absolute inset-x-1 bottom-1">
          <Badge class="text-wrap bg-secondary text-secondary-foreground hover:bg-secondary/70 hover:underline">
            <NuxtLink v-if="item.location" :to="`/location/${item.location.id}`">
              {{ locationString }}
            </NuxtLink>
          </Badge>
        </div>
      </div>
      <div class="col-span-4 flex grow flex-col gap-y-1 p-4 pt-2">
        <h2 class="line-clamp-2 text-ellipsis text-wrap text-lg font-bold">{{ item.name }}</h2>
        <Separator class="mb-1" />
        <TooltipProvider :delay-duration="0">
          <div class="flex items-center gap-2">
            <Tooltip v-if="item.insured">
              <TooltipTrigger>
                <MdiShieldCheck class="size-5 text-primary" />
              </TooltipTrigger>
              <TooltipContent>
                {{ $t("global.insured") }}
              </TooltipContent>
            </Tooltip>
            <Tooltip v-if="item.archived">
              <TooltipTrigger>
                <MdiArchive class="size-5 text-destructive" />
              </TooltipTrigger>
              <TooltipContent>
                {{ $t("global.archived") }}
              </TooltipContent>
            </Tooltip>
            <div class="grow" />
            <Tooltip>
              <TooltipTrigger>
                <Badge>
                  {{ item.quantity }}
                </Badge>
              </TooltipTrigger>
              <TooltipContent>
                {{ $t("global.quantity") }}
              </TooltipContent>
            </Tooltip>
          </div>
        </TooltipProvider>
        <Markdown class="mb-2 line-clamp-3 text-ellipsis" :source="item.description" />
        <div class="-mr-1 mt-auto flex flex-wrap justify-end gap-2">
          <LabelChip v-for="label in itemLabels" :key="label.id" :label="label" size="sm" />
        </div>
      </div>
    </NuxtLink>
  </Card>
</template>

<script setup lang="ts">
  import type { ItemOut, ItemSummary } from "~~/lib/api/types/data-contracts";
  import MdiShieldCheck from "~icons/mdi/shield-check";
  import MdiArchive from "~icons/mdi/archive";
  import { Badge } from "@/components/ui/badge";
  import { Card } from "@/components/ui/card";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
  import { Separator } from "@/components/ui/separator";
  import Markdown from "@/components/global/Markdown.vue";
  import LabelChip from "@/components/Label/Chip.vue";
  import type { Row } from "@tanstack/vue-table";
  import { Checkbox } from "@/components/ui/checkbox";

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

  const itemLabels = computed(() => {
    return props.item.labels || [];
  });

  const props = defineProps({
    item: {
      type: Object as () => ItemOut | ItemSummary,
      required: true,
    },
    locationFlatTree: {
      type: Array as () => FlatTreeItem[],
      required: false,
      default: () => [],
    },
    tableRow: {
      type: Object as () => Row<ItemSummary>,
      required: false,
      default: () => null,
    },
  });

  const locationString = computed(
    () => props.locationFlatTree.find(l => l.id === props.item.location?.id)?.treeString || props.item.location?.name
  );
</script>

<style lang="css"></style>
