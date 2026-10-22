// Types for client-side direct printing via WebUSB and Web Bluetooth

export type PrinterConnectionType = "usb" | "bluetooth";

export type PrinterProtocol = "brother-raster" | "escpos" | "zpl";

export interface LocalPrinter {
  id: string;
  name: string;
  connectionType: PrinterConnectionType;
  protocol: PrinterProtocol;
  vendorId?: number;
  productId?: number;
  // For reconnection
  deviceInfo?: USBDevice | BluetoothDevice;
}

export interface PrintJobOptions {
  copies?: number;
  labelWidth?: number; // mm
  labelHeight?: number; // mm
  dpi?: number;
}

export interface PrintResult {
  success: boolean;
  message: string;
}

// Brother QL specific types
export interface BrotherQLMediaInfo {
  mediaType: "continuous" | "die-cut";
  width: number; // mm
  length?: number; // mm, only for die-cut
}

// WebUSB device filter for common label printers
export const USB_PRINTER_FILTERS: USBDeviceFilter[] = [
  // Brother
  { vendorId: 0x04f9 }, // Brother Industries
  // Dymo
  { vendorId: 0x0922 }, // Dymo-CoStar
  // Zebra
  { vendorId: 0x0a5f }, // Zebra Technologies
];

// Web Bluetooth service UUIDs for printers
export const BLUETOOTH_PRINTER_SERVICES = {
  // Standard Serial Port Profile
  serialPort: "00001101-0000-1000-8000-00805f9b34fb",
  // Brother-specific
  brotherPrint: "e7a60000-6639-4b39-8f6c-8c85c11ac38e",
};

// Known printer models and their protocols
export const KNOWN_PRINTERS: Record<string, { protocol: PrinterProtocol; name: string }> = {
  // Brother QL series (vendorId: 0x04f9)
  "04f9:209b": { protocol: "brother-raster", name: "Brother QL-800" },
  "04f9:209c": { protocol: "brother-raster", name: "Brother QL-810W" },
  "04f9:209d": { protocol: "brother-raster", name: "Brother QL-820NWB" },
  "04f9:2042": { protocol: "brother-raster", name: "Brother QL-700" },
  "04f9:2049": { protocol: "brother-raster", name: "Brother QL-710W" },
  "04f9:2061": { protocol: "brother-raster", name: "Brother QL-720NW" },
  // Dymo LabelWriter series
  "0922:0028": { protocol: "escpos", name: "Dymo LabelWriter 450" },
  "0922:002a": { protocol: "escpos", name: "Dymo LabelWriter 450 Turbo" },
  // Generic thermal printers default to ESC/POS
};

export function identifyPrinter(
  vendorId: number,
  productId: number
): { protocol: PrinterProtocol; name: string } | null {
  const key = `${vendorId.toString(16).padStart(4, "0")}:${productId.toString(16).padStart(4, "0")}`;
  return KNOWN_PRINTERS[key] || null;
}
