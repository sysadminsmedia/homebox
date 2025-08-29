/* eslint-disable @typescript-eslint/no-explicit-any */
import type { UseConfirmDialogRevealResult, UseConfirmDialogReturn } from "@vueuse/core";
import type { Ref } from "vue";

type Store = UseConfirmDialogReturn<any, boolean, boolean> & {
  text: Ref<string>;
  setup: boolean;
  open: (text: string) => Promise<UseConfirmDialogRevealResult<boolean, boolean>>;
};

const store: Partial<Store> = {
  text: ref("Are you sure you want to delete this item? "),
  setup: false,
};

/**
 * This function is used to wrap the ModalConfirmation which is a "Singleton" component
 * that is used to confirm actions. It's mounded once on the root of the page and reused
 * for every confirmation action that is required.
 *
 * This is in an experimental phase of development and may have unknown or unexpected side effects.
 */
export function useConfirm(): Store {
  if (!store.setup) {
    store.setup = true;

    const { isRevealed, reveal, confirm, cancel } = useConfirmDialog<any, boolean, boolean>();
    store.isRevealed = isRevealed;
    store.reveal = reveal;
    store.confirm = confirm;
    store.cancel = cancel;
  }

  async function openDialog(msg: string): Promise<UseConfirmDialogRevealResult<boolean, boolean>> {
    if (!store.reveal) {
      throw new Error("reveal is not defined");
    }
    if (!store.text) {
      throw new Error("text is not defined");
    }

    store.text.value = msg;
    return await store.reveal();
  }

  return {
    ...(store as Store),
    open: openDialog,
  };
}
