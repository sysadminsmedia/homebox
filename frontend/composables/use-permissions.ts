import type { EffectivePermissionsOut } from "~~/lib/api/types/data-contracts";

type PermissionsState = {
  /** Tenant (collection) id the cached permissions belong to. */
  tenantKey: string | null;
  loading: boolean;
  data: EffectivePermissionsOut | null;
};

/**
 * usePermissions exposes the caller's effective permission set for the
 * active tenant, fetched once per tenant from /groups/permissions/self.
 *
 * UI gating only: the backend enforces every permission at the ORM layer.
 * While loading, `can()` returns false so privileged affordances stay hidden
 * rather than flashing.
 */
export function usePermissions() {
  const api = useUserApi();
  const prefs = useViewPreferences();

  const state = useState<PermissionsState>("permissions", () => ({
    tenantKey: null,
    loading: false,
    data: null,
  }));

  const activeTenant = computed(() => prefs.value?.collectionId ?? "default");

  async function fetchPermissions() {
    const key = activeTenant.value;
    state.value.loading = true;
    try {
      const { data, error } = await api.permissions.getSelf();
      if (!error && data) {
        state.value = { tenantKey: key, loading: false, data };
      } else {
        state.value = { tenantKey: key, loading: false, data: null };
      }
    } catch {
      state.value = { tenantKey: key, loading: false, data: null };
    }
  }

  function ensureFresh() {
    if (state.value.loading) {
      return;
    }
    if (state.value.tenantKey !== activeTenant.value) {
      void fetchPermissions();
    }
  }

  // Refetch when the active collection changes.
  watch(activeTenant, () => void fetchPermissions());

  ensureFresh();

  const ready = computed(() => state.value.tenantKey === activeTenant.value && state.value.data !== null);

  function can(permission: string): boolean {
    const d = state.value.data;
    if (!d || state.value.tenantKey !== activeTenant.value) {
      return false;
    }
    if (d.isSuperuser) {
      return true;
    }
    return d.permissions.includes(permission);
  }

  return {
    /** Reactive check for a tenant-wide permission key (e.g. "entity:update"). */
    can,
    canAny: (...permissions: string[]) => permissions.some(can),
    isOwner: computed(() => state.value.data?.isOwner ?? false),
    isSuperuser: computed(() => state.value.data?.isSuperuser ?? false),
    permissions: computed(() => state.value.data?.permissions ?? []),
    ready,
    /** Force a refetch, e.g. after permission management mutations. */
    refresh: fetchPermissions,
  };
}
