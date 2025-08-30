import { describe, expect, test } from "vitest";
import { scorePassword } from ".";

describe("scorePassword tests", () => {
  test("flagged words should return negative number", () => {
    const flaggedWords = ["password", "homebox", "admin", "qwerty", "login"];

    for (const word of flaggedWords) {
      expect(scorePassword(word)).toBe(0);
    }
  });

  test("should return 0 for empty string", () => {
    expect(scorePassword("")).toBe(0);
  });

  test("should return 0 for strings less than 6", () => {
    expect(scorePassword("12345")).toBe(0);
  });

  test("should return positive number for long string", () => {
    const result = expect(scorePassword("123456"));
    result.toBeGreaterThan(0);
    result.toBeLessThan(31);
  });

  test("should return max number for long string with all variations", () => {
    expect(scorePassword("3bYWcfYOwqxljqeOmQXTLlBwkrH6HV")).toBe(100);
  });
});
