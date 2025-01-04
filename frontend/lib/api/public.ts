import { ParsedAuth } from "ufo";
import { BaseAPI, route } from "./base";
import type { APISummary, LoginForm, OAuthForm, TokenResponse, UserRegistration } from "./types/data-contracts";

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

  public loginOauth(provider: string, iss: string, code: string, state: string | null = null) {
    return this.http.post<OAuthForm, TokenResponse>({
      url: route("/users/login", { provider }),
      body: {
        iss,
        code,
        state,
      },
    });
  }

  public register(body: UserRegistration) {
    return this.http.post<UserRegistration, TokenResponse>({ url: route("/users/register"), body });
  }
}
