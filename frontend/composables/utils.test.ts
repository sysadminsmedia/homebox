import { describe, expect, test, afterEach, vi } from "vitest";
import { maybeUrl, parseScanResult } from "./utils";

describe("maybeURL works as expected", () => {
  test("basic valid URL case", () => {
    const result = maybeUrl("https://example.com");
    expect(result.isUrl).toBe(true);
    expect(result.url).toBe("https://example.com");
    expect(result.text).toBe("https://example.com");
  });

  test("special URL syntax", () => {
    const result = maybeUrl("[My Text](http://example.com)");
    expect(result.isUrl).toBe(true);
    expect(result.url).toBe("http://example.com");
    expect(result.text).toBe("My Text");
  });

  test("not a url", () => {
    const result = maybeUrl("not a url");
    expect(result.isUrl).toBe(false);
    expect(result.url).toBe("");
    expect(result.text).toBe("");
  });

  test("malformed special syntax", () => {
    const result = maybeUrl("[My Text(http://example.com)");
    expect(result.isUrl).toBe(false);
    expect(result.url).toBe("");
    expect(result.text).toBe("");
  });
});

describe("parseScanResult", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  test("parses full https URL", () => {
    const url = parseScanResult("https://example.com/a/1");
    expect(url?.origin).toBe("https://example.com");
    expect(url?.pathname).toBe("/a/1");
  });

  test("parses full http URL", () => {
    const url = parseScanResult("http://example.com/a/1");
    expect(url?.origin).toBe("http://example.com");
    expect(url?.pathname).toBe("/a/1");
  });

  test("normalizes uppercase protocol and host", () => {
    const url = parseScanResult("HTTPS://EXAMPLE.COM/A/1");
    expect(url?.origin).toBe("https://example.com");
    // Path case is preserved by the URL spec.
    expect(url?.pathname).toBe("/A/1");
  });

  test("accepts protocol-less host+path payload", () => {
    const url = parseScanResult("example.com/a/1");
    expect(url?.pathname).toBe("/a/1");
    expect(url?.host).toBe("example.com");
  });

  test("accepts uppercase protocol-less payload (C40 Data Matrix)", () => {
    const url = parseScanResult("EXAMPLE.COM/A/1");
    expect(url?.host).toBe("example.com");
    expect(url?.pathname).toBe("/A/1");
  });

  test("uses current page protocol for protocol-less payloads", () => {
    vi.stubGlobal("location", { protocol: "http:", origin: "http://example.com" });
    const url = parseScanResult("example.com/a/1");
    expect(url?.origin).toBe("http://example.com");
  });

  test("rejects bare numeric barcode (no slash)", () => {
    expect(parseScanResult("1234567890123")).toBeNull();
  });

  test("rejects arbitrary text without dot+slash", () => {
    expect(parseScanResult("hello world")).toBeNull();
  });

  test("rejects bare path so callers can handle it themselves", () => {
    // parseHomeboxUrl handles raw "/a/1" inputs directly; this helper only
    // produces URL objects.
    expect(parseScanResult("/a/1")).toBeNull();
  });

  test("rejects protocol-relative URL (cannot be parsed without base)", () => {
    expect(parseScanResult("//evil.example/a/1")).toBeNull();
  });
});
