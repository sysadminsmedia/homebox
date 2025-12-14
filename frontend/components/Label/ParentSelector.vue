<template>
  <div class="flex flex-col gap-1">
    <Label :for="id">
      {{ $t('components.label.parent_selector.label') }}
    </Label>
    <Select v-model="modelValue">
      <SelectTrigger :id="id">
        <SelectValue :placeholder="$t('components.label.parent_selector.placeholder')" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="">{{ $t('components.label.parent_selector.no_parent') }}</SelectItem>
        <SelectItem v-for="label in props.labels" :key="label.id" :value="label.id">
          {{ label.name }}
        </SelectItem>
      </SelectContent>
    </Select>
  </div>
</template>

<script setup lang="ts">
  import { Label } from "@/components/ui/label";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import type { LabelOut } from "~/lib/api/types/data-contracts";

  const id = useId();

  const emit = defineEmits(["update:modelValue"]);
  const props = defineProps({
    modelValue: {
      type: String as () => string | null,
      default: null,
    },
    labels: {
      type: Array as () => LabelOut[],
      required: true,
    },
  });

  const modelValue = useVModel(props, "modelValue", emit);
</script>
