<template>
  <BaseModal v-model="modal">
    <template #title>🎉 {{ $t("components.app.outdated.new_version_available") }} 🎉</template>
    <div class="p-4">
      <p>{{ $t("components.app.outdated.current_version") }}: {{ current }}</p>
      <p>{{ $t("components.app.outdated.latest_version") }}: {{ latest }}</p>
      <p>
        <a href="https://github.com/sysadminsmedia/homebox/releases" target="_blank" rel="noopener" class="link">
          {{ $t("components.app.outdated.new_version_available_link") }}
        </a>
      </p>
    </div>
    <button class="btn btn-warning" @click="hide">
      {{ $t("components.app.outdated.dismiss") }}
    </button>
  </BaseModal>
</template>

<script setup lang="ts">
  const props = defineProps({
    modelValue: {
      type: Boolean,
      required: true,
    },
    current: {
      type: String,
      required: true,
    },
    latest: {
      type: String,
      required: true,
    },
  });

  const modal = useVModel(props, "modelValue");

  const hide = () => {
    modal.value = false;
    localStorage.setItem("latestVersion", props.latest);
  };
</script>
