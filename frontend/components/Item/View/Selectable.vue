<script setup lang="ts">
  import type { ViewType } from "~~/composables/use-preferences";
  import type { ItemSummary } from "~~/lib/api/types/data-contracts";
  import MdiCardTextOutline from "~icons/mdi/card-text-outline";
  import MdiTable from "~icons/mdi/table";
  import { Badge } from "@/components/ui/badge";
  import { Button, ButtonGroup } from "@/components/ui/button";
  import BaseSectionHeader from "@/components/Base/SectionHeader.vue";
  import DataTable from "./table/data-table.vue";
  import { makeColumns } from "./table/columns";
  import { useI18n } from "vue-i18n";

  type Props = {
    view?: ViewType;
    items: ItemSummary[];
  };

  const preferences = useViewPreferences();
  const { t } = useI18n();
  const columns = computed(() => makeColumns(t));

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
      <div class="flex gap-2 text-nowrap">
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

    <DataTable :view="itemView" :columns="columns" :data="items" />
  </section>
</template>

<style scoped></style>
