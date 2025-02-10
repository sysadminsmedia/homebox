<script setup lang="ts">
  import { ComboboxAnchor, ComboboxContent, ComboboxInput, ComboboxPortal, ComboboxRoot } from "radix-vue";
  import { computed, ref } from "vue";
  import { CommandEmpty, CommandGroup, CommandItem, CommandList } from "~/components/ui/command";
  import {
    TagsInput,
    TagsInputInput,
    TagsInputItem,
    TagsInputItemDelete,
    TagsInputItemText,
  } from "@/components/ui/tags-input";

  const frameworks = [
    { value: "next.js", label: "Next.js" },
    { value: "sveltekit", label: "SvelteKit" },
    { value: "nuxt", label: "Nuxt" },
    { value: "remix", label: "Remix" },
    { value: "astro", label: "Astro" },
  ];

  const modelValue = ref<string[]>([]);
  const open = ref(false);
  const searchTerm = ref("");

  const filteredFrameworks = computed(() => {
    const filtered = frameworks.filter(i => {
      return i.label.toLocaleLowerCase().includes(searchTerm.value.toLocaleLowerCase());
    });

    if (searchTerm.value.trim() !== "") {
      filtered.push({ value: "create-item", label: `Create ${searchTerm.value}` });
    }

    return filtered;
  });
  const filterFunction = (list: string[]) => {
    return list;
  };

  // TODO: when radix-vue 2 is release use hook to set cursor to end when label is added with click
  // TODO: get rid of no item found text, replace framework... placeholder
</script>

<template>
  <TagsInput v-model="modelValue" class="w-80 gap-0 px-0">
    <div class="flex flex-wrap items-center gap-2 px-3">
      <TagsInputItem v-for="item in modelValue" :key="item" :value="item">
        <TagsInputItemText />
        <TagsInputItemDelete />
      </TagsInputItem>
    </div>

    <ComboboxRoot
      v-model="modelValue"
      v-model:open="open"
      v-model:search-term="searchTerm"
      class="w-full"
      :filter-function="filterFunction"
    >
      <ComboboxAnchor as-child>
        <ComboboxInput placeholder="Framework..." as-child>
          <TagsInputInput class="w-full px-3" :class="modelValue.length > 0 ? 'mt-2' : ''" @focus="open = true" />
        </ComboboxInput>
      </ComboboxAnchor>

      <ComboboxPortal>
        <ComboboxContent>
          <CommandList
            position="popper"
            class="mt-2 w-[--radix-popper-anchor-width] rounded-md border bg-popover text-popover-foreground shadow-md outline-none data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2"
          >
            <CommandEmpty />
            <CommandGroup>
              <CommandItem
                v-for="framework in filteredFrameworks"
                :key="framework.value"
                :value="framework.value"
                @select.prevent="
                  ev => {
                    if (typeof ev.detail.value === 'string') {
                      searchTerm = '';
                      modelValue.push(ev.detail.value);
                    }
                  }
                "
              >
                {{ framework.label }}
              </CommandItem>
            </CommandGroup>
          </CommandList>
        </ComboboxContent>
      </ComboboxPortal>
    </ComboboxRoot>
  </TagsInput>
</template>
