<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { Input } from "@/components/ui/input";
  import { Label } from "@/components/ui/label";
  import { Textarea } from "@/components/ui/textarea";
  import { Switch } from "@/components/ui/switch";
  import { Button } from "@/components/ui/button";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import type { LabelmakerLabelPreset } from "~/lib/api/types/data-contracts";

  export interface EditorSettings {
    name: string;
    description: string;
    width: number;
    height: number;
    preset: string;
    dpi: number;
    isShared: boolean;
  }

  export interface DisplaySettings {
    showRuler: boolean;
    showGrid: boolean;
    snapToGrid: boolean;
    gridSize: number;
  }

  const settings = defineModel<EditorSettings>({ required: true });
  const display = defineModel<DisplaySettings>("display", { required: true });

  const props = defineProps<{
    presets: LabelmakerLabelPreset[];
  }>();

  const { t } = useI18n();

  // --- presets grouped by brand and kind -------------------------------------
  type PresetCategory = "sheet" | "continuous" | "die-cut";
  function categoryOf(preset: LabelmakerLabelPreset): PresetCategory {
    if (preset.sheetLayout) return "sheet";
    if (preset.continuous) return "continuous";
    return "die-cut";
  }

  const presetGroups = computed(() => {
    const brandMap: Record<string, Record<PresetCategory, LabelmakerLabelPreset[]>> = {};
    for (const preset of props.presets) {
      const brand = preset.brand || t("components.label_template.form.brand_custom");
      brandMap[brand] ??= { sheet: [], continuous: [], "die-cut": [] };
      brandMap[brand][categoryOf(preset)].push(preset);
    }
    const brandOrder = ["Avery", "Brother", "Dymo", "Brady", "Zebra", t("components.label_template.form.brand_custom")];
    for (const brand of Object.keys(brandMap)) if (!brandOrder.includes(brand)) brandOrder.push(brand);
    const categoryLabels: Record<PresetCategory, string> = {
      sheet: t("components.label_template.form.sheet_labels"),
      continuous: t("components.label_template.form.continuous_labels"),
      "die-cut": t("components.label_template.form.die_cut_labels"),
    };
    const groups: { key: string; label: string; presets: LabelmakerLabelPreset[] }[] = [];
    for (const brand of brandOrder) {
      const cats = brandMap[brand];
      if (!cats) continue;
      for (const category of ["sheet", "die-cut", "continuous"] as PresetCategory[]) {
        const list = cats[category];
        if (list.length === 0) continue;
        list.sort((a, b) => a.name.localeCompare(b.name));
        groups.push({ key: `${brand}-${category}`, label: `${brand} · ${categoryLabels[category]}`, presets: list });
      }
    }
    return groups;
  });

  const selectedPreset = computed(() => props.presets.find(p => p.key === settings.value.preset) ?? null);
  const isWidthLocked = computed(() => !!selectedPreset.value);
  const isHeightLocked = computed(() => (selectedPreset.value ? !selectedPreset.value.continuous : false));

  function handlePresetChange(value: unknown) {
    const v = String(value ?? "");
    // "custom" is the backend's free-size preset (0 x 0 mm); treat it as no preset.
    if (v === "none" || v === "" || v === "custom") {
      settings.value.preset = "";
      return;
    }
    settings.value.preset = v;
    const preset = props.presets.find(p => p.key === v);
    if (preset) {
      settings.value.width = preset.width;
      if (!preset.continuous) settings.value.height = preset.height;
    }
  }

  const orientation = computed(() => (settings.value.width >= settings.value.height ? "horizontal" : "vertical"));
  function toggleOrientation() {
    const w = settings.value.width;
    settings.value.width = settings.value.height;
    settings.value.height = w;
  }

  const dpiOptions = [150, 203, 300, 600];
</script>

<template>
  <div class="space-y-5">
    <div class="space-y-2">
      <Label for="ls-name">{{ $t("components.label_template.form.name") }}</Label>
      <Input id="ls-name" v-model="settings.name" />
    </div>
    <div class="space-y-2">
      <Label for="ls-description">{{ $t("components.label_template.form.description") }}</Label>
      <Textarea id="ls-description" v-model="settings.description" rows="2" />
    </div>

    <div class="space-y-2">
      <Label>{{ $t("components.label_template.form.label_preset") }}</Label>
      <Select :model-value="settings.preset || 'none'" @update:model-value="handlePresetChange">
        <SelectTrigger>
          <SelectValue />
        </SelectTrigger>
        <SelectContent class="max-h-80">
          <SelectItem value="none">{{ $t("components.label_template.form.preset_custom") }}</SelectItem>
          <template v-for="group in presetGroups" :key="group.key">
            <div class="px-2 py-1.5 text-xs font-semibold text-muted-foreground">{{ group.label }}</div>
            <SelectItem v-for="preset in group.presets" :key="preset.key" :value="preset.key">
              {{ preset.name }}
              <span v-if="preset.twoColor" class="ml-1 font-medium text-red-600 dark:text-red-400">
                {{ $t("components.label_template.form.two_color") }}
              </span>
              <span class="ml-1 text-muted-foreground">
                ({{ preset.width }}{{ preset.continuous ? " mm" : ` × ${preset.height} mm` }})
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
        <template v-else-if="selectedPreset?.twoColor">{{
          $t("components.label_template.form.two_color_hint")
        }}</template>
        <template v-else-if="selectedPreset?.continuous">{{
          $t("components.label_template.form.continuous_hint")
        }}</template>
        <template v-else-if="selectedPreset">{{ $t("components.label_template.form.die_cut_hint") }}</template>
        <template v-else>{{ $t("components.label_template.form.preset_hint") }}</template>
      </p>
    </div>

    <div class="grid grid-cols-2 gap-3">
      <div class="space-y-2">
        <Label for="ls-width">{{ $t("components.label_template.form.width") }}</Label>
        <div class="flex items-center gap-1">
          <Input
            id="ls-width"
            v-model.number="settings.width"
            type="number"
            step="0.1"
            min="1"
            :disabled="isWidthLocked"
          />
          <span class="shrink-0 text-xs text-muted-foreground">mm</span>
        </div>
      </div>
      <div class="space-y-2">
        <Label for="ls-height">{{ $t("components.label_template.form.height") }}</Label>
        <div class="flex items-center gap-1">
          <Input
            id="ls-height"
            v-model.number="settings.height"
            type="number"
            step="0.1"
            min="1"
            :disabled="isHeightLocked"
          />
          <span class="shrink-0 text-xs text-muted-foreground">mm</span>
        </div>
      </div>
    </div>

    <div v-if="!isWidthLocked && !isHeightLocked" class="flex items-center justify-between">
      <Label class="text-sm">{{ $t("components.label_template.editor.display.orientation") }}</Label>
      <Button variant="outline" size="sm" @click="toggleOrientation">
        {{
          orientation === "horizontal"
            ? $t("components.label_template.editor.display.horizontal")
            : $t("components.label_template.editor.display.vertical")
        }}
      </Button>
    </div>

    <div class="grid grid-cols-2 gap-3">
      <div class="space-y-2">
        <Label>{{ $t("components.label_template.form.dpi") }}</Label>
        <Select :model-value="String(settings.dpi)" @update:model-value="v => (settings.dpi = Number(v))">
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem v-for="d in dpiOptions" :key="d" :value="String(d)">{{ d }} dpi</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>

    <div class="flex items-center justify-between gap-4 rounded-md border p-3">
      <div class="space-y-0.5">
        <Label for="ls-shared">{{ $t("components.label_template.form.share_with_group") }}</Label>
        <p class="text-xs text-muted-foreground">{{ $t("components.label_template.form.share_description") }}</p>
      </div>
      <Switch id="ls-shared" v-model="settings.isShared" />
    </div>

    <div class="space-y-3 border-t pt-4">
      <h4 class="text-sm font-medium">{{ $t("components.label_template.editor.display.title") }}</h4>
      <label class="flex items-center justify-between text-sm">
        {{ $t("components.label_template.editor.display.show_ruler") }}
        <input v-model="display.showRuler" type="checkbox" class="size-4" />
      </label>
      <label class="flex items-center justify-between text-sm">
        {{ $t("components.label_template.editor.display.show_grid") }}
        <input v-model="display.showGrid" type="checkbox" class="size-4" />
      </label>
      <label class="flex items-center justify-between text-sm">
        {{ $t("components.label_template.editor.display.snap_to_grid") }}
        <input v-model="display.snapToGrid" type="checkbox" class="size-4" />
      </label>
      <div class="flex items-center justify-between gap-3 text-sm">
        <Label for="ls-grid" class="text-sm">{{ $t("components.label_template.editor.display.grid_size") }}</Label>
        <div class="flex items-center gap-1">
          <Input
            id="ls-grid"
            v-model.number="display.gridSize"
            type="number"
            min="5"
            max="50"
            step="5"
            class="h-8 w-20"
          />
          <span class="text-xs text-muted-foreground">px</span>
        </div>
      </div>
    </div>
  </div>
</template>
