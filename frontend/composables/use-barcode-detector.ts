import { ref, type Ref, watch, onUnmounted } from "vue";
import { useDocumentVisibility } from "@vueuse/core";
import { BarcodeDetector as BarcodeDetectorPolyfill } from "barcode-detector";
import type { ItemOut, ItemSummary, LocationOut } from "~~/lib/api/types/data-contracts";

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
  /** Center of the QR code in viewport coordinates */
  centerX: number;
  centerY: number;
  /** In-plane rotation in degrees (Z-axis) */
  rotateZ: number;
  /** Forward/backward tilt in degrees (X-axis) — positive = top tilted away */
  rotateX: number;
  /** Left/right tilt in degrees (Y-axis) — positive = right side closer */
  rotateY: number;
  /** Scale factor based on apparent QR code size (1.0 = reference size of ~200px) */
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
/** Smoothing factor: 0 = no smoothing (snap), 1 = frozen. 0.6 feels responsive but stable. */
const SMOOTH = 0.6;

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
        console.debug("[AR] QR value is not a URL or path, ignoring:", rawValue);
        return null;
      }
    }

    const sanitized = pathname.replace(/[^a-zA-Z0-9-_/]/g, "");

    // /a/{asset-id} - the primary QR code format
    const assetMatch = sanitized.match(/^\/a\/([a-zA-Z0-9-_]+)/);
    if (assetMatch) {
      return { entityType: "asset", id: assetMatch[1]! };
    }

    const itemMatch = sanitized.match(/^\/item\/([a-zA-Z0-9-_]+)/);
    if (itemMatch) {
      return { entityType: "item", id: itemMatch[1]! };
    }

    const locationMatch = sanitized.match(/^\/location\/([a-zA-Z0-9-_]+)/);
    if (locationMatch) {
      return { entityType: "location", id: locationMatch[1]! };
    }
    return null;
  } catch {
    return null;
  }
}

function mapBoundingBox(bbox: DOMRectReadOnly, video: HTMLVideoElement): DOMRect {
  const videoW = video.videoWidth;
  const videoH = video.videoHeight;
  const clientW = video.clientWidth;
  const clientH = video.clientHeight;

  if (!videoW || !videoH || !clientW || !clientH) {
    return new DOMRect(bbox.x, bbox.y, bbox.width, bbox.height);
  }

  const videoAspect = videoW / videoH;
  const clientAspect = clientW / clientH;

  let scaleX: number, scaleY: number, offsetX: number, offsetY: number;

  if (videoAspect > clientAspect) {
    // Video is wider than container - cropped horizontally (object-fit: cover)
    scaleY = clientH / videoH;
    scaleX = scaleY;
    offsetX = (clientW - videoW * scaleX) / 2;
    offsetY = 0;
  } else {
    // Video is taller than container - cropped vertically
    scaleX = clientW / videoW;
    scaleY = scaleX;
    offsetX = 0;
    offsetY = (clientH - videoH * scaleY) / 2;
  }

  return new DOMRect(bbox.x * scaleX + offsetX, bbox.y * scaleY + offsetY, bbox.width * scaleX, bbox.height * scaleY);
}

function getVideoTransform(video: HTMLVideoElement) {
  const videoW = video.videoWidth;
  const videoH = video.videoHeight;
  const clientW = video.clientWidth;
  const clientH = video.clientHeight;

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

function mapCornerPoints(corners: Array<{ x: number; y: number }>, video: HTMLVideoElement): Point2D[] {
  const { scaleX, scaleY, offsetX, offsetY } = getVideoTransform(video);
  return corners.map(c => ({
    x: c.x * scaleX + offsetX,
    y: c.y * scaleY + offsetY,
  }));
}

function dist(a: Point2D, b: Point2D): number {
  return Math.sqrt((a.x - b.x) ** 2 + (a.y - b.y) ** 2);
}

const REFERENCE_QR_SIZE = 200;

/**
 * Compute 3D pose from 4 corner points of a detected QR code.
 * Corner order: [topLeft, topRight, bottomRight, bottomLeft]
 */
function computePose(corners: Point2D[]): Pose3D {
  if (corners.length < 4) {
    return { centerX: 0, centerY: 0, rotateZ: 0, rotateX: 0, rotateY: 0, scale: 1 };
  }

  const [tl, tr, br, bl] = corners as [Point2D, Point2D, Point2D, Point2D];

  // Center
  const centerX = (tl.x + tr.x + br.x + bl.x) / 4;
  const centerY = (tl.y + tr.y + br.y + bl.y) / 4;

  // In-plane rotation: angle of the top edge relative to horizontal
  const rotateZ = Math.atan2(tr.y - tl.y, tr.x - tl.x) * (180 / Math.PI);

  // Perspective tilt (X-axis rotation: top tilted away from camera)
  // When QR is tilted back, the top edge appears shorter than the bottom edge
  const topEdge = dist(tl, tr);
  const bottomEdge = dist(bl, br);
  // Ratio > 1 means bottom is wider (top tilted away)
  const xRatio = bottomEdge / (topEdge || 1);
  // Convert to approximate angle — clamp to avoid extreme values
  const rotateX = Math.max(-60, Math.min(60, (xRatio - 1) * 50));

  // Perspective tilt (Y-axis rotation: right side closer to camera)
  // When QR is rotated right, the right edge appears taller than the left edge
  const leftEdge = dist(tl, bl);
  const rightEdge = dist(tr, br);
  const yRatio = rightEdge / (leftEdge || 1);
  const rotateY = Math.max(-60, Math.min(60, (1 - yRatio) * 50));

  // Scale based on average edge length relative to reference size
  const avgEdge = (topEdge + bottomEdge + leftEdge + rightEdge) / 4;
  const scale = Math.max(0.3, Math.min(2.0, avgEdge / REFERENCE_QR_SIZE));

  return { centerX, centerY, rotateZ, rotateX, rotateY, scale };
}

/**
 * Solve a 2D perspective transform (homography) from 4 point correspondences.
 * Returns a 3x3 matrix H (row-major, 9 elements) such that dst ≈ H * src in homogeneous coords.
 */
export function solveHomography(
  src: [Point2D, Point2D, Point2D, Point2D],
  dst: [Point2D, Point2D, Point2D, Point2D]
): number[] {
  // 8x8 linear system: for each (x,y)→(u,v):
  //   [x y 1 0 0 0 -xu -yu] [a]   [u]
  //   [0 0 0 x y 1 -xv -yv] [b] = [v]
  const aug: number[][] = [];
  for (let i = 0; i < 4; i++) {
    const { x, y } = src[i]!;
    const { x: u, y: v } = dst[i]!;
    aug.push([x, y, 1, 0, 0, 0, -x * u, -y * u, u]);
    aug.push([0, 0, 0, x, y, 1, -x * v, -y * v, v]);
  }

  const n = 8;
  // Gaussian elimination with partial pivoting
  for (let col = 0; col < n; col++) {
    let maxRow = col;
    for (let row = col + 1; row < n; row++) {
      if (Math.abs(aug[row]![col]!) > Math.abs(aug[maxRow]![col]!)) maxRow = row;
    }
    [aug[col], aug[maxRow]] = [aug[maxRow]!, aug[col]!];
    const pivot = aug[col]![col]!;
    if (Math.abs(pivot) < 1e-10) return [1, 0, 0, 0, 1, 0, 0, 0, 1]; // degenerate, return identity
    for (let row = col + 1; row < n; row++) {
      const factor = aug[row]![col]! / pivot;
      for (let j = col; j <= n; j++) {
        aug[row]![j]! -= factor * aug[col]![j]!;
      }
    }
  }

  // Back substitution
  const h = new Array(n).fill(0);
  for (let i = n - 1; i >= 0; i--) {
    h[i] = aug[i]![n]!;
    for (let j = i + 1; j < n; j++) {
      h[i] -= aug[i]![j]! * h[j];
    }
    h[i] /= aug[i]![i]!;
  }

  return [...h, 1]; // [a,b,c,d,e,f,g,h,1]
}

/**
 * Convert a 3x3 row-major homography matrix to a CSS matrix3d() string.
 * Embeds the 2D projective transform into 3D so CSS perspective division applies.
 */
export function homographyToMatrix3d(H: number[]): string {
  // H = [h0 h1 h2; h3 h4 h5; h6 h7 h8] row-major
  // CSS matrix3d is column-major 4x4:
  // [h0 h1 0 h2]     matrix3d(h0, h3, 0, h6,
  // [h3 h4 0 h5]  →           h1, h4, 0, h7,
  // [0  0  1 0 ]               0,  0, 1,  0,
  // [h6 h7 0 h8]              h2, h5, 0, h8)
  return `matrix3d(${H[0]}, ${H[3]}, 0, ${H[6]}, ${H[1]}, ${H[4]}, 0, ${H[7]}, 0, 0, 1, 0, ${H[2]}, ${H[5]}, 0, ${H[8]})`;
}

export function useBarcodeDetector(videoRef: Ref<HTMLVideoElement | undefined>) {
  const isSupported = ref(false);
  const isScanning = ref(false);
  const error = ref<string | null>(null);
  const detections = ref<DetectedEntity[]>([]);

  let detector: BarcodeDetectorPolyfill | null = null;
  let stream: MediaStream | null = null;
  let scanning = false;
  let currentDeviceIndex = 0;
  let videoDevices: MediaDeviceInfo[] = [];

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

  function init() {
    try {
      detector = new BarcodeDetectorPolyfill({ formats: ["qr_code"] });
      isSupported.value = true;
      console.debug("[AR] BarcodeDetector initialized successfully");
    } catch (e) {
      isSupported.value = false;
      console.error("[AR] BarcodeDetector init failed:", e);
    }
  }

  async function fetchEntityData(entityType: "item" | "location" | "asset", id: string): Promise<EntityData | null> {
    const cacheKey = `${entityType}:${id}`;

    const cached = entityCache.get(cacheKey);
    if (cached && Date.now() - cached.fetchedAt < CACHE_TTL) {
      return cached.data;
    }

    const pending = pendingRequests.get(cacheKey);
    if (pending) {
      return pending;
    }

    const api = useUserApi();
    const request = (async (): Promise<EntityData | null> => {
      try {
        if (entityType === "asset") {
          const { data } = await api.assets.get(id);
          if (data && data.items.length > 0) {
            if (data.items.length === 1) {
              // Single item - fetch full details + child items
              const [itemRes, childRes] = await Promise.all([
                api.items.get(data.items[0]!.id),
                api.items.getAll({ parentIds: [data.items[0]!.id] }),
              ]);
              const result: EntityData = {
                item: itemRes.data ?? undefined,
                childItems: childRes.data?.items ?? [],
              };
              entityCache.set(cacheKey, { data: result, fetchedAt: Date.now() });
              return result;
            }
            // Multiple items share this asset ID
            const result: EntityData = { childItems: data.items };
            entityCache.set(cacheKey, { data: result, fetchedAt: Date.now() });
            return result;
          }
          return null;
        }

        if (entityType === "item") {
          const [itemRes, childRes] = await Promise.all([api.items.get(id), api.items.getAll({ parentIds: [id] })]);
          if (itemRes.data) {
            const result: EntityData = {
              item: itemRes.data,
              childItems: childRes.data?.items ?? [],
            };
            entityCache.set(cacheKey, { data: result, fetchedAt: Date.now() });
            return result;
          }
          return null;
        }

        // Location - fetch location details + items in location
        const [locRes, itemsRes] = await Promise.all([api.locations.get(id), api.items.getAll({ locations: [id] })]);
        if (locRes.data) {
          const result: EntityData = {
            location: locRes.data,
            childItems: itemsRes.data?.items ?? [],
          };
          entityCache.set(cacheKey, { data: result, fetchedAt: Date.now() });
          return result;
        }
        return null;
      } catch {
        return null;
      } finally {
        pendingRequests.delete(cacheKey);
      }
    })();

    pendingRequests.set(cacheKey, request);
    return request;
  }

  async function startCamera() {
    error.value = null;
    console.debug("[AR] startCamera called");

    if (!navigator?.mediaDevices?.getUserMedia) {
      console.error("[AR] getUserMedia not available");
      error.value = "scanner.unsupported";
      return;
    }

    try {
      console.debug("[AR] Requesting camera stream...");
      stream = await navigator.mediaDevices.getUserMedia({
        video: {
          facingMode: { ideal: "environment" },
          width: { ideal: 1920 },
          height: { ideal: 1080 },
        },
      });
      console.debug(
        "[AR] Camera stream acquired, tracks:",
        stream.getTracks().map(t => ({ kind: t.kind, label: t.label, readyState: t.readyState }))
      );

      // Enumerate devices for camera switching
      const devices = await navigator.mediaDevices.enumerateDevices();
      videoDevices = devices.filter(d => d.kind === "videoinput");
      console.debug(
        "[AR] Video devices found:",
        videoDevices.length,
        videoDevices.map(d => d.label)
      );

      // Find current device index
      const currentTrack = stream.getVideoTracks()[0];
      if (currentTrack) {
        const settings = currentTrack.getSettings();
        currentDeviceIndex = videoDevices.findIndex(d => d.deviceId === settings.deviceId);
        if (currentDeviceIndex === -1) currentDeviceIndex = 0;
        console.debug("[AR] Current device index:", currentDeviceIndex, "settings:", {
          width: settings.width,
          height: settings.height,
          facingMode: settings.facingMode,
        });
      }

      if (videoRef.value) {
        videoRef.value.srcObject = stream;
        videoRef.value.setAttribute("playsinline", "true");
        await videoRef.value.play();
        console.debug(
          "[AR] Video playing, readyState:",
          videoRef.value.readyState,
          "videoWidth:",
          videoRef.value.videoWidth,
          "videoHeight:",
          videoRef.value.videoHeight
        );
      } else {
        console.error("[AR] videoRef.value is undefined, cannot attach stream");
      }

      init();

      if (isSupported.value) {
        console.debug("[AR] Starting detection loop");
        startDetection();
      } else {
        console.error("[AR] BarcodeDetector not supported");
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
    detectLoop();
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
    if (!stream || !detector) return;
    scanning = true;
    isScanning.value = true;
    detectLoop();
  }

  let frameCount = 0;
  let lastLogTime = 0;

  async function detectLoop() {
    if (!scanning || !detector || !videoRef.value) {
      console.debug(
        "[AR] detectLoop exiting early: scanning=",
        scanning,
        "detector=",
        !!detector,
        "videoRef=",
        !!videoRef.value
      );
      return;
    }

    const video = videoRef.value;
    if (video.readyState < 2) {
      console.debug("[AR] Video not ready yet, readyState:", video.readyState);
      requestAnimationFrame(detectLoop);
      return;
    }

    try {
      const barcodes = await detector.detect(video);
      frameCount++;
      const now = Date.now();

      // Log stats every 3 seconds
      if (now - lastLogTime > 3000) {
        console.debug(
          `[AR] Detection stats: ${frameCount} frames processed, ${barcodes.length} barcodes in latest frame, ${trackedEntities.size} tracked entities, video: ${video.videoWidth}x${video.videoHeight}, client: ${video.clientWidth}x${video.clientHeight}`
        );
        lastLogTime = now;
        frameCount = 0;
      }

      const seenIds = new Set<string>();

      for (const barcode of barcodes.slice(0, MAX_DETECTIONS)) {
        const parsed = parseHomeboxUrl(barcode.rawValue);
        if (!parsed) continue;

        const key = `${parsed.entityType}:${parsed.id}`;
        seenIds.add(key);

        const mappedBox = mapBoundingBox(barcode.boundingBox, video);
        const mappedCorners = mapCornerPoints(barcode.cornerPoints ?? [], video);
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
          console.debug(
            "[AR] New entity detected:",
            key,
            "pose:",
            `center=(${Math.round(pose.centerX)},${Math.round(pose.centerY)}) rotZ=${pose.rotateZ.toFixed(1)} rotX=${pose.rotateX.toFixed(1)} rotY=${pose.rotateY.toFixed(1)} scale=${pose.scale.toFixed(2)}`
          );
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

          // Fetch data in background
          console.debug("[AR] Fetching data for", key);
          fetchEntityData(parsed.entityType, parsed.id).then(data => {
            const tracked = trackedEntities.get(key);
            if (tracked) {
              tracked.data = data;
              tracked.loading = false;
              tracked.error = data === null;
              console.debug(
                "[AR] Data fetched for",
                key,
                "success:",
                data !== null,
                data
                  ? `item: ${data.item?.name ?? "-"}, location: ${data.location?.name ?? "-"}, children: ${data.childItems?.length ?? 0}`
                  : ""
              );
              updateDetections();
            }
          });
        }
      }

      // Remove stale entries
      for (const [key, entity] of trackedEntities) {
        if (now - entity.lastSeen > STALE_TIMEOUT) {
          console.debug("[AR] Removing stale entity:", key);
          trackedEntities.delete(key);
        }
      }

      updateDetections();
    } catch (e) {
      console.error("[AR] Detection error:", e);
    }

    if (scanning) {
      requestAnimationFrame(detectLoop);
    }
  }

  function updateDetections() {
    // Deep-copy tracked entities into plain objects so Vue reactivity picks up all changes
    const snapshot: DetectedEntity[] = Array.from(trackedEntities.values()).map(e => ({
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
    detections.value = snapshot;
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
