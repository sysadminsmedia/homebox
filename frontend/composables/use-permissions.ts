import { UserRole } from "~~/lib/api/types/data-contracts";

/**
 * usePermissions returns computed helpers based on the authenticated user's role.
 *
 * Role hierarchy (lowest → highest): viewer < user < manager < owner
 *
 * - isOwner:           Only "owner" role — full access including DELETE
 * - isManagerOrAbove:  "manager" or "owner" — can create/edit but not delete
 * - isViewer:          "viewer" role — read-only, cannot create/edit/delete
 */
export function usePermissions() {
  const ctx = useAuthContext();

  const role = computed<UserRole>(() => {
    if (!ctx.user) return UserRole.RoleViewer;
    return (ctx.user.role as UserRole) ?? UserRole.RoleUser;
  });

  const isOwner = computed(() => role.value === UserRole.RoleOwner);

  const isManagerOrAbove = computed(
    () => role.value === UserRole.RoleOwner || role.value === UserRole.RoleManager
  );

  const isViewer = computed(() => role.value === UserRole.RoleViewer);

  return {
    role,
    isOwner,
    isManagerOrAbove,
    isViewer,
  };
}
