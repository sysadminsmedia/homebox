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
        <div
          v-if="detectedBarcode"
          class="border-accent-foreground bg-accent text-accent-foreground mb-5 flex flex-col items-center gap-2 rounded-md border p-4"
          role="alert"
        >
          <div class="flex">
            <MdiBarcode class="text-default mr-2" />
            <span class="flex-1 text-center text-sm font-medium">
              {{ detectedBarcodeType }} {{ $t("scanner.barcode_detected_message") }}: <strong>{{ detectedBarcode }}</strong>
            </span>
          </div>

          <ButtonGroup>
            <Button :disabled="loading" type="submit" @click="handleButtonClick">
              {{ $t("scanner.barcode_fetch_data") }}
            </Button>
          </ButtonGroup>
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
  import { BrowserMultiFormatReader, NotFoundException, BarcodeFormat } from "@zxing/library";
  import { useI18n } from "vue-i18n";
  import { Dialog, DialogHeader, DialogTitle, DialogScrollContent } from "@/components/ui/dialog";
  import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "@/components/ui/select";
  import { Button } from "@/components/ui/button";
  import MdiBarcode from "~icons/mdi/barcode";
  import MdiAlertCircleOutline from "~icons/mdi/alert-circle-outline";
  import { useDialog } from "@/components/ui/dialog-provider";
  

  const { t } = useI18n();
  const { activeDialog, openDialog, closeDialog } = useDialog();
  const open = computed(() => activeDialog.value && activeDialog.value.id === "scanner");

  const sources = ref<MediaDeviceInfo[]>([]);
  const selectedSource = ref<string | null>(null);
  const loading = ref(false);
  const video = ref<HTMLVideoElement>();
  const codeReader = new BrowserMultiFormatReader();
  const errorMessage = ref<string | null>(null);
  const detectedBarcode = ref<string>("");
  const detectedBarcodeType = ref<string>("");
  const api = useUserApi();

  const handleError = (error: unknown) => {
    console.error("Scanner error:", error);
    errorMessage.value = t("scanner.error");
  };

  const checkPermissionsError = async () => {
    if (navigator.permissions) {
      const permissionStatus = await navigator.permissions.query({ name: "camera" as PermissionName });
      if (permissionStatus.state === "denied") {
        errorMessage.value = t("scanner.permission_denied");
        console.error("Camera permission denied");
        return true;
      }
    }
  };

  const handleButtonClick = () => {
    console.log("Button clicked!");

    getQRCodeUrl();
    // console.log("Value::: ", productEAN);

    /* const route2 = useRoute();

    const currentURL = window.location.href;
    // Adjust route import as needed
    console.log(route2(`/getproductfromean`)); */
  };

/*
  function openCreateModal(ItemCreate ic) {
      this.$emit('open-modal', ic)
  }
  */

  async function getQRCodeUrl() {
    /* const { isCanceled } = await confirm.open(
      "Are you sure you want to ensure all assets have an ID? This can take a while and cannot be undone."
    );

    if (isCanceled) {
      return;
    } */

    const result = await api.actions.getEAN(detectedBarcode.value);

    // this.$store.commit('setScannedData', result);

    if(result.error)
      return
    
    openDialog("create-item", result.data);

    /* if (result.error) {
      toast.error("Failed to ensure asset IDs.");
    } */

    // toast.success(`${result.data.completed} assets have been updated.`);
  };

  const startScanner = async () => {
    errorMessage.value = null;
    if (!(navigator && navigator.mediaDevices && "enumerateDevices" in navigator.mediaDevices)) {
      errorMessage.value = t("scanner.unsupported");
      return;
    }

    if (await checkPermissionsError()) {
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
            closeDialog("scanner");
            navigateTo(sanitizedPath);
          } catch (err) {
            // Check if it's a barcode for a new element
            const bcfmt = result.getBarcodeFormat();

            switch (bcfmt) {
              case BarcodeFormat.EAN_13:
              case BarcodeFormat.UPC_A:
              case BarcodeFormat.UPC_E:
              case BarcodeFormat.UPC_EAN_EXTENSION:
                console.info("Barcode detected");
                detectedBarcode.value = result.getText();
                detectedBarcodeType.value = BarcodeFormat[bcfmt].replaceAll("_","-");
                break;
              
              default:
                handleError(err);
            }

            loading.value = false;
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

  onUnmounted(() => {
    stopScanner();
  });
</script>

