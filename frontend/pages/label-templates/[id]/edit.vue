<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import MdiArrowLeft from "~icons/mdi/arrow-left";
  import MdiContentSave from "~icons/mdi/content-save";
  import MdiCog from "~icons/mdi/cog";
  import MdiMinus from "~icons/mdi/minus";
  import MdiPlus from "~icons/mdi/plus";
  import MdiFitToScreen from "~icons/mdi/fit-to-screen";
  import MdiMagnify from "~icons/mdi/magnify";
  import { Button } from "@/components/ui/button";
  import { Input } from "@/components/ui/input";
  import { Label } from "@/components/ui/label";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
  import BaseContainer from "@/components/Base/Container.vue";
  import LabelTemplateEditorCanvas from "~/components/LabelTemplate/Editor/Canvas.vue";
  import LabelTemplateEditorToolbar from "~/components/LabelTemplate/Editor/Toolbar.vue";
  import LabelTemplateEditorTextToolbar from "~/components/LabelTemplate/Editor/TextToolbar.vue";
  import LabelTemplateEditorTransformPanel from "~/components/LabelTemplate/Editor/TransformPanel.vue";
  import LabelTemplateEditorLayerPanel from "~/components/LabelTemplate/Editor/LayerPanel.vue";
  import LabelTemplateEditorRuler from "~/components/LabelTemplate/Editor/Ruler.vue";
  import LabelTemplateEditorPreviewPanel from "~/components/LabelTemplate/Editor/PreviewPanel.vue";
  import type { LabelmakerLabelPreset } from "~/lib/api/types/data-contracts";

  definePageMeta({
    middleware: ["auth"],
  });

  const { t } = useI18n();
  const route = useRoute();
  const api = useUserApi();
  const templateId = computed(() => route.params.id as string);

  useHead({
    title: computed(() => `HomeBox | ${t("components.label_template.editor.title")}`),
  });

  const { template, pending, refresh } = useLabelTemplate(templateId);
  const { update } = useLabelTemplateActions();

  // Label presets for dimension constraints
  const labelPresets = ref<LabelmakerLabelPreset[]>([]);

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

  // Local editable values for template settings
  const localWidth = ref(62);
  const localHeight = ref(29);
  const localPreset = ref<string | undefined>(undefined);

  // Initialize local values from template
  watch(
    template,
    newTemplate => {
      if (newTemplate) {
        localWidth.value = newTemplate.width;
        localHeight.value = newTemplate.height;
        localPreset.value = newTemplate.preset || undefined;
      }
    },
    { immediate: true }
  );

  // Get selected preset info
  const selectedPreset = computed(() => {
    if (!localPreset.value) return null;
    return labelPresets.value.find(p => p.key === localPreset.value) || null;
  });

  // Determine if dimensions should be locked
  const isWidthLocked = computed(() => !!selectedPreset.value);
  const isHeightLocked = computed(() => (selectedPreset.value ? !selectedPreset.value.continuous : false));

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

  function handlePresetChange(value: unknown) {
    const v = String(value || "");
    if (v === "none" || v === "" || v === "custom") {
      localPreset.value = undefined;
      return;
    }

    localPreset.value = v;

    const preset = labelPresets.value.find(p => p.key === v);
    if (preset) {
      localWidth.value = preset.width;
      if (!preset.continuous) {
        localHeight.value = preset.height;
      }
    }
  }

  const canvasRef = ref<InstanceType<typeof LabelTemplateEditorCanvas> | null>(null);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const selectedObject = ref<any>(null);
  // Track the object ID to detect when a different object is selected
  const selectedObjectId = ref<number | null>(null);
  const isSaving = ref(false);

  // Check if a text object is selected (for showing text toolbar)
  const isTextSelected = computed(() => {
    const type = selectedObject.value?.type?.toLowerCase();
    return type === "textbox" || type === "i-text" || type === "itext" || type === "text";
  });

  const localCanvasData = ref<Record<string, unknown>>({});

  // Layer management
  interface LayerItem {
    index: number;
    type: string;
    name: string;
    visible: boolean;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    object: any;
  }
  const layers = ref<LayerItem[]>([]);

  // Grid settings
  const showGrid = ref(false);
  const snapToGrid = ref(false);
  const gridSize = ref(10); // pixels

  // Ruler settings
  const showRuler = ref(true);
  const rulerUnit = ref<"metric" | "imperial">("metric");

  // Orientation computed - horizontal when width > height
  const orientation = computed(() => (localWidth.value >= localHeight.value ? "horizontal" : "vertical"));

  // Toggle orientation by swapping width and height
  function toggleOrientation() {
    const temp = localWidth.value;
    localWidth.value = localHeight.value;
    localHeight.value = temp;
  }

  // Conversion helpers
  const mmPerInch = 25.4;
  const formatDimension = (mm: number) => {
    if (rulerUnit.value === "imperial") {
      const inches = mm / mmPerInch;
      return `${inches.toFixed(2)}"`;
    }
    return `${mm}mm`;
  };

  // Computed pixel dimensions for ruler
  const screenDPI = 96;
  const pixelWidth = computed(() => Math.round((localWidth.value / 25.4) * screenDPI));
  const pixelHeight = computed(() => Math.round((localHeight.value / 25.4) * screenDPI));

  function refreshLayers() {
    layers.value = canvasRef.value?.getLayers() || [];
  }

  // Watch for template changes to initialize canvas data
  watch(
    template,
    newTemplate => {
      if (newTemplate?.canvasData) {
        localCanvasData.value = newTemplate.canvasData;
      }
    },
    { immediate: true }
  );

  function handleCanvasUpdate(data: Record<string, unknown>) {
    localCanvasData.value = data;
    // Refresh layers when canvas changes
    nextTick(() => refreshLayers());
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  function handleSelectionChanged(obj: any) {
    // Get the object's unique ID (Fabric.js assigns __id or we can use object reference)
    const newObjectId = obj ? (obj.__id ?? obj.id ?? null) : null;

    // Only update if it's a different object (not just property changes on the same object)
    if (newObjectId !== selectedObjectId.value || obj !== selectedObject.value) {
      selectedObjectId.value = newObjectId;
      selectedObject.value = obj;
    }
  }

  function handleAddText() {
    canvasRef.value?.addText();
  }

  function handleInsertField(fieldName: string) {
    // Get the fabric canvas and the active object
    const canvas = canvasRef.value?.getCanvas();
    if (!canvas) return;

    const activeObject = canvas.getActiveObject();
    if (!activeObject) return;

    // Check if it's a textbox
    const type = activeObject.type?.toLowerCase();
    if (type !== "textbox") return;

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const textbox = activeObject as any;
    const placeholder = `{{${fieldName}}}`;

    // If the textbox is being edited, insert at cursor position
    if (textbox.isEditing) {
      const selectionStart = textbox.selectionStart || 0;
      const selectionEnd = textbox.selectionEnd || 0;
      const currentText = textbox.text || "";

      // Replace selection or insert at cursor
      const newText = currentText.substring(0, selectionStart) + placeholder + currentText.substring(selectionEnd);
      textbox.set("text", newText);

      // Move cursor after the inserted placeholder
      textbox.selectionStart = selectionStart + placeholder.length;
      textbox.selectionEnd = selectionStart + placeholder.length;
    } else {
      // Append to end of text
      textbox.set("text", (textbox.text || "") + placeholder);
    }

    canvas.renderAll();
    canvasRef.value?.updateSelectedObject({}); // Trigger save
  }

  function handleAddBarcode(format: string, contentSource: string) {
    canvasRef.value?.addBarcode(format, contentSource);
  }

  function handleAddShape(type: "rect" | "line") {
    canvasRef.value?.addShape(type);
  }

  function handleDeleteSelected() {
    canvasRef.value?.deleteSelected();
    selectedObject.value = null;
  }

  function handleUpdateObject(updates: Record<string, unknown>) {
    canvasRef.value?.updateSelectedObject(updates);
  }

  // Layer management handlers
  function handleSelectLayer(index: number) {
    canvasRef.value?.selectLayerByIndex(index);
  }

  function handleBringForward(index: number) {
    canvasRef.value?.bringForward(index);
    nextTick(() => refreshLayers());
  }

  function handleSendBackward(index: number) {
    canvasRef.value?.sendBackward(index);
    nextTick(() => refreshLayers());
  }

  function handleBringToFront(index: number) {
    canvasRef.value?.bringToFront(index);
    nextTick(() => refreshLayers());
  }

  function handleSendToBack(index: number) {
    canvasRef.value?.sendToBack(index);
    nextTick(() => refreshLayers());
  }

  function handleToggleVisibility(index: number) {
    canvasRef.value?.toggleLayerVisibility(index);
    nextTick(() => refreshLayers());
  }

  async function handleSave() {
    if (!template.value) return;

    isSaving.value = true;

    try {
      await update(templateId.value, {
        id: templateId.value,
        name: template.value.name,
        description: template.value.description,
        width: localWidth.value,
        height: localHeight.value,
        preset: localPreset.value || "",
        isShared: template.value.isShared,
        outputFormat: template.value.outputFormat as "png" | "pdf",
        dpi: template.value.dpi,
        canvasData: localCanvasData.value,
      });

      toast.success(t("components.label_template.toast.updated"));
      await refresh();
    } catch {
      toast.error(t("components.label_template.toast.update_failed"));
    } finally {
      isSaving.value = false;
    }
  }

  // Track if dimensions have changed for canvas update
  const dimensionsChanged = computed(() => {
    if (!template.value) return false;
    return localWidth.value !== template.value.width || localHeight.value !== template.value.height;
  });

  // Serialized canvas data for live preview
  const canvasDataJson = computed(() => {
    if (!localCanvasData.value || Object.keys(localCanvasData.value).length === 0) {
      return "";
    }
    return JSON.stringify(localCanvasData.value);
  });
</script>

<template>
  <BaseContainer class="max-w-none">
    <div v-if="pending" class="flex items-center justify-center py-12">
      <div class="text-muted-foreground">{{ $t("global.loading") }}</div>
    </div>

    <template v-else-if="template">
      <div class="mb-4 flex items-center justify-between">
        <div class="flex items-center gap-4">
          <Button variant="ghost" size="icon" as-child>
            <NuxtLink to="/label-templates">
              <MdiArrowLeft class="size-5" />
            </NuxtLink>
          </Button>
          <div>
            <h1 class="text-xl font-semibold">{{ template.name }}</h1>
            <p class="text-sm text-muted-foreground">
              {{ formatDimension(localWidth) }} x {{ formatDimension(localHeight) }} @ {{ template.dpi }} DPI
              <span v-if="selectedPreset" class="ml-1">({{ selectedPreset.brand }} {{ selectedPreset.name }})</span>
              <span v-if="dimensionsChanged" class="ml-2 text-amber-600">*unsaved</span>
            </p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <Popover>
            <PopoverTrigger as-child>
              <Button variant="outline" size="icon">
                <MdiCog class="size-4" />
              </Button>
            </PopoverTrigger>
            <PopoverContent class="w-80" align="end">
              <div class="space-y-4">
                <h3 class="font-medium">{{ $t("components.label_template.editor.settings.title") }}</h3>

                <div class="space-y-2">
                  <Label>{{ $t("components.label_template.form.label_preset") }}</Label>
                  <Select :model-value="localPreset || 'none'" @update:model-value="handlePresetChange">
                    <SelectTrigger>
                      <SelectValue />
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
                </div>

                <div class="grid grid-cols-2 gap-4">
                  <div class="space-y-2">
                    <Label>{{ $t("components.label_template.form.width") }}</Label>
                    <div class="flex items-center gap-1">
                      <Input v-model.number="localWidth" type="number" step="0.1" min="1" :disabled="isWidthLocked" />
                      <span class="text-xs text-muted-foreground">mm</span>
                    </div>
                  </div>
                  <div class="space-y-2">
                    <Label>{{ $t("components.label_template.form.height") }}</Label>
                    <div class="flex items-center gap-1">
                      <Input v-model.number="localHeight" type="number" step="0.1" min="1" :disabled="isHeightLocked" />
                      <span class="text-xs text-muted-foreground">mm</span>
                    </div>
                  </div>
                </div>

                <!-- Orientation toggle (only when dimensions are not locked) -->
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

                <p v-if="selectedPreset?.sheetLayout" class="text-xs text-muted-foreground">
                  {{
                    $t("components.label_template.form.sheet_preset_hint", {
                      cols: selectedPreset.sheetLayout.columns,
                      rows: selectedPreset.sheetLayout.rows,
                    })
                  }}
                </p>
                <p v-else-if="selectedPreset?.continuous" class="text-xs text-muted-foreground">
                  {{ $t("components.label_template.form.continuous_hint") }}
                </p>
                <p v-else-if="selectedPreset" class="text-xs text-muted-foreground">
                  {{ $t("components.label_template.form.die_cut_hint") }}
                </p>

                <!-- Grid & Display Settings -->
                <div class="border-t pt-4">
                  <h4 class="mb-3 text-sm font-medium">{{ $t("components.label_template.editor.display.title") }}</h4>
                  <div class="space-y-3">
                    <div class="space-y-2">
                      <Label class="text-sm">{{ $t("components.label_template.editor.display.ruler_units") }}</Label>
                      <Select v-model="rulerUnit">
                        <SelectTrigger class="h-8">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="metric">{{
                            $t("components.label_template.editor.display.metric")
                          }}</SelectItem>
                          <SelectItem value="imperial">{{
                            $t("components.label_template.editor.display.imperial")
                          }}</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    <div class="flex items-center justify-between">
                      <Label for="showRuler" class="text-sm">{{
                        $t("components.label_template.editor.display.show_ruler")
                      }}</Label>
                      <input id="showRuler" v-model="showRuler" type="checkbox" class="size-4" />
                    </div>
                    <div class="flex items-center justify-between">
                      <Label for="showGrid" class="text-sm">{{
                        $t("components.label_template.editor.display.show_grid")
                      }}</Label>
                      <input id="showGrid" v-model="showGrid" type="checkbox" class="size-4" />
                    </div>
                    <div class="flex items-center justify-between">
                      <Label for="snapToGrid" class="text-sm">{{
                        $t("components.label_template.editor.display.snap_to_grid")
                      }}</Label>
                      <input id="snapToGrid" v-model="snapToGrid" type="checkbox" class="size-4" />
                    </div>
                    <div class="space-y-2">
                      <Label for="gridSize" class="text-sm">{{
                        $t("components.label_template.editor.display.grid_size")
                      }}</Label>
                      <div class="flex items-center gap-2">
                        <Input
                          id="gridSize"
                          v-model.number="gridSize"
                          type="number"
                          min="5"
                          max="50"
                          step="5"
                          class="h-8"
                        />
                        <span class="text-xs text-muted-foreground">px</span>
                      </div>
                    </div>
                  </div>
                </div>

                <p v-if="dimensionsChanged" class="text-xs text-amber-600">
                  {{ $t("components.label_template.editor.settings.save_reminder") }}
                </p>
              </div>
            </PopoverContent>
          </Popover>
          <Button :disabled="isSaving" @click="handleSave">
            <MdiContentSave class="mr-2 size-4" />
            {{ $t("global.save") }}
          </Button>
        </div>
      </div>

      <!-- Object Toolbar (Add Text, Barcode, Shape, Delete) -->
      <div class="mb-4">
        <LabelTemplateEditorToolbar
          @add-text="handleAddText"
          @add-barcode="handleAddBarcode"
          @add-shape="handleAddShape"
          @delete-selected="handleDeleteSelected"
        />
      </div>

      <div class="flex gap-4">
        <!-- Left Column: Text Toolbar + Canvas + Preview -->
        <div class="flex min-w-0 flex-1 flex-col gap-4">
          <!-- Text Toolbar (only visible when text is selected) -->
          <LabelTemplateEditorTextToolbar
            v-if="isTextSelected"
            :selected-object="selectedObject"
            @update-object="handleUpdateObject"
            @insert-field="handleInsertField"
          />

          <!-- Canvas with Ruler -->
          <div class="flex-1">
            <LabelTemplateEditorRuler
              v-if="showRuler"
              :width="localWidth"
              :height="localHeight"
              :pixel-width="Math.round(pixelWidth * (canvasRef?.zoomLevel ?? 1))"
              :pixel-height="Math.round(pixelHeight * (canvasRef?.zoomLevel ?? 1))"
              :unit="rulerUnit"
            >
              <LabelTemplateEditorCanvas
                ref="canvasRef"
                :width="localWidth"
                :height="localHeight"
                :dpi="template.dpi"
                :canvas-data="localCanvasData"
                :snap-to-grid="snapToGrid"
                :show-grid="showGrid"
                :grid-size="gridSize"
                @update:canvas-data="handleCanvasUpdate"
                @selection-changed="handleSelectionChanged"
              />
            </LabelTemplateEditorRuler>
            <LabelTemplateEditorCanvas
              v-else
              ref="canvasRef"
              :width="localWidth"
              :height="localHeight"
              :dpi="template.dpi"
              :canvas-data="localCanvasData"
              :snap-to-grid="snapToGrid"
              :show-grid="showGrid"
              :grid-size="gridSize"
              @update:canvas-data="handleCanvasUpdate"
              @selection-changed="handleSelectionChanged"
            />
          </div>

          <!-- Zoom Controls -->
          <div class="flex items-center gap-2 rounded-lg border bg-card px-3 py-2">
            <Button
              variant="outline"
              size="icon"
              class="size-7"
              :disabled="(canvasRef?.zoomLevel ?? 1) <= 0.25"
              @click="canvasRef?.zoomOut()"
            >
              <MdiMinus class="size-4" />
            </Button>
            <span class="w-12 text-center text-sm font-medium">
              {{ Math.round((canvasRef?.zoomLevel ?? 1) * 100) }}%
            </span>
            <Button
              variant="outline"
              size="icon"
              class="size-7"
              :disabled="(canvasRef?.zoomLevel ?? 1) >= 4"
              @click="canvasRef?.zoomIn()"
            >
              <MdiPlus class="size-4" />
            </Button>
            <div class="h-4 w-px bg-border" />
            <Button variant="outline" size="sm" class="h-7 px-2 text-xs" @click="canvasRef?.zoomToFit()">
              <MdiFitToScreen class="mr-1 size-3" />
              {{ $t("components.label_template.editor.zoom.fit") }}
            </Button>
            <Button variant="outline" size="sm" class="h-7 px-2 text-xs" @click="canvasRef?.zoomTo100()">
              <MdiMagnify class="mr-1 size-3" />
              100%
            </Button>
          </div>

          <!-- Preview Panel (moved below canvas) -->
          <LabelTemplateEditorPreviewPanel
            :template-id="templateId"
            :canvas-data="localCanvasData"
            :canvas-data-json="canvasDataJson"
          />
        </div>

        <!-- Right Sidebar: Transform + Layers only -->
        <div class="flex w-64 shrink-0 flex-col gap-4">
          <LabelTemplateEditorTransformPanel :selected-object="selectedObject" @update-object="handleUpdateObject" />
          <LabelTemplateEditorLayerPanel
            :layers="layers"
            :selected-object="selectedObject"
            @select-layer="handleSelectLayer"
            @bring-forward="handleBringForward"
            @send-backward="handleSendBackward"
            @bring-to-front="handleBringToFront"
            @send-to-back="handleSendToBack"
            @toggle-visibility="handleToggleVisibility"
          />
        </div>
      </div>
    </template>

    <div v-else class="flex flex-col items-center justify-center py-12">
      <p class="text-muted-foreground">Template not found</p>
      <Button variant="outline" class="mt-4" as-child>
        <NuxtLink to="/label-templates">Back to Templates</NuxtLink>
      </Button>
    </div>
  </BaseContainer>
</template>
