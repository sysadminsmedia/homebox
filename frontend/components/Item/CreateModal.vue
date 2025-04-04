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
        :label="$t('components.item.create_modal.item_name')"
        :max-length="255"
        :min-length="1"
      />
      <FormTextArea
        v-model="form.description"
        :label="$t('components.item.create_modal.item_description')"
        :max-length="1000"
      />
      <FormMultiselect v-model="form.labels" :label="$t('global.labels')" :items="labels ?? []" />

      <div class="modal-action mb-6">
        <div>
          <label for="photo" class="btn">{{ $t("components.item.create_modal.photo_button") }}</label>
          <input
            id="photo"
            class="hidden"
            type="file"
            accept="image/png,image/jpeg,image/gif,image/avif,image/webp"
            multiple
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
      <div class="border-t border-gray-300 px-4 pb-4">
        <div v-for="(photo, index) in form.photos" :key="index">
          <div class="indicator mt-8 w-auto">
            <div class="indicator-item right-2 top-2">
              <button type="button" class="btn btn-circle btn-primary btn-md" @click="deleteImage(index)">
                <MdiDelete class="size-5" />
              </button>
            </div>

            <img
              :src="photo.fileBase64"
              class="w-full rounded-t border-gray-300 object-fill shadow-sm"
              alt="Uploaded Photo"
            />
          </div>
          <p class="mt-1 text-sm" style="overflow-wrap: anywhere">File name: {{ photo.photoName }}</p>
        </div>
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
  import MdiDelete from "~icons/mdi/delete";
  import { AttachmentTypes } from "~~/lib/api/types/non-generated";

  interface PhotoPreview {
    photoName: string;
    file: File;
    fileBase64: string;
  }

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
    photos: [] as PhotoPreview[],
  });

  const { shift } = useMagicKeys();

  function deleteImage(index: number) {
    form.photos.splice(index, 1);
  }

  function previewImage(event: Event) {
    const input = event.target as HTMLInputElement;

    // We support uploading multiple files at once, so build up the list of files to preview and upload
    if (input.files && input.files.length > 0) {
      for (const file of input.files) {
        const reader = new FileReader();
        reader.onload = e => {
          form.photos.push({ photoName: file.name, fileBase64: e.target?.result as string, file });
        };

        reader.readAsDataURL(file);
      }
    }
  }

  watch(
    () => modal.value,
    open => {
      if (open) {
        useTimeoutFn(() => {
          focused.value = true;
        }, 50);

        if (locationId.value) {
          const found = locations.value.find(l => l.id === locationId.value);
          if (found) {
            form.location = found;
          }
        }

        if (labelId.value) {
          form.labels = labels.value.filter(l => l.id === labelId.value);
        }
      } else {
        focused.value = false;
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

    // If the photo was provided, upload it
    // NOTE: This is not transactional. It's entirely possible for some of the photos to successfully upload and the rest to fail, which will result in missing photos
    for (const photo of form.photos) {
      const { error } = await api.items.attachments.add(data.id, photo.file, photo.photoName, AttachmentTypes.Photo);

      if (error) {
        loading.value = false;
        toast.error("Failed to upload Photo " + photo.photoName);
        return;
      }

      toast.success("Photo uploaded");
    }

    // Reset
    form.name = "";
    form.description = "";
    form.color = "";
    form.photos = [];
    focused.value = false;
    loading.value = false;

    if (close) {
      modal.value = false;
      navigateTo(`/item/${data.id}`);
    }
  }
</script>
