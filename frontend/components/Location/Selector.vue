<!-- <template>
  <FormAutocomplete2
    v-if="locations"
    v-model="value"
    :items="locations"
    display="name"
    :label="$t('components.location.selector.parent_location')"
  >
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
</script> -->

<template>
  <div class="flex flex-col gap-1">
    <Label :for="id" class="px-1">
      {{ $t("components.location.selector.parent_location") }}
    </Label>

    <Popover v-model:open="open">
      <PopoverTrigger as-child>
        <Button :id="id" variant="outline" role="combobox" :aria-expanded="open" class="w-full justify-between">
          {{ value && value.name ? value.name : $t("components.location.selector.select_location") }}
          <ChevronsUpDown class="ml-2 size-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent class="w-[--radix-popper-anchor-width] p-0">
        <Command v-model="search">
          <CommandInput :placeholder="$t('components.location.selector.search_location')" />
          <CommandEmpty>{{ $t("components.location.selector.no_location_found") }}</CommandEmpty>
          <CommandList>
            <CommandGroup>
              <CommandItem
                v-for="location in locations"
                :key="location.id"
                :value="location.name"
                @select="selectLocation(location as unknown as LocationSummary)"
              >
                <Check :class="cn('mr-2 h-4 w-4', value?.id === location.id ? 'opacity-100' : 'opacity-0')" />
                <div>
                  <div class="flex w-full">
                    {{ location.name }}
                  </div>
                  <div v-if="location.name !== location.treeString" class="mt-1 text-xs text-muted-foreground">
                    {{ location.treeString }}
                  </div>
                </div>
              </CommandItem>
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  </div>
</template>

<script setup lang="ts">
  import { Check, ChevronsUpDown } from "lucide-vue-next";
  import { Button } from "~/components/ui/button";
  import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "~/components/ui/command";
  import { Label } from "~/components/ui/label";
  import { Popover, PopoverContent, PopoverTrigger } from "~/components/ui/popover";
  import { cn } from "~/lib/utils";
  import type { LocationSummary } from "~~/lib/api/types/data-contracts";
  import { useFlatLocations } from "~~/composables/use-location-helpers";

  type Props = {
    modelValue?: LocationSummary | null;
  };

  const props = defineProps<Props>();
  const emit = defineEmits(["update:modelValue"]);

  const open = ref(false);
  const search = ref("");
  const id = useId();
  const locations = useFlatLocations();
  const value = useVModel(props, "modelValue", emit);

  function selectLocation(location: LocationSummary) {
    if (value.value?.id !== location.id) {
      value.value = location;
    } else {
      value.value = null;
    }
    open.value = false;
  }

  // Reset search when value is cleared
  watch(
    () => value.value,
    () => {
      if (!value.value) {
        search.value = "";
      }
    }
  );
</script>
