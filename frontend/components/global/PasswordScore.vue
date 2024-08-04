<template>
  <div class="py-4">
    <p class="text-sm">{{ $t("password_strength") }}: {{ message }}</p>
    <progress
      class="progress w-full progress-bar"
      :value="score"
      max="100"
      :class="{
        'progress-success': score > 50,
        'progress-warning': score > 25 && score < 50,
        'progress-error': score < 25,
      }"
    />
  </div>
</template>

<script setup lang="ts">
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
