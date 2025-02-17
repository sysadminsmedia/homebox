<script setup lang="ts">
  import { ComboboxAnchor, ComboboxContent, ComboboxInput, ComboboxPortal, ComboboxRoot } from "radix-vue";
  import { computed, ref } from "vue";
  import { toast } from "vue-sonner";
  import { CommandEmpty, CommandGroup, CommandItem, CommandList } from "~/components/ui/command";
  import {
    TagsInput,
    TagsInputInput,
    TagsInputItem,
    TagsInputItemDelete,
    TagsInputItemText,
  } from "@/components/ui/tags-input";
  import type { LabelOut } from "~/lib/api/types/data-contracts";

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
    const filtered = props.labels
      .map(l => ({
        value: l.id,
        label: l.name,
      }))
      .filter(i => {
        return i.label.toLocaleLowerCase().includes(searchTerm.value.toLocaleLowerCase());
      })
      .filter(i => !modelValue.value.includes(i.value));

    if (searchTerm.value.trim() !== "") {
      filtered.push({ value: "create-item", label: `Create ${searchTerm.value}` });
    }

    return filtered;
  });

  const createAndAdd = async (name: string) => {
    const { error, data } = await api.labels.create({
      name,
      color: "", // Future!
      description: "",
    });

    if (error) {
      toast.error("Couldn't create label");
      return;
    }

    toast.success("Label created");
    modelValue.value.push(data.id);
  };

  // TODO: when radix-vue 2 is release use hook to set cursor to end when label is added with click
</script>

<template>
  <div class="flex flex-col gap-1">
    <Label :for="id" class="px-1">
      {{ $t("global.labels") }}
    </Label>

    <TagsInput
      v-model="modelValue"
      class="w-full gap-0 px-0"
      :display-value="v => props.labels.find(l => l.id === v)!.name"
    >
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
        :filter-function="l => l"
      >
        <ComboboxAnchor as-child>
          <ComboboxInput :placeholder="$t('components.label.selector.select_labels')" as-child>
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
              class="bg-popover text-popover-foreground data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 mt-2 w-[--radix-popper-anchor-width] rounded-md border shadow-md outline-none"
            >
              <CommandEmpty />
              <CommandGroup>
                <CommandItem
                  v-for="label in filteredLabels"
                  :key="label.value"
                  :value="label.value"
                  @select.prevent="
                    ev => {
                      console.log(ev);
                      if (typeof ev.detail.value === 'string') {
                        // TODO: this breaks everything, fix create-item
                        if (ev.detail.value === 'create-item') {
                          void createAndAdd(searchTerm);
                        } else {
                          modelValue.push(ev.detail.value);
                        }
                        searchTerm = '';
                      }
                    }
                  "
                >
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
