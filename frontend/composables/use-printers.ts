import type {
  PrinterCreate,
  PrinterOut,
  PrinterSummary,
  PrinterUpdate,
  PrinterStatusResponse,
  PrinterTestResponse,
  LabelTemplatePrintResponse,
  LabelPrintItem,
  LabelPrintLocation,
} from "~~/lib/api/types/data-contracts";

export function usePrinters() {
  const api = useUserApi();

  const { data, pending, refresh, error } = useAsyncData("printers", async () => {
    const { data, error } = await api.printers.getAll();
    if (error) {
      throw new Error("Failed to load printers");
    }
    return data;
  });

  return {
    printers: data as Ref<PrinterSummary[] | null>,
    pending,
    refresh,
    error,
  };
}

export function usePrinter(id: MaybeRefOrGetter<string>) {
  const api = useUserApi();

  const { data, pending, refresh, error } = useAsyncData(
    () => `printer-${toValue(id)}`,
    async () => {
      const printerId = toValue(id);
      if (!printerId) return null;

      const { data, error } = await api.printers.get(printerId);
      if (error) {
        throw new Error("Failed to load printer");
      }
      return data;
    },
    { watch: [() => toValue(id)] }
  );

  return {
    printer: data as Ref<PrinterOut | null>,
    pending,
    refresh,
    error,
  };
}

export function usePrinterActions() {
  const api = useUserApi();

  async function create(body: PrinterCreate): Promise<PrinterOut | null> {
    const { data, error } = await api.printers.create(body);
    if (error) {
      throw new Error("Failed to create printer");
    }
    return data;
  }

  async function update(id: string, body: PrinterUpdate): Promise<PrinterOut | null> {
    const { data, error } = await api.printers.update(id, body);
    if (error) {
      throw new Error("Failed to update printer");
    }
    return data;
  }

  async function remove(id: string): Promise<void> {
    const { error } = await api.printers.delete(id);
    if (error) {
      throw new Error("Failed to delete printer");
    }
  }

  async function setDefault(id: string): Promise<void> {
    const { error } = await api.printers.setDefault(id);
    if (error) {
      throw new Error("Failed to set default printer");
    }
  }

  async function getStatus(id: string): Promise<PrinterStatusResponse | null> {
    const { data, error } = await api.printers.getStatus(id);
    if (error) {
      throw new Error("Failed to get printer status");
    }
    return data;
  }

  async function testPrint(id: string): Promise<PrinterTestResponse | null> {
    const { data, error } = await api.printers.testPrint(id);
    if (error) {
      throw new Error("Failed to test printer");
    }
    return data;
  }

  return {
    create,
    update,
    remove,
    setDefault,
    getStatus,
    testPrint,
  };
}

export function useDirectPrint() {
  const api = useUserApi();

  /**
   * Print labels with optional per-item quantities
   * @param templateId - Template to use
   * @param itemsOrIds - Either array of item IDs (backward compatible) or items with quantities
   * @param printerId - Printer to use (optional)
   * @param defaultCopies - Default copies per label if not specified per-item
   */
  async function printLabels(
    templateId: string,
    itemsOrIds: string[] | LabelPrintItem[],
    printerId?: string | null,
    defaultCopies: number = 1
  ): Promise<LabelTemplatePrintResponse | null> {
    // Check if we have items with quantities or just IDs
    const hasQuantities = itemsOrIds.length > 0 && typeof itemsOrIds[0] === "object";

    const { data, error } = await api.labelTemplates.print(templateId, {
      itemIds: hasQuantities ? [] : (itemsOrIds as string[]),
      items: hasQuantities ? (itemsOrIds as LabelPrintItem[]) : [],
      printerId: printerId || undefined,
      copies: defaultCopies,
    });
    if (error) {
      throw new Error("Failed to print labels");
    }
    return data;
  }

  /**
   * Print location labels with optional per-location quantities
   * @param templateId - Template to use
   * @param locationsOrIds - Either array of location IDs or locations with quantities
   * @param printerId - Printer to use (optional)
   * @param defaultCopies - Default copies per label if not specified per-location
   */
  async function printLocationLabels(
    templateId: string,
    locationsOrIds: string[] | LabelPrintLocation[],
    printerId?: string | null,
    defaultCopies: number = 1
  ): Promise<LabelTemplatePrintResponse | null> {
    // Check if we have locations with quantities or just IDs
    const hasQuantities = locationsOrIds.length > 0 && typeof locationsOrIds[0] === "object";

    const { data, error } = await api.labelTemplates.printLocations(templateId, {
      locationIds: hasQuantities ? [] : (locationsOrIds as string[]),
      locations: hasQuantities ? (locationsOrIds as LabelPrintLocation[]) : [],
      printerId: printerId || undefined,
      copies: defaultCopies,
    });
    if (error) {
      throw new Error("Failed to print location labels");
    }
    return data;
  }

  return {
    printLabels,
    printLocationLabels,
  };
}

/**
 * Composable for quick printing with default printer and template.
 * Provides one-click printing when defaults are configured.
 */
export function useQuickPrint() {
  const { printers } = usePrinters();
  const preferences = useViewPreferences();
  const { printLabels, printLocationLabels } = useDirectPrint();

  // Get the default printer (marked as default on server)
  const defaultPrinter = computed(() => {
    if (!printers.value) return null;
    return printers.value.find(p => p.isDefault) || null;
  });

  // Get the default template ID from user preferences
  const defaultTemplateId = computed(() => preferences.value.defaultTemplateId);

  // Check if quick print is available (both defaults configured)
  const isQuickPrintAvailable = computed(() => {
    return !!defaultPrinter.value && !!defaultTemplateId.value;
  });

  /**
   * Quick print items using default printer and template.
   * @returns true if print was initiated, false if defaults not configured
   */
  async function quickPrintItems(itemIds: string[]): Promise<boolean> {
    if (!defaultPrinter.value || !defaultTemplateId.value) {
      return false;
    }

    await printLabels(defaultTemplateId.value, itemIds, defaultPrinter.value.id);
    return true;
  }

  /**
   * Quick print locations using default printer and template.
   * @returns true if print was initiated, false if defaults not configured
   */
  async function quickPrintLocations(locationIds: string[]): Promise<boolean> {
    if (!defaultPrinter.value || !defaultTemplateId.value) {
      return false;
    }

    await printLocationLabels(defaultTemplateId.value, locationIds, defaultPrinter.value.id);
    return true;
  }

  return {
    defaultPrinter,
    defaultTemplateId,
    isQuickPrintAvailable,
    quickPrintItems,
    quickPrintLocations,
  };
}
