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

      <!-- Photo upload -->
      <div class="flex w-full flex-col gap-1.5">
        <Label for="location-create-photo" class="flex w-full px-1">
          {{ $t("components.item.create_modal.item_photo") }}
        </Label>
        <div class="relative inline-block">
          <Button type="button" variant="outline" class="w-full" aria-hidden="true" @click.prevent="">
            {{ $t("components.item.create_modal.upload_photos") }}
          </Button>
          <Input
            id="location-create-photo"
            ref="fileInput"
            class="absolute left-0 top-0 size-full cursor-pointer opacity-0"
            type="file"
            accept="image/png,image/jpeg,image/gif,image/avif,image/webp,android/force-camera-workaround"
            multiple
            @change="previewImage"
          />
        </div>
      </div>

      <!-- Expanded fields (collapsible) -->
      <button
        type="button"
        class="mt-1 flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
        @click="showAdvanced = !showAdvanced"
      >
        <span v-if="!showAdvanced">{{ $t("global.show_more") }}</span>
        <span v-else>{{ $t("global.show_less") }}</span>
        <MdiChevronDown class="size-4 transition-transform" :class="{ 'rotate-180': showAdvanced }" />
      </button>

      <template v-if="showAdvanced">
        <TagSelector v-model="form.tags" :tags="tags ?? []" />
        <FormTextArea v-model="form.notes" label="Notes" :max-length="1000" />
      </template>

      <div class="mt-4 flex flex-row-reverse">
        <ButtonGroup>
          <Button :disabled="loading" type="submit">{{ $t("global.create") }}</Button>
          <Button variant="outline" :disabled="loading" type="button" @click="create(false)">{{
            $t("global.create_and_add")
          }}</Button>
        </ButtonGroup>
      </div>

      <!-- Photo preview (after buttons, like item create modal) -->
      <div v-if="form.photos.length > 0" class="mt-4 border-t px-4 pb-4">
        <div v-for="(photo, index) in form.photos" :key="index">
          <div class="mt-8 w-full">
            <img
              :src="photo.fileBase64"
              class="w-full rounded object-fill shadow-sm"
              :alt="$t('components.item.create_modal.uploaded')"
            />
          </div>
          <div class="mt-2 flex items-center gap-2">
            <TooltipProvider class="flex gap-2" :delay-duration="0">
              <Tooltip>
                <TooltipTrigger>
                  <Button size="icon" type="button" variant="destructive" @click.prevent="deleteImage(index)">
                    <MdiDelete />
                    <div class="sr-only">{{ $t("components.item.create_modal.delete_photo") }}</div>
                  </Button>
                </TooltipTrigger>
                <TooltipContent>
                  <p>{{ $t("components.item.create_modal.delete_photo") }}</p>
                </TooltipContent>
              </Tooltip>
              <Tooltip>
                <TooltipTrigger>
                  <Button
                    size="icon"
                    type="button"
                    :variant="photo.primary ? 'default' : 'outline'"
                    @click.prevent="setPrimary(index)"
                  >
                    <MdiStar v-if="photo.primary" />
                    <MdiStarOutline v-else />
                    <div class="sr-only">
                      {{ $t("components.item.create_modal.set_as_primary_photo", { isPrimary: photo.primary }) }}
                    </div>
                  </Button>
                </TooltipTrigger>
                <TooltipContent>
                  <p>{{ $t("components.item.create_modal.set_as_primary_photo", { isPrimary: photo.primary }) }}</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
            <p class="mt-1 text-sm" style="overflow-wrap: anywhere">{{ photo.photoName }}</p>
          </div>
        </div>
      </div>
    </form>
  </BaseModal>
</template>

<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { DialogID } from "@/components/ui/dialog-provider/utils";
  import { toast } from "@/components/ui/sonner";
  import { Button, ButtonGroup } from "~/components/ui/button";
  import { Label } from "@/components/ui/label";
  import { Input } from "@/components/ui/input";
  import BaseModal from "@/components/App/CreateModal.vue";
  import type { EntityTypeSummary, EntitySummary } from "~~/lib/api/types/data-contracts";
  import { AttachmentTypes } from "~~/lib/api/types/non-generated";
  import { useDialog, useDialogHotkey } from "~/components/ui/dialog-provider";
  import { useTagStore } from "~/stores/tags";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "~/components/ui/tooltip";
  import LocationSelector from "~/components/Location/Selector.vue";
  import TagSelector from "~/components/Tag/Selector.vue";
  import FormTextField from "~/components/Form/TextField.vue";
  import FormTextArea from "~/components/Form/TextArea.vue";
  import MdiChevronDown from "~icons/mdi/chevron-down";
  import MdiDelete from "~icons/mdi/delete";
  import MdiStar from "~icons/mdi/star";
  import MdiStarOutline from "~icons/mdi/star-outline";

  interface PhotoPreview {
    photoName: string;
    file: File;
    fileBase64: string;
    primary: boolean;
  }

  const { t } = useI18n();

  const { activeDialog, closeDialog } = useDialog();

  useDialogHotkey(DialogID.CreateLocation, { code: "Digit3", shift: true });

  // Entity type selection
  const locationTypes = ref<import("~~/lib/api/types/data-contracts").EntityTypeSummary[]>([]);
  const selectedEntityType = ref<import("~~/lib/api/types/data-contracts").EntityTypeSummary | null>(null);
  const showEntityTypeSelector = computed(() => locationTypes.value.length > 1);

  onMounted(async () => {
    const { data, error } = await api.entityTypes.getAll();
    if (!error && data) {
      locationTypes.value = data.filter(et => et.isLocation);
      if (locationTypes.value.length === 1) {
        selectedEntityType.value = locationTypes.value[0]!;
      }
    }
  });

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
  const showAdvanced = ref(false);
  const form = reactive({
    name: "",
    description: "",
    parent: null as EntitySummary | null,
    tags: [] as string[],
    notes: "",
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
    form.notes = "";
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
            primary: form.photos.length === 0,
          });
        };
        reader.readAsDataURL(file);
      }
      input.value = "";
    }
  }

  function deleteImage(index: number) {
    form.photos.splice(index, 1);
  }

  function setPrimary(index: number) {
    const primary = form.photos.findIndex(p => p.primary);
    if (primary !== -1) form.photos[primary]!.primary = false;
    if (primary !== index) form.photos[index]!.primary = true;
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
      entityTypeId: selectedEntityType.value?.id,
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
