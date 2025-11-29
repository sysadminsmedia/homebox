<script setup lang="ts">
  import { toast } from "@/components/ui/sonner";
  import MdiPencil from "~icons/mdi/pencil";
  import MdiDelete from "~icons/mdi/delete";
  import MdiPlus from "~icons/mdi/plus";
  import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { Card } from "@/components/ui/card";
  import { Button } from "@/components/ui/button";
  import { Separator } from "@/components/ui/separator";
  import { Switch } from "@/components/ui/switch";
  import { Label } from "@/components/ui/label";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import FormTextField from "~/components/Form/TextField.vue";
  import FormTextArea from "~/components/Form/TextArea.vue";
  import BaseContainer from "@/components/Base/Container.vue";
  import DateTime from "~/components/global/DateTime.vue";
  import Markdown from "~/components/global/Markdown.vue";

  definePageMeta({
    middleware: ["auth"],
  });

  const { openDialog, closeDialog } = useDialog();
  const route = useRoute();
  const api = useUserApi();
  const confirm = useConfirm();

  const templateId = computed<string>(() => route.params.id as string);

  const { data: template, refresh } = useAsyncData(templateId.value, async () => {
    const { data, error } = await api.templates.get(templateId.value);
    if (error) {
      toast.error("Failed to load template");
      navigateTo("/templates");
      return;
    }
    return data;
  });

  async function confirmDelete() {
    const { isCanceled } = await confirm.open("Delete this template?");
    if (isCanceled) return;

    const { error } = await api.templates.delete(templateId.value);
    if (error) {
      toast.error("Failed to delete template");
      return;
    }
    toast.success("Template deleted");
    navigateTo("/templates");
  }

  const updating = ref(false);
  const updateData = reactive({
    id: "",
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

  function openUpdate() {
    if (!template.value) return;
    Object.assign(updateData, {
      id: template.value.id,
      name: template.value.name,
      description: template.value.description,
      notes: template.value.notes,
      defaultQuantity: template.value.defaultQuantity,
      defaultInsured: template.value.defaultInsured,
      defaultManufacturer: template.value.defaultManufacturer,
      defaultLifetimeWarranty: template.value.defaultLifetimeWarranty,
      defaultWarrantyDetails: template.value.defaultWarrantyDetails,
      includeWarrantyFields: template.value.includeWarrantyFields,
      includePurchaseFields: template.value.includePurchaseFields,
      includeSoldFields: template.value.includeSoldFields,
      fields: template.value.fields.map(f => ({
        id: f.id,
        name: f.name,
        type: "text" as const,
        textValue: f.textValue,
      })),
    });
    openDialog(DialogID.UpdateTemplate);
  }

  async function update() {
    updating.value = true;
    const { error, data } = await api.templates.update(templateId.value, updateData);
    if (error) {
      updating.value = false;
      toast.error("Failed to update template");
      return;
    }
    toast.success("Template updated");
    template.value = data;
    closeDialog(DialogID.UpdateTemplate);
    updating.value = false;
    refresh();
  }

  const NIL_UUID = "00000000-0000-0000-0000-000000000000";
</script>

<template>
  <Dialog :dialog-id="DialogID.UpdateTemplate">
    <DialogContent class="max-h-[90vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>Edit Template</DialogTitle>
      </DialogHeader>

      <form v-if="template" class="flex flex-col gap-2" @submit.prevent="update">
        <FormTextField v-model="updateData.name" :autofocus="true" label="Name" :max-length="255" />
        <FormTextArea v-model="updateData.description" label="Description" :max-length="1000" />

        <Separator class="my-2" />
        <h3 class="text-sm font-medium">Default Item Values</h3>
        <div class="grid gap-2">
          <FormTextField v-model.number="updateData.defaultQuantity" label="Quantity" type="number" :min="1" />
          <FormTextField v-model="updateData.defaultManufacturer" label="Manufacturer" :max-length="255" />
          <div class="flex items-center gap-2">
            <Switch id="editInsured" v-model:checked="updateData.defaultInsured" />
            <Label for="editInsured" class="text-sm">Insured</Label>
          </div>
          <div class="flex items-center gap-2">
            <Switch id="editWarranty" v-model:checked="updateData.defaultLifetimeWarranty" />
            <Label for="editWarranty" class="text-sm">Lifetime Warranty</Label>
          </div>
        </div>

        <Separator class="my-2" />
        <div class="flex items-center justify-between">
          <h3 class="text-sm font-medium">Custom Fields</h3>
          <Button
            type="button"
            size="sm"
            variant="outline"
            @click="updateData.fields.push({ id: NIL_UUID, name: '', type: 'text', textValue: '' })"
          >
            <MdiPlus class="mr-1 size-4" />
            Add
          </Button>
        </div>
        <div v-if="updateData.fields.length > 0" class="space-y-2">
          <div v-for="(field, idx) in updateData.fields" :key="idx" class="flex items-end gap-2">
            <FormTextField v-model="field.name" label="Field Name" :max-length="255" class="flex-1" />
            <FormTextField v-model="field.textValue" label="Default Value" class="flex-1" />
            <Button type="button" size="icon" variant="ghost" @click="updateData.fields.splice(idx, 1)">
              <MdiDelete class="size-4" />
            </Button>
          </div>
        </div>

        <DialogFooter>
          <Button type="submit" :loading="updating">{{ $t("global.update") }}</Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>

  <BaseContainer v-if="template">
    <Title>{{ template.name }}</Title>

    <Card class="p-3">
      <header :class="{ 'mb-2': template.description }">
        <div class="flex flex-wrap items-end gap-2">
          <div>
            <h1 class="pb-1 text-2xl">{{ template.name }}</h1>
            <div class="flex flex-wrap gap-1 text-xs text-muted-foreground">
              <span>Created <DateTime :date="template.createdAt" /></span>
              <span>•</span>
              <span>Updated <DateTime :date="template.updatedAt" /></span>
            </div>
          </div>
          <div class="ml-auto flex gap-2">
            <Button @click="openUpdate">
              <MdiPencil class="mr-1" />
              {{ $t("global.edit") }}
            </Button>
            <Button variant="destructive" @click="confirmDelete">
              <MdiDelete class="mr-1" />
              {{ $t("global.delete") }}
            </Button>
          </div>
        </div>
      </header>

      <Separator v-if="template.description" class="my-3" />
      <Markdown v-if="template.description" :source="template.description" />

      <Separator class="my-3" />
      <div class="grid gap-4 text-sm md:grid-cols-2">
        <div>
          <h3 class="mb-2 font-medium">Default Values</h3>
          <dl class="space-y-1">
            <div class="flex justify-between">
              <dt class="text-muted-foreground">Quantity</dt>
              <dd>{{ template.defaultQuantity }}</dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-muted-foreground">Insured</dt>
              <dd>{{ template.defaultInsured ? "Yes" : "No" }}</dd>
            </div>
            <div v-if="template.defaultManufacturer" class="flex justify-between">
              <dt class="text-muted-foreground">Manufacturer</dt>
              <dd>{{ template.defaultManufacturer }}</dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-muted-foreground">Lifetime Warranty</dt>
              <dd>{{ template.defaultLifetimeWarranty ? "Yes" : "No" }}</dd>
            </div>
          </dl>
        </div>
        <div v-if="template.fields.length > 0">
          <h3 class="mb-2 font-medium">Custom Fields</h3>
          <dl class="space-y-1">
            <div v-for="field in template.fields" :key="field.id" class="flex justify-between">
              <dt class="text-muted-foreground">{{ field.name }}</dt>
              <dd>{{ field.textValue || "—" }}</dd>
            </div>
          </dl>
        </div>
      </div>
    </Card>
  </BaseContainer>
</template>
