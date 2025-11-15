export type TableHeader = {
  text: string;
  value: string;
  sortable?: boolean;
  align?: "left" | "center" | "right";
};

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type TableData = Record<string, any>;
