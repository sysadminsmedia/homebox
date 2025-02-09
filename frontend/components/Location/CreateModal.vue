<template>
  <Dialog dialog-id="create-location">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>{{ $t("components.location.create_modal.title") }}</DialogTitle>
      </DialogHeader>
      <form @submit.prevent="create()">
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
            <Button  :disabled="loading" type="submit">{{ $t("global.create") }}</Button>
            <Button variant="outline" :disabled="loading" type="button" @click="create(false)">{{ $t("global.create_and_add") }}</Button>
          </ButtonGroup>
        </div>
      </form>

      <DialogFooter>
        <span class="flex items-center gap-1 text-sm">
          Use <Shortcut size="sm" :keys="['Shift']" /> + <Shortcut size="sm" :keys="['Enter']" /> to create and add
          another.
        </span>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
  import { toast } from "vue-sonner";
  import { Button, ButtonGroup } from "~/components/ui/button";
  import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "~/components/ui/dialog";
  import type { LocationSummary } from "~~/lib/api/types/data-contracts";
  import MdiChevronDown from "~icons/mdi/chevron-down";
  import { useDialog, useDialogHotkey } from "~/components/ui/dialog-provider";

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
        // useTimeoutFn(() => {
        //   focused.value = true;
        // }, 50);

        if (locationId.value) {
          const found = locations.value.find(l => l.id === locationId.value);
          if (found) {
            form.parent = found;
          }
        }
      } else {
        // focused.value = false;
      }
    }
  );

  function reset() {
    form.name = "";
    form.description = "";
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
      toast.error("Already creating a location");
      return;
    }
    loading.value = true;

    if (shift.value) {
      close = false;
    }

    const { data, error } = await api.locations.create({
      name: form.name,
      description: form.description,
      parentId: form.parent ? form.parent.id : null,
    });

    if (error) {
      loading.value = false;
      toast.error("Couldn't create location");
    }

    if (data) {
      toast.success("Location created");
    }
    reset();

    if (close) {
      closeDialog("create-location");
      navigateTo(`/location/${data.id}`);
    }
  }
</script>
