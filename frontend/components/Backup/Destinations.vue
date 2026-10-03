<template>
  <div v-if="available" class="py-3">
    <div class="mb-3 flex flex-wrap items-start justify-between gap-2">
      <div>
        <h3 class="text-base font-semibold">{{ $t("tools.backup_destinations.title") }}</h3>
        <p class="text-sm text-muted-foreground">{{ $t("tools.backup_destinations.subtitle") }}</p>
      </div>
      <Button size="sm" @click="openCreate">
        <MdiPlus class="mr-1" />
        {{ $t("tools.backup_destinations.add") }}
      </Button>
    </div>

    <div
      v-for="d in problems"
      :key="d.id"
      class="mb-2 flex items-start gap-2 rounded-md border border-destructive bg-destructive/10 p-3 text-sm"
      role="alert"
    >
      <MdiAlert class="mt-0.5 shrink-0 text-destructive" />
      <div>
        <p class="font-semibold">{{ problemTitle(d) }}</p>
        <p class="text-muted-foreground">{{ d.healthStatus === "unreachable" ? d.healthError : d.lastError }}</p>
      </div>
    </div>

    <p v-if="destinations.length === 0" class="text-sm text-muted-foreground">
      {{ $t("tools.backup_destinations.empty") }}
    </p>

    <div v-for="d in destinations" :key="d.id" class="mb-3 rounded-md border p-3">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <span class="font-medium">{{ d.name }}</span>
            <Badge variant="outline">{{ typeLabel(d.type) }}</Badge>
            <Badge v-if="!d.enabled" variant="secondary">{{ $t("tools.backup_destinations.paused") }}</Badge>
            <Badge :variant="d.healthStatus === 'unreachable' ? 'destructive' : 'secondary'">
              {{ $t(`tools.backup_destinations.health.${d.healthStatus}`) }}
            </Badge>
          </div>
          <p class="mt-1 text-xs text-muted-foreground">{{ scheduleSummary(d) }}</p>
          <p class="text-xs text-muted-foreground">
            {{ $t("tools.backup_destinations.last_backup") }}:
            {{ d.lastSuccessAt ? formatDate(d.lastSuccessAt) : $t("tools.backup_destinations.never") }}
            <template v-if="d.scheduleEnabled && d.enabled && d.nextRunAt">
              · {{ $t("tools.backup_destinations.next_run") }}: {{ formatDate(d.nextRunAt) }}
            </template>
            <template v-if="d.lastSkippedAt">
              · {{ $t("tools.backup_destinations.last_skipped") }}: {{ formatDate(d.lastSkippedAt) }}
            </template>
          </p>
        </div>
        <div class="flex flex-wrap gap-2">
          <Button size="sm" variant="outline" :disabled="!d.enabled || busy[d.id]" @click="runNow(d)">
            {{ $t("tools.backup_destinations.run_now") }}
          </Button>
          <Button size="sm" variant="outline" :disabled="busy[d.id]" @click="testSaved(d)">
            {{ $t("tools.backup_destinations.test") }}
          </Button>
          <Button size="sm" variant="outline" @click="toggleVersions(d)">
            {{ $t("tools.backup_destinations.versions") }}
          </Button>
          <Button size="sm" variant="outline" @click="openEdit(d)">{{ $t("global.edit") }}</Button>
          <Button size="sm" variant="destructive" @click="remove(d)">{{ $t("global.delete") }}</Button>
        </div>
      </div>

      <div v-if="expanded === d.id" class="mt-3 border-t pt-3">
        <table v-if="(versions[d.id] ?? []).length > 0" class="w-full text-sm">
          <thead>
            <tr class="text-left text-muted-foreground">
              <th class="py-1">{{ $t("tools.backups_set.table.created") }}</th>
              <th class="py-1">{{ $t("tools.backups_set.table.status") }}</th>
              <th class="py-1">{{ $t("tools.backups_set.table.size") }}</th>
              <th class="py-1 text-right">{{ $t("tools.backups_set.table.actions") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="v in versions[d.id]" :key="v.id" class="border-t">
              <td class="py-2">
                {{ formatDate(v.createdAt) }}
                <Badge v-if="v.origin === 'scheduled'" variant="outline" class="ml-1">
                  {{ $t("tools.backup_destinations.scheduled") }}
                </Badge>
              </td>
              <td class="py-2">
                <span>{{ v.status }}</span>
                <span v-if="v.status === 'running'"> ({{ v.progress }}%)</span>
                <span v-if="v.status === 'failed' && v.error" class="block text-xs text-destructive" :title="v.error">
                  {{ v.error }}
                </span>
              </td>
              <td class="py-2">{{ v.status === "completed" ? formatBytes(v.sizeBytes) : "—" }}</td>
              <td class="space-x-2 py-2 text-right">
                <a
                  v-if="v.status === 'completed'"
                  :href="api.backups.downloadURL(v.id)"
                  class="text-primary underline"
                  :download="`homebox-export-${v.id}.zip`"
                >
                  {{ $t("tools.backups_set.download") }}
                </a>
                <button class="text-destructive underline" @click="deleteVersion(d, v.id)">
                  {{ $t("global.delete") }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        <p v-else class="text-sm text-muted-foreground">{{ $t("tools.backups_set.list_empty") }}</p>
      </div>
    </div>

    <Dialog :dialog-id="DialogID.BackupDestination">
      <DialogScrollContent class="max-w-xl">
        <DialogHeader>
          <DialogTitle>
            {{ editingId ? $t("tools.backup_destinations.edit_title") : $t("tools.backup_destinations.add_title") }}
          </DialogTitle>
        </DialogHeader>

        <form class="grid gap-4" @submit.prevent="save">
          <div class="grid gap-1.5">
            <Label for="bd-name">{{ $t("tools.backup_destinations.name") }}</Label>
            <Input id="bd-name" v-model="form.name" required maxlength="255" />
          </div>

          <div class="grid gap-1.5">
            <Label>{{ $t("tools.backup_destinations.type") }}</Label>
            <Select v-model="form.type">
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="ty in typeOptions" :key="ty" :value="ty">{{ typeLabel(ty) }}</SelectItem>
              </SelectContent>
            </Select>
            <p class="text-xs text-muted-foreground">{{ $t(`tools.backup_destinations.type_help.${form.type}`) }}</p>
          </div>

          <div v-if="form.type !== 'primary'" class="grid gap-1.5">
            <Label for="bd-conn">
              {{
                form.type === "local"
                  ? $t("tools.backup_destinations.local_dir")
                  : $t("tools.backup_destinations.conn_string")
              }}
            </Label>
            <Input
              id="bd-conn"
              v-model="form.connString"
              required
              :placeholder="connPlaceholder"
              autocomplete="off"
              spellcheck="false"
            />
            <p v-if="form.type !== 'local'" class="text-xs text-muted-foreground">
              {{ $t("tools.backup_destinations.conn_help") }}
            </p>
          </div>

          <div v-if="form.type !== 'primary' && form.type !== 'local'" class="grid gap-1.5">
            <Label for="bd-prefix">{{ $t("tools.backup_destinations.prefix") }}</Label>
            <Input id="bd-prefix" v-model="form.prefix" maxlength="255" />
          </div>

          <div class="flex items-center justify-between rounded-md border p-3">
            <Label for="bd-enabled">{{ $t("tools.backup_destinations.enabled") }}</Label>
            <Switch id="bd-enabled" v-model="form.enabled" />
          </div>

          <div class="grid gap-3 rounded-md border p-3">
            <div class="flex items-center justify-between">
              <Label for="bd-schedule">{{ $t("tools.backup_destinations.schedule_enabled") }}</Label>
              <Switch id="bd-schedule" v-model="form.scheduleEnabled" />
            </div>
            <template v-if="form.scheduleEnabled">
              <div class="grid grid-cols-2 gap-3">
                <div class="grid gap-1.5">
                  <Label>{{ $t("tools.backup_destinations.frequency") }}</Label>
                  <Select v-model="form.frequency">
                    <SelectTrigger><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem v-for="f in frequencies" :key="f" :value="f">
                        {{ $t(`tools.backup_destinations.frequencies.${f}`) }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div v-if="form.frequency === 'hourly'" class="grid gap-1.5">
                  <Label for="bd-interval">{{ $t("tools.backup_destinations.every_hours") }}</Label>
                  <Input id="bd-interval" v-model.number="form.intervalHours" type="number" min="1" max="168" />
                </div>
                <div v-else class="grid gap-1.5">
                  <Label for="bd-time">{{ $t("tools.backup_destinations.time") }}</Label>
                  <Input id="bd-time" v-model="time" type="time" required />
                </div>
                <div v-if="form.frequency === 'hourly'" class="grid gap-1.5">
                  <Label for="bd-minute">{{ $t("tools.backup_destinations.at_minute") }}</Label>
                  <Input id="bd-minute" v-model.number="form.atMinute" type="number" min="0" max="59" />
                </div>
                <div v-if="form.frequency === 'weekly'" class="grid gap-1.5">
                  <Label>{{ $t("tools.backup_destinations.weekday") }}</Label>
                  <Select v-model="weekdayStr">
                    <SelectTrigger><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem v-for="n in 7" :key="n" :value="String(n - 1)">
                        {{ weekdayName(n - 1) }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div v-if="form.frequency === 'monthly'" class="grid gap-1.5">
                  <Label for="bd-dom">{{ $t("tools.backup_destinations.day_of_month") }}</Label>
                  <Input id="bd-dom" v-model.number="form.dayOfMonth" type="number" min="1" max="28" />
                </div>
              </div>
              <div class="flex items-center gap-2">
                <Checkbox id="bd-skip" v-model="form.skipIfUnchanged" class="size-5" />
                <Label for="bd-skip" class="cursor-pointer">{{ $t("tools.backup_destinations.skip_unchanged") }}</Label>
              </div>
            </template>
          </div>

          <div class="grid gap-2 rounded-md border p-3">
            <div>
              <Label>{{ $t("tools.backup_destinations.retention") }}</Label>
              <p class="text-xs text-muted-foreground">{{ $t("tools.backup_destinations.retention_help") }}</p>
            </div>
            <div class="grid grid-cols-3 gap-3">
              <div class="grid gap-1.5">
                <Label for="bd-kd" class="text-xs">{{ $t("tools.backup_destinations.keep_daily") }}</Label>
                <Input id="bd-kd" v-model.number="form.keepDaily" type="number" min="0" max="3650" />
              </div>
              <div class="grid gap-1.5">
                <Label for="bd-kw" class="text-xs">{{ $t("tools.backup_destinations.keep_weekly") }}</Label>
                <Input id="bd-kw" v-model.number="form.keepWeekly" type="number" min="0" max="520" />
              </div>
              <div class="grid gap-1.5">
                <Label for="bd-km" class="text-xs">{{ $t("tools.backup_destinations.keep_monthly") }}</Label>
                <Input id="bd-km" v-model.number="form.keepMonthly" type="number" min="0" max="120" />
              </div>
            </div>
          </div>

          <div class="grid gap-3 rounded-md border p-3">
            <div class="flex items-center justify-between">
              <Label for="bd-alerts">{{ $t("tools.backup_destinations.alerts_enabled") }}</Label>
              <Switch id="bd-alerts" v-model="form.alertsEnabled" />
            </div>
            <p class="text-xs text-muted-foreground">{{ $t("tools.backup_destinations.alerts_help") }}</p>
            <div class="grid grid-cols-3 gap-3">
              <div class="grid gap-1.5">
                <Label for="bd-hi" class="text-xs">{{ $t("tools.backup_destinations.health_interval") }}</Label>
                <Input id="bd-hi" v-model.number="form.healthIntervalMinutes" type="number" min="1" max="1440" />
              </div>
              <div class="grid gap-1.5">
                <Label for="bd-ft" class="text-xs">{{ $t("tools.backup_destinations.failure_threshold") }}</Label>
                <Input id="bd-ft" v-model.number="form.alertFailureThreshold" type="number" min="1" max="100" />
              </div>
              <div class="grid gap-1.5">
                <Label for="bd-sh" class="text-xs">{{ $t("tools.backup_destinations.stale_hours") }}</Label>
                <Input id="bd-sh" v-model.number="form.alertStaleHours" type="number" min="0" max="8760" />
              </div>
            </div>
          </div>

          <div
            v-if="testResult"
            class="rounded-md border p-3 text-sm"
            :class="testResult.ok ? 'border-primary bg-primary/10' : 'border-destructive bg-destructive/10'"
            role="status"
          >
            <p class="font-semibold">
              {{
                testResult.ok
                  ? $t("tools.backup_destinations.test_ok", { ms: testResult.latencyMs })
                  : $t("tools.backup_destinations.test_failed")
              }}
            </p>
            <p class="break-words text-muted-foreground">{{ testResult.message }}</p>
          </div>

          <DialogFooter class="gap-2">
            <Button type="button" variant="outline" :disabled="testing" @click="testForm">
              <MdiLoading v-if="testing" class="mr-1 animate-spin" />
              {{ $t("tools.backup_destinations.test_connection") }}
            </Button>
            <Button type="submit" :disabled="saving || !form.name.trim()">
              <MdiLoading v-if="saving" class="mr-1 animate-spin" />
              {{ $t("global.save") }}
            </Button>
          </DialogFooter>
        </form>
      </DialogScrollContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import MdiPlus from "~icons/mdi/plus";
  import MdiAlert from "~icons/mdi/alert";
  import MdiLoading from "~icons/mdi/loading";
  import { Button } from "@/components/ui/button";
  import { Badge } from "@/components/ui/badge";
  import { Input } from "@/components/ui/input";
  import { Label } from "@/components/ui/label";
  import { Switch } from "@/components/ui/switch";
  import { Checkbox } from "@/components/ui/checkbox";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import { Dialog, DialogFooter, DialogHeader, DialogScrollContent, DialogTitle } from "@/components/ui/dialog";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { DialogID } from "@/components/ui/dialog-provider/utils";
  import { ServerEvent, onServerEvent } from "@/composables/use-server-events";
  import type {
    BackupDestinationOut,
    BackupOptions,
    BackupSettings,
    ExportOut,
    TestResult,
  } from "@/lib/api/types/data-contracts";

  const { t } = useI18n();
  const api = useUserApi();
  const confirm = useConfirm();
  const { openDialog, closeDialog } = useDialog();

  // The endpoints are owner-only; a 403 (or a disabled feature) hides the whole section.
  const available = ref(false);
  const options = ref<BackupOptions | null>(null);
  const destinations = ref<BackupDestinationOut[]>([]);
  const versions = ref<Record<string, ExportOut[]>>({});
  const expanded = ref<string | null>(null);
  const busy = ref<Record<string, boolean>>({});

  const frequencies = ["hourly", "daily", "weekly", "monthly"] as const;

  const typeOptions = computed(() => {
    const all = ["primary", "local", "s3", "gcs", "azblob"];
    return all.filter(ty => ty !== "local" || options.value?.localEnabled);
  });

  function typeLabel(ty: string) {
    return t(`tools.backup_destinations.types.${ty}`);
  }

  const connPlaceholder = computed(() => {
    switch (form.type) {
      case "local":
        return "nas/daily";
      case "gcs":
        return "gcs://my-bucket";
      case "azblob":
        return "azblob://my-container";
      default:
        return "s3://my-bucket?region=us-east-1";
    }
  });

  const problems = computed(() =>
    destinations.value.filter(d => d.enabled && (d.healthStatus === "unreachable" || d.lastError))
  );

  function problemTitle(d: BackupDestinationOut) {
    return d.healthStatus === "unreachable"
      ? t("tools.backup_destinations.problem_unreachable", { name: d.name })
      : t("tools.backup_destinations.problem_failed", { name: d.name });
  }

  async function refresh() {
    const opts = await api.backups.options();
    if (opts.error || !opts.data?.enabled) {
      available.value = false;
      return;
    }
    options.value = opts.data;
    const list = await api.backups.listDestinations();
    if (list.error || !list.data) {
      available.value = false;
      return;
    }
    available.value = true;
    destinations.value = list.data.items ?? [];
    if (expanded.value) {
      await loadVersions(expanded.value);
    }
  }

  async function loadVersions(id: string) {
    const res = await api.backups.destinationVersions(id);
    if (!res.error && res.data) {
      versions.value[id] = res.data.items ?? [];
    }
  }

  async function toggleVersions(d: BackupDestinationOut) {
    if (expanded.value === d.id) {
      expanded.value = null;
      return;
    }
    expanded.value = d.id;
    await loadVersions(d.id);
  }

  refresh();
  onServerEvent(ServerEvent.ExportMutation, refresh);

  // ---- formatting -----------------------------------------------------------

  function formatDate(iso: string | Date) {
    return new Date(iso).toLocaleString();
  }

  function formatBytes(n: number): string {
    if (!n) return "0 B";
    const units = ["B", "KB", "MB", "GB"];
    let i = 0;
    let v = n;
    while (v >= 1024 && i < units.length - 1) {
      v /= 1024;
      i++;
    }
    return `${v.toFixed(v >= 10 || i === 0 ? 0 : 1)} ${units[i]}`;
  }

  function weekdayName(n: number) {
    // 2024-01-07 is a Sunday, matching Go's time.Weekday numbering used by the API.
    return new Date(2024, 0, 7 + n).toLocaleDateString(undefined, { weekday: "long" });
  }

  function pad(n: number) {
    return String(n).padStart(2, "0");
  }

  function scheduleSummary(d: BackupDestinationOut) {
    if (!d.scheduleEnabled) {
      return t("tools.backup_destinations.manual_only");
    }
    const at = `${pad(d.atHour)}:${pad(d.atMinute)}`;
    let when: string;
    switch (d.frequency) {
      case "hourly":
        when = t("tools.backup_destinations.summary_hourly", { n: d.intervalHours, minute: pad(d.atMinute) });
        break;
      case "weekly":
        when = t("tools.backup_destinations.summary_weekly", { day: weekdayName(d.weekday), at });
        break;
      case "monthly":
        when = t("tools.backup_destinations.summary_monthly", { day: d.dayOfMonth, at });
        break;
      default:
        when = t("tools.backup_destinations.summary_daily", { at });
    }
    const keep = t("tools.backup_destinations.summary_keep", {
      d: d.keepDaily,
      w: d.keepWeekly,
      m: d.keepMonthly,
    });
    return `${when} · ${keep}`;
  }

  // ---- dialog form ------------------------------------------------------------

  function defaults(): BackupSettings {
    return {
      name: "",
      description: "",
      type: "primary",
      connString: "",
      prefix: "homebox-backups",
      enabled: true,
      scheduleEnabled: true,
      frequency: "daily",
      intervalHours: 1,
      atHour: 3,
      atMinute: 0,
      weekday: 0,
      dayOfMonth: 1,
      skipIfUnchanged: true,
      keepDaily: 7,
      keepWeekly: 4,
      keepMonthly: 6,
      healthIntervalMinutes: 15,
      alertsEnabled: true,
      alertFailureThreshold: 2,
      alertStaleHours: 48,
    };
  }

  const form = reactive<BackupSettings>(defaults());
  const editingId = ref<string | null>(null);
  const saving = ref(false);
  const testing = ref(false);
  const testResult = ref<TestResult | null>(null);

  // The time input and weekday select speak strings; the API speaks numbers.
  const time = computed({
    get: () => `${pad(form.atHour)}:${pad(form.atMinute)}`,
    set: (v: string) => {
      const [h = NaN, m = NaN] = v.split(":").map(Number);
      if (!Number.isNaN(h) && !Number.isNaN(m)) {
        form.atHour = h;
        form.atMinute = m;
      }
    },
  });
  const weekdayStr = computed({
    get: () => String(form.weekday),
    set: (v: string) => {
      form.weekday = Number(v);
    },
  });

  function openCreate() {
    Object.assign(form, defaults());
    editingId.value = null;
    testResult.value = null;
    openDialog(DialogID.BackupDestination);
  }

  function openEdit(d: BackupDestinationOut) {
    const { name, description, type, connString, prefix, enabled, scheduleEnabled, frequency } = d;
    Object.assign(form, defaults(), {
      name,
      description,
      type,
      connString,
      prefix,
      enabled,
      scheduleEnabled,
      frequency,
      intervalHours: d.intervalHours,
      atHour: d.atHour,
      atMinute: d.atMinute,
      weekday: d.weekday,
      dayOfMonth: d.dayOfMonth,
      skipIfUnchanged: d.skipIfUnchanged,
      keepDaily: d.keepDaily,
      keepWeekly: d.keepWeekly,
      keepMonthly: d.keepMonthly,
      healthIntervalMinutes: d.healthIntervalMinutes,
      alertsEnabled: d.alertsEnabled,
      alertFailureThreshold: d.alertFailureThreshold,
      alertStaleHours: d.alertStaleHours,
    });
    editingId.value = d.id;
    testResult.value = null;
    openDialog(DialogID.BackupDestination);
  }

  function payload(): BackupSettings {
    return { ...form, name: form.name.trim() };
  }

  async function testForm() {
    testing.value = true;
    testResult.value = null;
    const res = await api.backups.testSettings(payload());
    testing.value = false;
    if (res.error || !res.data) {
      testResult.value = { ok: false, latencyMs: 0, message: errorMessage(res.data) };
      return;
    }
    testResult.value = res.data;
  }

  // The API returns {error: "..."} on 4xx; fall back to a generic message.
  function errorMessage(data: unknown): string {
    const msg = (data as { error?: string } | null)?.error;
    return msg ?? t("tools.backup_destinations.request_failed");
  }

  async function save() {
    saving.value = true;
    const body = payload();
    const res = editingId.value
      ? await api.backups.updateDestination(editingId.value, body)
      : await api.backups.createDestination(body);
    saving.value = false;
    if (res.error) {
      toast.error(errorMessage(res.data));
      return;
    }
    toast.success(t("tools.backup_destinations.saved"));
    closeDialog(DialogID.BackupDestination);
    await refresh();
  }

  // ---- row actions ---------------------------------------------------------------

  async function withBusy(id: string, fn: () => Promise<void>) {
    busy.value[id] = true;
    try {
      await fn();
    } finally {
      busy.value[id] = false;
    }
  }

  function runNow(d: BackupDestinationOut) {
    return withBusy(d.id, async () => {
      const res = await api.backups.runDestination(d.id);
      if (res.error) {
        toast.error(errorMessage(res.data));
        return;
      }
      toast.success(t("tools.backup_destinations.run_started", { name: d.name }));
      await refresh();
    });
  }

  function testSaved(d: BackupDestinationOut) {
    return withBusy(d.id, async () => {
      const res = await api.backups.testDestination(d.id);
      if (res.error || !res.data) {
        toast.error(errorMessage(res.data));
        return;
      }
      if (res.data.ok) {
        toast.success(t("tools.backup_destinations.test_ok", { ms: res.data.latencyMs }));
      } else {
        toast.error(`${t("tools.backup_destinations.test_failed")}: ${res.data.message}`);
      }
      await refresh();
    });
  }

  async function remove(d: BackupDestinationOut) {
    const { isCanceled } = await confirm.open(t("tools.backup_destinations.delete_confirm", { name: d.name }));
    if (isCanceled) {
      return;
    }
    const res = await api.backups.deleteDestination(d.id);
    if (res.error) {
      toast.error(errorMessage(res.data));
      return;
    }
    if (expanded.value === d.id) {
      expanded.value = null;
    }
    await refresh();
  }

  async function deleteVersion(d: BackupDestinationOut, id: string) {
    const { isCanceled } = await confirm.open(t("tools.backups_set.delete_confirm"));
    if (isCanceled) {
      return;
    }
    const res = await api.backups.delete(id);
    if (res.error) {
      toast.error(t("tools.toast.backup_delete_failed"));
      return;
    }
    await loadVersions(d.id);
  }
</script>
