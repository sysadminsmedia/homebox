<template>
  <BaseModal :dialog-id="DialogID.CreateLabelTemplate" :title="$t('components.label_template.create_modal.title')">
    <form class="flex flex-col gap-4" @submit.prevent="handleSubmit">
      <div class="space-y-2">
        <Label for="name">{{ $t("components.label_template.form.name") }}</Label>
        <Input id="name" v-model="form.name" :placeholder="$t('components.label_template.form.name_placeholder')" />
      </div>

      <div class="space-y-2">
        <Label for="description">{{ $t("components.label_template.form.description") }}</Label>
        <Textarea
          id="description"
          v-model="form.description"
          :placeholder="$t('components.label_template.form.description_placeholder')"
          rows="2"
        />
      </div>

      <div class="space-y-2">
        <Label for="preset">{{ $t("components.label_template.form.label_preset") }}</Label>
        <Select :model-value="form.preset || 'none'" @update:model-value="handlePresetChange">
          <SelectTrigger>
            <SelectValue :placeholder="$t('components.label_template.form.preset_placeholder')" />
          </SelectTrigger>
          <SelectContent class="max-h-80">
            <SelectItem value="none">{{ $t("components.label_template.form.preset_custom") }}</SelectItem>
            <template v-for="group in presetGroups" :key="`${group.brand}-${group.category}`">
              <div class="px-2 py-1.5 text-xs font-semibold text-muted-foreground">
                {{ group.label }}
              </div>
              <SelectItem v-for="preset in group.presets" :key="preset.key" :value="preset.key">
                {{ preset.name }}
                <span v-if="preset.twoColor" class="ml-1 font-medium text-red-600 dark:text-red-400">
                  {{ $t("components.label_template.form.two_color") }}
                </span>
                <span class="text-muted-foreground">
                  ({{ preset.width }}{{ preset.continuous ? "mm wide" : `x${preset.height}mm` }})
                </span>
              </SelectItem>
            </template>
          </SelectContent>
        </Select>
        <p class="text-xs text-muted-foreground">
          <template v-if="selectedPreset?.sheetLayout">
            {{
              $t("components.label_template.form.sheet_preset_hint", {
                cols: selectedPreset.sheetLayout.columns,
                rows: selectedPreset.sheetLayout.rows,
              })
            }}
          </template>
          <template v-else-if="selectedPreset?.twoColor">
            {{ $t("components.label_template.form.two_color_hint") }}
          </template>
          <template v-else-if="selectedPreset?.continuous">
            {{ $t("components.label_template.form.continuous_hint") }}
          </template>
          <template v-else-if="selectedPreset">
            {{ $t("components.label_template.form.die_cut_hint") }}
          </template>
          <template v-else>
            {{ $t("components.label_template.form.preset_hint") }}
          </template>
        </p>
      </div>

      <div class="grid grid-cols-2 gap-4">
        <div class="space-y-2">
          <Label for="width">{{ $t("components.label_template.form.width") }}</Label>
          <div class="flex items-center gap-2">
            <Input id="width" v-model.number="form.width" type="number" step="0.1" min="1" :disabled="isWidthLocked" />
            <span class="text-sm text-muted-foreground">mm</span>
          </div>
        </div>
        <div class="space-y-2">
          <Label for="height">{{ $t("components.label_template.form.height") }}</Label>
          <div class="flex items-center gap-2">
            <Input
              id="height"
              v-model.number="form.height"
              type="number"
              step="0.1"
              min="1"
              :disabled="isHeightLocked"
            />
            <span class="text-sm text-muted-foreground">mm</span>
          </div>
        </div>
      </div>

      <div class="space-y-2">
        <Label for="dpi">{{ $t("components.label_template.form.dpi") }}</Label>
        <Select :model-value="String(form.dpi)" @update:model-value="v => (form.dpi = Number(v))">
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="150">150 DPI</SelectItem>
            <SelectItem value="203">203 DPI</SelectItem>
            <SelectItem value="300">300 DPI (Recommended)</SelectItem>
            <SelectItem value="600">600 DPI</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div class="flex items-center justify-between">
        <div class="space-y-0.5">
          <Label for="shared">{{ $t("components.label_template.form.share_with_group") }}</Label>
          <p class="text-xs text-muted-foreground">
            {{ $t("components.label_template.form.share_description") }}
          </p>
        </div>
        <Switch id="shared" v-model:checked="form.isShared" />
      </div>

      <div class="flex justify-end gap-2 pt-2">
        <Button type="button" variant="outline" @click="closeDialog(DialogID.CreateLabelTemplate)">
          {{ $t("global.cancel") }}
        </Button>
        <Button type="submit" :disabled="isSubmitting">
          {{ $t("global.create") }}
        </Button>
      </div>
    </form>
  </BaseModal>
</template>

<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import { Button } from "@/components/ui/button";
  import { Input } from "@/components/ui/input";
  import { Label } from "@/components/ui/label";
  import { Textarea } from "@/components/ui/textarea";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import { Switch } from "@/components/ui/switch";
  import { useDialog, DialogID } from "@/components/ui/dialog-provider/utils";
  import BaseModal from "@/components/App/CreateModal.vue";
  import type { LabelTemplateCreate, LabelmakerLabelPreset } from "~/lib/api/types/data-contracts";

  const emit = defineEmits<{
    created: [id: string];
  }>();

  const { t } = useI18n();
  const { closeDialog, registerOpenDialogCallback } = useDialog();
  const { create } = useLabelTemplateActions();
  const api = useUserApi();

  const form = reactive<LabelTemplateCreate>({
    name: "",
    description: "",
    width: 62,
    height: 29,
    preset: "",
    isShared: false,
    outputFormat: "png",
    dpi: 300,
    canvasData: {},
  });

  const isSubmitting = ref(false);
  const labelPresets = ref<LabelmakerLabelPreset[]>([]);

  // Fetch label presets on mount
  onMounted(async () => {
    try {
      const { data } = await api.labelTemplates.getPresets();
      if (data) {
        labelPresets.value = data;
      }
    } catch {
      // Silently fail - presets are optional
    }
  });

  // Categorize a preset
  type PresetCategory = "sheet" | "continuous" | "die-cut";
  function getPresetCategory(preset: LabelmakerLabelPreset): PresetCategory {
    if (preset.sheetLayout) return "sheet";
    if (preset.continuous) return "continuous";
    return "die-cut";
  }

  // Group presets by brand and category
  interface PresetGroup {
    brand: string;
    category: PresetCategory;
    label: string;
    presets: LabelmakerLabelPreset[];
  }

  const presetGroups = computed(() => {
    const groups: PresetGroup[] = [];
    const brandMap: Record<string, Record<PresetCategory, LabelmakerLabelPreset[]>> = {};

    for (const preset of labelPresets.value) {
      const brand = preset.brand || "Custom";
      const category = getPresetCategory(preset);

      if (!brandMap[brand]) {
        brandMap[brand] = { sheet: [], continuous: [], "die-cut": [] };
      }
      brandMap[brand][category].push(preset);
    }

    // Sort presets within each category
    for (const brand of Object.keys(brandMap)) {
      const brandCategories = brandMap[brand];
      if (!brandCategories) continue;
      for (const category of Object.keys(brandCategories) as PresetCategory[]) {
        const presets = brandCategories[category];
        if (presets) {
          presets.sort((a, b) => a.name.localeCompare(b.name));
        }
      }
    }

    // Define brand order
    const brandOrder = ["Avery", "Brother", "Dymo", "Brady", "Zebra", "Custom"];
    const categoryLabels: Record<PresetCategory, string> = {
      sheet: t("components.label_template.form.sheet_labels"),
      continuous: t("components.label_template.form.continuous_labels"),
      "die-cut": t("components.label_template.form.die_cut_labels"),
    };
    const categoryOrder: PresetCategory[] = ["sheet", "die-cut", "continuous"];

    // Build ordered groups
    const orderedBrands = [...brandOrder];
    for (const brand of Object.keys(brandMap)) {
      if (!orderedBrands.includes(brand)) {
        orderedBrands.push(brand);
      }
    }

    for (const brand of orderedBrands) {
      const brandCategories = brandMap[brand];
      if (!brandCategories) continue;

      for (const category of categoryOrder) {
        const presets = brandCategories[category];
        if (presets && presets.length > 0) {
          groups.push({
            brand,
            category,
            label: `${brand} - ${categoryLabels[category]}`,
            presets,
          });
        }
      }
    }

    return groups;
  });

  // Get the currently selected preset
  const selectedPreset = computed(() => {
    if (!form.preset) return null;
    return labelPresets.value.find(p => p.key === form.preset) || null;
  });

  // Determine if width should be locked (preset selected)
  const isWidthLocked = computed(() => {
    return !!selectedPreset.value;
  });

  // Determine if height should be locked (non-continuous preset selected)
  const isHeightLocked = computed(() => {
    return selectedPreset.value ? !selectedPreset.value.continuous : false;
  });

  function handlePresetChange(value: unknown) {
    const v = String(value || "");

    if (v === "none" || v === "custom") {
      form.preset = "";
      // Don't reset dimensions - let user keep custom values
      return;
    }

    form.preset = v;

    // Find the preset and update dimensions
    const preset = labelPresets.value.find(p => p.key === v);
    if (preset) {
      form.width = preset.width;

      if (!preset.continuous) {
        // Die-cut/sheet labels: set fixed height
        form.height = preset.height;
      }
      // For continuous, keep the current height or set a reasonable default
      else if (form.height < 10) {
        form.height = preset.height || 29; // Default minimum height
      }
    }
  }

  async function handleSubmit() {
    if (!form.name.trim()) {
      toast.error(t("components.label_template.validation.name_required"));
      return;
    }

    isSubmitting.value = true;

    try {
      const template = await create(form);
      if (template) {
        toast.success(t("components.label_template.toast.created"));
        emit("created", template.id);
        closeDialog(DialogID.CreateLabelTemplate);
        resetForm();

        // Navigate to editor
        navigateTo(`/label-templates/${template.id}/edit`);
      }
    } catch {
      toast.error(t("components.label_template.toast.create_failed"));
    } finally {
      isSubmitting.value = false;
    }
  }

  function resetForm() {
    form.name = "";
    form.description = "";
    form.width = 62;
    form.height = 29;
    form.preset = "";
    form.isShared = false;
    form.outputFormat = "png";
    form.dpi = 300;
    form.canvasData = {};
  }

  registerOpenDialogCallback(DialogID.CreateLabelTemplate, () => {
    resetForm();
  });
</script>
