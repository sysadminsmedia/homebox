import MdiHome from "~icons/mdi/home";
import MdiRoomService from "~icons/mdi/room-service";
import MdiChair from "~icons/mdi/chair";
import MdiLaptop from "~icons/mdi/laptop";
import MdiWrench from "~icons/mdi/wrench";
import MdiBook from "~icons/mdi/book";
import MdiFileCabinet from "~icons/mdi/file-cabinet";
import MdiDumbbell from "~icons/mdi/dumbbell";
import MdiBasketball from "~icons/mdi/basketball";
import MdiGamepad from "~icons/mdi/gamepad";
import MdiToolbox from "~icons/mdi/toolbox";
import MdiLightbulb from "~icons/mdi/lightbulb";
import MdiPlug from "~icons/mdi/plug";
import MdiCloset from "~icons/mdi/closet";
import MdiMicrowave from "~icons/mdi/microwave";
import MdiCoffee from "~icons/mdi/coffee";
import MdiTagOutline from "~icons/mdi/tag-outline";

export type IconName =
  | "home"
  | "room-service"
  | "chair"
  | "laptop"
  | "wrench"
  | "book"
  | "file-cabinet"
  | "dumbbell"
  | "basketball"
  | "gamepad"
  | "toolbox"
  | "lightbulb"
  | "plug"
  | "closet"
  | "microwave"
  | "coffee";

export const availableIcons = [
  { name: "home", component: MdiHome },
  { name: "room-service", component: MdiRoomService },
  { name: "chair", component: MdiChair },
  { name: "laptop", component: MdiLaptop },
  { name: "wrench", component: MdiWrench },
  { name: "book", component: MdiBook },
  { name: "file-cabinet", component: MdiFileCabinet },
  { name: "dumbbell", component: MdiDumbbell },
  { name: "basketball", component: MdiBasketball },
  { name: "gamepad", component: MdiGamepad },
  { name: "toolbox", component: MdiToolbox },
  { name: "lightbulb", component: MdiLightbulb },
  { name: "plug", component: MdiPlug },
  { name: "closet", component: MdiCloset },
  { name: "microwave", component: MdiMicrowave },
  { name: "coffee", component: MdiCoffee },
] as const;

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function getIconComponent(iconName: string | undefined): any {
  if (!iconName) {
    return defaultIcon;
  }
  const icon = availableIcons.find(i => i.name === iconName);
  return icon ? icon.component : defaultIcon;
}

export const defaultIcon = MdiTagOutline;
