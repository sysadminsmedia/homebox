/* eslint-disable @typescript-eslint/no-explicit-any */
import type { UseConfirmDialogReturn, UseConfirmDialogRevealResult } from "@vueuse/core";
import type { Ref } from "vue";

type Store = UseConfirmDialogReturn<any, boolean, boolean> & {
  text: Ref<string>;
  href: Ref<string>;
  setup: boolean;
  open: (text: string | { message: string; href?: string }) => Promise<UseConfirmDialogRevealResult<boolean, boolean>>;
};

const store: Partial<Store> = {
  text: ref("Are you sure you want to delete this item? "),
  href: ref(""),
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

  async function openDialog(
    msg: string | { message: string; href?: string }
  ): Promise<UseConfirmDialogRevealResult<boolean, boolean>> {
    if (!store.reveal) {
      throw new Error("reveal is not defined");
    }
    if (!store.text) {
      throw new Error("text is not defined");
    }
    if (store.href === undefined) {
      throw new Error("href is not defined");
    }

    if (typeof msg === "string") {
      store.text.value = msg;
    } else {
      store.text.value = msg.message;
      store.href.value = msg.href ?? "";
    }
    return await store.reveal();
  }

  return {
    ...(store as Store),
    open: openDialog,
  };
}
