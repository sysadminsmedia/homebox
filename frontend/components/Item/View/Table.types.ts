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

export const ItemSummaryHeaders = [
  { text: "items.asset_id", value: "assetId", enabled: false },
  {
    text: "items.name",
    value: "name",
    enabled: true,
    type: "name",
  },
  { text: "items.quantity", value: "quantity", align: "center", enabled: true },
  { text: "items.insured", value: "insured", align: "center", enabled: true, type: "boolean" },
  { text: "items.purchase_price", value: "purchasePrice", align: "center", enabled: true, type: "price" },
  { text: "items.location", value: "location", align: "center", enabled: false, type: "location" },
  { text: "items.archived", value: "archived", align: "center", enabled: false, type: "boolean" },
  { text: "items.created_at", value: "createdAt", align: "center", enabled: false, type: "date" },
  { text: "items.updated_at", value: "updatedAt", align: "center", enabled: false, type: "date" },
] satisfies TableHeaderType[];

export const BarcodeProductHeaders = [
  {
    text: "items.name",
    value: "item.name",
    enabled: true,
    align: "center",
    type: "name",
  },
  { text: "items.manufacturer", value: "item.manufacturer", align: "center", enabled: true },
  { text: "items.model_number", value: "item.modelNumber", align: "center", enabled: true },
  {
    text: "components.item.product_import.db_source",
    value: "search_engine_name",
    url: "search_engine_product_url",
    align: "center",
    enabled: true,
  },
] satisfies TableHeaderType[];

export type TableEmits = {
  (event: "update:selectedItem", value: ItemSummary | BarcodeProduct | null): void;
};

export type TableProps = {
  items: ItemSummary[] | BarcodeProduct[];
  defaultTableHeaders: TableHeaderType[];
  selectionMode?: boolean;
  disableControls?: boolean;
};

export type TableData = Record<string, any>;
