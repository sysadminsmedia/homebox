import { describe, expect, test } from "vitest";
import { factorRange, format, parse, zeroTime } from "./datelib";

describe("format", () => {
  test("should format a date as an ISO string", () => {
    const date = new Date(Date.UTC(2020, 1, 1));
    expect(format(date)).toBe("2020-02-01T00:00:00.000Z");
  });

  test("should return the string if a string is passed in", () => {
    expect(format("2020-02-01T00:00:00.000Z")).toBe("2020-02-01T00:00:00.000Z");
  });
});

describe("zeroTime", () => {
  test("should create a UTC date without time components", () => {
    const date = new Date(2020, 1, 1, 12, 30, 30);
    const utcDate = zeroTime(date);
    expect(utcDate.getUTCFullYear()).toBe(2020);
    expect(utcDate.getUTCMonth()).toBe(1);
    expect(utcDate.getUTCDate()).toBe(1);
    expect(utcDate.getUTCHours()).toBe(0);
    expect(utcDate.getUTCMinutes()).toBe(0);
    expect(utcDate.getUTCSeconds()).toBe(0);
    expect(utcDate.getUTCMilliseconds()).toBe(0);
  });
});

describe("factorRange", () => {
  test("should return a UTC date range", () => {
    const [start, end] = factorRange(10);

    // Verify UTC dates
    expect(start.getUTCHours()).toBe(0);
    expect(start.getUTCMinutes()).toBe(0);
    expect(start.getUTCSeconds()).toBe(0);
    expect(start.getUTCMilliseconds()).toBe(0);
    
    expect(end.getTime() - start.getTime()).toBe(10 * 24 * 60 * 60 * 1000);
  });
});

describe("parse", () => {
  test("should parse an ISO date string", () => {
    const date = parse("2020-02-01T00:00:00.000Z");
    expect(date.getUTCFullYear()).toBe(2020);
    expect(date.getUTCMonth()).toBe(1);
    expect(date.getUTCDate()).toBe(1);
  });
});
