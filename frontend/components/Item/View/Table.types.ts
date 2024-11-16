import type { ItemSummary } from "~~/lib/api/types/data-contracts";

export type TableHeader = {
  text: string;
  value: keyof ItemSummary;
  sortable?: boolean;
  align?: "left" | "center" | "right";
  enabled: boolean;
  type?: "price" | "boolean" | "name" | "location" | "date";
};

export type TableData = Record<string, any>;
