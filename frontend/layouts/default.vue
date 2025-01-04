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
    <AppToast />
    <div class="drawer drawer-mobile">
      <input id="my-drawer-2" v-model="drawerToggle" type="checkbox" class="drawer-toggle" />
      <div class="drawer-content justify-center bg-base-300 pt-20 lg:pt-0">
        <AppHeaderDecor v-if="preferences.displayHeaderDecor" class="-mt-10 hidden lg:block" />
        <!-- Button -->
        <div class="navbar drawer-button fixed top-0 z-[99] bg-primary shadow-md lg:hidden">
          <label for="my-drawer-2" class="btn btn-square btn-ghost drawer-button text-base-100 lg:hidden">
            <MdiMenu class="size-6" />
          </label>
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
            {{ $t("global.version", { version: status.build.version }) }} ~
            {{ $t("global.build", { build: status.build.commit }) }}
          </p>
        </footer>
      </div>

      <!-- Sidebar -->
      <div class="drawer-side shadow-lg">
        <label for="my-drawer-2" class="drawer-overlay"></label>

        <!-- Top Section -->
        <div class="flex min-w-40 max-w-min flex-col bg-base-200 p-5 md:py-10">
          <div class="space-y-8">
            <div class="flex flex-col items-center gap-4">
              <p>{{ $t("global.welcome", { username: username }) }}</p>
              <NuxtLink class="avatar placeholder" to="/home">
                <div class="w-24 rounded-full bg-base-300 p-4 text-neutral-content">
                  <AppLogo />
                </div>
              </NuxtLink>
            </div>
            <div class="flex flex-col bg-base-200">
              <div class="mb-6">
                <div class="dropdown visible w-full">
                  <label tabindex="0" class="text-no-transform btn btn-primary btn-block text-lg">
                    <span>
                      <MdiPlus class="-ml-1 mr-1" />
                    </span>
                    {{ $t("global.create") }}
                  </label>
                  <ul tabindex="0" class="dropdown-content menu rounded-box w-full bg-base-100 p-2 shadow">
                    <li v-for="btn in dropdown" :key="btn.id">
                      <button @click="btn.action">
                        {{ btn.name.value }}
                      </button>
                    </li>
                  </ul>
                </div>
              </div>
              <ul class="menu mx-auto flex flex-col gap-2">
                <li v-for="n in nav" :key="n.id" class="text-xl" @click="unfocus">
                  <NuxtLink
                    v-if="n.to"
                    class="rounded-btn"
                    :to="n.to"
                    :class="{
                      'bg-secondary text-secondary-content': n.active?.value,
                    }"
                  >
                    <component :is="n.icon" class="mr-4 size-6" />
                    {{ n.name.value }}
                  </NuxtLink>
                </li>
              </ul>
            </div>
          </div>

          <!-- Bottom -->
          <button class="rounded-btn mx-2 mt-auto p-3 hover:bg-base-300" @click="logout">
            {{ $t("global.sign_out") }}
          </button>
        </div>
      </div>
    </div>
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
  import MdiMenu from "~icons/mdi/menu";
  import MdiPlus from "~icons/mdi/plus";

  const { t } = useI18n();
  const username = computed(() => authCtx.user?.name || "User");

  const preferences = useViewPreferences();

  const pubApi = usePublicApi();
  const { data: status } = useAsyncData(async () => {
    const { data } = await pubApi.status();

    return data;
  });

  const latest = computed(() => status.value?.latest.version);
  const current = computed(() => status.value?.build.version);

  const isDev = computed(() => import.meta.dev || !current.value?.includes("."));
  const isOutdated = computed(() => current && latest.value && lt(current, latest.value));
  const hasHiddenLatest = computed(() => localStorage.getItem("latestVersion") === latest.value);

  const displayOutdatedWarning = computed(() => !isDev && !hasHiddenLatest.value && isOutdated.value);

  // Preload currency format
  useFormatCurrency();
  const modals = reactive({
    item: false,
    location: false,
    label: false,
    import: false,
    outdated: displayOutdatedWarning.value,
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
      action: () => {
        modals.item = true;
      },
    },
    {
      id: 1,
      name: computed(() => t("menu.create_location")),
      action: () => {
        modals.location = true;
      },
    },
    {
      id: 2,
      name: computed(() => t("menu.create_label")),
      action: () => {
        modals.label = true;
      },
    },
  ];

  const route = useRoute();

  const drawerToggle = ref();

  function unfocus() {
    // unfocus current element
    drawerToggle.value = false;
  }

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
