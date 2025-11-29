<template>
  <div class="flex flex-col gap-1">
    <Label :for="id" class="px-1"> Template (Optional) </Label>

    <Popover v-model:open="open">
      <PopoverTrigger as-child>
        <Button :id="id" variant="outline" role="combobox" :aria-expanded="open" class="w-full justify-between">
          {{ value && value.name ? value.name : "Select template..." }}
          <ChevronsUpDown class="ml-2 size-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent class="w-[--reka-popper-anchor-width] p-0">
        <Command :ignore-filter="true">
          <CommandInput v-model="search" placeholder="Search templates..." :display-value="_ => ''" />
          <CommandEmpty>No template found</CommandEmpty>
          <CommandList>
            <CommandGroup>
              <CommandItem
                v-for="template in filteredTemplates"
                :key="template.id"
                :value="template.id"
                @select="selectTemplate(template)"
              >
                <Check :class="cn('mr-2 h-4 w-4', value?.id === template.id ? 'opacity-100' : 'opacity-0')" />
                <div class="flex w-full flex-col">
                  <div>{{ template.name }}</div>
                  <div v-if="template.description" class="mt-1 line-clamp-1 text-xs text-muted-foreground">
                    {{ template.description }}
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
  import type { ItemTemplateSummary } from "~~/lib/api/types/data-contracts";

  type Props = {
    modelValue?: ItemTemplateSummary | null;
  };

  const props = defineProps<Props>();
  const emit = defineEmits(["update:modelValue", "template-selected"]);

  const open = ref(false);
  const search = ref("");
  const id = useId();
  const value = useVModel(props, "modelValue", emit);

  const api = useUserApi();

  const { data: templates } = useAsyncData("templates-selector", async () => {
    const { data, error } = await api.templates.getAll();
    if (error) {
      return [];
    }
    return data;
  });

  function selectTemplate(template: ItemTemplateSummary) {
    if (value.value?.id !== template.id) {
      value.value = template;
      emit("template-selected", template);
    } else {
      value.value = null;
      emit("template-selected", null);
    }
    open.value = false;
  }

  const filteredTemplates = computed(() => {
    if (!templates.value) return [];
    const filtered = fuzzysort.go(search.value, templates.value, { key: "name", all: true }).map(i => i.obj);
    return filtered;
  });

  watch(
    () => value.value,
    () => {
      if (!value.value) {
        search.value = "";
      }
    }
  );
</script>
