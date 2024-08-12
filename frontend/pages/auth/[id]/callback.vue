<script setup lang="ts">
  definePageMeta({
    layout: "empty",
    middleware: [
      () => {
        const ctx = useAuthContext();
        if (ctx.isAuthorized()) {
          return "/home";
        }
      },
    ],
  });

  const ctx = useAuthContext();
  const api = usePublicApi();
  const redirectTo = useState("authRedirect");

  async function login() {
    const urlParams = new URLSearchParams(window.location.search);
    const { error } = await ctx.loginOauth(api, "oidc", urlParams.get("iss")!, urlParams.get("code")!);

    if (error) {
      console.warn(error);
      navigateTo("/");
      return;
    }

    navigateTo(redirectTo || "/home");
    redirectTo.value = null;
  }

  onMounted(() => {
    login();
  });
</script>

<template></template>

<style scoped></style>
