<template>
  <Dialog dialog-id="scanner">
    <DialogScrollContent>
      <DialogHeader>
        <DialogTitle>{{ t("scanner.title") }}</DialogTitle>
      </DialogHeader>
      <div>
        <div
          v-if="errorMessage"
          class="mb-5 flex items-center gap-2 rounded-md border border-destructive bg-destructive/10 p-4 text-destructive"
          role="alert"
        >
          <MdiAlertCircleOutline class="text-destructive" />
          <span class="text-sm font-medium">{{ errorMessage }}</span>
        </div>
        <!-- eslint-disable-next-line tailwindcss/no-custom-classname -->
        <video ref="video" class="aspect-video w-full rounded-lg bg-muted shadow" poster="data:image/gif,AAAA"></video>
        <div class="mt-4">
          <Select v-model="selectedSource">
            <SelectTrigger class="w-full">
              <SelectValue :placeholder="t('scanner.select_video_source')" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="source in sources" :key="source.deviceId" :value="source.deviceId">
                {{ source.label }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>
    </DialogScrollContent>
  </Dialog>
</template>

<script setup lang="ts">
  import { ref, watch, computed } from "vue";
  import { BrowserMultiFormatReader, NotFoundException } from "@zxing/library";
  import { useI18n } from "vue-i18n";
  import { Dialog, DialogHeader, DialogTitle, DialogScrollContent } from "@/components/ui/dialog";
  import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "@/components/ui/select";
  import MdiAlertCircleOutline from "~icons/mdi/alert-circle-outline";
  import { useDialog } from "@/components/ui/dialog-provider";

  const { t } = useI18n();
  const { activeDialog } = useDialog();
  const open = computed(() => activeDialog.value === "scanner");

  const sources = ref<MediaDeviceInfo[]>([]);
  const selectedSource = ref<string | null>(null);
  const loading = ref(false);
  const video = ref<HTMLVideoElement>();
  const codeReader = new BrowserMultiFormatReader();
  const errorMessage = ref<string | null>(null);

  const handleError = (error: unknown) => {
    console.error("Scanner error:", error);
    errorMessage.value = t("scanner.error");
  };

  const startScanner = async () => {
    errorMessage.value = null;
    if (!(navigator && navigator.mediaDevices && "enumerateDevices" in navigator.mediaDevices)) {
      errorMessage.value = t("scanner.unsupported");
      return;
    }

    try {
      const devices = await codeReader.listVideoInputDevices();
      sources.value = devices;

      if (devices.length > 0) {
        for (let i = 0; i < devices.length; i++) {
          if (devices[i].label.toLowerCase().includes("back")) {
            selectedSource.value = devices[i].deviceId;
          }
        }
        if (!selectedSource.value) {
          selectedSource.value = devices[0].deviceId;
        }
      } else {
        errorMessage.value = t("scanner.no_sources");
      }
    } catch (err) {
      handleError(err);
    }
  };

  const stopScanner = () => {
    codeReader.reset();
    sources.value = [];
    selectedSource.value = null;
    loading.value = false;
  };

  watch(open, async isOpen => {
    if (isOpen) {
      await startScanner();
    } else {
      stopScanner();
    }
  });

  watch(selectedSource, async newSource => {
    if (!open.value || !newSource) return;
    codeReader.reset();

    try {
      await codeReader.decodeFromVideoDevice(newSource, video.value!, (result, err) => {
        if (result && !loading.value) {
          loading.value = true;
          try {
            const url = new URL(result.getText());
            if (!url.pathname.startsWith("/")) {
              throw new Error(t("scanner.invalid_url"));
            }
            const sanitizedPath = url.pathname.replace(/[^a-zA-Z0-9-_/]/g, "");
            navigateTo(sanitizedPath);
          } catch (err) {
            loading.value = false;
            handleError(err);
          }
        }
        if (err && !(err instanceof NotFoundException)) {
          console.error(err);
          handleError(err);
        }
      });
    } catch (err) {
      handleError(err);
    }
  });
</script>

<style scoped>
  video {
    object-fit: cover;
  }
</style>
