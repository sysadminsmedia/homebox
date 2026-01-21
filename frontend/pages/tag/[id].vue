<script setup lang="ts">
  import { useI18n } from "vue-i18n";
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
  import ColorSelector from "@/components/Form/ColorSelector.vue";
  import { getContrastTextColor } from "~/lib/utils";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import FormTextField from "~/components/Form/TextField.vue";
  import FormTextArea from "~/components/Form/TextArea.vue";
  import BaseContainer from "@/components/Base/Container.vue";
  import Currency from "~/components/global/Currency.vue";
  import DateTime from "~/components/global/DateTime.vue";
  import PageQRCode from "~/components/global/PageQRCode.vue";
  import Markdown from "~/components/global/Markdown.vue";
  import ItemViewSelectable from "~/components/Item/View/Selectable.vue";

  definePageMeta({
    middleware: ["auth"],
  });

  const { t } = useI18n();

  const { openDialog, closeDialog } = useDialog();

  const route = useRoute();
  const api = useUserApi();

  const tagId = computed<string>(() => route.params.id as string);

  const { data: tag } = useAsyncData(tagId.value, async () => {
    const { data, error } = await api.tags.get(tagId.value);
    if (error) {
      toast.error(t("tags.toast.failed_load_tag"));
      navigateTo("/home");
      return;
    }
    return data;
  });

  const confirm = useConfirm();

  async function confirmDelete() {
    const { isCanceled } = await confirm.open(t("tags.tag_delete_confirm"));

    if (isCanceled) {
      return;
    }

    const { error } = await api.tags.delete(tagId.value);

    if (error) {
      toast.error(t("tags.toast.failed_delete_tag"));
      return;
    }
    toast.success(t("tags.toast.tag_deleted"));
    navigateTo("/home");
  }

  const updating = ref(false);
  const updateData = reactive({
    name: "",
    description: "",
    color: "",
  });

  function openUpdate() {
    updateData.name = tag.value?.name || "";
    updateData.description = tag.value?.description || "";
    updateData.color = "";
    openDialog(DialogID.UpdateTag);
  }

  async function update() {
    updating.value = true;
    const { error, data } = await api.tags.update(tagId.value, updateData);

    if (error) {
      updating.value = false;
      toast.error(t("tags.toast.failed_update_tag"));
      return;
    }

    toast.success(t("tags.toast.tag_updated"));
    tag.value = data;
    closeDialog(DialogID.UpdateTag);
    updating.value = false;
  }

  const { data: items, refresh: refreshItemList } = useAsyncData(
    () => tagId.value + "_item_list",
    async () => {
      if (!tagId.value) {
        return {
          items: [],
          totalPrice: null,
        };
      }

      const resp = await api.items.getAll({
        tags: [tagId.value],
      });

      if (resp.error) {
        toast.error(t("items.toast.failed_load_items"));
        return {
          items: [],
          totalPrice: null,
        };
      }

      return resp.data;
    },
    {
      watch: [tagId],
    }
  );
</script>

<template>
  <!-- Update Dialog -->
  <Dialog :dialog-id="DialogID.UpdateTag">
    <DialogContent>
      <DialogHeader>
        <DialogTitle> {{ $t("tags.update_tag") }} </DialogTitle>
      </DialogHeader>

      <form v-if="tag" class="flex flex-col gap-2" @submit.prevent="update">
        <FormTextField
          v-model="updateData.name"
          :autofocus="true"
          :label="$t('components.tag.create_modal.tag_name')"
          :max-length="255"
          :min-length="1"
        />
        <FormTextArea
          v-model="updateData.description"
          :label="$t('components.tag.create_modal.tag_description')"
          :max-length="1000"
        />
        <ColorSelector
          v-model="updateData.color"
          :label="$t('components.tag.create_modal.tag_color')"
          :show-hex="true"
          :starting-color="tag.color"
        />
        <DialogFooter>
          <Button type="submit" :loading="updating"> {{ $t("global.update") }} </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>

  <BaseContainer v-if="tag">
    <!-- set page title -->
    <Title>{{ tag.name }}</Title>

    <Card class="p-3">
      <header :class="{ 'mb-2': tag.description }">
        <div class="flex flex-wrap items-end gap-2">
          <div
            class="mb-auto flex size-12 items-center justify-center rounded-full"
            :style="
              tag.color
                ? { backgroundColor: tag.color, color: getContrastTextColor(tag.color) }
                : { backgroundColor: 'hsl(var(--secondary))', color: 'hsl(var(--secondary-foreground))' }
            "
          >
            <MdiPackageVariant class="size-7" />
          </div>
          <div>
            <h1 class="flex items-center gap-3 pb-1 text-2xl">
              {{ tag ? tag.name : "" }}
              <Badge v-if="items && items.totalPrice" variant="secondary" class="ml-2">
                <Currency :amount="items.totalPrice" />
              </Badge>
            </h1>
            <div class="flex flex-wrap gap-1 text-xs">
              <div>
                {{ $t("global.created") }}
                <DateTime :date="tag?.createdAt" />
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
      <Separator v-if="tag && tag.description" />
      <Markdown v-if="tag && tag.description" class="mt-3 text-base" :source="tag.description" />
    </Card>
    <section v-if="tag && items">
      <ItemViewSelectable :items="items.items" @refresh="refreshItemList" />
    </section>
  </BaseContainer>
</template>
