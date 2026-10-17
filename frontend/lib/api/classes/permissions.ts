import { BaseAPI, route } from "../base";
import type {
  AccessGrantCreate,
  AccessGrantOut,
  EffectivePermissionsOut,
  MemberPermissions,
  PermissionDefinition,
  PermissionGroupCreate,
  PermissionGroupOut,
  PermissionGroupUpdate,
  AccessGrantUpdate,
  MemberPermissionsSet,
  PermissionGroupSetMembers,
} from "../types/data-contracts";

export class PermissionsApi extends BaseAPI {
  /**
   * Static catalog of all permission keys (resource x action matrix).
   */
  getCatalog() {
    return this.http.get<PermissionDefinition[]>({
      url: route("/permissions/catalog"),
    });
  }

  /**
   * The caller's effective permissions for the active tenant.
   */
  getSelf() {
    return this.http.get<EffectivePermissionsOut>({
      url: route("/groups/permissions/self"),
    });
  }

  getPermissionGroups() {
    return this.http.get<PermissionGroupOut[]>({
      url: route("/groups/permission-groups"),
    });
  }

  getPermissionGroup(id: string) {
    return this.http.get<PermissionGroupOut>({
      url: route(`/groups/permission-groups/${id}`),
    });
  }

  createPermissionGroup(data: PermissionGroupCreate) {
    return this.http.post<PermissionGroupCreate, PermissionGroupOut>({
      url: route("/groups/permission-groups"),
      body: data,
    });
  }

  updatePermissionGroup(id: string, data: PermissionGroupUpdate) {
    return this.http.put<PermissionGroupUpdate, PermissionGroupOut>({
      url: route(`/groups/permission-groups/${id}`),
      body: data,
    });
  }

  deletePermissionGroup(id: string) {
    return this.http.delete<void>({
      url: route(`/groups/permission-groups/${id}`),
    });
  }

  /**
   * Replace the member list of a permission group.
   */
  setPermissionGroupMembers(id: string, userIds: string[]) {
    return this.http.put<PermissionGroupSetMembers, PermissionGroupOut>({
      url: route(`/groups/permission-groups/${id}/members`),
      body: { userIds },
    });
  }

  /**
   * A member's direct permissions, permission groups, and effective set.
   */
  getMemberPermissions(userId: string) {
    return this.http.get<MemberPermissions>({
      url: route(`/groups/members/${userId}/permissions`),
    });
  }

  setMemberPermissions(userId: string, permissions: string[]) {
    return this.http.put<MemberPermissionsSet, MemberPermissions>({
      url: route(`/groups/members/${userId}/permissions`),
      body: { permissions },
    });
  }

  /**
   * Row-level access grants on one entity.
   */
  getEntityGrants(entityId: string) {
    return this.http.get<AccessGrantOut[]>({
      url: route(`/entities/${entityId}/permissions`),
    });
  }

  createEntityGrant(entityId: string, data: AccessGrantCreate) {
    return this.http.post<AccessGrantCreate, AccessGrantOut>({
      url: route(`/entities/${entityId}/permissions`),
      body: data,
    });
  }

  updateEntityGrant(entityId: string, grantId: string, actions: string[]) {
    return this.http.put<AccessGrantUpdate, AccessGrantOut>({
      url: route(`/entities/${entityId}/permissions/${grantId}`),
      body: { actions },
    });
  }

  deleteEntityGrant(entityId: string, grantId: string) {
    return this.http.delete<void>({
      url: route(`/entities/${entityId}/permissions/${grantId}`),
    });
  }
}
