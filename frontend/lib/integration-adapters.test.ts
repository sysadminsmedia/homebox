import { describe, expect, it } from "vitest";
import {
  SERVICE_ADAPTERS,
  classifyDroppedUrl,
  detectServiceFromUrl,
  extractPaperlessDocId,
  getAdapter,
  getAdapterByMimeType,
} from "./integration-adapters";

describe("integration adapters", () => {
  describe("extractPaperlessDocId", () => {
    it("extracts doc id from /documents/{id}", () => {
      expect(extractPaperlessDocId("https://paperless.local/documents/42", "https://paperless.local")).toBe("42");
    });

    it("extracts doc id from /documents/{id}/details", () => {
      expect(extractPaperlessDocId("https://paperless.local/documents/42/details", "https://paperless.local")).toBe(
        "42"
      );
    });

    it("returns null for foreign host", () => {
      expect(extractPaperlessDocId("https://paperless.local.evil/documents/42", "https://paperless.local")).toBeNull();
    });

    it("supports base path setups", () => {
      expect(
        extractPaperlessDocId(
          "https://example.local/paperless/documents/500/details",
          "https://example.local/paperless"
        )
      ).toBe("500");
    });

    it("extracts doc id without base URL", () => {
      expect(extractPaperlessDocId("https://paperless.local/documents/77/details")).toBe("77");
    });

    it("extracts doc id when configured base URL is invalid", () => {
      expect(extractPaperlessDocId("https://paperless.local/documents/99/details", "homebox")).toBe("99");
    });

    it("extracts doc id from bare-hostname URLs", () => {
      expect(extractPaperlessDocId("localhost/documents/2/details")).toBe("2");
    });

    it("trims whitespace around URL and base URL", () => {
      expect(extractPaperlessDocId("  http://localhost:8000/documents/7/  ", "  http://localhost:8000/  ")).toBe("7");
    });
  });

  describe("detectServiceFromUrl", () => {
    const settings = {
      paperless_url: "https://paperless.local",
    } as Record<string, string>;

    it("detects paperless via configured base URL", () => {
      expect(detectServiceFromUrl("https://paperless.local/documents/1", settings)?.name).toBe("paperless");
    });

    it("detects paperless with sub-path installation", () => {
      const subPathSettings = { paperless_url: "https://example.local/paperless" } as Record<string, string>;
      expect(detectServiceFromUrl("https://example.local/paperless/documents/1", subPathSettings)?.name).toBe(
        "paperless"
      );
    });

    it("returns null for paperless URL when settings are empty (no fallback)", () => {
      expect(detectServiceFromUrl("https://any.host/documents/1/details", {})).toBeNull();
    });

    it("does not match host-prefix attacks", () => {
      expect(detectServiceFromUrl("https://paperless.local.evil/documents/1", settings)).toBeNull();
    });

    it("returns null for unknown services", () => {
      expect(detectServiceFromUrl("https://example.org/something", settings)).toBeNull();
    });
  });

  describe("classifyDroppedUrl", () => {
    const settings = {
      paperless_url: "https://paperless.local",
    } as Record<string, string>;

    it("classifies paperless URL", () => {
      const result = classifyDroppedUrl("https://paperless.local/documents/42/details", settings);
      expect(result?.adapter.name).toBe("paperless");
      expect(result?.id).toBe("42");
    });

    it("returns null for paperless URL with no settings configured (no fallback)", () => {
      expect(classifyDroppedUrl("https://any.host/documents/7/details", {})).toBeNull();
    });

    it("returns null for unrecognised URL", () => {
      expect(classifyDroppedUrl("https://example.org/something", settings)).toBeNull();
    });

    it("returns null when URL matches service host but id cannot be extracted", () => {
      // URL host matches paperless but path has no document ID
      expect(classifyDroppedUrl("https://paperless.local/dashboard", settings)).toBeNull();
    });

    it("classifies paperless URL with whitespace around value", () => {
      const result = classifyDroppedUrl("  https://paperless.local/documents/42/details  ", settings);
      expect(result?.adapter.name).toBe("paperless");
      expect(result?.id).toBe("42");
    });
  });

  describe("getAdapter", () => {
    it("returns paperless adapter by name", () => {
      expect(getAdapter("paperless")?.name).toBe("paperless");
    });

    it("returns undefined for unknown name", () => {
      expect(getAdapter("unknown")).toBeUndefined();
    });
  });

  describe("getAdapterByMimeType", () => {
    it("returns paperless adapter for paperless/document", () => {
      expect(getAdapterByMimeType("paperless/document")?.name).toBe("paperless");
    });

    it("returns undefined for unknown MIME type", () => {
      expect(getAdapterByMimeType("application/pdf")).toBeUndefined();
    });
  });

  describe("adapter buildTitle", () => {
    it("paperless buildTitle includes the doc id", () => {
      expect(getAdapter("paperless")?.buildTitle("42")).toBe("Paperless Document 42");
    });
  });

  describe("SERVICE_ADAPTERS registry", () => {
    it("contains paperless", () => {
      const names = SERVICE_ADAPTERS.map(a => a.name);
      expect(names).toContain("paperless");
    });

    it("each adapter has required fields", () => {
      for (const adapter of SERVICE_ADAPTERS) {
        expect(adapter.name).toBeTruthy();
        expect(adapter.mimeType).toContain("/");
        expect(adapter.settingsUrlKey).toContain("_url");
        expect(adapter.settingsTokenKey).toContain("_token");
        expect(typeof adapter.extractId).toBe("function");
        expect(typeof adapter.buildTitle).toBe("function");
      }
    });
  });
});
