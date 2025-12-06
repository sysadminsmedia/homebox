<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import MdiPencil from "~icons/mdi/pencil";
  import MdiDelete from "~icons/mdi/delete";
  import MdiContentCopy from "~icons/mdi/content-copy";
  import { Card, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
  import { Button } from "@/components/ui/button";
  import type { ItemTemplateSummary, ItemTemplateCreate } from "~/lib/api/types/data-contracts";

  const props = defineProps<{
    template: ItemTemplateSummary;
  }>();

  const emit = defineEmits<{
    deleted: [];
    duplicated: [id: string];
  }>();

  const api = useUserApi();
  const confirm = useConfirm();
  const { t } = useI18n();

  async function handleDelete() {
    const { isCanceled } = await confirm.open(t("components.template.confirm_delete"));
    if (isCanceled) return;

    const { error } = await api.templates.delete(props.template.id);
    if (error) {
      toast.error(t("components.template.toast.delete_failed"));
      return;
    }

    toast.success(t("components.template.toast.deleted"));
    emit("deleted");
  }

  async function handleDuplicate() {
    // First, get the full template details
    const { data: fullTemplate, error: getError } = await api.templates.get(props.template.id);
    if (getError || !fullTemplate) {
      toast.error(t("components.template.toast.load_failed"));
      return;
    }

    const NIL_UUID = "00000000-0000-0000-0000-000000000000";

    // Create a duplicate with "(Copy)" suffix
    const duplicateData: ItemTemplateCreate = {
      name: `${fullTemplate.name} (Copy)`,
      description: fullTemplate.description,
      notes: fullTemplate.notes,
      defaultName: fullTemplate.defaultName,
      defaultDescription: fullTemplate.defaultDescription,
      defaultQuantity: fullTemplate.defaultQuantity,
      defaultInsured: fullTemplate.defaultInsured,
      defaultManufacturer: fullTemplate.defaultManufacturer,
      defaultModelNumber: fullTemplate.defaultModelNumber,
      defaultLifetimeWarranty: fullTemplate.defaultLifetimeWarranty,
      defaultWarrantyDetails: fullTemplate.defaultWarrantyDetails,
      defaultLocationId: fullTemplate.defaultLocation?.id ?? "",
      defaultLabelIds: fullTemplate.defaultLabels?.map(l => l.id) || [],
      includeWarrantyFields: fullTemplate.includeWarrantyFields,
      includePurchaseFields: fullTemplate.includePurchaseFields,
      includeSoldFields: fullTemplate.includeSoldFields,
      fields: fullTemplate.fields.map(field => ({
        id: NIL_UUID,
        name: field.name,
        type: field.type,
        textValue: field.textValue,
      })),
    };

    const { data, error } = await api.templates.create(duplicateData);
    if (error) {
      toast.error(t("components.template.toast.duplicate_failed"));
      return;
    }

    toast.success(t("components.template.toast.duplicated", { name: duplicateData.name }));
    emit("duplicated", data.id);
  }
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle class="truncate">{{ template.name }}</CardTitle>
      <CardDescription v-if="template.description" class="line-clamp-2">
        {{ template.description }}
      </CardDescription>
    </CardHeader>
    <CardFooter class="flex justify-end gap-1">
      <Button size="icon" variant="outline" class="size-8" as-child :title="$t('components.template.card.edit')">
        <NuxtLink :to="`/template/${template.id}`">
          <MdiPencil class="size-4" />
        </NuxtLink>
      </Button>
      <Button
        size="icon"
        variant="outline"
        class="size-8"
        :title="$t('components.template.card.duplicate')"
        @click="handleDuplicate"
      >
        <MdiContentCopy class="size-4" />
      </Button>
      <Button
        size="icon"
        variant="destructive"
        class="size-8"
        :title="$t('components.template.card.delete')"
        @click="handleDelete"
      >
        <MdiDelete class="size-4" />
      </Button>
    </CardFooter>
  </Card>
</template>
