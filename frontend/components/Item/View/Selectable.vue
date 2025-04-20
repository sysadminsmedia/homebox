<script setup lang="ts">
  import type { ViewType } from "~~/composables/use-preferences";
  import type { ItemSummary } from "~~/lib/api/types/data-contracts";
  import MdiCardTextOutline from "~icons/mdi/card-text-outline";
  import MdiTable from "~icons/mdi/table";
  import { Button, ButtonGroup } from "@/components/ui/button";

  type Props = {
    view?: ViewType;
    items: ItemSummary[];
  };

  const preferences = useViewPreferences();

  const props = defineProps<Props>();
  const viewSet = computed(() => {
    return !!props.view;
  });

  const itemView = computed(() => {
    return props.view ?? preferences.value.itemDisplayView;
  });

  function setViewPreference(view: ViewType) {
    preferences.value.itemDisplayView = view;
  }
</script>

<template>
  <section>
    <BaseSectionHeader class="mb-2 mt-4 flex items-center justify-between">
      {{ $t("components.item.view.selectable.items") }}
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
      <ItemViewTable :items="items" />
    </template>
    <template v-else>
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3">
        <ItemCard v-for="item in items" :key="item.id" :item="item" />
        <div class="hidden first:block">{{ $t("components.item.view.selectable.no_items") }}</div>
      </div>
    </template>
  </section>
</template>

<style scoped></style>
