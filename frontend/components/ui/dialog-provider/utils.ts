import { computed, type ComputedRef } from 'vue';
import { createContext } from 'reka-ui';
import { useMagicKeys, useActiveElement } from '@vueuse/core';
import type { BarcodeProduct } from '~~/lib/api/types/data-contracts';

export enum DialogID {
  AttachmentEdit = 'attachment-edit',
  ChangePassword = 'changePassword',
  CreateItem = 'create-item',
  CreateLocation = 'create-location',
  CreateLabel = 'create-label',
  CreateNotifier = 'create-notifier',
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

/**
 * - Keys present without ? => params required
 * - Keys present with ?    => params optional
 * - Keys not present       => no params allowed
 */
export type DialogParamsMap = {
  [DialogID.ItemImage]:
    | ({
        type: 'preloaded';
        originalSrc: string;
        originalType?: string;
        thumbnailSrc?: string;
      }
    | {
        type: 'attachment';
        mimeType: string;
        thumbnailId?: string;
      }) & {
        itemId: string;
        attachmentId: string;
      };
  [DialogID.CreateItem]?: { product?: BarcodeProduct };
  [DialogID.ProductImport]?: { barcode?: string };
};

/**
 * Defines the payload type for a dialog's onClose callback.
 */
export type DialogResultMap = {
  [DialogID.ItemImage]?: { action: 'delete', id: string };
};

/** Helpers to split IDs by requirement */
type OptionalKeys<T> = {
  [K in keyof T]-?: {} extends Pick<T, K> ? K : never;
}[keyof T];

type RequiredKeys<T> = Exclude<keyof T, OptionalKeys<T>>;

type SpecifiedDialogIDs = keyof DialogParamsMap;
export type NoParamDialogIDs = Exclude<DialogID, SpecifiedDialogIDs>;
export type RequiredDialogIDs = RequiredKeys<DialogParamsMap>;
export type OptionalDialogIDs = OptionalKeys<DialogParamsMap>;

type ParamsOf<T extends DialogID> = T extends SpecifiedDialogIDs
  ? DialogParamsMap[T]
  : never;

type ResultOf<T extends DialogID> = T extends keyof DialogResultMap
  ? DialogResultMap[T]
  : void;

type OpenDialog = {
  // Dialogs with no parameters
  <T extends NoParamDialogIDs>(
    dialogId: T,
    options?: { onClose?: (result?: ResultOf<T>) => void; params?: never }
  ): void;
  // Dialogs with required parameters
  <T extends RequiredDialogIDs>(
    dialogId: T,
    options: { params: ParamsOf<T>; onClose?: (result?: ResultOf<T>) => void }
  ): void;
  // Dialogs with optional parameters
  <T extends OptionalDialogIDs>(
    dialogId: T,
    options?: { params?: ParamsOf<T>; onClose?: (result?: ResultOf<T>) => void }
  ): void;
};

type CloseDialog = {
  // Close the currently active dialog, no ID specified. No result payload.
  (): void;
  // Close a specific dialog that has a defined result type.
  <T extends keyof DialogResultMap>(dialogId: T, result?: ResultOf<T>): void;
  // Close a specific dialog that has NO defined result type.
  <T extends Exclude<DialogID, keyof DialogResultMap>>(
    dialogId: T,
    result?: never
  ): void;
};

type OpenCallback = {
  <T extends NoParamDialogIDs>(dialogId: T, cb: () => void): () => void;
  <T extends RequiredDialogIDs>(
    dialogId: T,
    cb: (params: ParamsOf<T>) => void
  ): () => void;
  <T extends OptionalDialogIDs>(
    dialogId: T,
    cb: (params?: ParamsOf<T>) => void
  ): () => void;
};

export const [useDialog, provideDialogContext] = createContext<{
  activeDialog: ComputedRef<DialogID | null>;
  activeAlerts: ComputedRef<string[]>;
  registerOpenDialogCallback: OpenCallback;
  openDialog: OpenDialog;
  closeDialog: CloseDialog;
  addAlert: (alertId: string) => void;
  removeAlert: (alertId: string) => void;
}>('DialogProvider');

/**
 * Hotkey helper:
 * - No/optional params: pass dialogId + key
 * - Required params: pass dialogId + key + getParams()
 */
type HotkeyKey = {
  shift?: boolean;
  ctrl?: boolean;
  code: string;
};

export function useDialogHotkey<T extends NoParamDialogIDs | OptionalDialogIDs>(
  dialogId: T,
  key: HotkeyKey
): void;
export function useDialogHotkey<T extends RequiredDialogIDs>(
  dialogId: T,
  key: HotkeyKey,
  getParams: () => ParamsOf<T>
): void;
export function useDialogHotkey(
  dialogId: DialogID,
  key: HotkeyKey,
  getParams?: () => unknown
) {
  const { openDialog } = useDialog();

  const activeElement = useActiveElement();

  const notUsingInput = computed(
    () =>
      activeElement.value?.tagName !== 'INPUT' &&
      activeElement.value?.tagName !== 'TEXTAREA'
  );

  useMagicKeys({
    passive: false,
    onEventFired: (event) => {
      if (
        notUsingInput.value &&
        event.type === 'keydown' &&
        event.code === key.code &&
        (key.shift === undefined || event.shiftKey === key.shift) &&
        (key.ctrl === undefined || event.ctrlKey === key.ctrl)
      ) {
        if (getParams) {
          openDialog(dialogId as RequiredDialogIDs, {
            params: getParams() as never,
          });
        } else {
          openDialog(dialogId as NoParamDialogIDs);
        }
        event.preventDefault();
      }
    },
  });
}