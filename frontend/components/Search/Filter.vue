<template>
  <div ref="el" class="dropdown" :class="{ 'dropdown-open': dropdownOpen }">
    <button ref="btn" tabindex="0" class="btn btn-xs" @click="toggle">
      {{ label }} {{ len }} <MdiChevronDown class="h-4 w-4" />
    </button>
    <div tabindex="0" class="dropdown-content mt-1 w-64 shadow bg-base-100 rounded-md">
      <div class="pt-4 px-4 shadow-sm mb-1">
        <input v-model="search" type="text" placeholder="Search…" class="input input-sm input-bordered w-full mb-2" />
      </div>
      <div class="overflow-y-auto max-h-72 divide-y">
        <label
          v-for="v in selectedView"
          :key="v"
          class="cursor-pointer px-4 label flex justify-between hover:bg-base-200"
        >
          <span class="label-text mr-2">
            <slot name="display" v-bind="{ item: v }">
              {{ v[display] }}
            </slot>
          </span>
          <input v-model="selected" type="checkbox" :value="v" class="checkbox checkbox-sm checkbox-primary" />
        </label>
        <hr v-if="selected.length > 0" />
        <label
          v-for="v in unselected"
          :key="v"
          class="cursor-pointer px-4 label flex justify-between hover:bg-base-200"
        >
          <span class="label-text mr-2">
            <slot name="display" v-bind="{ item: v }">
              {{ v[display] }}
            </slot>
          </span>
          <input v-model="selected" type="checkbox" :value="v" class="checkbox checkbox-sm checkbox-primary" />
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
        return o[props.display].toLowerCase().includes(searchFold.value) && selected.value.every(s => s[props.uniqueField] !== o[props.uniqueField]);
      }
      return selected.value.every(s => s[props.uniqueField] !== o[props.uniqueField]);
    });
  });
</script>

<style scoped></style>
