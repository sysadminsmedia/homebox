import type { TResponse } from "./requests";

// Strip Go validator boilerplate so messages like
// "Key: 'ItemCreate.Name' Error:Field validation for 'Name' failed on the 'required' tag"
// become "Name is required" / "Name failed validation (min)".
function cleanValidatorMessage(raw: string): string {
  const match = raw.match(/Field validation for '([^']+)' failed on the '([^']+)' tag/);
  if (!match) return raw;
  const [, field, tag] = match;
  return tag === "required" ? `${field} is required` : `${field} failed validation (${tag})`;
}

function formatFields(fields: Record<string, unknown>): string {
  const parts: string[] = [];
  for (const [field, value] of Object.entries(fields)) {
    if (typeof value === "string" && value.length > 0) {
      parts.push(`${field}: ${cleanValidatorMessage(value)}`);
    }
  }
  return parts.join("; ");
}

/**
 * Extract a human-readable error message from an API response or a thrown exception.
 * Knows the HomeBox backend's `ErrorResponse { error, fields? }` shape and tolerates
 * non-JSON bodies, validation-error arrays, and unexpected payloads.
 */
export function extractErrorMessage(resp: TResponse<unknown> | undefined, thrown?: unknown): string {
  if (thrown !== undefined) {
    return thrown instanceof Error && thrown.message ? thrown.message : String(thrown);
  }
  if (!resp) return "Unknown error";

  const data = resp.data as unknown;

  if (data && typeof data === "object" && !Array.isArray(data)) {
    const obj = data as Record<string, unknown>;
    const fields = obj.fields;
    if (fields && typeof fields === "object" && Object.keys(fields).length > 0) {
      const formatted = formatFields(fields as Record<string, unknown>);
      if (formatted) return formatted;
    }
    for (const key of ["error", "message", "detail", "title"]) {
      const value = obj[key];
      if (typeof value === "string" && value.length > 0) return value;
    }
  }

  // Legacy/array shape some endpoints emit: [{ field, error }]
  if (Array.isArray(data) && data.length > 0) {
    const parts: string[] = [];
    for (const entry of data) {
      if (entry && typeof entry === "object") {
        const e = entry as Record<string, unknown>;
        if (typeof e.field === "string" && typeof e.error === "string") {
          parts.push(`${e.field}: ${cleanValidatorMessage(e.error)}`);
        } else if (typeof e.error === "string") {
          parts.push(e.error);
        }
      }
    }
    if (parts.length > 0) return parts.join("; ");
  }

  const statusText = resp.response?.statusText;
  if (statusText) return `${statusText} (HTTP ${resp.status})`;
  return `HTTP ${resp.status}`;
}
