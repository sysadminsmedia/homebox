import { describe, test, expect } from "vitest";
import { format, zeroTime, factorRange, parse } from "./datelib";

describe("format", () => {
  test("should format a date as a string", () => {
    const date = new Date(2020, 1, 1);
    expect(format(date)).toBe("2020-02-01");
  });

  test("should return the string if a string is passed in", () => {
    expect(format("2020-02-01")).toBe("2020-02-01");
  });
});

describe("zeroTime", () => {
  test("should zero out the time", () => {
    const date = new Date(2020, 1, 1, 12, 30, 30);
    const zeroed = zeroTime(date);
    expect(zeroed.getHours()).toBe(0);
    expect(zeroed.getMinutes()).toBe(0);
    expect(zeroed.getSeconds()).toBe(0);
  });
});

describe("factorRange", () => {
  test("should return a range of dates", () => {
    const [start, end] = factorRange(10);
    // Start should be today
    expect(start).toBeInstanceOf(Date);
    expect(start.getFullYear()).toBe(new Date().getFullYear());

    // End should be 10 days from now
    expect(end).toBeInstanceOf(Date);
    expect(end.getFullYear()).toBe(new Date().getFullYear());
  });
});

describe("parse", () => {
  test("should parse a date string", () => {
    const date = parse("2020-02-01");
    expect(date).toBeInstanceOf(Date);
  });
});
