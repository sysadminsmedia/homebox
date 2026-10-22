<script setup lang="ts">
  import { Button } from "@/components/ui/button";
  import MdiFormatText from "~icons/mdi/format-text";
  import MdiDatabase from "~icons/mdi/database";
  import MdiBarcode from "~icons/mdi/barcode";
  import MdiRectangleOutline from "~icons/mdi/rectangle-outline";
  import MdiMinus from "~icons/mdi/minus";
  import MdiImage from "~icons/mdi/image";
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
    if (data?.type === "data_field") return MdiDatabase;
    if (data?.type === "barcode") return MdiBarcode;
    if (lowerType.includes("text") || lowerType === "i-text" || lowerType === "itext") return MdiFormatText;
    if (lowerType === "rect") return MdiRectangleOutline;
    if (lowerType === "line") return MdiMinus;
    if (lowerType === "image") return MdiImage;
    if (lowerType === "group") return MdiBarcode;
    return MdiRectangleOutline;
  }

  function getLayerName(layer: LayerItem): string {
    const data = layer.object?.data;
    const type = layer.type?.toLowerCase() || "";

    if (data?.type === "data_field") return data.displayName || data.field || "Data Field";
    if (data?.type === "barcode")
      return `${data.format?.toUpperCase() || "Barcode"} (${data.contentSource || "custom"})`;
    if (type.includes("text") || type === "i-text" || type === "itext") {
      const text = layer.object?.text || "";
      return text.length > 24 ? text.substring(0, 24) + "…" : text || "Text";
    }
    if (type === "rect") return "Rectangle";
    if (type === "line") return "Line";
    if (type === "image") return "Image";
    if (type === "group") return "Group";
    return type || "Object";
  }

  function isSelected(layer: LayerItem): boolean {
    return props.selectedObject === layer.object;
  }

  const selectedDisplayIndex = computed(() => {
    if (!props.selectedObject) return -1;
    return props.layers.findIndex(l => l.object === props.selectedObject);
  });

  const selectedCanvasIndex = computed(() => {
    if (selectedDisplayIndex.value < 0) return -1;
    return props.layers[selectedDisplayIndex.value]?.index ?? -1;
  });

  const canBringForward = computed(() => selectedDisplayIndex.value > 0);
  const canSendBackward = computed(
    () => selectedDisplayIndex.value >= 0 && selectedDisplayIndex.value < props.layers.length - 1
  );
</script>

<template>
  <div class="flex flex-col gap-2">
    <div class="flex items-center justify-between px-1">
      <span class="text-xs text-muted-foreground">
        {{ $t("components.label_template.editor.layers.order_hint") }}
      </span>
      <div class="flex gap-0.5">
        <Button
          variant="ghost"
          size="icon"
          class="size-7"
          :disabled="!canBringForward"
          :title="$t('components.label_template.editor.layers.bring_to_front')"
          @click="emit('bringToFront', selectedCanvasIndex)"
        >
          <MdiArrowCollapseUp class="size-4" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          class="size-7"
          :disabled="!canBringForward"
          :title="$t('components.label_template.editor.layers.bring_forward')"
          @click="emit('bringForward', selectedCanvasIndex)"
        >
          <MdiArrowUp class="size-4" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          class="size-7"
          :disabled="!canSendBackward"
          :title="$t('components.label_template.editor.layers.send_backward')"
          @click="emit('sendBackward', selectedCanvasIndex)"
        >
          <MdiArrowDown class="size-4" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          class="size-7"
          :disabled="!canSendBackward"
          :title="$t('components.label_template.editor.layers.send_to_back')"
          @click="emit('sendToBack', selectedCanvasIndex)"
        >
          <MdiArrowCollapseDown class="size-4" />
        </Button>
      </div>
    </div>

    <div
      v-if="layers.length === 0"
      class="rounded-md border border-dashed p-4 text-center text-sm text-muted-foreground"
    >
      {{ $t("components.label_template.editor.layers.empty") }}
    </div>
    <ul v-else class="divide-y rounded-md border">
      <li
        v-for="layer in layers"
        :key="layer.index"
        class="flex cursor-pointer items-center gap-2 px-2 py-1.5 hover:bg-muted/50"
        :class="{ 'bg-primary/10': isSelected(layer) }"
        @click="emit('selectLayer', layer.index)"
      >
        <Button
          variant="ghost"
          size="icon"
          class="size-6 shrink-0"
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
        <component :is="getLayerIcon(layer.type, layer.object?.data)" class="size-4 shrink-0 text-muted-foreground" />
        <span class="truncate text-sm" :class="{ 'text-muted-foreground': !layer.visible }">
          {{ getLayerName(layer) }}
        </span>
      </li>
    </ul>
  </div>
</template>
