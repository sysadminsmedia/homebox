<template>
  <Dialog :dialog-id="DialogID.EditMaintenance">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>
          {{ entry.id ? $t("maintenance.modal.edit_title") : $t("maintenance.modal.new_title") }}
        </DialogTitle>
      </DialogHeader>

      <form class="flex flex-col gap-2" @submit.prevent="dispatchFormSubmit">
        <FormTextField v-model="entry.name" autofocus :label="$t('maintenance.modal.entry_name')" />
        <DatePicker v-model="entry.completedDate" date-only :label="$t('maintenance.modal.completed_date')" />
        <DatePicker v-model="entry.scheduledDate" date-only :label="$t('maintenance.modal.scheduled_date')" />
        <label class="flex items-center gap-2 text-sm">
          <input v-model="entry.isRecurring" type="checkbox" :true-value="true" :false-value="false" />
          {{ $t("maintenance.modal.recurring") }}
        </label>
        <div v-if="entry.isRecurring" class="grid grid-cols-2 gap-2">
          <FormTextField v-model="entry.intervalValue" :label="$t('maintenance.modal.interval_value')" />
          <select v-model="entry.intervalUnit" class="rounded-md border p-2 text-sm">
            <option :value="MaintenancePlanIntervalUnit.Hour">{{ $t("maintenance.modal.interval_hour") }}</option>
            <option :value="MaintenancePlanIntervalUnit.Day">{{ $t("maintenance.modal.interval_day") }}</option>
            <option :value="MaintenancePlanIntervalUnit.Week">{{ $t("maintenance.modal.interval_week") }}</option>
            <option :value="MaintenancePlanIntervalUnit.Month">{{ $t("maintenance.modal.interval_month") }}</option>
            <option :value="MaintenancePlanIntervalUnit.Year">{{ $t("maintenance.modal.interval_year") }}</option>
          </select>
        </div>
        <div v-if="entry.isRecurring && nextDuePreview.length > 0" class="rounded-md border p-3 text-sm">
          <p class="font-medium">{{ $t("maintenance.modal.next_scheduled_dates") }}</p>
          <ul class="mt-1 space-y-1">
            <li v-for="(dueDate, index) in nextDuePreview" :key="index">
              {{ dueDate.toLocaleDateString() }}
            </li>
          </ul>
        </div>
        <FormTextArea v-model="entry.description" :label="$t('maintenance.modal.notes')" />
        <FormTextField v-model="entry.cost" autofocus :label="$t('maintenance.modal.cost')" />

        <DialogFooter>
          <Button type="submit">
            <MdiPost />
            {{ entry.id ? $t("maintenance.modal.edit_action") : $t("maintenance.modal.new_action") }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { DialogID } from "@/components/ui/dialog-provider/utils";
  import { toast } from "@/components/ui/sonner";
  import MdiPost from "~icons/mdi/post";
  import DatePicker from "~~/components/Form/DatePicker.vue";
  import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import { useDialog } from "@/components/ui/dialog-provider";
  import FormTextField from "~/components/Form/TextField.vue";
  import FormTextArea from "~/components/Form/TextArea.vue";
  import { MaintenancePlanIntervalUnit } from "~~/lib/api/types/data-contracts";
  import type { MaintenanceEntryWithDetails, MaintenancePlanUpdate } from "~~/lib/api/types/data-contracts";
  import { getNextNDueDates } from "~/lib/maintenance/recurrence";
  import { parseDateOnly, toDateOnlyString } from "~/lib/datelib/dateOnly";
  import Button from "@/components/ui/button/Button.vue";

  const { closeDialog, registerOpenDialogCallback } = useDialog();

  const { t } = useI18n();
  const api = useUserApi();

  const entry = reactive({
    id: null as string | null,
    name: "",
    completedDate: "",
    scheduledDate: "",
    description: "",
    cost: "",
    planID: null as string | null,
    isRecurring: false,
    intervalValue: "30",
    intervalUnit: MaintenancePlanIntervalUnit.Day,
    itemIds: null as string[] | null,
    itemIdForPlanLookup: null as string | null,
  });

  /** Snapshot when the dialog opened / plan hydrated — used to avoid overwriting plan nextDueAt when unchanged */
  const planFieldBaseline = reactive({
    scheduledDate: "",
    completedDate: "",
    intervalValue: "30",
    intervalUnit: MaintenancePlanIntervalUnit.Day,
  });

  function capturePlanFieldBaseline() {
    planFieldBaseline.scheduledDate = entry.scheduledDate;
    planFieldBaseline.completedDate = entry.completedDate;
    planFieldBaseline.intervalValue = entry.intervalValue;
    planFieldBaseline.intervalUnit = entry.intervalUnit;
  }

  function sameCalendarDate(a: string, b: string): boolean {
    return a === b;
  }

  function planScheduleOrRecurrenceChanged(): boolean {
    return (
      !sameCalendarDate(entry.scheduledDate, planFieldBaseline.scheduledDate) ||
      !sameCalendarDate(entry.completedDate, planFieldBaseline.completedDate) ||
      entry.intervalValue !== planFieldBaseline.intervalValue ||
      entry.intervalUnit !== planFieldBaseline.intervalUnit
    );
  }

  const nextDuePreview = computed(() => {
    if (!entry.isRecurring) {
      return [];
    }

    const intervalValue = Math.max(parseInt(entry.intervalValue, 10) || 1, 1);
    const baseDate = parseDateOnly(entry.scheduledDate) ?? parseDateOnly(entry.completedDate) ?? new Date();

    return getNextNDueDates(baseDate, intervalValue, entry.intervalUnit, 3);
  });

  function getItemIdFromEntry(maintenanceEntry: unknown): string | null {
    if (!maintenanceEntry || typeof maintenanceEntry !== "object") {
      return null;
    }

    const candidate = (maintenanceEntry as MaintenanceEntryWithDetails).itemID;
    return typeof candidate === "string" && candidate.length > 0 ? candidate : null;
  }

  function normalizePlanId(planId: unknown): string | null {
    if (typeof planId !== "string") {
      return null;
    }

    const normalized = planId.trim().toLowerCase();
    if (!normalized || normalized === "00000000-0000-0000-0000-000000000000") {
      return null;
    }

    return planId;
  }

  async function hydrateRecurringPlan() {
    try {
      if (!entry.planID || !entry.itemIdForPlanLookup) {
        return;
      }

      const { data, error } = await api.items.maintenance.getPlans(entry.itemIdForPlanLookup);
      if (error || !data) {
        return;
      }

      const currentPlan = data.find(plan => plan.id === entry.planID);
      if (!currentPlan) {
        return;
      }

      entry.isRecurring = true;
      entry.intervalValue = currentPlan.intervalValue.toString();
      entry.intervalUnit = currentPlan.intervalUnit;
    } finally {
      capturePlanFieldBaseline();
    }
  }

  async function dispatchFormSubmit() {
    if (entry.id) {
      await editEntry();
      return;
    }

    await createEntry();
  }

  async function createEntry() {
    if (!entry.itemIds) {
      return;
    }

    const isRecurring = entry.isRecurring === true;

    await Promise.allSettled(
      entry.itemIds.map(async itemId => {
        if (isRecurring) {
          const firstDueDate = entry.scheduledDate || entry.completedDate || toDateOnlyString(new Date());
          const { error: planError } = await api.items.maintenance.createPlan(itemId, {
            name: entry.name,
            description: entry.description,
            active: true,
            intervalValue: parseInt(entry.intervalValue, 10) || 1,
            intervalUnit: entry.intervalUnit,
            startDate: firstDueDate,
          });

          if (planError) {
            toast.error(t("maintenance.toast.failed_to_create"));
          }

          return;
        }

        const { error } = await api.items.maintenance.create(itemId, {
          name: entry.name,
          completedDate: entry.completedDate,
          scheduledDate: entry.scheduledDate,
          description: entry.description,
          cost: parseFloat(entry.cost) ? entry.cost : "0",
        });

        if (error) {
          toast.error(t("maintenance.toast.failed_to_create"));
          return;
        }
      })
    );

    closeDialog(DialogID.EditMaintenance, true);
  }

  async function editEntry() {
    if (!entry.id) {
      return;
    }

    const isRecurring = entry.isRecurring === true;
    const recurringPlanId = entry.planID;
    const shouldDisableRecurringPlan = !isRecurring && !!recurringPlanId;
    const shouldCreateRecurringPlan = isRecurring && !recurringPlanId && !!entry.itemIdForPlanLookup;

    if (shouldCreateRecurringPlan && entry.itemIdForPlanLookup && entry.id) {
      const firstDueDate = entry.scheduledDate || entry.completedDate || toDateOnlyString(new Date());
      const { data: createdPlan, error: createPlanError } = await api.items.maintenance.createPlan(entry.itemIdForPlanLookup, {
        name: entry.name,
        description: entry.description,
        active: true,
        intervalValue: Math.max(parseInt(entry.intervalValue, 10) || 1, 1),
        intervalUnit: entry.intervalUnit,
        startDate: firstDueDate,
        linkExistingEntryID: entry.id,
      });

      if (createPlanError || !createdPlan) {
        toast.error(t("maintenance.toast.failed_to_update"));
        return;
      }

      entry.planID = createdPlan.id;
    }

    const { error } = await api.maintenance.update(entry.id, {
      name: entry.name,
      completedDate: entry.completedDate,
      scheduledDate: entry.scheduledDate,
      planID: isRecurring ? entry.planID ?? undefined : undefined,
      description: entry.description,
      cost: parseFloat(entry.cost) ? entry.cost : "0",
    });

    if (error) {
      toast.error(t("maintenance.toast.failed_to_update"));
      return;
    }

    if (shouldDisableRecurringPlan && recurringPlanId) {
      const { error: planDeleteError } = await api.maintenance.deletePlan(recurringPlanId);
      if (planDeleteError) {
        toast.error(t("maintenance.toast.failed_to_update"));
        return;
      }
    }

    if (isRecurring && entry.planID) {
      const intervalValue = Math.max(parseInt(entry.intervalValue, 10) || 1, 1);
      const planPayload: MaintenancePlanUpdate = {
        name: entry.name,
        description: entry.description,
        active: true,
        intervalValue,
        intervalUnit: entry.intervalUnit,
      };
      if (planScheduleOrRecurrenceChanged()) {
        const firstDueDate = entry.scheduledDate || entry.completedDate || toDateOnlyString(new Date());
        planPayload.nextDueAt = firstDueDate;
      }
      const { error: planError } = await api.maintenance.updatePlan(entry.planID, planPayload);

      if (planError) {
        toast.error(t("maintenance.toast.failed_to_update"));
        return;
      }
    }

    closeDialog(DialogID.EditMaintenance, true);
  }

  onMounted(() => {
    const cleanup = registerOpenDialogCallback(DialogID.EditMaintenance, params => {
      switch (params.type) {
        case "create":
          entry.id = null;
          entry.name = "";
          entry.completedDate = "";
          entry.scheduledDate = "";
          entry.description = "";
          entry.cost = "";
          entry.isRecurring = false;
          entry.intervalValue = "30";
          entry.intervalUnit = MaintenancePlanIntervalUnit.Day;
          entry.planID = null;
          entry.itemIds = typeof params.itemId === "string" ? [params.itemId] : params.itemId;
          entry.itemIdForPlanLookup = typeof params.itemId === "string" ? params.itemId : null;
          break;
        case "update":
          entry.id = params.maintenanceEntry.id;
          entry.name = params.maintenanceEntry.name;
          entry.completedDate = (params.maintenanceEntry.completedDate as string) ?? "";
          entry.scheduledDate = (params.maintenanceEntry.scheduledDate as string) ?? "";
          entry.description = params.maintenanceEntry.description;
          entry.cost = params.maintenanceEntry.cost;
          entry.planID = normalizePlanId(params.maintenanceEntry.planID);
          entry.isRecurring = !!entry.planID;
          entry.intervalValue = "30";
          entry.intervalUnit = MaintenancePlanIntervalUnit.Day;
          entry.itemIdForPlanLookup = params.itemId ?? getItemIdFromEntry(params.maintenanceEntry);
          entry.itemIds = null;
          if (!entry.planID) {
            capturePlanFieldBaseline();
          }
          void hydrateRecurringPlan();
          break;
        case "duplicate":
          entry.id = null;
          entry.name = params.maintenanceEntry.name;
          entry.completedDate = "";
          entry.scheduledDate = "";
          entry.description = params.maintenanceEntry.description;
          entry.cost = params.maintenanceEntry.cost;
          entry.planID = null;
          entry.isRecurring = false;
          entry.itemIds = [params.itemId];
          entry.itemIdForPlanLookup = params.itemId;
          break;
      }
    });

    onUnmounted(cleanup);
  });
</script>
