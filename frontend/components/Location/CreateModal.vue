<template>
  <BaseModal :dialog-id="DialogID.CreateLocation" :title="$t('components.location.create_modal.title')">
    <form class="flex min-w-0 flex-col gap-2" @submit.prevent="create()">
      <LocationSelector v-model="form.parent" />

      <!-- Entity Type selector (shown when multiple location types exist) -->
      <div v-if="showEntityTypeSelector" class="flex w-full flex-col gap-1.5">
        <Label for="location-type-select" class="px-1">Type</Label>
        <select
          id="location-type-select"
          class="w-full rounded-md border bg-background px-3 py-2 text-sm"
          :value="selectedEntityType?.id || ''"
          @change="onEntityTypeChanged(($event.target as HTMLSelectElement).value)"
        >
          <option value="">Select type...</option>
          <option v-for="et in locationTypes" :key="et.id" :value="et.id">{{ et.name }}</option>
        </select>
      </div>

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

      <TagSelector v-model="form.tags" :tags="tags ?? []" />
      <PhotoUploader
        :label="$t('components.location.create_modal.location_photo')"
        :button-label="$t('components.item.create_modal.upload_photos')"
        :existing-count="form.photos.length"
        @selected="appendPhotos"
      />

      <div class="mt-4 flex flex-row-reverse">
        <ButtonGroup>
          <Button :disabled="loading" type="submit">{{ $t("global.create") }}</Button>
          <Button variant="outline" :disabled="loading" type="button" @click="create(false)">{{
            $t("global.create_and_add")
          }}</Button>
        </ButtonGroup>
      </div>

      <PhotoUploaderPreview
        :photos="form.photos"
        @delete="deletePhotoAt"
        @rotate="rotatePhotoAt"
        @set-primary="setPrimaryPhotoAt"
      />
    </form>
  </BaseModal>
</template>

<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { DialogID } from "@/components/ui/dialog-provider/utils";
  import { toast } from "@/components/ui/sonner";
  import { Button, ButtonGroup } from "~/components/ui/button";
  import { Label } from "~/components/ui/label";
  import BaseModal from "@/components/App/CreateModal.vue";
  import type { EntitySummary } from "~~/lib/api/types/data-contracts";
  import { AttachmentTypes } from "~~/lib/api/types/non-generated";
  import { useDialog, useDialogHotkey } from "~/components/ui/dialog-provider";
  import { useTagStore } from "~/stores/tags";
  import LocationSelector from "~/components/Location/Selector.vue";
  import TagSelector from "~/components/Tag/Selector.vue";
  import FormTextField from "~/components/Form/TextField.vue";
  import FormTextArea from "~/components/Form/TextArea.vue";
  import { useEntityTypeStore } from "~~/stores/entityTypes";
  import PhotoUploader from "~/components/Form/PhotoUploader.vue";
  import PhotoUploaderPreview from "~/components/Form/PhotoUploaderPreview.vue";
  import {
    deletePhoto,
    rotatePhotoPreview,
    setPrimaryPhoto,
    type PhotoPreview,
  } from "~/components/Form/photo-uploader";

  const { t } = useI18n();

  const { activeDialog, closeDialog } = useDialog();

  useDialogHotkey(DialogID.CreateLocation, { code: "Digit3", shift: true });

  const entityTypeStore = useEntityTypeStore();

  // Entity type selection
  const locationTypes = computed(() => entityTypeStore.locationTypes);
  const selectedEntityType = ref<import("~~/lib/api/types/data-contracts").EntityTypeSummary | null>(null);
  const showEntityTypeSelector = computed(() => locationTypes.value.length > 1);

  async function onEntityTypeChanged(typeId: string) {
    const et = locationTypes.value.find(t => t.id === typeId);
    selectedEntityType.value = et || null;

    // If the selected type has a default template, auto-apply it
    if (et?.defaultTemplateId && et.defaultTemplate) {
      const { data: tplData, error: tplError } = await api.templates.get(et.defaultTemplateId);
      if (!tplError && tplData) {
        if (tplData.defaultName) form.name = tplData.defaultName;
        if (tplData.defaultDescription) form.description = tplData.defaultDescription;
        if (tplData.defaultTags && tplData.defaultTags.length > 0) {
          form.tags = tplData.defaultTags.map(l => l.id);
        }
        toast.success(t("components.template.toast.applied", { name: tplData.name }));
      }
    }
  }

  const loading = ref(false);
  const focused = ref(false);
  const form = reactive({
    name: "",
    description: "",
    parent: null as EntitySummary | null,
    tags: [] as string[],
    photos: [] as PhotoPreview[],
  });

  watch(
    () => activeDialog.value,
    active => {
      if (active && active === DialogID.CreateLocation) {
        if (locationId.value) {
          const found = locations.value.find(l => l.id === locationId.value);
          form.parent = found || null;
        }
      }
    }
  );

  function reset() {
    form.name = "";
    form.description = "";
    form.tags = [];
    form.photos = [];
    focused.value = false;
    loading.value = false;
  }

  const api = useUserApi();

  const locationsStore = useLocationStore();
  const locations = computed(() => locationsStore.allLocations);

  const tagStore = useTagStore();
  const tags = computed(() => tagStore.tags);

  const route = useRoute();

  const { shift } = useMagicKeys();

  const locationId = computed(() => {
    if (route.fullPath.includes("/location/")) {
      return route.params.id;
    }
    return null;
  });

  function appendPhotos(photos: PhotoPreview[]) {
    form.photos.push(...photos);
  }

  function deletePhotoAt(index: number) {
    form.photos = deletePhoto(form.photos, index);
  }

  function setPrimaryPhotoAt(index: number) {
    form.photos = setPrimaryPhoto(form.photos, index);
  }

  async function rotatePhotoAt(index: number) {
    const photo = form.photos[index];
    if (!photo) return;

    try {
      form.photos[index] = await rotatePhotoPreview(photo);
    } catch (error) {
      toast.error(t("components.item.create_modal.toast.rotate_process_failed"));
      console.error(error);
    }
  }

  async function create(close = true) {
    if (loading.value) {
      toast.error(t("components.location.create_modal.toast.already_creating"));
      return;
    }
    loading.value = true;

    if (shift?.value) close = false;

    const { data, error } = await api.items.createLocation({
      name: form.name,
      description: form.description,
      parentId: form.parent ? form.parent.id : null,
      entityTypeId: selectedEntityType.value?.id || "",
      quantity: 1,
      tagIds: form.tags,
    });

    if (error) {
      loading.value = false;
      toast.error(t("components.location.create_modal.toast.create_failed"));
      return;
    }

    if (data) {
      toast.success(t("components.location.create_modal.toast.create_success"));
    }

    // Upload photos if any
    if (form.photos.length > 0 && data) {
      toast.info(t("components.item.create_modal.toast.uploading_photos", { count: form.photos.length }));
      let uploadError = false;
      for (const photo of form.photos) {
        const { error: attachError } = await api.items.attachments.add(
          data.id,
          photo.file,
          photo.photoName,
          AttachmentTypes.Photo,
          photo.primary
        );
        if (attachError) {
          uploadError = true;
          toast.error(t("components.item.create_modal.toast.upload_failed", { photoName: photo.photoName }));
        }
      }
      if (!uploadError) {
        toast.success(t("components.item.create_modal.toast.upload_success", { count: form.photos.length }));
      }
    }

    reset();

    if (close) {
      closeDialog(DialogID.CreateLocation);
      navigateTo(`/location/${data.id}`);
    }
  }
</script>
