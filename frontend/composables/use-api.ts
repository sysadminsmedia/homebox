import { PublicApi } from "~~/lib/api/public";
import { UserClient } from "~~/lib/api/user";
import { Requests } from "~~/lib/requests";

export type Observer = {
  handler: (r: Response, req?: RequestInit) => void;
};

export type RemoveObserver = () => void;

const observers: Record<string, Observer> = {};

export function defineObserver(key: string, observer: Observer): RemoveObserver {
  observers[key] = observer;

  return () => {
    // eslint-disable-next-line @typescript-eslint/no-dynamic-delete
    delete observers[key];
  };
}

function logger(r: Response) {
  console.log(`${r.status}   ${r.url}   ${r.statusText}`);
}

export function usePublicApi(): PublicApi {
  const requests = new Requests("", "", {});
  return new PublicApi(requests);
}

export function useUserApi(): UserClient {
  const authCtx = useAuthContext();
  const prefs = useViewPreferences();

  const headers: Record<string, string> = {};
  if (prefs?.value?.collectionId) {
    headers["X-Tenant"] = prefs.value.collectionId;
  }

  const requests = new Requests("", "", headers);
  requests.addResponseInterceptor(logger);
  requests.addResponseInterceptor(async r => {
    if (r.status === 401) {
      console.error("unauthorized request, invalidating session");
      authCtx.invalidateSession();
      navigateTo("/");
    }

    if (r.status === 403) {
      try {
        const contentType = r.headers.get("Content-Type") ?? "";
        if (!contentType.startsWith("application/json")) {
          return;
        }

        const body = (await r.json().catch(() => null)) as { error?: string } | null;

        if (body?.error === "user does not have access to the requested tenant") {
          console.log("user does not have access to the requested tenant");
          if (window.location.pathname == "/") {
            // do nothing
            console.log("at root path, ignoring collectionId to prevent infinite redirect loop");
          } else if (!prefs?.value?.collectionId) {
            console.log("no collectionId set, ignoring");
          } else {
            console.log("clearing collectionId");
            prefs.value.collectionId = null;
          }
        }
      } catch {
        // ignore parsing errors to avoid breaking the interceptor chain
        console.log("failed to parse 403 response body");
      }
    }
  });

  for (const [_, observer] of Object.entries(observers)) {
    requests.addResponseInterceptor(observer.handler);
  }

  return new UserClient(requests, authCtx.attachmentToken || "");
}
