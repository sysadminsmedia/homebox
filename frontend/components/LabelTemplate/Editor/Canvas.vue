<script setup lang="ts">
  import { onMounted, onBeforeUnmount, watch, nextTick } from "vue";
  import { ActiveSelection, Canvas, IText, Textbox, Rect, Line, Group, FabricImage, FabricObject } from "fabric";
  import type { AlignTarget } from "./ToolRail.vue";
  import { MM_TO_PX, type LabelBorder } from "./label-border";

  // Fabric 7 anchors new objects at their centre by default. The server
  // renderer and the transform panel work in top-left coordinates, so keep
  // every object created in the editor on a left/top origin.
  FabricObject.ownDefaults.originX = "left";
  FabricObject.ownDefaults.originY = "top";

  // Type for textbox data properties
  interface TextBoxData {
    autofit?: boolean;
    maxFontSize?: number;
    fixedWidth?: number | null;
    fixedHeight?: number | null;
  }

  const props = defineProps<{
    width: number;
    height: number;
    dpi: number;
    canvasData?: Record<string, unknown>;
    snapToGrid?: boolean;
    gridSize?: number;
    showGrid?: boolean;
    labelBorder?: LabelBorder;
  }>();

  const emit = defineEmits<{
    "update:canvasData": [data: Record<string, unknown>];
    selectionChanged: [object: FabricObject | null];
    loaded: [];
  }>();

  const canvasRef = ref<HTMLCanvasElement | null>(null);
  const gridCanvasRef = ref<HTMLCanvasElement | null>(null);
  const canvasContainerRef = ref<HTMLDivElement | null>(null);
  let fabricCanvas: Canvas | null = null;
  let isLoadingData = false; // Prevent save during load
  const hasInitiallyLoaded = ref(false); // Track if initial data was loaded
  const selectionCount = ref(0);

  // Zoom controls
  const zoomLevel = ref(1);
  const MIN_ZOOM = 0.25;
  const MAX_ZOOM = 4;
  const ZOOM_STEP = 0.25;

  function setZoom(level: number) {
    zoomLevel.value = Math.max(MIN_ZOOM, Math.min(MAX_ZOOM, level));
    if (fabricCanvas) {
      fabricCanvas.setZoom(zoomLevel.value);
      fabricCanvas.setDimensions({
        width: displayWidth.value * zoomLevel.value,
        height: displayHeight.value * zoomLevel.value,
      });
      fabricCanvas.renderAll();
    }
    drawGrid();
  }

  function zoomIn() {
    setZoom(zoomLevel.value + ZOOM_STEP);
  }

  function zoomOut() {
    setZoom(zoomLevel.value - ZOOM_STEP);
  }

  function zoomToFit() {
    if (!canvasContainerRef.value) {
      setZoom(1);
      return;
    }

    const containerWidth = canvasContainerRef.value.clientWidth - 40;
    const containerHeight = Math.max(canvasContainerRef.value.clientHeight - 40, 200);

    const scaleX = containerWidth / displayWidth.value;
    const scaleY = containerHeight / displayHeight.value;
    const scale = Math.min(scaleX, scaleY, MAX_ZOOM);

    setZoom(Math.max(MIN_ZOOM, scale));
  }

  function zoomTo100() {
    setZoom(1);
  }

  // Default grid size in pixels (10 pixels = approx 2.6mm at 96 DPI)
  const effectiveGridSize = computed(() => props.gridSize || 10);

  // Convert mm to pixels for display (at 96 DPI for screen)
  const screenDPI = 96;
  const displayWidth = computed(() => Math.round((props.width / 25.4) * screenDPI));
  const displayHeight = computed(() => Math.round((props.height / 25.4) * screenDPI));

  // ---------------------------------------------------------------------------
  // History (undo / redo). Each entry is the serialized canvas; the current
  // state is history[historyIndex].
  // ---------------------------------------------------------------------------
  const history = ref<string[]>([]);
  const historyIndex = ref(-1);
  const MAX_HISTORY = 50;
  let isRestoringHistory = false;

  const canUndo = computed(() => historyIndex.value > 0);
  const canRedo = computed(() => historyIndex.value >= 0 && historyIndex.value < history.value.length - 1);

  function pushHistory(json: Record<string, unknown>) {
    if (isRestoringHistory) return;
    const serialized = JSON.stringify(json);
    if (history.value[historyIndex.value] === serialized) return;
    const next = history.value.slice(0, historyIndex.value + 1);
    next.push(serialized);
    if (next.length > MAX_HISTORY) next.shift();
    history.value = next;
    historyIndex.value = next.length - 1;
  }

  async function restoreHistory(index: number) {
    if (!fabricCanvas) return;
    const serialized = history.value[index];
    if (!serialized) return;
    isRestoringHistory = true;
    isLoadingData = true;
    try {
      const data = JSON.parse(serialized) as Record<string, unknown>;
      await fabricCanvas.loadFromJSON(data);
      fabricCanvas.discardActiveObject();
      fabricCanvas.renderAll();
      historyIndex.value = index;
      emit("update:canvasData", data);
      emit("selectionChanged", null);
    } finally {
      isLoadingData = false;
      isRestoringHistory = false;
      nextTick(() => drawGrid());
    }
  }

  function undo() {
    if (canUndo.value) restoreHistory(historyIndex.value - 1);
  }

  function redo() {
    if (canRedo.value) restoreHistory(historyIndex.value + 1);
  }

  function initCanvas() {
    if (!canvasRef.value) return;

    fabricCanvas = new Canvas(canvasRef.value, {
      width: displayWidth.value,
      height: displayHeight.value,
      backgroundColor: "transparent", // Grid canvas provides the white background
      selection: true,
      preserveObjectStacking: true,
    });

    // Load existing canvas data if provided
    if (props.canvasData && Object.keys(props.canvasData).length > 0) {
      hasInitiallyLoaded.value = true;
      isLoadingData = true;
      fabricCanvas.loadFromJSON(props.canvasData).then(() => {
        fabricCanvas?.renderAll();
        isLoadingData = false;
        onInitialLoad();
      });
    } else {
      nextTick(() => onInitialLoad());
    }

    // Listen for changes
    fabricCanvas.on("object:modified", saveCanvas);
    fabricCanvas.on("object:added", () => {
      saveCanvas();
      nextTick(() => drawGrid());
    });
    fabricCanvas.on("object:removed", () => {
      saveCanvas();
      nextTick(() => drawGrid());
    });

    // Listen for selection changes
    fabricCanvas.on("selection:created", e => {
      selectionCount.value = fabricCanvas?.getActiveObjects().length ?? 0;
      emit("selectionChanged", e.selected?.[0] || null);
    });
    fabricCanvas.on("selection:updated", e => {
      selectionCount.value = fabricCanvas?.getActiveObjects().length ?? 0;
      emit("selectionChanged", e.selected?.[0] || null);
    });
    fabricCanvas.on("selection:cleared", () => {
      selectionCount.value = 0;
      emit("selectionChanged", null);
    });

    // Snap to grid on object moving
    fabricCanvas.on("object:moving", e => {
      if (!props.snapToGrid || !e.target) return;

      const obj = e.target;
      const gridSize = effectiveGridSize.value;

      obj.set({
        left: Math.round((obj.left || 0) / gridSize) * gridSize,
        top: Math.round((obj.top || 0) / gridSize) * gridSize,
      });
    });

    // Handle textbox scaling - capture the final dimensions and reset scale
    fabricCanvas.on("object:scaling", e => {
      const obj = e.target as Textbox & { data?: TextBoxData };
      if (obj?.type !== "textbox") return;

      const scaleX = obj.scaleX || 1;
      const scaleY = obj.scaleY || 1;
      const newWidth = Math.round((obj.width || 150) * scaleX);
      const newHeight = Math.round((obj.height || 50) * scaleY);

      obj.set({
        width: newWidth,
        scaleX: 1,
        scaleY: 1,
      });

      obj.data = {
        ...obj.data,
        fixedWidth: newWidth,
        fixedHeight: newHeight,
      };

      if (obj.data?.autofit) {
        applyAutofit(obj);
      }
    });

    // Autofit text handler - adjust font to fit within bounds
    fabricCanvas.on("text:changed", e => {
      if (!fabricCanvas) return;

      const textObj = e.target as Textbox & { data?: TextBoxData };
      if (!textObj?.data?.autofit) return;

      applyAutofit(textObj);
      fabricCanvas.renderAll();
    });
  }

  // Apply autofit to a textbox - adjust font size to fit within fixed bounds
  function applyAutofit(textObj: Textbox & { data?: TextBoxData }) {
    if (!textObj?.data?.autofit) return;

    const maxWidth = textObj.data.fixedWidth || textObj.width || 150;
    const maxHeight = textObj.data.fixedHeight || 100;
    const maxFontSize = textObj.data.maxFontSize || 72;
    const minFontSize = 6;

    textObj.set({ width: maxWidth });

    const textOverflows = () => {
      if ((textObj.height || 0) > maxHeight) return true;

      const textLines = textObj.textLines || [];
      const ctx = textObj.canvas?.getContext();
      if (ctx && textLines.length > 0) {
        ctx.font = `${textObj.fontSize}px ${textObj.fontFamily}`;
        for (const line of textLines) {
          const lineWidth = ctx.measureText(line).width;
          if (lineWidth > maxWidth * 0.98) {
            return true;
          }
        }
      }
      return false;
    };

    while (textOverflows() && (textObj.fontSize || 16) > minFontSize) {
      textObj.set("fontSize", (textObj.fontSize || 16) - 1);
    }

    while ((textObj.fontSize || 16) < maxFontSize) {
      textObj.set("fontSize", (textObj.fontSize || 16) + 1);
      if (textOverflows()) {
        textObj.set("fontSize", (textObj.fontSize || 16) - 1);
        break;
      }
    }
  }

  // Draw grid on the grid canvas (also serves as background)
  function drawGrid() {
    if (!gridCanvasRef.value) return;

    const ctx = gridCanvasRef.value.getContext("2d");
    if (!ctx) return;

    const zoom = zoomLevel.value;
    const width = displayWidth.value * zoom;
    const height = displayHeight.value * zoom;
    const gridSize = effectiveGridSize.value * zoom;

    gridCanvasRef.value.width = width;
    gridCanvasRef.value.height = height;

    ctx.fillStyle = "#ffffff";
    ctx.fillRect(0, 0, width, height);

    if (props.showGrid) {
      ctx.strokeStyle = "#e0e0e0";
      ctx.lineWidth = 0.5;

      for (let x = 0; x <= width; x += gridSize) {
        ctx.beginPath();
        ctx.moveTo(x, 0);
        ctx.lineTo(x, height);
        ctx.stroke();
      }

      for (let y = 0; y <= height; y += gridSize) {
        ctx.beginPath();
        ctx.moveTo(0, y);
        ctx.lineTo(width, y);
        ctx.stroke();
      }
    }

    drawLabelBorder(ctx, width, height, zoom);
  }

  // The label border is not a canvas object: it is painted on the background
  // layer exactly where the server renderer draws it, behind every element.
  function drawLabelBorder(ctx: CanvasRenderingContext2D, width: number, height: number, zoom: number) {
    const border = props.labelBorder;
    if (!border?.enabled) return;
    const lineWidth = Math.max(0.5, border.width * MM_TO_PX * zoom);
    const inset = Math.max(0, border.inset * MM_TO_PX * zoom);
    const w = width - inset * 2 - lineWidth;
    const h = height - inset * 2 - lineWidth;
    if (w <= 0 || h <= 0) return;
    ctx.strokeStyle = border.color || "#000000";
    ctx.lineWidth = lineWidth;
    ctx.strokeRect(inset + lineWidth / 2, inset + lineWidth / 2, w, h);
  }

  function saveCanvas() {
    if (!fabricCanvas || isLoadingData) return;
    const json = fabricCanvas.toObject(["data"]) as Record<string, unknown>;
    pushHistory(json);
    emit("update:canvasData", json);
  }

  function addText(text: string = "Text", options: Record<string, unknown> = {}) {
    if (!fabricCanvas) return;

    const defaultWidth = 150;
    const defaultHeight = 30;

    const textObj = new Textbox(text, {
      left: 20,
      top: 20,
      width: defaultWidth,
      fontSize: 16,
      fontFamily: "Arial",
      fill: "#000000",
      textAlign: "left",
      splitByGrapheme: false,
      data: {
        autofit: false,
        maxFontSize: 16,
        fixedWidth: defaultWidth,
        fixedHeight: defaultHeight,
      },
      ...options,
    } as ConstructorParameters<typeof Textbox>[1] & { data: TextBoxData });

    fabricCanvas.add(textObj);
    fabricCanvas.setActiveObject(textObj);
    fabricCanvas.renderAll();
    saveCanvas();
  }

  function addBarcode(format: string, contentSource: string) {
    if (!fabricCanvas) return;

    const rect = new Rect({
      left: 0,
      top: 0,
      width: 100,
      height: 100,
      fill: "#f0f0f0",
      stroke: "#cccccc",
      strokeWidth: 1,
    });

    const label = new IText(`[${format.toUpperCase()}]`, {
      left: 5,
      top: 40,
      fontSize: 12,
      fontFamily: "Arial",
      fill: "#666666",
      selectable: false,
      evented: false,
    });

    const group = new Group([rect, label], {
      left: 20,
      top: 20,
      data: {
        type: "barcode",
        format,
        contentSource,
      },
    } as ConstructorParameters<typeof Group>[1] & {
      data: Record<string, unknown>;
    });

    fabricCanvas.add(group);
    fabricCanvas.setActiveObject(group);
    fabricCanvas.renderAll();
    saveCanvas();
  }

  function addShape(shapeType: "rect" | "line") {
    if (!fabricCanvas) return;

    let shape: FabricObject;

    if (shapeType === "rect") {
      // A new rectangle frames the label with a 1 mm inset, so a border is one
      // click; it can still be dragged smaller.
      const inset = Math.round(MM_TO_PX);
      shape = new Rect({
        left: inset,
        top: inset,
        width: Math.max(10, displayWidth.value - inset * 2),
        height: Math.max(10, displayHeight.value - inset * 2),
        fill: "transparent",
        stroke: "#000000",
        strokeWidth: 1,
      });
    } else {
      shape = new Line([0, 0, 100, 0], {
        left: 20,
        top: 20,
        stroke: "#000000",
        strokeWidth: 1,
      });
    }

    fabricCanvas.add(shape);
    fabricCanvas.setActiveObject(shape);
    fabricCanvas.renderAll();
    saveCanvas();
  }

  // Embed a raster image. The data URL travels inside the template JSON and is
  // decoded by the server renderer, so there is nothing to upload separately.
  async function addImage(file: File) {
    if (!fabricCanvas) return;

    const dataUrl = await new Promise<string>((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => resolve(String(reader.result));
      reader.onerror = () => reject(reader.error);
      reader.readAsDataURL(file);
    });

    const img = await FabricImage.fromURL(dataUrl);
    // Fit into 40% of the label, keeping aspect ratio
    const maxW = displayWidth.value * 0.4;
    const maxH = displayHeight.value * 0.8;
    const scale = Math.min(maxW / (img.width || 1), maxH / (img.height || 1), 1);
    img.set({ left: 20, top: 20, scaleX: scale, scaleY: scale });

    fabricCanvas.add(img);
    fabricCanvas.setActiveObject(img);
    fabricCanvas.renderAll();
    saveCanvas();
  }

  function deleteSelected() {
    if (!fabricCanvas) return;

    const active = fabricCanvas.getActiveObjects();
    if (active.length === 0) return;
    for (const obj of active) {
      fabricCanvas.remove(obj);
    }
    fabricCanvas.discardActiveObject();
    fabricCanvas.renderAll();
    saveCanvas();
  }

  interface Bounds {
    left: number;
    top: number;
    width: number;
    height: number;
  }

  // Move obj so that its bounding box sits against the given edge or centre
  // line of bounds. Works for rotated objects because it goes through the
  // bounding rect rather than left/top directly.
  function alignObjectTo(obj: FabricObject, target: AlignTarget, bounds: Bounds) {
    const rect = obj.getBoundingRect();
    const dx = (obj.left || 0) - rect.left;
    const dy = (obj.top || 0) - rect.top;

    switch (target) {
      case "left":
        obj.set({ left: bounds.left + dx });
        break;
      case "center":
        obj.set({ left: bounds.left + (bounds.width - rect.width) / 2 + dx });
        break;
      case "right":
        obj.set({ left: bounds.left + bounds.width - rect.width + dx });
        break;
      case "top":
        obj.set({ top: bounds.top + dy });
        break;
      case "middle":
        obj.set({ top: bounds.top + (bounds.height - rect.height) / 2 + dy });
        break;
      case "bottom":
        obj.set({ top: bounds.top + bounds.height - rect.height + dy });
        break;
    }
    obj.setCoords();
  }

  // Align the selected object (or a whole multi-selection) against the label edges / center.
  function alignSelected(target: AlignTarget) {
    if (!fabricCanvas) return;
    const obj = fabricCanvas.getActiveObject();
    if (!obj) return;

    alignObjectTo(obj, target, {
      left: 0,
      top: 0,
      width: displayWidth.value,
      height: displayHeight.value,
    });
    fabricCanvas.renderAll();
    saveCanvas();
    emit("selectionChanged", obj);
  }

  // Align every object of a multi-selection against the selection's own bounds,
  // e.g. left edges flush or centred on a common line.
  function alignWithinSelection(target: AlignTarget) {
    if (!fabricCanvas) return;
    const active = fabricCanvas.getActiveObject();
    if (!(active instanceof ActiveSelection)) return;
    const objects = active.getObjects();
    if (objects.length < 2) return;

    const bounds = active.getBoundingRect();
    // Objects inside a selection carry selection-relative coordinates; leave the
    // selection so each one can be moved in label coordinates, then reselect.
    fabricCanvas.discardActiveObject();
    for (const obj of objects) {
      alignObjectTo(obj, target, bounds);
    }
    const reselected = new ActiveSelection(objects, { canvas: fabricCanvas });
    fabricCanvas.setActiveObject(reselected);
    fabricCanvas.requestRenderAll();
    saveCanvas();
    emit("selectionChanged", objects[0] ?? null);
  }

  function nudgeSelected(dx: number, dy: number) {
    if (!fabricCanvas) return;
    const obj = fabricCanvas.getActiveObject();
    if (!obj) return;
    obj.set({ left: (obj.left || 0) + dx, top: (obj.top || 0) + dy });
    obj.setCoords();
    fabricCanvas.renderAll();
    saveCanvas();
    emit("selectionChanged", obj);
  }

  function updateSelectedObject(updates: Record<string, unknown>) {
    if (!fabricCanvas) return;

    const activeObject = fabricCanvas.getActiveObject();
    if (activeObject) {
      if ("data" in updates) {
        (activeObject as FabricObject & { data?: Record<string, unknown> }).data = updates.data as Record<
          string,
          unknown
        >;
        const { data: _, ...restUpdates } = updates;
        if (Object.keys(restUpdates).length > 0) {
          activeObject.set(restUpdates as Partial<FabricObject>);
        }
      } else {
        activeObject.set(updates as Partial<FabricObject>);
      }
      activeObject.setCoords();
      fabricCanvas.renderAll();
      saveCanvas();
      emit("selectionChanged", activeObject);
    }
  }

  function getCanvas() {
    return fabricCanvas;
  }

  // Layer management methods
  interface LayerItem {
    index: number;
    type: string;
    name: string;
    visible: boolean;
    object: FabricObject;
  }

  function getLayers(): LayerItem[] {
    if (!fabricCanvas) return [];

    const objects = fabricCanvas.getObjects();
    return objects
      .map((obj, index) => ({
        index,
        type: obj.type || "object",
        name: (obj as FabricObject & { text?: string }).text || obj.type || "Object",
        visible: obj.visible !== false,
        object: obj,
      }))
      .reverse();
  }

  function selectLayerByIndex(index: number) {
    if (!fabricCanvas) return;

    const obj = fabricCanvas.getObjects()[index];
    if (obj) {
      fabricCanvas.setActiveObject(obj);
      fabricCanvas.renderAll();
      emit("selectionChanged", obj);
    }
  }

  function bringForward(index: number) {
    if (!fabricCanvas) return;
    const objects = fabricCanvas.getObjects();
    const obj = objects[index];
    if (obj && index < objects.length - 1) {
      fabricCanvas.bringObjectForward(obj);
      fabricCanvas.renderAll();
      saveCanvas();
    }
  }

  function sendBackward(index: number) {
    if (!fabricCanvas) return;
    const obj = fabricCanvas.getObjects()[index];
    if (obj && index > 0) {
      fabricCanvas.sendObjectBackwards(obj);
      fabricCanvas.renderAll();
      saveCanvas();
    }
  }

  function bringToFront(index: number) {
    if (!fabricCanvas) return;
    const objects = fabricCanvas.getObjects();
    const obj = objects[index];
    if (obj && index < objects.length - 1) {
      fabricCanvas.bringObjectToFront(obj);
      fabricCanvas.renderAll();
      saveCanvas();
    }
  }

  function sendToBack(index: number) {
    if (!fabricCanvas) return;
    const obj = fabricCanvas.getObjects()[index];
    if (obj && index > 0) {
      fabricCanvas.sendObjectToBack(obj);
      fabricCanvas.renderAll();
      saveCanvas();
    }
  }

  function toggleLayerVisibility(index: number) {
    if (!fabricCanvas) return;
    const obj = fabricCanvas.getObjects()[index];
    if (obj) {
      obj.visible = !obj.visible;
      fabricCanvas.renderAll();
      saveCanvas();
    }
  }

  defineExpose({
    addText,
    addBarcode,
    addShape,
    addImage,
    deleteSelected,
    alignSelected,
    alignWithinSelection,
    selectionCount,
    updateSelectedObject,
    getCanvas,
    getLayers,
    selectLayerByIndex,
    bringForward,
    sendBackward,
    bringToFront,
    sendToBack,
    toggleLayerVisibility,
    undo,
    redo,
    canUndo,
    canRedo,
    zoomIn,
    zoomOut,
    zoomToFit,
    zoomTo100,
    zoomLevel,
    displayWidth,
    displayHeight,
  });

  function isTypingTarget(target: EventTarget | null): boolean {
    const el = target as HTMLElement | null;
    if (!el) return false;
    const tag = el.tagName?.toLowerCase();
    return tag === "input" || tag === "textarea" || tag === "select" || el.isContentEditable;
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (!fabricCanvas) return;
    if (isTypingTarget(e.target)) return;

    const activeObject = fabricCanvas.getActiveObject();
    if (activeObject && (activeObject as { isEditing?: boolean }).isEditing) {
      return;
    }

    const mod = e.metaKey || e.ctrlKey;
    if (mod && e.key.toLowerCase() === "z") {
      e.preventDefault();
      if (e.shiftKey) redo();
      else undo();
      return;
    }
    if (mod && e.key.toLowerCase() === "y") {
      e.preventDefault();
      redo();
      return;
    }

    if (e.key === "Delete" || e.key === "Backspace") {
      e.preventDefault();
      deleteSelected();
      return;
    }

    const step = e.shiftKey ? 10 : 1;
    switch (e.key) {
      case "ArrowLeft":
        e.preventDefault();
        nudgeSelected(-step, 0);
        break;
      case "ArrowRight":
        e.preventDefault();
        nudgeSelected(step, 0);
        break;
      case "ArrowUp":
        e.preventDefault();
        nudgeSelected(0, -step);
        break;
      case "ArrowDown":
        e.preventDefault();
        nudgeSelected(0, step);
        break;
    }
  }

  onMounted(() => {
    nextTick(() => {
      initCanvas();
      drawGrid();
    });

    window.addEventListener("keydown", handleKeyDown);
  });

  onBeforeUnmount(() => {
    window.removeEventListener("keydown", handleKeyDown);

    if (fabricCanvas) {
      fabricCanvas.dispose();
      fabricCanvas = null;
    }
  });

  // Watch for dimension changes
  watch([displayWidth, displayHeight], () => {
    if (fabricCanvas) {
      fabricCanvas.setZoom(zoomLevel.value);
      fabricCanvas.setDimensions({
        width: displayWidth.value * zoomLevel.value,
        height: displayHeight.value * zoomLevel.value,
      });
      fabricCanvas.renderAll();
    }
    drawGrid();
  });

  // Watch for grid and border settings changes
  watch(
    () => [props.showGrid, props.gridSize, props.labelBorder],
    () => {
      drawGrid();
    },
    { deep: true }
  );

  // Watch for canvas data changes (e.g., when template loads asynchronously)
  // Only trigger on initial load, not on every change
  watch(
    () => props.canvasData,
    newData => {
      if (fabricCanvas && newData && Object.keys(newData).length > 0 && !hasInitiallyLoaded.value) {
        hasInitiallyLoaded.value = true;
        isLoadingData = true;
        fabricCanvas.loadFromJSON(newData).then(() => {
          fabricCanvas?.renderAll();
          isLoadingData = false;
          onInitialLoad();
        });
      }
    },
    { immediate: true }
  );

  // After the template's objects are on the canvas: seed the history, fit the
  // label into the viewport (a 62x29mm label is only ~234px wide at 100%) and
  // let the parent refresh its layer list.
  function onInitialLoad() {
    if (fabricCanvas) {
      const json = fabricCanvas.toObject(["data"]) as Record<string, unknown>;
      history.value = [JSON.stringify(json)];
      historyIndex.value = 0;
    }
    nextTick(() => {
      zoomToFit();
      emit("loaded");
    });
  }
</script>

<template>
  <div ref="canvasContainerRef" class="min-h-[300px] overflow-auto rounded border bg-gray-100 p-4 dark:bg-zinc-800">
    <div class="relative inline-block shadow-lg">
      <!--
        Grid layer (behind fabric canvas). drawGrid sizes it; binding width/height
        here as well would let Vue reset the element, and with it the drawing,
        on the render that follows every zoom change.
      -->
      <canvas ref="gridCanvasRef" class="pointer-events-none absolute left-0 top-0" :style="{ zIndex: 0 }" />
      <!-- Fabric canvas (front, with transparent background area) -->
      <canvas ref="canvasRef" class="relative" :style="{ zIndex: 1 }" />
    </div>
  </div>
</template>
