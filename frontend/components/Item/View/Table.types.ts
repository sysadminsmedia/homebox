import type { ItemSummary, BarcodeProduct, ItemCreate } from "~~/lib/api/types/data-contracts";

export type TableProperties = keyof ItemSummary | keyof BarcodeProduct | `item.${keyof ItemCreate}`;

export type TableHeaderType = {
  text: string;
  value: TableProperties;
  url?: string;
  sortable?: boolean;
  align?: "left" | "center" | "right";
  enabled: boolean;
  type?: "price" | "boolean" | "name" | "location" | "date";
};

export type TableEmits = {
  (event: "update:selectedItem", value: ItemSummary | BarcodeProduct | null): void;
};

export type TableProps = {
  items: ItemSummary[] | BarcodeProduct[];
  itemType: "barcodeproduct" | "itemsummary";
  selectionMode?: boolean;
  disableControls?: boolean;
};

export type TableData = Record<string, any>;
