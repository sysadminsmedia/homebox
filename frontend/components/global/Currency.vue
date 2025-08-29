<template>
  <Suspense>
    <template #default>
      {{ formattedValue }}
    </template>
    <template #fallback> {{ $t("global.loading") }} </template>
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

  type AsyncReturnType<T extends (...args: unknown[]) => unknown> = Awaited<ReturnType<T>>;

  const fmt = ref<AsyncReturnType<typeof useFormatCurrency> | null>(null);

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
