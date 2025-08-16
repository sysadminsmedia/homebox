<template>
  <svg id="barcode"></svg>
</template>

<script setup lang="ts">
  import JsBarcode from "jsbarcode";

  const props = defineProps<{
    barcode: string;
  }>();

  onMounted(() => {
    if (!props.barcode) return;

    JsBarcode("#barcode", props.barcode, {
      // format: "auto",
      width: 1,
      height: 25,
      flat: false,
      fontSize: 12,
    });
  });
</script>
