import MdiTagOutline from "~icons/mdi/tag-outline";
import MdiTreeOutline from "~icons/mdi/tree-outline";
import MdiBagSuitcaseOutline from "~icons/mdi/bag-suitcase-outline";
import MdiBedOutline from "~icons/mdi/bed-outline";
import MdiKitchenCounterOutline from "~icons/mdi/kitchen-counter-outline";
import MdiBookOpenVariantOutline from "~icons/mdi/book-open-variant-outline";
import MdiLaptopOutline from "~icons/mdi/laptop";
import MdiToolboxOutline from "~icons/mdi/toolbox-outline";
import MdiFileCabinetOutline from "~icons/mdi/folder-outline";
import MdiDresserOutline from "~icons/mdi/dresser-outline";
import MdiLightbulbOutline from "~icons/mdi/lightbulb-outline";
import MdiPowerPlugOutline from "~icons/mdi/power-plug-outline";
import MdiWrenchOutline from "~icons/mdi/wrench-outline";
import MdiDumbbellOutline from "~icons/mdi/dumbbell";
import MdiSofaOutline from "~icons/mdi/sofa-outline";
import MdiPalleteOutline from "~icons/mdi/palette-outline";
import MdiBathtubOutline from "~icons/mdi/bathtub-outline";
import MdiMapMarkerOutline from "~icons/mdi/map-marker-outline";
import MdiBunkBedOutline from "~icons/mdi/bunk-bed-outline";
import MdiGarageVariant from "~icons/mdi/garage-variant";
import MdiPackageVariant from "~icons/mdi/package-variant";
import MdiPackageVariantClosed from "~icons/mdi/package-variant-closed";

export const availableIcons = [
  { name: "tag-outline", component: MdiTagOutline },
  { name: "tree-outline", component: MdiTreeOutline },
  { name: "bag-suitcase-outline", component: MdiBagSuitcaseOutline },
  { name: "bed-outline", component: MdiBedOutline },
  { name: "bunk-bed-outline", component: MdiBunkBedOutline },
  { name: "bathtub-outline", component: MdiBathtubOutline },
  { name: "kitchen-counter-outline", component: MdiKitchenCounterOutline },
  { name: "sofa-outline", component: MdiSofaOutline },
  { name: "dresser-outline", component: MdiDresserOutline },
  { name: "file-cabinet-outline", component: MdiFileCabinetOutline },
  { name: "garage-variant-outline", component: MdiGarageVariant },
  { name: "map-marker-outline", component: MdiMapMarkerOutline },
  { name: "folder-outline", component: MdiFileCabinetOutline },
  { name: "package-variant", component: MdiPackageVariant },
  { name: "package-variant-closed", component: MdiPackageVariantClosed },
  { name: "book-open-variant-outline", component: MdiBookOpenVariantOutline },
  { name: "laptop", component: MdiLaptopOutline },
  { name: "toolbox-outline", component: MdiToolboxOutline },
  { name: "lightbulb-outline", component: MdiLightbulbOutline },
  { name: "power-plug-outline", component: MdiPowerPlugOutline },
  { name: "wrench-outline", component: MdiWrenchOutline },
  { name: "dumbbell", component: MdiDumbbellOutline },
  { name: "palette-outline", component: MdiPalleteOutline },
] as const;

export type IconName = (typeof availableIcons)[number]["name"];

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function getIconComponent(iconName: string | undefined, defaultIcon: any | undefined): any {
  if (!iconName) {
    return defaultIcon || defaultTagIcon;
  }
  const icon = availableIcons.find(i => i.name === iconName);
  return icon ? icon.component : defaultIcon || defaultTagIcon;
}

export const defaultTagIcon = MdiTagOutline;
