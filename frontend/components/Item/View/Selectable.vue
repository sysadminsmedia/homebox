<script setup lang="ts">
  import type { ViewType } from "~~/composables/use-preferences";
  import type { ItemSummary } from "~~/lib/api/types/data-contracts";
  import MdiCardTextOutline from "~icons/mdi/card-text-outline";
  import MdiTable from "~icons/mdi/table";
  import { Badge } from "@/components/ui/badge";
  import { Button, ButtonGroup } from "@/components/ui/button";
  import BaseSectionHeader from "@/components/Base/SectionHeader.vue";
  import ItemCard from "@/components/Item/Card.vue";
  import ItemViewTable from "@/components/Item/View/Table.vue";

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

  const emit = defineEmits(["refreshItems"]);

  const selectedAllCards = ref<boolean | "indeterminate">(false);

  function setViewPreference(view: ViewType) {
    preferences.value.itemDisplayView = view;
  }
</script>

<template>
  <section>
    <BaseSectionHeader class="mb-2 mt-4 flex items-center justify-between">
      <div class="flex items-center gap-2 text-nowrap">
        {{ $t("components.item.view.selectable.items") }}
        <Badge>
          {{ items.length }}
        </Badge>
        <div v-if="itemView === 'card'" class="flex items-center gap-2">
          <Checkbox id="selectAll" v-model="selectedAllCards" class="size-6" />
          <Label for="selectAll" class="cursor-pointer"> Select all </Label>
        </div>
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
      <ItemViewTable :items="items" />
    </template>
    <template v-else>
      <div v-if="items.length === 0" class="flex flex-col items-center gap-2">
        <p>{{ $t("items.no_results") }}</p>
      </div>
      <ItemViewCardGrid v-else v-model="selectedAllCards" :items="items" @refresh-items="emit('refreshItems')" />
    </template>
  </section>
</template>

<style scoped></style>
