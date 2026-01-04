<template>
  <Popover v-model:open="open">
    <PopoverTrigger as-child>
      <Button
        variant="outline"
        role="combobox"
        :aria-expanded="open"
        :size="sidebar.state.value === 'collapsed' ? 'icon' : undefined"
        :class="sidebar.state.value === 'collapsed' ? 'size-10' : 'w-full justify-between'"
        aria-label="Collections"
        title="Collections"
      >
        <template v-if="sidebar.state.value === 'collapsed'">
          <MdiHomeGroup class="size-5" />
        </template>
        <template v-else>
          <span class="flex items-center truncate">
            <span class="truncate">
              {{
                selectedCollection && selectedCollection.name
                  ? selectedCollection.name
                  : t("components.collection.selector.select_collection")
              }}
            </span>
          </span>

          <ChevronsUpDown class="ml-2 size-4 shrink-0 opacity-50" />
        </template>
      </Button>
    </PopoverTrigger>
    <PopoverContent
      :class="[sidebar.state.value === 'collapsed' ? 'min-w-48 p-0' : 'w-[--reka-popper-anchor-width] p-0']"
    >
      <Command :ignore-filter="true">
        <CommandGroup>
          <CommandItem
            value="create-collection"
            @select="
              () => {
                openDialog(DialogID.CreateCollection);
              }
            "
          >
            <Plus class="mr-2 size-4" /> {{ t("components.collection.selector.create_collection") }}
          </CommandItem>
          <CommandItem
            value="join-collection"
            @select="
              () => {
                openDialog(DialogID.JoinCollection);
              }
            "
          >
            <Plus class="mr-2 size-4" /> {{ t("components.collection.selector.join_collection") }}
          </CommandItem>
          <CommandItem as-child value="collection-settings">
            <NuxtLink to="/collection" class="flex w-full items-center">
              <Settings class="mr-2 size-4" />
              {{ t("components.collection.selector.collection_settings") }}
            </NuxtLink>
          </CommandItem>
        </CommandGroup>
        <CommandInput v-model="search" placeholder="Search collections..." :display-value="_ => ''" />
        <CommandEmpty>{{ t("components.collection.selector.no_collection_found") }}</CommandEmpty>
        <CommandList>
          <CommandGroup :heading="t('components.collection.selector.your_collections')">
            <CommandItem
              v-for="collection in filteredCollections"
              :key="collection.id"
              :value="collection.id"
              @select="selectCollection(collection)"
            >
              <Check
                :class="cn('mr-2 h-4 w-4', selectedCollection?.id === collection.id ? 'opacity-100' : 'opacity-0')"
              />
              <div class="flex w-full items-center justify-between gap-2">
                {{ collection.name }}
                <div class="flex items-center gap-2"></div>
              </div>
            </CommandItem>
          </CommandGroup>
        </CommandList>
      </Command>
    </PopoverContent>
  </Popover>
</template>

<script setup lang="ts">
  import { Check, ChevronsUpDown, Plus, Settings } from "lucide-vue-next";
  import MdiHomeGroup from "~icons/mdi/home-group";
  import fuzzysort from "fuzzysort";
  import { Button } from "~/components/ui/button";
  import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "~/components/ui/command";
  import { Popover, PopoverContent, PopoverTrigger } from "~/components/ui/popover";
  import { cn } from "~/lib/utils";
  import { ref, computed, watch, onMounted } from "vue";
  import { useSidebar } from "@/components/ui/sidebar/utils";
  import { useI18n } from "vue-i18n";
  import { DialogID } from "@/components/ui/dialog-provider/utils";
  import { useDialog } from "~/components/ui/dialog-provider";

  const { openDialog } = useDialog();

  const { t } = useI18n();

  const open = ref(false);
  const search = ref("");

  const { collections, selectedCollection, load, set } = useCollections();
  const collectionsList = computed(() => collections.value);

  function selectCollection(collection: CollectionSummary) {
    if (selectedCollection.value?.id !== collection.id) {
      set(collection.id);
      window.location.reload();
    }
    open.value = false;
  }

  const sidebar = useSidebar();

  const filteredCollections = computed(() => {
    const filtered = fuzzysort.go(search.value, collectionsList.value, { key: "name", all: true }).map(i => i.obj);
    return filtered;
  });

  // Reset search when value is cleared
  watch(
    () => selectedCollection.value?.id,
    () => {
      if (!selectedCollection.value) {
        search.value = "";
      }
    }
  );

  onMounted(() => {
    load();
  });
</script>
