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

  const requests = new Requests("", "", {});
  requests.addResponseInterceptor(logger);
  requests.addResponseInterceptor(r => {
    if (r.status === 401) {
      console.error("unauthorized request, invalidating session");
      authCtx.invalidateSession();
      navigateTo("/");
    }
  });

  for (const [_, observer] of Object.entries(observers)) {
    requests.addResponseInterceptor(observer.handler);
  }

  return new UserClient(requests, authCtx.attachmentToken || "");
}
