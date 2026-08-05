import { beforeEach, describe, expect, it } from "vitest";
import { overrideParts, route } from ".";

describe("UrlBuilder", () => {
  beforeEach(() => {
    overrideParts("http://localhost.com", "/api/v1", "/");
  });

  it("basic query parameter", () => {
    const result = route("/test", { a: "b" });
    expect(result).toBe("/api/v1/test?a=b");
  });

  it("multiple query parameters", () => {
    const result = route("/test", { a: "b", c: "d" });
    expect(result).toBe("/api/v1/test?a=b&c=d");
  });

  it("no query parameters", () => {
    const result = route("/test");
    expect(result).toBe("/api/v1/test");
  });

  it("list-like query parameters", () => {
    const result = route("/test", { a: ["b", "c"] });
    expect(result).toBe("/api/v1/test?a=b&a=c");
  });

  it("respects custom base path", () => {
    overrideParts("http://localhost.com", "/api/v1", "/homebox/");
    const result = route("/test");
    expect(result).toBe("/homebox/api/v1/test");
  });
});
