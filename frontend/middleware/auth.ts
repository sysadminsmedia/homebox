export default defineNuxtRouteMiddleware(async () => {
  const ctx = useAuthContext();
  const api = useUserApi();
  const redirectTo = useState("authRedirect");

  if (!ctx.isAuthorized()) {
    if (window.location.pathname !== "/") {
      console.debug("[middleware/auth] isAuthorized returned false, redirecting to /");
      redirectTo.value = window.location.pathname;
      return navigateTo("/");
    }
  }

  if (!ctx.user) {
    console.log("Fetching user data");
    const { data, error } = await api.user.self();
    if (error) {
      if (window.location.pathname !== "/") {
        console.debug("[middleware/user] user is null and fetch failed, redirecting to /");
        redirectTo.value = window.location.pathname;
        return navigateTo("/");
      }
    }

    ctx.user = data.item;
  }
});
