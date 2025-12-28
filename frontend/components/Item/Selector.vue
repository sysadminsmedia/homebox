<template>
  <div class="flex flex-col gap-1">
    <Label :for="id" class="px-1">
      {{ tag }}
    </Label>
    <Popover v-model:open="open">
      <PopoverTrigger as-child>
        <Button :id="id" variant="outline" role="combobox" :aria-expanded="open" class="w-full justify-between">
          <span class="truncate text-left">
            <slot name="display" v-bind="{ item: value }">
              {{ displayValue(value) || localizedPlaceholder }}
            </slot>
          </span>

          <span class="ml-2 flex items-center">
            <button
              v-if="value"
              type="button"
              class="shrink-0 rounded p-1 hover:bg-primary/20"
              :aria-label="t('components.item.selector.clear')"
              @click.stop.prevent="clearSelection"
            >
              <X class="size-4" />
            </button>

            <ChevronsUpDown class="ml-2 size-4 shrink-0 opacity-50" />
          </span>
        </Button>
      </PopoverTrigger>
      <PopoverContent class="w-[--reka-popper-anchor-width] p-0">
        <Command :ignore-filter="true">
          <CommandInput v-model="search" :placeholder="localizedSearchPlaceholder" :display-value="_ => ''" />
          <CommandEmpty>
            <div v-if="isLoading" class="flex items-center justify-center p-4">
              <div class="size-4 animate-spin rounded-full border-2 border-primary border-t-transparent"></div>
              <span class="ml-2">{{ t("components.item.selector.searching") }}</span>
            </div>
            <div v-else>
              {{ localizedNoResultsText }}
            </div>
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
  import { computed, ref, watch } from "vue";
  import { Check, ChevronsUpDown, X } from "lucide-vue-next";
  import fuzzysort from "fuzzysort";
  import { useVModel } from "@vueuse/core";
  import { useI18n } from "vue-i18n";
  import { Button } from "~/components/ui/button";
  import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "~/components/ui/command";
  import { Label } from "~/components/ui/tag";
  import { Popover, PopoverContent, PopoverTrigger } from "~/components/ui/popover";
  import { cn } from "~/lib/utils";

  const { t } = useI18n();

  type ItemsObject = {
    [key: string]: unknown;
  };

  interface Props {
    tag?: string;
    modelValue?: string | ItemsObject | null | undefined;
    items?: ItemsObject[] | string[];
    itemText?: string;
    itemValue?: string;
    search?: string;
    searchPlaceholder?: string;
    noResultsText?: string;
    placeholder?: string;
    excludeItems?: ItemsObject[];
    isLoading?: boolean;
    triggerSearch?: () => Promise<boolean>;
  }

  const emit = defineEmits(["update:modelValue", "update:search"]);
  const props = withDefaults(defineProps<Props>(), {
    tag: "",
    modelValue: "",
    items: () => [],
    itemText: "text",
    itemValue: "value",
    search: "",
    searchPlaceholder: undefined,
    noResultsText: undefined,
    placeholder: undefined,
    excludeItems: undefined,
    isLoading: false,
    triggerSearch: undefined,
  });

  const id = useId();
  const open = ref(false);
  const search = ref(props.search);
  const value = useVModel(props, "modelValue", emit);
  const hasInitialSearch = ref(false);

  const localizedSearchPlaceholder = computed(
    () => props.searchPlaceholder ?? t("components.item.selector.search_placeholder")
  );
  const localizedNoResultsText = computed(() => props.noResultsText ?? t("components.item.selector.no_results"));
  const localizedPlaceholder = computed(() => props.placeholder ?? t("components.item.selector.placeholder"));

  // Trigger search when popover opens for the first time if no results exist
  async function handlePopoverOpen() {
    if (hasInitialSearch.value || props.items.length !== 0 || !props.triggerSearch) return;

    try {
      const success = await props.triggerSearch();
      if (success) {
        // Only mark as attempted after successful completion
        hasInitialSearch.value = true;
      }
      // If not successful, leave hasInitialSearch false to allow retries
    } catch (err) {
      console.error("triggerSearch failed:", err);
      // Leave hasInitialSearch false to allow retries on subsequent opens
    }
  }

  watch(
    () => open.value,
    isOpen => {
      if (isOpen) {
        handlePopoverOpen();
      }
    }
  );

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
    return arr.length > 0 && typeof arr[0] === "string";
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

  function clearSelection() {
    value.value = null;
    search.value = "";
    open.value = false;
  }

  const filtered = computed(() => {
    let baseItems = props.items;

    if (!isStrings(baseItems) && props.excludeItems) {
      const excludeIds = props.excludeItems.map(i => i.id);
      baseItems = baseItems.filter(item => !excludeIds?.includes(item.id));
    }
    if (!search.value) return baseItems;

    if (isStrings(baseItems)) {
      return baseItems.filter(item => item.toLowerCase().includes(search.value.toLowerCase()));
    } else {
      // Fuzzy search on itemText
      return fuzzysort.go(search.value, baseItems, { key: props.itemText, all: true }).map(i => i.obj);
    }
  });
</script>
