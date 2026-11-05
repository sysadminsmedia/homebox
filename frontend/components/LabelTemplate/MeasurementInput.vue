<script setup lang="ts">
  import { Input } from "@/components/ui/input";

  // Number input that edits a millimetre value in the user's measurement unit.
  // The typed text is kept locally so intermediate input like "2." isn't
  // rewritten by the round trip through millimetres.
  const props = defineProps<{
    id?: string;
    min?: number;
    max?: number;
    disabled?: boolean;
    inputClass?: string;
  }>();

  const modelValue = defineModel<number>({ required: true });

  const measurement = useMeasurement();
  const text = ref(String(measurement.toDisplay(modelValue.value)));

  function onInput(value: string | number) {
    text.value = String(value);
    const parsed = Number(value);
    if (value === "" || Number.isNaN(parsed)) return;
    modelValue.value = measurement.fromDisplay(parsed);
  }

  // Re-sync when the value changes from outside (preset picked, orientation
  // swapped) or the unit changes, but not when it's just our own edit echoing back.
  watch([modelValue, measurement.system], ([mm], [, previousSystem]) => {
    const parsed = Number(text.value);
    const echo =
      measurement.system.value === previousSystem &&
      !Number.isNaN(parsed) &&
      Math.abs(measurement.fromDisplay(parsed) - mm) < 0.01;
    if (!echo) {
      text.value = String(measurement.toDisplay(mm));
    }
  });
</script>

<template>
  <div class="flex items-center gap-1">
    <Input
      :id="props.id"
      :model-value="text"
      type="number"
      :step="measurement.step.value"
      :min="props.min !== undefined ? measurement.toDisplay(props.min) : undefined"
      :max="props.max !== undefined ? measurement.toDisplay(props.max) : undefined"
      :disabled="props.disabled"
      :class="props.inputClass"
      @update:model-value="onInput"
    />
    <span class="shrink-0 text-xs text-muted-foreground">{{ measurement.unit.value }}</span>
  </div>
</template>
