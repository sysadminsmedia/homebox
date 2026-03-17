import { ref, computed } from "vue";

// Lazy import to avoid SSR issues with Web Bluetooth
let niimbluelib: typeof import("@mmote/niimbluelib") | null = null;

async function loadNiimbluelib() {
  if (!niimbluelib) {
    try {
      niimbluelib = await import("@mmote/niimbluelib");
    } catch (e) {
      throw new Error("Failed to load @mmote/niimbluelib. Check that the package is installed.");
    }
  }
  return niimbluelib;
}

export type TapeSize = {
  label: string;
  width: number;
  height: number;
};

export const PRESET_TAPE_SIZES: TapeSize[] = [
  { label: "30 × 20 mm", width: 30, height: 20 },
  { label: "40 × 30 mm", width: 40, height: 30 },
  { label: "50 × 30 mm", width: 50, height: 30 },
  { label: "50 × 50 mm", width: 50, height: 50 },
  { label: "40 × 60 mm", width: 40, height: 60 },
];

const DPMM = 8; // 203 DPI = 8 dots per mm
const MIN_TAPE_MM = 10;
const MAX_TAPE_WIDTH_MM = 100;
const MAX_TAPE_HEIGHT_MM = 200;

const LS_TAPE_KEY = "niimbot_tape_size";

// Shared state across component instances
// eslint-disable-next-line @typescript-eslint/no-explicit-any -- niimbluelib alpha has no type declarations
const client = ref<any>(null);
const deviceName = ref<string | null>(null);
const connected = ref(false);
const printing = ref(false);
const printProgress = ref(0);

export function useNiimbot() {
  const isSupported = computed(() => typeof navigator !== "undefined" && "bluetooth" in navigator);

  function loadSavedTapeSize(): TapeSize {
    if (typeof localStorage === "undefined") return PRESET_TAPE_SIZES[2]; // 50x30 default
    const saved = localStorage.getItem(LS_TAPE_KEY);
    if (saved) {
      try {
        const parsed = JSON.parse(saved);
        if (
          Number.isFinite(parsed.width) &&
          Number.isFinite(parsed.height) &&
          parsed.width >= MIN_TAPE_MM &&
          parsed.width <= MAX_TAPE_WIDTH_MM &&
          parsed.height >= MIN_TAPE_MM &&
          parsed.height <= MAX_TAPE_HEIGHT_MM
        ) {
          return parsed;
        }
        localStorage.removeItem(LS_TAPE_KEY);
      } catch (e) {
        console.warn("Niimbot: corrupted tape size in localStorage, resetting", e);
        localStorage.removeItem(LS_TAPE_KEY);
      }
    }
    return PRESET_TAPE_SIZES[2];
  }

  function saveTapeSize(size: TapeSize) {
    if (typeof localStorage !== "undefined") {
      localStorage.setItem(LS_TAPE_KEY, JSON.stringify(size));
    }
  }

  async function connect(): Promise<boolean> {
    const lib = await loadNiimbluelib();
    const newClient = new lib.NiimbotBluetoothClient();

    newClient.on("disconnect", () => {
      connected.value = false;
      deviceName.value = null;
      client.value = null;
    });

    newClient.on("printprogress", event => {
      printProgress.value = event.pagePrintProgress;
    });

    try {
      const info = await newClient.connect();
      client.value = newClient;
      connected.value = true;
      deviceName.value = info.deviceName ?? "Niimbot";
      try {
        await newClient.fetchPrinterInfo();
      } catch (infoErr) {
        console.warn("Niimbot: connected but failed to fetch printer info", infoErr);
      }
      return true;
    } catch (e) {
      // User cancelled the Bluetooth picker — not an error
      if (e instanceof DOMException && e.name === "NotFoundError") {
        return false;
      }
      try {
        await newClient.disconnect();
      } catch (cleanupErr) {
        console.warn("Niimbot: cleanup disconnect failed", cleanupErr);
      }
      // Real error — throw so caller can show it
      throw e;
    }
  }

  async function disconnect() {
    if (client.value) {
      try {
        await client.value.disconnect();
      } catch (e) {
        console.warn("Niimbot: disconnect failed, cleaning up anyway", e);
      } finally {
        client.value = null;
        connected.value = false;
        deviceName.value = null;
      }
    }
  }

  async function printImage(imageUrl: string, tapeSize: TapeSize): Promise<void> {
    if (!client.value) {
      await connect(); // throws on real error, returns false on cancel
      if (!client.value) return; // user cancelled
    }

    const lib = await loadNiimbluelib();
    printing.value = true;
    printProgress.value = 0;

    // Capture client ref to avoid race condition if BLE disconnects mid-print
    const c = client.value;

    try {
      // With printDirection "top" (no rotation):
      //   canvas width  → cols (across printhead, must be multiple of 8)
      //   canvas height → rows (feed direction)
      // Printer does NOT auto-center. Canvas must span full printhead
      // width, with label content centered within it.
      if (!Number.isFinite(tapeSize.width) || !Number.isFinite(tapeSize.height)) {
        throw new Error("Invalid tape dimensions");
      }

      const meta = c.getModelMetadata?.();
      const printheadPx = meta?.printheadPixels ?? 384;
      const labelWidthPx = Math.round(tapeSize.width * DPMM);
      const labelHeightPx = Math.round(tapeSize.height * DPMM);

      if (labelWidthPx <= 0 || labelHeightPx <= 0) {
        throw new Error("Tape dimensions must be greater than zero");
      }
      if (labelWidthPx > printheadPx) {
        throw new Error(`Tape width exceeds printhead (${printheadPx / DPMM} mm max)`);
      }

      // Load and process label image
      const img = await loadImage(imageUrl);

      // First render label at tape dimensions
      const labelCanvas = resizeToCanvas(img, labelWidthPx, labelHeightPx);
      applyThreshold(labelCanvas, 128);

      // Then place centered on printhead-wide canvas
      const canvas = document.createElement("canvas");
      canvas.width = printheadPx;
      canvas.height = labelHeightPx;
      const ctx = canvas.getContext("2d");
      if (!ctx) throw new Error("Failed to create canvas context for label rendering");
      ctx.fillStyle = "white";
      ctx.fillRect(0, 0, canvas.width, canvas.height);
      const offsetX = Math.floor((printheadPx - labelWidthPx) / 2);
      ctx.drawImage(labelCanvas, offsetX, 0);

      // Encode for printer (no rotation for wide labels)
      const encoded = lib.ImageEncoder.encodeCanvas(canvas, "top");

      // Stop heartbeat during printing (can interfere with protocol)
      c.stopHeartbeat();

      // Get print task type from printer or fallback to B1
      const taskType = c.getPrintTaskType?.() ?? "B1";
      const printTask = c.abstraction.newPrintTask(taskType, {
        totalPages: 1,
        labelType: lib.LabelType.WithGaps,
        density: 3,
      });

      await printTask.printInit();
      await printTask.printPage(encoded, 1);
      await printTask.waitForFinished();

      printProgress.value = 100;
    } catch (e) {
      console.error("Niimbot print failed:", e);
      throw e;
    } finally {
      try {
        await c.abstraction?.printEnd();
      } catch (endErr) {
        console.warn("Niimbot: printEnd failed, printer may need power cycle", endErr);
      }
      c.startHeartbeat?.();
      printing.value = false;
    }
  }

  return {
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
  };
}

// --- Helper functions ---

function loadImage(url: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.crossOrigin = "anonymous";
    img.onload = () => resolve(img);
    img.onerror = () => reject(new Error(`Failed to load image: ${url}`));
    img.src = url;
  });
}

function resizeToCanvas(img: HTMLImageElement, targetW: number, targetH: number): HTMLCanvasElement {
  const canvas = document.createElement("canvas");
  canvas.width = targetW;
  canvas.height = targetH;
  const ctx = canvas.getContext("2d");
  if (!ctx) throw new Error("Failed to create canvas context");

  // White background
  ctx.fillStyle = "white";
  ctx.fillRect(0, 0, targetW, targetH);

  // Fit image preserving aspect ratio
  const scale = Math.min(targetW / img.width, targetH / img.height);
  const w = img.width * scale;
  const h = img.height * scale;
  const x = (targetW - w) / 2;
  const y = (targetH - h) / 2;
  ctx.drawImage(img, x, y, w, h);

  return canvas;
}

function applyThreshold(canvas: HTMLCanvasElement, threshold: number) {
  const ctx = canvas.getContext("2d");
  if (!ctx) throw new Error("Failed to get canvas context for threshold");
  const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height);
  const data = imageData.data;

  for (let i = 0; i < data.length; i += 4) {
    const luminance = data[i] * 0.299 + data[i + 1] * 0.587 + data[i + 2] * 0.114;
    const val = luminance < threshold ? 0 : 255;
    data[i] = val;
    data[i + 1] = val;
    data[i + 2] = val;
    data[i + 3] = 255;
  }

  ctx.putImageData(imageData, 0, 0);
}
