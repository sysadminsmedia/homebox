<template>
  <div ref="menu" class="form-control w-full">
    <label class="label">
      <span class="label-text">{{ label }}</span>
    </label>
    <div class="dropdown dropdown-top sm:dropdown-end">
      <div class="relative">
        <input
          v-model="internalSearch"
          tabindex="0"
          class="input flex w-full flex-wrap items-center rounded-lg border border-gray-400"
          @keyup.enter="selectFirst"
        />
        <button
          v-if="!!modelValue && Object.keys(modelValue).length !== 0"
          style="transform: translateY(-50%)"
          class="btn btn-circle btn-xs no-animation absolute right-2 top-1/2"
          @click="clear"
        >
          x
        </button>
      </div>
      <ul
        tabindex="0"
        style="display: inline"
        class="dropdown-content menu z-[9999] mb-1 max-h-60 w-full overflow-y-scroll rounded border border-gray-400 bg-base-100 shadow"
      >
        <li v-for="(obj, idx) in filtered" :key="idx">
          <div type="button" @click="select(obj)">
            <slot name="display" v-bind="{ item: obj }">
              {{ extractor(obj, itemText) }}
            </slot>
          </div>
        </li>
        <li class="hidden first:flex">
          <button disabled>
            {{ noResultsText }}
          </button>
        </li>
      </ul>
    </div>
  </div>
</template>

<script lang="ts" setup>
  type ItemsObject = {
    text?: string;
    value?: string;
    [key: string]: unknown;
  };

  interface Props {
    label: string;
    modelValue: string | ItemsObject | null | undefined;
    items: ItemsObject[] | string[];
    itemText?: keyof ItemsObject;
    itemSearch?: keyof ItemsObject | null;
    itemValue?: keyof ItemsObject;
    search?: string;
    noResultsText?: string;
  }

  const emit = defineEmits(["update:modelValue", "update:search"]);
  const props = withDefaults(defineProps<Props>(), {
    label: "",
    modelValue: "",
    items: () => [],
    itemText: "text",
    search: "",
    itemSearch: null,
    itemValue: "value",
    noResultsText: "No Results Found",
  });

  const searchKey = computed(() => props.itemSearch || props.itemText);

  function clear() {
    select(value.value);
  }

  const internalSearch = ref("");

  watch(
    () => props.search,
    val => {
      internalSearch.value = val;
    }
  );

  watch(
    () => internalSearch.value,
    val => {
      emit("update:search", val);
    }
  );

  function extractor(obj: string | ItemsObject, key: string | number): string {
    if (typeof obj === "string") {
      return obj;
    }

    return obj[key] as string;
  }

  const value = useVModel(props, "modelValue", emit);

  const usingObjects = computed(() => {
    return props.items.length > 0 && typeof props.items[0] === "object";
  });

  /**
   * isStrings is a type guard function to check if the items are an array of string
   */
  function isStrings(_arr: string[] | ItemsObject[]): _arr is string[] {
    return !usingObjects.value;
  }

  function selectFirst() {
    if (filtered.value.length > 0) {
      select(filtered.value[0]);
    }
  }

  watch(
    value,
    () => {
      if (value.value) {
        if (typeof value.value === "string") {
          internalSearch.value = value.value;
        } else {
          internalSearch.value = value.value[searchKey.value] as string;
        }
      }
    },
    {
      immediate: true,
    }
  );

  function select(obj: Props["modelValue"]) {
    if (isStrings(props.items)) {
      if (obj === value.value) {
        value.value = "";
        return;
      }
      // @ts-ignore
      value.value = obj;
    } else {
      if (obj === value.value) {
        value.value = {};
        return;
      }

      // @ts-ignore
      value.value = obj;
    }
  }

  const filtered = computed(() => {
    if (!internalSearch.value || internalSearch.value === "") {
      return props.items;
    }

    if (isStrings(props.items)) {
      return props.items.filter(item => item.toLowerCase().includes(internalSearch.value.toLowerCase()));
    } else {
      return props.items.filter(item => {
        if (searchKey.value && searchKey.value in item) {
          return (item[searchKey.value] as string).toLowerCase().includes(internalSearch.value.toLowerCase());
        }
        return false;
      });
    }
  });
</script>
