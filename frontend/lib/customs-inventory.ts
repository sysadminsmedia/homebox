import type {
  CustomsInventoryGroupBy,
  CustomsInventoryHeader,
  CustomsInventoryRequest,
  CustomsInventoryScopeQuery,
} from "~~/lib/api/types/data-contracts";

export type CustomsInventorySearchContext = { query: CustomsInventoryScopeQuery; total: number };

export type CustomsInventoryGrouping = "location" | "entityType" | "tag";

export type CustomsInventoryPrefs = {
  header: CustomsInventoryHeader;
  referenceDate: string;
  minMonths: number;
  groupBy: CustomsInventoryGrouping;
  groupTagIds: string[];
  dutyFreeTagIds: string[];
  dutiableTagIds: string[];
  mergeIdentical: boolean;
  includeReceipts: boolean;
  receiptTypes: string[];
};

export const CUSTOMS_INVENTORY_RECEIPT_TYPES = ["receipt", "attachment", "warranty", "manual"];
const GROUPINGS: CustomsInventoryGrouping[] = ["location", "entityType", "tag"];
const DEFAULT_MIN_MONTHS = 6;
const MAX_MIN_MONTHS = 60;

export const DEFAULT_CUSTOMS_INVENTORY_PREFS: CustomsInventoryPrefs = Object.freeze({
  header: { title: "", name: "", address: "", phone: "", email: "" },
  referenceDate: "",
  minMonths: DEFAULT_MIN_MONTHS,
  groupBy: "location",
  groupTagIds: [],
  dutyFreeTagIds: [],
  dutiableTagIds: [],
  mergeIdentical: true,
  includeReceipts: true,
  receiptTypes: ["receipt"],
}) as CustomsInventoryPrefs;

const DATE_RE = /^(\d{4})-(\d{2})-(\d{2})$/;

function parseDate(s: string): { y: number; m: number; d: number } | null {
  const match = DATE_RE.exec(s);
  if (!match) {
    return null;
  }
  const [y, m, d] = [Number(match[1]), Number(match[2]), Number(match[3])];
  const date = new Date(Date.UTC(y, m - 1, d));
  if (date.getUTCFullYear() !== y || date.getUTCMonth() !== m - 1 || date.getUTCDate() !== d) {
    return null;
  }
  return { y, m, d };
}

function validMonths(n: unknown): n is number {
  return typeof n === "number" && Number.isInteger(n) && n >= 1 && n <= MAX_MIN_MONTHS;
}

/** Reference date minus `months` months, clamped to the target month's last day (CALC-4). "" for invalid input. */
export function cutoffDate(referenceDate: string, months: number): string {
  const ref = parseDate(referenceDate);
  if (!ref || !validMonths(months)) {
    return "";
  }
  const index = ref.y * 12 + (ref.m - 1) - months;
  const y = Math.floor(index / 12);
  const m = (index % 12) + 1;
  const lastDay = new Date(Date.UTC(y, m, 0)).getUTCDate();
  const d = Math.min(ref.d, lastDay);
  return `${String(y).padStart(4, "0")}-${String(m).padStart(2, "0")}-${String(d).padStart(2, "0")}`;
}

function knownIds(ids: unknown, known: Set<string>): string[] {
  return Array.isArray(ids) ? ids.filter((id): id is string => typeof id === "string" && known.has(id)) : [];
}

function str(v: unknown): string {
  return typeof v === "string" ? v : "";
}

/** Merges saved options with the defaults and drops tag IDs this collection doesn't know (FR-21). */
export function sanitizePrefs(
  p: Partial<CustomsInventoryPrefs> | undefined | null,
  knownTagIds: Set<string>
): CustomsInventoryPrefs {
  const d = DEFAULT_CUSTOMS_INVENTORY_PREFS;
  const saved = p ?? {};
  const header = (saved.header ?? {}) as Partial<CustomsInventoryHeader>;
  const receiptTypes = Array.isArray(saved.receiptTypes)
    ? saved.receiptTypes.filter(t => CUSTOMS_INVENTORY_RECEIPT_TYPES.includes(t))
    : [];

  return {
    header: {
      title: str(header.title),
      name: str(header.name),
      address: str(header.address),
      phone: str(header.phone),
      email: str(header.email),
    },
    referenceDate: parseDate(str(saved.referenceDate)) ? str(saved.referenceDate) : d.referenceDate,
    minMonths: validMonths(saved.minMonths) ? saved.minMonths : d.minMonths,
    groupBy: GROUPINGS.includes(saved.groupBy as CustomsInventoryGrouping) ? saved.groupBy! : d.groupBy,
    groupTagIds: knownIds(saved.groupTagIds, knownTagIds),
    dutyFreeTagIds: knownIds(saved.dutyFreeTagIds, knownTagIds),
    dutiableTagIds: knownIds(saved.dutiableTagIds, knownTagIds),
    mergeIdentical: typeof saved.mergeIdentical === "boolean" ? saved.mergeIdentical : d.mergeIdentical,
    includeReceipts: typeof saved.includeReceipts === "boolean" ? saved.includeReceipts : d.includeReceipts,
    receiptTypes: receiptTypes.length > 0 ? receiptTypes : [...d.receiptTypes],
  };
}

export function buildRequest(a: {
  scope: "selected" | "query";
  itemIds: string[];
  search: CustomsInventorySearchContext | null;
  prefs: CustomsInventoryPrefs;
}): CustomsInventoryRequest {
  const { prefs } = a;
  let scope: CustomsInventoryRequest["scope"];
  if (a.scope === "query") {
    if (!a.search) {
      throw new Error("customs inventory: no search to export");
    }
    scope = { itemIds: [], query: { ...a.search.query } };
  } else {
    scope = { itemIds: [...a.itemIds] };
  }

  return {
    scope,
    options: {
      groupBy: prefs.groupBy as CustomsInventoryGroupBy,
      groupTagIds: [...prefs.groupTagIds],
      dutyFreeTagIds: [...prefs.dutyFreeTagIds],
      dutiableTagIds: [...prefs.dutiableTagIds],
      referenceDate: prefs.referenceDate,
      minMonths: prefs.minMonths,
      mergeIdentical: prefs.mergeIdentical,
      includeReceipts: prefs.includeReceipts,
      receiptTypes: [...prefs.receiptTypes],
    },
    header: { ...prefs.header },
  };
}

export function formatCents(cents: number, currency: string, locale: string): string {
  try {
    return new Intl.NumberFormat(locale, { style: "currency", currency }).format(cents / 100);
  } catch {
    return `${(cents / 100).toFixed(2)} ${currency}`;
  }
}

export function filenameFromDisposition(h: string | null, fallback: string): string {
  const match = h ? /filename="?([^";]+)"?/i.exec(h) : null;
  return match?.[1]?.trim() || fallback;
}
