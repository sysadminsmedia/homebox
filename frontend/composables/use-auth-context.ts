import type { CookieRef } from "nuxt/app";
import type { PublicApi } from "~~/lib/api/public";
import type { UserOut } from "~~/lib/api/types/data-contracts";
import type { UserClient } from "~~/lib/api/user";

export interface IAuthContext {
  get token(): boolean | null;
  get attachmentToken(): string | null;

  /**
   * The current user object for the session. This is undefined if the session is not authorized.
   */
  user?: UserOut;

  /**
   * Returns true if the session is authorized.
   */
  isAuthorized(): boolean;

  /**
   * Invalidates the session by removing the token and the expiresAt.
   */
  invalidateSession(): void;

  /**
   * Logs out the user and calls the invalidateSession method.
   */
  logout(api: UserClient): ReturnType<UserClient["user"]["logout"]>;

  /**
   * Logs in the user and sets the authorization context via cookies
   */
  login(api: PublicApi, email: string, password: string, stayLoggedIn: boolean): ReturnType<PublicApi["login"]>;
}

class AuthContext implements IAuthContext {
  private static _instance?: AuthContext;

  private static readonly cookieTokenKey = "hb.auth.session";
  private static readonly cookieAttachmentTokenKey = "hb.auth.attachment_token";

  user?: UserOut;
  private _token: CookieRef<string | null>;
  private _attachmentToken: CookieRef<string | null>;

  get token() {
    // @ts-ignore sometimes it's a boolean I guess?
    return this._token.value === "true" || this._token.value === true;
  }

  get attachmentToken() {
    return this._attachmentToken.value;
  }

  private constructor(token: string, attachmentToken: string) {
    this._token = useCookie(token);
    this._attachmentToken = useCookie(attachmentToken);
  }

  static get instance() {
    if (!this._instance) {
      this._instance = new AuthContext(AuthContext.cookieTokenKey, AuthContext.cookieAttachmentTokenKey);
    }

    return this._instance;
  }

  isExpired() {
    return !this.token;
  }

  isAuthorized() {
    console.debug("isAuthorized", this.token);
    return this.token;
  }

  invalidateSession() {
    this.user = undefined;

    // Delete the cookies
    this._token.value = null;
    this._attachmentToken.value = null;
    console.log("Session invalidated");
  }

  async login(api: PublicApi, email: string, password: string, stayLoggedIn: boolean) {
    const r = await api.login(email, password, stayLoggedIn);

    if (!r.error) {
      const expiresAt = new Date(r.data.expiresAt);
      this._token = useCookie(AuthContext.cookieTokenKey);
      this._attachmentToken = useCookie(AuthContext.cookieAttachmentTokenKey, {
        expires: expiresAt,
      });
      this._attachmentToken.value = r.data.attachmentToken;
    }

    return r;
  }

  async logout(api: UserClient) {
    const r = await api.user.logout();

    if (!r.error) {
      this.invalidateSession();
    }

    return r;
  }
}

export function useAuthContext(): IAuthContext {
  return AuthContext.instance;
}
