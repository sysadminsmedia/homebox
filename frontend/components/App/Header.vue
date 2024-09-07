<script lang="ts" setup>
  import MdiPlus from "~icons/mdi/plus";

  const ctx = useAuthContext();
  const api = useUserApi();

  async function logout() {
    const { error } = await ctx.logout(api);
    if (error) {
      return;
    }

    navigateTo("/");
  }

  const links = [
    {
      name: "Home",
      href: "/home",
    },
    {
      name: "Items",
      href: "/items",
    },
    {
      name: "Logout",
      action: logout,
      last: true,
    },
  ];

  const modals = reactive({
    item: false,
    location: false,
    label: false,
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
</script>

<template>
  <!--
    Confirmation Modal is a singleton used by all components so we render
    it here to ensure it's always available. Possibly could move this further
    up the tree
   -->
  <ModalConfirm />
  <ItemCreateModal v-model="modals.item" />
  <LabelCreateModal v-model="modals.label" />
  <LocationCreateModal v-model="modals.location" />

  <div class="absolute top-0 -z-10 h-80 max-h-96 w-full bg-neutral shadow-xl"></div>

  <BaseContainer cmp="header" class="max-w-none py-6">
    <BaseContainer>
      <NuxtLink to="/home">
        <h2 class="mt-1 flex text-4xl font-bold tracking-tight text-neutral-content sm:text-5xl lg:text-6xl">
          HomeB
          <AppLogo class="-mb-4 w-12" />
          x
        </h2>
      </NuxtLink>
      <div class="ml-1 mt-2 space-x-2 text-lg text-neutral-content/75">
        <template v-for="link in links">
          <NuxtLink
            v-if="!link.action"
            :key="link.name"
            class="italic transition-colors duration-200 hover:text-base-content"
            :to="link.href"
          >
            {{ link.name }}
          </NuxtLink>
          <button
            v-else
            :key="link.name + 'link'"
            for="location-form-modal"
            class="italic transition-colors duration-200 hover:text-base-content"
            @click="link.action"
          >
            {{ link.name }}
          </button>
          <span v-if="!link.last" :key="link.name"> / </span>
        </template>
      </div>
      <div class="mt-6 flex">
        <div class="dropdown">
          <label tabindex="0" class="btn btn-primary btn-sm">
            <span>
              <MdiPlus class="-ml-1 mr-1" />
            </span>
            Create
          </label>
          <ul tabindex="0" class="dropdown-content menu rounded-box w-52 bg-base-100 p-2 shadow">
            <li v-for="btn in dropdown" :key="btn.name">
              <button @click="btn.action">
                {{ btn.name }}
              </button>
            </li>
          </ul>
        </div>
      </div>
    </BaseContainer>
  </BaseContainer>
</template>
