<template>
  <BaseModal v-model="modal">
    <template #title> {{ $t("components.item.create_modal.title") }} </template>
    <form @submit.prevent="create()">
      <LocationSelector v-model="form.location" />
      <FormTextField
        ref="nameInput"
        v-model="form.name"
        :trigger-focus="focused"
        :autofocus="true"
        label="Item Name"
        :max-length="255"
        :min-length="1"
      />
      <FormTextArea v-model="form.description" label="Item Description" :max-length="1000" />
      <FormMultiselect v-model="form.labels" label="Labels" :items="labels ?? []" />

      <div class="modal-action mb-6">
        <div>
          <label for="photo" class="btn">{{ $t("components.item.create_modal.photo_button") }}</label>
          <input
            id="photo"
            class="hidden"
            type="file"
            accept="image/png,image/jpeg,image/gif,image/avif,image/webp"
            @change="previewImage"
          />
        </div>
        <div class="grow"></div>
        <div>
          <BaseButton class="rounded-r-none" :loading="loading" type="submit">
            <template #icon>
              <MdiPackageVariant class="swap-off size-5" />
              <MdiPackageVariantClosed class="swap-on size-5" />
            </template>
            {{ $t("global.create") }}
          </BaseButton>
          <div class="dropdown dropdown-top">
            <label tabindex="0" class="btn rounded-l-none rounded-r-xl">
              <MdiChevronDown class="size-5" name="mdi-chevron-down" />
            </label>
            <ul tabindex="0" class="dropdown-content menu rounded-box right-0 w-64 bg-base-100 p-2 shadow">
              <li>
                <button type="button" @click="create(false)">{{ $t("global.create_and_add") }}</button>
              </li>
            </ul>
          </div>
        </div>
      </div>

      <!-- photo preview area is AFTER the create button, to avoid pushing the button below the screen on small displays -->
      <div class="border-t border-gray-300 p-4">
        <template v-if="form.preview">
          <p class="mb-0">File name: {{ form.photo?.name }}</p>
          <img
            :src="form.preview"
            class="h-[100px] w-full rounded-t border-gray-300 object-cover shadow-sm"
            alt="Uploaded Photo"
          />
        </template>
      </div>
    </form>
    <p class="mt-4 text-center text-sm">
      use <kbd class="kbd kbd-xs">Shift</kbd> + <kbd class="kbd kbd-xs"> Enter </kbd> to create and add another
    </p>
  </BaseModal>
</template>

<script setup lang="ts">
  import type { ItemCreate, LabelOut, LocationOut } from "~~/lib/api/types/data-contracts";
  import { useLabelStore } from "~~/stores/labels";
  import { useLocationStore } from "~~/stores/locations";
  import MdiPackageVariant from "~icons/mdi/package-variant";
  import MdiPackageVariantClosed from "~icons/mdi/package-variant-closed";
  import MdiChevronDown from "~icons/mdi/chevron-down";
  import { AttachmentTypes } from "~~/lib/api/types/non-generated";

  const props = defineProps({
    modelValue: {
      type: Boolean,
      required: true,
    },
  });

  const api = useUserApi();
  const toast = useNotifier();

  const locationsStore = useLocationStore();
  const locations = computed(() => locationsStore.allLocations);

  const labelStore = useLabelStore();
  const labels = computed(() => labelStore.labels);

  const route = useRoute();

  const labelId = computed(() => {
    if (route.fullPath.includes("/label/")) {
      return route.params.id;
    }
    return null;
  });

  const locationId = computed(() => {
    if (route.fullPath.includes("/location/")) {
      return route.params.id;
    }
    return null;
  });

  const nameInput = ref<HTMLInputElement | null>(null);

  const modal = useVModel(props, "modelValue");
  const loading = ref(false);
  const focused = ref(false);
  const form = reactive({
    location: locations.value && locations.value.length > 0 ? locations.value[0] : ({} as LocationOut),
    name: "",
    description: "",
    color: "", // Future!
    labels: [] as LabelOut[],
    preview: null as string | null,
    photo: null as File | null,
  });

  const { shift } = useMagicKeys();

  function previewImage(event: Event) {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      const reader = new FileReader();
      reader.onload = e => {
        form.preview = e.target?.result as string;
      };
      const file = input.files[0];
      form.photo = file;
      reader.readAsDataURL(file);
    }
  }

  whenever(
    () => modal.value,
    () => {
      focused.value = true;

      if (locationId.value) {
        const found = locations.value.find(l => l.id === locationId.value);
        if (found) {
          form.location = found;
        }
      }

      if (labelId.value) {
        form.labels = labels.value.filter(l => l.id === labelId.value);
      }
    }
  );

  async function create(close = true) {
    if (!form.location) {
      return;
    }

    if (loading.value) {
      toast.error("Already creating an item");
      return;
    }

    loading.value = true;

    if (shift.value) {
      close = false;
    }

    const out: ItemCreate = {
      parentId: null,
      name: form.name,
      description: form.description,
      locationId: form.location.id as string,
      labelIds: form.labels.map(l => l.id) as string[],
    };

    const { error, data } = await api.items.create(out);
    loading.value = false;
    if (error) {
      loading.value = false;
      toast.error("Couldn't create item");
      return;
    }

    toast.success("Item created");

    // if the photo was provided, upload it
    if (form.photo) {
      const { error } = await api.items.attachments.add(data.id, form.photo, form.photo.name, AttachmentTypes.Photo);

      if (error) {
        loading.value = false;
        toast.error("Failed to upload Photo");
        return;
      }

      toast.success("Photo uploaded");
    }

    // Reset
    form.name = "";
    form.description = "";
    form.color = "";
    form.preview = null;
    form.photo = null;
    focused.value = false;
    loading.value = false;

    if (close) {
      modal.value = false;
      navigateTo(`/item/${data.id}`);
    }
  }
</script>
