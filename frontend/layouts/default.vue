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
      <div class="drawer-content justify-center bg-base-300 pt-20 lg:pt-0">
        <AppHeaderDecor class="-mt-10 hidden lg:block" />
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
        <div class="flex w-60 flex-col bg-base-200 py-5 md:py-10">
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
              <div class="mx-auto mb-6 w-40">
                <div class="dropdown visible w-40">
                  <label tabindex="0" class="text-no-transform btn btn-primary btn-block text-lg">
                    <span>
                      <MdiPlus class="-ml-1 mr-1" />
                    </span>
                    {{ $t("global.create") }}
                  </label>
                  <ul tabindex="0" class="dropdown-content menu rounded-box w-40 bg-base-100 p-2 shadow">
                    <li v-for="btn in dropdown" :key="btn.name">
                      <button @click="btn.action">
                        {{ btn.name }}
                      </button>
                    </li>
                  </ul>
                </div>
              </div>
              <ul class="menu mx-auto flex w-40 flex-col gap-2">
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
          <button class="rounded-btn mx-2 mt-auto p-3 hover:bg-base-300" @click="logout">
            {{ $t("global.sign_out") }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { useLabelStore } from "~~/stores/labels";
  import { useLocationStore } from "~~/stores/locations";
  import MdiMenu from "~icons/mdi/menu";
  import MdiPlus from "~icons/mdi/plus";

  import MdiHome from "~icons/mdi/home";
  import MdiFileTree from "~icons/mdi/file-tree";
  import MdiMagnify from "~icons/mdi/magnify";
  import MdiAccount from "~icons/mdi/account";
  import MdiCog from "~icons/mdi/cog";
  const username = computed(() => authCtx.user?.name || "User");

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
      name: "Home",
      to: "/home",
    },
    {
      icon: MdiFileTree,
      id: 4,
      active: computed(() => route.path === "/locations"),
      name: "Locations",
      to: "/locations",
    },
    {
      icon: MdiMagnify,
      id: 3,
      active: computed(() => route.path === "/items"),
      name: "Search",
      to: "/items",
    },
    {
      icon: MdiAccount,
      id: 1,
      active: computed(() => route.path === "/profile"),
      name: "Profile",
      to: "/profile",
    },
    {
      icon: MdiCog,
      id: 6,
      active: computed(() => route.path === "/tools"),
      name: "Tools",
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
