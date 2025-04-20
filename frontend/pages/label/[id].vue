<script setup lang="ts">
  import { toast } from "@/components/ui/sonner";
  import MdiPackageVariant from "~icons/mdi/package-variant";
  import MdiPencil from "~icons/mdi/pencil";
  import MdiDelete from "~icons/mdi/delete";
  import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { Card } from "@/components/ui/card";
  import { Button } from "@/components/ui/button";
  import { Badge } from "@/components/ui/badge";
  import { Separator } from "@/components/ui/separator";

  definePageMeta({
    middleware: ["auth"],
  });

  const { openDialog, closeDialog } = useDialog();

  const route = useRoute();
  const api = useUserApi();

  const labelId = computed<string>(() => route.params.id as string);

  const { data: label } = useAsyncData(labelId.value, async () => {
    const { data, error } = await api.labels.get(labelId.value);
    if (error) {
      toast.error("Failed to load label");
      navigateTo("/home");
      return;
    }
    return data;
  });

  const confirm = useConfirm();

  async function confirmDelete() {
    const { isCanceled } = await confirm.open(
      "Are you sure you want to delete this label? This action cannot be undone."
    );

    if (isCanceled) {
      return;
    }

    const { error } = await api.labels.delete(labelId.value);

    if (error) {
      toast.error("Failed to delete label");
      return;
    }
    toast.success("Label deleted");
    navigateTo("/home");
  }

  const updating = ref(false);
  const updateData = reactive({
    name: "",
    description: "",
    color: "",
  });

  function openUpdate() {
    updateData.name = label.value?.name || "";
    updateData.description = label.value?.description || "";
    updateData.color = "";
    openDialog("update-label");
  }

  async function update() {
    updating.value = true;
    const { error, data } = await api.labels.update(labelId.value, updateData);

    if (error) {
      updating.value = false;
      toast.error("Failed to update label");
      return;
    }

    toast.success("Label updated");
    label.value = data;
    closeDialog("update-label");
    updating.value = false;
  }

  const items = computedAsync(async () => {
    if (!label.value) {
      return {
        items: [],
        totalPrice: null,
      };
    }

    const resp = await api.items.getAll({
      labels: [label.value.id],
    });

    if (resp.error) {
      toast.error("Failed to load items");
      return {
        items: [],
        totalPrice: null,
      };
    }

    return resp.data;
  });
</script>

<template>
  <!-- Update Dialog -->
  <Dialog dialog-id="update-label">
    <DialogContent>
      <DialogHeader>
        <DialogTitle> {{ $t("labels.update_label") }} </DialogTitle>
      </DialogHeader>

      <form v-if="label" class="flex flex-col gap-2" @submit.prevent="update">
        <FormTextField
          v-model="updateData.name"
          :autofocus="true"
          :label="$t('components.label.create_modal.label_name')"
          :max-length="255"
          :min-length="1"
        />
        <FormTextArea
          v-model="updateData.description"
          :label="$t('components.label.create_modal.label_description')"
          :max-length="255"
        />
        <!-- TODO: color  -->
        <DialogFooter>
          <Button type="submit" :loading="updating"> {{ $t("global.update") }} </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>

  <BaseContainer v-if="label">
    <Card class="p-3">
      <header :class="{ 'mb-2': label.description }">
        <div class="flex flex-wrap items-end gap-2">
          <div
            class="mb-auto flex size-12 items-center justify-center rounded-full bg-neutral-focus text-neutral-content"
          >
            <MdiPackageVariant class="size-7" />
          </div>
          <div>
            <h1 class="flex items-center gap-3 pb-1 text-2xl">
              {{ label ? label.name : "" }}
              <Badge v-if="items && items.totalPrice" variant="secondary" class="ml-2">
                <Currency :amount="items.totalPrice" />
              </Badge>
            </h1>
            <div class="flex flex-wrap gap-1 text-xs">
              <div>
                {{ $t("global.created") }}
                <DateTime :date="label?.createdAt" />
              </div>
            </div>
          </div>
          <div class="ml-auto mt-2 flex flex-wrap items-center justify-between gap-3">
            <PageQRCode />
            <Button @click="openUpdate">
              <MdiPencil />
              {{ $t("global.edit") }}
            </Button>
            <Button variant="destructive" @click="confirmDelete()">
              <MdiDelete />
              {{ $t("global.delete") }}
            </Button>
          </div>
        </div>
      </header>
      <Separator v-if="label && label.description" />
      <Markdown v-if="label && label.description" class="mt-3 text-base" :source="label.description" />
    </Card>
    <section v-if="label && items">
      <ItemViewSelectable :items="items.items" />
    </section>
  </BaseContainer>
</template>
