// The order-by values the batch print page offers, in the order it lists them.
export const BATCH_ORDER_FIELDS = [
  "assetId",
  "name",
  "location",
  "createdAt",
  "updatedAt",
  "quantity",
  "purchasePrice",
  "insured",
] as const;

// True for an exact asset ID term such as "#000-123", not for a comparison or text.
export function isExactAssetIdSearch(q: string): boolean {
  return /^#[0-9-]+$/.test(q);
}

// Selected items in the list take the list's order. Selected items outside it stay
// selected, after them, in their old order.
export function orderSelection(selectedIds: string[], resultIds: string[]): string[] {
  const selected = new Set(selectedIds);
  const listed = new Set(resultIds);
  return [...resultIds.filter(id => selected.has(id)), ...selectedIds.filter(id => !listed.has(id))];
}
