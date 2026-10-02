export default defineNuxtRouteMiddleware(async to => {
  const ctx = useAuthContext();
  const api = useUserApi();
  const redirectTo = useState("authRedirect");

  if (!ctx.isAuthorized()) {
    const path = to.path;

    const asset = path.match(/^\/(?:a|assets)\/([^/]+)\/?$/);
    const item = path.match(/^\/item\/([^/]+)\/?$/);
    if (asset || item) {
      console.debug("[middleware/auth] unauthenticated label scan, redirecting to found page");
      redirectTo.value = path;
      const kind = item ? "item" : "asset";
      const match = item ?? asset;
      const id = match?.[1];
      return navigateTo(`/found/${kind}/${id}`);
    }

    if (path !== "/") {
      console.debug("[middleware/auth] isAuthorized returned false, redirecting to /");
      redirectTo.value = path;
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
