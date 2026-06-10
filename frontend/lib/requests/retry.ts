import type { TResponse } from "./requests";
import { extractErrorMessage } from "./extract-error";

export type RetryResult<T> = { ok: true; data: T } | { ok: false; reason: string; status: number };

export interface RetryOptions {
  /** Maximum number of retry attempts AFTER the first try (so total attempts = retries + 1). Default: 2. */
  retries?: number;
  /** Delays in ms between attempts. The last entry is reused for attempts beyond its length. Default: [1000, 2000, 4000]. */
  backoffMs?: number[];
  /** Override the default sleep implementation — primarily for tests. */
  sleep?: (ms: number) => Promise<void>;
}

const DEFAULT_BACKOFF_MS = [1000, 2000, 4000];
const defaultSleep = (ms: number) => new Promise<void>(resolve => setTimeout(resolve, ms));

// status 0 means the request never reached the server (caught exception);
// 408/429/5xx are transient by spec. Anything else is treated as final.
export function isRetryableStatus(status: number): boolean {
  return status === 0 || status === 408 || status === 429 || status >= 500;
}

/**
 * Invoke `call` and retry it on transient failures with exponential-ish backoff.
 * A `call` that throws is treated as a network failure (status 0, retryable).
 * Resolves to a tagged-union `RetryResult` so callers don't need to handle exceptions themselves.
 */
export async function withRetries<T>(
  call: () => Promise<TResponse<T>>,
  options: RetryOptions = {}
): Promise<RetryResult<T>> {
  const maxRetries = options.retries ?? 2;
  const backoff = options.backoffMs ?? DEFAULT_BACKOFF_MS;
  const sleep = options.sleep ?? defaultSleep;

  let lastReason = "Request failed";
  let lastStatus = 0;

  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    if (attempt > 0) {
      const delay = backoff[attempt - 1] ?? backoff[backoff.length - 1] ?? 4000;
      await sleep(delay);
    }

    let resp: TResponse<T> | undefined;
    let thrown: unknown;
    try {
      resp = await call();
    } catch (e) {
      thrown = e;
    }

    if (resp && !resp.error) {
      return { ok: true, data: resp.data };
    }

    lastReason = extractErrorMessage(resp, thrown);
    lastStatus = resp?.status ?? 0;

    const retryable = thrown !== undefined || isRetryableStatus(lastStatus);
    if (!retryable) break;
  }

  return { ok: false, reason: lastReason, status: lastStatus };
}
