<script setup lang="ts">
  import { Button } from "@/components/ui/button";
  import { Input } from "@/components/ui/input";
  import { Label } from "@/components/ui/label";
  import { Switch } from "@/components/ui/switch";
  import LabelTemplateMeasurementInput from "@/components/LabelTemplate/MeasurementInput.vue";
  import type { LabelmakerLabelPreset, LabelmakerSheetLayout } from "~/lib/api/types/data-contracts";

  // Page layout for printing several labels per sheet, plus the print offset
  // used to calibrate a printer. A null layout follows the preset's; editing
  // any value saves a custom copy on the template.
  const props = defineProps<{
    preset: LabelmakerLabelPreset | null;
    labelWidth: number; // mm
    labelHeight: number; // mm
  }>();

  const sheetLayout = defineModel<LabelmakerSheetLayout | null>("sheetLayout", { required: true });
  const offsetX = defineModel<number>("offsetX", { required: true });
  const offsetY = defineModel<number>("offsetY", { required: true });

  const presetLayout = computed(() => props.preset?.sheetLayout ?? null);
  const layout = computed(() => sheetLayout.value ?? presetLayout.value);
  const isCustom = computed(() => sheetLayout.value !== null);

  // Sheet printing can only be switched off when there's no preset layout to fall back to
  const enabled = computed({
    get: () => layout.value !== null,
    set: on => {
      sheetLayout.value = on ? defaultLayout() : null;
    },
  });

  // A Letter sheet filled with as many labels as fit inside 10 mm margins
  function defaultLayout(): LabelmakerSheetLayout {
    const pageWidth = 215.9;
    const pageHeight = 279.4;
    const margin = 10;
    const fit = (page: number, label: number) => Math.max(1, Math.floor((page - 2 * margin) / Math.max(label, 1)));
    return {
      pageWidth,
      pageHeight,
      columns: fit(pageWidth, props.labelWidth),
      rows: fit(pageHeight, props.labelHeight),
      marginTop: margin,
      marginLeft: margin,
      gutterH: 0,
      gutterV: 0,
    };
  }

  function setField<K extends keyof LabelmakerSheetLayout>(key: K, value: LabelmakerSheetLayout[K]) {
    if (!layout.value) return;
    sheetLayout.value = { ...layout.value, [key]: value };
  }

  function setCount(key: "columns" | "rows", value: string | number) {
    const n = Math.round(Number(value));
    if (Number.isFinite(n) && n >= 1) setField(key, n);
  }

  function resetToPreset() {
    sheetLayout.value = null;
  }

  // Labels that run past the page edge won't print
  const overflow = computed(() => {
    const l = layout.value;
    if (!l) return false;
    const usedWidth = l.marginLeft + l.columns * props.labelWidth + (l.columns - 1) * l.gutterH;
    const usedHeight = l.marginTop + l.rows * props.labelHeight + (l.rows - 1) * l.gutterV;
    return usedWidth > l.pageWidth + 0.01 || usedHeight > l.pageHeight + 0.01;
  });
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between gap-2">
      <h3 class="font-medium">{{ $t("components.label_template.editor.sheet.title") }}</h3>
      <Switch
        v-if="!presetLayout"
        v-model="enabled"
        :aria-label="$t('components.label_template.editor.sheet.enable')"
      />
    </div>

    <p v-if="!layout" class="text-xs text-muted-foreground">
      {{ $t("components.label_template.editor.sheet.disabled_hint") }}
    </p>

    <template v-else>
      <p class="text-xs text-muted-foreground">
        <template v-if="isCustom && presetLayout">{{
          $t("components.label_template.editor.sheet.custom_hint")
        }}</template>
        <template v-else-if="presetLayout">
          {{ $t("components.label_template.editor.sheet.preset_hint", { preset: preset?.name ?? "" }) }}
        </template>
        <template v-else>{{ $t("components.label_template.editor.sheet.manual_hint") }}</template>
      </p>

      <div class="grid grid-cols-2 gap-3">
        <div class="space-y-1">
          <Label class="text-xs">{{ $t("components.label_template.editor.sheet.page_width") }}</Label>
          <LabelTemplateMeasurementInput
            :model-value="layout.pageWidth"
            :min="1"
            @update:model-value="v => setField('pageWidth', v)"
          />
        </div>
        <div class="space-y-1">
          <Label class="text-xs">{{ $t("components.label_template.editor.sheet.page_height") }}</Label>
          <LabelTemplateMeasurementInput
            :model-value="layout.pageHeight"
            :min="1"
            @update:model-value="v => setField('pageHeight', v)"
          />
        </div>
        <div class="space-y-1">
          <Label class="text-xs">{{ $t("components.label_template.editor.sheet.columns") }}</Label>
          <Input
            :model-value="layout.columns"
            type="number"
            min="1"
            step="1"
            @update:model-value="v => setCount('columns', v)"
          />
        </div>
        <div class="space-y-1">
          <Label class="text-xs">{{ $t("components.label_template.editor.sheet.rows") }}</Label>
          <Input
            :model-value="layout.rows"
            type="number"
            min="1"
            step="1"
            @update:model-value="v => setCount('rows', v)"
          />
        </div>
        <div class="space-y-1">
          <Label class="text-xs">{{ $t("components.label_template.editor.sheet.margin_top") }}</Label>
          <LabelTemplateMeasurementInput
            :model-value="layout.marginTop"
            :min="0"
            @update:model-value="v => setField('marginTop', v)"
          />
        </div>
        <div class="space-y-1">
          <Label class="text-xs">{{ $t("components.label_template.editor.sheet.margin_left") }}</Label>
          <LabelTemplateMeasurementInput
            :model-value="layout.marginLeft"
            :min="0"
            @update:model-value="v => setField('marginLeft', v)"
          />
        </div>
        <div class="space-y-1">
          <Label class="text-xs">{{ $t("components.label_template.editor.sheet.gap_across") }}</Label>
          <LabelTemplateMeasurementInput
            :model-value="layout.gutterH"
            :min="0"
            @update:model-value="v => setField('gutterH', v)"
          />
        </div>
        <div class="space-y-1">
          <Label class="text-xs">{{ $t("components.label_template.editor.sheet.gap_down") }}</Label>
          <LabelTemplateMeasurementInput
            :model-value="layout.gutterV"
            :min="0"
            @update:model-value="v => setField('gutterV', v)"
          />
        </div>
      </div>

      <p v-if="overflow" class="text-xs text-amber-600">
        {{ $t("components.label_template.editor.sheet.overflow") }}
      </p>

      <Button v-if="isCustom && presetLayout" variant="outline" size="sm" class="w-full" @click="resetToPreset">
        {{ $t("components.label_template.editor.sheet.reset", { preset: preset?.name ?? "" }) }}
      </Button>
    </template>

    <div class="space-y-2 border-t pt-4">
      <h4 class="text-sm font-medium">{{ $t("components.label_template.editor.sheet.offset_title") }}</h4>
      <p class="text-xs text-muted-foreground">{{ $t("components.label_template.editor.sheet.offset_hint") }}</p>
      <div class="grid grid-cols-2 gap-3">
        <div class="space-y-1">
          <Label class="text-xs">{{ $t("components.label_template.editor.sheet.offset_x") }}</Label>
          <LabelTemplateMeasurementInput v-model="offsetX" />
        </div>
        <div class="space-y-1">
          <Label class="text-xs">{{ $t("components.label_template.editor.sheet.offset_y") }}</Label>
          <LabelTemplateMeasurementInput v-model="offsetY" />
        </div>
      </div>
    </div>
  </div>
</template>
