<script setup lang="ts">
  import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
  import { Button } from "@/components/ui/button";
  import MdiFormatText from "~icons/mdi/format-text";
  import MdiDatabase from "~icons/mdi/database";
  import MdiBarcode from "~icons/mdi/barcode";
  import MdiRectangleOutline from "~icons/mdi/rectangle-outline";
  import MdiMinus from "~icons/mdi/minus";
  import MdiArrowUp from "~icons/mdi/arrow-up";
  import MdiArrowDown from "~icons/mdi/arrow-down";
  import MdiArrowCollapseUp from "~icons/mdi/arrow-collapse-up";
  import MdiArrowCollapseDown from "~icons/mdi/arrow-collapse-down";
  import MdiEye from "~icons/mdi/eye";
  import MdiEyeOff from "~icons/mdi/eye-off";

  interface LayerItem {
    index: number;
    type: string;
    name: string;
    visible: boolean;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    object: any;
  }

  const props = defineProps<{
    layers: LayerItem[];
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    selectedObject: any;
  }>();

  const emit = defineEmits<{
    selectLayer: [index: number];
    bringForward: [index: number];
    sendBackward: [index: number];
    bringToFront: [index: number];
    sendToBack: [index: number];
    toggleVisibility: [index: number];
  }>();

  function getLayerIcon(type: string, data?: Record<string, unknown>) {
    const lowerType = type?.toLowerCase() || "";

    // Check custom data type first
    if (data?.type === "data_field") {
      return MdiDatabase;
    }
    if (data?.type === "barcode") {
      return MdiBarcode;
    }

    // Check Fabric.js type
    if (lowerType.includes("text") || lowerType === "i-text" || lowerType === "itext") {
      return MdiFormatText;
    }
    if (lowerType === "rect") {
      return MdiRectangleOutline;
    }
    if (lowerType === "line") {
      return MdiMinus;
    }
    if (lowerType === "group") {
      return MdiBarcode; // Groups are typically barcodes
    }

    return MdiRectangleOutline;
  }

  function getLayerName(layer: LayerItem): string {
    const data = layer.object?.data;
    const type = layer.type?.toLowerCase() || "";

    // Data field
    if (data?.type === "data_field") {
      return data.displayName || data.field || "Data Field";
    }

    // Barcode
    if (data?.type === "barcode") {
      return `${data.format?.toUpperCase() || "Barcode"} (${data.contentSource || "custom"})`;
    }

    // Text - show first part of text content
    if (type.includes("text") || type === "i-text" || type === "itext") {
      const text = layer.object?.text || "";
      if (text.length > 20) {
        return text.substring(0, 20) + "...";
      }
      return text || "Text";
    }

    // Shape
    if (type === "rect") {
      return "Rectangle";
    }
    if (type === "line") {
      return "Line";
    }
    if (type === "group") {
      return "Group";
    }

    return type || "Object";
  }

  function isSelected(layer: LayerItem): boolean {
    return props.selectedObject === layer.object;
  }

  // Position in the displayed (reversed) layers array
  const selectedDisplayIndex = computed(() => {
    if (!props.selectedObject) return -1;
    return props.layers.findIndex(l => l.object === props.selectedObject);
  });

  // Original canvas object index for the selected layer
  const selectedCanvasIndex = computed(() => {
    if (selectedDisplayIndex.value < 0) return -1;
    return props.layers[selectedDisplayIndex.value]?.index ?? -1;
  });

  // In the display, top layer is first, so "bring forward" means moving up in z-order (higher canvas index)
  const canBringForward = computed(() => selectedDisplayIndex.value > 0);
  const canSendBackward = computed(
    () => selectedDisplayIndex.value >= 0 && selectedDisplayIndex.value < props.layers.length - 1
  );
</script>

<template>
  <Card class="w-64">
    <CardHeader class="pb-2">
      <div class="flex items-center justify-between">
        <CardTitle class="text-sm">{{ $t("components.label_template.editor.layers.title") }}</CardTitle>
        <div class="flex gap-0.5">
          <Button
            variant="ghost"
            size="icon"
            class="size-6"
            :disabled="!canBringForward"
            :title="$t('components.label_template.editor.layers.bring_to_front')"
            @click="emit('bringToFront', selectedCanvasIndex)"
          >
            <MdiArrowCollapseUp class="size-3.5" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            class="size-6"
            :disabled="!canBringForward"
            :title="$t('components.label_template.editor.layers.bring_forward')"
            @click="emit('bringForward', selectedCanvasIndex)"
          >
            <MdiArrowUp class="size-3.5" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            class="size-6"
            :disabled="!canSendBackward"
            :title="$t('components.label_template.editor.layers.send_backward')"
            @click="emit('sendBackward', selectedCanvasIndex)"
          >
            <MdiArrowDown class="size-3.5" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            class="size-6"
            :disabled="!canSendBackward"
            :title="$t('components.label_template.editor.layers.send_to_back')"
            @click="emit('sendToBack', selectedCanvasIndex)"
          >
            <MdiArrowCollapseDown class="size-3.5" />
          </Button>
        </div>
      </div>
    </CardHeader>
    <CardContent class="p-0">
      <div class="max-h-48 overflow-y-auto">
        <div v-if="layers.length === 0" class="p-4 text-center text-sm text-muted-foreground">
          {{ $t("components.label_template.editor.layers.empty") }}
        </div>
        <div v-else class="divide-y">
          <!-- Layers are displayed in reverse order (top layer first) -->
          <div
            v-for="layer in layers"
            :key="layer.index"
            class="flex cursor-pointer items-center gap-2 px-3 py-2 hover:bg-muted/50"
            :class="{ 'bg-primary/10': isSelected(layer) }"
            @click="emit('selectLayer', layer.index)"
          >
            <Button
              variant="ghost"
              size="icon"
              class="size-5 shrink-0"
              :title="
                layer.visible
                  ? $t('components.label_template.editor.layers.hide')
                  : $t('components.label_template.editor.layers.show')
              "
              @click.stop="emit('toggleVisibility', layer.index)"
            >
              <component
                :is="layer.visible ? MdiEye : MdiEyeOff"
                class="size-3.5"
                :class="{ 'text-muted-foreground': !layer.visible }"
              />
            </Button>
            <component
              :is="getLayerIcon(layer.type, layer.object?.data)"
              class="size-4 shrink-0 text-muted-foreground"
            />
            <span class="truncate text-sm" :class="{ 'text-muted-foreground': !layer.visible }">
              {{ getLayerName(layer) }}
            </span>
          </div>
        </div>
      </div>
    </CardContent>
  </Card>
</template>
