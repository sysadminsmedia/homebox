import type { CustomsInventorySearchContext } from "~~/lib/customs-inventory";

/** The current search on /search, so the export dialog (mounted outside the page) can offer "all matching". */
export function useCustomsInventorySearch() {
  return useState<CustomsInventorySearchContext | null>("customs-inventory-search", () => null);
}
