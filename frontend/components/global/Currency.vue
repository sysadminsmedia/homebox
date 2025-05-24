<template>
  <Suspense>
    <template #default>
      {{ formattedValue }}
    </template>
    <template #fallback> Loading... </template>
  </Suspense>
</template>

<script setup lang="ts">
  import { ref, computed } from "vue";
  import { useI18n } from "vue-i18n";

  const { t } = useI18n();

  type Props = {
    amount: string | number;
  };

  const props = defineProps<Props>();

  const fmt = ref(null);

  const loadFormatter = async () => {
    fmt.value = await useFormatCurrency();
  };

  loadFormatter();

  const formattedValue = computed(() => {
    if (!fmt.value) {
      return t("global.loading");
    }

    if (!props.amount || props.amount === "0") {
      return fmt.value(0);
    }

    return fmt.value(props.amount);
  });
</script>
