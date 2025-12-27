<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import type { LocationSummary, LocationUpdate } from "~~/lib/api/types/data-contracts";
  import { useLocationStore } from "~~/stores/locations";
  import MdiLoading from "~icons/mdi/loading";
  import MdiPackageVariant from "~icons/mdi/package-variant";
  import MdiPencil from "~icons/mdi/pencil";
  import MdiDelete from "~icons/mdi/delete";
  import MdiLabel from "~icons/mdi/label";
  import MdiPrinterPos from "~icons/mdi/printer-pos";
  import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { Card } from "@/components/ui/card";
  import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbLink,
    BreadcrumbList,
    BreadcrumbSeparator,
  } from "@/components/ui/breadcrumb";
  import { Button } from "@/components/ui/button";
  import { Badge } from "@/components/ui/badge";
  import { Separator } from "@/components/ui/separator";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import FormTextField from "~/components/Form/TextField.vue";
  import FormTextArea from "~/components/Form/TextArea.vue";
  import LocationSelector from "~/components/Location/Selector.vue";
  import BaseContainer from "@/components/Base/Container.vue";
  import Currency from "~/components/global/Currency.vue";
  import DateTime from "~/components/global/DateTime.vue";
  import LabelMaker from "~/components/global/LabelMaker.vue";
  import Markdown from "~/components/global/Markdown.vue";
  import BaseSectionHeader from "@/components/Base/SectionHeader.vue";
  import ItemViewSelectable from "~/components/Item/View/Selectable.vue";
  import LocationCard from "~/components/Location/Card.vue";

  definePageMeta({
    middleware: ["auth"],
  });

  const { t } = useI18n();

  const { openDialog, closeDialog } = useDialog();

  const route = useRoute();
  const api = useUserApi();

  const locationId = computed<string>(() => route.params.id as string);

  const { data: location } = useAsyncData(locationId.value, async () => {
    const { data, error } = await api.locations.get(locationId.value);
    if (error) {
      toast.error(t("locations.toast.failed_load_location"));
      navigateTo("/home");
      return;
    }

    if (data.parent) {
      parent.value = locations.value.find(l => l.id === data.parent.id);
    }

    if (parent.value === undefined) {
      parent.value = data.parent;
    }

    return data;
  });

  const confirm = useConfirm();

  async function confirmDelete() {
    const { isCanceled } = await confirm.open(t("locations.location_items_delete_confirm"));
    if (isCanceled) {
      return;
    }

    const { error } = await api.locations.delete(locationId.value);
    if (error) {
      toast.error(t("locations.toast.failed_delete_location"));
      return;
    }

    toast.success(t("locations.toast.location_deleted"));
    navigateTo("/locations");
  }

  const updating = ref(false);
  const updateData = reactive<LocationUpdate>({
    id: locationId.value,
    name: "",
    description: "",
    parentId: null,
  });

  function openUpdate() {
    updateData.name = location.value?.name || "";
    updateData.description = location.value?.description || "";
    openDialog(DialogID.UpdateLocation);
  }

  async function update() {
    updating.value = true;
    updateData.parentId = parent.value?.id || null;
    const { error, data } = await api.locations.update(locationId.value, updateData);

    if (error) {
      updating.value = false;
      toast.error(t("locations.toast.failed_update_location"));
      return;
    }

    toast.success(t("locations.toast.location_updated"));
    location.value = data;
    closeDialog(DialogID.UpdateLocation);
    updating.value = false;
  }

  const locationStore = useLocationStore();
  const locations = computed(() => locationStore.allLocations);

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const parent = ref<LocationSummary | any>({});

  const { data: items, refresh: refreshItemList } = useAsyncData(
    () => locationId.value + "_item_list",
    async () => {
      if (!locationId.value) {
        return [];
      }

      const resp = await api.items.getAll({
        locations: [locationId.value],
      });

      if (resp.error) {
        toast.error(t("items.toast.failed_load_items"));
        return [];
      }

      return resp.data.items;
    },
    {
      watch: [locationId],
    }
  );

  // Quick Print functionality
  const { isQuickPrintAvailable, quickPrintLocations } = useQuickPrint();
  const isQuickPrinting = ref(false);

  async function handleQuickPrint() {
    isQuickPrinting.value = true;
    try {
      const success = await quickPrintLocations([locationId.value]);
      if (success) {
        toast.success(t("locations.quick_print_success"));
      } else {
        // Fallback to dialog if defaults not configured
        openDialog(DialogID.PrintLabelTemplate, { params: { locationIds: [locationId.value] } });
      }
    } catch {
      toast.error(t("locations.quick_print_failed"));
    } finally {
      isQuickPrinting.value = false;
    }
  }

  // Open the template-based print dialog for this location
  function openPrintLabelTemplate() {
    openDialog(DialogID.PrintLabelTemplate, { params: { locationIds: [locationId.value] } });
  }
</script>

<template>
  <div>
    <!-- Update Dialog -->
    <Dialog :dialog-id="DialogID.UpdateLocation">
      <DialogContent>
        <DialogHeader>
          <DialogTitle> {{ $t("locations.update_location") }} </DialogTitle>
        </DialogHeader>

        <form v-if="location" class="flex flex-col gap-2" @submit.prevent="update">
          <FormTextField
            v-model="updateData.name"
            :autofocus="true"
            :label="$t('components.location.create_modal.location_name')"
            :max-length="255"
            :min-length="1"
          />
          <FormTextArea
            v-model="updateData.description"
            :label="$t('components.location.create_modal.location_description')"
            :max-length="1000"
          />
          <LocationSelector v-model="parent" :current-location="location" />
          <DialogFooter>
            <Button type="submit" :disabled="updating">
              <MdiLoading v-if="updating" class="animate-spin" />
              {{ $t("global.update") }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <BaseContainer v-if="location">
      <!-- set page title -->
      <Title>{{ location.name }}</Title>

      <Card class="p-3">
        <header :class="{ 'mb-2': location?.description }">
          <div class="flex flex-wrap items-end gap-2">
            <div
              class="mb-auto flex size-12 items-center justify-center rounded-full bg-secondary text-secondary-foreground"
            >
              <MdiPackageVariant class="size-7" />
            </div>
            <div>
              <Breadcrumb v-if="location?.parent">
                <BreadcrumbList>
                  <BreadcrumbItem>
                    <BreadcrumbLink as-child class="text-foreground/70 hover:underline">
                      <NuxtLink :to="`/location/${location.parent.id}`">
                        {{ location.parent.name }}
                      </NuxtLink>
                    </BreadcrumbLink>
                  </BreadcrumbItem>
                  <BreadcrumbSeparator />
                  <BreadcrumbItem> {{ location.name }} </BreadcrumbItem>
                </BreadcrumbList>
              </Breadcrumb>
              <h1 class="flex items-center gap-3 pb-1 text-2xl">
                {{ location ? location.name : "" }}

                <Badge v-if="location && location.totalPrice" variant="secondary">
                  <Currency :amount="location.totalPrice" />
                </Badge>
              </h1>
              <div class="flex flex-wrap gap-1 text-xs">
                <div>
                  {{ $t("global.created") }}
                  <DateTime :date="location?.createdAt" />
                </div>
              </div>
            </div>
            <div class="ml-auto mt-2 flex flex-wrap items-center justify-between gap-3">
              <LabelMaker :id="location.id" type="location" />
              <Button
                v-if="isQuickPrintAvailable"
                variant="outline"
                :disabled="isQuickPrinting"
                @click="handleQuickPrint"
              >
                <MdiLoading v-if="isQuickPrinting" class="mr-1 size-4 animate-spin" />
                <MdiPrinterPos v-else class="mr-1 size-4" />
                {{ $t("locations.quick_print") }}
              </Button>
              <Button variant="outline" @click="openPrintLabelTemplate">
                <MdiLabel class="mr-1 size-4" />
                {{ $t("locations.print_label_template") }}
              </Button>
              <Button @click="openUpdate">
                <MdiPencil name="mdi-pencil" />
                {{ $t("global.edit") }}
              </Button>
              <Button variant="destructive" @click="confirmDelete()">
                <MdiDelete name="mdi-delete" />
                {{ $t("global.delete") }}
              </Button>
            </div>
          </div>
        </header>
        <Separator v-if="location && location.description" />
        <Markdown v-if="location && location.description" class="mt-3 text-base" :source="location.description" />
      </Card>
      <section v-if="location && items">
        <ItemViewSelectable :items="items" @refresh="refreshItemList" />
      </section>

      <section v-if="location && location.children.length > 0" class="mt-6">
        <BaseSectionHeader class="mb-5"> {{ $t("locations.child_locations") }} </BaseSectionHeader>
        <div class="grid grid-cols-1 gap-2 sm:grid-cols-3">
          <LocationCard v-for="item in location.children" :key="item.id" :location="item" />
        </div>
      </section>
    </BaseContainer>
  </div>
</template>
