function foundItemPath(pathname: string) {
  const itemMatch = pathname.match(/^\/item\/([^/]+)/);
  if (itemMatch) {
    return `/found/item/${encodeURIComponent(itemMatch[1]!)}`;
  }

  const assetMatch = pathname.match(/^\/(?:a|assets)\/([^/]+)/);
  if (assetMatch) {
    return `/found/asset/${encodeURIComponent(assetMatch[1]!)}`;
  }

  return null;
}

export default defineNuxtRouteMiddleware(async to => {
  const ctx = useAuthContext();
  const api = useUserApi();
  const redirectTo = useState("authRedirect");
  const currentPath = import.meta.client ? window.location.pathname : to.path;

  if (!ctx.isAuthorized()) {
    const foundPath = foundItemPath(to.path);
    if (foundPath) {
      return navigateTo(foundPath);
    }

    if (currentPath !== "/") {
      console.debug("[middleware/auth] isAuthorized returned false, redirecting to /");
      redirectTo.value = currentPath;
      return navigateTo("/");
    }
  }

  if (!ctx.user) {
    console.log("Fetching user data");
    const { data, error } = await api.user.self();
    if (error) {
      if (currentPath !== "/") {
        console.debug("[middleware/user] user is null and fetch failed, redirecting to /");
        redirectTo.value = currentPath;
        return navigateTo("/");
      }
    }

    ctx.user = data.item;
  }
});
