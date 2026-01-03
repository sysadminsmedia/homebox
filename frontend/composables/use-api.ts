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

    if (r.status === 403 && (await r.json()).error === "user does not have access to the requested tenant") {
      prefs.value.collectionId = null;
    }
  });

  for (const [_, observer] of Object.entries(observers)) {
    requests.addResponseInterceptor(observer.handler);
  }

  return new UserClient(requests, authCtx.attachmentToken || "");
}
