import { BaseAPI, route } from "./base";
import type {
  APISummary,
  ForgotPasswordRequest,
  LoginForm,
  ResetPasswordRequest,
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
    return this.http.post<UserRegistration, TokenResponse>({ url: route("/users/register"), body });
  }

  public forgotPassword(email: string) {
    return this.http.post<ForgotPasswordRequest, void>({
      url: route("/users/forgot-password"),
      body: { email },
    });
  }

  public resetPassword(token: string, password: string) {
    return this.http.post<ResetPasswordRequest, void>({
      url: route("/users/reset-password"),
      body: { token, password },
    });
  }
}
