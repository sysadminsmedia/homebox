import { BaseAPI, route } from "./base";
import type { APISummary, LoginForm, TokenResponse, UserRegistration } from "./types/data-contracts";

export type StatusResult = {
  health: boolean;
  versions: string[];
  title: string;
  message: string;
};

export type PublicFoundEntity = {
  assetId: string;
  contactName?: string;
  contactEmail?: string;
  multipleMatches?: boolean;
};

export class PublicApi extends BaseAPI {
  public status() {
    return this.http.get<APISummary>({ url: route("/status") });
  }

  public foundItem(id: string) {
    return this.http.get<PublicFoundEntity>({
      url: route(`/public/found/item/${id}`),
    });
  }

  public foundAsset(id: string) {
    return this.http.get<PublicFoundEntity>({
      url: route(`/public/found/asset/${id}`),
    });
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
}
