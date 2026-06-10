export enum Method {
  GET = "GET",
  POST = "POST",
  PUT = "PUT",
  DELETE = "DELETE",
  PATCH = "PATCH",
}

export type ResponseInterceptor = (r: Response, rq?: RequestInit) => void;

export interface TResponse<T> {
  status: number;
  error: boolean;
  data: T;
  response: Response;
}

export type RequestArgs<T> = {
  url: string;
  body?: T;
  data?: FormData;
  headers?: Record<string, string>;
};

export type ProgressUploadArgs = {
  url: string;
  data: FormData;
  headers?: Record<string, string>;
  /** Called with the fraction (0..1) of the body uploaded. The final call is always 1. */
  onProgress?: (fraction: number) => void;
};

export class Requests {
  private baseUrl: string;
  private token: () => string;
  private headers: Record<string, string> = {};
  private responseInterceptors: ResponseInterceptor[] = [];

  addResponseInterceptor(interceptor: ResponseInterceptor) {
    this.responseInterceptors.push(interceptor);
  }

  private callResponseInterceptors(response: Response, request?: RequestInit) {
    this.responseInterceptors.forEach(i => i(response.clone(), request));
  }

  private url(rest: string): string {
    return this.baseUrl + rest;
  }

  constructor(baseUrl: string, token: string | (() => string) = "", headers: Record<string, string> = {}) {
    this.baseUrl = baseUrl;
    this.token = typeof token === "string" ? () => token : token;
    this.headers = headers;
  }

  public get<T>(args: RequestArgs<T>): Promise<TResponse<T>> {
    return this.do<T>(Method.GET, args);
  }

  public post<T, U>(args: RequestArgs<T>): Promise<TResponse<U>> {
    return this.do<U>(Method.POST, args);
  }

  public put<T, U>(args: RequestArgs<T>): Promise<TResponse<U>> {
    return this.do<U>(Method.PUT, args);
  }

  public patch<T, U>(args: RequestArgs<T>): Promise<TResponse<U>> {
    return this.do<U>(Method.PATCH, args);
  }

  public delete<T>(args: RequestArgs<T>): Promise<TResponse<T>> {
    return this.do<T>(Method.DELETE, args);
  }

  /**
   * POST a FormData body with real upload progress via XHR. Mirrors `post`'s `TResponse<U>`
   * shape so call sites can opt in to progress without restructuring their result handling.
   * Resolves on every HTTP status (the `error` flag distinguishes 2xx from 4xx/5xx); rejects
   * only when the transport itself fails (network drop, abort) — same as `post()`.
   */
  public async postWithProgress<U>(args: ProgressUploadArgs): Promise<TResponse<U>> {
    const { xhrUpload } = await import("./xhr-upload");
    const headers: Record<string, string> = { ...args.headers, ...this.headers };
    const token = this.token();
    if (token !== "") headers["Authorization"] = token;

    const result = await xhrUpload<U>({
      url: this.url(args.url),
      data: args.data,
      headers,
      onProgress: args.onProgress,
    });
    this.callResponseInterceptors(result.response);
    return result;
  }

  private methodSupportsBody(method: Method): boolean {
    return method === Method.POST || method === Method.PUT || method === Method.PATCH;
  }

  private async do<T>(method: Method, rargs: RequestArgs<unknown>): Promise<TResponse<T>> {
    const payload: RequestInit = {
      method,
      headers: {
        ...rargs.headers,
        ...this.headers,
      } as Record<string, string>,
    };

    const token = this.token();
    if (token !== "" && payload.headers !== undefined) {
      // @ts-expect-error - we know that the header is there
      payload.headers["Authorization"] = token;
    }

    if (this.methodSupportsBody(method)) {
      if (rargs.data) {
        payload.body = rargs.data;
      } else {
        // @ts-expect-error - we know that the header is there
        payload.headers["Content-Type"] = "application/json";
        payload.body = JSON.stringify(rargs.body);
      }
    }

    const response = await fetch(this.url(rargs.url), payload);
    this.callResponseInterceptors(response, payload);

    const data: T = await (async () => {
      if (response.status === 204) {
        return {} as T;
      }

      if (response.headers.get("Content-Type")?.startsWith("application/json")) {
        try {
          return await response.json();
        } catch (e) {
          return {} as T;
        }
      }

      return response.body as unknown as T;
    })();

    return {
      status: response.status,
      error: !response.ok,
      data,
      response,
    };
  }
}
