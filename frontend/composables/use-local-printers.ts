import type { LocalPrinter, PrintResult, PrintJobOptions } from "~~/lib/printing/types";
import {
  isLocalPrintingSupported,
  getSavedLocalPrinters,
  removeLocalPrinter,
  pairLocalPrinter,
  printToLocalPrinter,
} from "~~/lib/printing";

export function useLocalPrinters() {
  const localPrinters = ref<LocalPrinter[]>([]);
  const isSupported = ref({ usb: false, bluetooth: false });

  // Check browser support on mount
  onMounted(() => {
    isSupported.value = isLocalPrintingSupported();
    loadSavedPrinters();
  });

  function loadSavedPrinters() {
    localPrinters.value = getSavedLocalPrinters();
  }

  async function pairUSBPrinter(): Promise<LocalPrinter | null> {
    if (!isSupported.value.usb) {
      throw new Error("WebUSB is not supported in this browser");
    }

    const printer = await pairLocalPrinter("usb");
    if (printer) {
      loadSavedPrinters();
    }
    return printer;
  }

  async function pairBluetoothPrinter(): Promise<LocalPrinter | null> {
    if (!isSupported.value.bluetooth) {
      throw new Error("Web Bluetooth is not supported in this browser");
    }

    const printer = await pairLocalPrinter("bluetooth");
    if (printer) {
      loadSavedPrinters();
    }
    return printer;
  }

  function removePrinter(printerId: string) {
    removeLocalPrinter(printerId);
    loadSavedPrinters();
  }

  return {
    localPrinters,
    isSupported,
    pairUSBPrinter,
    pairBluetoothPrinter,
    removePrinter,
    refresh: loadSavedPrinters,
  };
}

export function useLocalPrint() {
  const isPrinting = ref(false);
  const lastError = ref<string | null>(null);

  async function print(
    printer: LocalPrinter,
    imageData: Uint8Array,
    options: PrintJobOptions = {}
  ): Promise<PrintResult> {
    isPrinting.value = true;
    lastError.value = null;

    try {
      const result = await printToLocalPrinter(printer, imageData, options);
      if (!result.success) {
        lastError.value = result.message;
      }
      return result;
    } catch (error) {
      const message = (error as Error).message;
      lastError.value = message;
      return { success: false, message };
    } finally {
      isPrinting.value = false;
    }
  }

  return {
    print,
    isPrinting,
    lastError,
  };
}
