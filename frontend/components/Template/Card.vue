<script setup lang="ts">
  import { toast } from "@/components/ui/sonner";
  import MdiPencil from "~icons/mdi/pencil";
  import MdiDelete from "~icons/mdi/delete";
  import { Card, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
  import { Button, ButtonGroup } from "@/components/ui/button";
  import type { ItemTemplateSummary } from "~/lib/api/types/data-contracts";

  const props = defineProps<{
    template: ItemTemplateSummary;
  }>();

  const emit = defineEmits<{
    deleted: [];
  }>();

  const api = useUserApi();
  const confirm = useConfirm();

  async function handleDelete() {
    const { isCanceled } = await confirm.open("Delete this template?");
    if (isCanceled) return;

    const { error } = await api.templates.delete(props.template.id);
    if (error) {
      toast.error("Failed to delete template");
      return;
    }

    toast.success("Template deleted");
    emit("deleted");
  }
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle class="truncate">{{ template.name }}</CardTitle>
      <CardDescription v-if="template.description" class="line-clamp-2">
        {{ template.description }}
      </CardDescription>
    </CardHeader>
    <CardFooter class="flex justify-end gap-2">
      <ButtonGroup>
        <Button size="sm" variant="outline" as-child>
          <NuxtLink :to="`/template/${template.id}`">
            <MdiPencil class="size-4" />
          </NuxtLink>
        </Button>
        <Button size="sm" variant="destructive" @click="handleDelete">
          <MdiDelete class="size-4" />
        </Button>
      </ButtonGroup>
    </CardFooter>
  </Card>
</template>
