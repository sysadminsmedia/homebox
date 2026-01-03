import { BaseAPI, route } from "../base";
import type {
  CurrenciesCurrency,
  Group,
  GroupAcceptInvitationResponse,
  GroupInvitation,
  GroupInvitationCreate,
  GroupUpdate,
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

  currencies() {
    return this.http.get<CurrenciesCurrency[]>({
      url: route("/currencies"),
    });
  }
}
