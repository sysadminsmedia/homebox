<script setup lang="ts">
  import { BrowserMultiFormatReader, NotFoundException, BarcodeFormat } from "@zxing/library";
  import { useI18n } from "vue-i18n";
  import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
  import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "@/components/ui/select";
  import { Button } from "@/components/ui/button";
  import MdiAlertCircleOutline from "~icons/mdi/alert-circle-outline";

  import { useDialog } from "~/components/ui/dialog-provider";
  const { openDialog } = useDialog();

  definePageMeta({
    middleware: ["auth"],
  });
  useHead({
    title: "Homebox | Scanner",
  });

  const { t } = useI18n();

  const sources = ref<MediaDeviceInfo[]>([]);
  const selectedSource = ref<string | null>(null);
  const loading = ref(false);
  const video = ref<HTMLVideoElement>();
  const codeReader = new BrowserMultiFormatReader();
  const errorMessage = ref<string | null>(null);
  const detectedBarcode = ref<string | null>(null);
  const detectedBarcodeType = ref<string | null>();
  const api = useUserApi();

  const handleError = (error: unknown) => {
    console.error("Scanner error:", error);
    errorMessage.value = t("scanner.error");
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
  }

  onMounted(async () => {
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
  });

  // stop the code reader when navigating away
  onBeforeUnmount(() => codeReader.reset());

  watch(selectedSource, async newSource => {
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
</script>

<template>
  <Card class="mx-auto w-full max-w-screen-md">
    <CardHeader>
      <CardTitle>{{ t("scanner.title") }}</CardTitle>
    </CardHeader>
    <CardContent>
      <div
        v-if="errorMessage"
        class="border-destructive bg-destructive/10 text-destructive mb-5 flex items-center gap-2 rounded-md border p-4"
        role="alert"
      >
        <MdiAlertCircleOutline class="text-destructive" />
        <span class="text-sm font-medium">{{ errorMessage }}</span>
      </div>
      <div
        v-if="detectedBarcode"
        class="border-destructive bg-destructive/10 text-destructive mb-5 flex items-center gap-2 rounded-md border p-4"
        role="alert"
      >
        <MdiAlertCircleOutline class="text-default" />
        <span class="text-sm font-medium">{{ detectedBarcodeType }} product barcode detected: {{ detectedBarcode }} </span>

        <ButtonGroup>
          <Button :disabled="loading" type="submit" @click="handleButtonClick">Fetchdata and create</Button>
        </ButtonGroup>
      </div>

      <!-- eslint-disable-next-line tailwindcss/no-custom-classname -->
      <video ref="video" class="bg-muted aspect-video w-full rounded-lg shadow" poster="data:image/gif,AAAA"></video>
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
    </CardContent>
  </Card>
</template>

<style scoped>
  video {
    object-fit: cover;
  }
</style>
