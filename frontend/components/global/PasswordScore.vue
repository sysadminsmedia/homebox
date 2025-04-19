<template>
  <div class="py-4">
    <p class="text-sm">{{ $t("components.global.password_score.password_strength") }}: {{ message }}</p>
    <Progress class="w-full" :model-value="score" />
  </div>
</template>

<script setup lang="ts">
  import { Progress } from "@/components/ui/progress";

  const props = defineProps({
    password: {
      type: String,
      required: true,
    },
    valid: {
      type: Boolean,
      required: false,
    },
  });

  const emits = defineEmits(["update:valid"]);

  const { password } = toRefs(props);

  const { score, message, isValid } = usePasswordScore(password);

  watchEffect(() => {
    emits("update:valid", isValid.value);
  });
</script>

<style scoped></style>
