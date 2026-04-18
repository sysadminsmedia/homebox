<template>
  <div class="flex flex-col gap-1">
    <Label :for="id" class="px-1">
      {{ label }}
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
          <CommandList :key="commandListKey" class="max-h-[20.5rem]">
            <CommandGroup>
              <CommandItem
                v-for="item in paginatedItems"
                :key="itemKey(item)"
                :value="itemKey(item)"
                @select="select(item)"
              >
                <Check :class="cn('mr-2 h-4 w-4', isSelected(item) ? 'opacity-100' : 'opacity-0')" />
                <slot name="display" v-bind="{ item }">
                  {{ displayValue(item) }}
                </slot>
              </CommandItem>
            </CommandGroup>
          </CommandList>
          <div
            v-if="showPagination"
            class="flex items-center justify-between gap-1 border-t px-2 py-1.5"
            @pointerdown="handlePaginationToolbarPointerDown"
          >
            <div class="flex items-center gap-1">
              <Button
                type="button"
                variant="ghost"
                size="sm"
                class="size-8 p-0"
                :disabled="!canGoPrevious"
                :aria-label="t('items.first_page')"
                @click.stop.prevent="setPage(1)"
              >
                <ChevronsLeft class="size-4" />
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                class="size-8 p-0"
                :disabled="!canGoPrevious"
                :aria-label="t('items.prev_page')"
                @click.stop.prevent="setPage(currentPage - 1)"
              >
                <ChevronLeft class="size-4" />
              </Button>
            </div>

            <div class="flex items-center justify-center gap-1 text-xs text-muted-foreground">
              <Input
                v-model="pageInput"
                type="number"
                inputmode="numeric"
                :min="1"
                :max="pageCount"
                class="h-8 w-16 px-2 py-1 text-center text-xs md:text-xs"
                :aria-label="t('items.pages', { page: currentPage, totalPages: pageCount })"
                @click.stop
                @keydown.stop
                @keydown.enter.prevent="commitPageInput"
              />
              <span>/ {{ pageCount }}</span>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                class="size-8 p-0"
                :aria-label="t('global.confirm')"
                :disabled="!canConfirmPageInputState"
                @click.stop.prevent="commitPageInput"
              >
                <Check class="size-4" />
              </Button>
            </div>

            <div class="flex items-center gap-1">
              <Button
                type="button"
                variant="ghost"
                size="sm"
                class="size-8 p-0"
                :disabled="!canGoNext"
                :aria-label="t('items.next_page')"
                @click.stop.prevent="setPage(currentPage + 1)"
              >
                <ChevronRight class="size-4" />
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                class="size-8 p-0"
                :disabled="!canGoNext"
                :aria-label="t('items.last_page')"
                @click.stop.prevent="setPage(pageCount)"
              >
                <ChevronsRight class="size-4" />
              </Button>
            </div>
          </div>
        </Command>
      </PopoverContent>
    </Popover>
  </div>
</template>

<script setup lang="ts">
  import { computed, nextTick, ref, watch } from "vue";
  import { Check, ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, ChevronsUpDown, X } from "lucide-vue-next";
  import fuzzysort from "fuzzysort";
  import { useVModel } from "@vueuse/core";
  import { useI18n } from "vue-i18n";
  import { Button } from "~/components/ui/button";
  import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "~/components/ui/command";
  import { Input } from "~/components/ui/input";
  import { Label } from "~/components/ui/label";
  import { Popover, PopoverContent, PopoverTrigger } from "~/components/ui/popover";
  import { cn } from "~/lib/utils";

  const { t } = useI18n();

  type ItemsObject = {
    [key: string]: unknown;
  };

  interface Props {
    label?: string;
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
    pageSize?: number;
  }

  const emit = defineEmits(["update:modelValue", "update:search"]);
  const props = withDefaults(defineProps<Props>(), {
    label: "",
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
    pageSize: 10,
  });

  const id = useId();
  const open = ref(false);
  const search = useVModel(props, "search", emit);
  const value = useVModel(props, "modelValue", emit);
  const hasInitialSearch = ref(false);
  const currentPage = ref(1);
  const pageInput = ref<string | number>("1");

  const localizedSearchPlaceholder = computed(
    () => props.searchPlaceholder ?? t("components.item.selector.search_placeholder")
  );
  const localizedNoResultsText = computed(() => props.noResultsText ?? t("components.item.selector.no_results"));
  const localizedPlaceholder = computed(() => props.placeholder ?? t("components.item.selector.placeholder"));

  // Trigger search when popover opens for the first time if no results exist
  async function handlePopoverOpen() {
    resetPaginationState();

    if (hasInitialSearch.value || props.items.length !== 0 || !props.triggerSearch) return;

    try {
      const success = await props.triggerSearch();
      if (success) {
        // Only mark as attempted after successful completion
        hasInitialSearch.value = true;
        await nextTick();
        resetPaginationState();
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
        void handlePopoverOpen();
      }
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

  function findPageForSelectedKey(items: Array<string | ItemsObject>, selectedKey: string | null): number | null {
    if (!selectedKey) {
      return null;
    }

    const selectedIndex = items.findIndex(item => itemKey(item) === selectedKey);
    if (selectedIndex === -1) {
      return null;
    }

    return Math.floor(selectedIndex / props.pageSize) + 1;
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
    // Explicitly depend on all reactive sources
    const items = props.items;
    const searchTerm = search.value;
    const excludeItems = props.excludeItems;

    let baseItems = items;

    if (!isStrings(baseItems) && excludeItems) {
      const excludeIds = excludeItems.map(i => i.id);
      baseItems = baseItems.filter(item => !excludeIds?.includes(item.id));
    }

    if (!searchTerm) return baseItems;

    if (isStrings(baseItems)) {
      return baseItems.filter(item => item.toLowerCase().includes(searchTerm.toLowerCase()));
    } else {
      // Fuzzy search on itemText
      return fuzzysort.go(searchTerm, baseItems, { key: props.itemText, all: true }).map(i => i.obj);
    }
  });

  const pageCount = computed(() => Math.ceil(filtered.value.length / props.pageSize));
  const showPagination = computed(() => pageCount.value > 1);
  const canGoPrevious = computed(() => currentPage.value > 1);
  const canGoNext = computed(() => currentPage.value < pageCount.value);
  const canConfirmPageInputState = computed(() => {
    const page = Number(pageInput.value);

    return Number.isInteger(page) && page >= 1 && page <= pageCount.value && page !== currentPage.value;
  });
  const paginatedItems = computed(() => {
    const start = (currentPage.value - 1) * props.pageSize;
    return filtered.value.slice(start, start + props.pageSize);
  });

  function setPage(page: number) {
    if (page < 1 || page > pageCount.value) return;
    currentPage.value = page;
  }

  function commitPageInput() {
    const page = Number(pageInput.value);

    if (!Number.isInteger(page) || page < 1 || page > pageCount.value) {
      pageInput.value = String(currentPage.value);
      return;
    }

    setPage(page);
  }

  // Prevent pagination toolbar clicks from clearing the filter input
  function handlePaginationToolbarPointerDown(event: PointerEvent) {
    const toolbar = event.currentTarget as HTMLElement | null;
    const target = event.target as HTMLElement | null;
    if (target?.closest("input")) {
      return;
    }

    event.preventDefault();
    const inputElements = toolbar?.parentElement?.querySelectorAll<HTMLInputElement>("input");
    inputElements?.forEach(inputElement => {
      inputElement.blur();
    });
  }

  function resetPaginationState() {
    let selectedKey: string | null = null;
    if (value.value) {
      selectedKey = typeof value.value === "string" ? value.value : itemKey(value.value);
    }

    const initialPage = findPageForSelectedKey(filtered.value as Array<string | ItemsObject>, selectedKey) ?? 1;
    currentPage.value = initialPage;
    pageInput.value = String(initialPage);
  }

  watch(search, () => {
    currentPage.value = 1;
  });

  watch(pageCount, () => {
    if (pageCount.value === 0) {
      currentPage.value = 1;
      return;
    }

    if (currentPage.value > pageCount.value) {
      currentPage.value = pageCount.value;
    }
  });

  watch(currentPage, page => {
    pageInput.value = String(page);
  });

  // Generate a unique key to force CommandList re-render when visible items change
  const commandListKey = computed(() => {
    return JSON.stringify([currentPage.value, paginatedItems.value.map(item => itemKey(item))]);
  });
</script>
