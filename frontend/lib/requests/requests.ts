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

export class Requests {
  private baseUrl: string;
  private token: () => string;
  private headers: Record<string, string> = {};
  private responseInterceptors: ResponseInterceptor[] = [];

  addResponseInterceptor(interceptor: ResponseInterceptor) {
    this.responseInterceptors.push(interceptor);
  }

  private callResponseInterceptors(response: Response, request?: RequestInit) {
    this.responseInterceptors.forEach(i => i(response, request));
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
