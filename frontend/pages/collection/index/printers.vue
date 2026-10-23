<template>
  <div>
    <BaseContainer class="m-0 flex flex-col gap-4 px-0">
      <BaseCard>
        <template #title>
          <BaseSectionHeader>
            <MdiPrinter class="mr-2" />
            <span>{{ $t("pages.printers.title") }}</span>
            <template #description>{{ $t("pages.printers.subtitle") }}</template>
          </BaseSectionHeader>
        </template>

        <div class="border-t">
          <div v-if="printersPending" class="flex items-center justify-center py-8">
            <MdiLoading class="size-6 animate-spin text-muted-foreground" />
          </div>

          <table v-else class="w-full text-sm">
            <thead>
              <tr class="text-left text-xs uppercase tracking-wide text-muted-foreground">
                <th class="px-6 py-2 font-semibold">{{ $t("pages.printers.table.printer") }}</th>
                <th class="hidden px-3 py-2 font-semibold md:table-cell">
                  {{ $t("pages.printers.table.connection") }}
                </th>
                <th class="hidden px-3 py-2 font-semibold lg:table-cell">{{ $t("pages.printers.table.media") }}</th>
                <th class="px-3 py-2 font-semibold">{{ $t("pages.printers.table.status") }}</th>
                <th class="px-6 py-2 text-right font-semibold">{{ $t("pages.printers.table.actions") }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="printer in printers ?? []" :key="printer.id" class="border-t">
                <td class="px-6 py-3">
                  <div class="flex flex-col gap-0.5">
                    <span class="flex items-center gap-2 font-medium">
                      {{ printer.name }}
                      <Badge
                        v-if="printer.isDefault"
                        class="whitespace-nowrap bg-amber-100 text-amber-900 hover:bg-amber-100"
                      >
                        {{ $t("pages.printers.default") }}
                      </Badge>
                    </span>
                    <span class="text-xs text-muted-foreground">
                      {{ printerTypeLabel(printer.printerType) }} · {{ printer.dpi }} dpi
                    </span>
                  </div>
                </td>
                <td class="hidden p-3 font-mono text-xs text-muted-foreground md:table-cell">
                  {{ printer.address }}
                </td>
                <td class="hidden p-3 text-muted-foreground lg:table-cell">
                  <template v-if="printer.labelWidthMm">
                    {{ printer.labelWidthMm
                    }}<template v-if="printer.labelHeightMm"> × {{ printer.labelHeightMm }}</template> mm
                  </template>
                  <template v-else>—</template>
                </td>
                <td class="p-3">
                  <span class="inline-flex items-center gap-2" :class="statusClass(printer.status)">
                    <span class="size-2 rounded-full" :class="statusDot(printer.status)" />
                    {{ statusLabel(printer.status) }}
                  </span>
                </td>
                <td class="px-6 py-3 text-right">
                  <div class="inline-flex gap-1.5">
                    <Button
                      variant="outline"
                      size="sm"
                      :disabled="busyId === printer.id"
                      @click="handleTestPrint(printer.id)"
                    >
                      <MdiLoading v-if="busyId === printer.id" class="mr-1 size-4 animate-spin" />
                      {{ $t("pages.printers.test_print") }}
                    </Button>
                    <DropdownMenu>
                      <DropdownMenuTrigger as-child>
                        <Button
                          variant="outline"
                          size="icon"
                          class="size-8"
                          :aria-label="$t('pages.printers.more_actions', { name: printer.name })"
                        >
                          <MdiDotsHorizontal class="size-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end" class="w-48">
                        <DropdownMenuItem @click="handleCheckStatus(printer.id)">{{
                          $t("pages.printers.check_status")
                        }}</DropdownMenuItem>
                        <DropdownMenuItem @click="handleEditPrinter(printer)">{{ $t("global.edit") }}</DropdownMenuItem>
                        <DropdownMenuItem v-if="!printer.isDefault" @click="handleSetDefault(printer.id)">
                          {{ $t("pages.printers.set_default") }}
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          class="text-destructive focus:text-destructive"
                          @click="handleDeletePrinter(printer.id)"
                        >
                          {{ $t("global.delete") }}
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                </td>
              </tr>

              <!-- Local (browser-paired) printers -->
              <tr v-for="printer in localPrinters" :key="printer.id" class="border-t bg-muted/30">
                <td class="px-6 py-3">
                  <div class="flex flex-col gap-0.5">
                    <span class="flex items-center gap-2 font-medium">
                      {{ printer.name }}
                      <Badge variant="secondary">{{ $t("pages.printers.this_device_only") }}</Badge>
                    </span>
                    <span class="text-xs text-muted-foreground">
                      {{ printer.connectionType === "usb" ? "USB" : "Bluetooth" }} · {{ printer.protocol }}
                    </span>
                  </div>
                </td>
                <td class="hidden p-3 text-xs text-muted-foreground md:table-cell">
                  {{ printer.connectionType === "usb" ? "WebUSB" : "Web Bluetooth" }}
                </td>
                <td class="hidden p-3 text-muted-foreground lg:table-cell">—</td>
                <td class="p-3 text-muted-foreground">
                  <span class="inline-flex items-center gap-2"
                    ><span class="size-2 rounded-full bg-gray-400" />{{ $t("pages.printers.status_paired") }}</span
                  >
                </td>
                <td class="px-6 py-3 text-right">
                  <Button variant="outline" size="sm" @click="handleRemoveLocalPrinter(printer.id)">
                    {{ $t("pages.printers.forget") }}
                  </Button>
                </td>
              </tr>

              <tr v-if="(printers ?? []).length === 0 && localPrinters.length === 0" class="border-t">
                <td colspan="5" class="px-6 py-8 text-center text-muted-foreground">
                  {{ $t("pages.printers.no_printers") }}
                </td>
              </tr>
            </tbody>
          </table>

          <div
            class="flex flex-wrap items-center justify-between gap-3 border-t px-6 py-3 text-sm text-muted-foreground"
          >
            <span v-if="isSupported.usb || isSupported.bluetooth">{{ $t("pages.printers.local_note") }}</span>
            <span v-else>{{ $t("pages.printers.browser_not_supported") }}</span>
            <div class="flex flex-wrap gap-2">
              <Button v-if="isSupported.usb" variant="outline" size="sm" @click="handlePairUSB">
                <MdiUsb class="mr-2 size-4" />
                {{ $t("pages.printers.pair_usb") }}
              </Button>
              <Button v-if="isSupported.bluetooth" variant="outline" size="sm" @click="handlePairBluetooth">
                <MdiBluetooth class="mr-2 size-4" />
                {{ $t("pages.printers.pair_bluetooth") }}
              </Button>
              <Button size="sm" @click="showAddPrinterDialog = true">
                <MdiPlus class="mr-2 size-4" />
                {{ $t("pages.printers.add_printer") }}
              </Button>
            </div>
          </div>
        </div>
      </BaseCard>

      <!-- Printing defaults -->
      <BaseCard>
        <template #title>
          <BaseSectionHeader>
            <MdiPrinterPos class="mr-2" />
            <span>{{ $t("pages.printers.defaults_title") }}</span>
            <template #description>{{ $t("pages.printers.defaults_sub") }}</template>
          </BaseSectionHeader>
        </template>
        <div class="grid gap-6 border-t p-6 md:grid-cols-2">
          <div class="space-y-2">
            <Label>{{ $t("pages.printers.default_template") }}</Label>
            <Select
              :model-value="preferences.defaultTemplateId ?? 'none'"
              @update:model-value="v => (preferences.defaultTemplateId = v === 'none' ? null : String(v))"
            >
              <SelectTrigger>
                <SelectValue :placeholder="$t('pages.printers.default_template_none')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="none">{{ $t("pages.printers.default_template_none") }}</SelectItem>
                <SelectItem v-for="tpl in templates ?? []" :key="tpl.id" :value="tpl.id">{{ tpl.name }}</SelectItem>
              </SelectContent>
            </Select>
            <p class="text-xs text-muted-foreground">{{ $t("pages.printers.default_template_hint") }}</p>
          </div>
          <div class="space-y-2">
            <Label>{{ $t("pages.printers.address_policy") }}</Label>
            <p class="text-sm text-muted-foreground">
              {{ $t("pages.printers.address_policy_text") }}
              <code class="rounded bg-muted px-1 py-0.5 text-xs">HBOX_PRINTER_ALLOW_PUBLIC_ADDRESSES=true</code>
            </p>
          </div>
        </div>
      </BaseCard>
    </BaseContainer>

    <!-- Add/Edit Printer Dialog -->
    <DialogRoot :open="showAddPrinterDialog" @update:open="showAddPrinterDialog = $event">
      <DialogScrollContent class="max-w-md">
        <DialogHeader>
          <DialogTitle>
            {{ editingPrinter ? $t("pages.printers.edit_printer") : $t("pages.printers.add_printer") }}
          </DialogTitle>
        </DialogHeader>
        <div class="space-y-4">
          <div class="space-y-2">
            <Label for="printer-name">{{ $t("pages.printers.form.name") }}</Label>
            <Input
              id="printer-name"
              v-model="printerForm.name"
              :placeholder="$t('pages.printers.form.name_placeholder')"
            />
          </div>
          <div class="space-y-2">
            <Label for="printer-type">{{ $t("pages.printers.form.type") }}</Label>
            <Select v-model="printerForm.printerType">
              <SelectTrigger id="printer-type">
                <SelectValue :placeholder="$t('pages.printers.form.type_placeholder')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="ipp">IPP (Network Printer)</SelectItem>
                <SelectItem value="cups">CUPS (Local Printer Server)</SelectItem>
                <SelectItem value="brother_raster">Brother Raster (QL Series)</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div class="space-y-2">
            <Label for="printer-address">{{ $t("pages.printers.form.address") }}</Label>
            <Input
              id="printer-address"
              v-model="printerForm.address"
              :placeholder="$t('pages.printers.form.address_placeholder')"
            />
            <p class="text-xs text-muted-foreground">{{ $t("pages.printers.form.address_hint") }}</p>
          </div>
          <div class="space-y-2">
            <Label for="printer-dpi">{{ $t("pages.printers.form.dpi") }}</Label>
            <Input id="printer-dpi" v-model.number="printerForm.dpi" type="number" placeholder="300" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showAddPrinterDialog = false">{{ $t("global.cancel") }}</Button>
          <Button :disabled="isSaving" @click="handleSavePrinter">
            <MdiLoading v-if="isSaving" class="mr-2 size-4 animate-spin" />
            {{ $t("global.save") }}
          </Button>
        </DialogFooter>
      </DialogScrollContent>
    </DialogRoot>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import MdiPrinter from "~icons/mdi/printer";
  import MdiPrinterPos from "~icons/mdi/printer-pos";
  import MdiLoading from "~icons/mdi/loading";
  import MdiPlus from "~icons/mdi/plus";
  import MdiDotsHorizontal from "~icons/mdi/dots-horizontal";
  import MdiUsb from "~icons/mdi/usb";
  import MdiBluetooth from "~icons/mdi/bluetooth";
  import { Button } from "@/components/ui/button";
  import { Input } from "@/components/ui/input";
  import { Label } from "@/components/ui/label";
  import { Badge } from "@/components/ui/badge";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
  } from "@/components/ui/dropdown-menu";
  import { DialogRoot } from "reka-ui";
  import { DialogScrollContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import BaseContainer from "@/components/Base/Container.vue";
  import BaseCard from "@/components/Base/Card.vue";
  import BaseSectionHeader from "@/components/Base/SectionHeader.vue";
  import type { PrinterSummary, PrinterCreate, PrinterUpdate } from "~~/lib/api/types/data-contracts";

  const { t } = useI18n();

  definePageMeta({
    middleware: ["auth"],
  });

  useHead({
    title: "HomeBox | " + t("collection.tabs.printers"),
  });

  const confirm = useConfirm();
  const preferences = useViewPreferences();

  const { printers, pending: printersPending, refresh: refreshPrinters } = usePrinters();
  const { create, update, remove, setDefault, testPrint, getStatus } = usePrinterActions();
  const api = useUserApi();
  const { templates } = useLabelTemplates();

  const {
    localPrinters,
    isSupported,
    pairUSBPrinter,
    pairBluetoothPrinter,
    removePrinter: removeLocalPrinter,
  } = useLocalPrinters();

  const busyId = ref<string | null>(null);

  function printerTypeLabel(type: string) {
    switch (type) {
      case "brother_raster":
        return t("pages.printers.types.brother_raster");
      case "cups":
        return t("pages.printers.types.cups");
      default:
        return t("pages.printers.types.ipp");
    }
  }
  function statusLabel(status: string) {
    return t(`pages.printers.status_${status || "unknown"}`);
  }
  function statusDot(status: string) {
    return { online: "bg-green-500", offline: "bg-red-500" }[status] ?? "bg-gray-400";
  }
  function statusClass(status: string) {
    return { online: "text-green-700", offline: "text-red-700" }[status] ?? "text-muted-foreground";
  }

  // Dialog state
  const showAddPrinterDialog = ref(false);
  const editingPrinter = ref<PrinterSummary | null>(null);
  const isSaving = ref(false);

  const printerForm = ref<PrinterCreate>({
    name: "",
    printerType: "ipp",
    address: "",
    description: "",
    dpi: 300,
    isDefault: false,
  });

  function resetForm() {
    printerForm.value = { name: "", printerType: "ipp", address: "", description: "", dpi: 300, isDefault: false };
    editingPrinter.value = null;
  }

  async function handleEditPrinter(printer: PrinterSummary) {
    editingPrinter.value = printer;
    // Update replaces every field and the summary has no media type, so load
    // the full printer first to keep its label size and media type intact.
    const { data: full } = await api.printers.get(printer.id);
    printerForm.value = {
      name: printer.name,
      printerType: printer.printerType as "ipp" | "cups" | "brother_raster",
      address: printer.address,
      description: printer.description || "",
      dpi: printer.dpi || 300,
      isDefault: printer.isDefault,
      labelWidthMm: full?.labelWidthMm || printer.labelWidthMm || null,
      labelHeightMm: full?.labelHeightMm || printer.labelHeightMm || null,
      mediaType: full?.mediaType || null,
    };
    showAddPrinterDialog.value = true;
  }

  async function handleSavePrinter() {
    if (!printerForm.value.name || !printerForm.value.address) {
      toast.error(t("pages.printers.toast.required_fields"));
      return;
    }
    isSaving.value = true;
    try {
      if (editingPrinter.value) {
        await update(editingPrinter.value.id, printerForm.value as PrinterUpdate);
        toast.success(t("pages.printers.toast.updated"));
      } else {
        await create(printerForm.value);
        toast.success(t("pages.printers.toast.created"));
      }
      showAddPrinterDialog.value = false;
      resetForm();
      await refreshPrinters();
    } catch {
      toast.error(t("pages.printers.toast.save_failed"));
    } finally {
      isSaving.value = false;
    }
  }

  async function handleDeletePrinter(id: string) {
    const { isCanceled } = await confirm.open(t("pages.printers.delete_confirm"));
    if (isCanceled) return;
    try {
      await remove(id);
      toast.success(t("pages.printers.toast.deleted"));
      await refreshPrinters();
    } catch {
      toast.error(t("pages.printers.toast.delete_failed"));
    }
  }

  async function handleSetDefault(id: string) {
    try {
      await setDefault(id);
      toast.success(t("pages.printers.toast.set_default"));
      await refreshPrinters();
    } catch {
      toast.error(t("pages.printers.toast.set_default_failed"));
    }
  }

  async function handleTestPrint(id: string) {
    busyId.value = id;
    try {
      const result = await testPrint(id);
      if (result?.success) {
        toast.success(t("pages.printers.toast.test_success"));
      } else {
        toast.error(result?.message || t("pages.printers.toast.test_failed"));
      }
    } catch {
      toast.error(t("pages.printers.toast.test_failed"));
    } finally {
      busyId.value = null;
    }
  }

  async function handleCheckStatus(id: string) {
    busyId.value = id;
    try {
      const result = await getStatus(id);
      toast.success(t("pages.printers.toast.status_checked", { status: statusLabel(result?.status ?? "unknown") }));
      await refreshPrinters();
    } catch {
      toast.error(t("pages.printers.toast.status_failed"));
    } finally {
      busyId.value = null;
    }
  }

  async function handlePairUSB() {
    try {
      const printer = await pairUSBPrinter();
      if (printer) toast.success(t("pages.printers.toast.paired", { name: printer.name }));
    } catch (error) {
      toast.error((error as Error).message);
    }
  }

  async function handlePairBluetooth() {
    try {
      const printer = await pairBluetoothPrinter();
      if (printer) toast.success(t("pages.printers.toast.paired", { name: printer.name }));
    } catch (error) {
      toast.error((error as Error).message);
    }
  }

  function handleRemoveLocalPrinter(id: string) {
    removeLocalPrinter(id);
    toast.success(t("pages.printers.toast.removed"));
  }

  watch(showAddPrinterDialog, isOpen => {
    if (!isOpen) resetForm();
  });
</script>
