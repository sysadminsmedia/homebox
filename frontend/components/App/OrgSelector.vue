<template>
  <Popover v-model:open="open">
    <PopoverTrigger as-child>
      <Button
        variant="outline"
        role="combobox"
        :aria-expanded="open"
        class="w-full justify-between"
      >
        {{ value && value.name ? value.name : "Select inventory" }}
                        <div class="flex items-center gap-2" v-if="value">
                  <Badge
                    class="whitespace-nowrap"
                    :variant="value.role === 'owner' ? 'default' : value.role === 'admin' ? 'secondary' : 'outline'"
                  >
                    {{ value.role }}
                  </Badge>
                </div>

        <ChevronsUpDown class="ml-2 size-4 shrink-0 opacity-50" />
      </Button>
    </PopoverTrigger>
    <PopoverContent class="w-[--reka-popper-anchor-width] p-0">
      <Command :ignore-filter="true">
        <CommandInput v-model="search" placeholder="Search collections..." :display-value="(_) => ''" />
        <CommandEmpty>No inventory found</CommandEmpty>
        <CommandList>
          <CommandGroup heading="Your Collections">
            <CommandItem
              v-for="org in filteredOrgs"
              :key="org.id"
              :value="org.id"
              @select="selectOrg(org as unknown as OrgSummary)"
            >
              <Check :class="cn('mr-2 h-4 w-4', value?.id === org.id ? 'opacity-100' : 'opacity-0')" />
              <div class="flex w-full items-center justify-between gap-2">
                {{ org.name }}
                <div class="flex items-center gap-2">
                  <Badge
                    class="whitespace-nowrap"
                    variant="outline"
                  >
                    {{ org.count }}
                  </Badge>
                  <Badge
                    class="whitespace-nowrap"
                    :variant="org.role === 'owner' ? 'default' : org.role === 'admin' ? 'secondary' : 'outline'"
                  >
                    {{ org.role }}
                  </Badge>
                </div>
              </div>
            </CommandItem>
          </CommandGroup>
          <CommandGroup>
            <CommandItem @select="() => {}">
              <Plus class="mr-2 size-4" /> Create New Collection
            </CommandItem>
            <CommandItem @select="() => {}">
              <Plus class="mr-2 size-4" /> Join Existing Collection
            </CommandItem>
          </CommandGroup>
        </CommandList>
      </Command>
    </PopoverContent>
  </Popover>
</template>

<script setup lang="ts">
  import { Check, ChevronsUpDown, Lock, Users, Plus } from "lucide-vue-next";
  import fuzzysort from "fuzzysort";
  import { Button } from "~/components/ui/button";
  import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "~/components/ui/command";
  import { Popover, PopoverContent, PopoverTrigger } from "~/components/ui/popover";
  import { Badge } from "@/components/ui/badge";
  import { cn } from "~/lib/utils";
  import { ref, computed, watch } from "vue";
  import { useVModel } from "@vueuse/core";

  type OrgSummary = {
    id: string;
    name: string;
    count: number;
    role: "owner" | "admin" | "editor" | "viewer";
    type: "personal" | "org";
  };

  type Props = {
    modelValue?: OrgSummary | null;
  };

  const props = defineProps<Props>();
  const emit = defineEmits(["update:modelValue"]);

  const open = ref(false);
  const search = ref("");
  const value = useVModel(props, "modelValue", emit);

  // Mock data for demonstration purposes
  const orgs = ref<OrgSummary[]>([
    { id: "1", name: "Personal Inventory", count: 1, role: "owner", type: "personal" },
    { id: "2", name: "Family Home", count: 4, role: "admin", type: "org" },
    { id: "3", name: "Office Equipment", count: 12, role: "editor", type: "org" },
    { id: "4", name: "Workshop Tools", count: 3, role: "viewer", type: "org" },
  ]);

  function selectOrg(org: OrgSummary) {
    if (value.value?.id !== org.id) {
      value.value = org;
    } else {
      value.value = null;
    }
    open.value = false;
  }

  const filteredOrgs = computed(() => {
    const filtered = fuzzysort.go(search.value, orgs.value, { key: "name", all: true }).map((i) => i.obj);
    return filtered;
  });

  // Reset search when value is cleared
  watch(
    () => value.value,
    () => {
      if (!value.value) {
        search.value = "";
      }
    },
  );
</script>