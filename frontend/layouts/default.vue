<template>
  <div id="app">
    <!--
    Confirmation Modal is a singleton used by all components so we render
    it here to ensure it's always available. Possibly could move this further
    up the tree
   -->
    <ModalConfirm />
    <AppOutdatedModal v-model="modals.outdated" :current="current ?? ''" :latest="latest ?? ''" />
    <ItemCreateModal v-model="modals.item" />
    <LabelCreateModal v-model="modals.label" />
    <LocationCreateModal v-model="modals.location" />
    <QuickMenuModal v-model="modals.quickMenu" :actions="quickMenuActions" />
    <AppToast />
    <SidebarProvider>
      <Sidebar collapsible="icon">
        <SidebarHeader class="items-center bg-base-200">
          <SidebarGroupLabel class="text-base">{{ $t("global.welcome", { username: username }) }}</SidebarGroupLabel>
          <NuxtLink class="avatar placeholder group-data-[collapsible=icon]:hidden" to="/home">
            <div class="w-24 rounded-full bg-base-300 p-4 text-neutral-content">
              <AppLogo />
            </div>
          </NuxtLink>
          <DropdownMenu>
            <!--TODO: change tooltip to be shadcn -->
            <DropdownMenuTrigger as-child>
              <SidebarMenuButton
                class="flex justify-center bg-primary text-primary-foreground shadow hover:bg-primary/90 group-data-[collapsible=icon]:justify-start"
                data-tip="Shortcut: Ctrl+`"
              >
                <MdiPlus />
                <span>
                  {{ $t("global.create") }}
                </span>
              </SidebarMenuButton>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuItem v-for="btn in dropdown" :key="btn.id" @click="btn.action">
                <button>
                  {{ btn.name.value }}

                  <kbd v-if="btn.shortcut" class="ml-auto text-neutral-400">{{ btn.shortcut }}</kbd>
                </button>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </SidebarHeader>

        <SidebarContent class="bg-base-200">
          <SidebarGroup>
            <SidebarMenu>
              <SidebarMenuItem v-for="n in nav" :key="n.id">
                <SidebarMenuLink :href="n.to" class="hover:bg-accent hover:text-accent-foreground">
                  <component :is="n.icon" />
                  <span>{{ n.name.value }}</span>
                </SidebarMenuLink>
              </SidebarMenuItem>
            </SidebarMenu></SidebarGroup
          >
        </SidebarContent>

        <SidebarFooter class="bg-base-200">
          <SidebarMenuButton
            class="flex justify-center bg-destructive text-destructive-foreground shadow-sm hover:bg-destructive/90 group-data-[collapsible=icon]:justify-start"
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
          <!-- <AppHeaderDecor v-if="preferences.displayHeaderDecor" class="-mt-10 hidden lg:block" /> -->
          <!-- Button -->
          <div class="navbar drawer-button z-[99] bg-primary shadow-md">
            <!-- <label for="my-drawer-2" class="btn btn-square btn-ghost drawer-button text-base-100 lg:hidden">
              <MdiMenu class="size-6" />
            </label> -->
            <SidebarTrigger class="-ml-1" />
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
  import { lt } from "semver";
  import { useLabelStore } from "~~/stores/labels";
  import { useLocationStore } from "~~/stores/locations";

  import MdiHome from "~icons/mdi/home";
  import MdiFileTree from "~icons/mdi/file-tree";
  import MdiMagnify from "~icons/mdi/magnify";
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

  const { t } = useI18n();
  const username = computed(() => authCtx.user?.name || "User");

  // const preferences = useViewPreferences();

  const pubApi = usePublicApi();
  const { data: status } = useAsyncData(async () => {
    const { data } = await pubApi.status();

    return data;
  });

  const latest = computed(() => status.value?.latest.version);
  const current = computed(() => status.value?.build.version);

  const isDev = computed(() => import.meta.dev || !current.value?.includes("."));
  const isOutdated = computed(() => current.value && latest.value && lt(current.value, latest.value));
  const hasHiddenLatest = computed(() => localStorage.getItem("latestVersion") === latest.value);

  const displayOutdatedWarning = computed(() => !isDev.value && !hasHiddenLatest.value && isOutdated.value);

  const keys = useMagicKeys({
    aliasMap: {
      "⌃": "control_",
    },
  });

  // Preload currency format
  useFormatCurrency();
  const modals = reactive({
    item: false,
    location: false,
    label: false,
    import: false,
    outdated: displayOutdatedWarning.value,
    quickMenu: false,
  });

  watch(displayOutdatedWarning, () => {
    console.log("displayOutdatedWarning", displayOutdatedWarning.value);
    if (displayOutdatedWarning.value) {
      modals.outdated = true;
    }
  });

  const dropdown = [
    {
      id: 0,
      name: computed(() => t("menu.create_item")),
      shortcut: "⌃1",
      action: () => {
        modals.item = true;
      },
    },
    {
      id: 1,
      name: computed(() => t("menu.create_location")),
      shortcut: "⌃2",
      action: () => {
        modals.location = true;
      },
    },
    {
      id: 2,
      name: computed(() => t("menu.create_label")),
      shortcut: "⌃3",
      action: () => {
        modals.label = true;
      },
    },
  ];

  dropdown.forEach(option => {
    if (option.shortcut) {
      whenever(keys[option.shortcut], () => {
        option.action();
      });
    }
  });

  const route = useRoute();

  // const drawerToggle = ref();

  // function unfocus() {
  //   // unfocus current element
  //   drawerToggle.value = false;
  // }

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
      id: 3,
      active: computed(() => route.path === "/maintenance"),
      name: computed(() => t("menu.maintenance")),
      to: "/maintenance",
    },
    {
      icon: MdiAccount,
      id: 4,
      active: computed(() => route.path === "/profile"),
      name: computed(() => t("menu.profile")),
      to: "/profile",
    },
    {
      icon: MdiCog,
      id: 5,
      active: computed(() => route.path === "/tools"),
      name: computed(() => t("menu.tools")),
      to: "/tools",
    },
  ];

  const quickMenuShortcut = keys.control_Backquote;
  whenever(quickMenuShortcut, () => {
    modals.quickMenu = true;
    modals.item = false;
    modals.location = false;
    modals.label = false;
    modals.import = false;
  });

  const quickMenuActions = ref(
    [
      {
        text: computed(() => `${t("global.create")}: ${t("menu.create_item")}`),
        action: () => {
          modals.item = true;
        },
        shortcut: "1",
      },
      {
        text: computed(() => `${t("global.create")}: ${t("menu.create_location")}`),
        action: () => {
          modals.location = true;
        },
        shortcut: "2",
      },
      {
        text: computed(() => `${t("global.create")}: ${t("menu.create_label")}`),
        action: () => {
          modals.label = true;
        },
        shortcut: "3",
      },
    ].concat(
      nav.map(v => {
        return {
          text: computed(() => `${t("global.navigate")}: ${v.name.value}`),
          action: () => {
            navigateTo(v.to);
          },
          shortcut: "",
        };
      })
    )
  );

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
