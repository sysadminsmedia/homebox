<script setup lang="ts">
  import { BrowserMultiFormatReader, NotFoundException } from "@zxing/library";
  import { useI18n } from "vue-i18n";

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

  const handleError = (error: unknown) => {
    console.error("Scanner error:", error);
    errorMessage.value = t("scanner.error");
  };

  onMounted(async () => {
    if (!(navigator && navigator.mediaDevices && "enumerateDevices" in navigator.mediaDevices)) {
      errorMessage.value = t("scanner.unsupported");
      return;
    }

    try {
      const devices = await codeReader.listVideoInputDevices();
      sources.value = devices;

      if (devices.length > 0) {
        selectedSource.value = devices[0].deviceId;
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

<template>
  <div class="flex flex-col gap-12 pb-16">
    <section>
      <div class="mx-auto">
        <div class="max-w-screen-md">
          <div v-if="errorMessage" role="alert" class="alert alert-error mb-5 shadow-lg">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="size-6 shrink-0 stroke-current"
              fill="none"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <span class="text-sm">{{ errorMessage }}</span>
          </div>
          <video ref="video" class="rounded-box shadow-lg" poster="data:image/gif,AAAA"></video>
          <select v-model="selectedSource" class="select mt-4 w-full shadow-lg">
            <option disabled selected :value="null">{{ t("scanner.select_video_source") }}</option>
            <option v-for="source in sources" :key="source.deviceId" :value="source.deviceId">
              {{ source.label }}
            </option>
          </select>
        </div>
      </div>
    </section>
  </div>
</template>

<style lang="css" scoped>
  video {
    width: 100%;
    object-fit: cover;
    margin-left: auto;
    margin-right: auto;
  }
</style>
