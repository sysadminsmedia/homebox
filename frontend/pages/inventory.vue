<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import type { ItemSummary } from "~~/lib/api/types/data-contracts";
  import BaseContainer from "@/components/Base/Container.vue";
  import { Button } from "@/components/ui/button";
  import { Badge } from "@/components/ui/badge";
  import MdiCheckCircle from "~icons/mdi/check-circle";
  import MdiAlertCircle from "~icons/mdi/alert-circle";
  import MdiClockAlert from "~icons/mdi/clock-alert";
  import MdiLoading from "~icons/mdi/loading";

  const { t, d } = useI18n();
  const api = useUserApi();
  const { isManagerOrAbove } = usePermissions();

  definePageMeta({
    middleware: ["auth"],
  });

  useHead({
    title: "HomeBox | 盤點管理",
  });

  // Fetch all items (non-archived)
  const { data: itemsResp, refresh } = await useAsyncData("inventory-items", () =>
    api.items.getAll({ pageSize: 5000 })
  );

  const items = computed(() => itemsResp.value?.data?.items ?? []);

  // Compute days since last inventory for each item
  function daysSinceInventory(item: ItemSummary & { lastInventoryAt?: string | null }): number | null {
    if (!item.lastInventoryAt) return null;
    const diff = Date.now() - new Date(item.lastInventoryAt).getTime();
    return Math.floor(diff / (1000 * 60 * 60 * 24));
  }

  type InventoryStatus = "ok" | "warning" | "overdue" | "never";

  function inventoryStatus(item: ItemSummary & { lastInventoryAt?: string | null }): InventoryStatus {
    const days = daysSinceInventory(item);
    if (days === null) return "never";
    if (days <= 30) return "ok";
    if (days <= 90) return "warning";
    return "overdue";
  }

  const statusColors: Record<InventoryStatus, string> = {
    ok: "bg-green-100 text-green-800",
    warning: "bg-yellow-100 text-yellow-800",
    overdue: "bg-red-100 text-red-800",
    never: "bg-gray-100 text-gray-600",
  };

  const statusLabels: Record<InventoryStatus, string> = {
    ok: "30天內已清點",
    warning: "30-90天未清點",
    overdue: "超過90天未清點",
    never: "從未清點",
  };

  // Filter state
  const filterStatus = ref<InventoryStatus | "all">("all");

  const filteredItems = computed(() => {
    if (filterStatus.value === "all") return items.value;
    return items.value.filter(item => inventoryStatus(item as any) === filterStatus.value);
  });

  // Sorted by overdue first
  const sortedItems = computed(() =>
    [...filteredItems.value].sort((a, b) => {
      const order: InventoryStatus[] = ["never", "overdue", "warning", "ok"];
      return order.indexOf(inventoryStatus(a as any)) - order.indexOf(inventoryStatus(b as any));
    })
  );

  const checkingId = ref<string | null>(null);

  async function markInventoried(id: string) {
    if (!isManagerOrAbove.value) return;
    checkingId.value = id;
    try {
      const resp = await api.items.inventoryCheck(id);
      if (resp.error) {
        toast.error("盤點標記失敗");
        return;
      }
      toast.success("已標記為今日清點");
      await refresh();
    } finally {
      checkingId.value = null;
    }
  }

  // Summary stats
  const stats = computed(() => {
    const all = items.value;
    return {
      total: all.length,
      ok: all.filter(i => inventoryStatus(i as any) === "ok").length,
      warning: all.filter(i => inventoryStatus(i as any) === "warning").length,
      overdue: all.filter(i => inventoryStatus(i as any) === "overdue").length,
      never: all.filter(i => inventoryStatus(i as any) === "never").length,
    };
  });
</script>

<template>
  <BaseContainer>
    <div class="mb-6">
      <h1 class="text-2xl font-bold mb-1">盤點管理</h1>
      <p class="text-muted-foreground text-sm">追蹤所有資產的最後清點日期，確保定期盤點資產在位情況</p>
    </div>

    <!-- Summary Stats -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
      <div class="rounded-lg border p-4 text-center cursor-pointer" @click="filterStatus = 'all'">
        <div class="text-2xl font-bold">{{ stats.total }}</div>
        <div class="text-sm text-muted-foreground">全部資產</div>
      </div>
      <div class="rounded-lg border p-4 text-center bg-green-50 cursor-pointer" @click="filterStatus = 'ok'">
        <div class="text-2xl font-bold text-green-700">{{ stats.ok }}</div>
        <div class="text-sm text-green-600">30天內已清點</div>
      </div>
      <div class="rounded-lg border p-4 text-center bg-yellow-50 cursor-pointer" @click="filterStatus = 'warning'">
        <div class="text-2xl font-bold text-yellow-700">{{ stats.warning }}</div>
        <div class="text-sm text-yellow-600">30-90天未清點</div>
      </div>
      <div class="rounded-lg border p-4 text-center bg-red-50 cursor-pointer" @click="filterStatus = 'overdue'">
        <div class="text-2xl font-bold text-red-700">{{ stats.overdue + stats.never }}</div>
        <div class="text-sm text-red-600">需要清點</div>
      </div>
    </div>

    <!-- Filter Buttons -->
    <div class="flex gap-2 mb-4 flex-wrap">
      <Button variant="outline" :class="filterStatus === 'all' ? 'border-primary' : ''" size="sm" @click="filterStatus = 'all'">
        全部 ({{ stats.total }})
      </Button>
      <Button variant="outline" :class="filterStatus === 'never' ? 'border-primary' : ''" size="sm" @click="filterStatus = 'never'">
        從未清點 ({{ stats.never }})
      </Button>
      <Button variant="outline" :class="filterStatus === 'overdue' ? 'border-primary' : ''" size="sm" @click="filterStatus = 'overdue'">
        逾期 ({{ stats.overdue }})
      </Button>
      <Button variant="outline" :class="filterStatus === 'warning' ? 'border-primary' : ''" size="sm" @click="filterStatus = 'warning'">
        即將逾期 ({{ stats.warning }})
      </Button>
      <Button variant="outline" :class="filterStatus === 'ok' ? 'border-primary' : ''" size="sm" @click="filterStatus = 'ok'">
        正常 ({{ stats.ok }})
      </Button>
    </div>

    <!-- Items Table -->
    <div class="rounded-lg border overflow-hidden">
      <table class="w-full text-sm">
        <thead class="bg-muted/50">
          <tr>
            <th class="text-left px-4 py-3 font-medium">資產名稱</th>
            <th class="text-left px-4 py-3 font-medium hidden md:table-cell">存放位置</th>
            <th class="text-left px-4 py-3 font-medium">最後清點日期</th>
            <th class="text-left px-4 py-3 font-medium">狀態</th>
            <th v-if="isManagerOrAbove" class="text-right px-4 py-3 font-medium">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="item in sortedItems"
            :key="item.id"
            class="border-t hover:bg-muted/30 transition-colors"
          >
            <td class="px-4 py-3">
              <NuxtLink :to="`/item/${item.id}`" class="font-medium hover:underline">
                {{ item.name }}
              </NuxtLink>
            </td>
            <td class="px-4 py-3 hidden md:table-cell text-muted-foreground">
              {{ item.location?.name ?? "—" }}
            </td>
            <td class="px-4 py-3 text-muted-foreground">
              <span v-if="(item as any).lastInventoryAt">
                {{ d(new Date((item as any).lastInventoryAt), "short") }}
                <span class="text-xs ml-1">({{ daysSinceInventory(item as any) }} 天前)</span>
              </span>
              <span v-else class="text-gray-400">從未清點</span>
            </td>
            <td class="px-4 py-3">
              <span
                class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium"
                :class="statusColors[inventoryStatus(item as any)]"
              >
                <MdiCheckCircle v-if="inventoryStatus(item as any) === 'ok'" class="h-3 w-3" />
                <MdiAlertCircle v-else-if="inventoryStatus(item as any) === 'warning'" class="h-3 w-3" />
                <MdiClockAlert v-else class="h-3 w-3" />
                {{ statusLabels[inventoryStatus(item as any)] }}
              </span>
            </td>
            <td v-if="isManagerOrAbove" class="px-4 py-3 text-right">
              <Button
                size="sm"
                variant="outline"
                :disabled="checkingId === item.id"
                @click="markInventoried(item.id)"
              >
                <MdiLoading v-if="checkingId === item.id" class="h-4 w-4 animate-spin mr-1" />
                標記已清點
              </Button>
            </td>
          </tr>
          <tr v-if="sortedItems.length === 0">
            <td colspan="5" class="px-4 py-8 text-center text-muted-foreground">無符合條件的資產</td>
          </tr>
        </tbody>
      </table>
    </div>
  </BaseContainer>
</template>
