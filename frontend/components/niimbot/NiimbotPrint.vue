<script setup lang="ts">
  import { ref, onMounted } from "vue";
  import { useI18n } from "vue-i18n";
  import { type QueryValue, route } from "../../lib/api/base/urls";
  import { toast } from "@/components/ui/sonner";
  import { Button } from "@/components/ui/button";
  import { PRESET_TAPE_SIZES, type TapeSize } from "@/composables/use-niimbot";
  import MdiLoading from "~icons/mdi/loading";
  import MdiBluetooth from "~icons/mdi/bluetooth";
  import MdiBluetoothConnect from "~icons/mdi/bluetooth-connect";

  const { t } = useI18n();

  const props = defineProps<{
    type: string;
    id: string;
  }>();

  const {
    isSupported,
    connected,
    deviceName,
    printing,
    printProgress,
    connect,
    disconnect,
    printImage,
    loadSavedTapeSize,
    saveTapeSize,
  } = useNiimbot();

  // --- Label variant ---
  type LabelVariant = "full" | "qr";
  const labelVariant = ref<LabelVariant>("full");

  // --- Tape size ---
  const selectedPresetIndex = ref(-1);
  const customWidth = ref(50);
  const customHeight = ref(30);
  const isCustomSize = ref(false);

  onMounted(() => {
    const saved = loadSavedTapeSize();
    const presetIdx = PRESET_TAPE_SIZES.findIndex(s => s.width === saved.width && s.height === saved.height);
    if (presetIdx >= 0) {
      selectedPresetIndex.value = presetIdx;
      isCustomSize.value = false;
    } else {
      isCustomSize.value = true;
      customWidth.value = saved.width;
      customHeight.value = saved.height;
    }
  });

  function getCurrentTapeSize(): TapeSize {
    if (isCustomSize.value) {
      return {
        label: `${customWidth.value} × ${customHeight.value} mm`,
        width: customWidth.value,
        height: customHeight.value,
      };
    }
    return PRESET_TAPE_SIZES[selectedPresetIndex.value] ?? PRESET_TAPE_SIZES[2];
  }

  function getLabelImageUrl(): string {
    const { selectedId } = useCollections();
    const params: Record<string, QueryValue> = {};
    if (selectedId.value) {
      params.tenant = selectedId.value;
    }

    if (labelVariant.value === "qr") {
      const pageUrl = `${window.location.origin}${window.location.pathname}`;
      return route("/qrcode", { data: pageUrl });
    }

    const validTypes = ["item", "location", "asset"];
    if (!validTypes.includes(props.type)) {
      throw new Error(`Unexpected type: ${props.type}`);
    }
    return route(`/labelmaker/${props.type}/${props.id}`, params);
  }

  async function handlePrint() {
    try {
      const tapeSize = getCurrentTapeSize();
      saveTapeSize(tapeSize);

      const url = getLabelImageUrl();
      await printImage(url, tapeSize);
      toast.success(t("components.niimbot.print_success"));
    } catch (e) {
      console.error("Niimbot: print error", e);
      const msg = e instanceof Error ? e.message : String(e);
      toast.error(t("components.niimbot.print_failed", { error: msg }));
    }
  }

  async function handleConnect() {
    try {
      if (connected.value) {
        await disconnect();
      } else {
        await connect();
      }
    } catch (e) {
      console.error("Niimbot: connection error", e);
      const msg = e instanceof Error ? e.message : String(e);
      toast.error(t("components.niimbot.connection_failed", { error: msg }));
    }
  }

  function onPresetChange(event: Event) {
    const value = (event.target as HTMLSelectElement).value;
    if (value === "custom") {
      isCustomSize.value = true;
      selectedPresetIndex.value = -1;
    } else {
      isCustomSize.value = false;
      selectedPresetIndex.value = Number(value);
    }
  }
</script>

<template>
  <div v-if="isSupported" class="mt-3 flex flex-col gap-3 border-t pt-3">
    <div class="flex items-center justify-between">
      <span class="text-sm font-medium">Niimbot</span>
      <Button size="sm" variant="outline" class="gap-1" @click="handleConnect">
        <MdiBluetoothConnect v-if="connected" class="text-blue-500" />
        <MdiBluetooth v-else />
        {{ connected ? deviceName : $t("components.niimbot.connect") }}
      </Button>
    </div>

    <!-- Label variant -->
    <div class="flex items-center gap-2">
      <label class="text-sm">{{ $t("components.niimbot.label") }}:</label>
      <select v-model="labelVariant" class="flex-1 rounded border bg-background px-2 py-1 text-sm">
        <option value="full">{{ $t("components.niimbot.label_full") }}</option>
        <option value="qr">{{ $t("components.niimbot.label_qr") }}</option>
      </select>
    </div>

    <!-- Tape size -->
    <div class="flex items-center gap-2">
      <label class="text-sm">{{ $t("components.niimbot.tape") }}:</label>
      <select
        class="flex-1 rounded border bg-background px-2 py-1 text-sm"
        :value="isCustomSize ? 'custom' : selectedPresetIndex"
        @change="onPresetChange"
      >
        <option v-for="(size, idx) in PRESET_TAPE_SIZES" :key="idx" :value="idx">
          {{ size.label }}
        </option>
        <option value="custom">{{ $t("components.niimbot.custom") }}</option>
      </select>
    </div>

    <!-- Custom size inputs -->
    <div v-if="isCustomSize" class="flex items-center gap-2">
      <input
        v-model.number="customWidth"
        type="number"
        min="10"
        max="100"
        class="w-16 rounded border bg-background px-2 py-1 text-sm"
        placeholder="W"
      />
      <span class="text-sm">x</span>
      <input
        v-model.number="customHeight"
        type="number"
        min="10"
        max="200"
        class="w-16 rounded border bg-background px-2 py-1 text-sm"
        placeholder="H"
      />
      <span class="text-sm text-muted-foreground">mm</span>
    </div>

    <!-- Print button -->
    <Button class="w-full gap-2" :disabled="printing" @click="handlePrint">
      <MdiLoading v-if="printing" class="animate-spin" />
      {{ printing ? $t("components.niimbot.printing", { progress: printProgress }) : $t("components.niimbot.print") }}
    </Button>
  </div>
</template>
