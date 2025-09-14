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
    const code = urlParams.get("code");
    const state = urlParams.get("state");
    
    if (!code) {
      console.error("No authorization code received");
      navigateTo("/");
      return;
    }

    // For OIDC, we need to get the issuer from the provider configuration
    // The issuer will be determined by the backend from the ID token
    const { error } = await ctx.loginOauth(api, "oidc", "", code, state);

    if (error) {
      console.warn(error);
      navigateTo("/");
      return;
    }

    navigateTo(redirectTo.value || "/home");
    redirectTo.value = null;
  }

  onMounted(() => {
    login();
  });
</script>

<template></template>

<style scoped></style>
