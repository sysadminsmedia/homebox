import type { TagOut } from "~~/lib/api/types/data-contracts";

export type TagTreeItem = Omit<TagOut, "children"> & {
  children: TagTreeItem[];
};
