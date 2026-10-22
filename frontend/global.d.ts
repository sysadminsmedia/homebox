/// <reference types="unplugin-icons/types/vue" />

declare module "#app" {
  interface NuxtApp {
    $otelEnabled: boolean;
  }
}

declare module "vue" {
  interface ComponentCustomProperties {
    $otelEnabled: boolean;
  }
}

declare global {
  /**
   * WebUSB and Web Bluetooth API Type Declarations
   *
   * These type declarations are necessary because TypeScript does not include
   * WebUSB or Web Bluetooth APIs by default - they are experimental browser APIs
   * that are not part of the standard lib.dom.d.ts types.
   *
   * These APIs are used by the label printing feature to enable direct USB and
   * Bluetooth connections to label printers (e.g., Brother QL-series) from the
   * browser, without requiring server-side print handling.
   *
   * Used by:
   * - frontend/lib/printing/webusb-printer.ts
   * - frontend/lib/printing/webbluetooth-printer.ts
   *
   * References:
   * - WebUSB: https://wicg.github.io/webusb/
   * - Web Bluetooth: https://webbluetoothcg.github.io/web-bluetooth/
   */

  // WebUSB API Type Declarations
  // https://wicg.github.io/webusb/

  interface USBDeviceFilter {
    vendorId?: number;
    productId?: number;
    classCode?: number;
    subclassCode?: number;
    protocolCode?: number;
    serialNumber?: string;
  }

  interface USBDeviceRequestOptions {
    filters: USBDeviceFilter[];
  }

  interface USBEndpoint {
    endpointNumber: number;
    direction: "in" | "out";
    type: "bulk" | "interrupt" | "isochronous";
    packetSize: number;
  }

  interface USBAlternateInterface {
    alternateSetting: number;
    interfaceClass: number;
    interfaceSubclass: number;
    interfaceProtocol: number;
    interfaceName?: string;
    endpoints: USBEndpoint[];
  }

  interface USBInterface {
    interfaceNumber: number;
    alternate: USBAlternateInterface;
    alternates: USBAlternateInterface[];
    claimed: boolean;
  }

  interface USBConfiguration {
    configurationValue: number;
    configurationName?: string;
    interfaces: USBInterface[];
  }

  interface USBInTransferResult {
    data?: DataView;
    status: "ok" | "stall" | "babble";
  }

  interface USBOutTransferResult {
    bytesWritten: number;
    status: "ok" | "stall";
  }

  interface USBDevice {
    readonly vendorId: number;
    readonly productId: number;
    readonly deviceClass: number;
    readonly deviceSubclass: number;
    readonly deviceProtocol: number;
    readonly deviceVersionMajor: number;
    readonly deviceVersionMinor: number;
    readonly deviceVersionSubminor: number;
    readonly manufacturerName?: string;
    readonly productName?: string;
    readonly serialNumber?: string;
    readonly configuration?: USBConfiguration;
    readonly configurations: USBConfiguration[];
    readonly opened: boolean;
    open(): Promise<void>;
    close(): Promise<void>;
    selectConfiguration(configurationValue: number): Promise<void>;
    claimInterface(interfaceNumber: number): Promise<void>;
    releaseInterface(interfaceNumber: number): Promise<void>;
    selectAlternateInterface(interfaceNumber: number, alternateSetting: number): Promise<void>;
    transferIn(endpointNumber: number, length: number): Promise<USBInTransferResult>;
    transferOut(endpointNumber: number, data: BufferSource): Promise<USBOutTransferResult>;
    clearHalt(direction: "in" | "out", endpointNumber: number): Promise<void>;
    reset(): Promise<void>;
  }

  interface USB {
    getDevices(): Promise<USBDevice[]>;
    requestDevice(options: USBDeviceRequestOptions): Promise<USBDevice>;
  }

  // Web Bluetooth API Type Declarations
  // https://webbluetoothcg.github.io/web-bluetooth/

  interface BluetoothRemoteGATTDescriptor {
    readonly characteristic: BluetoothRemoteGATTCharacteristic;
    readonly uuid: string;
    readonly value?: DataView;
    readValue(): Promise<DataView>;
    writeValue(value: BufferSource): Promise<void>;
  }

  interface BluetoothCharacteristicProperties {
    readonly broadcast: boolean;
    readonly read: boolean;
    readonly writeWithoutResponse: boolean;
    readonly write: boolean;
    readonly notify: boolean;
    readonly indicate: boolean;
    readonly authenticatedSignedWrites: boolean;
    readonly reliableWrite: boolean;
    readonly writableAuxiliaries: boolean;
  }

  interface BluetoothRemoteGATTCharacteristic extends EventTarget {
    readonly service: BluetoothRemoteGATTService;
    readonly uuid: string;
    readonly properties: BluetoothCharacteristicProperties;
    readonly value?: DataView;
    getDescriptor(descriptor: string): Promise<BluetoothRemoteGATTDescriptor>;
    getDescriptors(descriptor?: string): Promise<BluetoothRemoteGATTDescriptor[]>;
    readValue(): Promise<DataView>;
    writeValue(value: BufferSource): Promise<void>;
    writeValueWithResponse(value: BufferSource): Promise<void>;
    writeValueWithoutResponse(value: BufferSource): Promise<void>;
    startNotifications(): Promise<BluetoothRemoteGATTCharacteristic>;
    stopNotifications(): Promise<BluetoothRemoteGATTCharacteristic>;
  }

  interface BluetoothRemoteGATTService extends EventTarget {
    readonly device: BluetoothDevice;
    readonly uuid: string;
    readonly isPrimary: boolean;
    getCharacteristic(characteristic: string): Promise<BluetoothRemoteGATTCharacteristic>;
    getCharacteristics(characteristic?: string): Promise<BluetoothRemoteGATTCharacteristic[]>;
    getIncludedService(service: string): Promise<BluetoothRemoteGATTService>;
    getIncludedServices(service?: string): Promise<BluetoothRemoteGATTService[]>;
  }

  interface BluetoothRemoteGATTServer {
    readonly device: BluetoothDevice;
    readonly connected: boolean;
    connect(): Promise<BluetoothRemoteGATTServer>;
    disconnect(): void;
    getPrimaryService(service: string): Promise<BluetoothRemoteGATTService>;
    getPrimaryServices(service?: string): Promise<BluetoothRemoteGATTService[]>;
  }

  interface BluetoothDevice extends EventTarget {
    readonly id: string;
    readonly name?: string;
    readonly gatt?: BluetoothRemoteGATTServer;
    watchAdvertisements(): Promise<void>;
    unwatchAdvertisements(): void;
  }

  interface BluetoothLEScanFilterInit {
    services?: string[];
    name?: string;
    namePrefix?: string;
    manufacturerData?: Map<number, DataView>;
    serviceData?: Map<string, DataView>;
  }

  interface RequestDeviceOptions {
    filters?: BluetoothLEScanFilterInit[];
    optionalServices?: string[];
    acceptAllDevices?: boolean;
  }

  interface Bluetooth extends EventTarget {
    getDevices(): Promise<BluetoothDevice[]>;
    requestDevice(options: RequestDeviceOptions): Promise<BluetoothDevice>;
  }

  // Extend Navigator interface
  interface Navigator {
    usb: USB;
    bluetooth: Bluetooth;
  }
}

export {};
