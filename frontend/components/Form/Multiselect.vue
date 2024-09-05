<template>
  <div ref="menu" class="form-control w-full">
    <label class="label">
      <span class="label-text">{{ label }}</span>
    </label>
    <div class="dropdown dropdown-top sm:dropdown-end">
      <div tabindex="0" class="w-full min-h-[48px] flex gap-2 p-4 flex-wrap border border-gray-400 rounded-lg">
        <span v-for="itm in value" :key="name != '' ? itm[name] : itm" class="badge">
          {{ name != "" ? itm[name] : itm }}
        </span>
      </div>
      <div
        tabindex="0"
        style="display: inline"
        class="dropdown-content mb-1 menu w-full z-[9999] shadow border border-gray-400 rounded bg-base-100"
      >
        <div class="m-2">
          <input v-model="search" placeholder="Search…" class="input input-sm input-bordered w-full" />
        </div>
        <ul class="overflow-y-scroll max-h-60">
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
