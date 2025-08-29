<template>
  <Popover>
    <PopoverTrigger as-child>
      <Button size="sm" variant="outline" class="group/filter">
        {{ label }} {{ len }}
        <MdiChevronDown class="transition-transform group-data-[state=open]/filter:rotate-180" />
      </Button>
    </PopoverTrigger>
    <PopoverContent class="z-40 p-0">
      <div class="p-4 shadow-sm">
        <Input v-model="search" type="text" placeholder="Search…" />
      </div>
      <div class="max-h-72 divide-y overflow-y-auto">
        <Label
          v-for="v in selectedView"
          :key="v.id"
          class="flex cursor-pointer justify-between px-4 py-2 text-sm hover:bg-accent hover:text-accent-foreground"
        >
          <div>
            <span>{{ v.name }}</span>
            <span v-if="v.treeString && v.treeString !== v.name" class="ml-auto text-xs">{{ v.treeString }}</span>
          </div>
          <Checkbox :model-value="true" @update:model-value="_ => (selected = selected.filter(s => s.id !== v.id))" />
        </Label>
        <hr v-if="selected.length > 0" />
        <Label
          v-for="v in unselected"
          :key="v.id"
          class="flex cursor-pointer justify-between px-4 py-2 text-sm hover:bg-accent hover:text-accent-foreground"
        >
          <div>
            <div>{{ v.name }}</div>
            <div v-if="v.treeString && v.treeString !== v.name" class="ml-auto text-xs">
              {{ v.treeString }}
            </div>
          </div>
          <Checkbox :model-value="false" @update:model-value="_ => (selected = [...selected, v])" />
        </Label>
      </div>
    </PopoverContent>
  </Popover>
</template>

<script setup lang="ts">
  import MdiChevronDown from "~icons/mdi/chevron-down";
  import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
  import { Button } from "@/components/ui/button";
  import { Checkbox } from "@/components/ui/checkbox";
  import { Input } from "@/components/ui/input";
  import { Label } from "@/components/ui/label";

  type Props = {
    label?: string;
    options: {
      name: string;
      id: string;
      treeString?: string;
    }[];
    modelValue?: {
      name: string;
      id: string;
      treeString?: string;
    }[];
  };

  const search = ref("");
  const searchFold = computed(() => search.value.toLowerCase());

  const emit = defineEmits(["update:modelValue"]);
  const props = withDefaults(defineProps<Props>(), {
    label: "",
    modelValue: () => [],
  });

  const len = computed(() => {
    return selected.value.length > 0 ? `(${selected.value.length})` : "";
  });

  const selectedView = computed(() => {
    return selected.value.filter(o => {
      if (searchFold.value.length > 0) {
        return o.name.toLowerCase().includes(searchFold.value);
      }
      return true;
    });
  });

  const selected = useVModel(props, "modelValue", emit);

  const unselected = computed(() => {
    return props.options.filter(o => {
      if (searchFold.value.length > 0) {
        return o.name.toLowerCase().includes(searchFold.value) && selected.value.every(s => s.id !== o.id);
      }
      return selected.value.every(s => s.id !== o.id);
    });
  });
</script>
