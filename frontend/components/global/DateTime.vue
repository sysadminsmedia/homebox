<template>
  {{ value }}
</template>

<script setup lang="ts">
  import type { DateTimeFormat, DateTimeType } from "~~/composables/use-formatters";

  type Props = {
    date?: Date | string;
    format?: DateTimeFormat;
    datetimeType?: DateTimeType;
  };

  const props = withDefaults(defineProps<Props>(), {
    date: undefined,
    format: "relative",
    datetimeType: "date",
  });

  const value = computed(() => {
    if (!props.date || !validDate(props.date)) {
      return "";
    }

    return fmtDate(props.date, props.format, props.datetimeType);
  });
</script>
