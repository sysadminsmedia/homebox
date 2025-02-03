import type { ComputedRef } from "vue";
import { createContext } from "radix-vue";

export const [useDialog, provideDialogContext] = createContext<{
  activeDialog: ComputedRef<string | null>;
  openDialog: (dialogId: string) => void;
  closeDialog: (dialogId?: string) => void;
}>("DialogProvider");
