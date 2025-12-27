<script setup lang="ts">
  import { Input } from "@/components/ui/input";
  import { Label } from "@/components/ui/label";
  import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

  const props = defineProps<{
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    selectedObject: any;
  }>();

  const emit = defineEmits<{
    updateObject: [updates: Record<string, unknown>];
  }>();

  // Reactive properties
  const objectProps = computed(() => {
    if (!props.selectedObject) return null;

    const obj = props.selectedObject;

    return {
      left: Math.round(obj.left || 0),
      top: Math.round(obj.top || 0),
      width: Math.round((obj.width || 0) * (obj.scaleX || 1)),
      height: Math.round((obj.height || 0) * (obj.scaleY || 1)),
      angle: Math.round(obj.angle || 0),
      type: obj.type,
    };
  });

  function handlePositionChange(prop: "left" | "top", value: string) {
    const numValue = parseInt(value, 10);
    if (!isNaN(numValue)) {
      emit("updateObject", { [prop]: numValue });
    }
  }

  function handleAngleChange(value: string) {
    let numValue = parseInt(value, 10);
    if (isNaN(numValue)) return;

    // Normalize angle to 0-360 range
    numValue = ((numValue % 360) + 360) % 360;
    emit("updateObject", { angle: numValue });
  }
</script>

<template>
  <Card>
    <CardHeader class="pb-2">
      <CardTitle class="text-sm">{{ $t("components.label_template.editor.transform.title") }}</CardTitle>
    </CardHeader>
    <CardContent>
      <div v-if="objectProps" class="space-y-4">
        <div class="space-y-2">
          <Label class="text-xs text-muted-foreground">
            {{ $t("components.label_template.editor.properties.position") }}
          </Label>
          <div class="grid grid-cols-2 gap-2">
            <div>
              <Label for="left" class="text-xs">X</Label>
              <Input
                id="left"
                type="number"
                :model-value="objectProps.left"
                class="h-8"
                @update:model-value="handlePositionChange('left', String($event))"
              />
            </div>
            <div>
              <Label for="top" class="text-xs">Y</Label>
              <Input
                id="top"
                type="number"
                :model-value="objectProps.top"
                class="h-8"
                @update:model-value="handlePositionChange('top', String($event))"
              />
            </div>
          </div>
        </div>

        <div class="space-y-2">
          <Label class="text-xs text-muted-foreground">
            {{ $t("components.label_template.editor.properties.size") }}
          </Label>
          <div class="grid grid-cols-2 gap-2">
            <div>
              <Label class="text-xs">W</Label>
              <Input type="number" :model-value="objectProps.width" class="h-8" disabled />
            </div>
            <div>
              <Label class="text-xs">H</Label>
              <Input type="number" :model-value="objectProps.height" class="h-8" disabled />
            </div>
          </div>
        </div>

        <div class="space-y-2">
          <Label for="angle" class="text-xs text-muted-foreground">
            {{ $t("components.label_template.editor.properties.rotation") }}
          </Label>
          <div class="flex items-center gap-2">
            <Input
              id="angle"
              type="number"
              :model-value="objectProps.angle"
              class="h-8"
              min="0"
              max="360"
              step="1"
              @update:model-value="handleAngleChange(String($event))"
            />
            <span class="text-xs text-muted-foreground">°</span>
          </div>
        </div>
      </div>

      <div v-else class="py-4 text-center text-sm text-muted-foreground">
        {{ $t("components.label_template.editor.transform.no_selection") }}
      </div>
    </CardContent>
  </Card>
</template>
