import { BaseAPI, route } from "../base";
import type {
  CurrenciesCurrency,
  Group,
  GroupAcceptInvitationResponse,
  GroupInvitation,
  GroupInvitationCreate,
  GroupMemberAdd,
  GroupUpdate,
  UserUpdate,
} from "../types/data-contracts";

export class GroupApi extends BaseAPI {
  createInvitation(data: GroupInvitationCreate) {
    return this.http.post<GroupInvitationCreate, GroupInvitation>({
      url: route("/groups/invitations"),
      body: data,
    });
  }

  acceptInvitation(id: string) {
    return this.http.post<null, GroupAcceptInvitationResponse>({
      url: route(`/groups/invitations/${id}`),
    });
  }

  getInvitations() {
    return this.http.get<GroupInvitation[]>({
      url: route("/groups/invitations"),
    });
  }

  deleteInvitation(id: string) {
    return this.http.delete<void>({
      url: route(`/groups/invitations/${id}`),
    });
  }

  /**
   * Get all members of the current (or specified) group.
   */
  getMembers(groupId?: string) {
    return this.http.get<UserUpdate[]>({
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

  update(data: GroupUpdate, groupId?: string) {
    return this.http.put<GroupUpdate, Group>({
      url: route(`/groups/${groupId || ""}`),
      body: data,
    });
  }

  get(groupId?: string) {
    return this.http.get<Group>({
      url: route(`/groups/${groupId || ""}`),
    });
  }

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

  delete(groupId: string) {
    return this.http.delete<void>({
      url: route(`/groups/${groupId}`),
    });
  }

  currencies() {
    return this.http.get<CurrenciesCurrency[]>({
      url: route("/currencies"),
    });
  }
}
