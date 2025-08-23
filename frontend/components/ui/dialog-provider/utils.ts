import type { ComputedRef } from "vue";
import { createContext } from "reka-ui";
import { useMagicKeys, useActiveElement } from "@vueuse/core";
import type { BarcodeProduct } from "~~/lib/api/types/data-contracts";

export enum DialogID {
  AttachmentEdit = 'attachment-edit',
  ChangePassword = 'changePassword',
  CreateItem = 'create-item',
  CreateLocation = 'create-location',
  CreateLabel = 'create-label',
  CreateNotifier = 'create-notifier',
  DuplicateSettings = 'duplicate-settings',
  DuplicateTemporarySettings = 'duplicate-temporary-settings',
  EditMaintenance = 'edit-maintenance',
  Import = 'import',
  ItemImage = 'item-image',
  ItemTableSettings = 'item-table-settings',
  PrintLabel = 'print-label',
  ProductImport = 'product-import',
  QuickMenu = 'quick-menu',
  Scanner = 'scanner',
  PageQRCode = 'page-qr-code',
  UpdateLabel = 'update-label',
  UpdateLocation = 'update-location',
}

export type DialogParamsMap = {
  [DialogID.CreateItem]: { product?: BarcodeProduct };
  [DialogID.ProductImport]: { barcode?: string };
};

type DialogsWithParams = keyof DialogParamsMap;

type OpenDialog = {
  <T extends DialogID>(dialogId: T, params?: T extends DialogsWithParams ? DialogParamsMap[T] : undefined): void;
};

type OpenCallback = {
  <T extends DialogID>(dialogId: T, callback: (params?: T extends keyof DialogParamsMap ? DialogParamsMap[T] : undefined) => void): void;
}

export const [useDialog, provideDialogContext] = createContext<{
  activeDialog: ComputedRef<DialogID | null>;
  activeAlerts: ComputedRef<string[]>;
  registerOpenDialogCallback: OpenCallback;
  openDialog: OpenDialog;
  closeDialog: (dialogId?: DialogID) => void;
  addAlert: (alertId: string) => void;
  removeAlert: (alertId: string) => void;
}>("DialogProvider");

export const useDialogHotkey = (
  dialogId: DialogID,
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
