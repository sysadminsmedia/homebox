import type { Ref } from "vue";
import type { TableHeaderType } from "~/components/Item/View/Table.types";
import type { DaisyTheme } from "~~/lib/data/themes";

export type ViewType = "table" | "card" | "tree";

export type LocationViewPreferences = {
  showDetails: boolean;
  showEmpty: boolean;
  editorAdvancedView: boolean;
  itemDisplayView: ViewType;
  theme: DaisyTheme;
  itemsPerTablePage: number;
  tableHeaders?: TableHeaderType[];
  displayHeaderDecor: boolean;
  language?: string;
  overrideFormatLocale?: string;
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
      displayHeaderDecor: true,
      language: null,
      overrideFormatLocale: null,
    },
    { mergeDefaults: true }
  );

  // casting is required because the type returned is removable, however since we
  // use `mergeDefaults` the result _should_ always be present.
  return results as unknown as Ref<LocationViewPreferences>;
}
