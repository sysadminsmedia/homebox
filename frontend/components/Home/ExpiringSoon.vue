<script setup lang="ts">
  import Subtitle from "~/components/global/Subtitle.vue";
  import BaseCard from "@/components/Base/Card.vue";

  const api = useUserApi();

  const { data: items } = useAsyncData("expiring-soon", async () => {
    const { data } = await api.items.fields.expiring(30);
    return data ?? [];
  });

  function daysUntil(dateStr: string): number {
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const d = new Date(`${dateStr}T00:00:00`);
    return Math.round((d.getTime() - today.getTime()) / 86_400_000);
  }

  function relativeLabel(n: number): string {
    if (n < 0) {
      return `${Math.abs(n)} day${Math.abs(n) === 1 ? "" : "s"} overdue`;
    }
    if (n === 0) {
      return "due today";
    }
    if (n === 1) {
      return "due tomorrow";
    }
    return `in ${n} days`;
  }

  // Urgency drives the pill colour; theme-aware for light/dark.
  function urgencyClass(n: number): string {
    if (n < 0) {
      return "bg-red-100 text-red-700 dark:bg-red-950 dark:text-red-300";
    }
    if (n <= 7) {
      return "bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300";
    }
    return "bg-muted text-muted-foreground";
  }
</script>

<template>
  <section v-if="items && items.length > 0">
    <Subtitle>Coming Due</Subtitle>
    <BaseCard>
      <ul class="divide-y">
        <li v-for="(row, i) in items" :key="`${row.id}-${i}`" class="flex items-center justify-between gap-3 px-4 py-3">
          <div class="min-w-0">
            <NuxtLink :to="`/item/${row.id}`" class="block truncate font-medium hover:underline">
              {{ row.name }}
            </NuxtLink>
            <span class="text-xs text-muted-foreground">{{ row.fieldName }} · {{ row.date }}</span>
          </div>
          <span class="shrink-0 rounded-full px-2.5 py-0.5 text-xs font-medium" :class="urgencyClass(daysUntil(row.date))">
            {{ relativeLabel(daysUntil(row.date)) }}
          </span>
        </li>
      </ul>
    </BaseCard>
  </section>
</template>
