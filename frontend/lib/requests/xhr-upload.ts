import type { TResponse } from "./requests";

export interface XhrUploadArgs {
  url: string;
  data: FormData;
  headers?: Record<string, string>;
  /** Called with the fraction (0..1) of the body uploaded. May fire many times; the final call is always 1. */
  onProgress?: (fraction: number) => void;
}

/**
 * POST a FormData body with real upload progress via `XMLHttpRequest.upload.onprogress`.
 * Mirrors the `TResponse<T>` shape returned by `Requests.post` so call sites can swap
 * fetch for XHR transparently. Resolves on every HTTP outcome (including 4xx/5xx) so
 * callers can read `status` / `error`; rejects only when the request itself fails to send
 * (network drop, CORS, abort) — matching `fetch()` semantics.
 */
export function xhrUpload<T>(args: XhrUploadArgs): Promise<TResponse<T>> {
  return new Promise<TResponse<T>>((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open("POST", args.url, true);

    if (args.headers) {
      for (const [name, value] of Object.entries(args.headers)) {
        xhr.setRequestHeader(name, value);
      }
    }

    if (args.onProgress) {
      xhr.upload.addEventListener("progress", e => {
        if (e.lengthComputable && e.total > 0) {
          args.onProgress!(Math.min(1, e.loaded / e.total));
        }
      });
      xhr.upload.addEventListener("load", () => args.onProgress!(1));
    }

    xhr.addEventListener("load", () => {
      const contentType = xhr.getResponseHeader("Content-Type") ?? "";
      let data: T = {} as T;
      if (xhr.status !== 204) {
        if (contentType.startsWith("application/json")) {
          try {
            data = xhr.responseText ? (JSON.parse(xhr.responseText) as T) : ({} as T);
          } catch {
            data = {} as T;
          }
        } else {
          data = xhr.responseText as unknown as T;
        }
      }

      // Build a Response-shaped object so consumers that read `resp.response.statusText` still work.
      // We use a real Response when available (browser) — falling back to a duck-typed object in test envs.
      // Status codes 1xx/204/205/304 must have a null body per the Response spec.
      const nullBodyStatuses = new Set([101, 204, 205, 304]);
      const body = nullBodyStatuses.has(xhr.status) ? null : xhr.responseText;
      const response =
        typeof Response !== "undefined"
          ? new Response(body, { status: xhr.status, statusText: xhr.statusText })
          : ({ statusText: xhr.statusText } as Response);

      resolve({ status: xhr.status, error: xhr.status < 200 || xhr.status >= 300, data, response });
    });

    xhr.addEventListener("error", () => reject(new TypeError("Network request failed")));
    xhr.addEventListener("abort", () => reject(new DOMException("Request aborted", "AbortError")));
    xhr.addEventListener("timeout", () => reject(new TypeError("Request timed out")));

    xhr.send(args.data);
  });
}
