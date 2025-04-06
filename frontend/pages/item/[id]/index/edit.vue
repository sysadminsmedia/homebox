<script setup lang="ts">
  import type { ItemAttachment, ItemField, ItemOut, ItemUpdate } from "~~/lib/api/types/data-contracts";
  import { AttachmentTypes } from "~~/lib/api/types/non-generated";
  import { useLabelStore } from "~~/stores/labels";
  import { useLocationStore } from "~~/stores/locations";
  import Autocomplete from "~~/components/Form/Autocomplete.vue";
  import MdiDelete from "~icons/mdi/delete";
  import MdiPencil from "~icons/mdi/pencil";
  import MdiContentSaveOutline from "~icons/mdi/content-save-outline";
  import MdiContentCopy from "~icons/mdi/content-copy";

  definePageMeta({
    middleware: ["auth"],
  });

  const route = useRoute();
  const api = useUserApi();
  const toast = useNotifier();
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
      toast.error("Failed to load item");
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

  const item = computed<ItemOut>(() => nullableItem.value as ItemOut);

  onMounted(() => {
    refresh();
  });

  async function duplicateItem() {
    const { error, data } = await api.items.create({
      name: `${item.value.name} Copy`,
      description: item.value.description,
      locationId: item.value.location!.id,
      parentId: item.value.parent?.id,
      labelIds: item.value.labels.map(l => l.id),
    });

    if (error) {
      toast.error("Failed to duplicate item");
      return;
    }

    // add extra fields
    const { error: updateError } = await api.items.update(data.id, {
      ...item.value,
      id: data.id,
      labelIds: data.labels.map(l => l.id),
      locationId: data.location!.id,
      name: data.name,
    });

    if (updateError) {
      toast.error("Failed to duplicate item");
      return;
    }

    navigateTo(`/item/${data.id}`);
  }

  async function saveItem() {
    if (!item.value.location?.id) {
      toast.error("Failed to save item: no location selected");
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
      labelIds: item.value.labels.map(l => l.id),
      parentId: parent.value ? parent.value.id : null,
      assetId: item.value.assetId,
      purchasePrice,
      soldPrice,
      purchaseTime: item.value.purchaseTime as Date,
    };

    const { error } = await api.items.update(itemId.value, payload);

    if (error) {
      toast.error("Failed to save item");
      return;
    }

    toast.success("Item saved");
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
      toast.error("Failed to upload attachment");
      return;
    }

    toast.success("Attachment uploaded");

    item.value.attachments = data.attachments;
  }

  const confirm = useConfirm();

  async function deleteAttachment(attachmentId: string) {
    const confirmed = await confirm.open("Are you sure you want to delete this attachment?");

    if (confirmed.isCanceled) {
      return;
    }

    const { error } = await api.items.attachments.delete(itemId.value, attachmentId);

    if (error) {
      toast.error("Failed to delete attachment");
      return;
    }

    toast.success("Attachment deleted");
    item.value.attachments = item.value.attachments.filter(a => a.id !== attachmentId);
  }

  const editState = reactive({
    modal: false,
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
    editState.modal = true;

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
      toast.error("Failed to update attachment");
      return;
    }

    item.value.attachments = data.attachments;

    editState.loading = false;
    editState.modal = false;

    editState.id = "";
    editState.title = "";
    editState.type = "";

    toast.success("Attachment updated");
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

  async function deleteItem() {
    const confirmed = await confirm.open("Are you sure you want to delete this item?");

    if (!confirmed.data) {
      return;
    }

    const { error } = await api.items.delete(itemId.value);
    if (error) {
      toast.error("Failed to delete item");
      return;
    }
    toast.success("Item deleted");
    navigateTo("/home");
  }

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
        toast.error("Something went wrong trying to load parent data");
        return;
      }

      if (data.syncChildItemsLocations) {
        toast.info("Selected parent syncs its children's locations to its own. The location has been updated.");
        item.value.location = data.location;
      }
    }
  }

  async function informAboutDesyncingLocationFromParent() {
    if (parent.value && parent.value.id) {
      const { data, error } = await api.items.get(parent.value.id);

      if (error) {
        toast.error("Something went wrong trying to load parent data");
        return;
      }

      if (data.syncChildItemsLocations) {
        toast.info("Changing location will de-sync it from the parent's location");
      }
    }
  }

  async function syncChildItemsLocations() {
    if (!item.value.location?.id) {
      toast.error("Failed to save item: no location selected");
      return;
    }

    const payload: ItemUpdate = {
      ...item.value,
      locationId: item.value.location?.id,
      labelIds: item.value.labels.map(l => l.id),
      parentId: parent.value ? parent.value.id : null,
      assetId: item.value.assetId,
    };

    const { error } = await api.items.update(itemId.value, payload);

    if (error) {
      toast.error("Failed to save item");
      return;
    }

    if (!item.value.syncChildItemsLocations) {
      toast.success("Child items' locations will no longer be synced with this item.");
    } else {
      toast.success("Child items' locations have been synced with this item");
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
    <BaseModal v-model="editState.modal">
      <template #title> Attachment Edit </template>

      <FormTextField v-model="editState.title" label="Attachment Title" />
      <FormSelect
        v-model:value="editState.type"
        label="Attachment Type"
        value-key="value"
        name="text"
        :items="attachmentOpts"
      />
      <div v-if="editState.type == 'photo'" class="mt-3 flex gap-2">
        <input v-model="editState.primary" type="checkbox" class="checkbox" />
        <p class="text-sm">
          <span class="font-semibold">Primary Photo</span>
          This options is only available for photos. Only one photo can be primary. If you select this option, the
          current primary photo, if any will be unselected.
        </p>
      </div>
      <div class="modal-action">
        <BaseButton :loading="editState.loading" @click="updateAttachment"> Update </BaseButton>
      </div>
    </BaseModal>

    <section class="relative">
      <div class="sticky top-1 z-10 my-4 flex items-center justify-end gap-2">
        <div class="tooltip tooltip-right mr-auto" :data-tip="$t('items.show_advanced_view_options')">
          <label class="label mr-auto cursor-pointer">
            <input v-model="preferences.editorAdvancedView" type="checkbox" class="toggle toggle-primary" />
            <span class="label-text ml-4"> {{ $t("items.advanced") }} </span>
          </label>
        </div>
        <BaseButton size="sm" class="btn" @click="duplicateItem">
          <template #icon>
            <MdiContentCopy />
          </template>
          {{ $t("global.duplicate") }}
        </BaseButton>
        <BaseButton size="sm" @click="saveItem">
          <template #icon>
            <MdiContentSaveOutline />
          </template>
          {{ $t("global.save") }}
        </BaseButton>
        <BaseButton class="btn btn-error btn-sm" @click="deleteItem()">
          <MdiDelete class="mr-2" />
          {{ $t("global.delete") }}
        </BaseButton>
      </div>
      <div v-if="!requestPending" class="space-y-6">
        <BaseCard class="overflow-visible">
          <template #title> {{ $t("items.edit_details") }} </template>
          <template #title-actions>
            <div class="mt-2 flex flex-wrap items-center justify-between gap-4"></div>
          </template>
          <div class="mb-6 grid gap-4 border-t px-5 pt-2 md:grid-cols-2">
            <LocationSelector v-model="item.location" @update:model-value="informAboutDesyncingLocationFromParent()" />
            <FormMultiselect v-model="item.labels" :label="$t('global.labels')" :items="labels ?? []" />
            <FormToggle
              v-model="item.syncChildItemsLocations"
              label="Sync child items' locations"
              inline
              @update:model-value="syncChildItemsLocations()"
            />
            <Autocomplete
              v-if="preferences.editorAdvancedView"
              v-model="parent"
              v-model:search="query"
              :items="results"
              item-text="name"
              :label="$t('items.parent_item')"
              no-results-text="Type to search..."
              @update:model-value="maybeSyncWithParentLocation()"
            />
          </div>

          <div class="border-t border-gray-300 sm:p-0">
            <div v-for="field in mainFields" :key="field.ref" class="grid grid-cols-1 sm:divide-y sm:divide-gray-300">
              <div class="border-b border-gray-300 px-4 pb-4 pt-2 sm:px-6">
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

        <BaseCard>
          <template #title> {{ $t("items.custom_fields") }} </template>
          <div class="space-y-4 divide-y divide-gray-300 border-t px-5">
            <div
              v-for="(field, idx) in item.fields"
              :key="`field-${idx}`"
              class="grid grid-cols-2 gap-2 md:grid-cols-4"
            >
              <!-- <FormSelect v-model:value="field.type" label="Field Type" :items="fieldTypes" value-key="value" /> -->
              <FormTextField v-model="field.name" :label="$t('global.name')" />
              <div class="col-span-3 flex items-end">
                <FormTextField v-model="field.textValue" :label="$t('global.value')" :max-length="500" />
                <div class="tooltip" :data-tip="$t('global.delete')">
                  <button class="btn btn-square btn-sm mb-2 ml-2" @click="item.fields.splice(idx, 1)">
                    <MdiDelete />
                  </button>
                </div>
              </div>
            </div>
          </div>
          <div class="mt-4 flex justify-end px-5 pb-4">
            <BaseButton size="sm" @click="addField"> {{ $t("global.add") }} </BaseButton>
          </div>
        </BaseCard>

        <div
          v-if="preferences.editorAdvancedView"
          ref="attDropZone"
          class="card overflow-visible bg-base-100 shadow-xl sm:rounded-lg"
        >
          <div class="px-4 py-5 sm:px-6">
            <h3 class="text-lg font-medium leading-6">{{ $t("items.attachments") }}</h3>
            <p class="text-xs">{{ $t("items.changes_persisted_immediately") }}</p>
          </div>
          <div class="border-t border-gray-300 p-4">
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

          <div class="border-t border-gray-300 p-4">
            <ul role="list" class="divide-y divide-gray-400 rounded-md border border-gray-400">
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
                  <div class="tooltip" :data-tip="$t('global.delete')">
                    <button class="btn btn-square btn-sm" @click="deleteAttachment(attachment.id)">
                      <MdiDelete />
                    </button>
                  </div>
                  <div class="tooltip" :data-tip="$t('global.edit')">
                    <button class="btn btn-square btn-sm" @click="openAttachmentEditDialog(attachment)">
                      <MdiPencil />
                    </button>
                  </div>
                </div>
              </li>
            </ul>
          </div>
        </div>

        <div v-if="preferences.editorAdvancedView" class="card overflow-visible bg-base-100 shadow-xl sm:rounded-lg">
          <div class="px-4 py-5 sm:px-6">
            <h3 class="text-lg font-medium leading-6">{{ $t("items.purchase_details") }}</h3>
          </div>
          <div class="border-t border-gray-300 sm:p-0">
            <div
              v-for="field in purchaseFields"
              :key="field.ref"
              class="grid grid-cols-1 sm:divide-y sm:divide-gray-300"
            >
              <div class="border-b border-gray-300 px-4 pb-4 pt-2 sm:px-6">
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
        </div>

        <div v-if="preferences.editorAdvancedView" class="card overflow-visible bg-base-100 shadow-xl sm:rounded-lg">
          <div class="px-4 py-5 sm:px-6">
            <h3 class="text-lg font-medium leading-6">{{ $t("items.warranty_details") }}</h3>
          </div>
          <div class="border-t border-gray-300 sm:p-0">
            <div
              v-for="field in warrantyFields"
              :key="field.ref"
              class="grid grid-cols-1 sm:divide-y sm:divide-gray-300"
            >
              <div class="border-b border-gray-300 px-4 pb-4 pt-2 sm:px-6">
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
        </div>

        <div v-if="preferences.editorAdvancedView" class="card overflow-visible bg-base-100 shadow-xl sm:rounded-lg">
          <div class="px-4 py-5 sm:px-6">
            <h3 class="text-lg font-medium leading-6">Sold Details</h3>
          </div>
          <div class="border-t border-gray-300 sm:p-0">
            <div v-for="field in soldFields" :key="field.ref" class="grid grid-cols-1 sm:divide-y sm:divide-gray-300">
              <div class="border-b border-gray-300 px-4 pb-4 pt-2 sm:px-6">
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
        </div>
      </div>
    </section>
  </div>
</template>
