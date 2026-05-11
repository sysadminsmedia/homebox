import { describe, expect, it } from "vitest";
import {
  SERVICE_ADAPTERS,
  classifyDroppedUrl,
  detectServiceFromUrl,
  extractImmichAssetId,
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

  describe("extractImmichAssetId", () => {
    it("extracts asset id from /assets/{id}", () => {
      expect(
        extractImmichAssetId("https://immich.local/assets/1df4f848-c8fc-4e72-bfba-20159757da8f", "https://immich.local")
      ).toBe("1df4f848-c8fc-4e72-bfba-20159757da8f");
    });

    it("returns null for host mismatch", () => {
      expect(extractImmichAssetId("https://immich.local.evil/assets/abc", "https://immich.local")).toBeNull();
    });

    it("supports base path setups", () => {
      expect(
        extractImmichAssetId(
          "https://example.local/photos/assets/1df4f848-c8fc-4e72-bfba-20159757da8f",
          "https://example.local/photos"
        )
      ).toBe("1df4f848-c8fc-4e72-bfba-20159757da8f");
    });

    it("extracts asset id without base URL", () => {
      expect(extractImmichAssetId("https://immich.local/assets/1df4f848-c8fc-4e72-bfba-20159757da8f")).toBe(
        "1df4f848-c8fc-4e72-bfba-20159757da8f"
      );
    });

    it("extracts asset id from bare-hostname URLs", () => {
      expect(extractImmichAssetId("localhost/assets/1df4f848-c8fc-4e72-bfba-20159757da8f")).toBe(
        "1df4f848-c8fc-4e72-bfba-20159757da8f"
      );
    });

    it("trims whitespace around URL and base URL", () => {
      expect(
        extractImmichAssetId(
          "  https://immich.local/assets/1df4f848-c8fc-4e72-bfba-20159757da8f  ",
          "  https://immich.local  "
        )
      ).toBe("1df4f848-c8fc-4e72-bfba-20159757da8f");
    });
  });

  describe("detectServiceFromUrl", () => {
    const settings = {
      paperless_url: "https://paperless.local",
      immich_url: "https://immich.local",
    } as Record<string, string>;

    it("detects paperless via configured base URL", () => {
      expect(detectServiceFromUrl("https://paperless.local/documents/1", settings)?.name).toBe("paperless");
    });

    it("detects immich via configured base URL", () => {
      expect(detectServiceFromUrl("https://immich.local/assets/1", settings)?.name).toBe("immich");
    });

    it("detects paperless via URL pattern when settings empty", () => {
      expect(detectServiceFromUrl("https://any.host/documents/1/details", {})?.name).toBe("paperless");
    });

    it("detects immich via URL pattern when settings empty", () => {
      expect(detectServiceFromUrl("https://any.host/assets/1df4f848-c8fc-4e72-bfba-20159757da8f", {})?.name).toBe(
        "immich"
      );
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
      immich_url: "https://immich.local",
    } as Record<string, string>;

    it("classifies paperless URL", () => {
      const result = classifyDroppedUrl("https://paperless.local/documents/42/details", settings);
      expect(result?.adapter.name).toBe("paperless");
      expect(result?.id).toBe("42");
    });

    it("classifies immich URL", () => {
      const result = classifyDroppedUrl("https://immich.local/assets/1df4f848-c8fc-4e72-bfba-20159757da8f", settings);
      expect(result?.adapter.name).toBe("immich");
      expect(result?.id).toBe("1df4f848-c8fc-4e72-bfba-20159757da8f");
    });

    it("classifies paperless URL with missing settings (pattern fallback)", () => {
      const result = classifyDroppedUrl("https://any.host/documents/7/details", {});
      expect(result?.adapter.name).toBe("paperless");
      expect(result?.id).toBe("7");
    });

    it("returns null for unrecognised URL", () => {
      expect(classifyDroppedUrl("https://example.org/something", settings)).toBeNull();
    });

    it("classifies paperless URL with whitespace around value", () => {
      const result = classifyDroppedUrl("  https://paperless.local/documents/42/details  ", settings);
      expect(result?.adapter.name).toBe("paperless");
      expect(result?.id).toBe("42");
    });

    it("classifies immich URL with no settings (pattern fallback)", () => {
      const result = classifyDroppedUrl("https://any.host/assets/1df4f848-c8fc-4e72-bfba-20159757da8f", {});
      expect(result?.adapter.name).toBe("immich");
      expect(result?.id).toBe("1df4f848-c8fc-4e72-bfba-20159757da8f");
    });
  });

  describe("getAdapter", () => {
    it("returns paperless adapter by name", () => {
      expect(getAdapter("paperless")?.name).toBe("paperless");
    });

    it("returns immich adapter by name", () => {
      expect(getAdapter("immich")?.name).toBe("immich");
    });

    it("returns undefined for unknown name", () => {
      expect(getAdapter("unknown")).toBeUndefined();
    });
  });

  describe("getAdapterByMimeType", () => {
    it("returns paperless adapter for paperless/document", () => {
      expect(getAdapterByMimeType("paperless/document")?.name).toBe("paperless");
    });

    it("returns immich adapter for immich/asset", () => {
      expect(getAdapterByMimeType("immich/asset")?.name).toBe("immich");
    });

    it("returns undefined for unknown MIME type", () => {
      expect(getAdapterByMimeType("application/pdf")).toBeUndefined();
    });
  });

  describe("adapter buildTitle", () => {
    it("paperless buildTitle includes the doc id", () => {
      expect(getAdapter("paperless")?.buildTitle("42")).toBe("Paperless Document 42");
    });

    it("immich buildTitle includes the asset id", () => {
      expect(getAdapter("immich")?.buildTitle("1df4f848")).toBe("Immich Asset 1df4f848");
    });
  });

  describe("SERVICE_ADAPTERS registry", () => {
    it("contains paperless and immich", () => {
      const names = SERVICE_ADAPTERS.map(a => a.name);
      expect(names).toContain("paperless");
      expect(names).toContain("immich");
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
