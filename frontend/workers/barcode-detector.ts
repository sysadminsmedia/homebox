import { BarcodeDetector } from "barcode-detector";

interface DetectRequest {
  type: "detect";
  frame: ImageBitmap;
  id: number;
}

interface DetectResult {
  type: "result";
  id: number;
  barcodes: Array<{
    rawValue: string;
    format: string;
    boundingBox: { x: number; y: number; width: number; height: number };
    cornerPoints: Array<{ x: number; y: number }>;
  }>;
}

interface ReadyMessage {
  type: "ready";
}

interface ErrorMessage {
  type: "error";
  message: string;
}

export type WorkerMessage = DetectRequest;
export type WorkerResponse = DetectResult | ReadyMessage | ErrorMessage;

let detector: BarcodeDetector | null = null;

async function init() {
  try {
    detector = new BarcodeDetector({ formats: ["qr_code"] });
    self.postMessage({ type: "ready" } satisfies ReadyMessage);
  } catch (e) {
    self.postMessage({ type: "error", message: `BarcodeDetector init failed: ${e}` } satisfies ErrorMessage);
  }
}

self.onmessage = async (event: MessageEvent<WorkerMessage>) => {
  const msg = event.data;

  if (msg.type === "detect") {
    if (!detector) {
      self.postMessage({ type: "result", id: msg.id, barcodes: [] } satisfies DetectResult);
      msg.frame.close();
      return;
    }

    try {
      const results = await detector.detect(msg.frame);

      const barcodes = results.map(b => ({
        rawValue: b.rawValue,
        format: b.format,
        boundingBox: {
          x: b.boundingBox.x,
          y: b.boundingBox.y,
          width: b.boundingBox.width,
          height: b.boundingBox.height,
        },
        cornerPoints: Array.from(b.cornerPoints).map(p => ({ x: p.x, y: p.y })),
      }));

      self.postMessage({ type: "result", id: msg.id, barcodes } satisfies DetectResult);
    } catch {
      self.postMessage({ type: "result", id: msg.id, barcodes: [] } satisfies DetectResult);
    } finally {
      msg.frame.close();
    }
  }
};

init();
