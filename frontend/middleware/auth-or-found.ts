export default defineNuxtRouteMiddleware(async to => {
  const ctx = useAuthContext();
  const api = useUserApi();
  const redirectTo = useState("authRedirect");

  if (!ctx.isAuthorized()) {
    return navigateTo(`/found/${to.params.id}`);
  }

  if (!ctx.user) {
    const { data, error } = await api.user.self();
    if (error) {
      redirectTo.value = `/item/${to.params.id}`;
      return navigateTo("/");
    }

    ctx.user = data.item;
  }
});
