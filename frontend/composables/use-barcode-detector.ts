import { ref, type Ref, watch, onUnmounted } from "vue";
import { useDocumentVisibility } from "@vueuse/core";
import type { ItemOut, ItemSummary, LocationOut } from "~~/lib/api/types/data-contracts";
import type { WorkerResponse } from "~~/workers/barcode-detector";

export interface EntityData {
  item?: ItemOut;
  location?: LocationOut;
  childItems?: ItemSummary[];
}

export interface Point2D {
  x: number;
  y: number;
}

/** 3D pose derived from QR code corner points */
export interface Pose3D {
  centerX: number;
  centerY: number;
  rotateZ: number;
  rotateX: number;
  rotateY: number;
  scale: number;
}

export interface DetectedEntity {
  id: string;
  entityType: "item" | "location" | "asset";
  rawValue: string;
  boundingBox: DOMRect;
  cornerPoints: Point2D[];
  pose: Pose3D;
  data: EntityData | null;
  loading: boolean;
  error: boolean;
  lastSeen: number;
}

interface CachedEntity {
  data: EntityData;
  fetchedAt: number;
}

const CACHE_TTL = 60_000;
const STALE_TIMEOUT = 1500;
const MAX_DETECTIONS = 10;
const SMOOTH = 0.6;
/** Minimum ms between detection requests sent to worker */
const DETECT_INTERVAL = 50; // ~20fps detection, rAF loop runs at display rate for smooth lerp

function lerpVal(prev: number, next: number, t: number): number {
  return prev * t + next * (1 - t);
}

function lerpPoint(prev: Point2D, next: Point2D, t: number): Point2D {
  return { x: lerpVal(prev.x, next.x, t), y: lerpVal(prev.y, next.y, t) };
}

function lerpRect(prev: DOMRect, next: DOMRect, t: number): DOMRect {
  return new DOMRect(
    lerpVal(prev.x, next.x, t),
    lerpVal(prev.y, next.y, t),
    lerpVal(prev.width, next.width, t),
    lerpVal(prev.height, next.height, t)
  );
}

function lerpCorners(prev: Point2D[], next: Point2D[], t: number): Point2D[] {
  if (prev.length !== next.length) return next;
  return prev.map((p, i) => lerpPoint(p, next[i]!, t));
}

function parseHomeboxUrl(rawValue: string): { entityType: "item" | "location" | "asset"; id: string } | null {
  try {
    let pathname: string;
    try {
      const url = new URL(rawValue);
      pathname = url.pathname;
    } catch {
      if (rawValue.startsWith("/")) {
        pathname = rawValue;
      } else {
        return null;
      }
    }

    const sanitized = pathname.replace(/[^a-zA-Z0-9-_/]/g, "");

    const assetMatch = sanitized.match(/^\/a\/([a-zA-Z0-9-_]+)/);
    if (assetMatch) return { entityType: "asset", id: assetMatch[1]! };

    const itemMatch = sanitized.match(/^\/item\/([a-zA-Z0-9-_]+)/);
    if (itemMatch) return { entityType: "item", id: itemMatch[1]! };

    const locationMatch = sanitized.match(/^\/location\/([a-zA-Z0-9-_]+)/);
    if (locationMatch) return { entityType: "location", id: locationMatch[1]! };

    return null;
  } catch {
    return null;
  }
}

interface VideoTransform {
  scaleX: number;
  scaleY: number;
  offsetX: number;
  offsetY: number;
}

function getVideoTransform(videoW: number, videoH: number, clientW: number, clientH: number): VideoTransform {
  if (!videoW || !videoH || !clientW || !clientH) {
    return { scaleX: 1, scaleY: 1, offsetX: 0, offsetY: 0 };
  }

  const videoAspect = videoW / videoH;
  const clientAspect = clientW / clientH;

  if (videoAspect > clientAspect) {
    const s = clientH / videoH;
    return { scaleX: s, scaleY: s, offsetX: (clientW - videoW * s) / 2, offsetY: 0 };
  } else {
    const s = clientW / videoW;
    return { scaleX: s, scaleY: s, offsetX: 0, offsetY: (clientH - videoH * s) / 2 };
  }
}

function mapBox(bbox: { x: number; y: number; width: number; height: number }, t: VideoTransform): DOMRect {
  return new DOMRect(
    bbox.x * t.scaleX + t.offsetX,
    bbox.y * t.scaleY + t.offsetY,
    bbox.width * t.scaleX,
    bbox.height * t.scaleY
  );
}

function mapCorners(corners: Array<{ x: number; y: number }>, t: VideoTransform): Point2D[] {
  return corners.map(c => ({
    x: c.x * t.scaleX + t.offsetX,
    y: c.y * t.scaleY + t.offsetY,
  }));
}

function ptDist(a: Point2D, b: Point2D): number {
  return Math.sqrt((a.x - b.x) ** 2 + (a.y - b.y) ** 2);
}

const REFERENCE_QR_SIZE = 200;

function computePose(corners: Point2D[]): Pose3D {
  if (corners.length < 4) {
    return { centerX: 0, centerY: 0, rotateZ: 0, rotateX: 0, rotateY: 0, scale: 1 };
  }

  const [tl, tr, br, bl] = corners as [Point2D, Point2D, Point2D, Point2D];

  const centerX = (tl.x + tr.x + br.x + bl.x) / 4;
  const centerY = (tl.y + tr.y + br.y + bl.y) / 4;
  const rotateZ = Math.atan2(tr.y - tl.y, tr.x - tl.x) * (180 / Math.PI);

  const topEdge = ptDist(tl, tr);
  const bottomEdge = ptDist(bl, br);
  const xRatio = bottomEdge / (topEdge || 1);
  const rotateX = Math.max(-60, Math.min(60, (xRatio - 1) * 50));

  const leftEdge = ptDist(tl, bl);
  const rightEdge = ptDist(tr, br);
  const yRatio = rightEdge / (leftEdge || 1);
  const rotateY = Math.max(-60, Math.min(60, (1 - yRatio) * 50));

  const avgEdge = (topEdge + bottomEdge + leftEdge + rightEdge) / 4;
  const scale = Math.max(0.3, Math.min(2.0, avgEdge / REFERENCE_QR_SIZE));

  return { centerX, centerY, rotateZ, rotateX, rotateY, scale };
}

export function solveHomography(
  src: [Point2D, Point2D, Point2D, Point2D],
  dst: [Point2D, Point2D, Point2D, Point2D]
): number[] {
  const aug: number[][] = [];
  for (let i = 0; i < 4; i++) {
    const { x, y } = src[i]!;
    const { x: u, y: v } = dst[i]!;
    aug.push([x, y, 1, 0, 0, 0, -x * u, -y * u, u]);
    aug.push([0, 0, 0, x, y, 1, -x * v, -y * v, v]);
  }

  const n = 8;
  for (let col = 0; col < n; col++) {
    let maxRow = col;
    for (let row = col + 1; row < n; row++) {
      if (Math.abs(aug[row]![col]!) > Math.abs(aug[maxRow]![col]!)) maxRow = row;
    }
    [aug[col], aug[maxRow]] = [aug[maxRow]!, aug[col]!];
    const pivot = aug[col]![col]!;
    if (Math.abs(pivot) < 1e-10) return [1, 0, 0, 0, 1, 0, 0, 0, 1];
    for (let row = col + 1; row < n; row++) {
      const factor = aug[row]![col]! / pivot;
      for (let j = col; j <= n; j++) {
        aug[row]![j]! -= factor * aug[col]![j]!;
      }
    }
  }

  const h = new Array(n).fill(0);
  for (let i = n - 1; i >= 0; i--) {
    h[i] = aug[i]![n]!;
    for (let j = i + 1; j < n; j++) {
      h[i] -= aug[i]![j]! * h[j];
    }
    h[i] /= aug[i]![i]!;
  }

  return [...h, 1];
}

export function homographyToMatrix3d(H: number[]): string {
  return `matrix3d(${H[0]}, ${H[3]}, 0, ${H[6]}, ${H[1]}, ${H[4]}, 0, ${H[7]}, 0, 0, 1, 0, ${H[2]}, ${H[5]}, 0, ${H[8]})`;
}

export function useBarcodeDetector(videoRef: Ref<HTMLVideoElement | undefined>) {
  const isSupported = ref(false);
  const isScanning = ref(false);
  const error = ref<string | null>(null);
  const detections = ref<DetectedEntity[]>([]);

  let worker: Worker | null = null;
  let stream: MediaStream | null = null;
  let scanning = false;
  let currentDeviceIndex = 0;
  let videoDevices: MediaDeviceInfo[] = [];
  let detectRequestId = 0;
  let detectInFlight = false;
  let lastDetectTime = 0;

  const entityCache = new Map<string, CachedEntity>();
  const pendingRequests = new Map<string, Promise<EntityData | null>>();
  const trackedEntities = new Map<string, DetectedEntity>();

  const visibility = useDocumentVisibility();
  let wasScanning = false;

  watch(visibility, vis => {
    if (vis === "hidden" && scanning) {
      wasScanning = true;
      pauseDetection();
    } else if (vis === "visible" && wasScanning) {
      wasScanning = false;
      resumeDetection();
    }
  });

  function initWorker(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        worker = new Worker(new URL("~~/workers/barcode-detector.ts", import.meta.url), { type: "module" });

        worker.onmessage = (event: MessageEvent<WorkerResponse>) => {
          const msg = event.data;

          if (msg.type === "ready") {
            isSupported.value = true;
            console.debug("[AR] Worker ready, BarcodeDetector initialized");
            resolve();
          } else if (msg.type === "error") {
            console.error("[AR] Worker init error:", msg.message);
            isSupported.value = false;
            reject(new Error(msg.message));
          } else if (msg.type === "result") {
            handleDetectionResult(msg.barcodes);
          }
        };

        worker.onerror = e => {
          console.error("[AR] Worker error:", e);
          isSupported.value = false;
          reject(e);
        };
      } catch (e) {
        console.error("[AR] Failed to create worker:", e);
        reject(e);
      }
    });
  }

  function handleDetectionResult(
    barcodes: Array<{
      rawValue: string;
      format: string;
      boundingBox: { x: number; y: number; width: number; height: number };
      cornerPoints: Array<{ x: number; y: number }>;
    }>
  ) {
    detectInFlight = false;

    const video = videoRef.value;
    if (!video || !scanning) return;

    const vt = getVideoTransform(video.videoWidth, video.videoHeight, video.clientWidth, video.clientHeight);
    const now = Date.now();

    for (const barcode of barcodes.slice(0, MAX_DETECTIONS)) {
      const parsed = parseHomeboxUrl(barcode.rawValue);
      if (!parsed) continue;

      const key = `${parsed.entityType}:${parsed.id}`;
      const mappedBox = mapBox(barcode.boundingBox, vt);
      const mappedCorners = mapCorners(barcode.cornerPoints, vt);
      const pose = computePose(mappedCorners);
      const existing = trackedEntities.get(key);

      if (existing) {
        existing.boundingBox = lerpRect(existing.boundingBox, mappedBox, SMOOTH);
        existing.cornerPoints = lerpCorners(existing.cornerPoints, mappedCorners, SMOOTH);
        existing.pose = {
          centerX: lerpVal(existing.pose.centerX, pose.centerX, SMOOTH),
          centerY: lerpVal(existing.pose.centerY, pose.centerY, SMOOTH),
          rotateZ: lerpVal(existing.pose.rotateZ, pose.rotateZ, SMOOTH),
          rotateX: lerpVal(existing.pose.rotateX, pose.rotateX, SMOOTH),
          rotateY: lerpVal(existing.pose.rotateY, pose.rotateY, SMOOTH),
          scale: lerpVal(existing.pose.scale, pose.scale, SMOOTH),
        };
        existing.lastSeen = now;
      } else {
        console.debug("[AR] New entity detected:", key);
        const entity: DetectedEntity = {
          id: parsed.id,
          entityType: parsed.entityType,
          rawValue: barcode.rawValue,
          boundingBox: mappedBox,
          cornerPoints: mappedCorners,
          pose,
          data: null,
          loading: true,
          error: false,
          lastSeen: now,
        };
        trackedEntities.set(key, entity);

        fetchEntityData(parsed.entityType, parsed.id).then(data => {
          const tracked = trackedEntities.get(key);
          if (tracked) {
            tracked.data = data;
            tracked.loading = false;
            tracked.error = data === null;
            console.debug(
              "[AR] Data fetched for",
              key,
              data
                ? `item: ${data.item?.name ?? "-"}, location: ${data.location?.name ?? "-"}, children: ${data.childItems?.length ?? 0}`
                : "not found"
            );
            updateDetections();
          }
        });
      }
    }

    // Remove stale entries
    for (const [key, entity] of trackedEntities) {
      if (now - entity.lastSeen > STALE_TIMEOUT) {
        trackedEntities.delete(key);
      }
    }

    updateDetections();
  }

  async function fetchEntityData(entityType: "item" | "location" | "asset", id: string): Promise<EntityData | null> {
    const cacheKey = `${entityType}:${id}`;

    const cached = entityCache.get(cacheKey);
    if (cached && Date.now() - cached.fetchedAt < CACHE_TTL) {
      return cached.data;
    }

    const pending = pendingRequests.get(cacheKey);
    if (pending) return pending;

    const api = useUserApi();
    const fetcher = async (): Promise<EntityData | null> => {
      try {
        if (entityType === "asset") {
          const { data } = await api.assets.get(id);
          if (data && data.items.length > 0) {
            if (data.items.length === 1) {
              const [itemRes, childRes] = await Promise.all([
                api.items.get(data.items[0]!.id),
                api.items.getAll({ parentIds: [data.items[0]!.id] }),
              ]);
              const result: EntityData = { item: itemRes.data ?? undefined, childItems: childRes.data?.items ?? [] };
              entityCache.set(cacheKey, { data: result, fetchedAt: Date.now() });
              return result;
            }
            const result: EntityData = { childItems: data.items };
            entityCache.set(cacheKey, { data: result, fetchedAt: Date.now() });
            return result;
          }
          return null;
        }

        if (entityType === "item") {
          const [itemRes, childRes] = await Promise.all([api.items.get(id), api.items.getAll({ parentIds: [id] })]);
          if (itemRes.data) {
            const result: EntityData = { item: itemRes.data, childItems: childRes.data?.items ?? [] };
            entityCache.set(cacheKey, { data: result, fetchedAt: Date.now() });
            return result;
          }
          return null;
        }

        const [locRes, itemsRes] = await Promise.all([api.locations.get(id), api.items.getAll({ locations: [id] })]);
        if (locRes.data) {
          const result: EntityData = { location: locRes.data, childItems: itemsRes.data?.items ?? [] };
          entityCache.set(cacheKey, { data: result, fetchedAt: Date.now() });
          return result;
        }
        return null;
      } catch {
        return null;
      } finally {
        pendingRequests.delete(cacheKey);
      }
    };
    const request = fetcher();

    pendingRequests.set(cacheKey, request);
    return request;
  }

  async function startCamera() {
    error.value = null;
    console.debug("[AR] startCamera called");

    if (!navigator?.mediaDevices?.getUserMedia) {
      error.value = "scanner.unsupported";
      return;
    }

    try {
      stream = await navigator.mediaDevices.getUserMedia({
        video: {
          facingMode: { ideal: "environment" },
          width: { ideal: 1920 },
          height: { ideal: 1080 },
        },
      });
      console.debug("[AR] Camera stream acquired");

      const devices = await navigator.mediaDevices.enumerateDevices();
      videoDevices = devices.filter(d => d.kind === "videoinput");

      const currentTrack = stream.getVideoTracks()[0];
      if (currentTrack) {
        const settings = currentTrack.getSettings();
        currentDeviceIndex = videoDevices.findIndex(d => d.deviceId === settings.deviceId);
        if (currentDeviceIndex === -1) currentDeviceIndex = 0;
      }

      if (videoRef.value) {
        videoRef.value.srcObject = stream;
        videoRef.value.setAttribute("playsinline", "true");
        await videoRef.value.play();
        console.debug("[AR] Video playing:", videoRef.value.videoWidth, "x", videoRef.value.videoHeight);
      }

      await initWorker();

      if (isSupported.value) {
        startDetection();
      } else {
        error.value = "scanner_ar.unsupported";
      }
    } catch (err: unknown) {
      console.error("[AR] startCamera error:", err);
      if (err instanceof Error && err.name === "NotAllowedError") {
        error.value = "scanner.permission_denied";
      } else {
        error.value = "scanner.error";
      }
    }
  }

  function stopCamera() {
    stopDetection();
    if (worker) {
      worker.terminate();
      worker = null;
    }
    if (stream) {
      stream.getTracks().forEach(t => t.stop());
      stream = null;
    }
    if (videoRef.value) {
      videoRef.value.srcObject = null;
    }
    trackedEntities.clear();
    detections.value = [];
  }

  async function switchCamera() {
    if (videoDevices.length < 2) return;

    currentDeviceIndex = (currentDeviceIndex + 1) % videoDevices.length;
    const nextDevice = videoDevices[currentDeviceIndex];
    if (!nextDevice) return;

    stopDetection();
    if (stream) {
      stream.getTracks().forEach(t => t.stop());
    }

    try {
      stream = await navigator.mediaDevices.getUserMedia({
        video: { deviceId: { exact: nextDevice.deviceId } },
      });

      if (videoRef.value) {
        videoRef.value.srcObject = stream;
        await videoRef.value.play();
      }

      startDetection();
    } catch {
      error.value = "scanner.error";
    }
  }

  function startDetection() {
    scanning = true;
    isScanning.value = true;
    detectInFlight = false;
    lastDetectTime = 0;
    renderLoop();
  }

  function stopDetection() {
    scanning = false;
    isScanning.value = false;
  }

  function pauseDetection() {
    scanning = false;
    isScanning.value = false;
  }

  function resumeDetection() {
    if (!stream || !worker) return;
    scanning = true;
    isScanning.value = true;
    renderLoop();
  }

  let frameCount = 0;
  let lastLogTime = 0;

  function renderLoop() {
    if (!scanning || !videoRef.value) return;

    const video = videoRef.value;
    const now = Date.now();

    // Log stats periodically
    frameCount++;
    if (now - lastLogTime > 3000) {
      console.debug(
        `[AR] Stats: ${frameCount} frames, ${trackedEntities.size} tracked, detect in-flight: ${detectInFlight}`
      );
      lastLogTime = now;
      frameCount = 0;
    }

    // Send frame to worker for detection at throttled rate
    if (!detectInFlight && worker && video.readyState >= 2 && now - lastDetectTime >= DETECT_INTERVAL) {
      try {
        const frame = createImageBitmap(video);
        frame.then(bitmap => {
          if (!worker || !scanning) {
            bitmap.close();
            return;
          }
          detectInFlight = true;
          lastDetectTime = now;
          detectRequestId++;
          worker.postMessage({ type: "detect", frame: bitmap, id: detectRequestId }, [bitmap]);
        });
      } catch {
        // createImageBitmap can fail if video not ready
      }
    }

    // Remove stale entries every frame (so cards disappear promptly)
    let removedStale = false;
    for (const [key, entity] of trackedEntities) {
      if (now - entity.lastSeen > STALE_TIMEOUT) {
        trackedEntities.delete(key);
        removedStale = true;
      }
    }
    if (removedStale) {
      updateDetections();
    }

    if (scanning) {
      requestAnimationFrame(renderLoop);
    }
  }

  function updateDetections() {
    detections.value = Array.from(trackedEntities.values()).map(e => ({
      id: e.id,
      entityType: e.entityType,
      rawValue: e.rawValue,
      boundingBox: new DOMRect(e.boundingBox.x, e.boundingBox.y, e.boundingBox.width, e.boundingBox.height),
      cornerPoints: e.cornerPoints.map(p => ({ ...p })),
      pose: { ...e.pose },
      data: e.data,
      loading: e.loading,
      error: e.error,
      lastSeen: e.lastSeen,
    }));
  }

  onUnmounted(() => {
    stopCamera();
  });

  return {
    isSupported,
    isScanning,
    error,
    detections,
    startCamera,
    stopCamera,
    switchCamera,
  };
}
