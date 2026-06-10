import { describe, expect, test } from "vitest";
import type { TResponse } from "./requests";
import { isRetryableStatus, withRetries } from "./retry";

function ok<T>(data: T): TResponse<T> {
  return { status: 200, error: false, data, response: {} as Response };
}

function fail(status: number, data: unknown = null, statusText = ""): TResponse<unknown> {
  return { status, error: true, data, response: { statusText } as Response };
}

const noSleep = () => Promise.resolve();

describe("isRetryableStatus", () => {
  test("treats status 0 (thrown / no response) as retryable", () => {
    expect(isRetryableStatus(0)).toBe(true);
  });

  test("retries on 408, 429, and all 5xx", () => {
    expect(isRetryableStatus(408)).toBe(true);
    expect(isRetryableStatus(429)).toBe(true);
    expect(isRetryableStatus(500)).toBe(true);
    expect(isRetryableStatus(503)).toBe(true);
  });

  test("does not retry on 2xx, 3xx, or other 4xx", () => {
    expect(isRetryableStatus(200)).toBe(false);
    expect(isRetryableStatus(304)).toBe(false);
    expect(isRetryableStatus(400)).toBe(false);
    expect(isRetryableStatus(404)).toBe(false);
    expect(isRetryableStatus(422)).toBe(false);
  });
});

describe("withRetries", () => {
  test("returns ok immediately on success", async () => {
    let calls = 0;
    const result = await withRetries(
      async () => {
        calls++;
        return ok({ id: "abc" });
      },
      { sleep: noSleep }
    );
    expect(result).toEqual({ ok: true, data: { id: "abc" } });
    expect(calls).toBe(1);
  });

  test("retries thrown errors then succeeds", async () => {
    let calls = 0;
    const result = await withRetries(
      async () => {
        calls++;
        if (calls < 3) throw new Error("network down");
        return ok({ id: "abc" });
      },
      { sleep: noSleep }
    );
    expect(result).toEqual({ ok: true, data: { id: "abc" } });
    expect(calls).toBe(3);
  });

  test("retries on 5xx then succeeds", async () => {
    let calls = 0;
    const result = await withRetries(
      async () => {
        calls++;
        if (calls < 2) return fail(503, { error: "upstream down" });
        return ok({ id: "abc" });
      },
      { sleep: noSleep }
    );
    expect(result.ok).toBe(true);
    expect(calls).toBe(2);
  });

  test("does NOT retry on 4xx — returns the failure with extracted reason", async () => {
    let calls = 0;
    const result = await withRetries(
      async () => (
        calls++,
        fail(422, {
          error: "Validation Error",
          fields: { Name: "Field validation for 'Name' failed on the 'required' tag" },
        })
      ),
      { sleep: noSleep }
    );
    expect(calls).toBe(1);
    expect(result).toEqual({ ok: false, reason: "Name: Name is required", status: 422 });
  });

  test("gives up after maxRetries and returns the last reason", async () => {
    let calls = 0;
    const result = await withRetries(
      async () => {
        calls++;
        throw new Error("network down");
      },
      { sleep: noSleep, retries: 2 }
    );
    expect(calls).toBe(3); // 1 initial + 2 retries
    expect(result).toEqual({ ok: false, reason: "network down", status: 0 });
  });

  test("uses the configured backoff schedule", async () => {
    const delays: number[] = [];
    await withRetries(
      async () => {
        throw new Error("again");
      },
      {
        retries: 3,
        backoffMs: [10, 20, 30],
        sleep: async ms => {
          delays.push(ms);
        },
      }
    );
    expect(delays).toEqual([10, 20, 30]);
  });

  test("reuses the last backoff entry when attempts exceed schedule length", async () => {
    const delays: number[] = [];
    await withRetries(
      async () => {
        throw new Error("again");
      },
      {
        retries: 4,
        backoffMs: [10],
        sleep: async ms => {
          delays.push(ms);
        },
      }
    );
    expect(delays).toEqual([10, 10, 10, 10]);
  });

  test("network failure surfaces the thrown Error's message", async () => {
    const result = await withRetries(
      async () => {
        throw new TypeError("Failed to fetch");
      },
      { sleep: noSleep, retries: 0 }
    );
    expect(result).toEqual({ ok: false, reason: "Failed to fetch", status: 0 });
  });
});
