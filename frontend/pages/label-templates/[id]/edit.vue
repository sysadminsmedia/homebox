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
  import LabelTemplateEditorTransformPanel, {
    type EditorUnit,
  } from "~/components/LabelTemplate/Editor/TransformPanel.vue";
  import LabelTemplateEditorLayerPanel from "~/components/LabelTemplate/Editor/LayerPanel.vue";
  import LabelTemplateEditorRuler from "~/components/LabelTemplate/Editor/Ruler.vue";
  import LabelTemplateEditorPreviewPanel from "~/components/LabelTemplate/Editor/PreviewPanel.vue";
  import LabelTemplateEditorLabelSettings, {
    type EditorSettings,
    type DisplaySettings,
  } from "~/components/LabelTemplate/Editor/LabelSettings.vue";
  import type { EntitySummary } from "~/lib/api/types/data-contracts";

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
  const { defaultPrinter } = useQuickPrint();
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
    mediaType: "",
    isShared: false,
  });

  const display = reactive<DisplaySettings>({
    showRuler: true,
    showGrid: false,
    snapToGrid: false,
    gridSize: 10,
  });

  const unit = ref<EditorUnit>("mm");
  const localCanvasData = ref<Record<string, unknown>>({});

  function loadFromTemplate() {
    const tpl = template.value;
    if (!tpl) return;
    settings.name = tpl.name;
    settings.description = tpl.description ?? "";
    settings.width = tpl.width;
    settings.height = tpl.height;
    settings.preset = tpl.preset || "";
    settings.dpi = tpl.dpi;
    settings.mediaType = tpl.mediaType || "";
    settings.isShared = tpl.isShared;
    if (tpl.canvasData) localCanvasData.value = tpl.canvasData;
  }

  watch(template, loadFromTemplate, { immediate: true });

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
      settings.mediaType !== (tpl.mediaType || "") ||
      settings.isShared !== tpl.isShared;
    const canvasChanged = JSON.stringify(localCanvasData.value) !== JSON.stringify(tpl.canvasData ?? {});
    return settingsChanged || canvasChanged;
  });

  useEventListener(window, "beforeunload", (e: BeforeUnloadEvent) => {
    if (isDirty.value) e.preventDefault();
  });

  const sizeLabel = computed(() => {
    if (unit.value === "in") {
      return `${(settings.width / 25.4).toFixed(2)} × ${(settings.height / 25.4).toFixed(2)} in`;
    }
    return formatLabelSize(settings.width, settings.height);
  });

  // ---------------------------------------------------------------------------
  // Canvas interaction
  // ---------------------------------------------------------------------------
  const canvasRef = ref<InstanceType<typeof LabelTemplateEditorCanvas> | null>(null);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const selectedObject = ref<any>(null);
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

  const canUndo = computed(() => canvasRef.value?.canUndo ?? false);
  const canRedo = computed(() => canvasRef.value?.canRedo ?? false);
  const zoomLevel = computed(() => canvasRef.value?.zoomLevel ?? 1);

  // Ruler pixel sizes follow the canvas zoom
  const screenDPI = 96;
  const pixelWidth = computed(() => Math.round((settings.width / 25.4) * screenDPI * zoomLevel.value));
  const pixelHeight = computed(() => Math.round((settings.height / 25.4) * screenDPI * zoomLevel.value));

  function refreshLayers() {
    layers.value = canvasRef.value?.getLayers() || [];
  }

  function handleCanvasUpdate(data: Record<string, unknown>) {
    localCanvasData.value = data;
    nextTick(() => refreshLayers());
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  function handleSelectionChanged(obj: any) {
    selectedObject.value = obj;
    // Jump to the element tab on selection unless the user is working in the layer list
    if (obj && activeTab.value !== "layers") activeTab.value = "element";
  }

  const quickFields = ["item_name", "asset_id", "location", "location_path", "serial_number", "tags", "current_date"];

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
        mediaType: settings.mediaType || null,
        canvasData: localCanvasData.value,
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

  const canTestPrint = computed(() => !!previewItem.value && !!defaultPrinter.value);
  const testPrintHint = computed(() => {
    if (!defaultPrinter.value) return t("components.label_template.editor.toast.test_print_needs_printer");
    if (!previewItem.value) return t("components.label_template.editor.toast.test_print_needs_item");
    return t("components.label_template.editor.header.test_print_hint", { printer: defaultPrinter.value.name });
  });

  async function handleTestPrint() {
    if (!previewItem.value || !defaultPrinter.value) return;
    isTestPrinting.value = true;
    try {
      // The printer renders the saved template, so persist pending edits first.
      if (isDirty.value) {
        const ok = await handleSave();
        if (!ok) return;
      }
      const result = await printLabels(templateId.value, [previewItem.value.id], defaultPrinter.value.id, 1);
      if (result?.success) {
        toast.success(t("components.label_template.editor.toast.test_print_sent", { printer: result.printerName }));
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
    { id: "element" as const, label: t("components.label_template.editor.tabs.element"), count: undefined },
    { id: "layers" as const, label: t("components.label_template.editor.tabs.layers"), count: layers.value.length },
    { id: "label" as const, label: t("components.label_template.editor.tabs.label"), count: undefined },
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
          :can-undo="canUndo"
          :can-redo="canRedo"
          @add-text="canvasRef?.addText()"
          @add-barcode="(format, source) => canvasRef?.addBarcode(format, source)"
          @add-shape="type => canvasRef?.addShape(type)"
          @add-image="handleAddImage"
          @align="handleAlign"
          @undo="canvasRef?.undo()"
          @redo="canvasRef?.redo()"
          @delete-selected="handleDeleteSelected"
        />

        <!-- Workspace -->
        <div class="flex min-w-0 flex-1 flex-col gap-3">
          <LabelTemplateEditorTextToolbar
            v-if="isTextSelected"
            :selected-object="selectedObject"
            @update-object="handleUpdateObject"
            @insert-field="handleInsertField"
          />

          <div class="flex flex-col gap-6 rounded-lg border bg-muted/40 p-4 2xl:flex-row 2xl:items-start">
            <!-- Design -->
            <section
              class="flex min-w-0 shrink-0 flex-col gap-2"
              :aria-label="$t('components.label_template.editor.design')"
            >
              <span class="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
                {{ $t("components.label_template.editor.design") }}
              </span>
              <LabelTemplateEditorRuler
                v-if="display.showRuler"
                :width="settings.width"
                :height="settings.height"
                :pixel-width="pixelWidth"
                :pixel-height="pixelHeight"
                :unit="unit === 'mm' ? 'metric' : 'imperial'"
              >
                <LabelTemplateEditorCanvas
                  ref="canvasRef"
                  :width="settings.width"
                  :height="settings.height"
                  :dpi="settings.dpi"
                  :canvas-data="localCanvasData"
                  :snap-to-grid="display.snapToGrid"
                  :show-grid="display.showGrid"
                  :grid-size="display.gridSize"
                  @update:canvas-data="handleCanvasUpdate"
                  @selection-changed="handleSelectionChanged"
                  @loaded="refreshLayers"
                />
              </LabelTemplateEditorRuler>
              <LabelTemplateEditorCanvas
                v-else
                ref="canvasRef"
                :width="settings.width"
                :height="settings.height"
                :dpi="settings.dpi"
                :canvas-data="localCanvasData"
                :snap-to-grid="display.snapToGrid"
                :show-grid="display.showGrid"
                :grid-size="display.gridSize"
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
                  :unit="unit"
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
                      + {{ $t(`components.label_template.editor.data_fields.${field}`) }}
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
            <LabelTemplateEditorLabelSettings
              v-else
              v-model="settings"
              v-model:display="display"
              :presets="presets ?? []"
            />
          </div>
        </aside>
      </div>
    </template>

    <div v-else class="flex flex-col items-center justify-center py-12">
      <p class="text-muted-foreground">{{ $t("components.label_template.editor.not_found") }}</p>
      <Button variant="outline" class="mt-4" as-child>
        <NuxtLink to="/label-templates">{{ $t("components.label_template.editor.header.back") }}</NuxtLink>
      </Button>
    </div>
  </BaseContainer>
</template>
