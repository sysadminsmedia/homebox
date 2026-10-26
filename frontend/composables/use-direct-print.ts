import { useI18n } from "vue-i18n";
import type { LabelTemplatePrintResponse, LabelPrintItem, LabelPrintLocation } from "~~/lib/api/types/data-contracts";

/**
 * Whether the server has an admin-configured label print command
 * (HBOX_LABEL_MAKER_PRINT_COMMAND). Reads the same status the basic label uses.
 */
export function useServerLabelPrinting() {
  const pubApi = usePublicApi();

  const { data: status, pending } = useAsyncData("server-label-printing-status", async () => {
    const { data, error } = await pubApi.status();
    if (error) return null;
    return data;
  });

  const labelPrinting = computed(() => status.value?.labelPrinting === true);

  return { labelPrinting, pending };
}

export function useDirectPrint() {
  const api = useUserApi();
  const { t } = useI18n();

  /**
   * Print labels with optional per-item quantities
   * @param templateId - Template to use
   * @param itemsOrIds - Either array of item IDs (backward compatible) or items with quantities
   * @param defaultCopies - Default copies per label if not specified per-item
   */
  async function printLabels(
    templateId: string,
    itemsOrIds: string[] | LabelPrintItem[],
    defaultCopies: number = 1
  ): Promise<LabelTemplatePrintResponse | null> {
    // Check if we have items with quantities or just IDs
    const hasQuantities = itemsOrIds.length > 0 && typeof itemsOrIds[0] === "object";

    const { data, error } = await api.labelTemplates.print(templateId, {
      itemIds: hasQuantities ? [] : (itemsOrIds as string[]),
      items: hasQuantities ? (itemsOrIds as LabelPrintItem[]) : [],
      copies: defaultCopies,
    });
    if (error) {
      throw new Error(t("components.label_template.toast.print_failed"));
    }
    return data;
  }

  /**
   * Print location labels with optional per-location quantities
   * @param templateId - Template to use
   * @param locationsOrIds - Either array of location IDs or locations with quantities
   * @param defaultCopies - Default copies per label if not specified per-location
   */
  async function printLocationLabels(
    templateId: string,
    locationsOrIds: string[] | LabelPrintLocation[],
    defaultCopies: number = 1
  ): Promise<LabelTemplatePrintResponse | null> {
    // Check if we have locations with quantities or just IDs
    const hasQuantities = locationsOrIds.length > 0 && typeof locationsOrIds[0] === "object";

    const { data, error } = await api.labelTemplates.printLocations(templateId, {
      locationIds: hasQuantities ? [] : (locationsOrIds as string[]),
      locations: hasQuantities ? (locationsOrIds as LabelPrintLocation[]) : [],
      copies: defaultCopies,
    });
    if (error) {
      throw new Error(t("components.label_template.toast.print_failed"));
    }
    return data;
  }

  return {
    printLabels,
    printLocationLabels,
  };
}

/**
 * Composable for quick printing with the default template, sent to the
 * server-configured label printer. Available when a default template is chosen
 * and the server has a print command configured.
 */
export function useQuickPrint() {
  const preferences = useViewPreferences();
  const { t } = useI18n();
  const { printLabels, printLocationLabels } = useDirectPrint();
  const { labelPrinting } = useServerLabelPrinting();
  const { templates } = useLabelTemplates();
  const { find: findPreset } = useLabelPresetLookup();

  // Get the default template ID from user preferences
  const defaultTemplateId = computed(() => preferences.value.defaultTemplateId);

  // The server prints one label at a time, so a sheet template (Avery and the
  // like) cannot be quick printed; the print dialog offers PDF instead.
  const defaultIsSheet = computed(() => {
    const tpl = templates.value?.find(item => item.id === defaultTemplateId.value);
    return findPreset(tpl?.preset)?.kind === "sheet";
  });

  const isQuickPrintAvailable = computed(
    () => labelPrinting.value && !!defaultTemplateId.value && !defaultIsSheet.value
  );

  /**
   * Quick print items using the default template.
   * @returns true if print was initiated, false if quick print is unavailable
   */
  async function quickPrintItems(itemIds: string[]): Promise<boolean> {
    if (!isQuickPrintAvailable.value || !defaultTemplateId.value) {
      return false;
    }

    const result = await printLabels(defaultTemplateId.value, itemIds);
    if (!result?.success) {
      throw new Error(result?.message || t("components.label_template.toast.print_failed"));
    }
    return true;
  }

  /**
   * Quick print locations using the default template.
   * @returns true if print was initiated, false if quick print is unavailable
   */
  async function quickPrintLocations(locationIds: string[]): Promise<boolean> {
    if (!isQuickPrintAvailable.value || !defaultTemplateId.value) {
      return false;
    }

    const result = await printLocationLabels(defaultTemplateId.value, locationIds);
    if (!result?.success) {
      throw new Error(result?.message || t("components.label_template.toast.print_failed"));
    }
    return true;
  }

  return {
    labelPrinting,
    defaultTemplateId,
    isQuickPrintAvailable,
    quickPrintItems,
    quickPrintLocations,
  };
}
