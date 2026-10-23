<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import MdiDotsHorizontal from "~icons/mdi/dots-horizontal";
  import MdiStar from "~icons/mdi/star";
  import { Card } from "@/components/ui/card";
  import { Button } from "@/components/ui/button";
  import { Badge } from "@/components/ui/badge";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
  } from "@/components/ui/dropdown-menu";
  import DateTime from "~/components/global/DateTime.vue";
  import type { LabelTemplateSummary } from "~/lib/api/types/data-contracts";

  const props = defineProps<{
    template: LabelTemplateSummary;
  }>();

  const emit = defineEmits<{
    deleted: [];
    duplicated: [id: string];
  }>();

  const api = useUserApi();
  const { duplicate, remove } = useLabelTemplateActions();
  const { find: findPreset, title: presetTitle } = useLabelPresetLookup();
  const confirm = useConfirm();
  const { t } = useI18n();
  const preferences = useViewPreferences();

  const isDefault = computed(() => preferences.value.defaultTemplateId === props.template.id);
  const preset = computed(() => findPreset(props.template.preset));
  const previewUrl = computed(() => api.labelTemplates.getPreviewUrl(props.template.id));
  const sizeLabel = computed(() => formatLabelSize(props.template.width, props.template.height));

  const presetKind = computed(() => {
    const p = preset.value;
    if (!p) return "";
    if (p.kind === "sheet") return t("components.label_template.card.sheet_up", { count: p.perSheet ?? 0 });
    if (p.kind === "continuous") return t("components.label_template.card.continuous");
    return t("components.label_template.card.die_cut");
  });

  // Thumbnail aspect ratio follows the label so a 90x29 strip and a 62x29 tag read differently.
  const thumbStyle = computed(() => {
    const ratio = props.template.width / Math.max(props.template.height, 1);
    // Clamp so very wide strips still fit the card.
    const clamped = Math.min(Math.max(ratio, 1), 3.2);
    return { aspectRatio: String(clamped) };
  });

  function toggleDefault() {
    if (isDefault.value) {
      preferences.value.defaultTemplateId = null;
      toast.success(t("components.label_template.toast.default_cleared"));
    } else {
      preferences.value.defaultTemplateId = props.template.id;
      toast.success(t("components.label_template.toast.default_set", { name: props.template.name }));
    }
  }

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

  async function handleExport() {
    const { data, error } = await api.labelTemplates.get(props.template.id);
    if (error || !data) {
      toast.error(t("components.label_template.toast.export_failed"));
      return;
    }
    const payload = {
      name: data.name,
      description: data.description,
      width: data.width,
      height: data.height,
      preset: data.preset,
      dpi: data.dpi,
      outputFormat: data.outputFormat,
      canvasData: data.canvasData,
    };
    const blob = new Blob([JSON.stringify(payload, null, 2)], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `${data.name.replace(/[^a-z0-9-_]+/gi, "_").toLowerCase() || "label-template"}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }
</script>

<template>
  <Card class="flex flex-col overflow-hidden">
    <!-- Rendered sample of the label -->
    <NuxtLink
      :to="template.isOwner ? `/label-templates/${template.id}/edit` : undefined"
      class="flex h-36 items-center justify-center border-b bg-muted/60 p-4"
    >
      <img
        :src="previewUrl"
        :alt="$t('components.label_template.card.thumbnail_alt', { name: template.name })"
        class="max-h-full rounded-sm bg-white shadow"
        :style="thumbStyle"
        loading="lazy"
      />
    </NuxtLink>

    <div class="flex flex-1 flex-col gap-2 px-4 py-3">
      <div class="flex items-start justify-between gap-2">
        <h3 class="truncate text-base font-semibold">{{ template.name }}</h3>
        <Badge v-if="isDefault" class="shrink-0 gap-1 bg-amber-100 text-amber-900 hover:bg-amber-100">
          <MdiStar class="size-3" />
          {{ $t("components.label_template.default") }}
        </Badge>
      </div>

      <div class="flex flex-wrap gap-1.5 text-xs">
        <Badge v-if="preset" variant="outline" class="font-normal">
          {{ presetTitle(template.preset) }} · {{ presetKind }}
        </Badge>
        <Badge variant="outline" class="font-normal">{{ sizeLabel }}</Badge>
        <Badge v-if="template.isShared" variant="secondary" class="font-normal">
          {{ $t("components.label_template.shared") }}
        </Badge>
        <Badge v-if="!template.isOwner" variant="secondary" class="font-normal">
          {{ $t("components.label_template.not_owner") }}
        </Badge>
      </div>

      <p v-if="template.description" class="line-clamp-2 text-sm text-muted-foreground">
        {{ template.description }}
      </p>
    </div>

    <div class="flex items-center justify-between gap-2 border-t px-3 py-2">
      <span class="text-xs text-muted-foreground">
        {{ $t("components.label_template.card.edited") }}
        <DateTime :date="template.updatedAt" format="relative" datetime-type="date" />
      </span>
      <div class="flex gap-1.5">
        <Button v-if="template.isOwner" size="sm" variant="outline" as-child>
          <NuxtLink :to="`/label-templates/${template.id}/edit`">{{ $t("global.edit") }}</NuxtLink>
        </Button>
        <DropdownMenu>
          <DropdownMenuTrigger as-child>
            <Button
              size="icon"
              variant="outline"
              class="size-8"
              :aria-label="$t('components.label_template.card.more_actions', { name: template.name })"
            >
              <MdiDotsHorizontal class="size-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" class="w-52">
            <DropdownMenuItem @click="handleDuplicate">
              {{ $t("components.label_template.card.duplicate") }}
            </DropdownMenuItem>
            <DropdownMenuItem @click="handleExport">
              {{ $t("components.label_template.card.export_json") }}
            </DropdownMenuItem>
            <DropdownMenuItem @click="toggleDefault">
              {{
                isDefault
                  ? $t("components.label_template.card.clear_default")
                  : $t("components.label_template.card.set_default")
              }}
            </DropdownMenuItem>
            <template v-if="template.isOwner">
              <DropdownMenuSeparator />
              <DropdownMenuItem class="text-destructive focus:text-destructive" @click="handleDelete">
                {{ $t("components.label_template.card.delete") }}
              </DropdownMenuItem>
            </template>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  </Card>
</template>
