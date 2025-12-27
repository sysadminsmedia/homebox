<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import MdiPencil from "~icons/mdi/pencil";
  import MdiDelete from "~icons/mdi/delete";
  import MdiContentCopy from "~icons/mdi/content-copy";
  import MdiStar from "~icons/mdi/star";
  import MdiStarOutline from "~icons/mdi/star-outline";
  import { Card, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
  import { Button } from "@/components/ui/button";
  import { Badge } from "@/components/ui/badge";
  import type { LabelTemplateSummary } from "~/lib/api/types/data-contracts";

  const props = defineProps<{
    template: LabelTemplateSummary;
  }>();

  const emit = defineEmits<{
    deleted: [];
    duplicated: [id: string];
  }>();

  const { duplicate, remove } = useLabelTemplateActions();
  const confirm = useConfirm();
  const { t } = useI18n();
  const preferences = useViewPreferences();

  const isDefault = computed(() => preferences.value.defaultTemplateId === props.template.id);

  function toggleDefault() {
    if (isDefault.value) {
      preferences.value.defaultTemplateId = null;
      toast.success(t("components.label_template.toast.default_cleared"));
    } else {
      preferences.value.defaultTemplateId = props.template.id;
      toast.success(t("components.label_template.toast.default_set", { name: props.template.name }));
    }
  }

  const sizeLabel = computed(() => {
    return `${props.template.width}mm x ${props.template.height}mm`;
  });

  async function handleDelete() {
    const { isCanceled } = await confirm.open(t("components.label_template.confirm_delete"));
    if (isCanceled) return;

    try {
      await remove(props.template.id);
      toast.success(t("components.label_template.toast.deleted"));
      emit("deleted");
    } catch {
      toast.error(t("components.label_template.toast.delete_failed"));
    }
  }

  async function handleDuplicate() {
    try {
      const newTemplate = await duplicate(props.template.id);
      if (newTemplate) {
        toast.success(t("components.label_template.toast.duplicated", { name: newTemplate.name }));
        emit("duplicated", newTemplate.id);
      }
    } catch {
      toast.error(t("components.label_template.toast.duplicate_failed"));
    }
  }
</script>

<template>
  <Card>
    <CardHeader>
      <div class="flex items-start justify-between gap-2">
        <CardTitle class="truncate">{{ template.name }}</CardTitle>
        <div class="flex gap-1">
          <Badge v-if="isDefault" variant="default" class="bg-amber-500 hover:bg-amber-500">
            {{ $t("components.label_template.default") }}
          </Badge>
          <Badge v-if="template.isShared" variant="secondary">
            {{ $t("components.label_template.shared") }}
          </Badge>
          <Badge v-if="!template.isOwner" variant="outline">
            {{ $t("components.label_template.not_owner") }}
          </Badge>
        </div>
      </div>
      <CardDescription>
        <div class="text-xs text-muted-foreground">{{ sizeLabel }}</div>
        <div v-if="template.preset" class="text-xs text-muted-foreground">{{ template.preset }}</div>
        <div v-if="template.description" class="mt-1 line-clamp-2">
          {{ template.description }}
        </div>
      </CardDescription>
    </CardHeader>
    <CardFooter class="flex justify-end gap-1">
      <Button
        size="icon"
        :variant="isDefault ? 'default' : 'outline'"
        :class="['size-8', isDefault ? 'bg-amber-500 hover:bg-amber-600' : '']"
        :title="
          isDefault
            ? $t('components.label_template.card.clear_default')
            : $t('components.label_template.card.set_default')
        "
        @click="toggleDefault"
      >
        <MdiStar v-if="isDefault" class="size-4" />
        <MdiStarOutline v-else class="size-4" />
      </Button>
      <Button
        v-if="template.isOwner"
        size="icon"
        variant="outline"
        class="size-8"
        as-child
        :title="$t('components.label_template.card.edit')"
      >
        <NuxtLink :to="`/label-templates/${template.id}/edit`">
          <MdiPencil class="size-4" />
        </NuxtLink>
      </Button>
      <Button
        size="icon"
        variant="outline"
        class="size-8"
        :title="$t('components.label_template.card.duplicate')"
        @click="handleDuplicate"
      >
        <MdiContentCopy class="size-4" />
      </Button>
      <Button
        v-if="template.isOwner"
        size="icon"
        variant="destructive"
        class="size-8"
        :title="$t('components.label_template.card.delete')"
        @click="handleDelete"
      >
        <MdiDelete class="size-4" />
      </Button>
    </CardFooter>
  </Card>
</template>
