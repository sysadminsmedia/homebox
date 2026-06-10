<script setup lang="ts">
  import { toast } from "@/components/ui/sonner";
  import type {
    EntityTypeCreate,
    EntityTypeSummary,
    EntityTypeUpdate,
    EntityTemplateSummary,
  } from "~~/lib/api/types/data-contracts";
  import MdiPlus from "~icons/mdi/plus";
  import MdiPencil from "~icons/mdi/pencil";
  import MdiDelete from "~icons/mdi/delete";
  import MdiMapMarkerOutline from "~icons/mdi/map-marker-outline";
  import MdiPackageVariantClosed from "~icons/mdi/package-variant-closed";
  import { Button } from "@/components/ui/button";
  import { Badge } from "@/components/ui/badge";
  import { Card } from "@/components/ui/card";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
  import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import FormTextField from "~/components/Form/TextField.vue";
  import FormCheckbox from "~/components/Form/Checkbox.vue";
  import TemplateSelector from "~/components/Template/Selector.vue";

  const api = useUserApi();
  const confirm = useConfirm();
  const { openDialog, closeDialog } = useDialog();

  const { data: entityTypes, refresh } = useAsyncData("entity-types", async () => {
    const { data, error } = await api.entityTypes.getAll();
    if (error) {
      toast.error("Failed to load entity types");
      return [];
    }
    return data;
  });

  // Create form
  const createForm = reactive({
    name: "",
    icon: "",
    isLocation: false,
  });
  const createTemplate = ref<EntityTemplateSummary | null>(null);

  function resetCreateForm() {
    createForm.name = "";
    createForm.icon = "";
    createForm.isLocation = false;
    createTemplate.value = null;
  }

  async function create() {
    if (!createForm.name.trim()) {
      toast.error("Name is required");
      return;
    }

    const payload = {
      name: createForm.name,
      icon: createForm.icon,
      isLocation: createForm.isLocation,
      ...(createTemplate.value?.id ? { defaultTemplateId: createTemplate.value.id } : {}),
    } as EntityTypeCreate;

    const { error } = await api.entityTypes.create(payload);
    if (error) {
      toast.error("Failed to create entity type");
      return;
    }

    toast.success("Entity type created");
    resetCreateForm();
    closeDialog(DialogID.CreateEntityType);
    refresh();
  }

  // Update form
  const updateForm = reactive({
    id: "",
    name: "",
    icon: "",
    isLocation: false,
  });
  const updateTemplate = ref<EntityTemplateSummary | null>(null);

  function openEdit(et: EntityTypeSummary) {
    updateForm.id = et.id;
    updateForm.name = et.name;
    updateForm.icon = et.icon;
    updateForm.isLocation = et.isLocation;
    updateTemplate.value = et.defaultTemplate
      ? ({
          id: et.defaultTemplate.id,
          name: et.defaultTemplate.name,
          description: et.defaultTemplate.description,
        } as EntityTemplateSummary)
      : null;
    openDialog(DialogID.UpdateEntityType);
  }

  async function update() {
    if (!updateForm.name.trim()) {
      toast.error("Name is required");
      return;
    }

    const payload = {
      id: updateForm.id,
      name: updateForm.name,
      icon: updateForm.icon,
      isLocation: updateForm.isLocation,
      ...(updateTemplate.value?.id ? { defaultTemplateId: updateTemplate.value.id } : {}),
    } as EntityTypeUpdate;

    const { error } = await api.entityTypes.update(updateForm.id, payload);
    if (error) {
      toast.error("Failed to update entity type");
      return;
    }

    toast.success("Entity type updated");
    closeDialog(DialogID.UpdateEntityType);
    refresh();
  }

  async function deleteEntityType(et: EntityTypeSummary) {
    const { isCanceled } = await confirm.open(
      `Are you sure you want to delete "${et.name}"? Entities using this type will need to be reassigned.`
    );
    if (isCanceled) return;

    const { error } = await api.entityTypes.delete(et.id);
    if (error) {
      toast.error("Failed to delete entity type. Make sure no entities are using it.");
      return;
    }

    toast.success("Entity type deleted");
    refresh();
  }
</script>

<template>
  <div>
    <!-- Create Dialog -->
    <Dialog :dialog-id="DialogID.CreateEntityType">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Entity Type</DialogTitle>
        </DialogHeader>
        <form class="flex flex-col gap-3" @submit.prevent="create">
          <FormTextField v-model="createForm.name" :autofocus="true" label="Name" :max-length="255" :min-length="1" />
          <FormCheckbox v-model="createForm.isLocation" label="Is a container / location type" />
          <TemplateSelector v-model="createTemplate" />

          <DialogFooter>
            <Button type="submit">Create</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <!-- Update Dialog -->
    <Dialog :dialog-id="DialogID.UpdateEntityType">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Update Entity Type</DialogTitle>
        </DialogHeader>
        <form class="flex flex-col gap-3" @submit.prevent="update">
          <FormTextField v-model="updateForm.name" :autofocus="true" label="Name" :max-length="255" :min-length="1" />
          <FormCheckbox v-model="updateForm.isLocation" label="Is a container / location type" />
          <TemplateSelector v-model="updateTemplate" />

          <DialogFooter>
            <Button type="submit">Update</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <!-- Page Content -->
    <div class="mb-4 flex items-center justify-between">
      <h3 class="text-lg font-medium">Entity Types</h3>
      <Button size="sm" @click="openDialog(DialogID.CreateEntityType)">
        <MdiPlus class="mr-1 size-4" />
        Create
      </Button>
    </div>

    <div v-if="entityTypes && entityTypes.length > 0" class="space-y-2">
      <Card v-for="et in entityTypes" :key="et.id" class="p-4">
        <div class="flex items-center gap-3">
          <div
            class="flex size-10 shrink-0 items-center justify-center rounded-full bg-secondary text-secondary-foreground"
          >
            <MdiMapMarkerOutline v-if="et.isLocation" class="size-5" />
            <MdiPackageVariantClosed v-else class="size-5" />
          </div>

          <div class="mr-auto min-w-0">
            <div class="flex items-center gap-2">
              <span class="font-medium">{{ et.name }}</span>
              <Badge v-if="et.isLocation" variant="secondary" class="text-xs">Container</Badge>
            </div>
            <p v-if="et.defaultTemplate" class="text-xs text-muted-foreground">
              Default template: {{ et.defaultTemplate.name }}
            </p>
          </div>

          <TooltipProvider :delay-duration="0">
            <div class="flex gap-1">
              <Tooltip>
                <TooltipTrigger as-child>
                  <Button variant="ghost" size="icon" class="size-8" @click="openEdit(et)">
                    <MdiPencil class="size-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Edit</TooltipContent>
              </Tooltip>
              <Tooltip>
                <TooltipTrigger as-child>
                  <Button variant="ghost" size="icon" class="size-8 text-destructive" @click="deleteEntityType(et)">
                    <MdiDelete class="size-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Delete</TooltipContent>
              </Tooltip>
            </div>
          </TooltipProvider>
        </div>
      </Card>
    </div>

    <div v-else class="flex flex-col items-center justify-center py-12 text-center">
      <p class="mb-4 text-muted-foreground">No entity types defined yet.</p>
      <Button @click="openDialog(DialogID.CreateEntityType)">
        <MdiPlus class="mr-2" />
        Create Entity Type
      </Button>
    </div>
  </div>
</template>
