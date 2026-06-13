<template>
  <BaseModal :dialog-id="DialogID.CreateEntity">
    <template #title>
      <div class="flex items-center gap-2 text-nowrap">
        <span>Create</span>
        <EntitySelector
          :selected-entity-type="selectedEntityType?.id"
          :entity-types="subItemCreate ? entityTypes.filter(t => !t.isLocation) : entityTypes"
          size="sm"
          @entity-type-changed="onEntityTypeChanged"
        />
      </div>
    </template>
    <template #header-actions>
      <div class="flex gap-2">
        <TooltipProvider :delay-duration="0">
          <!-- Template selector button -->
          <Tooltip v-if="!selectedEntityType?.isLocation">
            <TooltipTrigger>
              <TemplateSelector v-model="selectedTemplate" compact @template-selected="handleTemplateSelected" />
            </TooltipTrigger>
            <TooltipContent>
              <p>{{ $t("components.template.apply_template") }}</p>
            </TooltipContent>
          </Tooltip>

          <ButtonGroup>
            <Tooltip>
              <TooltipTrigger>
                <Button variant="outline" :disabled="loading" size="icon" data-pos="start" @click="openQrScannerPage()">
                  <MdiBarcodeScan class="size-5" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>{{ $t("components.entity.create_modal.product_tooltip_scan_barcode") }}</p>
              </TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger>
                <Button variant="outline" :disabled="loading" size="icon" data-pos="end" @click="openBarcodeDialog()">
                  <MdiBarcode class="size-5" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>{{ $t("components.entity.create_modal.product_tooltip_input_barcode") }}</p>
              </TooltipContent>
            </Tooltip>
          </ButtonGroup>
        </TooltipProvider>
      </div>
    </template>

    <form class="flex min-w-0 flex-col gap-2" @submit.prevent="create()">
      <LocationSelector v-model="form.location" />

      <!-- Template Info Display - Collapsible banner with distinct styling -->
      <div v-if="templateData" class="rounded-lg border-l-4 border-l-primary bg-primary/5 p-3">
        <div class="flex items-start justify-between gap-2">
          <div class="flex flex-1 items-start gap-2">
            <MdiFileDocumentOutline class="mt-0.5 size-4 shrink-0 text-primary" />
            <div class="flex-1">
              <h4 class="text-sm font-medium text-foreground">
                {{ $t("components.template.using_template", { name: templateData.name }) }}
              </h4>
              <button
                type="button"
                class="mt-1 flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
                @click="showTemplateDetails = !showTemplateDetails"
              >
                <span v-if="!showTemplateDetails">{{ $t("components.template.show_defaults") }}</span>
                <span v-else>{{ $t("components.template.hide_defaults") }}</span>
                <MdiChevronDown class="size-4 transition-transform" :class="{ 'rotate-180': showTemplateDetails }" />
              </button>
            </div>
          </div>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            class="size-7 shrink-0"
            :aria-label="$t('components.entity.create_modal.clear_template')"
            @click="clearTemplate"
          >
            <MdiClose class="size-4" />
          </Button>
        </div>

        <!-- Collapsible details section -->
        <div v-if="showTemplateDetails" class="mt-3 border-t border-primary/20 pt-3">
          <div class="flex flex-col gap-2 text-xs text-muted-foreground">
            <p v-if="templateData.description" class="text-foreground/80">{{ templateData.description }}</p>
            <div class="grid grid-cols-2 gap-x-4 gap-y-1">
              <div v-if="templateData.defaultName">
                <span class="font-medium">{{ $t("global.name") }}:</span> {{ templateData.defaultName }}
              </div>
              <div>
                <span class="font-medium">{{ $t("global.quantity") }}:</span> {{ templateData.defaultQuantity }}
              </div>
              <div>
                <span class="font-medium">{{ $t("global.insured") }}:</span>
                {{ templateData.defaultInsured ? $t("global.yes") : $t("global.no") }}
              </div>
              <div v-if="templateData.defaultManufacturer">
                <span class="font-medium">{{ $t("components.template.form.manufacturer") }}:</span>
                {{ templateData.defaultManufacturer }}
              </div>
              <div v-if="templateData.defaultModelNumber">
                <span class="font-medium">{{ $t("components.template.form.model_number") }}:</span>
                {{ templateData.defaultModelNumber }}
              </div>
              <div v-if="templateData.defaultLifetimeWarranty">
                <span class="font-medium">{{ $t("components.template.form.lifetime_warranty") }}:</span>
                {{ $t("global.yes") }}
              </div>
              <div v-if="templateData.defaultLocation">
                <span class="font-medium">{{ $t("components.template.form.location") }}:</span>
                {{ templateData.defaultLocation.name }}
              </div>
            </div>
            <div v-if="templateData.defaultTags && templateData.defaultTags.length > 0" class="mt-1">
              <span class="font-medium">{{ $t("global.tags") }}:</span>
              {{ templateData.defaultTags.map((t: any) => t.name).join(", ") }}
            </div>
            <div v-if="templateData.defaultDescription" class="mt-1">
              <p class="font-medium">{{ $t("components.template.form.item_description") }}:</p>
              <p class="ml-2">{{ templateData.defaultDescription }}</p>
            </div>
            <div v-if="templateData.fields && templateData.fields.length > 0" class="mt-1">
              <p class="font-medium">{{ $t("components.template.form.custom_fields") }}:</p>
              <ul class="ml-4 flex list-none flex-col gap-1">
                <li v-for="field in templateData.fields" :key="field.id">
                  <span class="font-medium">{{ field.name }}:</span>
                  <span> {{ field.textValue || $t("components.template.empty_value") }}</span>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      <ItemSelector
        v-if="subItemCreate"
        v-model="parent"
        v-model:search="query"
        :label="$t('components.entity.create_modal.parent_item')"
        :items="results"
        item-text="name"
        :no-results-text="$t('components.entity.create_modal.item_selector_no_results_text')"
        :is-loading="isLoading"
        :trigger-search="triggerSearch"
      />
      <FormTextField
        ref="nameInput"
        v-model="form.name"
        :trigger-focus="focused"
        :autofocus="true"
        :label="
          $t('components.entity.create_modal.entity_name', {
            type: selectedEntityType ? t(selectedEntityType.name) : '',
          })
        "
        :max-length="255"
        :min-length="1"
      />
      <FormTextField
        v-if="!selectedEntityType?.isLocation"
        v-model.number="form.quantity"
        :label="
          $t('components.entity.create_modal.entity_quantity', {
            type: t(selectedEntityType ? selectedEntityType.name : 'global.entity'),
          })
        "
        type="number"
        step="any"
      />
      <FormTextArea
        v-model="form.description"
        :label="
          $t('components.entity.create_modal.entity_description', {
            type: t(selectedEntityType ? selectedEntityType.name : 'global.entity'),
          })
        "
        :max-length="1000"
      />
      <TagSelector v-model="form.tags" :tags="tags ?? []" />
      <PhotoUploader
        :label="
          $t('components.entity.create_modal.entity_photo', {
            type: t(selectedEntityType ? selectedEntityType.name : 'global.entity'),
          })
        "
        :button-label="$t('components.entity.create_modal.upload_photos')"
        :existing-count="form.photos.length"
        @selected="appendPhotos"
      />
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
  import BaseModal from "@/components/App/CreateModal.vue";
  import type {
    EntityCreate,
    EntityTemplateOut,
    EntityTemplateSummary,
    EntityOut,
    EntityTypeSummary,
  } from "~~/lib/api/types/data-contracts";
  import { useTagStore } from "~/stores/tags";
  import { useLocationStore } from "~~/stores/locations";
  import MdiBarcode from "~icons/mdi/barcode";
  import MdiBarcodeScan from "~icons/mdi/barcode-scan";
  import MdiPackageVariant from "~icons/mdi/package-variant";
  import MdiPackageVariantClosed from "~icons/mdi/package-variant-closed";
  import MdiFileDocumentOutline from "~icons/mdi/file-document-outline";
  import MdiChevronDown from "~icons/mdi/chevron-down";
  import MdiClose from "~icons/mdi/close";
  import { AttachmentTypes } from "~~/lib/api/types/non-generated";
  import { useDialog, useDialogHotkey } from "~/components/ui/dialog-provider";
  import TagSelector from "~/components/Tag/Selector.vue";
  import ItemSelector from "~/components/Item/Selector.vue";
  import TemplateSelector from "~/components/Template/Selector.vue";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "~/components/ui/tooltip";
  import LocationSelector from "~/components/Location/Selector.vue";
  import FormTextField from "~/components/Form/TextField.vue";
  import FormTextArea from "~/components/Form/TextArea.vue";
  import PhotoUploader from "~/components/Form/PhotoUploader.vue";
  import PhotoUploaderPreview from "~/components/Form/PhotoUploaderPreview.vue";
  import {
    deletePhoto,
    dataURLtoFile,
    rotatePhotoPreview,
    setPrimaryPhoto,
    type PhotoPreview,
  } from "~/components/Form/photo-uploader";
  import { useEntityTypeStore } from "~~/stores/entityTypes";
  import EntitySelector from "~/components/Entity/Selector.vue";

  const { t } = useI18n();
  const { openDialog, closeDialog, registerOpenDialogCallback } = useDialog();

  useDialogHotkey(DialogID.CreateEntity, { code: "Digit1", shift: true }, () => ({
    baseType: "item",
  }));
  useDialogHotkey(DialogID.CreateEntity, { code: "Digit2", shift: true }, () => ({
    baseType: "location",
  }));

  const entityTypeStore = useEntityTypeStore();

  const api = useUserApi();

  const locationsStore = useLocationStore();
  const locations = computed(() => locationsStore.allLocations);

  const tagStore = useTagStore();
  const tags = computed(() => tagStore.tags);

  const route = useRoute();

  const parent = ref();
  const { query, results, isLoading, triggerSearch } = useItemSearch(api, { immediate: false });
  const subItemCreate = ref();

  const tagId = computed(() => {
    if (route.fullPath.includes("/tag/")) {
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

  const itemId = computed(() => {
    if (route.fullPath.includes("/item/")) {
      return route.params.id;
    }
    return null;
  });

  const nameInput = ref<HTMLInputElement | null>(null);

  // Entity type selection
  const entityTypes = computed(() => entityTypeStore.allTypes);
  const selectedEntityType = ref<EntityTypeSummary | null>(null);

  async function onEntityTypeChanged(typeId: string) {
    const et = entityTypes.value.find(t => t.id === typeId);
    selectedEntityType.value = et || null;

    // If the selected type has a default template and is not a location, auto-apply it
    if (et?.isLocation || !et?.defaultTemplateId || !et.defaultTemplate) {
      clearTemplate();
    } else {
      const { data, error } = await api.templates.get(et.defaultTemplateId);
      if (!error && data) {
        selectedTemplate.value = {
          id: data.id,
          name: data.name,
          description: data.description,
        } as EntityTemplateSummary;
        templateData.value = data;
        form.quantity = data.defaultQuantity;
        if (data.defaultName) form.name = data.defaultName;
        if (data.defaultDescription) form.description = data.defaultDescription;
        if (data.defaultLocation) {
          const found = locations.value.find(l => l.id === data.defaultLocation!.id);
          if (found) form.location = found;
        }
        if (data.defaultTags && data.defaultTags.length > 0) {
          form.tags = data.defaultTags.map(l => l.id);
        }
        toast.success(t("components.template.toast.applied", { name: data.name }));
      }
    }
  }

  const LAST_TEMPLATE_KEY = "homebox:lastUsedTemplate";

  const loading = ref(false);
  const focused = ref(false);
  const selectedTemplate = ref<EntityTemplateSummary | null>(null);
  const templateData = ref<EntityTemplateOut | null>(null);
  const showTemplateDetails = ref(false);
  const form = reactive({
    location: locations.value && locations.value.length > 0 ? locations.value[0] : ({} as EntityOut),
    parentId: null,
    name: "",
    quantity: 1,
    description: "",
    color: "",
    tags: [] as string[],
    photos: [] as PhotoPreview[],
  });

  async function handleTemplateSelected(template: EntityTemplateSummary | null) {
    if (!template) {
      // Template was deselected, clear template data and remove from storage
      templateData.value = null;
      form.quantity = 1;
      localStorage.removeItem(LAST_TEMPLATE_KEY);
      return;
    }

    // Load full template details
    const { data, error } = await api.templates.get(template.id);
    if (error || !data) {
      toast.error(t("components.template.toast.load_failed"));
      return;
    }

    // Store template data for display and item creation
    templateData.value = data;

    // Pre-fill form with template defaults
    form.quantity = data.defaultQuantity;
    if (data.defaultName) {
      form.name = data.defaultName;
    }
    if (data.defaultDescription) {
      form.description = data.defaultDescription;
    }
    // Pre-fill location if template has one and current form doesn't
    if (data.defaultLocation && !form.location?.id) {
      const found = locations.value.find(l => l.id === data.defaultLocation!.id);
      if (found) {
        form.location = found;
      }
    }
    // Pre-fill tags from template
    if (data.defaultTags && data.defaultTags.length > 0) {
      form.tags = data.defaultTags.map(l => l.id);
    }

    // Save template ID to localStorage for persistence
    localStorage.setItem(LAST_TEMPLATE_KEY, template.id);

    toast.success(t("components.template.toast.applied", { name: data.name }));
  }

  async function restoreLastTemplate() {
    const lastTemplateId = localStorage.getItem(LAST_TEMPLATE_KEY);
    if (!lastTemplateId) return;

    // Load the template details
    const { data, error } = await api.templates.get(lastTemplateId);
    if (error || !data) {
      // Template might have been deleted, clear the stored ID
      localStorage.removeItem(LAST_TEMPLATE_KEY);
      return;
    }

    // Set the template
    selectedTemplate.value = { id: data.id, name: data.name, description: data.description } as EntityTemplateSummary;
    templateData.value = data;
    form.quantity = data.defaultQuantity;
    if (data.defaultName) {
      form.name = data.defaultName;
    }
    if (data.defaultDescription) {
      form.description = data.defaultDescription;
    }
    // Pre-fill location if template has one
    if (data.defaultLocation) {
      const found = locations.value.find(l => l.id === data.defaultLocation!.id);
      if (found) {
        form.location = found;
      }
    }
    // Pre-fill tags from template
    if (data.defaultTags && data.defaultTags.length > 0) {
      form.tags = data.defaultTags.map(l => l.id);
    }
  }

  function clearTemplate() {
    selectedTemplate.value = null;
    templateData.value = null;
    showTemplateDetails.value = false;
    form.quantity = 1;
    localStorage.removeItem(LAST_TEMPLATE_KEY);
  }

  watch(
    parent,
    newParent => {
      if (newParent && newParent.id && subItemCreate.value) {
        form.parentId = newParent.id;
      } else {
        form.parentId = null;
      }
    },
    { immediate: true }
  );

  const { shift } = useMagicKeys();
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
      toast.error(t("components.entity.create_modal.toast.rotate_process_failed"));
      console.error(error);
    }
  }

  onMounted(() => {
    const cleanup = registerOpenDialogCallback(DialogID.CreateEntity, async params => {
      subItemCreate.value = false;
      let parentItemLocationId = null;
      parent.value = {};
      form.parentId = null;

      if (params.baseType === "item") {
        selectedEntityType.value = entityTypes.value.find(t => !t.isLocation) || null;

        subItemCreate.value = params.subItem;

        if (subItemCreate.value && itemId.value) {
          const itemIdRead = typeof itemId.value === "string" ? (itemId.value as string) : itemId.value[0]!;
          const { data, error } = await api.items.get(itemIdRead);
          if (error || !data) {
            toast.error(t("components.entity.create_modal.toast.failed_load_parent"));
            console.error("Parent item fetch error:", error);
          }

          if (data) {
            parent.value = data;
          }

          if (data.parent) {
            const loc = data.parent;
            parentItemLocationId = loc.id;
          }
        }

        if (params.product) {
          form.name = params.product.item.name;
          form.description = params.product.item.description;

          if (params.product.imageURL) {
            appendPhotos([
              {
                photoName: "product_view.jpg",
                fileBase64: params.product.imageBase64,
                primary: form.photos.length === 0,
                file: dataURLtoFile(params.product.imageBase64, "product_view.jpg"),
              },
            ]);
          }
        }

        // Restore last used template if available
        await restoreLastTemplate();
      } else {
        selectedEntityType.value = entityTypes.value.find(t => t.isLocation) || null;
      }

      const locId = locationId.value ? locationId.value : parentItemLocationId;

      if (locId) {
        const found = locations.value.find(l => l.id === locId);
        if (found) {
          form.location = found;
        }
      }

      if (tagId.value) {
        form.tags = tags.value.filter(l => l.id === tagId.value).map(l => l.id);
      }
    });

    onUnmounted(cleanup);
  });

  async function create(close = true) {
    if (!form.location?.id) {
      toast.error(t("components.entity.create_modal.toast.please_select_location"));
      return;
    }

    if (loading.value) {
      toast.error(
        t("components.entity.create_modal.toast.already_creating", {
          type: t(selectedEntityType.value ? selectedEntityType.value.name : "global.entity"),
        })
      );
      return;
    }

    loading.value = true;

    if (shift?.value) close = false;

    let error, data;

    // If the selected entity type is a location, use the location creation endpoint
    if (selectedEntityType.value?.isLocation) {
      const result = await api.items.createLocation({
        name: form.name,
        description: form.description,
        parentId: form.location ? form.location.id : null,
        entityTypeId: selectedEntityType.value?.id || "",
        quantity: 1,
        tagIds: form.tags,
      });
      error = result.error;
      data = result.data;
    } else if (templateData.value) {
      // If a template is selected, use the template creation endpoint
      const templateRequest = {
        name: form.name,
        description: form.description,
        parentId: form.location.id as string,
        tagIds: form.tags,
        quantity: form.quantity,
      };

      const result = await api.templates.createItem(templateData.value.id, templateRequest);
      error = result.error;
      data = result.data;
    } else {
      // Normal item creation without template
      const out: EntityCreate = {
        parentId: form.parentId || (form.location.id as string),
        name: form.name,
        quantity: form.quantity,
        description: form.description,
        tagIds: form.tags,
        entityTypeId: selectedEntityType.value?.id || "",
      };

      const result = await api.items.create(out);
      error = result.error;
      data = result.data;
    }

    if (error) {
      loading.value = false;
      toast.error(
        t("components.entity.create_modal.toast.create_failed", {
          type: t(selectedEntityType.value ? selectedEntityType.value.name : "global.entity"),
        })
      );
      return;
    }

    toast.success(
      t("components.entity.create_modal.toast.create_success", {
        type: t(selectedEntityType.value ? selectedEntityType.value.name : "global.entity"),
      })
    );

    if (form.photos.length > 0) {
      toast.info(t("components.entity.create_modal.toast.uploading_photos", { count: form.photos.length }));
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
          toast.error(t("components.entity.create_modal.toast.upload_failed", { photoName: photo.photoName }));
          console.error(attachError);
        }
      }
      if (uploadError) {
        toast.warning(t("components.entity.create_modal.toast.some_photos_failed", { count: form.photos.length }));
      } else {
        toast.success(t("components.entity.create_modal.toast.upload_success", { count: form.photos.length }));
      }
    }

    form.name = "";
    form.quantity = 1;
    form.description = "";
    form.color = "";
    form.photos = [];
    form.tags = [];
    selectedTemplate.value = null;
    templateData.value = null;
    showTemplateDetails.value = false;
    focused.value = false;
    loading.value = false;

    if (close) {
      closeDialog(DialogID.CreateEntity);
      if (selectedEntityType.value?.isLocation) {
        navigateTo(`/location/${data.id}`);
      } else {
        navigateTo(`/item/${data.id}`);
      }
    }
  }

  function openQrScannerPage() {
    openDialog(DialogID.Scanner);
  }

  function openBarcodeDialog() {
    openDialog(DialogID.ProductImport);
  }
</script>
