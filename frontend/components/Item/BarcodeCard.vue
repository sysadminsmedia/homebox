<template>
  <Card class="selectable-card cursor-pointer overflow-hidden" :class="{ selected }" @click="toggleSelect">
    <div class="relative h-[200px]">
      <img
        v-if="props.item?.imageBase64"
        class="h-[200px] w-full object-cover shadow-md"
        loading="lazy"
        :src="props.item.imageBase64"
        alt=""
      />
    </div>
    <div class="col-span-4 flex grow flex-col gap-y-1 p-4 pt-2">
      <h2 class="line-clamp-2 text-ellipsis text-wrap text-lg font-bold">{{ item.item.name }}</h2>
      <Separator class="mb-1" />
      <TooltipProvider :delay-duration="0">
        <div class="flex items-center space-x-4">
          <Tooltip>
            <TooltipTrigger>
              <MdiFactory class="size-5 text-destructive" />
            </TooltipTrigger>
            <TooltipContent>
              {{ $t("items.manufacturer") }}
            </TooltipContent>
          </Tooltip>
          <span class="text-sm font-medium">
            {{ item.item.manufacturer }}
          </span>
        </div>
        <div class="flex items-center space-x-4">
          <Tooltip>
            <TooltipTrigger>
              <MdiPound class="size-5 text-destructive" />
            </TooltipTrigger>
            <TooltipContent>
              {{ $t("items.model_number") }}
            </TooltipContent>
          </Tooltip>
          <span class="text-sm font-medium">
            {{ item.item.modelNumber }}
          </span>
        </div>
        <div class="flex items-center space-x-4">
          <Tooltip>
            <TooltipTrigger>
              <MdiImageText class="size-5 text-destructive" />
            </TooltipTrigger>
            <TooltipContent>
              {{ $t("items.description") }}
            </TooltipContent>
          </Tooltip>
          <Markdown class="mb-2 line-clamp-1 text-ellipsis" :source="item.item.description" />
        </div>
      </TooltipProvider>
    </div>
  </Card>
</template>

<script setup lang="ts">
  import { defineProps, defineEmits } from "vue";
  import type { BarcodeProduct } from "~~/lib/api/types/data-contracts";
  import MdiPound from "~icons/mdi/pound";
  import MdiFactory from "~icons/mdi/factory";
  import MdiImageText from "~icons/mdi/image-text";
  import { Card } from "@/components/ui/card";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
  import { Separator } from "@/components/ui/separator";

  const props = defineProps({
    item: {
      type: Object as () => BarcodeProduct,
      required: true,
    },
    locationFlatTree: {
      type: Array as () => FlatTreeItem[],
      required: false,
      default: () => [],
    },
    modelValue: {
      type: Boolean,
      required: false,
      default: false,
    },
  });

  const selected = computed(() => props.modelValue);

  const emit = defineEmits<{
    (e: "update:modelValue", value: boolean): void;
  }>();

  function toggleSelect() {
    emit("update:modelValue", !selected.value);
  }
</script>
