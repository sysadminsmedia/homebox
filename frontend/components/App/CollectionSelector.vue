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
              {{ selectedCollection && selectedCollection.name ? selectedCollection.name : "Select collection" }}
            </span>
            <span v-if="selectedCollection?.role" class="ml-2">
              <Badge class="whitespace-nowrap" :variant="roleVariant(selectedCollection?.role)">
                {{ selectedCollection?.role }}
              </Badge>
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
          <CommandItem as-child value="collection-settings">
            <NuxtLink to="/collection" class="flex w-full items-center">
              <Settings class="mr-2 size-4" />
              Collection Settings
            </NuxtLink>
          </CommandItem>
          <CommandItem value="create-collection" @select="() => {}">
            <Plus class="mr-2 size-4" /> Create New Collection
          </CommandItem>
          <CommandItem value="join-collection" @select="() => {}">
            <Plus class="mr-2 size-4" /> Join Existing Collection
          </CommandItem>
        </CommandGroup>
        <CommandInput v-model="search" placeholder="Search collections..." :display-value="_ => ''" />
        <CommandEmpty>No inventory found</CommandEmpty>
        <CommandList>
          <CommandGroup heading="Your Collections">
            <CommandItem
              v-for="collection in filteredCollections"
              :key="collection.id"
              :value="collection.id"
              @select="selectCollection(collection)"
            >
              <Check :class="cn('mr-2 h-4 w-4', value === collection.id ? 'opacity-100' : 'opacity-0')" />
              <div class="flex w-full items-center justify-between gap-2">
                {{ collection.name }}
                <div class="flex items-center gap-2">
                  <Badge class="whitespace-nowrap" variant="outline">{{ collection.count }}</Badge>
                  <Badge class="whitespace-nowrap" :variant="roleVariant(collection.role)">{{ collection.role }}</Badge>
                </div>
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
  import { api } from "~/mock/collections";
  import type { Collection as MockCollection, User as MockUser } from "~/mock/collections";
  import { Button } from "~/components/ui/button";
  import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "~/components/ui/command";
  import { Popover, PopoverContent, PopoverTrigger } from "~/components/ui/popover";
  import { Badge } from "@/components/ui/badge";
  import { cn } from "~/lib/utils";
  import { ref, computed, watch } from "vue";
  import { useVModel } from "@vueuse/core";
  import { useSidebar } from "@/components/ui/sidebar/utils";

  // api.getCollections returns collection objects augmented with `count` and `role` for the current user
  type CollectionSummary = MockCollection & { count: number; role: MockUser["collections"][number]["role"] };

  type Props = {
    modelValue?: string | null;
  };

  const props = defineProps<Props>();
  const emit = defineEmits(["update:modelValue"]);

  const open = ref(false);
  const search = ref("");
  const value = useVModel(props, "modelValue", emit);

  // Use shared mock collections data via fake api (for current user)
  const collectionsList = ref<CollectionSummary[]>(api.getCollections() as CollectionSummary[]);

  function roleVariant(role: string | undefined) {
    if (role === "owner") return "default";
    if (role === "admin") return "secondary";
    return "outline";
  }

  function selectCollection(collection: CollectionSummary) {
    if (value.value !== collection.id) {
      value.value = collection.id;
      console.log(collection);
    }
    open.value = false;
  }

  const selectedCollection = computed(() => {
    return collectionsList.value.find(o => o.id === value.value) ?? null;
  });

  const sidebar = useSidebar();

  const filteredCollections = computed(() => {
    const filtered = fuzzysort.go(search.value, collectionsList.value, { key: "name", all: true }).map(i => i.obj);
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
