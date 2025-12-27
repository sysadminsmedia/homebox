// Web Bluetooth printer handling for direct Bluetooth printing from the browser
// Note: Web Bluetooth only works in Chrome/Edge browsers

import { type LocalPrinter, type PrintResult, type PrintJobOptions, BLUETOOTH_PRINTER_SERVICES } from "./types";
import { BrotherRasterProtocol } from "./protocols/brother-raster";

export class WebBluetoothPrinter {
  private device: BluetoothDevice | null = null;
  private server: BluetoothRemoteGATTServer | null = null;
  private characteristic: BluetoothRemoteGATTCharacteristic | null = null;

  static isSupported(): boolean {
    return "bluetooth" in navigator;
  }

  async requestDevice(): Promise<LocalPrinter | null> {
    if (!WebBluetoothPrinter.isSupported()) {
      throw new Error("Web Bluetooth is not supported in this browser. Please use Chrome or Edge.");
    }

    try {
      const device = await navigator.bluetooth.requestDevice({
        // Accept any device with a name containing "Brother", "Dymo", "Zebra", or "Printer"
        filters: [
          { namePrefix: "Brother" },
          { namePrefix: "QL-" },
          { namePrefix: "Dymo" },
          { namePrefix: "Zebra" },
          { namePrefix: "Printer" },
        ],
        optionalServices: [BLUETOOTH_PRINTER_SERVICES.serialPort, BLUETOOTH_PRINTER_SERVICES.brotherPrint],
      });

      // Determine protocol based on device name
      let protocol: "brother-raster" | "escpos" | "zpl" = "escpos";
      const name = device.name?.toLowerCase() || "";
      if (name.includes("brother") || name.includes("ql-")) {
        protocol = "brother-raster";
      } else if (name.includes("zebra")) {
        protocol = "zpl";
      }

      return {
        id: `bt-${device.id}`,
        name: device.name || "Unknown Bluetooth Printer",
        connectionType: "bluetooth",
        protocol,
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
    // Check that we have a Bluetooth device by checking for Bluetooth-specific properties
    if (!printer.deviceInfo || !("gatt" in printer.deviceInfo)) {
      throw new Error("Invalid Bluetooth device");
    }

    this.device = printer.deviceInfo as BluetoothDevice;

    // Connect to GATT server
    this.server = (await this.device.gatt?.connect()) || null;
    if (!this.server) {
      throw new Error("Failed to connect to Bluetooth device");
    }

    // Try to find a writable characteristic
    // First try Brother-specific service
    try {
      const service = await this.server.getPrimaryService(BLUETOOTH_PRINTER_SERVICES.brotherPrint);
      const characteristics = await service.getCharacteristics();

      for (const char of characteristics) {
        if (char.properties.write || char.properties.writeWithoutResponse) {
          this.characteristic = char;
          break;
        }
      }
    } catch {
      // Brother service not found, try Serial Port Profile
    }

    // Try Serial Port Profile if Brother service didn't work
    if (!this.characteristic) {
      try {
        const service = await this.server.getPrimaryService(BLUETOOTH_PRINTER_SERVICES.serialPort);
        const characteristics = await service.getCharacteristics();

        for (const char of characteristics) {
          if (char.properties.write || char.properties.writeWithoutResponse) {
            this.characteristic = char;
            break;
          }
        }
      } catch {
        // Serial port service not found
      }
    }

    if (!this.characteristic) {
      throw new Error("No writable characteristic found on printer");
    }
  }

  async disconnect(): Promise<void> {
    if (this.server?.connected) {
      this.server.disconnect();
    }
    this.device = null;
    this.server = null;
    this.characteristic = null;
  }

  async print(imageData: Uint8Array, protocol: string, options: PrintJobOptions = {}): Promise<PrintResult> {
    if (!this.characteristic) {
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
        default:
          printData = imageData;
      }

      // Send data in chunks (Bluetooth typically has smaller MTU)
      const chunkSize = 512; // Conservative chunk size for BLE
      for (let offset = 0; offset < printData.length; offset += chunkSize) {
        const chunk = printData.slice(offset, offset + chunkSize);

        if (this.characteristic.properties.writeWithoutResponse) {
          await this.characteristic.writeValueWithoutResponse(chunk);
        } else {
          await this.characteristic.writeValue(chunk);
        }

        // Small delay between chunks to prevent buffer overflow
        await new Promise(resolve => setTimeout(resolve, 20));
      }

      return { success: true, message: "Print job sent successfully" };
    } catch (error) {
      return {
        success: false,
        message: `Print failed: ${(error as Error).message}`,
      };
    }
  }

  isConnected(): boolean {
    return this.server?.connected || false;
  }
}

// Singleton instance
let webBluetoothPrinter: WebBluetoothPrinter | null = null;

export function getWebBluetoothPrinter(): WebBluetoothPrinter {
  if (!webBluetoothPrinter) {
    webBluetoothPrinter = new WebBluetoothPrinter();
  }
  return webBluetoothPrinter;
}
