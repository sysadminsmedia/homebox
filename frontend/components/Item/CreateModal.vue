<template>
  <BaseModal :dialog-id="DialogID.CreateItem" :title="$t('components.item.create_modal.title')">
    <template #header-actions>
      <div class="flex gap-2">
        <TooltipProvider :delay-duration="0">
          <!-- Template selector button -->
          <Tooltip>
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
                <p>{{ $t("components.item.create_modal.product_tooltip_scan_barcode") }}</p>
              </TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger>
                <Button variant="outline" :disabled="loading" size="icon" data-pos="end" @click="openBarcodeDialog()">
                  <MdiBarcode class="size-5" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>{{ $t("components.item.create_modal.product_tooltip_input_barcode") }}</p>
              </TooltipContent>
            </Tooltip>
          </ButtonGroup>
        </TooltipProvider>
      </div>
    </template>

    <form class="flex flex-col gap-2" @submit.prevent="create()">
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
            :aria-label="$t('components.item.create_modal.clear_template')"
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
            <div v-if="templateData.defaultLabels && templateData.defaultLabels.length > 0" class="mt-1">
              <span class="font-medium">{{ $t("global.labels") }}:</span>
              {{ templateData.defaultLabels.map((l: any) => l.name).join(", ") }}
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
        :label="$t('components.item.create_modal.parent_item')"
        :items="results"
        item-text="name"
        no-results-text="Type to search..."
        :is-loading="isLoading"
        :trigger-search="triggerSearch"
      />
      <FormTextField
        ref="nameInput"
        v-model="form.name"
        :trigger-focus="focused"
        :autofocus="true"
        :label="$t('components.item.create_modal.item_name')"
        :max-length="255"
        :min-length="1"
      />
      <FormTextField v-model="form.quantity" :label="$t('components.item.create_modal.item_quantity')" type="number" />
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
        <div class="relative inline-block">
          <Button type="button" variant="outline" class="w-full" aria-hidden="true" @click.prevent="">
            {{ $t("components.item.create_modal.upload_photos") }}
          </Button>
          <Input
            id="image-create-photo"
            ref="fileInput"
            class="absolute left-0 top-0 size-full cursor-pointer opacity-0"
            type="file"
            accept="image/png,image/jpeg,image/gif,image/avif,image/webp,android/force-camera-workaround"
            multiple
            @change="previewImage"
          />
        </div>
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
                    variant="default"
                    @click.prevent="
                      async () => {
                        await rotateBase64Image90Deg(photo.fileBase64, index);
                      }
                    "
                  >
                    <MdiRotateClockwise />
                    <div class="sr-only">{{ $t("components.item.create_modal.rotate_photo") }}</div>
                  </Button>
                </TooltipTrigger>
                <TooltipContent>
                  <p>{{ $t("components.item.create_modal.rotate_photo") }}</p>
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
                  <p>
                    {{ $t("components.item.create_modal.set_as_primary_photo", { isPrimary: photo.primary }) }}
                  </p>
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
  import BaseModal from "@/components/App/CreateModal.vue";
  import { Label } from "@/components/ui/label";
  import { Input } from "@/components/ui/input";
  import type { ItemCreate, ItemTemplateOut, ItemTemplateSummary, LocationOut } from "~~/lib/api/types/data-contracts";
  import { useLabelStore } from "~~/stores/labels";
  import { useLocationStore } from "~~/stores/locations";
  import MdiBarcode from "~icons/mdi/barcode";
  import MdiBarcodeScan from "~icons/mdi/barcode-scan";
  import MdiPackageVariant from "~icons/mdi/package-variant";
  import MdiPackageVariantClosed from "~icons/mdi/package-variant-closed";
  import MdiDelete from "~icons/mdi/delete";
  import MdiRotateClockwise from "~icons/mdi/rotate-clockwise";
  import MdiStarOutline from "~icons/mdi/star-outline";
  import MdiStar from "~icons/mdi/star";
  import MdiFileDocumentOutline from "~icons/mdi/file-document-outline";
  import MdiChevronDown from "~icons/mdi/chevron-down";
  import MdiClose from "~icons/mdi/close";
  import { AttachmentTypes } from "~~/lib/api/types/non-generated";
  import { useDialog, useDialogHotkey } from "~/components/ui/dialog-provider";
  import LabelSelector from "~/components/Label/Selector.vue";
  import ItemSelector from "~/components/Item/Selector.vue";
  import TemplateSelector from "~/components/Template/Selector.vue";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "~/components/ui/tooltip";
  import LocationSelector from "~/components/Location/Selector.vue";
  import FormTextField from "~/components/Form/TextField.vue";
  import FormTextArea from "~/components/Form/TextArea.vue";

  interface PhotoPreview {
    photoName: string;
    file: File;
    fileBase64: string;
    primary: boolean;
  }

  const { t } = useI18n();
  const { openDialog, closeDialog, registerOpenDialogCallback } = useDialog();

  useDialogHotkey(DialogID.CreateItem, { code: "Digit1", shift: true });

  const api = useUserApi();

  const locationsStore = useLocationStore();
  const locations = computed(() => locationsStore.allLocations);

  const labelStore = useLabelStore();
  const labels = computed(() => labelStore.labels);

  const route = useRoute();
  const router = useRouter();

  const parent = ref();
  const { query, results, isLoading, triggerSearch } = useItemSearch(api, { immediate: false });
  const subItemCreateParam = useRouteQuery("subItemCreate", "n");
  const subItemCreate = ref();

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

  const itemId = computed(() => {
    if (route.fullPath.includes("/item/")) {
      return route.params.id;
    }
    return null;
  });

  const nameInput = ref<HTMLInputElement | null>(null);

  const LAST_TEMPLATE_KEY = "homebox:lastUsedTemplate";

  const loading = ref(false);
  const focused = ref(false);
  const selectedTemplate = ref<ItemTemplateSummary | null>(null);
  const templateData = ref<ItemTemplateOut | null>(null);
  const showTemplateDetails = ref(false);
  const form = reactive({
    location: locations.value && locations.value.length > 0 ? locations.value[0] : ({} as LocationOut),
    parentId: null,
    name: "",
    quantity: 1,
    description: "",
    color: "",
    labels: [] as string[],
    photos: [] as PhotoPreview[],
  });

  async function handleTemplateSelected(template: ItemTemplateSummary | null) {
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
    // Pre-fill labels from template
    if (data.defaultLabels && data.defaultLabels.length > 0) {
      form.labels = data.defaultLabels.map(l => l.id);
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
    selectedTemplate.value = { id: data.id, name: data.name, description: data.description } as ItemTemplateSummary;
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
    // Pre-fill labels from template
    if (data.defaultLabels && data.defaultLabels.length > 0) {
      form.labels = data.defaultLabels.map(l => l.id);
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

  function deleteImage(index: number) {
    form.photos.splice(index, 1);
  }

  function setPrimary(index: number) {
    const primary = form.photos.findIndex(p => p.primary);

    if (primary !== -1) form.photos[primary]!.primary = false;
    if (primary !== index) form.photos[index]!.primary = true;
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
            primary: form.photos.length === 0,
          });
        };
        reader.readAsDataURL(file);
      }
      input.value = "";
    }
  }

  onMounted(() => {
    const cleanup = registerOpenDialogCallback(DialogID.CreateItem, async params => {
      // needed since URL will be cleared in the next step => ParentId Selection should stay though
      subItemCreate.value = subItemCreateParam.value === "y";
      let parentItemLocationId = null;

      if (subItemCreate.value && itemId.value) {
        const itemIdRead = typeof itemId.value === "string" ? (itemId.value as string) : itemId.value[0]!;
        const { data, error } = await api.items.get(itemIdRead);
        if (error || !data) {
          toast.error(t("components.item.create_modal.toast.failed_load_parent"));
          console.error("Parent item fetch error:", error);
        }

        if (data) {
          parent.value = data;
        }

        if (data.location) {
          const { location } = data;
          parentItemLocationId = location.id;
        }

        // clear URL Parameter (subItemCreate) since intention was communicated and received
        const currentQuery = { ...route.query };
        delete currentQuery.subItemCreate;
        await router.push({ query: currentQuery });
      } else {
        // since Input is hidden in this case, make sure no accidental parent information is sent out
        parent.value = {};
        form.parentId = null;
      }

      const locId = locationId.value ? locationId.value : parentItemLocationId;

      if (locId) {
        const found = locations.value.find(l => l.id === locId);
        if (found) {
          form.location = found;
        }
      }

      if (params?.product) {
        form.name = params.product.item.name;
        form.description = params.product.item.description;

        if (params.product.imageURL) {
          form.photos.push({
            photoName: "product_view.jpg",
            fileBase64: params.product.imageBase64,
            primary: form.photos.length === 0,
            file: dataURLtoFile(params.product.imageBase64, "product_view.jpg"),
          });
        }
      }

      if (labelId.value) {
        form.labels = labels.value.filter(l => l.id === labelId.value).map(l => l.id);
      }

      // Restore last used template if available
      await restoreLastTemplate();
    });

    onUnmounted(cleanup);
  });

  async function create(close = true) {
    if (!form.location?.id) {
      toast.error(t("components.item.create_modal.toast.please_select_location"));
      return;
    }

    if (loading.value) {
      toast.error(t("components.item.create_modal.toast.already_creating"));
      return;
    }

    loading.value = true;

    if (shift?.value) close = false;

    let error, data;

    // If a template is selected, use the template creation endpoint
    if (templateData.value) {
      const templateRequest = {
        name: form.name,
        description: form.description,
        locationId: form.location.id as string,
        labelIds: form.labels,
        quantity: form.quantity,
      };

      const result = await api.templates.createItem(templateData.value.id, templateRequest);
      error = result.error;
      data = result.data;
    } else {
      // Normal item creation without template
      const out: ItemCreate = {
        parentId: form.parentId,
        name: form.name,
        quantity: form.quantity,
        description: form.description,
        locationId: form.location.id as string,
        labelIds: form.labels,
      };

      const result = await api.items.create(out);
      error = result.error;
      data = result.data;
    }

    if (error) {
      loading.value = false;
      toast.error(t("components.item.create_modal.toast.create_failed"));
      return;
    }

    toast.success(t("components.item.create_modal.toast.create_success"));

    if (form.photos.length > 0) {
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
          console.error(attachError);
        }
      }
      if (uploadError) {
        toast.warning(t("components.item.create_modal.toast.some_photos_failed", { count: form.photos.length }));
      } else {
        toast.success(t("components.item.create_modal.toast.upload_success", { count: form.photos.length }));
      }
    }

    form.name = "";
    form.quantity = 1;
    form.description = "";
    form.color = "";
    form.photos = [];
    form.labels = [];
    selectedTemplate.value = null;
    templateData.value = null;
    showTemplateDetails.value = false;
    focused.value = false;
    loading.value = false;

    if (close) {
      closeDialog(DialogID.CreateItem);
      navigateTo(`/item/${data.id}`);
    }
  }

  function dataURLtoFile(dataURL: string, fileName: string) {
    try {
      const arr = dataURL.split(",");
      const mimeMatch = arr[0]!.match(/:(.*?);/);
      if (!mimeMatch || !mimeMatch[1]) {
        throw new Error("Invalid data URL format");
      }
      const mime = mimeMatch[1];

      // Validate mime type is an image
      if (!mime.startsWith("image/")) {
        throw new Error("Invalid mime type, expected image");
      }

      const bstr = atob(arr[arr.length - 1]!);
      let n = bstr.length;
      const u8arr = new Uint8Array(n);
      while (n--) {
        u8arr[n] = bstr.charCodeAt(n);
      }
      return new File([u8arr], fileName, { type: mime });
    } catch (error) {
      console.error("Error converting data URL to file:", error);
      // Return a fallback or rethrow based on your error handling strategy
      throw error;
    }
  }

  async function rotateBase64Image90Deg(base64Image: string, index: number) {
    // Create an off-screen canvas
    const offScreenCanvas = document.createElement("canvas");
    const offScreenCanvasCtx = offScreenCanvas.getContext("2d");

    if (!offScreenCanvasCtx) {
      toast.error(t("components.item.create_modal.toast.no_canvas_support"));
      return;
    }

    // Create an image
    const img = new Image();

    // Create a promise to handle the image loading
    await new Promise<void>((resolve, reject) => {
      img.onload = () => resolve();
      img.onerror = () => reject(new Error("Failed to load image"));
      img.src = base64Image;
    }).catch(error => {
      toast.error(t("components.item.create_modal.toast.rotate_failed", { error: error.message }));
    });

    // Set its dimensions to rotated size
    offScreenCanvas.height = img.width;
    offScreenCanvas.width = img.height;

    // Rotate and draw source image into the off-screen canvas
    offScreenCanvasCtx.rotate((90 * Math.PI) / 180);
    offScreenCanvasCtx.translate(0, -offScreenCanvas.width);
    offScreenCanvasCtx.drawImage(img, 0, 0);

    const imageType = base64Image.match(/^data:(.+);base64/)?.[1] || "image/jpeg";

    // Encode image to data-uri with base64
    try {
      form.photos[index]!.fileBase64 = offScreenCanvas.toDataURL(imageType, 100);
      form.photos[index]!.file = dataURLtoFile(form.photos[index]!.fileBase64, form.photos[index]!.photoName);
    } catch (error) {
      toast.error(t("components.item.create_modal.toast.rotate_process_failed"));
      console.error(error);
    } finally {
      // Clean up resources
      offScreenCanvas.width = 0;
      offScreenCanvas.height = 0;
    }
  }

  function openQrScannerPage() {
    openDialog(DialogID.Scanner);
  }

  function openBarcodeDialog() {
    openDialog(DialogID.ProductImport);
  }
</script>
