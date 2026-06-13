<script setup lang="ts">
  import { computed, onMounted, reactive, ref } from "vue";
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
  import { Dialog, DialogScrollContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { DialogID } from "~/components/ui/dialog-provider/utils";

  import { Button } from "@/components/ui/button";
  import { Badge } from "@/components/ui/badge";
  import { Card } from "@/components/ui/card";
  import { Input } from "@/components/ui/input";
  import { Checkbox } from "@/components/ui/checkbox";
  import { Separator } from "@/components/ui/separator";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import { Label } from "@/components/ui/label";

  import MdiPlus from "~icons/mdi/plus";
  import MdiPencil from "~icons/mdi/pencil";
  import MdiDelete from "~icons/mdi/delete";
  import MdiMapMarkerOutline from "~icons/mdi/map-marker-outline";
  import MdiPackageVariantClosed from "~icons/mdi/package-variant-closed";

  import MdiArrowUp from "~icons/mdi/arrow-up";
  import MdiArrowDown from "~icons/mdi/arrow-down";
  import MdiDragVertical from "~icons/mdi/drag-vertical";
  import MdiClose from "~icons/mdi/close";

  import TemplateSelector from "~/components/Template/Selector.vue";
  import FormTextField from "~/components/Form/TextField.vue";
  import { useEntityTypeStore } from "~/stores/entityTypes";
  import { useUserApi } from "~/composables/use-api";

  import type {
    EntityTemplateSummary,
    EntityTypeCreate,
    EntityTypeSummary,
    EntityTypeUpdate,
  } from "~~/lib/api/types/data-contracts";

  const { t } = useI18n();

  useHead({ title: `HomeBox | ${t("collection.tabs.entity_types")}` });

  const api = useUserApi();
  const entityTypeStore = useEntityTypeStore();
  const { openDialog, closeDialog } = useDialog();
  const confirm = useConfirm();

  const entityTypes = computed(() => entityTypeStore.allTypes);

  // -----------------------
  // Create form
  // -----------------------
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
      toast.error(t("components.entityTypes.toasts.name_required"));
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
      toast.error(t("components.entityTypes.toasts.create_failed"));
      return;
    }

    toast.success(t("components.entityTypes.toasts.create_success"));
    resetCreateForm();
    closeDialog(DialogID.CreateEntityType);
    entityTypeStore.refresh();
  }

  // -----------------------
  // Edit POC model
  // -----------------------
  type FieldType = "text" | "number" | "date" | "boolean";

  type EntityTypeField = {
    id?: string;
    name: string;
    fieldType: FieldType;
    order: number;
  };

  type EntityTypeFieldGroup = {
    id?: string;
    name: string;
    order: number;
    isRequiredGroup: boolean;
    fields: EntityTypeField[];
  };

  const REQUIRED_GROUP_ORDER = 0;

  function makeRequiredGroup(): EntityTypeFieldGroup {
    return {
      id: "required",
      name: "Required fields",
      order: REQUIRED_GROUP_ORDER,
      isRequiredGroup: true,
      fields: [],
    };
  }

  function normaliseGroupOrders(groups: EntityTypeFieldGroup[]) {
    const sorted = [...groups].sort((a, b) => a.order - b.order);
    sorted.forEach((g, idx) => {
      g.order = idx;
    });
    return sorted;
  }

  function normaliseFieldOrders(fields: EntityTypeField[]) {
    const sorted = [...fields].sort((a, b) => a.order - b.order);
    sorted.forEach((f, idx) => {
      f.order = idx;
    });
    return sorted;
  }

  const updateForm = reactive({
    id: "",
    name: "",
    icon: "",
    isLocation: false,
    originalIsItem: false,
    // New: optional allow-list
    allowedParentTypeIds: [] as string[],
  });

  const updateTemplate = ref<EntityTemplateSummary | null>(null);

  // Field groups state for POC
  const fieldGroups = ref<EntityTypeFieldGroup[]>([makeRequiredGroup()]);

  function resetFieldGroupsForEdit() {
    fieldGroups.value = [makeRequiredGroup()];
  }

  function openEdit(et: EntityTypeSummary) {
    updateForm.id = et.id;
    updateForm.name = et.name;
    updateForm.icon = et.icon;
    updateForm.isLocation = et.isLocation;
    updateForm.originalIsItem = !et.isLocation;
    updateForm.allowedParentTypeIds = []; // optional allow-list starts empty for POC

    updateTemplate.value = et.defaultTemplate
      ? ({
          id: et.defaultTemplate.id,
          name: et.defaultTemplate.name,
          description: et.defaultTemplate.description,
        } as EntityTemplateSummary)
      : null;

    // POC initial field config. Replace with real data hydration later.
    resetFieldGroupsForEdit();

    // Example: add 2 required fields by default for demonstration
    const required = fieldGroups.value.find(g => g.isRequiredGroup)!;
    required.fields.push(
      { id: "rf1", name: "Name", fieldType: "text", order: 0 },
      { id: "rf2", name: "Quantity", fieldType: "number", order: 1 }
    );
    required.fields = normaliseFieldOrders(required.fields);

    // Example optional group (non-required)
    fieldGroups.value.push({
      id: "g1",
      name: "Extra details",
      order: 1,
      isRequiredGroup: false,
      fields: [
        { id: "f1", name: "Expiry date", fieldType: "date", order: 0 },
        { id: "f2", name: "Is fragile", fieldType: "boolean", order: 1 },
      ],
    });

    fieldGroups.value = normaliseGroupOrders(fieldGroups.value);

    openDialog(DialogID.UpdateEntityType);
  }

  function reorderGroupsAfter(index: number, delta: number) {
    // Only allow moving groups after required group (required group is order 0 / index 0 after sort)
    const groups = [...fieldGroups.value].sort((a, b) => a.order - b.order);
    if (index <= 0) return; // required group index 0
    const newIndex = index + delta;
    if (newIndex <= 0 || newIndex >= groups.length) return;
    const tmp = groups[index];
    groups[index] = groups[newIndex];
    groups[newIndex] = tmp;
    fieldGroups.value = normaliseGroupOrders(groups);
  }

  function reorderFieldsInGroup(groupIndex: number, fieldIndex: number, delta: number) {
    const groups = [...fieldGroups.value].sort((a, b) => a.order - b.order);
    const g = groups[groupIndex];
    if (!g) return;

    const fields = normaliseFieldOrders([...g.fields]);
    if (fieldIndex < 0 || fieldIndex >= fields.length) return;

    const newIndex = fieldIndex + delta;
    if (newIndex < 0 || newIndex >= fields.length) return;

    const tmp = fields[fieldIndex];
    fields[fieldIndex] = fields[newIndex];
    fields[newIndex] = tmp;

    g.fields = normaliseFieldOrders(fields);
    fieldGroups.value = normaliseGroupOrders(groups);
  }

  function addGroup() {
    // Add after last group
    const groups = [...fieldGroups.value].sort((a, b) => a.order - b.order);
    const lastOrder = groups.length ? Math.max(...groups.map(g => g.order)) : 0;

    groups.push({
      id: crypto.randomUUID?.() ?? `g-${Date.now()}`,
      name: "New group",
      order: lastOrder + 1,
      isRequiredGroup: false,
      fields: [],
    });

    fieldGroups.value = normaliseGroupOrders(groups);
  }

  function deleteGroup(groupIndex: number) {
    const groups = [...fieldGroups.value].sort((a, b) => a.order - b.order);
    const g = groups[groupIndex];
    if (!g) return;
    if (g.isRequiredGroup) return; // immutable

    groups.splice(groupIndex, 1);
    fieldGroups.value = normaliseGroupOrders(groups);
  }

  function addField(groupIndex: number) {
    const groups = [...fieldGroups.value].sort((a, b) => a.order - b.order);
    const g = groups[groupIndex];
    if (!g) return;

    const nextOrder = g.fields.length ? Math.max(...g.fields.map(f => f.order)) + 1 : 0;
    g.fields.push({
      id: crypto.randomUUID?.() ?? `f-${Date.now()}`,
      name: "",
      fieldType: "text",
      order: nextOrder,
    });

    g.fields = normaliseFieldOrders(g.fields);
    fieldGroups.value = normaliseGroupOrders(groups);
  }

  function deleteField(groupIndex: number, fieldIndex: number) {
    const groups = [...fieldGroups.value].sort((a, b) => a.order - b.order);
    const g = groups[groupIndex];
    if (!g) return;
    const fields = normaliseFieldOrders([...g.fields]);
    if (fieldIndex < 0 || fieldIndex >= fields.length) return;

    fields.splice(fieldIndex, 1);
    g.fields = normaliseFieldOrders(fields);

    fieldGroups.value = normaliseGroupOrders(groups);
  }

  function validateBeforeSave(): string | null {
    const groups = [...fieldGroups.value].sort((a, b) => a.order - b.order);
    if (groups.length === 0) return "Missing field groups";

    const required = groups.find(g => g.isRequiredGroup);
    if (!required) return "Required group must exist";
    if (required.order !== REQUIRED_GROUP_ORDER) return "Required group must be first";

    for (const g of groups) {
      if (!g.isRequiredGroup && !g.name.trim()) return "Every group must have a name";
      for (const f of g.fields) {
        if (!f.name.trim()) return "Every field must have a name";
      }
    }
    return null;
  }

  async function update() {
    if (!updateForm.name.trim()) {
      toast.error(t("components.entityTypes.toasts.name_required"));
      return;
    }

    const err = validateBeforeSave();
    if (err) {
      toast.error(err);
      return;
    }

    const payload = {
      id: updateForm.id,
      name: updateForm.name,
      icon: updateForm.icon,
      isLocation: updateForm.isLocation,
      ...(updateTemplate.value?.id ? { defaultTemplateId: updateTemplate.value.id } : {}),

      // New POC fields
      allowedParentTypeIds: updateForm.allowedParentTypeIds.length > 0 ? updateForm.allowedParentTypeIds : null,

      // Field groups payload
      fieldGroups: [...fieldGroups.value]
        .sort((a, b) => a.order - b.order)
        .map(g => ({
          id: g.id,
          name: g.isRequiredGroup ? null : g.name,
          order: g.order,
          isRequiredGroup: g.isRequiredGroup,
          fields: normaliseFieldOrders([...g.fields]).map(f => ({
            id: f.id,
            name: f.name,
            fieldType: f.fieldType,
            order: f.order,
          })),
        })),
    } as EntityTypeUpdate & {
      allowedParentTypeIds: string[] | null;
      fieldGroups: any[];
    };

    // Convert this into your real API call later.
    // For now, show payload validity only.
    console.debug("EntityType update POC payload:", payload);

    const proceedIfConvertNeeded = async () => {
      if (updateForm.originalIsItem && updateForm.isLocation) {
        const { isCanceled } = await confirm.open(t("components.entityTypes.confirm.convert_item_to_location"));
        if (isCanceled) return false;
      }
      return true;
    };

    const canProceed = await proceedIfConvertNeeded();
    if (!canProceed) return;

    // Keep existing behaviour (no backend changes in POC)
    const { error } = await api.entityTypes.update(updateForm.id, {
      id: updateForm.id,
      name: updateForm.name,
      icon: updateForm.icon,
      isLocation: updateForm.isLocation,
      ...(updateTemplate.value?.id ? { defaultTemplateId: updateTemplate.value.id } : {}),
    } as EntityTypeUpdate);

    if (error) {
      toast.error(t("components.entityTypes.toasts.update_failed"));
      return;
    }

    toast.success(t("components.entityTypes.toasts.update_success"));
    closeDialog(DialogID.UpdateEntityType);
    entityTypeStore.refresh();
  }

  async function deleteEntityType(et: EntityTypeSummary) {
    const { isCanceled } = await confirm.open(
      t("components.entityTypes.confirm.delete_entity_type", { name: et.name })
    );
    if (isCanceled) return;

    const { error } = await api.entityTypes.delete(et.id);
    if (error) {
      toast.error(t("components.entityTypes.toasts.delete_confirm_failed"));
      return;
    }

    toast.success(t("components.entityTypes.toasts.delete_success"));
    entityTypeStore.refresh();
  }

  function toggleAllowedParentType(typeId: string) {
    const idx = updateForm.allowedParentTypeIds.indexOf(typeId);
    if (idx >= 0) {
      updateForm.allowedParentTypeIds.splice(idx, 1);
    } else {
      updateForm.allowedParentTypeIds.push(typeId);
    }
  }

  function moveUp(i: number, maxIndexExclusive: number) {
    return i > 0 ? i - 1 : i;
  }

  function moveDown(i: number, maxIndexExclusive: number) {
    return i < maxIndexExclusive - 1 ? i + 1 : i;
  }

  // -----------------------
  // Hot fix for crypto in older browsers
  // -----------------------
  onMounted(() => {
    if (typeof crypto === "undefined") {
      // nothing - POC ids will still work with Date.now fallback
    }
  });

  // -----------------------
  // For parity with old file
  // -----------------------
  onMounted(() => {
    // no-op
  });

  const allowedTypeOptions = computed(() => entityTypes.value);
</script>

<template>
  <div>
    <!-- Create Dialog -->
    <Dialog :dialog-id="DialogID.CreateEntityType">
      <DialogScrollContent>
        <DialogHeader>
          <DialogTitle>{{ t("components.entityTypes.create_dialog.title") }}</DialogTitle>
        </DialogHeader>

        <form class="flex flex-col gap-3" @submit.prevent="create">
          <FormTextField
            v-model="createForm.name"
            :autofocus="true"
            :label="t('components.entityTypes.create_dialog.name_label')"
            :max-length="255"
            :min-length="1"
          />

          <div class="flex items-center gap-3">
            <Checkbox
              id="create-isLocation"
              :checked="createForm.isLocation"
              @update:checked="(v: boolean) => (createForm.isLocation = v)"
            />
            <Label for="create-isLocation">
              {{ t("components.entityTypes.create_dialog.is_container_location_type_label") }}
            </Label>
          </div>

          <TemplateSelector v-if="!createForm.isLocation" v-model="createTemplate" />

          <DialogFooter>
            <Button type="submit">{{ t("components.entityTypes.create_dialog.button") }}</Button>
          </DialogFooter>
        </form>
      </DialogScrollContent>
    </Dialog>

    <!-- Update Dialog -->
    <Dialog :dialog-id="DialogID.UpdateEntityType">
      <DialogScrollContent class="max-w-4xl">
        <DialogHeader>
          <DialogTitle>{{ t("components.entityTypes.update_dialog.title") }}</DialogTitle>
        </DialogHeader>

        <form class="flex flex-col gap-4" @submit.prevent="update">
          <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
            <FormTextField
              v-model="updateForm.name"
              :autofocus="true"
              :label="t('components.entityTypes.update_dialog.name_label')"
              :max-length="255"
              :min-length="1"
            />

            <div class="flex items-center gap-3">
              <Checkbox
                id="update-isLocation"
                :checked="updateForm.isLocation"
                @update:checked="(v: boolean) => (updateForm.isLocation = v)"
              />
              <Label for="update-isLocation">
                {{ t("components.entityTypes.update_dialog.is_container_location_type_label") }}
              </Label>
            </div>
          </div>

          <div class="flex flex-col gap-2">
            <TemplateSelector v-if="!updateForm.isLocation" v-model="updateTemplate" />
          </div>

          <Separator />

          <!-- Parent allow-list -->
          <div class="flex flex-col gap-2">
            <div class="flex items-center justify-between">
              <h3 class="text-sm font-medium">Parent entity type allow-list (optional)</h3>
              <Badge v-if="updateForm.allowedParentTypeIds.length === 0" variant="secondary"> Not restricted </Badge>
            </div>

            <p class="text-xs text-muted-foreground">
              If left empty, any entity type can be a parent of this entity type.
            </p>

            <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
              <div v-for="et in allowedTypeOptions" :key="et.id" class="flex items-center gap-3 rounded-md border p-2">
                <Checkbox
                  :id="`parent-${et.id}`"
                  :checked="updateForm.allowedParentTypeIds.includes(et.id)"
                  @update:checked="() => toggleAllowedParentType(et.id)"
                />
                <Label :for="`parent-${et.id}`" class="min-w-0">
                  <div class="flex items-center gap-2">
                    <span class="truncate font-medium">{{ et.name }}</span>
                    <span v-if="et.isLocation" class="text-xs text-muted-foreground"> (location) </span>
                  </div>
                </Label>
              </div>
            </div>
          </div>

          <Separator />

          <!-- Field groups builder -->
          <div class="flex flex-col gap-3">
            <div class="flex items-center justify-between">
              <h3 class="text-sm font-medium">Entity type fields</h3>
              <Button type="button" variant="outline" @click="addGroup">
                <span class="mr-2 inline-flex items-center justify-center">
                  <MdiPlus class="size-4" />
                </span>
                Add group
              </Button>
            </div>

            <p class="text-xs text-muted-foreground">
              Group 1 is the required group. It is pre-defined and cannot be renamed or reordered.
            </p>

            <div class="flex flex-col gap-3">
              <Card
                v-for="(group, groupIndex) in [...fieldGroups].sort((a, b) => a.order - b.order)"
                :key="group.id ?? group.order"
                class="p-3"
              >
                <!-- Group header -->
                <div class="flex items-center gap-3">
                  <div class="flex flex-1 items-center gap-3">
                    <div
                      class="flex size-9 items-center justify-center rounded-md bg-secondary text-secondary-foreground"
                    >
                      <MdiDragVertical class="size-4" />
                    </div>

                    <div class="min-w-0 flex-1">
                      <div class="flex items-center gap-2">
                        <Input
                          v-if="!group.isRequiredGroup"
                          v-model="group.name"
                          :placeholder="`Group name`"
                          class="h-8"
                        />
                        <div v-else class="text-sm font-medium">Required fields</div>
                      </div>
                      <div class="mt-0.5 text-xs text-muted-foreground">
                        Group {{ group.order + 1 }}
                        <span v-if="group.isRequiredGroup">- required</span>
                      </div>
                    </div>
                  </div>

                  <!-- Group reorder buttons (only for non-required groups) -->
                  <div v-if="!group.isRequiredGroup" class="flex items-center gap-2">
                    <Button
                      type="button"
                      variant="outline"
                      size="icon"
                      class="size-8"
                      :disabled="groupIndex === 0"
                      @click="reorderGroupsAfter(groupIndex, -1)"
                    >
                      <MdiArrowUp class="size-4" />
                    </Button>
                    <Button
                      type="button"
                      variant="outline"
                      size="icon"
                      class="size-8"
                      :disabled="groupIndex === fieldGroups.length - 1"
                      @click="reorderGroupsAfter(groupIndex, 1)"
                    >
                      <MdiArrowDown class="size-4" />
                    </Button>

                    <TooltipProvider :delay-duration="0">
                      <Tooltip>
                        <TooltipTrigger as-child>
                          <Button
                            type="button"
                            variant="destructive"
                            size="icon"
                            class="size-8"
                            @click="deleteGroup(groupIndex)"
                          >
                            <MdiDelete class="size-4" />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>Delete group</TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </div>
                </div>

                <Separator class="my-3" />

                <!-- Fields -->
                <div class="flex items-center justify-between">
                  <h4 class="text-sm font-medium">Fields ({{ group.fields.length }})</h4>

                  <Button type="button" variant="outline" size="sm" @click="addField(groupIndex)">
                    <MdiPlus class="mr-2 size-4" />
                    Add field
                  </Button>
                </div>

                <div class="mt-3 flex flex-col gap-2">
                  <div
                    v-for="(field, fieldIndex) in normaliseFieldOrders(group.fields)"
                    :key="field.id ?? `${groupIndex}-${fieldIndex}`"
                    class="rounded-md border p-2"
                  >
                    <div class="flex items-start gap-3">
                      <div class="mt-1 flex flex-col gap-2">
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          class="size-7"
                          :disabled="fieldIndex === 0"
                          @click="reorderFieldsInGroup(groupIndex, fieldIndex, -1)"
                        >
                          <MdiArrowUp class="size-4" />
                        </Button>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          class="size-7"
                          :disabled="fieldIndex === group.fields.length - 1"
                          @click="reorderFieldsInGroup(groupIndex, fieldIndex, 1)"
                        >
                          <MdiArrowDown class="size-4" />
                        </Button>
                      </div>

                      <div class="grid flex-1 grid-cols-1 gap-2 sm:grid-cols-[1fr_180px]">
                        <div class="flex flex-col gap-1">
                          <Label class="text-xs">Field name</Label>
                          <Input v-model="group.fields[fieldIndex].name" placeholder="e.g. Serial number" />
                        </div>

                        <div class="flex flex-col gap-1">
                          <Label class="text-xs">Field type</Label>
                          <Select
                            :model-value="group.fields[fieldIndex].fieldType"
                            @update:model-value="(v: FieldType) => (group.fields[fieldIndex].fieldType = v)"
                          >
                            <SelectTrigger>
                              <SelectValue placeholder="Select type" />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="text">Text</SelectItem>
                              <SelectItem value="number">Number</SelectItem>
                              <SelectItem value="date">Date</SelectItem>
                              <SelectItem value="boolean">Boolean</SelectItem>
                            </SelectContent>
                          </Select>
                        </div>
                      </div>

                      <div class="flex items-center">
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          class="size-8 text-destructive"
                          :disabled="group.isRequiredGroup && group.fields.length === 0"
                          @click="deleteField(groupIndex, fieldIndex)"
                        >
                          <MdiClose class="size-4" />
                        </Button>
                      </div>
                    </div>
                  </div>

                  <div v-if="group.fields.length === 0" class="text-xs text-muted-foreground">
                    No fields in this group yet.
                  </div>
                </div>
              </Card>
            </div>
          </div>

          <DialogFooter class="mt-2">
            <Button type="submit">
              {{ t("components.entityTypes.update_dialog.button") }}
            </Button>
          </DialogFooter>
        </form>
      </DialogScrollContent>
    </Dialog>

    <!-- Page Content -->
    <div class="mb-4 flex items-center justify-between">
      <h3 class="text-lg font-medium">{{ t("components.entityTypes.page.title") }}</h3>
      <Button size="sm" @click="openDialog(DialogID.CreateEntityType)">
        <MdiPlus class="mr-1 size-4" />
        {{ t("components.entityTypes.page.create") }}
      </Button>
    </div>

    <div v-if="entityTypes && entityTypes.length > 0" class="flex flex-col gap-2">
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
              <span class="font-medium">{{ t(et.name) }}</span>
              <Badge v-if="et.isLocation" variant="secondary" class="text-xs">
                {{ t("components.entityTypes.card.badge_container") }}
              </Badge>
            </div>
            <p v-if="et.defaultTemplate && !et.isLocation" class="text-xs text-muted-foreground">
              {{ t("components.entityTypes.card.default_template", { name: et.defaultTemplate.name }) }}
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
                <TooltipContent>{{ t("components.entityTypes.card.tooltip.edit") }}</TooltipContent>
              </Tooltip>
              <Tooltip>
                <TooltipTrigger as-child>
                  <Button variant="ghost" size="icon" class="size-8 text-destructive" @click="deleteEntityType(et)">
                    <MdiDelete class="size-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>{{ t("components.entityTypes.card.tooltip.delete") }}</TooltipContent>
              </Tooltip>
            </div>
          </TooltipProvider>
        </div>
      </Card>
    </div>

    <div v-else class="flex flex-col items-center justify-center py-12 text-center">
      <p class="mb-4 text-muted-foreground">{{ t("components.entityTypes.page.empty_title") }}</p>
      <Button @click="openDialog(DialogID.CreateEntityType)">
        <MdiPlus class="mr-2" />
        {{ t("components.entityTypes.page.empty_button") }}
      </Button>
    </div>
  </div>
</template>
