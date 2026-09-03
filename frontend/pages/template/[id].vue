<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import MdiPencil from "~icons/mdi/pencil";
  import MdiDelete from "~icons/mdi/delete";
  import MdiPlus from "~icons/mdi/plus";
  import MdiImagePlus from "~icons/mdi/image-plus";
  import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { Card } from "@/components/ui/card";
  import { Button } from "@/components/ui/button";
  import { Separator } from "@/components/ui/separator";
  import { Switch } from "@/components/ui/switch";
  import { Label } from "@/components/ui/label";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import FormTextField from "~/components/Form/TextField.vue";
  import FormTextArea from "~/components/Form/TextArea.vue";
  import BaseContainer from "@/components/Base/Container.vue";
  import DateTime from "~/components/global/DateTime.vue";
  import Markdown from "~/components/global/Markdown.vue";
  import LocationSelector from "~/components/Location/Selector.vue";
  import TagSelector from "~/components/Tag/Selector.vue";
  import { useTagStore } from "~/stores/tags";
  import type { EntityOut } from "~~/lib/api/types/data-contracts";

  definePageMeta({
    middleware: ["auth"],
  });

  const { openDialog, closeDialog } = useDialog();
  const route = useRoute();
  const api = useUserApi();
  const confirm = useConfirm();

  const tagStore = useTagStore();
  const tags = computed(() => tagStore.tags);

  const templateId = computed<string>(() => route.params.id as string);

  const { t } = useI18n();

  const { data: template, refresh } = useAsyncData(templateId.value, async () => {
    const { data, error } = await api.templates.get(templateId.value);
    if (error) {
      toast.error(t("components.template.toast.load_failed"));
      navigateTo("/templates");
      return;
    }
    return data;
  });

  async function confirmDelete() {
    const { isCanceled } = await confirm.open(t("components.template.confirm_delete"));
    if (isCanceled) return;

    const { error } = await api.templates.delete(templateId.value);
    if (error) {
      toast.error(t("components.template.toast.delete_failed"));
      return;
    }
    toast.success(t("components.template.toast.deleted"));
    navigateTo("/templates");
  }

  const updating = ref(false);
  const updateData = reactive({
    id: "",
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
    defaultLocation: null as EntityOut | null,
    defaultTagIds: [] as string[],
    includeWarrantyFields: false,
    includePurchaseFields: false,
    includeSoldFields: false,
    fields: [] as Array<{ id: string; name: string; type: "text"; textValue: string }>,
  });

  function openUpdate() {
    if (!template.value) return;
    Object.assign(updateData, {
      id: template.value.id,
      name: template.value.name,
      description: template.value.description,
      notes: template.value.notes,
      defaultName: template.value.defaultName ?? "",
      defaultDescription: template.value.defaultDescription ?? "",
      defaultQuantity: template.value.defaultQuantity,
      defaultInsured: template.value.defaultInsured,
      defaultManufacturer: template.value.defaultManufacturer,
      defaultModelNumber: template.value.defaultModelNumber ?? "",
      defaultLifetimeWarranty: template.value.defaultLifetimeWarranty,
      defaultWarrantyDetails: template.value.defaultWarrantyDetails,
      defaultLocation: template.value.defaultLocation ?? null,
      defaultTagIds: template.value.defaultTags?.map(l => l.id) ?? [],
      includeWarrantyFields: template.value.includeWarrantyFields,
      includePurchaseFields: template.value.includePurchaseFields,
      includeSoldFields: template.value.includeSoldFields,
      fields: template.value.fields.map(f => ({
        id: f.id,
        name: f.name,
        type: "text" as const,
        textValue: f.textValue,
      })),
    });
    openDialog(DialogID.UpdateTemplate);
  }

  async function update() {
    updating.value = true;

    // Prepare the data with proper format for API
    const payload = {
      ...updateData,
      defaultLocationId: updateData.defaultLocation?.id ?? null,
    };

    const { error, data } = await api.templates.update(templateId.value, payload);
    if (error) {
      updating.value = false;
      toast.error(t("components.template.toast.update_failed"));
      return;
    }
    toast.success(t("components.template.toast.updated"));
    template.value = data;
    closeDialog(DialogID.UpdateTemplate);
    updating.value = false;
    refresh();
  }

  const NIL_UUID = "00000000-0000-0000-0000-000000000000";

  const imageUploading = ref(false);
  const imageInput = ref<HTMLInputElement | null>(null);

  const hasImage = computed(() => !!template.value?.defaultImage);

  const imageUrl = computed(() => {
    const image = template.value?.defaultImage;
    if (!image) return null;
    // Cache-bust on the attachment id so replacing the image repaints.
    // authURL only adds a query string when an attachment token is in play.
    const url = api.authURL(`/templates/${templateId.value}/image`);
    return `${url}${url.includes("?") ? "&" : "?"}v=${image.id}`;
  });

  async function onImageSelected(event: Event) {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    input.value = "";
    if (!file) return;

    imageUploading.value = true;
    const { error, data } = await api.templates.setImage(templateId.value, file, file.name);
    imageUploading.value = false;

    if (error) {
      toast.error(t("components.template.toast.image_update_failed"));
      return;
    }
    toast.success(t("components.template.toast.image_updated"));
    template.value = data;
  }

  async function removeImage() {
    const { isCanceled } = await confirm.open(t("components.template.confirm_remove_image"));
    if (isCanceled) return;

    const { error, data } = await api.templates.deleteImage(templateId.value);
    if (error) {
      toast.error(t("components.template.toast.image_remove_failed"));
      return;
    }
    toast.success(t("components.template.toast.image_removed"));
    template.value = data;
  }
</script>

<template>
  <Dialog :dialog-id="DialogID.UpdateTemplate">
    <DialogContent class="max-h-[90vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>{{ $t("components.template.edit_modal.title") }}</DialogTitle>
      </DialogHeader>

      <form v-if="template" class="flex flex-col gap-2" @submit.prevent="update">
        <FormTextField
          v-model="updateData.name"
          :autofocus="true"
          :label="$t('components.template.form.template_name')"
          :max-length="255"
        />
        <FormTextArea
          v-model="updateData.description"
          :label="$t('components.template.form.template_description')"
          :max-length="1000"
        />

        <Separator class="my-2" />
        <h3 class="text-sm font-medium">{{ $t("components.template.form.default_item_values") }}</h3>
        <div class="grid gap-2">
          <FormTextField
            v-model="updateData.defaultName"
            :label="$t('components.template.form.item_name')"
            :max-length="255"
          />
          <FormTextArea
            v-model="updateData.defaultDescription"
            :label="$t('components.template.form.item_description')"
            :max-length="1000"
          />
          <div class="grid grid-cols-2 gap-2">
            <FormTextField
              v-model.number="updateData.defaultQuantity"
              :label="$t('global.quantity')"
              type="number"
              :min="1"
              step="any"
            />
            <FormTextField
              v-model="updateData.defaultModelNumber"
              :label="$t('components.template.form.model_number')"
              :max-length="255"
            />
          </div>
          <FormTextField
            v-model="updateData.defaultManufacturer"
            :label="$t('components.template.form.manufacturer')"
            :max-length="255"
          />
          <LocationSelector
            v-model="updateData.defaultLocation"
            :label="$t('components.template.form.default_location')"
          />
          <TagSelector v-model="updateData.defaultTagIds" :tags="tags ?? []" />
          <div class="flex items-center gap-4">
            <div class="flex items-center gap-2">
              <Switch id="editInsured" v-model:checked="updateData.defaultInsured" />
              <Label for="editInsured" class="text-sm">{{ $t("global.insured") }}</Label>
            </div>
            <div class="flex items-center gap-2">
              <Switch id="editWarranty" v-model:checked="updateData.defaultLifetimeWarranty" />
              <Label for="editWarranty" class="text-sm">{{ $t("components.template.form.lifetime_warranty") }}</Label>
            </div>
          </div>
        </div>

        <Separator class="my-2" />
        <div class="flex items-center justify-between">
          <h3 class="text-sm font-medium">{{ $t("components.template.form.custom_fields") }}</h3>
          <Button
            type="button"
            size="sm"
            variant="outline"
            @click="updateData.fields.push({ id: NIL_UUID, name: '', type: 'text', textValue: '' })"
          >
            <MdiPlus class="mr-1 size-4" />
            {{ $t("global.add") }}
          </Button>
        </div>
        <div v-if="updateData.fields.length > 0" class="flex flex-col gap-2">
          <div v-for="(field, idx) in updateData.fields" :key="idx" class="flex items-end gap-2">
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
            <Button type="button" size="icon" variant="ghost" @click="updateData.fields.splice(idx, 1)">
              <MdiDelete class="size-4" />
            </Button>
          </div>
        </div>

        <DialogFooter>
          <Button type="submit" :loading="updating">{{ $t("global.update") }}</Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>

  <BaseContainer v-if="template">
    <Title>{{ template.name }}</Title>

    <Card class="p-3">
      <header :class="{ 'mb-2': template.description }">
        <div class="flex flex-wrap items-end gap-2">
          <div>
            <h1 class="pb-1 text-2xl">{{ template.name }}</h1>
            <div class="flex flex-wrap gap-1 text-xs text-muted-foreground">
              <span>{{ $t("global.created") }} <DateTime :date="template.createdAt" /></span>
              <span>•</span>
              <span>{{ $t("components.template.detail.updated") }} <DateTime :date="template.updatedAt" /></span>
            </div>
          </div>
          <div class="ml-auto flex gap-2">
            <Button @click="openUpdate">
              <MdiPencil class="mr-1" />
              {{ $t("global.edit") }}
            </Button>
            <Button variant="destructive" @click="confirmDelete">
              <MdiDelete class="mr-1" />
              {{ $t("global.delete") }}
            </Button>
          </div>
        </div>
      </header>

      <Separator v-if="template.description" class="my-3" />
      <Markdown v-if="template.description" :source="template.description" />

      <Separator class="my-3" />
      <div class="flex flex-wrap items-start gap-4">
        <div>
          <h3 class="mb-2 text-sm font-medium">{{ $t("components.template.detail.default_image") }}</h3>
          <img
            v-if="imageUrl"
            :src="imageUrl"
            :alt="$t('components.template.detail.default_image')"
            class="size-32 rounded border object-cover"
          />
          <p v-else class="text-sm text-muted-foreground">
            {{ $t("components.template.form.no_default_image") }}
          </p>
        </div>
        <div class="flex flex-col gap-2 pt-7">
          <p class="max-w-xs text-xs text-muted-foreground">
            {{ $t("components.template.form.default_image_hint") }}
          </p>
          <div class="flex gap-2">
            <Button type="button" size="sm" variant="outline" :loading="imageUploading" @click="imageInput?.click()">
              <MdiImagePlus class="mr-1 size-4" />
              {{
                hasImage ? $t("components.template.form.replace_image") : $t("components.template.form.upload_image")
              }}
            </Button>
            <Button v-if="hasImage" type="button" size="sm" variant="ghost" @click="removeImage">
              <MdiDelete class="mr-1 size-4" />
              {{ $t("components.template.form.remove_image") }}
            </Button>
          </div>
          <input
            ref="imageInput"
            type="file"
            class="hidden"
            accept="image/png,image/jpeg,image/gif,image/avif,image/webp"
            @change="onImageSelected"
          />
        </div>
      </div>

      <Separator class="my-3" />
      <div class="grid gap-4 text-sm md:grid-cols-2">
        <div>
          <h3 class="mb-2 font-medium">{{ $t("components.template.detail.default_values") }}</h3>
          <dl class="flex flex-col gap-1">
            <div v-if="template.defaultName" class="flex justify-between">
              <dt class="text-muted-foreground">{{ $t("components.template.form.item_name") }}</dt>
              <dd>{{ template.defaultName }}</dd>
            </div>
            <div v-if="template.defaultDescription" class="flex justify-between">
              <dt class="text-muted-foreground">{{ $t("components.template.form.item_description") }}</dt>
              <dd class="max-w-[200px] truncate">{{ template.defaultDescription }}</dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-muted-foreground">{{ $t("global.quantity") }}</dt>
              <dd>{{ template.defaultQuantity }}</dd>
            </div>
            <div v-if="template.defaultModelNumber" class="flex justify-between">
              <dt class="text-muted-foreground">{{ $t("components.template.form.model_number") }}</dt>
              <dd>{{ template.defaultModelNumber }}</dd>
            </div>
            <div v-if="template.defaultManufacturer" class="flex justify-between">
              <dt class="text-muted-foreground">{{ $t("components.template.form.manufacturer") }}</dt>
              <dd>{{ template.defaultManufacturer }}</dd>
            </div>
            <div v-if="template.defaultLocation" class="flex justify-between">
              <dt class="text-muted-foreground">{{ $t("components.template.form.location") }}</dt>
              <dd>{{ template.defaultLocation.name }}</dd>
            </div>
            <div v-if="template.defaultTags && template.defaultTags.length > 0" class="flex justify-between">
              <dt class="text-muted-foreground">{{ $t("global.tags") }}</dt>
              <dd>{{ template.defaultTags.map(t => t.name).join(", ") }}</dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-muted-foreground">{{ $t("global.insured") }}</dt>
              <dd>{{ template.defaultInsured ? $t("global.yes") : $t("global.no") }}</dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-muted-foreground">{{ $t("components.template.form.lifetime_warranty") }}</dt>
              <dd>{{ template.defaultLifetimeWarranty ? $t("global.yes") : $t("global.no") }}</dd>
            </div>
          </dl>
        </div>
        <div v-if="template.fields.length > 0">
          <h3 class="mb-2 font-medium">{{ $t("components.template.form.custom_fields") }}</h3>
          <dl class="flex flex-col gap-1">
            <div v-for="field in template.fields" :key="field.id" class="flex justify-between">
              <dt class="text-muted-foreground">{{ field.name }}</dt>
              <dd>{{ field.textValue || "—" }}</dd>
            </div>
          </dl>
        </div>
      </div>
    </Card>
  </BaseContainer>
</template>
