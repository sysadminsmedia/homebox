import type { ComputedRef } from "vue";
import { createContext } from "radix-vue";
import { useMagicKeys, useActiveElement } from "@vueuse/core";

export const [useDialog, provideDialogContext] = createContext<{
  activeDialog: ComputedRef<string | null>;
  activeAlerts: ComputedRef<string[]>;
  openDialog: (dialogId: string) => void;
  closeDialog: (dialogId?: string) => void;
  addAlert: (alertId: string) => void;
  removeAlert: (alertId: string) => void;
}>("DialogProvider");

export const useDialogHotkey = (
  dialogId: string,
  key: {
    shift?: boolean;
    ctrl?: boolean;
    code: string;
  }
) => {
  const { openDialog } = useDialog();

  const activeElement = useActiveElement();

  const notUsingInput = computed(
    () => activeElement.value?.tagName !== "INPUT" && activeElement.value?.tagName !== "TEXTAREA"
  );

  useMagicKeys({
    passive: false,
    onEventFired: event => {
      // console.log({
      //   event,
      //   notUsingInput: notUsingInput.value,
      //   eventType: event.type,
      //   keyCode: event.code,
      //   matchingKeyCode: key.code === event.code,
      //   shift: event.shiftKey,
      //   matchingShift: key.shift === undefined || event.shiftKey === key.shift,
      //   ctrl: event.ctrlKey,
      //   matchingCtrl: key.ctrl === undefined || event.ctrlKey === key.ctrl,
      // });
      if (
        notUsingInput.value &&
        event.type === "keydown" &&
        event.code === key.code &&
        (key.shift === undefined || event.shiftKey === key.shift) &&
        (key.ctrl === undefined || event.ctrlKey === key.ctrl)
      ) {
        openDialog(dialogId);
        event.preventDefault();
      }
    },
  });
};
