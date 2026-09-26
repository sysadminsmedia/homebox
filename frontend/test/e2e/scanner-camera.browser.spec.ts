import type { Page } from "@playwright/test";
import { expect, test } from "@playwright/test";
import { BarcodeFormat, EncodeHintType, QRCodeWriter } from "@zxing/library";

const SCANNER_ERROR = "An error occurred while scanning";
const PERMISSION_DENIED = "Camera permission denied, please allow access to the camera in your browser settings";
const UNSUPPORTED = "Media Stream API is not supported without HTTPS";
const NO_SOURCES = "No video sources available";
const LAST_USED_DEVICE_ID_KEY = "homebox:lastUsedDeviceId";
const PREFERENCES_KEY = "homebox/preferences/location";

const CAMERAS = {
  frontId: "camera-front",
  backId: "camera-back",
  frontLabel: "Front Camera",
  backLabel: "Back Camera",
};

type CameraMode = "allow" | "deny" | "no-devices" | "missing-media";

type CameraAcquisition = {
  deviceId: string | null;
  liveBefore: number;
};

type CameraStub = {
  installed: string;
  installError: string;
  mode: CameraMode;
  busyRejections: number;
  maxLiveStreams: number;
  acquisitions: CameraAcquisition[];
  setMode: (mode: CameraMode) => void;
  setQr: (modules: number[][]) => void;
  clearQr: () => void;
  hideMediaDevices: () => void;
  liveStreamCount: () => number;
  liveTrackCount: () => number;
};

declare global {
  interface Window {
    __hbCameraStub?: CameraStub;
  }
}

type StubState = {
  liveStreams: number;
  liveTracks: number;
  busyRejections: number;
  maxLiveStreams: number;
  acquisitions: CameraAcquisition[];
};

function qrModules(text: string): number[][] {
  const hints = new Map<EncodeHintType, string | number>([
    [EncodeHintType.ERROR_CORRECTION, "H"],
    [EncodeHintType.MARGIN, 4],
  ]);
  const matrix = new QRCodeWriter().encode(text, BarcodeFormat.QR_CODE, 0, 0, hints);
  const rows: number[][] = [];
  for (let y = 0; y < matrix.getHeight(); y++) {
    const row: number[] = [];
    for (let x = 0; x < matrix.getWidth(); x++) row.push(matrix.get(x, y) ? 1 : 0);
    rows.push(row);
  }
  return rows;
}

// Page-init stub. A canvas captureStream stands in for the camera, and a second
// getUserMedia is rejected while any earlier track is still live.
function installCameraStub(config: typeof CAMERAS) {
  const outputs: Array<{ ctx: CanvasRenderingContext2D; stream: MediaStream }> = [];
  let hidden = false;
  let mode: CameraMode = "allow";
  let qr: number[][] | null = null;
  let maxLiveStreams = 0;
  let busyRejections = 0;
  let loopStarted = false;
  const acquisitions: CameraAcquisition[] = [];
  let installed = "failed";
  let installError = "";

  function paint(ctx: CanvasRenderingContext2D) {
    const width = ctx.canvas.width;
    const height = ctx.canvas.height;
    ctx.fillStyle = "#ffffff";
    ctx.fillRect(0, 0, width, height);
    const modules = qr;
    if (modules && modules.length > 0) {
      const size = modules.length;
      const scale = Math.max(1, Math.floor(Math.min(width, height) / size));
      const drawn = scale * size;
      const ox = Math.floor((width - drawn) / 2);
      const oy = Math.floor((height - drawn) / 2);
      ctx.fillStyle = "#000000";
      for (let y = 0; y < size; y++) {
        const row = modules[y];
        if (!row) continue;
        for (let x = 0; x < row.length; x++) {
          if (row[x]) ctx.fillRect(ox + x * scale, oy + y * scale, scale, scale);
        }
      }
      if (ox > 2) {
        ctx.fillStyle = Math.floor(performance.now() / 80) % 2 === 0 ? "#000000" : "#ffffff";
        ctx.fillRect(0, 0, 1, 1);
      }
      return;
    }
    ctx.fillStyle = Math.floor(performance.now() / 80) % 2 === 0 ? "#f3f3f3" : "#ffffff";
    ctx.fillRect(0, 0, width, height);
  }

  function paintLive() {
    for (const output of outputs) {
      const track = output.stream.getVideoTracks()[0] as (MediaStreamTrack & { requestFrame?: () => void }) | undefined;
      if (track?.readyState !== "live") continue;
      paint(output.ctx);
      track.requestFrame?.();
    }
  }

  function tick() {
    paintLive();
    requestAnimationFrame(tick);
  }

  function startLoop() {
    if (loopStarted) return;
    loopStarted = true;
    requestAnimationFrame(tick);
    setInterval(paintLive, 50);
  }

  function mountCanvas(): HTMLCanvasElement {
    const canvas = document.createElement("canvas");
    canvas.width = 480;
    canvas.height = 480;
    canvas.setAttribute("aria-hidden", "true");
    canvas.style.cssText =
      "position:fixed;top:0;left:0;width:480px;height:480px;opacity:0.01;pointer-events:none;z-index:-1;";
    (document.body || document.documentElement).appendChild(canvas);
    return canvas;
  }

  function createStream(): MediaStream {
    const canvas = mountCanvas();
    const ctx = canvas.getContext("2d", { alpha: false });
    if (!ctx) throw new Error("canvas 2d context unavailable");
    paint(ctx);
    const stream = canvas.captureStream(30);
    const track = stream.getVideoTracks()[0] as (MediaStreamTrack & { requestFrame?: () => void }) | undefined;
    track?.requestFrame?.();
    outputs.push({ ctx, stream });
    return stream;
  }

  // Real cameras are not instantaneous. Yield until the scanner dialog has
  // mounted its video so ZXing attaches the decode stream to that element.
  // A probe that runs before any dialog (the old header path) gives up quickly.
  async function waitForDialogVideo() {
    const started = performance.now();
    while (performance.now() - started < 1000) {
      if (document.querySelector("[role='dialog'] video")) return;
      // A probe that starts before any dialog (the old header path) must not wait forever.
      if (!document.querySelector("[role='dialog']") && performance.now() - started > 400) return;
      await new Promise(resolve => requestAnimationFrame(resolve));
    }
  }

  function liveStreamCount() {
    return outputs.filter(output => output.stream.getTracks().some(track => track.readyState === "live")).length;
  }

  function liveTrackCount() {
    return outputs.reduce(
      (total, output) => total + output.stream.getTracks().filter(track => track.readyState === "live").length,
      0
    );
  }

  function noteLive() {
    const live = liveStreamCount();
    if (live > maxLiveStreams) maxLiveStreams = live;
    return live;
  }

  function deviceIdFrom(constraints: MediaStreamConstraints | undefined) {
    const video = constraints?.video;
    if (!video || video === true) return null;
    const deviceId = video.deviceId;
    if (typeof deviceId === "string") return deviceId;
    if (deviceId && typeof deviceId === "object" && "exact" in deviceId && typeof deviceId.exact === "string") {
      return deviceId.exact;
    }
    return null;
  }

  function videoDevice(deviceId: string, label: string) {
    return {
      deviceId,
      groupId: `${deviceId}-group`,
      kind: "videoinput" as const,
      label,
      toJSON() {
        return { deviceId: this.deviceId, groupId: this.groupId, kind: this.kind, label: this.label };
      },
    };
  }

  async function getUserMedia(constraints?: MediaStreamConstraints): Promise<MediaStream> {
    if (mode === "deny") throw new DOMException("Permission denied", "NotAllowedError");
    startLoop();
    await waitForDialogVideo();
    const liveBefore = noteLive();
    if (liveBefore > 0) {
      busyRejections += 1;
      throw new DOMException("Could not start video source", "NotReadableError");
    }
    const stream = createStream();
    acquisitions.push({ deviceId: deviceIdFrom(constraints), liveBefore });
    noteLive();
    return stream;
  }

  async function enumerateDevices() {
    if (mode === "no-devices") {
      return [
        {
          deviceId: "microphone",
          groupId: "microphone-group",
          kind: "audioinput" as const,
          label: "Microphone",
          toJSON() {
            return { deviceId: this.deviceId, groupId: this.groupId, kind: this.kind, label: this.label };
          },
        },
      ];
    }
    return [videoDevice(config.frontId, config.frontLabel), videoDevice(config.backId, config.backLabel)];
  }

  const fake = {
    getUserMedia,
    enumerateDevices,
    getSupportedConstraints: () => ({ deviceId: true, facingMode: true, width: true, height: true }),
    addEventListener: () => undefined,
    removeEventListener: () => undefined,
    dispatchEvent: () => true,
    ondevicechange: null,
  };

  try {
    Object.defineProperty(navigator, "mediaDevices", {
      configurable: true,
      enumerable: true,
      get() {
        return hidden ? undefined : fake;
      },
    });
    installed = "instance";
  } catch (error) {
    installError = String(error);
    const media = navigator.mediaDevices;
    if (media) {
      media.getUserMedia = getUserMedia;
      media.enumerateDevices = enumerateDevices;
      installed = "patch";
    }
  }

  window.__hbCameraStub = {
    installed,
    installError,
    get mode() {
      return mode;
    },
    set mode(next: CameraMode) {
      mode = next;
    },
    get busyRejections() {
      return busyRejections;
    },
    get maxLiveStreams() {
      return maxLiveStreams;
    },
    get acquisitions() {
      return acquisitions;
    },
    setMode(next) {
      mode = next;
    },
    setQr(modules) {
      qr = modules;
    },
    clearQr() {
      qr = null;
    },
    hideMediaDevices() {
      hidden = true;
      if (installed === "patch") {
        Object.defineProperty(navigator, "mediaDevices", {
          configurable: true,
          get() {
            return undefined;
          },
        });
      }
    },
    liveStreamCount,
    liveTrackCount,
  };

  startLoop();
}

async function login(page: Page) {
  await page.goto("/home");
  await expect(page).toHaveURL("/");
  await page.fill("input[type='text']", "demo@example.com");
  await page.fill("input[type='password']", "demodemo");
  await page.click("button[type='submit']");
  await expect(page).toHaveURL("/home");
}

function scannerDialog(page: Page) {
  return page.getByRole("dialog").filter({ has: page.getByRole("heading", { name: "Scanner", exact: true }) });
}

async function stubState(page: Page): Promise<StubState> {
  return page.evaluate(() => {
    const stub = window.__hbCameraStub;
    if (!stub) throw new Error("camera stub missing");
    return {
      liveStreams: stub.liveStreamCount(),
      liveTracks: stub.liveTrackCount(),
      busyRejections: stub.busyRejections,
      maxLiveStreams: stub.maxLiveStreams,
      acquisitions: stub.acquisitions.map(item => ({ deviceId: item.deviceId, liveBefore: item.liveBefore })),
    };
  });
}

async function openScannerButton(page: Page) {
  await page.getByRole("button", { name: "Scanner", exact: true }).click();
  await expect(scannerDialog(page)).toBeVisible();
}

async function closeScanner(page: Page) {
  await scannerDialog(page).getByRole("button", { name: "Close", exact: true }).click();
  await expect(scannerDialog(page)).toBeHidden();
}

async function expectSingleLivePreview(page: Page) {
  const video = scannerDialog(page).locator("video");
  await expect
    .poll(
      async () => {
        const attached = (await video.count()) > 0;
        const playing = attached
          ? await video.evaluate((el: HTMLVideoElement) => {
              return (
                el.srcObject instanceof MediaStream &&
                el.readyState >= 2 &&
                !el.paused &&
                !el.ended &&
                el.videoWidth > 0 &&
                el.currentTime > 0
              );
            })
          : false;
        const errorVisible = await scannerDialog(page).getByText(SCANNER_ERROR).isVisible();
        const state = await stubState(page);
        return {
          playing,
          errorVisible,
          liveStreams: state.liveStreams,
          busyRejections: state.busyRejections,
          maxLiveStreams: state.maxLiveStreams,
        };
      },
      { timeout: 15_000 }
    )
    .toEqual({
      playing: true,
      errorVisible: false,
      liveStreams: 1,
      busyRejections: 0,
      maxLiveStreams: 1,
    });
}

async function expectMessage(page: Page, message: string) {
  await expect(scannerDialog(page).getByRole("alert")).toContainText(message);
  await expect(page.getByText(message)).toHaveCount(1);
  await expect.poll(async () => (await stubState(page)).liveTracks).toBe(0);
}

test.describe("scanner camera", () => {
  test.describe.configure({ timeout: 120_000 });
  test.use({ viewport: { width: 1280, height: 800 } });

  test.beforeEach(async ({ page }) => {
    await page.addInitScript(installCameraStub, CAMERAS);
    await login(page);
    const installed = await page.evaluate(() => window.__hbCameraStub?.installed ?? "missing");
    expect(installed, "mediaDevices stub did not install").not.toBe("failed");
  });

  test("header scan plays one stream, releases it, and reopens cleanly", async ({ page }) => {
    await openScannerButton(page);
    await expect(scannerDialog(page).getByRole("combobox")).toContainText(CAMERAS.backLabel);
    await expectSingleLivePreview(page);

    const opened = await stubState(page);
    expect(opened.acquisitions.map(item => item.deviceId)).toEqual([null, CAMERAS.backId]);
    expect(opened.acquisitions.every(item => item.liveBefore === 0)).toBe(true);

    await closeScanner(page);
    await expect.poll(async () => (await stubState(page)).liveTracks).toBe(0);

    await openScannerButton(page);
    await expectSingleLivePreview(page);
    const reopened = await stubState(page);
    expect(reopened.liveStreams).toBe(1);
    expect(reopened.maxLiveStreams).toBe(1);
    expect(reopened.acquisitions.at(-1)?.liveBefore).toBe(0);
  });

  test("switching cameras stops the previous stream before the next acquisition", async ({ page }) => {
    await openScannerButton(page);
    await expectSingleLivePreview(page);

    await scannerDialog(page).getByRole("combobox").click();
    await page.getByRole("option", { name: CAMERAS.frontLabel, exact: true }).click();

    await expect(scannerDialog(page).getByRole("combobox")).toContainText(CAMERAS.frontLabel);
    await expectSingleLivePreview(page);

    const state = await stubState(page);
    expect(state.acquisitions.at(-1)).toEqual({ deviceId: CAMERAS.frontId, liveBefore: 0 });
    expect(state.busyRejections).toBe(0);
    expect(state.liveStreams).toBe(1);
  });

  test("permission denial stays inside the scanner dialog", async ({ page }) => {
    await page.evaluate(() => window.__hbCameraStub?.setMode("deny"));
    await openScannerButton(page);
    await expectMessage(page, PERMISSION_DENIED);
    await expect(page.locator("[data-sonner-toast]")).toHaveCount(0);
  });

  test("missing mediaDevices shows the unsupported message", async ({ page }) => {
    await page.evaluate(() => window.__hbCameraStub?.hideMediaDevices());
    await openScannerButton(page);
    await expectMessage(page, UNSUPPORTED);
  });

  test("an empty video device list shows the no-sources message", async ({ page }) => {
    await page.evaluate(() => window.__hbCameraStub?.setMode("no-devices"));
    await openScannerButton(page);
    await expectMessage(page, NO_SOURCES);
  });

  test("sidebar scanner entry uses the same single-owner path", async ({ page }) => {
    await page.evaluate(key => {
      const raw = localStorage.getItem(key);
      const current = raw ? (JSON.parse(raw) as Record<string, unknown>) : {};
      current.displayLegacyHeader = true;
      localStorage.setItem(key, JSON.stringify(current));
    }, PREFERENCES_KEY);
    await page.reload();
    await expect(page).toHaveURL(/\/home$/);
    await expect(page.getByRole("button", { name: "Scanner", exact: true })).toBeVisible();
    await openScannerButton(page);
    await expectSingleLivePreview(page);
    await closeScanner(page);
    await expect.poll(async () => (await stubState(page)).liveTracks).toBe(0);
  });

  test("quick menu scanner entry uses the same single-owner path", async ({ page }) => {
    await page.getByRole("heading", { name: "Quick Statistics" }).click();
    await page.keyboard.down("Control");
    await page.keyboard.press("Backquote");
    await page.keyboard.up("Control");
    await expect(page.getByPlaceholder("Use the number keys to quickly select an action.")).toBeVisible();
    await page.getByRole("option", { name: "Scanner", exact: true }).click();
    await expect(scannerDialog(page)).toBeVisible();
    await expectSingleLivePreview(page);
  });

  test("restores the saved camera and writes the next choice", async ({ page }) => {
    await page.evaluate(({ key, deviceId }) => localStorage.setItem(key, deviceId), {
      key: LAST_USED_DEVICE_ID_KEY,
      deviceId: CAMERAS.frontId,
    });
    await openScannerButton(page);
    await expect(scannerDialog(page).getByRole("combobox")).toContainText(CAMERAS.frontLabel);
    await expectSingleLivePreview(page);
    expect((await stubState(page)).acquisitions.at(-1)?.deviceId).toBe(CAMERAS.frontId);

    await scannerDialog(page).getByRole("combobox").click();
    await page.getByRole("option", { name: CAMERAS.backLabel, exact: true }).click();
    await expect(scannerDialog(page).getByRole("combobox")).toContainText(CAMERAS.backLabel);
    await expectSingleLivePreview(page);

    await closeScanner(page);
    await openScannerButton(page);
    await expect(scannerDialog(page).getByRole("combobox")).toContainText(CAMERAS.backLabel);
    await expectSingleLivePreview(page);
    const state = await stubState(page);
    expect(state.acquisitions.at(-1)).toEqual({ deviceId: CAMERAS.backId, liveBefore: 0 });
    expect(state.maxLiveStreams).toBe(1);
  });

  test("a QR payload navigates and releases the camera", async ({ page }) => {
    const target = `${new URL(page.url()).origin}/items`;
    await page.evaluate(modules => window.__hbCameraStub?.setQr(modules), qrModules(target));
    await page.getByRole("button", { name: "Scanner", exact: true }).click();
    await expect.poll(() => page.evaluate(() => location.pathname), { timeout: 20_000 }).toBe("/items");
    await expect(scannerDialog(page)).toBeHidden();
    await expect.poll(async () => (await stubState(page)).liveTracks).toBe(0);
    await expect.poll(async () => (await stubState(page)).busyRejections).toBe(0);
  });
});
