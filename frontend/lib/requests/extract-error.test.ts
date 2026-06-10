import { describe, expect, test } from "vitest";
import type { TResponse } from "./requests";
import { extractErrorMessage } from "./extract-error";

function resp(status: number, data: unknown, statusText = ""): TResponse<unknown> {
  return {
    status,
    error: status >= 400,
    data,
    response: { statusText } as Response,
  };
}

describe("extractErrorMessage", () => {
  test("returns thrown Error message when provided", () => {
    expect(extractErrorMessage(undefined, new Error("network timed out"))).toBe("network timed out");
  });

  test("falls back to String() for non-Error throws", () => {
    expect(extractErrorMessage(undefined, "boom")).toBe("boom");
  });

  test("returns Unknown error when nothing is given", () => {
    expect(extractErrorMessage(undefined)).toBe("Unknown error");
  });

  test("extracts top-level error key from backend ErrorResponse", () => {
    expect(extractErrorMessage(resp(400, { error: "failed to parse multipart form" }))).toBe(
      "failed to parse multipart form"
    );
  });

  test("formats validation fields and strips Go validator boilerplate", () => {
    const data = {
      error: "Validation Error",
      fields: {
        Name: "Key: 'ItemCreate.Name' Error:Field validation for 'Name' failed on the 'required' tag",
        Quantity: "Key: 'ItemCreate.Quantity' Error:Field validation for 'Quantity' failed on the 'min' tag",
      },
    };
    const msg = extractErrorMessage(resp(422, data));
    expect(msg).toContain("Name: Name is required");
    expect(msg).toContain("Quantity: Quantity failed validation (min)");
  });

  test("prefers fields over the generic top-level error when both are present", () => {
    const data = {
      error: "Validation Error",
      fields: { Name: "Field validation for 'Name' failed on the 'required' tag" },
    };
    expect(extractErrorMessage(resp(422, data))).toBe("Name: Name is required");
  });

  test("handles legacy [{ field, error }] array shape", () => {
    const data = [
      { field: "Name", error: "Field validation for 'Name' failed on the 'required' tag" },
      { field: "Quantity", error: "must be positive" },
    ];
    const msg = extractErrorMessage(resp(422, data));
    expect(msg).toContain("Name: Name is required");
    expect(msg).toContain("Quantity: must be positive");
  });

  test("falls back to message / detail / title keys", () => {
    expect(extractErrorMessage(resp(500, { message: "boom" }))).toBe("boom");
    expect(extractErrorMessage(resp(500, { detail: "internal" }))).toBe("internal");
    expect(extractErrorMessage(resp(500, { title: "Server Error" }))).toBe("Server Error");
  });

  test("falls back to statusText when no body fields match", () => {
    expect(extractErrorMessage(resp(503, {}, "Service Unavailable"))).toBe("Service Unavailable (HTTP 503)");
  });

  test("falls back to HTTP status when neither body nor statusText is useful", () => {
    expect(extractErrorMessage(resp(418, null))).toBe("HTTP 418");
  });

  test("ignores empty fields map", () => {
    const data = { error: "Bad Request", fields: {} };
    expect(extractErrorMessage(resp(400, data))).toBe("Bad Request");
  });

  test("ignores empty strings in fields map", () => {
    const data = { error: "Validation Error", fields: { Name: "" } };
    expect(extractErrorMessage(resp(422, data))).toBe("Validation Error");
  });
});
