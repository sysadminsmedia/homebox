<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import type { ItemAttachment, ItemField, ItemOut, ItemUpdate } from "~~/lib/api/types/data-contracts";
  import { AttachmentTypes } from "~~/lib/api/types/non-generated";
  import { useLabelStore } from "~~/stores/labels";
  import { useLocationStore } from "~~/stores/locations";
  import MdiDelete from "~icons/mdi/delete";
  import MdiPencil from "~icons/mdi/pencil";
  import MdiContentSaveOutline from "~icons/mdi/content-save-outline";
  import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import { Button } from "@/components/ui/button";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { Checkbox } from "@/components/ui/checkbox";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import { Tooltip, TooltipContent, TooltipTrigger, TooltipProvider } from "@/components/ui/tooltip";
  import { Switch } from "@/components/ui/switch";
  import { Label } from "@/components/ui/label";

  const { t } = useI18n();

  const { openDialog, closeDialog } = useDialog();

  definePageMeta({
    middleware: ["auth"],
  });

  const route = useRoute();
  const api = useUserApi();
  const preferences = useViewPreferences();

  const itemId = computed<string>(() => route.params.id as string);

  const locationStore = useLocationStore();
  const locations = computed(() => locationStore.allLocations);

  const labelStore = useLabelStore();
  const labels = computed(() => labelStore.labels);

  const {
    data: nullableItem,
    refresh,
    pending: requestPending,
  } = useAsyncData(async () => {
    const { data, error } = await api.items.get(itemId.value);
    if (error) {
      toast.error(t("items.toast.failed_load_item"));
      navigateTo("/home");
      return;
    }

    if (locations && data.location?.id) {
      // @ts-expect-error - we know the locations is valid
      const location = locations.value.find(l => l.id === data.location.id);
      if (location) {
        data.location = location;
      }
    }

    if (data.parent) {
      parent.value = data.parent;
    }

    return data;
  });

  const item = ref<ItemOut & { labelIds: string[] }>(null as any);

  watchEffect(() => {
    if (nullableItem.value) {
      item.value = {
        ...nullableItem.value,
        labelIds: nullableItem.value.labels.map(l => l.id) ?? [],
      };
    }
  });

  // const item = computed(() => nullableItem.value as ItemOut);

  onMounted(() => {
    refresh();
  });

  async function saveItem() {
    if (!item.value.location?.id) {
      toast.error(t("items.toast.failed_save_no_location"));
      return;
    }

    let purchasePrice = 0;
    let soldPrice = 0;
    if (item.value.purchasePrice) {
      purchasePrice = item.value.purchasePrice;
    }
    if (item.value.soldPrice) {
      soldPrice = item.value.soldPrice;
    }

    console.log((item.value.purchasePrice ??= 0));
    console.log((item.value.soldPrice ??= 0));

    const payload: ItemUpdate = {
      ...item.value,
      locationId: item.value.location?.id,
      labelIds: item.value.labelIds,
      parentId: parent.value ? parent.value.id : null,
      assetId: item.value.assetId,
      purchasePrice,
      soldPrice,
      purchaseTime: item.value.purchaseTime as Date,
    };

    const { error } = await api.items.update(itemId.value, payload);

    if (error) {
      toast.error(t("items.toast.failed_save"));
      return;
    }

    toast.success(t("items.toast.item_saved"));
    navigateTo("/item/" + itemId.value);
  }
  type NoUndefinedField<T> = { [P in keyof T]-?: NoUndefinedField<NonNullable<T[P]>> };

  type StringKeys<T> = { [k in keyof T]: T[k] extends string ? k : never }[keyof T];
  type OnlyString<T> = { [k in StringKeys<T>]: string };

  type NumberKeys<T> = { [k in keyof T]: T[k] extends number ? k : never }[keyof T];
  type OnlyNumber<T> = { [k in NumberKeys<T>]: number };

  type TextFormField = {
    type: "text" | "textarea";
    label: string;
    // key of ItemOut where the value is a string
    ref: keyof OnlyString<NoUndefinedField<ItemOut>>;
    maxLength?: number;
    minLength?: number;
  };

  type NumberFormField = {
    type: "number";
    label: string;
    ref: keyof OnlyNumber<NoUndefinedField<ItemOut>> | keyof OnlyString<NoUndefinedField<ItemOut>>;
  };

  // https://stackoverflow.com/questions/50851263/how-do-i-require-a-keyof-to-be-for-a-property-of-a-specific-type
  // I don't know why typescript can't just be normal
  type BooleanKeys<T> = { [k in keyof T]: T[k] extends boolean ? k : never }[keyof T];
  type OnlyBoolean<T> = { [k in BooleanKeys<T>]: boolean };

  interface BoolFormField {
    type: "checkbox";
    label: string;
    ref: keyof OnlyBoolean<NoUndefinedField<ItemOut>>;
  }

  type DateKeys<T> = { [k in keyof T]: T[k] extends Date | string ? k : never }[keyof T];
  type OnlyDate<T> = { [k in DateKeys<T>]: Date | string };

  type DateFormField = {
    type: "date";
    label: string;
    ref: keyof OnlyDate<NoUndefinedField<ItemOut>>;
  };

  type FormField = TextFormField | BoolFormField | DateFormField | NumberFormField;

  const mainFields: FormField[] = [
    {
      type: "text",
      label: "items.name",
      ref: "name",
      maxLength: 255,
      minLength: 1,
    },
    {
      type: "number",
      label: "items.quantity",
      ref: "quantity",
    },
    {
      type: "textarea",
      label: "items.description",
      ref: "description",
      maxLength: 1000,
    },
    {
      type: "text",
      label: "items.serial_number",
      ref: "serialNumber",
      maxLength: 255,
    },
    {
      type: "text",
      label: "items.model_number",
      ref: "modelNumber",
      maxLength: 255,
    },
    {
      type: "text",
      label: "items.manufacturer",
      ref: "manufacturer",
      maxLength: 255,
    },
    {
      type: "textarea",
      label: "items.notes",
      ref: "notes",
      maxLength: 1000,
    },
    {
      type: "checkbox",
      label: "items.insured",
      ref: "insured",
    },
    {
      type: "checkbox",
      label: "items.archived",
      ref: "archived",
    },
    {
      type: "text",
      label: "items.asset_id",
      ref: "assetId",
    },
  ];

  const purchaseFields: FormField[] = [
    {
      type: "text",
      label: "items.purchased_from",
      ref: "purchaseFrom",
      maxLength: 255,
    },
    {
      type: "number",
      label: "items.purchase_price",
      ref: "purchasePrice",
    },
    {
      type: "date",
      label: "items.purchase_date",
      // @ts-expect-error - we know this is a date
      ref: "purchaseTime",
    },
  ];

  const warrantyFields: FormField[] = [
    {
      type: "checkbox",
      label: "items.lifetime_warranty",
      ref: "lifetimeWarranty",
    },
    {
      type: "date",
      label: "items.warranty_expires",
      // @ts-expect-error - we know this is a date
      ref: "warrantyExpires",
    },
    {
      type: "textarea",
      label: "items.warranty_details",
      ref: "warrantyDetails",
      maxLength: 1000,
    },
  ];

  const soldFields: FormField[] = [
    {
      type: "text",
      label: "items.sold_to",
      ref: "soldTo",
      maxLength: 255,
    },
    {
      type: "number",
      label: "items.sold_price",
      ref: "soldPrice",
    },
    {
      type: "date",
      label: "items.sold_at",
      // @ts-expect-error - we know this is a date
      ref: "soldTime",
    },
  ];

  // - Attachments
  const attDropZone = ref<HTMLDivElement>();
  const { isOverDropZone: attDropZoneActive } = useDropZone(attDropZone);

  const refAttachmentInput = ref<HTMLInputElement>();

  function clickUpload() {
    if (!refAttachmentInput.value) {
      return;
    }
    refAttachmentInput.value.click();
  }

  function uploadImage(e: Event) {
    const files = (e.target as HTMLInputElement).files;
    if (!files || !files.item(0)) {
      return;
    }

    const first = files.item(0);
    if (!first) {
      return;
    }

    uploadAttachment([first], null);
  }

  const dropPhoto = (files: File[] | null) => uploadAttachment(files, AttachmentTypes.Photo);
  const dropAttachment = (files: File[] | null) => uploadAttachment(files, AttachmentTypes.Attachment);
  const dropWarranty = (files: File[] | null) => uploadAttachment(files, AttachmentTypes.Warranty);
  const dropManual = (files: File[] | null) => uploadAttachment(files, AttachmentTypes.Manual);
  const dropReceipt = (files: File[] | null) => uploadAttachment(files, AttachmentTypes.Receipt);

  async function uploadAttachment(files: File[] | null, type: AttachmentTypes | null) {
    if (!files || files.length === 0) {
      return;
    }

    const { data, error } = await api.items.attachments.add(itemId.value, files[0], files[0].name, type);

    if (error) {
      toast.error(t("items.toast.failed_upload_attachment"));
      return;
    }

    toast.success(t("items.toast.attachment_uploaded"));

    item.value.attachments = data.attachments;
  }

  const confirm = useConfirm();

  async function deleteAttachment(attachmentId: string) {
    const confirmed = await confirm.open(t("items.delete_attachment_confirm"));

    if (confirmed.isCanceled) {
      return;
    }

    const { error } = await api.items.attachments.delete(itemId.value, attachmentId);

    if (error) {
      toast.error(t("items.toast.failed_delete_attachment"));
      return;
    }

    toast.success(t("items.toast.attachment_deleted"));
    item.value.attachments = item.value.attachments.filter(a => a.id !== attachmentId);
  }

  const editState = reactive({
    loading: false,

    // Values
    obj: {},
    id: "",
    title: "",
    type: "",
    primary: false,
  });

  const attachmentOpts = Object.entries(AttachmentTypes).map(([key, value]) => ({
    text: key[0].toUpperCase() + key.slice(1),
    value,
  }));

  function openAttachmentEditDialog(attachment: ItemAttachment) {
    editState.id = attachment.id;
    editState.title = attachment.title;
    editState.type = attachment.type;
    editState.primary = attachment.primary;
    openDialog("attachment-edit");

    editState.obj = attachmentOpts.find(o => o.value === attachment.type) || attachmentOpts[0];
  }

  async function updateAttachment() {
    editState.loading = true;
    const { error, data } = await api.items.attachments.update(itemId.value, editState.id, {
      title: editState.title,
      type: editState.type,
      primary: editState.primary,
    });

    if (error) {
      toast.error(t("items.toast.failed_delete_attachment"));
      return;
    }

    item.value.attachments = data.attachments;

    editState.loading = false;
    closeDialog("attachment-edit");

    editState.id = "";
    editState.title = "";
    editState.type = "";

    toast.success(t("items.toast.attachment_updated"));
  }

  function addField() {
    item.value.fields.push({
      id: null,
      name: "Field Name",
      type: "text",
      textValue: "",
      numberValue: 0,
      booleanValue: false,
      timeValue: null,
    } as unknown as ItemField);
  }

  const { query, results } = useItemSearch(api, { immediate: false });
  const parent = ref();

  async function keyboardSave(e: KeyboardEvent) {
    // Cmd + S
    if (e.metaKey && e.key === "s") {
      e.preventDefault();
      await saveItem();
    }

    // Ctrl + S
    if (e.ctrlKey && e.key === "s") {
      e.preventDefault();
      await saveItem();
    }
  }

  async function maybeSyncWithParentLocation() {
    if (parent.value && parent.value.id) {
      const { data, error } = await api.items.get(parent.value.id);

      if (error) {
        toast.error(t("items.toast.error_loading_parent_data"));
        return;
      }

      if (data.syncChildItemsLocations) {
        toast.info(t("items.toast.sync_child_location"));
        item.value.location = data.location;
      }
    }
  }

  async function informAboutDesyncingLocationFromParent() {
    if (parent.value && parent.value.id) {
      const { data, error } = await api.items.get(parent.value.id);

      if (error) {
        toast.error(t("items.toast.error_loading_parent_data"));
        return;
      }

      if (data.syncChildItemsLocations) {
        toast.info(t("items.toast.child_location_desync"));
      }
    }
  }

  async function syncChildItemsLocations() {
    if (!item.value.location?.id) {
      toast.error(t("items.toast.failed_save_no_location"));
      return;
    }

    const payload: ItemUpdate = {
      ...item.value,
      locationId: item.value.location?.id,
      labelIds: item.value.labelIds,
      parentId: parent.value ? parent.value.id : null,
      assetId: item.value.assetId,
    };

    const { error } = await api.items.update(itemId.value, payload);

    if (error) {
      toast.error("Failed to save item");
      return;
    }

    if (!item.value.syncChildItemsLocations) {
      toast.success(t("items.toast.child_items_location_no_longer_synced"));
    } else {
      toast.success(t("items.toast.child_items_location_synced"));
    }
  }

  onMounted(() => {
    window.addEventListener("keydown", keyboardSave);
  });

  onUnmounted(() => {
    window.removeEventListener("keydown", keyboardSave);
  });
</script>

<template>
  <div v-if="item" class="pb-8">
    <Dialog dialog-id="attachment-edit">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ $t("items.edit.edit_attachment_dialog.title") }}</DialogTitle>
        </DialogHeader>

        <FormTextField v-model="editState.title" :label="$t('items.edit.edit_attachment_dialog.attachment_title')" />
        <div>
          <Label for="attachment-type"> {{ $t("items.edit.edit_attachment_dialog.attachment_type") }} </Label>
          <Select id="attachment-type" v-model:model-value="editState.type">
            <SelectTrigger>
              <SelectValue :placeholder="$t('items.edit.edit_attachment_dialog.select_type')" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="opt in attachmentOpts" :key="opt.value" :value="opt.value">
                {{ opt.text }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div v-if="editState.type == 'photo'" class="mt-3 flex items-center gap-2">
          <Checkbox
            id="primary"
            v-model="editState.primary"
            :label="$t('items.edit.edit_attachment_dialog.primary_photo')"
          />
          <label class="cursor-pointer text-sm" for="primary">
            <span class="font-semibold">{{ $t("items.edit.edit_attachment_dialog.primary_photo") }}</span>
            {{ $t("items.edit.edit_attachment_dialog.primary_photo_sub") }}
          </label>
        </div>

        <DialogFooter>
          <Button :loading="editState.loading" @click="updateAttachment"> {{ $t("global.update") }} </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <section class="relative">
      <!-- IMPORTANT: this is based on the height of the topbar + 0.25rem -->
      <div class="sticky top-[7.25rem] z-10 my-4 flex items-center justify-between gap-2 sm:top-[4.25rem]">
        <TooltipProvider :delay-duration="0">
          <Tooltip>
            <TooltipTrigger as-child>
              <Label class="flex cursor-pointer items-center gap-2 backdrop-blur-sm">
                <Switch v-model="preferences.editorAdvancedView" />
                {{ $t("items.advanced") }}
              </Label>
            </TooltipTrigger>
            <TooltipContent>{{ $t("items.show_advanced_view_options") }}</TooltipContent>
          </Tooltip>
        </TooltipProvider>
        <Button size="sm" @click="saveItem">
          <MdiContentSaveOutline />
          {{ $t("global.save") }}
        </Button>
      </div>
      <div v-if="!requestPending" class="space-y-6">
        <BaseCard class="overflow-visible">
          <template #title> {{ $t("items.edit_details") }} </template>
          <div class="mb-6 grid gap-4 border-t px-5 pt-2 md:grid-cols-2">
            <LocationSelector v-model="item.location" @update:model-value="informAboutDesyncingLocationFromParent()" />
            <ItemSelector
              v-model="parent"
              v-model:search="query"
              :items="results"
              item-text="name"
              :label="$t('items.parent_item')"
              no-results-text="Type to search..."
              @update:model-value="maybeSyncWithParentLocation()"
            />
            <div class="flex flex-col gap-2">
              <Label class="px-1">{{ $t("items.sync_child_locations") }}</Label>
              <Switch v-model="item.syncChildItemsLocations" @update:model-value="syncChildItemsLocations()" />
            </div>
            <LabelSelector v-model="item.labelIds" :labels="labels" />
          </div>

          <div class="border-t sm:p-0">
            <div v-for="field in mainFields" :key="field.ref" class="grid grid-cols-1 sm:divide-y">
              <div class="border-b px-4 pb-4 pt-2 sm:px-6">
                <FormTextArea
                  v-if="field.type === 'textarea'"
                  v-model="item[field.ref]"
                  :label="$t(field.label)"
                  inline
                  :max-length="field.maxLength"
                  :min-length="field.minLength"
                />
                <FormTextField
                  v-else-if="field.type === 'text'"
                  v-model="item[field.ref]"
                  :label="$t(field.label)"
                  inline
                  type="text"
                  :max-length="field.maxLength"
                  :min-length="field.minLength"
                />
                <FormTextField
                  v-else-if="field.type === 'number'"
                  v-model.number="item[field.ref]"
                  type="number"
                  :label="$t(field.label)"
                  inline
                />
                <FormDatePicker
                  v-else-if="field.type === 'date'"
                  v-model="item[field.ref]"
                  :label="$t(field.label)"
                  inline
                />
                <FormCheckbox
                  v-else-if="field.type === 'checkbox'"
                  v-model="item[field.ref]"
                  :label="$t(field.label)"
                  inline
                />
              </div>
            </div>
          </div>
        </BaseCard>

        <BaseCard v-if="preferences.editorAdvancedView">
          <template #title> {{ $t("items.custom_fields") }} </template>
          <div class="space-y-4 divide-y border-t px-5">
            <div
              v-for="(field, idx) in item.fields"
              :key="`field-${idx}`"
              class="grid grid-cols-2 gap-2 pt-4 md:grid-cols-4"
            >
              <!-- <FormSelect v-model:value="field.type" label="Field Type" :items="fieldTypes" value-key="value" /> -->
              <FormTextField v-model="field.name" :label="$t('global.name')" />
              <div class="col-span-3 flex items-end">
                <FormTextField v-model="field.textValue" :label="$t('global.value')" :max-length="500" />
                <Tooltip>
                  <TooltipTrigger as-child>
                    <Button size="icon" variant="destructive" class="ml-2" @click="item.fields.splice(idx, 1)">
                      <MdiDelete />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>{{ $t("global.delete") }}</TooltipContent>
                </Tooltip>
              </div>
            </div>
          </div>
          <div class="mt-4 flex justify-end px-5 pb-4">
            <Button size="sm" @click="addField"> {{ $t("global.add") }} </Button>
          </div>
        </BaseCard>

        <Card ref="attDropZone" class="overflow-visible shadow-xl">
          <div class="px-4 py-5 sm:px-6">
            <h3 class="text-lg font-medium leading-6">{{ $t("items.attachments") }}</h3>
            <p class="text-xs">{{ $t("items.changes_persisted_immediately") }}</p>
          </div>
          <div class="border-t p-4">
            <div v-if="attDropZoneActive" class="grid grid-cols-4 gap-4">
              <DropZone @drop="dropPhoto"> {{ $t("items.photos") }} </DropZone>
              <DropZone @drop="dropWarranty"> {{ $t("items.warranty") }} </DropZone>
              <DropZone @drop="dropManual"> {{ $t("items.manuals") }} </DropZone>
              <DropZone @drop="dropAttachment"> {{ $t("items.attachments") }} </DropZone>
              <DropZone @drop="dropReceipt"> {{ $t("items.receipts") }} </DropZone>
            </div>
            <button
              v-else
              class="grid h-24 w-full place-content-center border-2 border-dashed border-primary"
              @click="clickUpload"
            >
              <input ref="refAttachmentInput" hidden type="file" @change="uploadImage" />
              <p>{{ $t("items.drag_and_drop") }}</p>
            </button>
          </div>

          <div class="border-t p-4">
            <ul role="list" class="divide-y rounded-md border">
              <li
                v-for="attachment in item.attachments"
                :key="attachment.id"
                class="grid grid-cols-6 justify-between py-3 pl-3 pr-4 text-sm"
              >
                <p class="col-span-4 my-auto">
                  {{ attachment.title }}
                </p>
                <p class="my-auto">
                  {{ $t(`items.${attachment.type}`) }}
                </p>
                <div class="flex justify-end gap-2">
                  <Tooltip>
                    <TooltipTrigger as-child>
                      <Button variant="destructive" size="icon" @click="deleteAttachment(attachment.id)">
                        <MdiDelete />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>{{ $t("global.delete") }}</TooltipContent>
                  </Tooltip>
                  <Tooltip>
                    <TooltipTrigger as-child>
                      <Button size="icon" @click="openAttachmentEditDialog(attachment)">
                        <MdiPencil />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>{{ $t("global.edit") }}</TooltipContent>
                  </Tooltip>
                </div>
              </li>
            </ul>
          </div>
        </Card>

        <Card v-if="preferences.editorAdvancedView" class="overflow-visible shadow-xl">
          <div class="px-4 py-5 sm:px-6">
            <h3 class="text-lg font-medium leading-6">{{ $t("items.purchase_details") }}</h3>
          </div>
          <div class="border-t sm:p-0">
            <div v-for="field in purchaseFields" :key="field.ref" class="grid grid-cols-1 sm:divide-y">
              <div class="border-b px-4 pb-4 pt-2 sm:px-6">
                <FormTextArea
                  v-if="field.type === 'textarea'"
                  v-model="item[field.ref]"
                  :label="$t(field.label)"
                  inline
                  :max-length="field.maxLength"
                  :min-length="field.minLength"
                />
                <FormTextField
                  v-else-if="field.type === 'text'"
                  v-model="item[field.ref]"
                  :label="$t(field.label)"
                  inline
                  :max-length="field.maxLength"
                  :min-length="field.minLength"
                />
                <FormTextField
                  v-else-if="field.type === 'number'"
                  v-model.number="item[field.ref]"
                  type="number"
                  :label="$t(field.label)"
                  inline
                />
                <FormDatePicker
                  v-else-if="field.type === 'date'"
                  v-model="item[field.ref]"
                  :label="$t(field.label)"
                  inline
                />
                <FormCheckbox
                  v-else-if="field.type === 'checkbox'"
                  v-model="item[field.ref]"
                  :label="$t(field.label)"
                  inline
                />
              </div>
            </div>
          </div>
        </Card>

        <Card v-if="preferences.editorAdvancedView" class="overflow-visible shadow-xl">
          <div class="px-4 py-5 sm:px-6">
            <h3 class="text-lg font-medium leading-6">{{ $t("items.warranty_details") }}</h3>
          </div>
          <div class="border-t sm:p-0">
            <div v-for="field in warrantyFields" :key="field.ref" class="grid grid-cols-1 sm:divide-y">
              <div class="border-b px-4 pb-4 pt-2 sm:px-6">
                <FormTextArea
                  v-if="field.type === 'textarea'"
                  v-model="item[field.ref]"
                  :label="$t(field.label)"
                  inline
                  :max-length="field.maxLength"
                  :min-length="field.minLength"
                />
                <FormTextField
                  v-else-if="field.type === 'text'"
                  v-model="item[field.ref]"
                  :label="$t(field.label)"
                  inline
                  :max-length="field.maxLength"
                  :min-length="field.minLength"
                />
                <FormTextField
                  v-else-if="field.type === 'number'"
                  v-model.number="item[field.ref]"
                  type="number"
                  :label="$t(field.label)"
                  inline
                />
                <FormDatePicker
                  v-else-if="field.type === 'date'"
                  v-model="item[field.ref]"
                  :label="$t(field.label)"
                  inline
                />
                <FormCheckbox
                  v-else-if="field.type === 'checkbox'"
                  v-model="item[field.ref]"
                  :label="$t(field.label)"
                  inline
                />
              </div>
            </div>
          </div>
        </Card>

        <Card v-if="preferences.editorAdvancedView" class="overflow-visible shadow-xl">
          <div class="px-4 py-5 sm:px-6">
            <h3 class="text-lg font-medium leading-6">{{ $t("items.sold_details") }}</h3>
          </div>
          <div class="border-t sm:p-0">
            <div v-for="field in soldFields" :key="field.ref" class="grid grid-cols-1 sm:divide-y">
              <div class="border-b px-4 pb-4 pt-2 sm:px-6">
                <FormTextArea
                  v-if="field.type === 'textarea'"
                  v-model="item[field.ref]"
                  :label="$t(field.label)"
                  inline
                  :max-length="field.maxLength"
                  :min-length="field.minLength"
                />
                <FormTextField
                  v-else-if="field.type === 'text'"
                  v-model="item[field.ref]"
                  :label="$t(field.label)"
                  inline
                  :max-length="field.maxLength"
                  :min-length="field.minLength"
                />
                <FormTextField
                  v-else-if="field.type === 'number'"
                  v-model.number="item[field.ref]"
                  type="number"
                  :label="$t(field.label)"
                  inline
                />
                <FormDatePicker
                  v-else-if="field.type === 'date'"
                  v-model="item[field.ref]"
                  :label="$t(field.label)"
                  inline
                />
                <FormCheckbox
                  v-else-if="field.type === 'checkbox'"
                  v-model="item[field.ref]"
                  :label="$t(field.label)"
                  inline
                />
              </div>
            </div>
          </div>
        </Card>
      </div>
    </section>
  </div>
</template>
