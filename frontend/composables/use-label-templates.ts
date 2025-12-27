import type {
  LabelTemplateCreate,
  LabelTemplateOut,
  LabelTemplateSummary,
  LabelTemplateUpdate,
  LabelmakerLabelPreset,
  LabelmakerBarcodeFormatInfo,
} from "~~/lib/api/types/data-contracts";

export function useLabelTemplates() {
  const api = useUserApi();

  const { data, pending, refresh, error } = useAsyncData("label-templates", async () => {
    const { data, error } = await api.labelTemplates.getAll();
    if (error) {
      throw new Error("Failed to load label templates");
    }
    return data;
  });

  return {
    templates: data as Ref<LabelTemplateSummary[] | null>,
    pending,
    refresh,
    error,
  };
}

export function useLabelTemplate(id: MaybeRefOrGetter<string>) {
  const api = useUserApi();

  const { data, pending, refresh, error } = useAsyncData(
    () => `label-template-${toValue(id)}`,
    async () => {
      const templateId = toValue(id);
      if (!templateId) return null;

      const { data, error } = await api.labelTemplates.get(templateId);
      if (error) {
        throw new Error("Failed to load label template");
      }
      return data;
    },
    { watch: [() => toValue(id)] }
  );

  return {
    template: data as Ref<LabelTemplateOut | null>,
    pending,
    refresh,
    error,
  };
}

export function useLabelPresets() {
  const api = useUserApi();

  const { data, pending, error } = useAsyncData("label-presets", async () => {
    const { data, error } = await api.labelTemplates.getPresets();
    if (error) {
      throw new Error("Failed to load label presets");
    }
    return data;
  });

  return {
    presets: data as Ref<LabelmakerLabelPreset[] | null>,
    pending,
    error,
  };
}

export function useBarcodeFormats() {
  const api = useUserApi();

  const { data, pending, error } = useAsyncData("barcode-formats", async () => {
    const { data, error } = await api.labelTemplates.getBarcodeFormats();
    if (error) {
      throw new Error("Failed to load barcode formats");
    }
    return data;
  });

  return {
    formats: data as Ref<LabelmakerBarcodeFormatInfo[] | null>,
    pending,
    error,
  };
}

export function useLabelTemplateActions() {
  const api = useUserApi();

  async function create(body: LabelTemplateCreate): Promise<LabelTemplateOut | null> {
    const { data, error } = await api.labelTemplates.create(body);
    if (error) {
      throw new Error("Failed to create label template");
    }
    return data;
  }

  async function update(id: string, body: LabelTemplateUpdate): Promise<LabelTemplateOut | null> {
    const { data, error } = await api.labelTemplates.update(id, body);
    if (error) {
      throw new Error("Failed to update label template");
    }
    return data;
  }

  async function remove(id: string): Promise<void> {
    const { error } = await api.labelTemplates.delete(id);
    if (error) {
      throw new Error("Failed to delete label template");
    }
  }

  async function duplicate(id: string): Promise<LabelTemplateOut | null> {
    const { data, error } = await api.labelTemplates.duplicate(id);
    if (error) {
      throw new Error("Failed to duplicate label template");
    }
    return data;
  }

  async function render(
    id: string,
    itemIds: string[],
    format: string = "png",
    pageSize: string = "Letter",
    showCutGuides: boolean = false,
    canvasData?: string
  ): Promise<Blob> {
    return api.labelTemplates.render(id, { itemIds, format, pageSize, showCutGuides, canvasData: canvasData || "" });
  }

  async function renderLocations(
    id: string,
    locationIds: string[],
    format: string = "png",
    pageSize: string = "Letter",
    showCutGuides: boolean = false
  ): Promise<Blob> {
    return api.labelTemplates.renderLocations(id, { locationIds, format, pageSize, showCutGuides });
  }

  function getPreviewUrl(id: string): string {
    return api.labelTemplates.getPreviewUrl(id);
  }

  return {
    create,
    update,
    remove,
    duplicate,
    render,
    renderLocations,
    getPreviewUrl,
  };
}
