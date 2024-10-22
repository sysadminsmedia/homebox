<template>
  <FormAutocomplete2 v-if="locations" v-model="value" :items="locations" display="name" :label="$t('components.location.selector.parent_location')">
    <template #display="{ item, selected, active }">
      <div>
        <div class="flex w-full">
          {{ cast(item.value).name }}
          <span
            v-if="selected"
            :class="['absolute inset-y-0 right-0 flex  items-center pr-4', active ? 'text-white' : 'text-primary']"
          >
            <MdiCheck class="size-5" aria-hidden="true" />
          </span>
        </div>
        <div v-if="cast(item.value).name != cast(item.value).treeString" class="mt-1 text-xs">
          {{ cast(item.value).treeString }}
        </div>
      </div>
    </template>
  </FormAutocomplete2>
</template>

<script lang="ts" setup>
  import type { FlatTreeItem } from "~~/composables/use-location-helpers";
  import { useFlatLocations } from "~~/composables/use-location-helpers";
  import type { LocationSummary } from "~~/lib/api/types/data-contracts";
  import MdiCheck from "~icons/mdi/check";

  type Props = {
    modelValue?: LocationSummary | null;
  };

  // Cast the type of the item to a FlatTreeItem so we can get type "safety" in the template
  // Note that this does not actually change the type of the item, it just tells the compiler
  // that the type is FlatTreeItem. We must keep this in sync with the type of the items
  function cast(value: any): FlatTreeItem {
    return value as FlatTreeItem;
  }

  const props = defineProps<Props>();
  const value = useVModel(props, "modelValue");

  const locations = useFlatLocations();
  const form = ref({
    parent: null as LocationSummary | null,
    search: "",
  });

  // Whenever parent goes from value to null reset search
  watch(
    () => value.value,
    () => {
      if (!value.value) {
        form.value.search = "";
      }
    }
  );
</script>
