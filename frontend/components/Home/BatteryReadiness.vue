<script setup lang="ts">
  import Subtitle from "~/components/global/Subtitle.vue";
  import BaseCard from "@/components/Base/Card.vue";
  import type { BatteryReadinessRow } from "~~/lib/api/types/data-contracts";

  const api = useUserApi();

  const { data: rows } = useAsyncData("battery-readiness", async () => {
    const { data } = await api.items.fields.batteryReadiness();
    return data ?? [];
  });

  type Status = { label: string; class: string };

  // Supply vs. threshold drives the status pill; theme-aware for light/dark.
  function status(row: BatteryReadinessRow): Status {
    const red = "bg-red-100 text-red-700 dark:bg-red-950 dark:text-red-300";
    const amber = "bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300";
    const green = "bg-green-100 text-green-700 dark:bg-green-950 dark:text-green-300";

    if (!row.hasMinStock) {
      // Devices depend on this type but no stock item exists.
      return { label: "no stock item", class: red };
    }
    if (row.stock <= 0) {
      return { label: "reorder", class: red };
    }
    if (row.stock <= row.minStock) {
      return { label: "low", class: amber };
    }
    return { label: "ok", class: green };
  }
</script>

<template>
  <section v-if="rows && rows.length > 0">
    <Subtitle>Battery Readiness</Subtitle>
    <BaseCard>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b text-left text-xs uppercase text-muted-foreground">
              <th class="px-4 py-2 font-medium">Type</th>
              <th class="px-4 py-2 text-right font-medium">In stock</th>
              <th class="px-4 py-2 text-right font-medium">Min</th>
              <th class="px-4 py-2 text-right font-medium">Devices</th>
              <th class="px-4 py-2 text-right font-medium">Status</th>
            </tr>
          </thead>
          <tbody class="divide-y">
            <tr v-for="row in rows" :key="row.type">
              <td class="px-4 py-2 font-medium">{{ row.type }}</td>
              <td class="px-4 py-2 text-right tabular-nums">{{ row.stock }}</td>
              <td class="px-4 py-2 text-right tabular-nums text-muted-foreground">
                {{ row.hasMinStock ? row.minStock : "—" }}
              </td>
              <td class="px-4 py-2 text-right tabular-nums">{{ row.deviceCount }}</td>
              <td class="px-4 py-2 text-right">
                <span class="rounded-full px-2.5 py-0.5 text-xs font-medium" :class="status(row).class">
                  {{ status(row).label }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </BaseCard>
  </section>
</template>
