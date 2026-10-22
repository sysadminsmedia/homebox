<script setup lang="ts">
  import { onMounted, onBeforeUnmount, watch, nextTick } from "vue";
  import { Canvas, IText, Textbox, Rect, Line, Group, type FabricObject } from "fabric";

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
  }>();

  const emit = defineEmits<{
    "update:canvasData": [data: Record<string, unknown>];
    selectionChanged: [object: FabricObject | null];
  }>();

  const canvasRef = ref<HTMLCanvasElement | null>(null);
  const gridCanvasRef = ref<HTMLCanvasElement | null>(null);
  const canvasContainerRef = ref<HTMLDivElement | null>(null);
  let fabricCanvas: Canvas | null = null;
  let isLoadingData = false; // Prevent save during load
  const hasInitiallyLoaded = ref(false); // Track if initial data was loaded

  // Zoom controls
  const zoomLevel = ref(1);
  const MIN_ZOOM = 0.25;
  const MAX_ZOOM = 4;
  const ZOOM_STEP = 0.25;

  function setZoom(level: number) {
    zoomLevel.value = Math.max(MIN_ZOOM, Math.min(MAX_ZOOM, level));
    if (fabricCanvas) {
      fabricCanvas.setZoom(zoomLevel.value);
      // Update canvas viewport dimensions to account for zoom
      fabricCanvas.setDimensions({
        width: displayWidth.value * zoomLevel.value,
        height: displayHeight.value * zoomLevel.value,
      });
      fabricCanvas.renderAll();
    }
    // Redraw grid at new zoom level
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

    const containerWidth = canvasContainerRef.value.clientWidth - 40; // Account for padding
    const containerHeight = canvasContainerRef.value.clientHeight - 40;

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

  function initCanvas() {
    if (!canvasRef.value) return;

    fabricCanvas = new Canvas(canvasRef.value, {
      width: displayWidth.value,
      height: displayHeight.value,
      backgroundColor: "transparent", // Grid canvas provides the white background
      selection: true,
    });

    // Load existing canvas data if provided
    if (props.canvasData && Object.keys(props.canvasData).length > 0) {
      hasInitiallyLoaded.value = true;
      isLoadingData = true;
      fabricCanvas.loadFromJSON(props.canvasData).then(() => {
        fabricCanvas?.renderAll();
        isLoadingData = false;
      });
    }

    // Listen for changes
    fabricCanvas.on("object:modified", saveCanvas);
    fabricCanvas.on("object:added", () => {
      saveCanvas();
      // Ensure grid stays visible after object addition
      nextTick(() => drawGrid());
    });
    fabricCanvas.on("object:removed", () => {
      saveCanvas();
      nextTick(() => drawGrid());
    });

    // Listen for selection changes
    fabricCanvas.on("selection:created", e => {
      emit("selectionChanged", e.selected?.[0] || null);
    });
    fabricCanvas.on("selection:updated", e => {
      emit("selectionChanged", e.selected?.[0] || null);
    });
    fabricCanvas.on("selection:cleared", () => {
      emit("selectionChanged", null);
    });

    // Snap to grid on object moving
    fabricCanvas.on("object:moving", e => {
      if (!props.snapToGrid || !e.target) return;

      const obj = e.target;
      const gridSize = effectiveGridSize.value;

      // Snap left and top to grid
      obj.set({
        left: Math.round((obj.left || 0) / gridSize) * gridSize,
        top: Math.round((obj.top || 0) / gridSize) * gridSize,
      });
    });

    // Handle textbox scaling - capture the final dimensions and reset scale
    fabricCanvas.on("object:scaling", e => {
      const obj = e.target as Textbox & { data?: TextBoxData };
      if (obj?.type !== "textbox") return;

      // Calculate the new dimensions based on scale
      const scaleX = obj.scaleX || 1;
      const scaleY = obj.scaleY || 1;
      const newWidth = Math.round((obj.width || 150) * scaleX);
      const newHeight = Math.round((obj.height || 50) * scaleY);

      // Update the textbox width (this controls word wrap)
      obj.set({
        width: newWidth,
        scaleX: 1,
        scaleY: 1,
      });

      // Store the fixed dimensions in data for autofit
      obj.data = {
        ...obj.data,
        fixedWidth: newWidth,
        fixedHeight: newHeight,
      };

      // If autofit is enabled, adjust font size
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

    // Ensure the textbox has the correct width for word wrapping
    textObj.set({ width: maxWidth });

    // Calculate if text overflows
    const textOverflows = () => {
      // Check if text height exceeds max height
      if ((textObj.height || 0) > maxHeight) return true;

      // For textbox, check if any line exceeds the width
      // This is indicated by text being cut off or wrapped excessively
      const textLines = textObj.textLines || [];
      const ctx = textObj.canvas?.getContext();
      if (ctx && textLines.length > 0) {
        ctx.font = `${textObj.fontSize}px ${textObj.fontFamily}`;
        for (const line of textLines) {
          const lineWidth = ctx.measureText(line).width;
          if (lineWidth > maxWidth * 0.98) {
            // Text line is close to or exceeding width
            return true;
          }
        }
      }
      return false;
    };

    // Shrink font if text exceeds bounds
    while (textOverflows() && (textObj.fontSize || 16) > minFontSize) {
      textObj.set("fontSize", (textObj.fontSize || 16) - 1);
    }

    // Grow font back if there's room (up to max)
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

    // Update grid canvas size
    gridCanvasRef.value.width = width;
    gridCanvasRef.value.height = height;

    // Draw white background (label area)
    ctx.fillStyle = "#ffffff";
    ctx.fillRect(0, 0, width, height);

    // Only draw grid lines if enabled
    if (!props.showGrid) return;

    // Draw grid lines
    ctx.strokeStyle = "#e0e0e0";
    ctx.lineWidth = 0.5;

    // Vertical lines
    for (let x = 0; x <= width; x += gridSize) {
      ctx.beginPath();
      ctx.moveTo(x, 0);
      ctx.lineTo(x, height);
      ctx.stroke();
    }

    // Horizontal lines
    for (let y = 0; y <= height; y += gridSize) {
      ctx.beginPath();
      ctx.moveTo(0, y);
      ctx.lineTo(width, y);
      ctx.stroke();
    }
  }

  function saveCanvas() {
    if (!fabricCanvas || isLoadingData) return;
    const json = fabricCanvas.toObject(["data"]);
    emit("update:canvasData", json);
  }

  function addText(text: string = "Text", options: Record<string, unknown> = {}) {
    if (!fabricCanvas) return;

    const defaultWidth = 150;
    const defaultHeight = 30;

    const textObj = new Textbox(text, {
      left: 20,
      top: 20,
      width: defaultWidth, // Width controls word wrap
      fontSize: 16,
      fontFamily: "Arial",
      fill: "#000000",
      textAlign: "left",
      splitByGrapheme: false, // Wrap by word, not character
      // Custom data for autofit settings
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

    // Create a placeholder rectangle for the barcode
    const rect = new Rect({
      left: 0,
      top: 0,
      width: 100,
      height: 100,
      fill: "#f0f0f0",
      stroke: "#cccccc",
      strokeWidth: 1,
    });

    // Add label text
    const label = new IText(`[${format.toUpperCase()}]`, {
      left: 5,
      top: 40,
      fontSize: 12,
      fontFamily: "Arial",
      fill: "#666666",
      selectable: false,
      evented: false,
    });

    // Group them with data in constructor for proper serialization
    const group = new Group([rect, label], {
      left: 20,
      top: 20,
      data: {
        type: "barcode",
        format,
        contentSource,
      },
    } as ConstructorParameters<typeof Group>[1] & { data: Record<string, unknown> });

    fabricCanvas.add(group);
    fabricCanvas.setActiveObject(group);
    fabricCanvas.renderAll();
    saveCanvas();
  }

  function addShape(shapeType: "rect" | "line") {
    if (!fabricCanvas) return;

    let shape: FabricObject;

    if (shapeType === "rect") {
      shape = new Rect({
        left: 20,
        top: 20,
        width: 100,
        height: 50,
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

  function deleteSelected() {
    if (!fabricCanvas) return;

    const activeObject = fabricCanvas.getActiveObject();
    if (activeObject) {
      fabricCanvas.remove(activeObject);
      fabricCanvas.renderAll();
      saveCanvas();
    }
  }

  function updateSelectedObject(updates: Record<string, unknown>) {
    if (!fabricCanvas) return;

    const activeObject = fabricCanvas.getActiveObject();
    if (activeObject) {
      // Handle custom 'data' property separately as Fabric.js set() may not handle it properly
      if ("data" in updates) {
        (activeObject as FabricObject & { data?: Record<string, unknown> }).data = updates.data as Record<
          string,
          unknown
        >;
        // Remove data from updates so we don't try to set() it again
        const { data: _, ...restUpdates } = updates;
        if (Object.keys(restUpdates).length > 0) {
          activeObject.set(restUpdates as Partial<FabricObject>);
        }
      } else {
        activeObject.set(updates as Partial<FabricObject>);
      }
      fabricCanvas.renderAll();
      saveCanvas();
      // Re-emit selection to trigger reactivity update in properties panel
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
    // Return in reverse order so top layer appears first in the list
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

    const objects = fabricCanvas.getObjects();
    const obj = objects[index];
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
    // Can bring forward if not already at front (highest index)
    if (obj && index < objects.length - 1) {
      fabricCanvas.bringObjectForward(obj);
      fabricCanvas.renderAll();
      saveCanvas();
    }
  }

  function sendBackward(index: number) {
    if (!fabricCanvas) return;

    const objects = fabricCanvas.getObjects();
    const obj = objects[index];
    // Can send backward if not already at back (index 0)
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
    // Can bring to front if not already at front
    if (obj && index < objects.length - 1) {
      fabricCanvas.bringObjectToFront(obj);
      fabricCanvas.renderAll();
      saveCanvas();
    }
  }

  function sendToBack(index: number) {
    if (!fabricCanvas) return;

    const objects = fabricCanvas.getObjects();
    const obj = objects[index];
    // Can send to back if not already at back
    if (obj && index > 0) {
      fabricCanvas.sendObjectToBack(obj);
      fabricCanvas.renderAll();
      saveCanvas();
    }
  }

  function toggleLayerVisibility(index: number) {
    if (!fabricCanvas) return;

    const objects = fabricCanvas.getObjects();
    const obj = objects[index];
    if (obj) {
      obj.visible = !obj.visible;
      fabricCanvas.renderAll();
      saveCanvas();
    }
  }

  // Expose methods for parent components
  defineExpose({
    addText,
    addBarcode,
    addShape,
    deleteSelected,
    updateSelectedObject,
    getCanvas,
    getLayers,
    selectLayerByIndex,
    bringForward,
    sendBackward,
    bringToFront,
    sendToBack,
    toggleLayerVisibility,
    // Zoom controls
    zoomIn,
    zoomOut,
    zoomToFit,
    zoomTo100,
    zoomLevel,
  });

  function handleKeyDown(e: KeyboardEvent) {
    if (!fabricCanvas) return;

    // Delete or Backspace to remove selected object
    if (e.key === "Delete" || e.key === "Backspace") {
      // Don't delete if we're editing text
      const activeObject = fabricCanvas.getActiveObject();
      if (activeObject && (activeObject as { isEditing?: boolean }).isEditing) {
        return;
      }

      e.preventDefault();
      deleteSelected();
    }
  }

  onMounted(() => {
    nextTick(() => {
      initCanvas();
      drawGrid();
    });

    // Add keyboard listener
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
    // Redraw grid with zoom
    drawGrid();
  });

  // Watch for grid settings changes
  watch(
    () => [props.showGrid, props.gridSize],
    () => {
      drawGrid();
    }
  );

  // Watch for canvas data changes (e.g., when template loads asynchronously)
  // Only trigger on initial load, not on every change
  watch(
    () => props.canvasData,
    newData => {
      // Only load once when data first arrives
      if (fabricCanvas && newData && Object.keys(newData).length > 0 && !hasInitiallyLoaded.value) {
        hasInitiallyLoaded.value = true;
        isLoadingData = true;
        fabricCanvas.loadFromJSON(newData).then(() => {
          fabricCanvas?.renderAll();
          isLoadingData = false;
        });
      }
    },
    { immediate: true }
  );
</script>

<template>
  <div ref="canvasContainerRef" class="canvas-container overflow-auto rounded border bg-gray-100 p-4">
    <div class="relative inline-block shadow-lg">
      <!-- Grid layer (behind fabric canvas) -->
      <canvas
        ref="gridCanvasRef"
        class="pointer-events-none absolute left-0 top-0"
        :style="{ zIndex: 0 }"
        :width="displayWidth * zoomLevel"
        :height="displayHeight * zoomLevel"
      />
      <!-- Fabric canvas (front, with transparent background area) -->
      <canvas ref="canvasRef" class="relative" :style="{ zIndex: 1 }" />
    </div>
  </div>
</template>

<style scoped>
  .canvas-container {
    min-height: 300px;
  }
</style>
