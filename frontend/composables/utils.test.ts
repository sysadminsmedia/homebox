import { describe, expect, test } from "vitest";
import { maybeUrl } from "./utils";

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
