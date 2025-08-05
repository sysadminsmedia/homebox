export type StringLike = string | number | boolean;

type BaseDetail = {
  name: string;
  slot?: string;
};

type Barcode = BaseDetail & {
  type: "barcode";
  text: string;
  copyable?: boolean;
};

type DateDetail = BaseDetail & {
  type: "date";
  text: Date | string;
  date: boolean;
};

type CurrencyDetail = BaseDetail & {
  type: "currency";
  text: string;
};

type LinkDetail = BaseDetail & {
  type: "link";
  text: string;
  href: string;
};

type MarkdownDetail = BaseDetail & {
  type: "markdown";
  text: string;
};

export type Detail = BaseDetail & {
  text: StringLike;
  type?: "text";
  copyable?: boolean;
};

export type AnyDetail = Barcode | DateDetail | CurrencyDetail | LinkDetail | MarkdownDetail | Detail;

export type Details = Array<Detail | AnyDetail>;

export function filterZeroValues(details: Details): Details {
  return details.filter(detail => {
    switch (detail.type) {
      case "date":
        return validDate(detail.text);
      case "currency":
        return !!detail.text;
      case "link":
        return !!detail.text && !!detail.href;
      case undefined:
      case "text":
      case "markdown":
        return detail.text !== null && detail.text !== "" && detail.text !== undefined;
      default:
        console.warn("Unknown detail type (this should never happen)", detail);
        return false;
    }
  });
}
