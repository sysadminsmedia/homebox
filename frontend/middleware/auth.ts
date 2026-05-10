export default defineNuxtRouteMiddleware(async to => {
  const ctx = useAuthContext();
  const api = useUserApi();
  const redirectTo = useState("authRedirect");

  if (!ctx.isAuthorized()) {
    const foundPath = foundLabelPath(to.path);
    if (foundPath) {
      return navigateTo(foundPath);
    }

    if (to.path !== "/") {
      console.debug("[middleware/auth] isAuthorized returned false, redirecting to /");
      redirectTo.value = to.path;
      return navigateTo("/");
    }
  }

  if (!ctx.user) {
    console.log("Fetching user data");
    const { data, error } = await api.user.self();
    if (error) {
      if (to.path !== "/") {
        console.debug("[middleware/user] user is null and fetch failed, redirecting to /");
        redirectTo.value = to.path;
        return navigateTo("/");
      }
    }

    ctx.user = data.item;
  }
});

function foundLabelPath(path: string): string | null {
  const itemMatch = path.match(/^\/item\/([^/]+)/);
  if (itemMatch) {
    return `/found/item/${encodeURIComponent(itemMatch[1]!)}`;
  }

  const assetMatch = path.match(/^\/(?:a|assets)\/([^/]+)/);
  if (assetMatch) {
    return `/found/asset/${encodeURIComponent(assetMatch[1]!)}`;
  }

  return null;
}
