<template>
  <div>
    <Combobox v-model="value">
      <ComboboxLabel class="label">
        <span class="label-text">{{ label }}</span>
      </ComboboxLabel>
      <div class="relative">
        <ComboboxInput
          :display-value="i => extractDisplay(i as SupportValues)"
          class="input input-bordered w-full"
          @change="search = $event.target.value"
        />
        <button
          v-if="!!value"
          type="button"
          class="absolute inset-y-0 right-6 flex items-center rounded-r-md px-2 focus:outline-none"
          @click="clear"
        >
          <MdiClose class="size-5" />
        </button>
        <ComboboxButton class="absolute inset-y-0 right-0 flex items-center rounded-r-md px-2 focus:outline-none">
          <MdiChevronDown class="size-5" />
        </ComboboxButton>
        <ComboboxOptions
          v-if="computedItems.length > 0"
          class="card dropdown-content absolute z-10 mt-2 max-h-60 w-full overflow-auto rounded-md border border-gray-400 bg-base-100"
        >
          <ComboboxOption
            v-for="item in computedItems"
            :key="item.id"
            v-slot="{ active, selected }"
            :value="item.value"
            as="template"
          >
            <li
              :class="[
                'relative cursor-default select-none py-2 pl-3 pr-9 transition-colors duration-75 ease-in-out',
                active ? 'bg-primary text-primary-content' : 'text-base-content',
              ]"
            >
              <slot name="display" v-bind="{ item: item, selected, active }">
                <span :class="['block truncate', selected && 'font-semibold']">
                  {{ item.display }}
                </span>
                <span
                  v-if="selected"
                  :class="[
                    'absolute inset-y-0 right-0 flex items-center pr-4 text-primary',
                    active ? 'text-primary-content' : 'bg-primary',
                  ]"
                >
                  <MdiCheck class="size-5" aria-hidden="true" />
                </span>
              </slot>
            </li>
          </ComboboxOption>
        </ComboboxOptions>
      </div>
    </Combobox>
  </div>
</template>

<script setup lang="ts">
  import lunr from "lunr";
  import {
    Combobox,
    ComboboxInput,
    ComboboxOptions,
    ComboboxOption,
    ComboboxButton,
    ComboboxLabel,
  } from "@headlessui/vue";
  import MdiClose from "~icons/mdi/close";
  import MdiChevronDown from "~icons/mdi/chevron-down";
  import MdiCheck from "~icons/mdi/check";

  type SupportValues = string | { [key: string]: any };

  type ComboItem = {
    display: string;
    value: SupportValues;
    id: number;
  };

  type Props = {
    label: string;
    modelValue: SupportValues | null | undefined;
    items: {
      id: string;
      treeString: string;
    }[];
    display?: string;
    multiple?: boolean;
  };

  const emit = defineEmits(["update:modelValue", "update:search"]);
  const props = withDefaults(defineProps<Props>(), {
    label: "",
    modelValue: "",
    display: "text",
    multiple: false,
  });

  function clear() {
    emit("update:modelValue", null);
  }

  const search = ref("");
  const value = useVModel(props, "modelValue", emit);

  function extractDisplay(item?: SupportValues): string {
    if (!item) {
      return "";
    }

    if (typeof item === "string") {
      return item;
    }

    if (props.display in item) {
      return item[props.display] as string;
    }

    // Try these options as well
    const fallback = ["name", "title", "display", "value"];
    for (let i = 0; i < fallback.length; i++) {
      const key = fallback[i];
      if (key in item) {
        return item[key] as string;
      }
    }

    return "";
  }

  function lunrFactory() {
    return lunr(function () {
      this.ref("id");
      this.field("display");

      for (let i = 0; i < props.items.length; i++) {
        const item = props.items[i];
        const display = extractDisplay(item);
        this.add({ id: i, display });
      }
    });
  }

  const index = ref<ReturnType<typeof lunrFactory>>(lunrFactory());

  watchEffect(() => {
    if (props.items) {
      index.value = lunrFactory();
    }
  });

  const computedItems = computed<ComboItem[]>(() => {
    const list: ComboItem[] = [];

    const matches = index.value.search("*" + search.value + "*");

    const resultIDs = [];
    for (let i = 0; i < matches.length; i++) {
      const match = matches[i];
      const item = props.items[parseInt(match.ref)];
      const display = extractDisplay(item);
      list.push({ id: i, display, value: item });
      resultIDs.push(item.id);
    }

    /**
     * Supplementary search,
     * Resolve the issue of language not being supported
     */
    for (let i = 0; i < props.items.length; i++) {
      const item = props.items[i];
      if (resultIDs.find(item_ => item_ === item.id) !== undefined) {
        continue;
      }
      if (item.treeString.includes(search.value)) {
        const display = extractDisplay(item);
        list.push({ id: i, display, value: item });
      }
    }

    return list;
  });
</script>
