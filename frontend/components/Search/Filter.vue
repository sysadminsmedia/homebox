<template>
  <div ref="el" class="dropdown" :class="{ 'dropdown-open': dropdownOpen }">
    <button ref="btn" tabindex="0" class="btn btn-xs" @click="toggle">
      {{ label }} {{ len }} <MdiChevronDown class="size-4" />
    </button>
    <div tabindex="0" class="dropdown-content mt-1 w-64 rounded-md bg-base-100 shadow">
      <div class="mb-1 px-4 pt-4 shadow-sm">
        <input v-model="search" type="text" placeholder="Search…" class="input input-bordered input-sm mb-2 w-full" />
      </div>
      <div class="max-h-72 divide-y overflow-y-auto">
        <label
          v-for="v in selectedView"
          :key="v"
          class="label flex cursor-pointer justify-between px-4 hover:bg-base-200"
        >
          <span class="label-text mr-2">
            <slot name="display" v-bind="{ item: v }">
              {{ v[display] }}
            </slot>
          </span>
          <input v-model="selected" type="checkbox" :value="v" class="checkbox checkbox-primary checkbox-sm" />
        </label>
        <hr v-if="selected.length > 0" />
        <label
          v-for="v in unselected"
          :key="v"
          class="label flex cursor-pointer justify-between px-4 hover:bg-base-200"
        >
          <span class="label-text mr-2">
            <slot name="display" v-bind="{ item: v }">
              {{ v[display] }}
            </slot>
          </span>
          <input v-model="selected" type="checkbox" :value="v" class="checkbox checkbox-primary checkbox-sm" />
        </label>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import MdiChevronDown from "~icons/mdi/chevron-down";
  type Props = {
    label: string;
    options: any[];
    display?: string;
    modelValue: any[];
    uniqueField: string;
  };

  const btn = ref<HTMLButtonElement>();

  const search = ref("");
  const searchFold = computed(() => search.value.toLowerCase());
  const dropdownOpen = ref(false);
  const el = ref();

  function toggle() {
    dropdownOpen.value = !dropdownOpen.value;

    if (!dropdownOpen.value) {
      btn.value?.blur();
    }
  }

  onClickOutside(el, () => {
    dropdownOpen.value = false;
  });

  watch(dropdownOpen, val => {
    console.log(val);
  });

  const emit = defineEmits(["update:modelValue"]);
  const props = withDefaults(defineProps<Props>(), {
    label: "",
    display: "name",
    modelValue: () => [],
    uniqueField: "id",
  });

  const len = computed(() => {
    return selected.value.length > 0 ? `(${selected.value.length})` : "";
  });

  const selectedView = computed(() => {
    return selected.value.filter(o => {
      if (searchFold.value.length > 0) {
        return o[props.display].toLowerCase().includes(searchFold.value);
      }
      return true;
    });
  });

  const selected = useVModel(props, "modelValue", emit);

  const unselected = computed(() => {
    return props.options.filter(o => {
      if (searchFold.value.length > 0) {
        return (
          o[props.display].toLowerCase().includes(searchFold.value) &&
          selected.value.every(s => s[props.uniqueField] !== o[props.uniqueField])
        );
      }
      return selected.value.every(s => s[props.uniqueField] !== o[props.uniqueField]);
    });
  });
</script>

<style scoped></style>
