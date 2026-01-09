import { BaseAPI, route } from "../base";
import type {
  CurrenciesCurrency,
  Group,
  GroupAcceptInvitationResponse,
  GroupInvitation,
  GroupInvitationCreate,
  GroupMemberAdd,
  GroupUpdate,
  UserSummary,
} from "../types/data-contracts";

export class GroupApi extends BaseAPI {
  /**
   * Create a new invitation for the current group.
   */
  createInvitation(data: GroupInvitationCreate) {
    return this.http.post<GroupInvitationCreate, GroupInvitation>({
      url: route("/groups/invitations"),
      body: data,
    });
  }

  /**
   * Accept an invitation.
   */
  acceptInvitation(id: string) {
    return this.http.post<null, GroupAcceptInvitationResponse>({
      url: route(`/groups/invitations/${id}`),
    });
  }

  /**
   * Get all invitations for the current group.
   */
  getInvitations() {
    return this.http.get<GroupInvitation[]>({
      url: route("/groups/invitations"),
    });
  }

  /**
   * Delete an invitation by ID.
   */
  deleteInvitation(id: string) {
    return this.http.delete<void>({
      url: route(`/groups/invitations/${id}`),
    });
  }

  /**
   * Get all members of the current (or specified) group.
   */
  getMembers(groupId?: string) {
    return this.http.get<UserSummary[]>({
      url: route(`/groups/${groupId || ""}/members`),
    });
  }

  /**
   * Add a user to the current (or specified) group.
   */
  addMember(data: GroupMemberAdd, groupId?: string) {
    return this.http.post<GroupMemberAdd, void>({
      url: route(`/groups/${groupId || ""}/members`),
      body: data,
    });
  }

  /**
   * Remove a user from the current (or specified) group.
   */
  removeMember(userId: string, groupId?: string) {
    return this.http.delete<void>({
      url: route(`/groups/${groupId || ""}/members/${userId}`),
    });
  }

  /**
   * Update a user's role in the current (or specified) group.
   */
  update(data: GroupUpdate, groupId?: string) {
    return this.http.put<GroupUpdate, Group>({
      url: route(`/groups/${groupId || ""}`),
      body: data,
    });
  }

  /**
   * Get a group by ID, if no ID is provided, get the current group.
   */
  get(groupId?: string) {
    return this.http.get<Group>({
      url: route(`/groups/${groupId || ""}`),
    });
  }

  /**
   * Create a new group with the given name.
   */
  create(name: string) {
    return this.http.post<
      {
        name: string;
      },
      Group
    >({
      url: route("/groups"),
      body: { name },
    });
  }

  /**
   * Delete a group by ID.
   */
  delete(groupId: string) {
    return this.http.delete<void>({
      url: route(`/groups/${groupId}`),
    });
  }

  /**
   * Get all currencies.
   */
  currencies() {
    return this.http.get<CurrenciesCurrency[]>({
      url: route("/currencies"),
    });
  }
}
