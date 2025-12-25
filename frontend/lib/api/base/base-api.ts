/* eslint-disable @typescript-eslint/no-explicit-any */
import type { Requests } from "../../requests";

const ZERO_DATE = "0001-01-01T00:00:00Z";

type BaseApiType = {
  createdAt: string;
  updatedAt: string;

  [key: string]: any;
};

export function hasKey(obj: Record<string, any>, key: string): obj is Required<BaseApiType> {
  return key in obj ? typeof obj[key] === "string" : false;
}

export function parseDate<T>(obj: T, keys: Array<keyof T> = []): T {
  const result = { ...obj };
  [...keys, "createdAt", "updatedAt"].forEach(key => {
    // @ts-expect-error - TS doesn't know that we're checking for the key above
    if (hasKey(result, key)) {
      const value = result[key] as string;

      if (value === undefined || value === "" || value.startsWith(ZERO_DATE)) {
        const dt = new Date();
        dt.setFullYear(1);

        result[key] = dt;
        return;
      }

      // Possible Formats
      // Date Only: YYYY-MM-DD
      // Timestamp: 0001-01-01T00:00:00Z

      // Parse timestamps with default date
      if (value.includes("T")) {
        result[key] = new Date(value);
        return;
      }

      // Parse dates with default time
      const split = value.split("-");

      if (split.length !== 3) {
        console.log(`Invalid date format: ${value}`);
        throw new Error(`Invalid date format: ${value}`);
      }

      const [year, month, day] = split as [string, string, string];

      const dt = new Date();

      dt.setFullYear(parseInt(year, 10));
      dt.setMonth(parseInt(month, 10) - 1);
      dt.setDate(parseInt(day, 10));

      result[key] = dt;
    }
  });

  return result;
}

export class BaseAPI {
  http: Requests;
  attachmentToken: string;

  constructor(requests: Requests, attachmentToken = "") {
    this.http = requests;
    this.attachmentToken = attachmentToken;
  }

  // if an attachmentToken is present, it will be added to URL as a query param
  authURL(url: string): string {
    if (this.attachmentToken) {
      const separator = url.includes("?") ? "&" : "?";
      return `${url}${separator}access_token=${encodeURIComponent(this.attachmentToken)}`;
    }
    return url;
  }

  /**
   * dropFields will remove any fields that are specified in the fields array
   * additionally, it will remove the `createdAt` and `updatedAt` fields if they
   * are present. This is useful for when you want to send a subset of fields to
   * the server like when performing an update.
   */
  protected dropFields<T>(obj: T, keys: Array<keyof T> = []): T {
    const result = { ...obj };
    [...keys, "createdAt", "updatedAt"].forEach(key => {
      // @ts-expect-error - TS doesn't know that we're checking for the key above
      if (hasKey(result, key)) {
        // eslint-disable-next-line @typescript-eslint/no-dynamic-delete
        delete result[key];
      }
    });
    return result;
  }
}
