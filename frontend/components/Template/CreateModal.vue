<template>
  <BaseModal :dialog-id="DialogID.CreateTemplate" :title="$t('components.template.create_modal.title')">
    <form class="flex min-w-0 flex-col gap-2" @submit.prevent="create()">
      <FormTextField
        v-model="form.name"
        :autofocus="true"
        :label="$t('components.template.form.template_name')"
        :max-length="255"
        :min-length="1"
      />
      <FormTextArea
        v-model="form.description"
        :label="$t('components.template.form.template_description')"
        :max-length="1000"
      />

      <Separator class="my-2" />
      <h3 class="text-sm font-medium">{{ $t("components.template.form.default_item_values") }}</h3>
      <div class="flex min-w-0 flex-col gap-2">
        <FormTextField v-model="form.defaultName" :label="$t('components.template.form.item_name')" :max-length="255" />
        <FormTextArea
          v-model="form.defaultDescription"
          :label="$t('components.template.form.item_description')"
          :max-length="1000"
        />
        <div class="grid grid-cols-2 gap-2">
          <FormTextField
            v-model.number="form.defaultQuantity"
            :label="$t('global.quantity')"
            type="number"
            :min="1"
            step="any"
          />
          <FormTextField
            v-model="form.defaultModelNumber"
            :label="$t('components.template.form.model_number')"
            :max-length="255"
          />
        </div>
        <FormTextField
          v-model="form.defaultManufacturer"
          :label="$t('components.template.form.manufacturer')"
          :max-length="255"
        />
        <LocationSelector
          v-model="form.defaultLocationObject"
          :label="$t('components.template.form.default_location')"
        />
        <TagSelector v-model="form.defaultTagIds" :tags="tags ?? []" />
        <div class="flex items-center gap-4">
          <div class="flex items-center gap-2">
            <Switch id="defaultInsured" v-model:checked="form.defaultInsured" />
            <Label for="defaultInsured" class="text-sm">{{ $t("global.insured") }}</Label>
          </div>
          <div class="flex items-center gap-2">
            <Switch id="defaultLifetimeWarranty" v-model:checked="form.defaultLifetimeWarranty" />
            <Label for="defaultLifetimeWarranty" class="text-sm">{{
              $t("components.template.form.lifetime_warranty")
            }}</Label>
          </div>
        </div>
      </div>

      <Separator class="my-2" />
      <div class="flex flex-col gap-2">
        <h3 class="text-sm font-medium">{{ $t("components.template.form.default_image") }}</h3>
        <p class="text-xs text-muted-foreground">{{ $t("components.template.form.default_image_hint") }}</p>
        <div class="flex items-center gap-3">
          <img
            v-if="imagePreview"
            :src="imagePreview"
            :alt="$t('components.template.form.default_image')"
            class="size-16 rounded border object-cover"
          />
          <Button type="button" size="sm" variant="outline" @click="imageInput?.click()">
            {{ image ? $t("components.template.form.replace_image") : $t("components.template.form.upload_image") }}
          </Button>
          <Button v-if="image" type="button" size="sm" variant="ghost" @click="clearImage">
            {{ $t("components.template.form.remove_image") }}
          </Button>
          <input
            ref="imageInput"
            type="file"
            class="hidden"
            accept="image/png,image/jpeg,image/gif,image/avif,image/webp"
            @change="onImageSelected"
          />
        </div>
      </div>

      <Separator class="my-2" />
      <div class="flex items-center justify-between">
        <h3 class="text-sm font-medium">{{ $t("components.template.form.custom_fields") }}</h3>
        <Button type="button" size="sm" variant="outline" @click="addField">
          <MdiPlus class="mr-1 size-4" />
          {{ $t("global.add") }}
        </Button>
      </div>
      <div v-if="form.fields.length > 0" class="flex flex-col gap-2">
        <div v-for="(field, idx) in form.fields" :key="idx" class="flex items-end gap-2">
          <FormTextField
            v-model="field.name"
            :label="$t('components.template.form.field_name')"
            :max-length="255"
            class="flex-1"
          />
          <FormTextField
            v-model="field.textValue"
            :label="$t('components.template.form.default_value')"
            class="flex-1"
          />
          <Button type="button" size="icon" variant="ghost" @click="form.fields.splice(idx, 1)">
            <MdiDelete class="size-4" />
          </Button>
        </div>
      </div>
      <p v-else class="text-sm text-muted-foreground">{{ $t("components.template.form.no_custom_fields") }}</p>

      <div class="mt-4 flex justify-end">
        <Button type="submit" :loading="loading">{{ $t("global.create") }}</Button>
      </div>
    </form>
  </BaseModal>
</template>

<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import MdiPlus from "~icons/mdi/plus";
  import MdiDelete from "~icons/mdi/delete";
  import { DialogID } from "@/components/ui/dialog-provider/utils";
  import BaseModal from "@/components/App/CreateModal.vue";
  import { useDialog } from "~/components/ui/dialog-provider";
  import FormTextField from "~/components/Form/TextField.vue";
  import FormTextArea from "~/components/Form/TextArea.vue";
  import { Button } from "~/components/ui/button";
  import { Separator } from "@/components/ui/separator";
  import { Switch } from "@/components/ui/switch";
  import { Label } from "@/components/ui/label";
  import LocationSelector from "~/components/Location/Selector.vue";
  import TagSelector from "~/components/Tag/Selector.vue";
  import { useTagStore } from "~/stores/tags";
  import type { EntitySummary } from "~~/lib/api/types/data-contracts";

  const emit = defineEmits<{ created: [] }>();
  const { closeDialog } = useDialog();

  const tagStore = useTagStore();
  const tags = computed(() => tagStore.tags);

  const loading = ref(false);
  const form = reactive({
    name: "",
    description: "",
    notes: "",
    defaultName: "",
    defaultDescription: "",
    defaultQuantity: 1,
    defaultInsured: false,
    defaultManufacturer: "",
    defaultModelNumber: "",
    defaultLifetimeWarranty: false,
    defaultWarrantyDetails: "",
    defaultLocationId: null as string | null,
    defaultLocationObject: null as EntitySummary | null,
    defaultTagIds: [] as string[],
    includeWarrantyFields: false,
    includePurchaseFields: false,
    includeSoldFields: false,
    fields: [] as Array<{ id: string; name: string; type: "text"; textValue: string }>,
  });

  const NIL_UUID = "00000000-0000-0000-0000-000000000000";

  const image = ref<File | null>(null);
  const imagePreview = ref<string | null>(null);
  const imageInput = ref<HTMLInputElement | null>(null);

  function onImageSelected(event: Event) {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    input.value = "";
    if (!file) return;

    clearImage();
    image.value = file;
    imagePreview.value = URL.createObjectURL(file);
  }

  function clearImage() {
    if (imagePreview.value) {
      URL.revokeObjectURL(imagePreview.value);
    }
    image.value = null;
    imagePreview.value = null;
  }

  onBeforeUnmount(clearImage);

  function addField() {
    form.fields.push({ id: NIL_UUID, name: "", type: "text", textValue: "" });
  }

  function reset() {
    Object.assign(form, {
      name: "",
      description: "",
      notes: "",
      defaultName: "",
      defaultDescription: "",
      defaultQuantity: 1,
      defaultInsured: false,
      defaultManufacturer: "",
      defaultModelNumber: "",
      defaultLifetimeWarranty: false,
      defaultWarrantyDetails: "",
      defaultLocationId: null,
      defaultLocationObject: null,
      defaultTagIds: [],
      includeWarrantyFields: false,
      includePurchaseFields: false,
      includeSoldFields: false,
      fields: [],
    });
    clearImage();
    loading.value = false;
  }

  const api = useUserApi();

  const { t } = useI18n();

  async function create() {
    if (loading.value) return;
    loading.value = true;

    // Prepare the data with proper format for API
    const createData = {
      name: form.name,
      description: form.description,
      notes: form.notes,
      defaultName: form.defaultName || null,
      defaultDescription: form.defaultDescription || null,
      defaultQuantity: form.defaultQuantity,
      defaultInsured: form.defaultInsured,
      defaultManufacturer: form.defaultManufacturer || null,
      defaultModelNumber: form.defaultModelNumber || null,
      defaultLifetimeWarranty: form.defaultLifetimeWarranty,
      defaultWarrantyDetails: form.defaultWarrantyDetails || null,
      defaultLocationId: form.defaultLocationObject?.id ?? null,
      defaultTagIds: form.defaultTagIds,
      includeWarrantyFields: form.includeWarrantyFields,
      includePurchaseFields: form.includePurchaseFields,
      includeSoldFields: form.includeSoldFields,
      fields: form.fields,
    };

    const { error, data } = await api.templates.create(createData);
    if (error) {
      toast.error(t("components.template.toast.create_failed"));
      loading.value = false;
      return;
    }

    // The image needs a template to belong to, so it follows the create call.
    // A failure here leaves a usable template, so it warns rather than aborts.
    if (image.value && data) {
      const { error: imageError } = await api.templates.setImage(data.id, image.value, image.value.name);
      if (imageError) {
        toast.error(t("components.template.toast.image_update_failed"));
      }
    }

    toast.success(t("components.template.toast.created"));
    reset();
    closeDialog(DialogID.CreateTemplate);
    emit("created");
  }
</script>
