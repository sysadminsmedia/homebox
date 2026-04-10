<template>
  <div ref="container" class="relative h-dvh w-dvw overflow-hidden bg-black">
    <!-- Camera feed -->
    <video ref="videoEl" class="size-full object-cover" autoplay muted playsinline poster="data:image/gif,AAAA" />

    <!-- Error state -->
    <div v-if="error" class="absolute inset-0 flex items-center justify-center bg-background/90 p-6">
      <div class="flex max-w-sm flex-col items-center gap-4 text-center">
        <MdiAlertCircleOutline class="size-12 text-destructive" />
        <p class="text-sm text-foreground">{{ t(error) }}</p>
        <Button variant="outline" @click="startCamera">
          {{ t("scanner_ar.retry") }}
        </Button>
      </div>
    </div>

    <!-- Overlay cards -->
    <ScannerAROverlayCard
      v-for="detection in detections"
      :key="`${detection.entityType}:${detection.id}`"
      :position="detection.boundingBox"
      :pose="detection.pose"
      :corner-points="detection.cornerPoints"
      :entity="detection.data"
      :entity-type="detection.entityType"
      :loading="detection.loading"
      :error="detection.error"
    />

    <!-- Controls -->
    <ScannerARControls :is-scanning="isScanning" @back="handleBack" @switch-camera="switchCamera" />
  </div>
</template>

<script setup lang="ts">
  import { ref, onMounted, onBeforeUnmount, watch } from "vue";
  import { useI18n } from "vue-i18n";
  import { useBarcodeDetector } from "@/composables/use-barcode-detector";
  import { Button } from "@/components/ui/button";
  import ScannerAROverlayCard from "@/components/Scanner/AROverlayCard.vue";
  import ScannerARControls from "@/components/Scanner/ARControls.vue";
  import MdiAlertCircleOutline from "~icons/mdi/alert-circle-outline";

  const { t } = useI18n();
  const videoEl = ref<HTMLVideoElement>();
  const container = ref<HTMLDivElement>();

  const { isScanning, error, detections, startCamera, stopCamera, switchCamera } = useBarcodeDetector(videoEl);

  let lastDetectionCount = 0;
  watch(detections, val => {
    if (val.length !== lastDetectionCount) {
      console.debug(
        "[AR:View] detections count changed:",
        lastDetectionCount,
        "->",
        val.length,
        val.map(d => `${d.entityType}:${d.id} loading=${d.loading} error=${d.error}`)
      );
      lastDetectionCount = val.length;
    }
  });

  watch(error, val => {
    if (val) console.error("[AR:View] error state:", val);
  });

  function handleBack() {
    const router = useRouter();
    router.back();
  }

  onMounted(() => {
    console.debug("[AR:View] mounted, videoEl:", !!videoEl.value);
    startCamera();
  });

  onBeforeUnmount(() => {
    stopCamera();
  });
</script>
