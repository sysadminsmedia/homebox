<template>
  <div ref="menu" class="form-control w-full">
    <label class="label">
      <span class="label-text">{{ label }}</span>
    </label>
    <div class="dropdown dropdown-top sm:dropdown-end">
      <div tabindex="0" class="flex min-h-[48px] w-full flex-wrap gap-2 rounded-lg border border-gray-400 p-4">
        <span v-for="itm in value" :key="name != '' ? itm[name] : itm" class="badge">
          {{ name != "" ? itm[name] : itm }}
        </span>
      </div>
      <div
        tabindex="0"
        style="display: inline"
        class="dropdown-content menu z-[9999] mb-1 w-full rounded border border-gray-400 bg-base-100 shadow"
      >
        <div class="m-2">
          <input v-model="search" placeholder="Search…" class="input input-bordered input-sm w-full" />
        </div>
        <ul class="max-h-60 overflow-y-scroll">
          <li
            v-for="(obj, idx) in filteredItems"
            :key="idx"
            :class="{
              bordered: selected.includes(obj[props.uniqueField]),
            }"
          >
            <button type="button" @click="toggle(obj[props.uniqueField])">
              {{ name != "" ? obj[name] : obj }}
            </button>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  const emit = defineEmits(["update:modelValue"]);
  const props = defineProps({
    label: {
      type: String,
      default: "",
    },
    modelValue: {
      type: Array as () => any[],
      default: null,
    },
    items: {
      type: Array as () => any[],
      required: true,
    },
    name: {
      type: String,
      default: "name",
    },
    uniqueField: {
      type: String,
      default: "id",
    },
    selectFirst: {
      type: Boolean,
      default: false,
    },
  });

  const value = useVModel(props, "modelValue", emit);

  const search = ref("");

  const filteredItems = computed(() => {
    if (!search.value) {
      return props.items;
    }

    return props.items.filter(item => {
      return item[props.name].toLowerCase().includes(search.value.toLowerCase());
    });
  });

  const selected = computed<string[]>(() => {
    return value.value.map(itm => itm[props.uniqueField]);
  });

  function toggle(uniqueField: string) {
    const item = props.items.find(itm => itm[props.uniqueField] === uniqueField);
    if (selected.value.includes(item[props.uniqueField])) {
      value.value = value.value.filter(itm => itm[props.uniqueField] !== item[props.uniqueField]);
    } else {
      value.value = [...value.value, item];
    }
  }
</script>
