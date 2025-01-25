<template>
  <NuxtLink class="group card rounded-md border border-gray-300" :to="`/item/${item.id}`">
    <div class="relative h-[200px]">
      <img
        v-if="imageUrl"
        class="h-[200px] w-full rounded-t border-gray-300 object-cover shadow-sm"
        :src="imageUrl"
        alt=""
      />
      <div class="absolute inset-x-1 bottom-1 text-wrap">
        <NuxtLink
          v-if="item.location"
          class="badge h-auto rounded-md text-sm shadow-md hover:link"
          :to="`/location/${item.location.id}`"
          loading="lazy"
        >
          {{ locationString }}
        </NuxtLink>
      </div>
    </div>
    <div class="col-span-4 flex grow flex-col gap-y-1 rounded-b bg-base-100 p-4 pt-2">
      <h2 class="line-clamp-2 text-ellipsis text-wrap text-lg font-bold">{{ item.name }}</h2>
      <div class="divider my-0"></div>
      <div class="flex gap-2">
        <div v-if="item.insured" class="tooltip z-10" data-tip="Insured">
          <MdiShieldCheck class="size-5 text-primary" />
        </div>
        <div v-if="item.archived" class="tooltip z-10" data-tip="Archived">
          <MdiArchive class="size-5 text-red-700" />
        </div>
        <div class="grow"></div>
        <div class="tooltip" data-tip="Quantity">
          <span class="badge badge-primary badge-sm size-5 text-xs">
            {{ item.quantity }}
          </span>
        </div>
      </div>
      <Markdown class="mb-2 line-clamp-3 text-ellipsis" :source="item.description" />
      <div class="-mr-1 mt-auto flex flex-wrap justify-end gap-2">
        <LabelChip v-for="label in top3" :key="label.id" :label="label" size="sm" />
      </div>
    </div>
  </NuxtLink>
</template>

<script setup lang="ts">
  import type { ItemOut, ItemSummary } from "~~/lib/api/types/data-contracts";
  import MdiShieldCheck from "~icons/mdi/shield-check";
  import MdiArchive from "~icons/mdi/archive";

  const api = useUserApi();

  const imageUrl = computed(() => {
    if (!props.item.imageId) {
      return "/no-image.jpg";
    }

    return api.authURL(`/items/${props.item.id}/attachments/${props.item.imageId}`);
  });

  const top3 = computed(() => {
    return props.item.labels.slice(0, 3) || [];
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
  });

  const locationString = computed(
    () => props.locationFlatTree.find(l => l.id === props.item.location?.id)?.treeString || props.item.location?.name
  );
</script>

<style lang="css"></style>
