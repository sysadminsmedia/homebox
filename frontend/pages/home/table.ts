import type { UserClient } from "~~/lib/api/user";

export function itemsTable(api: UserClient) {
  const { data: items, refresh } = useAsyncData(
    "items",
    async () => {
      const { data } = await api.items.getAll({
        page: 1,
        pageSize: 5,
        orderBy: "createdAt",
      });
      return data.items;
    },
    {
      deep: true,
    }
  );

  onServerEvent(ServerEvent.ItemMutation, () => {
    console.log("item mutation");
    refresh();
  });

  return computed(() => {
    return {
      items: items.value || [],
    };
  });
}
