// Client-side printing module for direct USB and Bluetooth printing
// Supports Brother QL series, and can be extended for other thermal printers

import type { LocalPrinter, PrintResult, PrintJobOptions } from "./types";
import { WebUSBPrinter, getWebUSBPrinter } from "./webusb-printer";
import { WebBluetoothPrinter, getWebBluetoothPrinter } from "./webbluetooth-printer";

export * from "./types";
export * from "./webusb-printer";
export * from "./webbluetooth-printer";

// Storage key for saved printers
const STORAGE_KEY = "homebox-local-printers";

/**
 * Check if local printing is supported in this browser
 */
export function isLocalPrintingSupported(): { usb: boolean; bluetooth: boolean } {
  return {
    usb: WebUSBPrinter.isSupported(),
    bluetooth: WebBluetoothPrinter.isSupported(),
  };
}

/**
 * Get saved local printers from localStorage
 */
export function getSavedLocalPrinters(): LocalPrinter[] {
  if (typeof window === "undefined") return [];

  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved) {
      return JSON.parse(saved);
    }
  } catch {
    // Ignore parse errors
  }
  return [];
}

/**
 * Save a local printer to localStorage
 */
export function saveLocalPrinter(printer: LocalPrinter): void {
  if (typeof window === "undefined") return;

  const printers = getSavedLocalPrinters();

  // Remove device info before saving (can't serialize)
  const printerToSave = { ...printer, deviceInfo: undefined };

  // Check if printer already exists
  const existingIndex = printers.findIndex(p => p.id === printer.id);
  if (existingIndex >= 0) {
    printers[existingIndex] = printerToSave;
  } else {
    printers.push(printerToSave);
  }

  localStorage.setItem(STORAGE_KEY, JSON.stringify(printers));
}

/**
 * Remove a saved local printer
 */
export function removeLocalPrinter(printerId: string): void {
  if (typeof window === "undefined") return;

  const printers = getSavedLocalPrinters();
  const filtered = printers.filter(p => p.id !== printerId);
  localStorage.setItem(STORAGE_KEY, JSON.stringify(filtered));
}

/**
 * Request and pair a new local printer (USB or Bluetooth)
 */
export async function pairLocalPrinter(type: "usb" | "bluetooth"): Promise<LocalPrinter | null> {
  if (type === "usb") {
    const usbPrinter = getWebUSBPrinter();
    const printer = await usbPrinter.requestDevice();
    if (printer) {
      saveLocalPrinter(printer);
    }
    return printer;
  } else {
    const btPrinter = getWebBluetoothPrinter();
    const printer = await btPrinter.requestDevice();
    if (printer) {
      saveLocalPrinter(printer);
    }
    return printer;
  }
}

/**
 * Print to a local printer
 */
export async function printToLocalPrinter(
  printer: LocalPrinter,
  imageData: Uint8Array,
  options: PrintJobOptions = {}
): Promise<PrintResult> {
  if (printer.connectionType === "usb") {
    const usbPrinter = getWebUSBPrinter();
    await usbPrinter.connect(printer);
    try {
      return await usbPrinter.print(imageData, printer.protocol, options);
    } finally {
      await usbPrinter.disconnect();
    }
  } else {
    const btPrinter = getWebBluetoothPrinter();
    await btPrinter.connect(printer);
    try {
      return await btPrinter.print(imageData, printer.protocol, options);
    } finally {
      await btPrinter.disconnect();
    }
  }
}
