<template>
  <BaseModal :dialog-id="DialogID.CreateTemplate" title="Create Template">
    <form class="flex flex-col gap-2" @submit.prevent="create()">
      <FormTextField v-model="form.name" :autofocus="true" label="Name" :max-length="255" :min-length="1" />
      <FormTextArea v-model="form.description" label="Description" :max-length="1000" />

      <Separator class="my-2" />
      <h3 class="text-sm font-medium">Default Item Values</h3>
      <div class="grid gap-2">
        <FormTextField v-model.number="form.defaultQuantity" label="Quantity" type="number" :min="1" />
        <FormTextField v-model="form.defaultManufacturer" label="Manufacturer" :max-length="255" />
        <div class="flex items-center gap-2">
          <Switch id="defaultInsured" v-model:checked="form.defaultInsured" />
          <Label for="defaultInsured" class="text-sm">Insured</Label>
        </div>
        <div class="flex items-center gap-2">
          <Switch id="defaultLifetimeWarranty" v-model:checked="form.defaultLifetimeWarranty" />
          <Label for="defaultLifetimeWarranty" class="text-sm">Lifetime Warranty</Label>
        </div>
      </div>

      <Separator class="my-2" />
      <div class="flex items-center justify-between">
        <h3 class="text-sm font-medium">Custom Fields</h3>
        <Button type="button" size="sm" variant="outline" @click="addField">
          <MdiPlus class="mr-1 size-4" />
          Add
        </Button>
      </div>
      <div v-if="form.fields.length > 0" class="space-y-2">
        <div v-for="(field, idx) in form.fields" :key="idx" class="flex items-end gap-2">
          <FormTextField v-model="field.name" label="Field Name" :max-length="255" class="flex-1" />
          <FormTextField v-model="field.textValue" label="Default Value" class="flex-1" />
          <Button type="button" size="icon" variant="ghost" @click="form.fields.splice(idx, 1)">
            <MdiDelete class="size-4" />
          </Button>
        </div>
      </div>
      <p v-else class="text-sm text-muted-foreground">No custom fields.</p>

      <div class="mt-4 flex justify-end">
        <Button type="submit" :loading="loading">{{ $t("global.create") }}</Button>
      </div>
    </form>
  </BaseModal>
</template>

<script setup lang="ts">
  import { toast } from "@/components/ui/sonner";
  import MdiPlus from "~icons/mdi/plus";
  import MdiDelete from "~icons/mdi/delete";
  import { DialogID } from "@/components/ui/dialog-provider/utils";
  import BaseModal from "@/components/App/CreateModal.vue";
  import { useDialog } from "~/components/ui/dialog-provider";
  import FormTextField from "~/components/Form/TextField.vue";
  import FormTextArea from "~/components/Form/TextArea.vue";
  import { Button } from "~/components/ui/button";
  import { Separator } from "@/components/ui/separator";
  import { Switch } from "@/components/ui/switch";
  import { Label } from "@/components/ui/label";

  const emit = defineEmits<{ created: [] }>();
  const { closeDialog } = useDialog();

  const loading = ref(false);
  const form = reactive({
    name: "",
    description: "",
    notes: "",
    defaultQuantity: 1,
    defaultInsured: false,
    defaultManufacturer: "",
    defaultLifetimeWarranty: false,
    defaultWarrantyDetails: "",
    includeWarrantyFields: false,
    includePurchaseFields: false,
    includeSoldFields: false,
    fields: [] as Array<{ id: string; name: string; type: "text"; textValue: string }>,
  });

  function addField() {
    form.fields.push({ id: "", name: "", type: "text", textValue: "" });
  }

  function reset() {
    Object.assign(form, {
      name: "",
      description: "",
      notes: "",
      defaultQuantity: 1,
      defaultInsured: false,
      defaultManufacturer: "",
      defaultLifetimeWarranty: false,
      defaultWarrantyDetails: "",
      includeWarrantyFields: false,
      includePurchaseFields: false,
      includeSoldFields: false,
      fields: [],
    });
    loading.value = false;
  }

  const api = useUserApi();

  async function create() {
    if (loading.value) return;
    loading.value = true;

    const { error } = await api.templates.create(form);
    if (error) {
      toast.error("Failed to create template");
      loading.value = false;
      return;
    }

    toast.success("Template created");
    reset();
    closeDialog(DialogID.CreateTemplate);
    emit("created");
  }
</script>
