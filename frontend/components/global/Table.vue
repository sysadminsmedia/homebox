<template>
  <div class="overflow-x-auto">
    <table class="table w-full">
      <thead>
        <tr class="bg-primary">
          <th
            v-for="h in headers"
            :key="h.value"
            class="text-no-transform bg-neutral text-sm text-neutral-content"
            :class="{
              'text-center': h.align === 'center',
              'text-right': h.align === 'right',
              'text-left': h.align === 'left',
            }"
          >
            <template v-if="typeof h === 'string'">{{ h }}</template>
            <template v-else>{{ h.text }}</template>
          </th>
        </tr>
      </thead>
      <tbody>
        <!-- row 1 -->
        <tr v-for="(d, i) in data" :key="i">
          <td
            v-for="h in headers"
            :key="`${h.value}-${i}`"
            class="bg-base-100"
            :class="{
              'text-center': h.align === 'center',
              'text-right': h.align === 'right',
              'text-left': h.align === 'left',
            }"
          >
            <slot :name="cell(h)" v-bind="{ item: d }">
              {{ extractValue(d, h.value) }}
            </slot>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
  import type { TableData, TableHeader } from "./Table.types";

  type Props = {
    headers: TableHeader[];
    data: TableData[];
  };

  function extractValue(data: TableData, value: string) {
    const parts = value.split(".");
    let current = data;
    for (const part of parts) {
      current = current[part];
    }
    return current;
  }

  function cell(h: TableHeader) {
    return `cell-${h.value.replace(".", "_")}`;
  }

  defineProps<Props>();
</script>

<style scoped></style>
