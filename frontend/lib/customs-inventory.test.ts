import { describe, expect, test } from "vitest";
import {
  buildRequest,
  cutoffDate,
  DEFAULT_CUSTOMS_INVENTORY_PREFS,
  fallbackExportFilename,
  filenameFromDisposition,
  formatCents,
  formatQty,
  sanitizePrefs,
  serverErrorMessage,
  type CustomsInventoryPrefs,
  type CustomsInventorySearchContext,
} from "./customs-inventory";

const prefs: CustomsInventoryPrefs = {
  ...DEFAULT_CUSTOMS_INVENTORY_PREFS,
  header: { title: "Shipment A", name: "Jane Doe", address: "1 Example Street", phone: "", email: "" },
  referenceDate: "2026-11-15",
  groupBy: "tag",
  groupTagIds: ["b", "a"],
  dutyFreeTagIds: ["owned"],
};

const search: CustomsInventorySearchContext = {
  query: {
    q: "lamp",
    parentIds: [],
    entityTypeIds: [],
    tags: ["shipment"],
    negateTags: false,
    matchAllTags: false,
    onlyWithPhoto: false,
    onlyWithoutPhoto: false,
    includeArchived: false,
    fields: [],
  },
  total: 42,
};

describe("customs-inventory", () => {
  test("buildRequest selected", () => {
    const req = buildRequest({ scope: "selected", itemIds: ["1", "2"], search, prefs });
    expect(req.scope.itemIds).toEqual(["1", "2"]);
    expect(req.scope.query).toBeUndefined();
    expect(req.options.referenceDate).toBe("2026-11-15");
    expect(req.options.minMonths).toBe(6);
    expect(req.options.dutyFreeTagIds).toEqual(["owned"]);
    expect(req.options.receiptTypes).toEqual(["receipt"]);
    expect(req.header).toEqual(prefs.header);
    expect(req.header).not.toBe(prefs.header);
  });

  test("buildRequest query", () => {
    const req = buildRequest({ scope: "query", itemIds: ["1"], search, prefs });
    expect(req.scope.query).toEqual(search.query);
    expect(req.scope.itemIds).toEqual([]);
  });

  test("buildRequest query without search throws", () => {
    expect(() => buildRequest({ scope: "query", itemIds: [], search: null, prefs })).toThrow();
  });

  test("buildRequest keeps group tag order", () => {
    expect(buildRequest({ scope: "selected", itemIds: ["1"], search: null, prefs }).options.groupTagIds).toEqual([
      "b",
      "a",
    ]);
  });

  test("sanitizePrefs drops unknown tags and fills defaults", () => {
    const known = new Set(["known", "a"]);
    const p = sanitizePrefs(
      {
        dutyFreeTagIds: ["x", "known"],
        dutiableTagIds: ["gone"],
        groupTagIds: ["a", "z"],
        minMonths: 0,
        header: { title: "T" } as CustomsInventoryPrefs["header"],
      },
      known
    );
    expect(p.dutyFreeTagIds).toEqual(["known"]);
    expect(p.dutiableTagIds).toEqual([]);
    expect(p.groupTagIds).toEqual(["a"]);
    expect(p.receiptTypes).toEqual(["receipt"]);
    expect(p.minMonths).toBe(6);
    expect(p.header).toEqual({ title: "T", name: "", address: "", phone: "", email: "" });
    expect(p.mergeIdentical).toBe(true);
    expect(p.includeReceipts).toBe(true);

    expect(sanitizePrefs({ minMonths: 61 }, known).minMonths).toBe(6);
    expect(sanitizePrefs({ minMonths: 12 }, known).minMonths).toBe(12);
    expect(sanitizePrefs({ receiptTypes: ["photo", "warranty"] }, known).receiptTypes).toEqual(["warranty"]);
    expect(sanitizePrefs({ groupBy: "room" as CustomsInventoryPrefs["groupBy"] }, known).groupBy).toBe("location");
    expect(sanitizePrefs(undefined, known)).toEqual(DEFAULT_CUSTOMS_INVENTORY_PREFS);
  });

  test("cutoffDate", () => {
    expect(cutoffDate("2026-08-31", 6)).toBe("2026-02-28");
    expect(cutoffDate("2028-08-31", 6)).toBe("2028-02-29");
    expect(cutoffDate("2026-11-15", 6)).toBe("2026-05-15");
    expect(cutoffDate("2026-11-15", 12)).toBe("2025-11-15");
    expect(cutoffDate("2026-01-31", 6)).toBe("2025-07-31");
    expect(cutoffDate("garbage", 6)).toBe("");
    expect(cutoffDate("2026-13-01", 6)).toBe("");
    expect(cutoffDate("2026-11-15", 0)).toBe("");
  });

  test("formatCents", () => {
    expect(formatCents(1924990, "EUR", "de-DE")).toBe(
      new Intl.NumberFormat("de-DE", { style: "currency", currency: "EUR" }).format(19249.9)
    );
    expect(formatCents(-101, "USD", "en-US")).toBe(
      new Intl.NumberFormat("en-US", { style: "currency", currency: "USD" }).format(-1.01)
    );
  });

  test("filenameFromDisposition", () => {
    expect(filenameFromDisposition('attachment; filename="Inventarliste_x_2026-10-31.zip"', "f")).toBe(
      "Inventarliste_x_2026-10-31.zip"
    );
    expect(filenameFromDisposition("attachment; filename=Inventarliste_x.xlsx", "f")).toBe("Inventarliste_x.xlsx");
    expect(filenameFromDisposition(null, "fallback.zip")).toBe("fallback.zip");
    expect(filenameFromDisposition("attachment", "fallback.zip")).toBe("fallback.zip");
  });

  test("fallbackExportFilename picks the extension from the content type", () => {
    expect(fallbackExportFilename("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")).toBe(
      "Inventarliste.xlsx"
    );
    expect(fallbackExportFilename("application/zip")).toBe("Inventarliste.zip");
    expect(fallbackExportFilename(null)).toBe("Inventarliste.zip");
  });

  test("serverErrorMessage", () => {
    expect(serverErrorMessage({ error: "customs inventory: scope too large: 3 entities, at most 2" })).toBe(
      "customs inventory: scope too large: 3 entities, at most 2"
    );
    expect(serverErrorMessage({})).toBe("");
    expect(serverErrorMessage({ error: 42 })).toBe("");
    expect(serverErrorMessage(null)).toBe("");
  });
});

describe("formatQty", () => {
  test("formats pieces for the locale without float noise", () => {
    expect(formatQty(12, "en")).toBe("12");
    expect(formatQty(2.5, "en")).toBe("2.5");
    expect(formatQty(2.5, "de")).toBe("2,5");
    expect(formatQty(0.1 + 0.2, "en")).toBe("0.3");
  });
});

describe("mergeFirstWord", () => {
  test("defaults to false and keeps a saved value", () => {
    expect(sanitizePrefs(undefined, new Set()).mergeFirstWord).toBe(false);
    expect(sanitizePrefs({ mergeFirstWord: true }, new Set()).mergeFirstWord).toBe(true);
    expect(sanitizePrefs({ mergeFirstWord: "yes" as unknown as boolean }, new Set()).mergeFirstWord).toBe(false);
  });

  test("is sent with the request", () => {
    const prefs = { ...sanitizePrefs(undefined, new Set()), mergeFirstWord: true };
    expect(buildRequest({ scope: "selected", itemIds: ["a"], search: null, prefs }).options.mergeFirstWord).toBe(true);
  });
});
