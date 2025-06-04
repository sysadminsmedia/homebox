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
      <PopoverContent class="w-[--reka-popper-anchor-width] p-0">
        <Command :ignore-filter="true">
          <CommandInput
            v-model="search"
            :placeholder="$t('components.location.selector.search_location')"
            :display-value="_ => ''"
          />
          <CommandEmpty>{{ $t("components.location.selector.no_location_found") }}</CommandEmpty>
          <CommandList>
            <CommandGroup>
              <CommandItem
                v-for="location in filteredLocations"
                :key="location.id"
                :value="location.id"
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
  import fuzzysort from "fuzzysort";
  import { Button } from "~/components/ui/button";
  import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "~/components/ui/command";
  import { Label } from "~/components/ui/label";
  import { Popover, PopoverContent, PopoverTrigger } from "~/components/ui/popover";
  import { cn } from "~/lib/utils";
  import type { LocationSummary } from "~~/lib/api/types/data-contracts";
  import { useFlatLocations } from "~~/composables/use-location-helpers";

  type Props = {
    modelValue?: LocationSummary | null;
    currentLocationId?: string | undefined;
  };

  const props = defineProps<Props>();
  const emit = defineEmits(["update:modelValue"]);

  const open = ref(false);
  const search = ref("");
  const id = useId();
  const locations = useFlatLocations(props.currentLocationId);
  const value = useVModel(props, "modelValue", emit);

  function selectLocation(location: LocationSummary) {
    if (value.value?.id !== location.id) {
      value.value = location;
    } else {
      value.value = null;
    }
    open.value = false;
  }

  const filteredLocations = computed(() => {
    const filtered = fuzzysort.go(search.value, locations.value, { key: "name", all: true }).map(i => i.obj).filter(loc => loc.id !== props.currentLocationId);

    return filtered;
  });

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
