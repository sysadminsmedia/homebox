import type { HTMLAttributes } from "vue";
import { cva, type VariantProps } from "class-variance-authority";

export interface SidebarProps {
  side?: "left" | "right";
  variant?: "sidebar" | "floating" | "inset";
  collapsible?: "offcanvas" | "icon" | "none";
  class?: HTMLAttributes["class"];
}

export { default as Sidebar } from "./Sidebar.vue";
export { default as SidebarContent } from "./SidebarContent.vue";
export { default as SidebarFooter } from "./SidebarFooter.vue";
export { default as SidebarGroup } from "./SidebarGroup.vue";
export { default as SidebarGroupAction } from "./SidebarGroupAction.vue";
export { default as SidebarGroupContent } from "./SidebarGroupContent.vue";
export { default as SidebarGroupLabel } from "./SidebarGroupLabel.vue";
export { default as SidebarHeader } from "./SidebarHeader.vue";
export { default as SidebarInput } from "./SidebarInput.vue";
export { default as SidebarInset } from "./SidebarInset.vue";
export { default as SidebarMenu } from "./SidebarMenu.vue";
export { default as SidebarMenuAction } from "./SidebarMenuAction.vue";
export { default as SidebarMenuBadge } from "./SidebarMenuBadge.vue";
export { default as SidebarMenuButton } from "./SidebarMenuButton.vue";
export { default as SidebarMenuLink } from "./SidebarMenuLink.vue";
export { default as SidebarMenuItem } from "./SidebarMenuItem.vue";
export { default as SidebarMenuSkeleton } from "./SidebarMenuSkeleton.vue";
export { default as SidebarMenuSub } from "./SidebarMenuSub.vue";
export { default as SidebarMenuSubButton } from "./SidebarMenuSubButton.vue";
export { default as SidebarMenuSubItem } from "./SidebarMenuSubItem.vue";
export { default as SidebarProvider } from "./SidebarProvider.vue";
export { default as SidebarRail } from "./SidebarRail.vue";
export { default as SidebarSeparator } from "./SidebarSeparator.vue";
export { default as SidebarTrigger } from "./SidebarTrigger.vue";

export { useSidebar } from "./utils";

export const sidebarMenuButtonVariants = cva(
  "peer/menu-button ring-sidebar-ring hover:bg-sidebar-accent hover:text-sidebar-accent-foreground active:bg-sidebar-accent active:text-sidebar-accent-foreground data-[active=true]:bg-sidebar-accent data-[active=true]:text-sidebar-accent-foreground data-[state=open]:hover:bg-sidebar-accent data-[state=open]:hover:text-sidebar-accent-foreground flex w-full items-center gap-2 overflow-hidden rounded-md p-2 text-left text-sm outline-none transition-[width,height,padding] focus-visible:ring-2 disabled:pointer-events-none disabled:opacity-50 group-has-[[data-sidebar=menu-action]]/menu-item:pr-8 aria-disabled:pointer-events-none aria-disabled:opacity-50 data-[active=true]:font-medium group-data-[collapsible=icon]:!size-10 group-data-[collapsible=icon]:!p-2 [&>span:last-child]:truncate [&>svg]:size-6 [&>svg]:shrink-0",
  {
    variants: {
      variant: {
        default: "hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
        outline:
          "hover:bg-sidebar-accent hover:text-sidebar-accent-foreground bg-background shadow-[0_0_0_1px_hsl(var(--sidebar-border))] hover:shadow-[0_0_0_1px_hsl(var(--sidebar-accent))]",
      },
      size: {
        default: "h-8 text-sm",
        sm: "h-7 text-xs",
        lg: "h-12 text-sm group-data-[collapsible=icon]:!p-0",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
);

export type SidebarMenuButtonVariants = VariantProps<typeof sidebarMenuButtonVariants>;
