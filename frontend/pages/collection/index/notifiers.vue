<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import { Button } from "@/components/ui/button";
  import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
  import { badgeVariants } from "@/components/ui/badge";
  import FormCheckbox from "~/components/Form/Checkbox.vue";
  import FormTextField from "~/components/Form/TextField.vue";
  import DateTime from "@/components/global/DateTime.vue";
  import MdiMegaphone from "~icons/mdi/megaphone";
  import MdiDelete from "~icons/mdi/delete";
  import MdiPencil from "~icons/mdi/pencil";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import type { NotifierCreate, NotifierOut } from "~~/lib/api/types/data-contracts";

  definePageMeta({
    middleware: ["auth"],
  });

  const { t } = useI18n();

    useHead({ title: `HomeBox | ${t("collection.tabs.notifiers")}` });

  const api = useUserApi();
  const confirm = useConfirm();
  const { openDialog, closeDialog } = useDialog();

  const notifiers = useAsyncData(async () => {
    const { data } = await api.notifiers.getAll();
    return data;
  });

  const targetID = ref("");
  const notifier = ref<NotifierCreate | null>(null);

  const openNotifierDialog = (v: NotifierOut | null) => {
    if (v) {
      targetID.value = v.id;
      notifier.value = {
        name: v.name,
        url: v.url,
        isActive: v.isActive,
      };
    } else {
      targetID.value = "";
      notifier.value = {
        name: "",
        url: "",
        isActive: true,
      };
    }

    openDialog(DialogID.CreateNotifier);
  };

  const createNotifier = async () => {
    if (!notifier.value) return;

    if (targetID.value) {
      await editNotifier();
      return;
    }

    const result = await api.notifiers.create({
      name: notifier.value.name,
      url: notifier.value.url || "",
      isActive: notifier.value.isActive,
    });

    if (result.error) {
      toast.error(t("profile.toast.failed_create_notifier"));
    }

    notifier.value = null;
    closeDialog(DialogID.CreateNotifier);
    await notifiers.refresh();
  };

  const editNotifier = async () => {
    if (!notifier.value) return;

    const result = await api.notifiers.update(targetID.value, {
      name: notifier.value.name,
      url: notifier.value.url || "",
      isActive: notifier.value.isActive,
    });

    if (result.error) {
      toast.error(t("profile.toast.failed_update_notifier"));
    }

    notifier.value = null;
    targetID.value = "";
    closeDialog(DialogID.CreateNotifier);
    await notifiers.refresh();
  };

  const deleteNotifier = async (id: string) => {
    const result = await confirm.open(t("profile.delete_notifier_confirm"));

    if (result.isCanceled) return;

    const { error } = await api.notifiers.delete(id);

    if (error) {
      toast.error(t("profile.toast.failed_delete_notifier"));
      return;
    }

    await notifiers.refresh();
  };

  const testNotifier = async () => {
    if (!notifier.value) return;

    const { error } = await api.notifiers.test(notifier.value.url);

    if (error) {
      toast.error(t("profile.toast.failed_test_notifier"));
      return;
    }

    toast.success(t("profile.toast.notifier_test_success"));
  };
</script>

<template>
  <div class="space-y-4">
    <Dialog :dialog-id="DialogID.CreateNotifier">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ $t("profile.notifier_modal", { type: notifier != null }) }}</DialogTitle>
        </DialogHeader>

        <form @submit.prevent="createNotifier">
          <template v-if="notifier">
            <FormTextField v-model="notifier.name" :label="$t('global.name')" class="mb-2" />
            <FormTextField v-model="notifier.url" :label="$t('profile.url')" class="mb-2" />
            <div class="max-w-[100px]">
              <FormCheckbox v-model="notifier.isActive" :label="$t('profile.enabled')" />
            </div>
          </template>
          <div class="mt-4 flex justify-between gap-2">
            <DialogFooter class="flex w-full">
              <Button variant="secondary" :disabled="!(notifier && notifier.url)" type="button" @click="testNotifier">
                {{ $t("profile.test") }}
              </Button>
              <div class="grow" />
              <Button type="submit">{{ $t("global.submit") }}</Button>
            </DialogFooter>
          </div>
        </form>
      </DialogContent>
    </Dialog>

    <div class="rounded-md border bg-card p-4">
      <header class="mb-2 flex items-center gap-2">
        <MdiMegaphone class="-mt-1" />
        <div>
          <h2 class="text-lg font-semibold">{{ $t("profile.notifiers") }}</h2>
          <p class="text-sm text-muted-foreground">{{ $t("profile.notifiers_sub") }}</p>
        </div>
      </header>

      <div v-if="notifiers.data.value" class="mx-1 divide-y rounded-md border">
        <p v-if="notifiers.data.value.length === 0" class="p-2 text-center text-sm">
          {{ $t("profile.no_notifiers") }}
        </p>
        <article v-for="n in notifiers.data.value" v-else :key="n.id" class="p-2">
          <div class="flex flex-wrap items-center gap-2">
            <p class="mr-auto text-lg">{{ n.name }}</p>
            <TooltipProvider :delay-duration="0" class="flex justify-end gap-2">
              <Tooltip>
                <TooltipTrigger>
                  <Button variant="destructive" size="icon" @click="deleteNotifier(n.id)">
                    <MdiDelete />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>{{ $t("global.delete") }}</TooltipContent>
              </Tooltip>
              <Tooltip>
                <TooltipTrigger>
                  <Button variant="outline" size="icon" @click="openNotifierDialog(n)">
                    <MdiPencil />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>{{ $t("global.edit") }}</TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
          <div class="flex flex-wrap justify-between py-1 text-sm">
            <p>
              <span v-if="n.isActive" :class="badgeVariants()">{{ $t("profile.active") }}</span>
              <span v-else :class="badgeVariants({ variant: 'destructive' })">{{ $t("profile.inactive") }}</span>
            </p>
            <p>
              {{ $t("global.created") }}
              <DateTime format="relative" datetime-type="time" :date="n.createdAt" />
            </p>
          </div>
        </article>
      </div>

      <div class="mt-4">
        <Button variant="secondary" size="sm" @click="openNotifierDialog(null)">
          {{ $t("global.create") }}
        </Button>
      </div>
    </div>
  </div>
</template>
