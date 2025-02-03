<template>
  <Dialog v-model:open="modal">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>{{ $t("components.location.create_modal.title") }}</DialogTitle>
      </DialogHeader>
      <form @submit.prevent="create()">
        <LocationSelector v-model="form.parent" />
        <FormTextField
          ref="locationNameRef"
          v-model="form.name"
          :trigger-focus="focused"
          :autofocus="true"
          :required="true"
          :label="$t('components.location.create_modal.location_name')"
          :max-length="255"
          :min-length="1"
        />
        <FormTextArea
          v-model="form.description"
          :label="$t('components.location.create_modal.location_description')"
          :max-length="1000"
        />
        <div class="modal-action">
          <div class="flex justify-center">
            <BaseButton class="rounded-r-none" type="submit" :loading="loading">{{ $t("global.create") }}</BaseButton>
            <div class="dropdown dropdown-top">
              <label tabindex="0" class="btn rounded-l-none rounded-r-xl">
                <MdiChevronDown class="size-5" />
              </label>
              <ul tabindex="0" class="dropdown-content menu rounded-box right-0 w-64 bg-base-100 p-2 shadow">
                <li>
                  <button type="button" @click="create(false)">{{ $t("global.create_and_add") }}</button>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </form>

      <DialogFooter>
        use <kbd class="kbd kbd-xs">Shift</kbd> + <kbd class="kbd kbd-xs"> Enter </kbd> to create and add another
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
  import { toast } from "vue-sonner";
  import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import type { LocationSummary } from "~~/lib/api/types/data-contracts";
  import MdiChevronDown from "~icons/mdi/chevron-down";
  const props = defineProps({
    modelValue: {
      type: Boolean,
      required: true,
    },
  });

  const modal = useVModel(props, "modelValue");
  const loading = ref(false);
  const focused = ref(false);
  const form = reactive({
    name: "",
    description: "",
    parent: null as LocationSummary | null,
  });

  watch(
    () => modal.value,
    open => {
      if (open) {
        // useTimeoutFn(() => {
        //   focused.value = true;
        // }, 50);

        if (locationId.value) {
          const found = locations.value.find(l => l.id === locationId.value);
          if (found) {
            form.parent = found;
          }
        }
      } else {
        // focused.value = false;
      }
    }
  );

  function reset() {
    form.name = "";
    form.description = "";
    form.parent = null;
    focused.value = false;
    loading.value = false;
  }

  const api = useUserApi();

  const locationsStore = useLocationStore();
  const locations = computed(() => locationsStore.allLocations);

  const route = useRoute();

  const { shift } = useMagicKeys();

  const locationId = computed(() => {
    if (route.fullPath.includes("/location/")) {
      return route.params.id;
    }
    return null;
  });

  async function create(close = true) {
    if (loading.value) {
      toast.error("Already creating a location");
      return;
    }
    loading.value = true;

    if (shift.value) {
      close = false;
    }

    const { data, error } = await api.locations.create({
      name: form.name,
      description: form.description,
      parentId: form.parent ? form.parent.id : null,
    });

    if (error) {
      loading.value = false;
      toast.error("Couldn't create location");
    }

    if (data) {
      toast.success("Location created");
    }
    reset();

    if (close) {
      modal.value = false;
      navigateTo(`/location/${data.id}`);
    }
  }
</script>
