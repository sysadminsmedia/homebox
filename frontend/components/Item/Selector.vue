<template>
  <div class="flex flex-col gap-1">
    <Label :for="id" class="px-1">
      {{ label }}
    </Label>
    <Popover v-model:open="open">
      <PopoverTrigger as-child>
        <Button :id="id" variant="outline" role="combobox" :aria-expanded="open" class="w-full justify-between">
          <span>
            <slot name="display" v-bind="{ item: value }">
              {{ displayValue(value) || localizedPlaceholder }}
            </slot>
          </span>
          <ChevronsUpDown class="ml-2 size-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent class="w-[--reka-popper-anchor-width] p-0">
        <Command :ignore-filter="true">
          <CommandInput v-model="search" :placeholder="localizedSearchPlaceholder" :display-value="_ => ''" />
          <CommandEmpty>
            {{ localizedNoResultsText }}
          </CommandEmpty>
          <CommandList>
            <CommandGroup>
              <CommandItem v-for="item in filtered" :key="itemKey(item)" :value="itemKey(item)" @select="select(item)">
                <Check :class="cn('mr-2 h-4 w-4', isSelected(item) ? 'opacity-100' : 'opacity-0')" />
                <slot name="display" v-bind="{ item }">
                  {{ displayValue(item) }}
                </slot>
              </CommandItem>
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, watch } from "vue";
  import { Check, ChevronsUpDown } from "lucide-vue-next";
  import fuzzysort from "fuzzysort";
  import { useVModel } from "@vueuse/core";
  import { useI18n } from "vue-i18n";
  import { Button } from "~/components/ui/button";
  import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "~/components/ui/command";
  import { Label } from "~/components/ui/label";
  import { Popover, PopoverContent, PopoverTrigger } from "~/components/ui/popover";
  import { cn } from "~/lib/utils";
  import { useId } from "#imports";

  const { t } = useI18n();

  type ItemsObject = {
    [key: string]: unknown;
  };

  interface Props {
    label: string;
    modelValue: string | ItemsObject | null | undefined;
    items: ItemsObject[] | string[];
    itemText?: string;
    itemValue?: string;
    search?: string;
    searchPlaceholder?: string;
    noResultsText?: string;
    placeholder?: string;
  }

  const emit = defineEmits(["update:modelValue", "update:search"]);
  const props = withDefaults(defineProps<Props>(), {
    label: "",
    modelValue: "",
    items: () => [],
    itemText: "text",
    itemValue: "value",
    search: "",
    searchPlaceholder: "Type to search...",
    noResultsText: "No Results Found",
    placeholder: "Select...",
  });

  const id = useId();
  const open = ref(false);
  const search = ref(props.search);
  const value = useVModel(props, "modelValue", emit);

  const localizedSearchPlaceholder = computed(
    () => props.searchPlaceholder ?? t("components.item.selector.search_placeholder")
  );
  const localizedNoResultsText = computed(() => props.noResultsText ?? t("components.item.selector.no_results"));
  const localizedPlaceholder = computed(() => props.placeholder ?? t("components.item.selector.placeholder"));

  watch(
    () => props.search,
    val => {
      search.value = val;
    }
  );

  watch(
    () => search.value,
    val => {
      emit("update:search", val);
    }
  );

  function isStrings(arr: string[] | ItemsObject[]): arr is string[] {
    return typeof arr[0] === "string";
  }

  function displayValue(item: string | ItemsObject | null | undefined): string {
    if (!item) return "";
    if (typeof item === "string") return item;
    return (item[props.itemText] as string) || "";
  }

  function itemKey(item: string | ItemsObject): string {
    if (typeof item === "string") return item;
    return (item[props.itemValue] as string) || displayValue(item);
  }

  function isSelected(item: string | ItemsObject): boolean {
    if (!value.value) return false;
    if (typeof item === "string") return value.value === item;
    if (typeof value.value === "string") return itemKey(item) === value.value;
    return itemKey(item) === itemKey(value.value);
  }

  function select(item: string | ItemsObject) {
    if (isSelected(item)) {
      value.value = null;
    } else {
      value.value = item;
    }
    open.value = false;
  }

  const filtered = computed(() => {
    if (!search.value) return props.items;
    if (isStrings(props.items)) {
      return props.items.filter(item => item.toLowerCase().includes(search.value.toLowerCase()));
    } else {
      // Fuzzy search on itemText
      return fuzzysort.go(search.value, props.items, { key: props.itemText, all: true }).map(i => i.obj);
    }
  });
</script>
