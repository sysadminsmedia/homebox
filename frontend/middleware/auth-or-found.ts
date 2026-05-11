export default defineNuxtRouteMiddleware(async to => {
  const ctx = useAuthContext();
  const api = useUserApi();
  const redirectTo = useState("authRedirect");
  const routeId = Array.isArray(to.params.id) ? to.params.id[0] : to.params.id;

  if (!routeId || typeof routeId !== "string") {
    return navigateTo("/");
  }

  if (!ctx.isAuthorized()) {
    return navigateTo(`/found/${encodeURIComponent(routeId)}`);
  }

  if (!ctx.user) {
    const { data, error } = await api.user.self();
    if (error) {
      redirectTo.value = `/item/${encodeURIComponent(routeId)}`;
      return navigateTo("/");
    }

    ctx.user = data.item;
  }
});
