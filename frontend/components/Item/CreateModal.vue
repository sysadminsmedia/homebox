<template>
  <BaseModal dialog-id="create-item" :title="$t('components.item.create_modal.title')">
    <form class="flex flex-col gap-2" @submit.prevent="create()">
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
      <LabelSelector v-model="form.labels" :labels="labels ?? []" />
      <div class="flex w-full flex-col gap-1.5">
        <Label for="image-create-photo" class="flex w-full px-1">
          {{ $t("components.item.create_modal.item_photo") }}
        </Label>
        <Input
          id="image-create-photo"
          class="w-full"
          type="file"
          accept="image/png,image/jpeg,image/gif,image/avif,image/webp;capture=camera"
          multiple
          @change="previewImage"
        />
      </div>
      <div class="mt-4 flex flex-row-reverse">
        <ButtonGroup>
          <Button :disabled="loading" type="submit" class="group">
            <div class="relative mx-2">
              <div
                class="absolute inset-0 flex items-center justify-center transition-transform duration-300 group-hover:rotate-[360deg]"
              >
                <MdiPackageVariant class="size-5 group-hover:hidden" />
                <MdiPackageVariantClosed class="hidden size-5 group-hover:block" />
              </div>
            </div>
            {{ $t("global.create") }}
          </Button>
          <Button variant="outline" :disabled="loading" type="button" @click="create(false)">
            {{ $t("global.create_and_add") }}
          </Button>
        </ButtonGroup>
      </div>

      <!-- photo preview area is AFTER the create button, to avoid pushing the button below the screen on small displays -->
      <div v-if="form.photos.length > 0" class="mt-4 border-t border-gray-300 px-4 pb-4">
        <div v-for="(photo, index) in form.photos" :key="index">
          <!-- TODO: remove the indicator, and btn classes from daisyui -->
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
  </BaseModal>
</template>

<script setup lang="ts">
  import { toast } from "vue-sonner";
  import { Button, ButtonGroup } from "~/components/ui/button";
  import BaseModal from "@/components/App/CreateModal.vue";
  import { Label } from "@/components/ui/label";
  import { Input } from "@/components/ui/input";
  import type { ItemCreate, LocationOut } from "~~/lib/api/types/data-contracts";
  import { useLabelStore } from "~~/stores/labels";
  import { useLocationStore } from "~~/stores/locations";
  import MdiPackageVariant from "~icons/mdi/package-variant";
  import MdiPackageVariantClosed from "~icons/mdi/package-variant-closed";
  import MdiDelete from "~icons/mdi/delete";
  import { AttachmentTypes } from "~~/lib/api/types/non-generated";
  import { useDialog, useDialogHotkey } from "~/components/ui/dialog-provider";
  import LabelSelector from "~/components/Label/Selector.vue";

  interface PhotoPreview {
    photoName: string;
    file: File;
    fileBase64: string;
  }

  const { activeDialog, closeDialog } = useDialog();

  useDialogHotkey("create-item", { code: "Digit1", shift: true });

  const api = useUserApi();

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

  const loading = ref(false);
  const focused = ref(false);
  const form = reactive({
    location: locations.value && locations.value.length > 0 ? locations.value[0] : ({} as LocationOut),
    name: "",
    description: "",
    color: "",
    labels: [] as string[],
    photos: [] as PhotoPreview[],
  });

  const { shift } = useMagicKeys();

  function deleteImage(index: number) {
    form.photos.splice(index, 1);
  }

  function previewImage(event: Event) {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      for (const file of input.files) {
        const reader = new FileReader();
        reader.onload = e => {
          form.photos.push({
            photoName: file.name,
            fileBase64: e.target?.result as string,
            file,
          });
        };
        reader.readAsDataURL(file);
      }
    }
  }

  watch(
    () => activeDialog.value,
    active => {
      if (active === "create-item") {
        if (locationId.value) {
          const found = locations.value.find(l => l.id === locationId.value);
          if (found) {
            form.location = found;
          }
        }
        if (labelId.value) {
          form.labels = labels.value.filter(l => l.id === labelId.value).map(l => l.id);
        }
      }
    }
  );

  async function create(close = true) {
    if (!form.location?.id) {
      toast.error("Please select a location.");
      return;
    }

    if (loading.value) {
      toast.error("Already creating an item");
      return;
    }

    loading.value = true;

    if (shift.value) close = false;

    const out: ItemCreate = {
      parentId: null,
      name: form.name,
      description: form.description,
      locationId: form.location.id as string,
      labelIds: form.labels,
    };

    const { error, data } = await api.items.create(out);

    if (error) {
      loading.value = false;
      toast.error("Couldn't create item");
      return;
    }

    toast.success("Item created");

    if (form.photos.length > 0) {
      toast.info(`Uploading ${form.photos.length} photo(s)...`);
      let uploadError = false;
      for (const photo of form.photos) {
        const { error: attachError } = await api.items.attachments.add(
          data.id,
          photo.file,
          photo.photoName,
          AttachmentTypes.Photo
        );

        if (attachError) {
          uploadError = true;
          toast.error(`Failed to upload Photo: ${photo.photoName}`);
        } else {
          toast.success(`Photo uploaded: ${photo.photoName}`);
        }
      }
      if (uploadError) {
        toast.warning("Some photos failed to upload.");
      } else {
        toast.success("All photos uploaded successfully.");
      }
    }

    form.name = "";
    form.description = "";
    form.color = "";
    form.photos = [];
    form.labels = [];
    focused.value = false;
    loading.value = false;

    if (close) {
      closeDialog("create-item");
      navigateTo(`/item/${data.id}`);
    }
  }
</script>
