<template>
  <div class="flex flex-col gap-1">
    <Label :for="id" class="px-1">
      {{ $t("global.labels") }}
    </Label>

    <TagsInput
      v-model="modelValue"
      class="w-full gap-0 px-0"
      :display-value="v => props.labels.find(l => l.id === v)?.name ?? 'Loading...'"
    >
      <div class="flex flex-wrap items-center gap-2 px-3" style="overflow: hidden;">
        <TagsInputItem 
          v-for="item in modelValue"
          :key="item"
          :value="item"
          :style="{
            overflow: 'hidden',
            'text-wrap': 'nowrap',
          }"
        >
          <span
            v-if="props.labels.find(l => l.id === item)?.color"
            class="ml-2 inline-block size-4 rounded-full"
            :style="{ backgroundColor: props.labels.find(l => l.id === item)?.color }"
          />
          <TagsInputItemText />
          <TagsInputItemDelete />
        </TagsInputItem>
      </div>

      <ComboboxRoot v-model="modelValue" v-model:open="open" class="w-full" :ignore-filter="true">
        <ComboboxAnchor as-child>
          <ComboboxInput v-model="searchTerm" :placeholder="$t('components.label.selector.select_labels')" as-child>
            <TagsInputInput
              :id="id"
              class="w-full px-3"
              :class="modelValue.length > 0 ? 'mt-2' : ''"
              @focus="open = true"
            />
          </ComboboxInput>
        </ComboboxAnchor>

        <ComboboxPortal>
          <ComboboxContent :side-offset="4" position="popper" class="z-50">
            <CommandList
              position="popper"
              class="mt-2 w-[--reka-popper-anchor-width] rounded-md border bg-popover text-popover-foreground shadow-md outline-none data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2"
            >
              <CommandEmpty />
              <CommandGroup>
                <CommandItem
                  v-for="label in filteredLabels"
                  :key="label.value"
                  :value="label.value"
                  @select.prevent="
                    ev => {
                      if (typeof ev.detail.value === 'string') {
                        if (ev.detail.value === 'create-item') {
                          void createAndAdd(searchTerm);
                        } else {
                          if (!modelValue.includes(ev.detail.value)) {
                            modelValue = [...modelValue, ev.detail.value];
                          }
                        }
                        searchTerm = '';
                      }
                    }
                  "
                >
                  <span
                    class="mr-2 inline-block size-4 rounded-full align-middle"
                    :class="{ border: props.labels.find(l => l.id === label.value)?.color }"
                    :style="{ backgroundColor: props.labels.find(l => l.id === label.value)?.color }"
                  />
                  {{ label.label }}
                </CommandItem>
              </CommandGroup>
            </CommandList>
          </ComboboxContent>
        </ComboboxPortal>
      </ComboboxRoot>
    </TagsInput>
  </div>
</template>
<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { ComboboxAnchor, ComboboxContent, ComboboxInput, ComboboxPortal, ComboboxRoot } from "reka-ui";
  import { computed, ref } from "vue";
  import fuzzysort from "fuzzysort";
  import { toast } from "@/components/ui/sonner";
  import { CommandEmpty, CommandGroup, CommandItem, CommandList } from "~/components/ui/command";
  import {
    TagsInput,
    TagsInputInput,
    TagsInputItem,
    TagsInputItemDelete,
    TagsInputItemText,
  } from "@/components/ui/tags-input";
  import type { LabelOut } from "~/lib/api/types/data-contracts";

  const { t } = useI18n();

  const id = useId();

  const api = useUserApi();

  const emit = defineEmits(["update:modelValue"]);
  const props = defineProps({
    modelValue: {
      type: Array as () => string[],
      default: null,
    },
    labels: {
      type: Array as () => LabelOut[],
      required: true,
    },
  });

  const modelValue = useVModel(props, "modelValue", emit);

  const open = ref(false);
  const searchTerm = ref("");

  const filteredLabels = computed(() => {
    const filtered = fuzzysort
      .go(searchTerm.value, props.labels, { key: "name", all: true })
      .map(l => ({
        value: l.obj.id,
        label: l.obj.name,
      }))
      .filter(i => !modelValue.value.includes(i.value));

    // Only show "Create" option if search term is not empty and no exact match exists
    if (searchTerm.value.trim() !== "") {
      const trimmedSearchTerm = searchTerm.value.trim();
      const hasExactMatch = props.labels.some(label => label.name.toLowerCase() === trimmedSearchTerm.toLowerCase());

      if (!hasExactMatch) {
        filtered.push({ value: "create-item", label: `${t("global.create")} ${searchTerm.value}` });
      }
    }

    return filtered;
  });

  const createAndAdd = async (name: string, color = "") => {
    if (name.length > 50) {
      toast.error(t("components.label.create_modal.toast.label_name_too_long"));
      return;
    }
    const { error, data } = await api.labels.create({
      name,
      color,
      description: "",
    });

    if (error) {
      toast.error(t("components.label.create_modal.toast.create_failed"));
      return;
    }

    toast.success(t("components.label.create_modal.toast.create_success"));

    modelValue.value = [...modelValue.value, data.id];
  };

  // TODO: when reka-ui 2 is release use hook to set cursor to end when label is added with click
</script>
