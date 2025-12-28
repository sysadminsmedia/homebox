import { useI18n } from "vue-i18n";
import type { UserClient } from "~~/lib/api/user";

type StatCard = {
  tag: string;
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
        tag: t("home.total_value"),
        value: statistics.value?.totalItemPrice || 0,
        type: "currency",
      },
      {
        tag: t("home.total_items"),
        value: statistics.value?.totalItems || 0,
        type: "number",
      },
      {
        tag: t("home.total_locations"),
        value: statistics.value?.totalLocations || 0,
        type: "number",
      },
      {
        tag: t("home.total_tags"),
        value: statistics.value?.totalLabels || 0,
        type: "number",
      },
    ] as StatCard[];
  });
}
