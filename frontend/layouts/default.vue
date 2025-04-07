<template>
  <div id="app">
    <Toaster />
    <!--
    Confirmation Modal is a singleton used by all components so we render
    it here to ensure it's always available. Possibly could move this further
    up the tree
    -->
    <ModalConfirm />
    <AppOutdatedModal v-if="status" :status="status" />
    <ItemCreateModal />
    <LabelCreateModal />
    <LocationCreateModal />
    <AppQuickMenuModal :actions="quickMenuActions" />
    <SidebarProvider :default-open="sidebarState">
      <Sidebar collapsible="icon">
        <SidebarHeader class="items-center bg-base-200">
          <SidebarGroupLabel class="text-base">{{ $t("global.welcome", { username: username }) }}</SidebarGroupLabel>
          <NuxtLink class="avatar placeholder group-data-[collapsible=icon]:hidden" to="/home">
            <div class="w-24 rounded-full bg-base-300 p-4 text-neutral-content">
              <AppLogo />
            </div>
          </NuxtLink>
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <SidebarMenuButton
                class="flex justify-center bg-primary text-primary-foreground shadow hover:bg-primary/90 group-data-[collapsible=icon]:justify-start"
                :tooltip="$t('global.create')"
                hotkey="Shortcut: Ctrl+`"
              >
                <MdiPlus />
                <span>
                  {{ $t("global.create") }}
                </span>
              </SidebarMenuButton>
            </DropdownMenuTrigger>
            <DropdownMenuContent class="min-w-[var(--radix-dropdown-menu-trigger-width)]">
              <DropdownMenuItem
                v-for="btn in dropdown"
                :key="btn.id"
                class="group cursor-pointer text-lg"
                @click="openDialog(btn.dialogId)"
              >
                {{ btn.name.value }}
                <Shortcut
                  v-if="btn.shortcut"
                  class="ml-auto hidden group-hover:inline"
                  :keys="btn.shortcut.replace('Shift', '⇧').split('+')"
                />
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </SidebarHeader>

        <SidebarContent class="bg-base-200">
          <SidebarGroup>
            <SidebarMenu>
              <SidebarMenuItem v-for="n in nav" :key="n.id">
                <SidebarMenuLink
                  :href="n.to"
                  :class="{
                    'bg-secondary text-secondary-foreground': n.active?.value,
                    'text-nowrap': typeof locale === 'string' && locale.startsWith('zh-'),
                    'hover:bg-base-300': !n.active?.value,
                  }"
                  :tooltip="n.name.value"
                >
                  <component :is="n.icon" />
                  <span>{{ n.name.value }}</span>
                </SidebarMenuLink>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroup>
        </SidebarContent>

        <SidebarFooter class="bg-base-200">
          <SidebarMenuButton
            class="flex justify-center hover:bg-base-300 group-data-[collapsible=icon]:justify-start group-data-[collapsible=icon]:bg-destructive group-data-[collapsible=icon]:text-destructive-foreground group-data-[collapsible=icon]:shadow-sm group-data-[collapsible=icon]:hover:bg-destructive/90"
            :tooltip="$t('global.sign_out')"
            @click="logout"
          >
            <MdiLogout />
            <span>
              {{ $t("global.sign_out") }}
            </span>
          </SidebarMenuButton>
        </SidebarFooter>

        <SidebarRail />
      </Sidebar>
      <SidebarInset class="min-h-screen bg-base-300">
        <div class="justify-center pt-20 lg:pt-0">
          <AppHeaderDecor v-if="preferences.displayHeaderDecor" class="-mt-10 hidden lg:block" />
          <SidebarTrigger class="absolute left-2 top-2 hidden lg:flex" variant="default" />
          <div class="fixed top-0 z-20 flex h-16 w-full items-center gap-2 bg-primary p-2 shadow-md lg:hidden">
            <SidebarTrigger />
            <NuxtLink to="/home">
              <h2 class="flex text-3xl font-bold tracking-tight text-base-100">
                HomeB
                <AppLogo class="-mb-3 w-8" />
                x
              </h2>
            </NuxtLink>
          </div>

          <slot></slot>
          <footer v-if="status" class="bottom-0 w-full bg-base-300 pb-4 text-center text-secondary-content">
            <p class="text-center text-sm">
              <a
                href="https://github.com/sysadminsmedia/homebox/releases/tag/{{ status.build.version }}"
                target="_blank"
              >
                {{ $t("global.version", { version: status.build.version }) }} ~
                {{ $t("global.build", { build: status.build.commit }) }}</a
              >
              ~
              <a href="https://homebox.software/en/api.html" target="_blank">API</a>
            </p>
          </footer>
        </div>
      </SidebarInset>
    </SidebarProvider>
  </div>
</template>

<script lang="ts" setup>
  import { useI18n } from "vue-i18n";
  import { useLabelStore } from "~~/stores/labels";
  import { useLocationStore } from "~~/stores/locations";

  import MdiHome from "~icons/mdi/home";
  import MdiFileTree from "~icons/mdi/file-tree";
  import MdiMagnify from "~icons/mdi/magnify";
  import MdiQrcodeScan from "~icons/mdi/qrcode-scan";
  import MdiAccount from "~icons/mdi/account";
  import MdiCog from "~icons/mdi/cog";
  import MdiWrench from "~icons/mdi/wrench";
  import MdiPlus from "~icons/mdi/plus";
  import MdiLogout from "~icons/mdi/logout";

  import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarHeader,
    SidebarInset,
    SidebarRail,
    SidebarTrigger,
    SidebarGroup,
    SidebarGroupLabel,
    SidebarMenu,
    SidebarMenuItem,
    SidebarMenuButton,
    SidebarMenuLink,
  } from "@/components/ui/sidebar";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
  } from "@/components/ui/dropdown-menu";
  import { Toaster } from "~/components/ui/sonner";
  import { Shortcut } from "~/components/ui/shortcut";
  import { useDialog } from "~/components/ui/dialog-provider";

  const { t, locale } = useI18n();
  const username = computed(() => authCtx.user?.name || "User");

  const { openDialog } = useDialog();

  const preferences = useViewPreferences();

  // get sidebar state from cookies
  const sidebarState = useCookie("sidebar:state", {
    readonly: true,
    decode: value => value !== "false",
  });

  const pubApi = usePublicApi();
  const { data: status } = useAsyncData(async () => {
    const { data } = await pubApi.status();

    return data;
  });

  // Preload currency format
  useFormatCurrency();

  const dropdown = [
    {
      id: 0,
      name: computed(() => t("menu.create_item")),
      shortcut: "Shift+1",
      dialogId: "create-item",
    },
    {
      id: 1,
      name: computed(() => t("menu.create_location")),
      shortcut: "Shift+2",
      dialogId: "create-location",
    },
    {
      id: 2,
      name: computed(() => t("menu.create_label")),
      shortcut: "Shift+3",
      dialogId: "create-label",
    },
  ];

  const route = useRoute();

  const nav = [
    {
      icon: MdiHome,
      active: computed(() => route.path === "/home"),
      id: 0,
      name: computed(() => t("menu.home")),
      to: "/home",
    },
    {
      icon: MdiFileTree,
      id: 1,
      active: computed(() => route.path === "/locations"),
      name: computed(() => t("menu.locations")),
      to: "/locations",
    },
    {
      icon: MdiMagnify,
      id: 2,
      active: computed(() => route.path === "/items"),
      name: computed(() => t("menu.search")),
      to: "/items",
    },
    {
      icon: MdiQrcodeScan,
      id: 3,
      active: computed(() => route.path === "/scanner"),
      name: computed(() => t("menu.scanner")),
      to: "/scanner",
    },
    {
      icon: MdiWrench,
      id: 4,
      active: computed(() => route.path === "/maintenance"),
      name: computed(() => t("menu.maintenance")),
      to: "/maintenance",
    },
    {
      icon: MdiAccount,
      id: 5,
      active: computed(() => route.path === "/profile"),
      name: computed(() => t("menu.profile")),
      to: "/profile",
    },
    {
      icon: MdiCog,
      id: 6,
      active: computed(() => route.path === "/tools"),
      name: computed(() => t("menu.tools")),
      to: "/tools",
    },
  ];

  const quickMenuActions = reactive([
    ...dropdown.map(v => ({
      text: computed(() => v.name.value),
      dialogId: v.dialogId,
      shortcut: v.shortcut.split("+")[1],
      type: "create" as const,
    })),
    ...nav.map(v => ({
      text: computed(() => v.name.value),
      href: v.to,
      type: "navigate" as const,
    })),
  ]);

  const labelStore = useLabelStore();

  const locationStore = useLocationStore();

  onServerEvent(ServerEvent.LabelMutation, () => {
    labelStore.refresh();
  });

  onServerEvent(ServerEvent.LocationMutation, () => {
    locationStore.refreshChildren();
    locationStore.refreshParents();
    locationStore.refreshTree();
  });

  onServerEvent(ServerEvent.ItemMutation, () => {
    // item mutations can affect locations counts
    // so we need to refresh those as well
    locationStore.refreshChildren();
    locationStore.refreshParents();
  });

  const authCtx = useAuthContext();
  const api = useUserApi();

  async function logout() {
    await authCtx.logout(api);
    navigateTo("/");
  }
</script>
