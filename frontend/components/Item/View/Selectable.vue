<script setup lang="ts">
  import { useForwardPropsEmits } from "reka-ui";
  import type { TableProps, TableEmits } from "./Table.types";
  import type { ViewType } from "~~/composables/use-preferences";
  import MdiCardTextOutline from "~icons/mdi/card-text-outline";
  import MdiTable from "~icons/mdi/table";
  import { Badge } from "@/components/ui/badge";
  import { Button, ButtonGroup } from "@/components/ui/button";
  import type { BarcodeProduct, ItemSummary } from "~~/lib/api/types/data-contracts";

  type Props = {
    view?: ViewType;
  };

  const preferences = useViewPreferences();

  const props = defineProps<Props & TableProps>();

  const emits = defineEmits<TableEmits>();

  const forwardedPropsEmits = useForwardPropsEmits(props, emits);

  const selectedCard = ref<number>(-1);

  const viewSet = computed(() => {
    return !!props.view;
  });

  const itemView = computed(() => {
    return props.view ?? preferences.value.itemDisplayView;
  });

  watch(selectedCard, index => {
    if (index === -1) {
      emits("update:selectedItem", null);
      return;
    }
    emits("update:selectedItem", props.items[index]);
  });

  function setViewPreference(view: ViewType) {
    preferences.value.itemDisplayView = view;
  }
</script>

<template>
  <section>
    <BaseSectionHeader class="mb-2 mt-4 flex items-center justify-between">
      <div class="flex gap-2">
        {{ $t("components.item.view.selectable.items") }}
        <Badge>
          {{ items.length }}
        </Badge>
      </div>
      <template #description>
        <div v-if="!viewSet">
          <ButtonGroup>
            <Button size="sm" :variant="itemView === 'card' ? 'default' : 'outline'" @click="setViewPreference('card')">
              <MdiCardTextOutline class="size-5" />
              {{ $t("components.item.view.selectable.card") }}
            </Button>
            <Button
              size="sm"
              :variant="itemView === 'table' ? 'default' : 'outline'"
              @click="setViewPreference('table')"
            >
              <MdiTable class="size-5" />
              {{ $t("components.item.view.selectable.table") }}
            </Button>
          </ButtonGroup>
        </div>
      </template>
    </BaseSectionHeader>

    <template v-if="itemView === 'table'">
      <ItemViewTable v-bind="forwardedPropsEmits" />
    </template>
    <template v-else>
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3">
        <template v-if="itemType === 'itemsummary'">
          <ItemCard v-for="item in items" :key="(item as ItemSummary).id" :item="item as ItemSummary" />
        </template>
        <template v-if="itemType === 'barcodeproduct'">
          <ItemBarcodeCard
            v-for="(item, index) in items"
            :key="index"
            :item="item as BarcodeProduct"
            :model-value="selectedCard === index"
            @update:model-value="val => (selectedCard = val ? index : -1)"
          />
        </template>
        <div class="hidden first:block">{{ $t("components.item.view.selectable.no_items") }}</div>
      </div>
    </template>
  </section>
</template>

<style scoped></style>
