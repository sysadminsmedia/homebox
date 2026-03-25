import { useI18n } from "vue-i18n";
import type { UserClient } from "~~/lib/api/user";

type StatCard = {
  label: string;
  value: number;
  type: "currency" | "number";
};

export function statCardData(api: UserClient) {
  const { t } = useI18n();

  const { data: statistics } = useAsyncData(
    "statistics",
    async () => {
      const { data } = await api.stats.group();
      return data;
    },
    {
      deep: true,
    }
  );

  return computed(() => {
    return [
      {
        label: t("home.total_value"),
        value: statistics.value?.totalItemPrice || 0,
        type: "currency",
      },
      {
        label: t("home.total_items"),
        value: statistics.value?.totalItems || 0,
        type: "number",
      },
      {
        label: t("home.total_locations"),
        value: statistics.value?.totalLocations || 0,
        type: "number",
      },
      {
        label: t("home.total_tags"),
        value: statistics.value?.totalTags || 0,
        type: "number",
      },
    ] as StatCard[];
  });
}
