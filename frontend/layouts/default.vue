<template>
  <div id="app">
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
    <AppScannerModal />
    <SidebarProvider :default-open="sidebarState">
      <Sidebar collapsible="icon">
        <SidebarHeader class="items-center">
          <SidebarGroupLabel class="text-base group-data-[collapsible=icon]:hidden">{{
            $t("global.welcome", { username: username })
          }}</SidebarGroupLabel>
          <NuxtLink class="group-data-[collapsible=icon]:hidden" to="/home">
            <div class="flex size-24 items-center justify-center rounded-full bg-background-accent p-4">
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
            <DropdownMenuContent class="z-40 min-w-[var(--reka-dropdown-menu-trigger-width)]">
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

        <SidebarContent>
          <SidebarGroup>
            <SidebarMenu>
              <SidebarMenuItem v-for="n in nav" :key="n.id">
                <SidebarMenuLink
                  :href="n.to"
                  :class="{
                    'bg-accent text-accent-foreground': n.active?.value,
                    'text-nowrap': typeof locale === 'string' && locale.startsWith('zh-'),
                  }"
                  :tooltip="n.name.value"
                >
                  <component :is="n.icon" />
                  <span>{{ n.name.value }}</span>
                </SidebarMenuLink>
              </SidebarMenuItem>

              <!-- makes scanner accessible easily if using legacy header -->
              <SidebarMenuItem v-if="preferences.displayLegacyHeader">
                <SidebarMenuButton
                  :class="{
                    'text-nowrap': typeof locale === 'string' && locale.startsWith('zh-'),
                  }"
                  :tooltip="$t('menu.scanner')"
                  @click.prevent="openDialog('scanner')"
                >
                  <MdiQrcodeScan />
                  <span>{{ $t("menu.scanner") }}</span>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroup>
        </SidebarContent>

        <SidebarFooter>
          <SidebarMenuButton
            class="flex justify-center group-data-[collapsible=icon]:justify-start group-data-[collapsible=icon]:bg-destructive group-data-[collapsible=icon]:text-destructive-foreground group-data-[collapsible=icon]:shadow-sm group-data-[collapsible=icon]:hover:bg-destructive/90"
            :tooltip="$t('global.sign_out')"
            data-testid="logout-button"
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
      <SidebarInset class="min-h-dvh bg-background-accent">
        <div class="relative flex h-full flex-col justify-center">
          <div v-if="preferences.displayLegacyHeader">
            <AppHeaderDecor class="-mt-10 hidden lg:block" />
            <SidebarTrigger class="absolute left-2 top-2 hidden lg:flex" variant="default" />
          </div>
          <!-- IMPORTANT: if you change the height of this div, alter the top value in the item edit page-->
          <div
            class="sticky top-0 z-20 flex h-28 translate-y-[-0.5px] flex-col bg-secondary p-2 shadow-md sm:h-16 sm:flex-row"
            :class="{
              'lg:hidden': preferences.displayLegacyHeader,
            }"
          >
            <div class="flex h-1/2 items-center gap-2 sm:h-auto">
              <SidebarTrigger variant="default" />
              <NuxtLink to="/home">
                <AppHeaderText class="h-6" />
              </NuxtLink>
            </div>
            <div class="sm:grow"></div>
            <div class="flex h-1/2 grow items-center justify-end gap-2 sm:h-auto">
              <Input
                v-model:model-value="search"
                class="h-9 grow sm:max-w-sm"
                :placeholder="$t('global.search')"
                type="search"
                @keyup.enter="triggerSearch"
              />
              <div>
                <Button size="icon" @click="triggerSearch">
                  <MdiMagnify />
                </Button>
              </div>
              <div>
                <Button size="icon" @click="openScanner">
                  <MdiQrcodeScan />
                </Button>
              </div>
            </div>
          </div>

          <slot></slot>
          <div class="grow"></div>

          <footer v-if="status" class="bottom-0 w-full pb-4 text-center">
            <p class="text-center text-sm">
              <span
                v-html="
                  DOMPurify.sanitize(
                    $t('global.footer.version_link', { version: status.build.version, build: status.build.commit })
                  )
                "
              ></span>
              ~
              <span v-html="DOMPurify.sanitize($t('global.footer.api_link'))"></span>
            </p>
          </footer>
        </div>
      </SidebarInset>
    </SidebarProvider>
  </div>
</template>

<script lang="ts" setup>
  import { useI18n } from "vue-i18n";
  import DOMPurify from "dompurify";
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
  import { Shortcut } from "~/components/ui/shortcut";
  import { useDialog } from "~/components/ui/dialog-provider";
  import { Input } from "~/components/ui/input";
  import { Button } from "~/components/ui/button";
  import { toast } from "@/components/ui/sonner";

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

  const search = ref("");

  const triggerSearch = () => {
    if (search.value) {
      navigateTo(`/items?q=${encodeURIComponent(search.value)}`);
      search.value = "";
      // remove focus from input
      if (document.activeElement && "blur" in document.activeElement) {
        (document.activeElement as HTMLElement).blur();
      }
    }
  };

  const openScanner = () => {
    // request permission
    if (navigator.mediaDevices) {
      navigator.mediaDevices
        .getUserMedia({ video: true })
        .then(() => {
          openDialog("scanner");
        })
        .catch(err => {
          console.error(err);
          toast.error(t("scanner.permission_denied"));
        });
    } else {
      toast.error(t("scanner.unsupported"));
    }
  };

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
      shortcut: "Shift+3",
      dialogId: "create-location",
    },
    {
      id: 2,
      name: computed(() => t("menu.create_label")),
      shortcut: "Shift+2",
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
