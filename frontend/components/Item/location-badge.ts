/**
 * The descendant-location API includes locationPath on its rows. Hiding the
 * path in the location view sets its value to undefined but keeps the
 * property so the UI knows this is a scoped physical-location view.
 *
 * A descendant item's logical parent may belong to another room, so that
 * parent must not be labelled or linked as its physical storage location.
 */
export function isScopedLocationItem(item: { locationPath?: string }): boolean {
  return Object.prototype.hasOwnProperty.call(item, "locationPath");
}

export function shouldShowParentLocationBadge(item: { parent?: unknown; locationPath?: string }): boolean {
  return Boolean(item.parent) && !isScopedLocationItem(item);
}
