<script setup lang="ts">
  import { MoreHorizontal } from "lucide-vue-next";
  import { Button } from "@/components/ui/button";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
  } from "@/components/ui/dropdown-menu";
  import type { ItemSummary } from "~/lib/api/types/data-contracts";

  const props = defineProps<{
    item: ItemSummary | ItemSummary[];
  }>();

  defineEmits<{
    (e: "expand"): void;
  }>();

  const multiple = computed(() => {
    return Array.isArray(props.item);
  });

  function copy(id: string) {
    navigator.clipboard.writeText(id);
  }
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <Button variant="ghost" class="size-8 p-0">
        <span class="sr-only">Open menu</span>
        <MoreHorizontal class="size-4" />
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="end">
      <DropdownMenuLabel>Actions</DropdownMenuLabel>
      <DropdownMenuItem v-if="!multiple" as-child>
        <NuxtLink :to="`/item/${item.id}`" class="hover:underline"> View item </NuxtLink>
      </DropdownMenuItem>
      <DropdownMenuItem v-else @click="console.log('needs to be implemented')"> View items </DropdownMenuItem>
      <DropdownMenuItem @click="copy(payment.id)"> Copy payment ID </DropdownMenuItem>
      <DropdownMenuItem @click="$emit('expand')"> Expand </DropdownMenuItem>
      <DropdownMenuSeparator />
      <DropdownMenuItem>View customer</DropdownMenuItem>
      <DropdownMenuItem>View payment details</DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>
