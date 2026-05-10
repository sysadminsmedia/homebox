import { BaseAPI, route } from "./base";
import type {
  APISummary,
  FoundEntityContact,
  LoginForm,
  TokenResponse,
  UserRegistration,
} from "./types/data-contracts";

export type StatusResult = {
  health: boolean;
  versions: string[];
  title: string;
  message: string;
};

export class PublicApi extends BaseAPI {
  public status() {
    return this.http.get<APISummary>({ url: route("/status") });
  }

  public login(username: string, password: string, stayLoggedIn = false) {
    return this.http.post<LoginForm, TokenResponse>({
      url: route("/users/login"),
      body: {
        username,
        password,
        stayLoggedIn,
      },
    });
  }

  public register(body: UserRegistration) {
    return this.http.post<UserRegistration, TokenResponse>({
      url: route("/users/register"),
      body,
    });
  }

  public foundEntityContact(id: string) {
    return this.http.get<FoundEntityContact>({
      url: route(`/found/entities/${id}`),
    });
  }

  public foundAssetContact(id: string) {
    return this.http.get<FoundEntityContact>({
      url: route(`/found/assets/${id}`),
    });
  }
}
