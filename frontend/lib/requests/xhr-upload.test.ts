import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { xhrUpload } from "./xhr-upload";

interface Listener {
  (e: unknown): void;
}

interface UploadStub {
  listeners: Record<string, Listener[]>;
  addEventListener(event: string, fn: Listener): void;
}

class FakeXHR {
  upload: UploadStub = {
    listeners: {},
    addEventListener(event, fn) {
      (this.listeners[event] ??= []).push(fn);
    },
  };
  listeners: Record<string, Listener[]> = {};
  headers: Record<string, string> = {};
  method = "";
  url = "";
  status = 0;
  statusText = "";
  responseText = "";
  private responseHeaders: Record<string, string> = {};
  sentBody: FormData | null = null;

  open(method: string, url: string) {
    this.method = method;
    this.url = url;
  }
  setRequestHeader(name: string, value: string) {
    this.headers[name] = value;
  }
  addEventListener(event: string, fn: Listener) {
    (this.listeners[event] ??= []).push(fn);
  }
  getResponseHeader(name: string) {
    return this.responseHeaders[name] ?? null;
  }
  send(body: FormData) {
    this.sentBody = body;
  }

  // Test helpers
  fireUploadProgress(loaded: number, total: number) {
    for (const fn of this.upload.listeners.progress ?? []) {
      fn({ lengthComputable: true, loaded, total });
    }
  }
  fireUploadLoad() {
    for (const fn of this.upload.listeners.load ?? []) fn({});
  }
  complete(status: number, statusText: string, body: string, contentType = "application/json") {
    this.status = status;
    this.statusText = statusText;
    this.responseText = body;
    this.responseHeaders["Content-Type"] = contentType;
    for (const fn of this.listeners.load ?? []) fn({});
  }
  fail(event: "error" | "abort" | "timeout") {
    for (const fn of this.listeners[event] ?? []) fn({});
  }
}

let lastXhr: FakeXHR;

beforeEach(() => {
  vi.stubGlobal(
    "XMLHttpRequest",
    vi.fn(() => {
      lastXhr = new FakeXHR();
      return lastXhr;
    })
  );
});

afterEach(() => {
  vi.unstubAllGlobals();
});

const fd = () => {
  const f = new FormData();
  f.append("file", "x");
  return f;
};

describe("xhrUpload", () => {
  test("sets method, url, headers, and sends the FormData body", async () => {
    const promise = xhrUpload({
      url: "/api/v1/items/123/attachments",
      data: fd(),
      headers: { Authorization: "Bearer abc", "X-Trace": "t1" },
    });
    lastXhr.complete(201, "Created", JSON.stringify({ id: "att1" }));
    const result = await promise;

    expect(lastXhr.method).toBe("POST");
    expect(lastXhr.url).toBe("/api/v1/items/123/attachments");
    expect(lastXhr.headers.Authorization).toBe("Bearer abc");
    expect(lastXhr.headers["X-Trace"]).toBe("t1");
    expect(lastXhr.sentBody).toBeInstanceOf(FormData);
    expect(result.status).toBe(201);
    expect(result.error).toBe(false);
    expect(result.data).toEqual({ id: "att1" });
  });

  test("forwards upload progress as fractions and emits 1 on completion", async () => {
    const fractions: number[] = [];
    const promise = xhrUpload({
      url: "/u",
      data: fd(),
      onProgress: f => fractions.push(f),
    });
    lastXhr.fireUploadProgress(0, 1000);
    lastXhr.fireUploadProgress(500, 1000);
    lastXhr.fireUploadProgress(1000, 1000);
    lastXhr.fireUploadLoad();
    lastXhr.complete(200, "OK", "{}");
    await promise;

    expect(fractions).toEqual([0, 0.5, 1, 1]);
  });

  test("flags 4xx as error and returns parsed JSON body", async () => {
    const promise = xhrUpload<{ error: string }>({ url: "/u", data: fd() });
    lastXhr.complete(422, "Unprocessable Entity", JSON.stringify({ error: "Validation Error" }));
    const result = await promise;

    expect(result.status).toBe(422);
    expect(result.error).toBe(true);
    expect(result.data).toEqual({ error: "Validation Error" });
  });

  test("returns raw text when Content-Type is not JSON", async () => {
    const promise = xhrUpload<string>({ url: "/u", data: fd() });
    lastXhr.complete(500, "Internal Server Error", "oops", "text/plain");
    const result = await promise;

    expect(result.status).toBe(500);
    expect(result.error).toBe(true);
    expect(result.data).toBe("oops");
  });

  test("returns empty object on 204 No Content", async () => {
    const promise = xhrUpload({ url: "/u", data: fd() });
    lastXhr.complete(204, "No Content", "");
    const result = await promise;

    expect(result.status).toBe(204);
    expect(result.error).toBe(false);
    expect(result.data).toEqual({});
  });

  test("rejects on network error", async () => {
    const promise = xhrUpload({ url: "/u", data: fd() });
    lastXhr.fail("error");
    await expect(promise).rejects.toThrow("Network request failed");
  });

  test("rejects on timeout", async () => {
    const promise = xhrUpload({ url: "/u", data: fd() });
    lastXhr.fail("timeout");
    await expect(promise).rejects.toThrow("timed out");
  });

  test("ignores non-computable progress events", async () => {
    const fractions: number[] = [];
    const promise = xhrUpload({ url: "/u", data: fd(), onProgress: f => fractions.push(f) });
    // Fire a non-computable progress event by calling the listener directly with lengthComputable=false
    for (const fn of lastXhr.upload.listeners.progress ?? []) {
      fn({ lengthComputable: false, loaded: 100, total: 0 });
    }
    lastXhr.complete(200, "OK", "{}");
    await promise;

    // Only the final load event fires a value
    expect(fractions).toEqual([]);
  });
});
