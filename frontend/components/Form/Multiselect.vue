<template>
  <div ref="menu" class="form-control w-full">
    <label class="label">
      <span class="label-text">{{ label }}</span>
    </label>
    <div class="dropdown dropdown-top sm:dropdown-end">
      <div tabindex="0" class="flex min-h-[48px] w-full flex-wrap gap-2 rounded-lg border border-gray-400 p-4">
        <span v-for="itm in value" :key="itm.id" class="badge">
          {{ itm.name }}
        </span>
        <button
          v-if="value.length > 0"
          type="button"
          class="absolute inset-y-0 right-6 flex items-center rounded-r-md px-2 focus:outline-none"
          @click="clear"
        >
          <MdiClose class="size-5" />
        </button>
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
              bordered: selected.includes(obj.id),
            }"
          >
            <button type="button" @click="toggle(obj.id)">
              {{ obj.name }}
            </button>
          </li>
          <li v-if="!filteredItems.some(itm => itm.name === search) && search.length > 0">
            <button type="button" @click="createAndAdd(search)">{{ $t("global.create") }} {{ search }}</button>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import MdiClose from "~icons/mdi/close";

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
      return item.name.toLowerCase().includes(search.value.toLowerCase());
    });
  });

  function clear() {
    value.value = [];
  }

  const selected = computed<string[]>(() => {
    return value.value.map(itm => itm.id);
  });

  function toggle(uniqueField: string) {
    const item = props.items.find(itm => itm.id === uniqueField);
    if (selected.value.includes(item.id)) {
      value.value = value.value.filter(itm => itm.id !== item.id);
    } else {
      value.value = [...value.value, item];
    }
  }

  const api = useUserApi();
  const toast = useNotifier();

  async function createAndAdd(name: string) {
    const { error, data } = await api.labels.create({
      name,
      color: "", // Future!
      description: "",
    });

    if (error) {
      console.error(error);
      toast.error(`Failed to create label: ${name}`);
    } else {
      value.value = [...value.value, data];
    }
  }
</script>
