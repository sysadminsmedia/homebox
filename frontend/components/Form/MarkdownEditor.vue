<script setup lang="ts">
  import { ref, watch } from "vue";
  import Markdown from "@/components/global/Markdown.vue";
  import { Checkbox } from "@/components/ui/checkbox";
  import { Textarea } from "@/components/ui/textarea";

  const props = withDefaults(
    defineProps<{
      modelValue?: string | null;
      label?: string | null;
    }>(),
    {
      modelValue: null,
      label: null,
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
</script>

<template>
  <div class="w-full">
    <div class="mb-2 flex items-center justify-between">
      <label v-if="props.label" :for="id" class="text-sm font-medium">{{ props.label }}</label>
      <div class="flex items-center gap-2">
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
