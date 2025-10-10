<script setup lang="ts">
  import { ref, watch } from "vue";
  import Markdown from "@/components/global/Markdown.vue";
  import { Checkbox } from "@/components/ui/checkbox";
  import { Textarea } from "@/components/ui/textarea";
  import { Label } from "@/components/ui/label";

  const props = withDefaults(
    defineProps<{
      modelValue?: string | null;
      label?: string | null;
      maxLength?: number;
      minLength?: number;
    }>(),
    {
      modelValue: null,
      label: null,
      maxLength: -1,
      minLength: -1,
    }
  );

  const emit = defineEmits(["update:modelValue"]);

  const local = ref(props.modelValue ?? "");

  watch(
    () => props.modelValue,
    v => {
      if (v !== local.value) local.value = v ?? "";
    }
  );

  watch(local, v => emit("update:modelValue", v === "" ? null : v));

  const showPreview = ref(false);

  const id = useId();

  const isLengthInvalid = computed(() => {
    if (typeof local.value !== "string") return false;
    const len = local.value.length;
    const max = props.maxLength ?? -1;
    const min = props.minLength ?? -1;
    return (max !== -1 && len > max) || (min !== -1 && len < min);
  });

  const lengthIndicator = computed(() => {
    if (typeof local.value !== "string") return "";
    const max = props.maxLength ?? -1;
    if (max !== -1) {
      return `${local.value.length}/${max}`;
    }
    return "";
  });
</script>

<template>
  <div class="w-full">
    <div class="mb-2 grid grid-cols-1 items-center gap-2 md:grid-cols-4">
      <div class="min-w-0">
        <Label :for="id" class="flex min-w-0 items-center gap-2 px-1">
          <span class="truncate" :title="props.label ?? ''">{{ props.label }}</span>
          <span class="grow" />
          <span class="ml-2 text-sm" :class="{ 'text-destructive': isLengthInvalid }">{{ lengthIndicator }}</span>
        </Label>
      </div>

      <div class="col-span-1 flex items-center justify-start gap-2 md:col-span-3 md:justify-end">
        <label class="text-xs text-slate-500">{{ $t("global.preview") }}</label>
        <Checkbox v-model="showPreview" />
      </div>
    </div>

    <div class="flex w-full flex-col gap-4">
      <Textarea :id="id" v-model="local" autosize class="resize-none" />

      <div v-if="showPreview">
        <Markdown :source="local" />
      </div>
    </div>
  </div>
</template>
