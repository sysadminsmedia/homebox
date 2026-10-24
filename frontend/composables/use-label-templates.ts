import type { LabelTemplateCreate, LabelTemplateOut, LabelTemplateUpdate } from "~~/lib/api/types/data-contracts";

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
    templates: data,
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
    template: data,
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
    presets: data,
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
    formats: data,
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
    return api.labelTemplates.render(id, {
      itemIds,
      format,
      pageSize,
      showCutGuides,
      skipLabels: 0,
      canvasData: canvasData || "",
    });
  }

  async function renderLocations(
    id: string,
    locationIds: string[],
    format: string = "png",
    pageSize: string = "Letter",
    showCutGuides: boolean = false
  ): Promise<Blob> {
    return api.labelTemplates.renderLocations(id, { locationIds, format, pageSize, showCutGuides, skipLabels: 0 });
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

export type PresetKind = "sheet" | "continuous" | "die-cut";

export interface PresetInfo {
  key: string;
  brand: string;
  name: string;
  kind: PresetKind;
  /** labels per sheet, for sheet presets */
  perSheet?: number;
  width: number;
  height: number;
}

/**
 * Looks up label presets by key so the UI can show "Brother DK-22205" instead of
 * the raw preset key stored on a template.
 */
export function useLabelPresetLookup() {
  const { presets, pending } = useLabelPresets();

  const byKey = computed(() => {
    const map = new Map<string, PresetInfo>();
    for (const p of presets.value ?? []) {
      const kind: PresetKind = p.sheetLayout ? "sheet" : p.continuous ? "continuous" : "die-cut";
      map.set(p.key, {
        key: p.key,
        brand: p.brand || "",
        name: p.name,
        kind,
        perSheet: p.sheetLayout ? p.sheetLayout.columns * p.sheetLayout.rows : undefined,
        width: p.width,
        height: p.height,
      });
    }
    return map;
  });

  function find(key?: string | null): PresetInfo | null {
    if (!key) return null;
    return byKey.value.get(key) ?? null;
  }

  /** "Brother DK-22205" or the raw key while presets are still loading */
  function title(key?: string | null): string {
    const p = find(key);
    if (!p) return key ?? "";
    return [p.brand, p.name].filter(Boolean).join(" ");
  }

  return { presets, pending, find, title };
}

/** "62 × 29 mm" */
export function formatLabelSize(width: number, height: number): string {
  const fmt = (n: number) => (Number.isInteger(n) ? String(n) : n.toFixed(1));
  return `${fmt(width)} × ${fmt(height)} mm`;
}
