import type { Ref } from "vue";
import type { TableHeaderType } from "~/components/Item/View/Table.types";
import type { DaisyTheme } from "~~/lib/data/themes";

export type ViewType = "table" | "card" | "tree";

export type DuplicateSettings = {
  copyMaintenance: boolean;
  copyAttachments: boolean;
  copyCustomFields: boolean;
  copyPrefixOverride: string | null;
};

export type LocationViewPreferences = {
  showDetails: boolean;
  showEmpty: boolean;
  editorAdvancedView: boolean;
  itemDisplayView: ViewType;
  theme: DaisyTheme;
  itemsPerTablePage: number;
  tableHeaders?: TableHeaderType[];
  displayLegacyHeader: boolean;
  language?: string;
  overrideFormatLocale?: string;
  duplicateSettings: DuplicateSettings;
};

/**
 * useViewPreferences loads the view preferences from local storage and hydrates
 * them. These are reactive and will update the local storage when changed.
 */
export function useViewPreferences(): Ref<LocationViewPreferences> {
  const results = useLocalStorage(
    "homebox/preferences/location",
    {
      showDetails: true,
      showEmpty: true,
      editorAdvancedView: false,
      itemDisplayView: "card",
      theme: "homebox",
      itemsPerTablePage: 10,
      displayLegacyHeader: false,
      language: null,
      overrideFormatLocale: null,
      duplicateSettings: {
        copyMaintenance: false,
        copyAttachments: true,
        copyCustomFields: true,
        copyPrefixOverride: null,
      },
    },
    { mergeDefaults: true }
  );

  // casting is required because the type returned is removable, however since we
  // use `mergeDefaults` the result _should_ always be present.
  return results as unknown as Ref<LocationViewPreferences>;
}
