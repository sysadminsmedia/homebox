const parts = {
  base: "/",
  host: "http://localhost.com",
  prefix: "/api/v1",
};

export function overrideParts(host: string, prefix: string, base: string = "/") {
  parts.base = base;
  parts.host = host;
  parts.prefix = prefix;
}

export type QueryValue = string | string[] | number | number[] | boolean | null | undefined;

/**
 * route is the main URL builder for the API. It will use a predefined host and prefix (global)
 * in the urls.ts file and then append the passed-in path parameter using the `URL` class from the
 * browser. It will also append any query parameters passed in as the second parameter.
 *
 * The default host `http://localhost.com` is removed from the path if it is present. This allows us
 * to bootstrap the API with different hosts as needed (like for testing) but still allows us to use
 * relative URLs in production because the API and client bundle are served from the same server/host.
 */
export function route(rest: string, params: Record<string, QueryValue> = {}): string {
  const base = parts.base.replace(/\/$/, "");
  const url = new URL(base + parts.prefix + rest, parts.host);

  for (const [key, value] of Object.entries(params)) {
    if (Array.isArray(value)) {
      for (const item of value) {
        url.searchParams.append(key, String(item));
      }
    } else {
      url.searchParams.append(key, String(value));
    }
  }

  return url.toString().replace("http://localhost.com", "");
}
