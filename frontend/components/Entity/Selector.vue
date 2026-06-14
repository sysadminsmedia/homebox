<template>
  <Select :model-value="selectedEntityType" @update:model-value="id => onEntityTypeChanged(id as string)">
    <SelectTrigger :class="{ 'h-7 p-1': size === 'sm' }">
      <SelectValue :class="{ 'text-xl': size === 'sm' }" :placeholder="$t('components.entity.selector.placeholder')" />
    </SelectTrigger>
    <SelectContent>
      <SelectItem v-for="type in entityTypes" :key="type.id" :value="type.id">
        <div class="flex items-center gap-2">
          <MdiMapMarkerOutline v-if="type.isLocation" class="size-4" />
          <MdiPackageVariantClosed v-else class="size-4" />
          <span>{{ t(type.name) }}</span>
        </div>
      </SelectItem>
    </SelectContent>
  </Select>
</template>

<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import MdiMapMarkerOutline from "~icons/mdi/map-marker-outline";
  import MdiPackageVariantClosed from "~icons/mdi/package-variant-closed";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "~/components/ui/select";
  import type { EntityTypeSummary } from "~/lib/api/types/data-contracts";

  const { t } = useI18n();

  defineProps<{
    entityTypes: EntityTypeSummary[];
    selectedEntityType?: string;
    onEntityTypeChanged: (id: string) => void;
    size?: "sm" | "md";
  }>();
</script>
