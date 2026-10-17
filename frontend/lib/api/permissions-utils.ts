import type { PermissionDefinition } from "~~/lib/api/types/data-contracts";

/**
 * Stored permission lists support wildcards, mirroring the backend:
 * "*" covers every permission (present and future), "<resource>:*" covers
 * every action on one resource. The UI always *displays* concrete catalog
 * keys (expandPermissions) and *persists* the wildcard whenever the whole
 * catalog is selected (normalizePermissions), so full-access assignments
 * stay valid when new permissions are added to the catalog later.
 */
export const PERMISSION_WILDCARD = "*";

function covers(stored: string[], key: string): boolean {
  if (stored.includes(PERMISSION_WILDCARD) || stored.includes(key)) {
    return true;
  }
  const resource = key.split(":")[0];
  return stored.includes(`${resource}:*`);
}

/** Expand a stored list (which may contain wildcards) into concrete catalog keys. */
export function expandPermissions(stored: string[], catalog: PermissionDefinition[]): string[] {
  return catalog.filter(def => covers(stored, def.key)).map(def => def.key);
}

/**
 * Normalize a checkbox selection for storage: when every catalog key is
 * selected, persist ["*"] instead of an enumerated snapshot.
 */
export function normalizePermissions(selected: string[], catalog: PermissionDefinition[]): string[] {
  if (catalog.length > 0 && catalog.every(def => selected.includes(def.key))) {
    return [PERMISSION_WILDCARD];
  }
  // Keep only known keys, in catalog order.
  return catalog.filter(def => selected.includes(def.key)).map(def => def.key);
}
