<script setup lang="ts">
  import { Input } from "@/components/ui/input";
  import { Label } from "@/components/ui/label";

  export type EditorUnit = "mm" | "in";

  const props = defineProps<{
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    selectedObject: any;
    unit: EditorUnit;
  }>();

  const emit = defineEmits<{
    updateObject: [updates: Record<string, unknown>];
  }>();

  // The canvas works in CSS pixels at 96 dpi; the panel shows real-world units.
  const PX_PER_MM = 96 / 25.4;
  const PX_PER_IN = 96;

  const pxPerUnit = computed(() => (props.unit === "mm" ? PX_PER_MM : PX_PER_IN));
  const decimals = computed(() => (props.unit === "mm" ? 1 : 2));
  const step = computed(() => (props.unit === "mm" ? 0.5 : 0.05));

  function toUnit(px: number): string {
    return (px / pxPerUnit.value).toFixed(decimals.value);
  }

  function toPx(value: string): number | null {
    const n = parseFloat(value);
    if (Number.isNaN(n)) return null;
    return n * pxPerUnit.value;
  }

  const type = computed(() => String(props.selectedObject?.type || "").toLowerCase());
  const canResize = computed(() => type.value === "rect" || type.value === "image" || type.value === "group");

  const objectProps = computed(() => {
    if (!props.selectedObject) return null;
    const obj = props.selectedObject;
    return {
      left: toUnit(obj.left || 0),
      top: toUnit(obj.top || 0),
      width: toUnit((obj.width || 0) * (obj.scaleX || 1)),
      height: toUnit((obj.height || 0) * (obj.scaleY || 1)),
      angle: Math.round(obj.angle || 0),
    };
  });

  function handlePositionChange(prop: "left" | "top", value: string) {
    const px = toPx(value);
    if (px !== null) emit("updateObject", { [prop]: Math.round(px * 100) / 100 });
  }

  // Width/height edits change the scale so Fabric keeps the object's native size;
  // the renderer multiplies width by scaleX the same way.
  function handleSizeChange(prop: "width" | "height", value: string) {
    const px = toPx(value);
    if (px === null || px <= 0 || !props.selectedObject) return;
    const base = prop === "width" ? props.selectedObject.width || 1 : props.selectedObject.height || 1;
    emit("updateObject", { [prop === "width" ? "scaleX" : "scaleY"]: px / base });
  }

  function handleAngleChange(value: string) {
    let numValue = parseInt(value, 10);
    if (Number.isNaN(numValue)) return;
    numValue = ((numValue % 360) + 360) % 360;
    emit("updateObject", { angle: numValue });
  }
</script>

<template>
  <div v-if="objectProps" class="space-y-4">
    <div class="grid grid-cols-2 gap-2">
      <div class="space-y-1">
        <Label for="tp-left" class="text-xs text-muted-foreground">X</Label>
        <div class="flex items-center gap-1">
          <Input
            id="tp-left"
            type="number"
            :step="step"
            :model-value="objectProps.left"
            class="h-8"
            @change="handlePositionChange('left', ($event.target as HTMLInputElement).value)"
          />
          <span class="w-6 shrink-0 text-xs text-muted-foreground">{{ unit }}</span>
        </div>
      </div>
      <div class="space-y-1">
        <Label for="tp-top" class="text-xs text-muted-foreground">Y</Label>
        <div class="flex items-center gap-1">
          <Input
            id="tp-top"
            type="number"
            :step="step"
            :model-value="objectProps.top"
            class="h-8"
            @change="handlePositionChange('top', ($event.target as HTMLInputElement).value)"
          />
          <span class="w-6 shrink-0 text-xs text-muted-foreground">{{ unit }}</span>
        </div>
      </div>
      <div class="space-y-1">
        <Label for="tp-width" class="text-xs text-muted-foreground">
          {{ $t("components.label_template.form.width") }}
        </Label>
        <div class="flex items-center gap-1">
          <Input
            id="tp-width"
            type="number"
            :step="step"
            :model-value="objectProps.width"
            class="h-8"
            :disabled="!canResize"
            @change="handleSizeChange('width', ($event.target as HTMLInputElement).value)"
          />
          <span class="w-6 shrink-0 text-xs text-muted-foreground">{{ unit }}</span>
        </div>
      </div>
      <div class="space-y-1">
        <Label for="tp-height" class="text-xs text-muted-foreground">
          {{ $t("components.label_template.form.height") }}
        </Label>
        <div class="flex items-center gap-1">
          <Input
            id="tp-height"
            type="number"
            :step="step"
            :model-value="objectProps.height"
            class="h-8"
            :disabled="!canResize"
            @change="handleSizeChange('height', ($event.target as HTMLInputElement).value)"
          />
          <span class="w-6 shrink-0 text-xs text-muted-foreground">{{ unit }}</span>
        </div>
      </div>
    </div>

    <div class="space-y-1">
      <Label for="tp-angle" class="text-xs text-muted-foreground">
        {{ $t("components.label_template.editor.properties.rotation") }}
      </Label>
      <div class="flex items-center gap-1">
        <Input
          id="tp-angle"
          type="number"
          :model-value="objectProps.angle"
          class="h-8 w-28"
          min="0"
          max="360"
          step="1"
          @change="handleAngleChange(($event.target as HTMLInputElement).value)"
        />
        <span class="text-xs text-muted-foreground">°</span>
      </div>
    </div>
  </div>
</template>
