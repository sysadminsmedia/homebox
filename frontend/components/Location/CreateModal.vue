<template>
  <BaseModal dialog-id="create-location" :title="$t('components.location.create_modal.title')">
    <form class="flex flex-col gap-2" @submit.prevent="create()">
      <LocationSelector v-model="form.parent" />
      <FormTextField
        ref="locationNameRef"
        v-model="form.name"
        :trigger-focus="focused"
        :autofocus="true"
        :required="true"
        :label="$t('components.location.create_modal.location_name')"
        :max-length="255"
        :min-length="1"
      />
      <FormTextArea
        v-model="form.description"
        :label="$t('components.location.create_modal.location_description')"
        :max-length="1000"
      />
      <div class="mt-4 flex flex-row-reverse">
        <ButtonGroup>
          <Button :disabled="loading" type="submit">{{ $t("global.create") }}</Button>
          <Button variant="outline" :disabled="loading" type="button" @click="create(false)">{{
            $t("global.create_and_add")
          }}</Button>
        </ButtonGroup>
      </div>
    </form>
  </BaseModal>
</template>

<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import { Button, ButtonGroup } from "~/components/ui/button";
  import BaseModal from "@/components/App/CreateModal.vue";
  import type { LocationSummary } from "~~/lib/api/types/data-contracts";
  import { useDialog, useDialogHotkey } from "~/components/ui/dialog-provider";

  const { t } = useI18n();

  const { activeDialog, closeDialog } = useDialog();

  useDialogHotkey("create-location", { code: "Digit3", shift: true });

  const loading = ref(false);
  const focused = ref(false);
  const form = reactive({
    name: "",
    description: "",
    parent: null as LocationSummary | null,
  });

  watch(
    () => activeDialog.value,
    active => {
      if (active === "create-location") {
        initializeForm();
      }
    }
  );

  function initializeForm() {
    if (locationId.value) {
      const found = locations.value.find(l => l.id === locationId.value);
      form.parent = found || null;
    } else {
      form.parent = null;
    }
  }

  function reset(preserveParent = false) {
    form.name = "";
    form.description = "";
    if (!preserveParent || !locationId.value)
      form.parent = null;
    focused.value = false;
    loading.value = false;
  }

  const api = useUserApi();

  const locationsStore = useLocationStore();
  const locations = computed(() => locationsStore.allLocations);

  const route = useRoute();

  const { shift } = useMagicKeys();

  const locationId = computed(() => {
    if (route.fullPath.includes("/location/")) {
      return route.params.id;
    }
    return null;
  });

  async function create(close = true) {
    if (loading.value) {
      toast.error(t("components.location.create_modal.toast.already_creating"));
      return;
    }
    loading.value = true;

    if (shift.value) close = false;

    const { data, error } = await api.locations.create({
      name: form.name,
      description: form.description,
      parentId: form.parent ? form.parent.id : null,
    });

    if (error) {
      loading.value = false;
      toast.error(t("components.location.create_modal.toast.create_failed"));
    }

    if (data) {
      toast.success(t("components.location.create_modal.toast.create_success"));
    }
    
    const shouldPreserveParent = !close && !!locationId.value;
    reset(shouldPreserveParent);

    if (close) {
      closeDialog("create-location");
      navigateTo(`/location/${data.id}`);
    }
  }
</script>
