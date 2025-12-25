// WebUSB printer handling for direct USB printing from the browser
// Note: WebUSB only works in Chrome/Edge browsers

import {
  type LocalPrinter,
  type PrintResult,
  type PrintJobOptions,
  USB_PRINTER_FILTERS,
  identifyPrinter,
} from "./types";
import { BrotherRasterProtocol } from "./protocols/brother-raster";

export class WebUSBPrinter {
  private device: USBDevice | null = null;
  private interfaceNumber: number = 0;
  private endpointOut: number = 0;
  private endpointIn: number = 0;

  static isSupported(): boolean {
    return "usb" in navigator;
  }

  async requestDevice(): Promise<LocalPrinter | null> {
    if (!WebUSBPrinter.isSupported()) {
      throw new Error("WebUSB is not supported in this browser. Please use Chrome or Edge.");
    }

    try {
      const device = await navigator.usb.requestDevice({
        filters: USB_PRINTER_FILTERS,
      });

      const printerInfo = identifyPrinter(device.vendorId, device.productId);

      return {
        id: `usb-${device.vendorId}-${device.productId}-${device.serialNumber || "unknown"}`,
        name: printerInfo?.name || device.productName || "Unknown USB Printer",
        connectionType: "usb",
        protocol: printerInfo?.protocol || "escpos",
        vendorId: device.vendorId,
        productId: device.productId,
        deviceInfo: device,
      };
    } catch (error) {
      if ((error as Error).name === "NotFoundError") {
        return null; // User cancelled
      }
      throw error;
    }
  }

  async connect(printer: LocalPrinter): Promise<void> {
    // Check that we have a USB device by checking for USB-specific properties
    if (!printer.deviceInfo || !("vendorId" in printer.deviceInfo)) {
      throw new Error("Invalid USB device");
    }

    this.device = printer.deviceInfo as USBDevice;

    await this.device.open();

    // Select configuration (usually 1)
    if (this.device.configuration === null) {
      await this.device.selectConfiguration(1);
    }

    // Find the printer interface
    const configuration = this.device.configuration;
    if (!configuration) {
      throw new Error("No USB configuration found");
    }

    for (const iface of configuration.interfaces) {
      // Look for printer class (7) or vendor-specific interface
      const alternate = iface.alternates[0];
      if (!alternate) continue;

      if (alternate.interfaceClass === 7 || alternate.interfaceClass === 255) {
        this.interfaceNumber = iface.interfaceNumber;

        // Find bulk endpoints
        for (const endpoint of alternate.endpoints) {
          if (endpoint.type === "bulk") {
            if (endpoint.direction === "out") {
              this.endpointOut = endpoint.endpointNumber;
            } else {
              this.endpointIn = endpoint.endpointNumber;
            }
          }
        }
        break;
      }
    }

    await this.device.claimInterface(this.interfaceNumber);
  }

  async disconnect(): Promise<void> {
    if (this.device) {
      try {
        await this.device.releaseInterface(this.interfaceNumber);
        await this.device.close();
      } catch {
        // Ignore errors during disconnect
      }
      this.device = null;
    }
  }

  async print(imageData: Uint8Array, protocol: string, options: PrintJobOptions = {}): Promise<PrintResult> {
    if (!this.device) {
      return { success: false, message: "Printer not connected" };
    }

    try {
      let printData: Uint8Array;

      switch (protocol) {
        case "brother-raster":
          printData = BrotherRasterProtocol.createPrintJob(imageData, {
            width: options.labelWidth || 62,
            height: options.labelHeight,
            copies: options.copies || 1,
          });
          break;
        // Add other protocols here
        default:
          // For unknown protocols, send raw data
          printData = imageData;
      }

      // Send data in chunks (max 64KB per transfer)
      const chunkSize = 64 * 1024;
      for (let offset = 0; offset < printData.length; offset += chunkSize) {
        const chunk = printData.slice(offset, offset + chunkSize);
        await this.device.transferOut(this.endpointOut, chunk);
      }

      return { success: true, message: "Print job sent successfully" };
    } catch (error) {
      return {
        success: false,
        message: `Print failed: ${(error as Error).message}`,
      };
    }
  }

  async getStatus(): Promise<string> {
    if (!this.device || !this.endpointIn) {
      return "disconnected";
    }

    try {
      const result = await this.device.transferIn(this.endpointIn, 32);
      if (result.data) {
        // Parse status based on protocol
        // For now, just return connected
        return "connected";
      }
      return "unknown";
    } catch {
      return "error";
    }
  }
}

// Singleton instance for easy access
let webUSBPrinter: WebUSBPrinter | null = null;

export function getWebUSBPrinter(): WebUSBPrinter {
  if (!webUSBPrinter) {
    webUSBPrinter = new WebUSBPrinter();
  }
  return webUSBPrinter;
}
