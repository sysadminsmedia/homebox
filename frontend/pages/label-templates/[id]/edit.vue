<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { useEventListener } from "@vueuse/core";
  import { toast } from "@/components/ui/sonner";
  import MdiArrowLeft from "~icons/mdi/arrow-left";
  import MdiContentSave from "~icons/mdi/content-save";
  import MdiMinus from "~icons/mdi/minus";
  import MdiPlus from "~icons/mdi/plus";
  import MdiFitToScreen from "~icons/mdi/fit-to-screen";
  import MdiPrinterPos from "~icons/mdi/printer-pos";
  import MdiDelete from "~icons/mdi/delete";
  import MdiLoading from "~icons/mdi/loading";
  import { Button } from "@/components/ui/button";
  import { Input } from "@/components/ui/input";
  import { Badge } from "@/components/ui/badge";
  import BaseContainer from "@/components/Base/Container.vue";
  import LabelTemplateEditorCanvas from "~/components/LabelTemplate/Editor/Canvas.vue";
  import LabelTemplateEditorToolRail, { type AlignTarget } from "~/components/LabelTemplate/Editor/ToolRail.vue";
  import LabelTemplateEditorTextToolbar from "~/components/LabelTemplate/Editor/TextToolbar.vue";
  import LabelTemplateEditorBarcodePanel from "~/components/LabelTemplate/Editor/BarcodePanel.vue";
  import LabelTemplateEditorTransformPanel, {
    type EditorUnit,
  } from "~/components/LabelTemplate/Editor/TransformPanel.vue";
  import LabelTemplateEditorLayerPanel from "~/components/LabelTemplate/Editor/LayerPanel.vue";
  import LabelTemplateEditorPreviewPanel from "~/components/LabelTemplate/Editor/PreviewPanel.vue";
  import LabelTemplateEditorSheetLayoutPanel from "~/components/LabelTemplate/Editor/SheetLayoutPanel.vue";
  import LabelTemplateEditorLabelSettings, {
    type EditorSettings,
    type DisplaySettings,
  } from "~/components/LabelTemplate/Editor/LabelSettings.vue";
  import {
    DEFAULT_LABEL_BORDER,
    readLabelBorder,
    withLabelBorder,
    type LabelBorder,
  } from "~/components/LabelTemplate/Editor/label-border";
  import type { EntitySummary, LabelmakerSheetLayout } from "~/lib/api/types/data-contracts";

  definePageMeta({
    middleware: ["auth"],
  });

  const { t } = useI18n();
  const route = useRoute();
  const templateId = computed(() => route.params.id as string);

  useHead({
    title: computed(() => `HomeBox | ${t("components.label_template.editor.title")}`),
  });

  const { template, pending, refresh } = useLabelTemplate(templateId);
  const { update } = useLabelTemplateActions();
  const { presets } = useLabelPresets();
  const { title: presetTitle } = useLabelPresetLookup();
  const { labelPrinting } = useServerLabelPrinting();
  const { printLabels } = useDirectPrint();

  // ---------------------------------------------------------------------------
  // Editable copies of the template's settings and canvas
  // ---------------------------------------------------------------------------
  const settings = reactive<EditorSettings>({
    name: "",
    description: "",
    width: 62,
    height: 29,
    preset: "",
    dpi: 300,
    isShared: false,
  });

  const display = reactive<DisplaySettings>({
    showRuler: true,
    showGrid: false,
    snapToGrid: false,
    gridSize: 10,
  });

  // The border is part of canvasData, so it is saved with the design and
  // rendered by the server; here it is edited as its own small object.
  const border = reactive<LabelBorder>({ ...DEFAULT_LABEL_BORDER });

  // The unit follows the user's measurement preference; the toggle below changes it
  const measurement = useMeasurement();
  const preferences = useViewPreferences();
  const unit = computed<EditorUnit>({
    get: () => (measurement.isImperial.value ? "in" : "mm"),
    set: value => {
      preferences.value.measurementSystem = value === "in" ? "imperial" : "metric";
    },
  });
  const localCanvasData = ref<Record<string, unknown>>({});

  // Page layout for sheet printing (null follows the preset) and the printer calibration offset
  const sheetLayout = ref<LabelmakerSheetLayout | null>(null);
  const offsetX = ref(0);
  const offsetY = ref(0);
  const selectedPreset = computed(() => presets.value?.find(p => p.key === settings.preset) ?? null);

  function loadFromTemplate() {
    const tpl = template.value;
    if (!tpl) return;
    settings.name = tpl.name;
    settings.description = tpl.description ?? "";
    settings.width = tpl.width;
    settings.height = tpl.height;
    settings.preset = tpl.preset || "";
    settings.dpi = tpl.dpi;
    settings.isShared = tpl.isShared;
    if (tpl.canvasData) localCanvasData.value = tpl.canvasData;
    Object.assign(border, readLabelBorder(tpl.canvasData));
    sheetLayout.value = tpl.sheetLayout ? { ...tpl.sheetLayout } : null;
    offsetX.value = tpl.offsetX ?? 0;
    offsetY.value = tpl.offsetY ?? 0;
  }

  watch(template, loadFromTemplate, { immediate: true });

  // A custom layout is measured for one sheet; picking another preset drops it
  // so the new preset's layout applies.
  watch(
    () => settings.preset,
    preset => {
      if (preset !== (template.value?.preset || "")) sheetLayout.value = null;
    }
  );

  watch(border, () => {
    localCanvasData.value = withLabelBorder(localCanvasData.value, border);
  });

  // The API returns JSON objects with sorted keys while the editor builds them
  // in insertion order, so compare canvas data independent of key order.
  function stableStringify(value: unknown): string {
    if (Array.isArray(value)) return `[${value.map(stableStringify).join(",")}]`;
    if (value && typeof value === "object") {
      const o = value as Record<string, unknown>;
      return `{${Object.keys(o)
        .sort()
        .map(k => `${JSON.stringify(k)}:${stableStringify(o[k])}`)
        .join(",")}}`;
    }
    return JSON.stringify(value) ?? "null";
  }

  const isDirty = computed(() => {
    const tpl = template.value;
    if (!tpl) return false;
    const settingsChanged =
      settings.name !== tpl.name ||
      settings.description !== (tpl.description ?? "") ||
      settings.width !== tpl.width ||
      settings.height !== tpl.height ||
      settings.preset !== (tpl.preset || "") ||
      settings.dpi !== tpl.dpi ||
      settings.isShared !== tpl.isShared ||
      offsetX.value !== (tpl.offsetX ?? 0) ||
      offsetY.value !== (tpl.offsetY ?? 0) ||
      stableStringify(sheetLayout.value) !== stableStringify(tpl.sheetLayout ?? null);
    const canvasChanged = stableStringify(localCanvasData.value) !== stableStringify(tpl.canvasData ?? {});
    return settingsChanged || canvasChanged;
  });

  useEventListener(window, "beforeunload", (e: BeforeUnloadEvent) => {
    if (isDirty.value) e.preventDefault();
  });

  const sizeLabel = computed(() => measurement.formatSize(settings.width, settings.height));

  // ---------------------------------------------------------------------------
  // Canvas interaction
  // ---------------------------------------------------------------------------
  const canvasRef = ref<InstanceType<typeof LabelTemplateEditorCanvas> | null>(null);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const selectedObject = ref<any>(null);
  // Fabric mutates objects in place, which Vue cannot see; the panels re-read
  // the selection whenever this counter moves.
  const selectionRevision = ref(0);
  const isSaving = ref(false);
  const isTestPrinting = ref(false);
  const activeTab = ref<"element" | "layers" | "label">("element");
  const previewItem = ref<EntitySummary | null>(null);

  interface LayerItem {
    index: number;
    type: string;
    name: string;
    visible: boolean;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    object: any;
  }
  const layers = ref<LayerItem[]>([]);

  const selectedType = computed(() => String(selectedObject.value?.type || "").toLowerCase());
  const isTextSelected = computed(() => ["textbox", "i-text", "itext", "text"].includes(selectedType.value));
  const selectedTypeLabel = computed(() => {
    if (!selectedObject.value) return "";
    if (selectedObject.value.data?.type === "barcode") {
      return t("components.label_template.editor.element.type_barcode");
    }
    const map: Record<string, string> = {
      textbox: "type_text",
      "i-text": "type_text",
      itext: "type_text",
      text: "type_text",
      rect: "type_rect",
      line: "type_line",
      image: "type_image",
      group: "type_group",
    };
    return t(`components.label_template.editor.element.${map[selectedType.value] ?? "type_group"}`);
  });

  const selectionCount = computed(() => canvasRef.value?.selectionCount ?? 0);
  const canUndo = computed(() => canvasRef.value?.canUndo ?? false);
  const canRedo = computed(() => canvasRef.value?.canRedo ?? false);
  const zoomLevel = computed(() => canvasRef.value?.zoomLevel ?? 1);

  function refreshLayers() {
    layers.value = canvasRef.value?.getLayers() || [];
  }

  function handleCanvasUpdate(data: Record<string, unknown>) {
    // Fabric serialises only its own objects; keep the label-level settings with them.
    localCanvasData.value = withLabelBorder(data, border);
    nextTick(() => refreshLayers());
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  function handleSelectionChanged(obj: any) {
    selectedObject.value = obj;
    selectionRevision.value++;
    // Jump to the element tab on selection unless the user is working in the layer list
    if (obj && activeTab.value !== "layers") activeTab.value = "element";
  }

  const quickFields = ["name", "asset_id", "location", "location_path", "serial_number", "tags", "current_date"];

  function handleInsertField(fieldName: string) {
    const canvas = canvasRef.value?.getCanvas();
    if (!canvas) return;

    const activeObject = canvas.getActiveObject();
    if (!activeObject || activeObject.type?.toLowerCase() !== "textbox") return;

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const textbox = activeObject as any;
    const placeholder = `{{${fieldName}}}`;

    if (textbox.isEditing) {
      const selectionStart = textbox.selectionStart || 0;
      const selectionEnd = textbox.selectionEnd || 0;
      const currentText = textbox.text || "";
      textbox.set("text", currentText.substring(0, selectionStart) + placeholder + currentText.substring(selectionEnd));
      textbox.selectionStart = selectionStart + placeholder.length;
      textbox.selectionEnd = selectionStart + placeholder.length;
    } else {
      textbox.set("text", (textbox.text || "") + placeholder);
    }

    canvas.renderAll();
    canvasRef.value?.updateSelectedObject({});
  }

  function handleAlign(target: AlignTarget) {
    canvasRef.value?.alignSelected(target);
  }

  async function handleAddImage(file: File) {
    try {
      await canvasRef.value?.addImage(file);
    } catch {
      toast.error(t("components.label_template.editor.tools.image_failed"));
    }
  }

  function handleDeleteSelected() {
    canvasRef.value?.deleteSelected();
    selectedObject.value = null;
  }

  function handleUpdateObject(updates: Record<string, unknown>) {
    canvasRef.value?.updateSelectedObject(updates);
  }

  function withLayerRefresh(fn: (index: number) => void) {
    return (index: number) => {
      fn(index);
      nextTick(() => refreshLayers());
    };
  }
  const bringForward = withLayerRefresh(i => canvasRef.value?.bringForward(i));
  const sendBackward = withLayerRefresh(i => canvasRef.value?.sendBackward(i));
  const bringToFront = withLayerRefresh(i => canvasRef.value?.bringToFront(i));
  const sendToBack = withLayerRefresh(i => canvasRef.value?.sendToBack(i));
  const toggleVisibility = withLayerRefresh(i => canvasRef.value?.toggleLayerVisibility(i));

  // ---------------------------------------------------------------------------
  // Save / test print
  // ---------------------------------------------------------------------------
  async function handleSave(): Promise<boolean> {
    if (!template.value) return false;
    if (!settings.name.trim()) {
      toast.error(t("components.label_template.validation.name_required"));
      return false;
    }

    isSaving.value = true;
    try {
      await update(templateId.value, {
        id: templateId.value,
        name: settings.name.trim(),
        description: settings.description,
        width: settings.width,
        height: settings.height,
        preset: settings.preset || "",
        isShared: settings.isShared,
        outputFormat: (template.value.outputFormat as "png" | "pdf") || "png",
        dpi: settings.dpi,
        canvasData: localCanvasData.value,
        sheetLayout: sheetLayout.value,
        offsetX: offsetX.value,
        offsetY: offsetY.value,
      });
      toast.success(t("components.label_template.toast.updated"));
      await refresh();
      return true;
    } catch {
      toast.error(t("components.label_template.toast.update_failed"));
      return false;
    } finally {
      isSaving.value = false;
    }
  }

  const canTestPrint = computed(() => !!previewItem.value && labelPrinting.value);
  const testPrintHint = computed(() => {
    if (!labelPrinting.value) return t("components.label_template.editor.toast.test_print_needs_server");
    if (!previewItem.value) return t("components.label_template.editor.toast.test_print_needs_item");
    return t("components.label_template.editor.header.test_print_hint");
  });

  async function handleTestPrint() {
    if (!previewItem.value || !labelPrinting.value) return;
    isTestPrinting.value = true;
    try {
      // The printer renders the saved template, so persist pending edits first.
      if (isDirty.value) {
        const ok = await handleSave();
        if (!ok) return;
      }
      const result = await printLabels(templateId.value, [previewItem.value.id], 1);
      if (result?.success) {
        toast.success(result.message || t("components.label_template.editor.toast.test_print_sent"));
      } else {
        toast.error(result?.message || t("components.label_template.editor.toast.test_print_failed"));
      }
    } catch {
      toast.error(t("components.label_template.editor.toast.test_print_failed"));
    } finally {
      isTestPrinting.value = false;
    }
  }

  const canvasDataJson = computed(() =>
    localCanvasData.value && Object.keys(localCanvasData.value).length > 0 ? JSON.stringify(localCanvasData.value) : ""
  );

  const tabs = computed(() => [
    {
      id: "element" as const,
      label: t("components.label_template.editor.tabs.element"),
      count: undefined,
    },
    {
      id: "layers" as const,
      label: t("components.label_template.editor.tabs.layers"),
      count: layers.value.length,
    },
    {
      id: "label" as const,
      label: t("components.label_template.editor.tabs.label"),
      count: undefined,
    },
  ]);
</script>

<template>
  <BaseContainer class="max-w-none">
    <div v-if="pending" class="flex items-center justify-center py-12">
      <div class="text-muted-foreground">{{ $t("global.loading") }}</div>
    </div>

    <template v-else-if="template">
      <!-- Header -->
      <div class="mb-3 flex flex-wrap items-center gap-3 rounded-lg border bg-card px-3 py-2">
        <Button variant="ghost" size="icon" as-child>
          <NuxtLink to="/label-templates" :aria-label="$t('components.label_template.editor.header.back')">
            <MdiArrowLeft class="size-5" />
          </NuxtLink>
        </Button>
        <Input
          v-model="settings.name"
          :aria-label="$t('components.label_template.editor.header.name_label')"
          class="h-9 w-56 border-transparent text-base font-semibold shadow-none hover:border-input focus-visible:border-input"
        />
        <Badge v-if="isDirty" variant="outline" class="border-amber-300 bg-amber-50 text-amber-800">
          {{ $t("components.label_template.editor.header.unsaved") }}
        </Badge>

        <div class="ml-auto flex flex-wrap items-center gap-2">
          <Button variant="outline" size="sm" class="h-9 gap-2" @click="activeTab = 'label'">
            <span class="font-semibold">
              {{ settings.preset ? presetTitle(settings.preset) : $t("components.label_template.form.preset_custom") }}
            </span>
            <span class="text-muted-foreground">{{ sizeLabel }} · {{ settings.dpi }} dpi</span>
          </Button>
          <div
            role="group"
            :aria-label="$t('components.label_template.editor.header.units')"
            class="inline-flex h-9 overflow-hidden rounded-md border text-sm"
          >
            <button
              type="button"
              class="px-3"
              :class="unit === 'mm' ? 'bg-accent font-semibold' : 'text-muted-foreground'"
              :aria-pressed="unit === 'mm'"
              @click="unit = 'mm'"
            >
              mm
            </button>
            <button
              type="button"
              class="border-l px-3"
              :class="unit === 'in' ? 'bg-accent font-semibold' : 'text-muted-foreground'"
              :aria-pressed="unit === 'in'"
              @click="unit = 'in'"
            >
              in
            </button>
          </div>
          <Button
            variant="outline"
            size="sm"
            class="h-9"
            :disabled="!canTestPrint || isTestPrinting"
            :title="testPrintHint"
            @click="handleTestPrint"
          >
            <MdiLoading v-if="isTestPrinting" class="mr-2 size-4 animate-spin" />
            <MdiPrinterPos v-else class="mr-2 size-4" />
            {{ $t("components.label_template.editor.header.test_print") }}
          </Button>
          <Button :disabled="isSaving" class="h-9" @click="handleSave">
            <MdiContentSave class="mr-2 size-4" />
            {{ $t("global.save") }}
          </Button>
        </div>
      </div>

      <div class="flex flex-col gap-3 lg:flex-row lg:items-start">
        <!-- Tool rail -->
        <LabelTemplateEditorToolRail
          :has-selection="!!selectedObject"
          :selection-count="selectionCount"
          :can-undo="canUndo"
          :can-redo="canRedo"
          @add-text="canvasRef?.addText()"
          @add-barcode="(format, source) => canvasRef?.addBarcode(format, source)"
          @add-shape="type => canvasRef?.addShape(type)"
          @add-image="handleAddImage"
          @align="handleAlign"
          @align-selection="target => canvasRef?.alignWithinSelection(target)"
          @undo="canvasRef?.undo()"
          @redo="canvasRef?.redo()"
          @delete-selected="handleDeleteSelected"
        />

        <!-- Workspace -->
        <div class="flex min-w-0 flex-1 flex-col gap-3">
          <!--
            The slot keeps its height when nothing is selected; otherwise the canvas
            jumps down when the toolbar appears mid-drag and the text lands off the label.
          -->
          <div class="min-h-[50px]">
            <LabelTemplateEditorTextToolbar
              v-if="isTextSelected"
              :selected-object="selectedObject"
              :revision="selectionRevision"
              :unit="unit"
              @update-object="handleUpdateObject"
              @insert-field="handleInsertField"
            />
            <div
              v-else
              class="flex h-[50px] items-center rounded-lg border border-dashed px-3 text-sm text-muted-foreground"
            >
              {{ $t("components.label_template.editor.text_toolbar_hint") }}
            </div>
          </div>

          <div class="flex flex-col gap-6 rounded-lg border bg-muted/40 p-4 2xl:flex-row 2xl:items-start">
            <!-- Design -->
            <section
              class="flex min-w-0 shrink-0 flex-col gap-2"
              :aria-label="$t('components.label_template.editor.design')"
            >
              <span class="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
                {{ $t("components.label_template.editor.design") }}
              </span>
              <!-- One canvas instance: toggling the rulers must not recreate it and lose the undo history -->
              <LabelTemplateEditorCanvas
                ref="canvasRef"
                :width="settings.width"
                :height="settings.height"
                :dpi="settings.dpi"
                :canvas-data="localCanvasData"
                :snap-to-grid="display.snapToGrid"
                :show-grid="display.showGrid"
                :grid-size="display.gridSize"
                :label-border="border"
                :show-ruler="display.showRuler"
                :ruler-unit="unit === 'mm' ? 'metric' : 'imperial'"
                @update:canvas-data="handleCanvasUpdate"
                @selection-changed="handleSelectionChanged"
                @loaded="refreshLayers"
              />

              <!-- Zoom + canvas toggles -->
              <div class="flex flex-wrap items-center gap-2 text-sm">
                <Button
                  variant="outline"
                  size="icon"
                  class="size-7"
                  :disabled="zoomLevel <= 0.25"
                  :aria-label="$t('components.label_template.editor.zoom.out')"
                  @click="canvasRef?.zoomOut()"
                >
                  <MdiMinus class="size-4" />
                </Button>
                <span class="w-12 text-center font-medium">{{ Math.round(zoomLevel * 100) }}%</span>
                <Button
                  variant="outline"
                  size="icon"
                  class="size-7"
                  :disabled="zoomLevel >= 4"
                  :aria-label="$t('components.label_template.editor.zoom.in')"
                  @click="canvasRef?.zoomIn()"
                >
                  <MdiPlus class="size-4" />
                </Button>
                <Button variant="outline" size="sm" class="h-7 px-2 text-xs" @click="canvasRef?.zoomToFit()">
                  <MdiFitToScreen class="mr-1 size-3" />
                  {{ $t("components.label_template.editor.zoom.fit") }}
                </Button>
                <Button variant="outline" size="sm" class="h-7 px-2 text-xs" @click="canvasRef?.zoomTo100()">
                  100%
                </Button>
                <span class="mx-1 h-4 w-px bg-border" />
                <label class="flex items-center gap-1.5 text-xs text-muted-foreground">
                  <input v-model="display.showGrid" type="checkbox" class="size-3.5" />
                  {{ $t("components.label_template.editor.display.grid") }}
                </label>
                <label class="flex items-center gap-1.5 text-xs text-muted-foreground">
                  <input v-model="display.snapToGrid" type="checkbox" class="size-3.5" />
                  {{ $t("components.label_template.editor.display.snap") }}
                </label>
                <label class="flex items-center gap-1.5 text-xs text-muted-foreground">
                  <input v-model="display.showRuler" type="checkbox" class="size-3.5" />
                  {{ $t("components.label_template.editor.display.rulers") }}
                </label>
              </div>
            </section>

            <!-- Live print preview -->
            <LabelTemplateEditorPreviewPanel
              :template-id="templateId"
              :canvas-data="localCanvasData"
              :canvas-data-json="canvasDataJson"
              :width="settings.width"
              :height="settings.height"
              :dpi="settings.dpi"
              @update:item="previewItem = $event"
            />
          </div>
        </div>

        <!-- Inspector -->
        <aside
          class="flex w-full shrink-0 flex-col rounded-lg border bg-card lg:w-80"
          :aria-label="$t('components.label_template.editor.inspector')"
        >
          <div role="tablist" class="flex border-b text-sm">
            <button
              v-for="tab in tabs"
              :key="tab.id"
              type="button"
              role="tab"
              :aria-selected="activeTab === tab.id"
              class="flex h-10 flex-1 items-center justify-center gap-1.5 border-b-2 px-2"
              :class="
                activeTab === tab.id ? 'border-primary font-semibold' : 'border-transparent text-muted-foreground'
              "
              @click="activeTab = tab.id"
            >
              {{ tab.label }}
              <span v-if="tab.count !== undefined" class="text-xs text-muted-foreground">{{ tab.count }}</span>
            </button>
          </div>

          <div class="p-4">
            <!-- Element -->
            <div v-if="activeTab === 'element'">
              <div
                v-if="!selectedObject"
                class="rounded-md border border-dashed p-4 text-center text-sm text-muted-foreground"
              >
                {{ $t("components.label_template.editor.element.no_selection") }}
              </div>
              <div v-else class="space-y-4">
                <div class="flex items-center justify-between">
                  <span class="text-sm font-semibold">{{ selectedTypeLabel }}</span>
                  <Button
                    variant="outline"
                    size="icon"
                    class="size-7 text-destructive hover:text-destructive"
                    :aria-label="$t('components.label_template.editor.tools.delete')"
                    @click="handleDeleteSelected"
                  >
                    <MdiDelete class="size-4" />
                  </Button>
                </div>
                <LabelTemplateEditorTransformPanel
                  :selected-object="selectedObject"
                  :revision="selectionRevision"
                  :unit="unit"
                  @update-object="handleUpdateObject"
                />
                <LabelTemplateEditorBarcodePanel
                  v-if="selectedObject?.data?.type === 'barcode'"
                  :selected-object="selectedObject"
                  :revision="selectionRevision"
                  :dpi="settings.dpi"
                  @update-object="handleUpdateObject"
                />
                <div v-if="isTextSelected" class="space-y-2">
                  <span class="text-sm font-semibold">
                    {{ $t("components.label_template.editor.element.insert_field") }}
                  </span>
                  <div class="flex flex-wrap gap-1.5">
                    <Button
                      v-for="field in quickFields"
                      :key="field"
                      variant="secondary"
                      size="sm"
                      class="h-7 rounded-full px-2.5 text-xs"
                      @click="handleInsertField(field)"
                    >
                      +
                      {{ $t(`components.label_template.editor.data_fields.${field}`) }}
                    </Button>
                  </div>
                </div>
              </div>
            </div>

            <!-- Layers -->
            <LabelTemplateEditorLayerPanel
              v-else-if="activeTab === 'layers'"
              :layers="layers"
              :selected-object="selectedObject"
              @select-layer="canvasRef?.selectLayerByIndex($event)"
              @bring-forward="bringForward"
              @send-backward="sendBackward"
              @bring-to-front="bringToFront"
              @send-to-back="sendToBack"
              @toggle-visibility="toggleVisibility"
            />

            <!-- Label -->
            <div v-else class="space-y-5">
              <LabelTemplateEditorLabelSettings
                v-model="settings"
                v-model:display="display"
                v-model:border="border"
                :presets="presets ?? []"
              />
              <div class="border-t pt-4">
                <LabelTemplateEditorSheetLayoutPanel
                  v-model:sheet-layout="sheetLayout"
                  v-model:offset-x="offsetX"
                  v-model:offset-y="offsetY"
                  :preset="selectedPreset"
                  :label-width="settings.width"
                  :label-height="settings.height"
                />
              </div>
            </div>
          </div>
        </aside>
      </div>
    </template>

    <div v-else class="flex flex-col items-center justify-center py-12">
      <p class="text-muted-foreground">
        {{ $t("components.label_template.editor.not_found") }}
      </p>
      <Button variant="outline" class="mt-4" as-child>
        <NuxtLink to="/label-templates">{{ $t("components.label_template.editor.header.back") }}</NuxtLink>
      </Button>
    </div>
  </BaseContainer>
</template>
