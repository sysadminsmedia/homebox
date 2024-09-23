<template>
  <div>
    <!--
    Confirmation Modal is a singleton used by all components so we render
    it here to ensure it's always available. Possibly could move this further
    up the tree
   -->
    <ModalConfirm />
    <ItemCreateModal v-model="modals.item" />
    <LabelCreateModal v-model="modals.label" />
    <LocationCreateModal v-model="modals.location" />
    <AppToast />
    <div class="drawer drawer-mobile">
      <input id="my-drawer-2" v-model="drawerToggle" type="checkbox" class="drawer-toggle" />
      <div class="drawer-content bg-base-300 justify-center pt-20 lg:pt-0">
        <AppHeaderDecor v-if="preferences.displayHeaderDecor" class="-mt-10 hidden lg:block" />
        <!-- Button -->
        <div class="navbar drawer-button bg-primary fixed top-0 z-[99] shadow-md lg:hidden">
          <label for="my-drawer-2" class="btn btn-square btn-ghost drawer-button text-base-100 lg:hidden">
            <MdiMenu class="size-6" />
          </label>
          <NuxtLink to="/home">
            <h2 class="text-base-100 flex text-3xl font-bold tracking-tight">
              HomeB
              <AppLogo class="-mb-3 w-8" />
              x
            </h2>
          </NuxtLink>
        </div>

        <slot></slot>
        <footer v-if="status" class="bg-base-300 text-secondary-content bottom-0 w-full pb-4 text-center">
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
        <div class="bg-base-200 flex min-w-40 max-w-min flex-col p-5 md:py-10">
          <div class="space-y-8">
            <div class="flex flex-col items-center gap-4">
              <p>{{ $t("global.welcome", { username: username }) }}</p>
              <NuxtLink class="avatar placeholder" to="/home">
                <div class="bg-base-300 text-neutral-content w-24 rounded-full p-4">
                  <AppLogo />
                </div>
              </NuxtLink>
            </div>
            <div class="bg-base-200 flex flex-col">
              <div class="mb-6">
                <div class="dropdown visible w-full">
                  <label tabindex="0" class="text-no-transform btn btn-primary btn-block text-lg">
                    <span>
                      <MdiPlus class="-ml-1 mr-1" />
                    </span>
                    {{ $t("global.create") }}
                  </label>
                  <ul tabindex="0" class="dropdown-content menu rounded-box bg-base-100 w-full p-2 shadow">
                    <li v-for="btn in dropdown" :key="btn.name">
                      <button @click="btn.action">
                        {{ btn.name }}
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
                    {{ n.name }}
                  </NuxtLink>
                </li>
              </ul>
            </div>
          </div>

          <!-- Bottom -->
          <button class="rounded-btn hover:bg-base-300 mx-2 mt-auto p-3" @click="logout">
            {{ $t("global.sign_out") }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { useI18n } from "vue-i18n";
  import { useLabelStore } from "~~/stores/labels";
  import { useLocationStore } from "~~/stores/locations";
  import MdiMenu from "~icons/mdi/menu";
  import MdiPlus from "~icons/mdi/plus";

  import MdiHome from "~icons/mdi/home";
  import MdiFileTree from "~icons/mdi/file-tree";
  import MdiMagnify from "~icons/mdi/magnify";
  import MdiAccount from "~icons/mdi/account";
  import MdiCog from "~icons/mdi/cog";
  import MdiWrench from "~icons/mdi/wrench";

  const { t } = useI18n();
  const username = computed(() => authCtx.user?.name || "User");

  const preferences = useViewPreferences();

  const pubApi = usePublicApi();
  const { data: status } = useAsyncData(async () => {
    const { data } = await pubApi.status();

    return data;
  });

  // Preload currency format
  useFormatCurrency();
  const modals = reactive({
    item: false,
    location: false,
    label: false,
    import: false,
  });

  const dropdown = [
    {
      name: "Item / Asset",
      action: () => {
        modals.item = true;
      },
    },
    {
      name: "Location",
      action: () => {
        modals.location = true;
      },
    },
    {
      name: "Label",
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
      name: t("menu.home"),
      to: "/home",
    },
    {
      icon: MdiFileTree,
      id: 1,
      active: computed(() => route.path === "/locations"),
      name: t("menu.locations"),
      to: "/locations",
    },
    {
      icon: MdiMagnify,
      id: 2,
      active: computed(() => route.path === "/items"),
      name: t("menu.search"),
      to: "/items",
    },
    {
      icon: MdiWrench,
      id: 3,
      active: computed(() => route.path === "/maintenance"),
      name: t("menu.maintenance"),
      to: "/maintenance",
    },
    {
      icon: MdiAccount,
      id: 4,
      active: computed(() => route.path === "/profile"),
      name: t("menu.profile"),
      to: "/profile",
    },
    {
      icon: MdiCog,
      id: 5,
      active: computed(() => route.path === "/tools"),
      name: t("menu.tools"),
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
