<template>
  <div>
    <BaseContainer class="flex flex-col gap-4">
      <!-- Server Printers Section -->
      <BaseCard>
        <template #title>
          <BaseSectionHeader>
            <MdiPrinter class="mr-2" />
            <span>{{ $t("pages.printers.network_printers") }}</span>
            <template #description>{{ $t("pages.printers.network_printers_sub") }}</template>
          </BaseSectionHeader>
        </template>
        <div class="border-t p-4">
          <div v-if="printersPending" class="flex items-center justify-center py-8">
            <MdiLoading class="size-6 animate-spin text-muted-foreground" />
          </div>

          <div v-else-if="!printers || printers.length === 0" class="py-8 text-center text-muted-foreground">
            {{ $t("pages.printers.no_printers") }}
          </div>

          <div v-else class="divide-y">
            <div v-for="printer in printers" :key="printer.id" class="flex items-center justify-between py-3">
              <div class="flex items-center gap-3">
                <div class="flex size-10 items-center justify-center rounded-lg bg-muted">
                  <MdiPrinter class="size-5" />
                </div>
                <div>
                  <div class="flex items-center gap-2">
                    <span class="font-medium">{{ printer.name }}</span>
                    <span v-if="printer.isDefault" :class="badgeVariants({ variant: 'secondary' })">
                      {{ $t("pages.printers.default") }}
                    </span>
                  </div>
                  <div class="text-sm text-muted-foreground">
                    {{ printer.address }}
                  </div>
                </div>
              </div>
              <div class="flex items-center gap-2">
                <Button variant="ghost" size="icon" @click="handleTestPrint(printer.id)">
                  <MdiTestTube class="size-4" />
                </Button>
                <Button variant="ghost" size="icon" @click="handleEditPrinter(printer)">
                  <MdiPencil class="size-4" />
                </Button>
                <Button v-if="!printer.isDefault" variant="ghost" size="icon" @click="handleSetDefault(printer.id)">
                  <MdiStar class="size-4" />
                </Button>
                <Button variant="ghost" size="icon" @click="handleDeletePrinter(printer.id)">
                  <MdiDelete class="size-4" />
                </Button>
              </div>
            </div>
          </div>

          <div class="mt-4 flex justify-end">
            <Button @click="showAddPrinterDialog = true">
              <MdiPlus class="mr-2 size-4" />
              {{ $t("pages.printers.add_printer") }}
            </Button>
          </div>
        </div>
      </BaseCard>

      <!-- Local Printers Section (WebUSB/Bluetooth) -->
      <BaseCard>
        <template #title>
          <BaseSectionHeader>
            <MdiUsb class="mr-2" />
            <span>{{ $t("pages.printers.local_printers") }}</span>
            <template #description>{{ $t("pages.printers.local_printers_sub") }}</template>
          </BaseSectionHeader>
        </template>
        <div class="border-t p-4">
          <div
            v-if="!isSupported.usb && !isSupported.bluetooth"
            class="rounded-lg bg-amber-50 p-4 text-amber-800 dark:bg-amber-950 dark:text-amber-200"
          >
            <div class="flex items-center gap-2">
              <MdiAlert class="size-5" />
              <span>{{ $t("pages.printers.browser_not_supported") }}</span>
            </div>
            <p class="mt-2 text-sm">{{ $t("pages.printers.browser_not_supported_hint") }}</p>
          </div>

          <template v-else>
            <div v-if="localPrinters.length === 0" class="py-8 text-center text-muted-foreground">
              {{ $t("pages.printers.no_local_printers") }}
            </div>

            <div v-else class="divide-y">
              <div v-for="printer in localPrinters" :key="printer.id" class="flex items-center justify-between py-3">
                <div class="flex items-center gap-3">
                  <div class="flex size-10 items-center justify-center rounded-lg bg-muted">
                    <MdiUsb v-if="printer.connectionType === 'usb'" class="size-5" />
                    <MdiBluetooth v-else class="size-5" />
                  </div>
                  <div>
                    <div class="font-medium">{{ printer.name }}</div>
                    <div class="text-sm text-muted-foreground">
                      {{ printer.connectionType === "usb" ? "USB" : "Bluetooth" }}
                      <span class="ml-2">({{ printer.protocol }})</span>
                    </div>
                  </div>
                </div>
                <div class="flex items-center gap-2">
                  <Button variant="ghost" size="icon" @click="handleRemoveLocalPrinter(printer.id)">
                    <MdiDelete class="size-4" />
                  </Button>
                </div>
              </div>
            </div>

            <div class="mt-4 flex justify-end gap-2">
              <Button v-if="isSupported.usb" variant="outline" @click="handlePairUSB">
                <MdiUsb class="mr-2 size-4" />
                {{ $t("pages.printers.pair_usb") }}
              </Button>
              <Button v-if="isSupported.bluetooth" variant="outline" @click="handlePairBluetooth">
                <MdiBluetooth class="mr-2 size-4" />
                {{ $t("pages.printers.pair_bluetooth") }}
              </Button>
            </div>
          </template>
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
            <p class="text-xs text-muted-foreground">
              {{ $t("pages.printers.form.address_hint") }}
            </p>
          </div>
          <div class="space-y-2">
            <Label for="printer-dpi">{{ $t("pages.printers.form.dpi") }}</Label>
            <Input id="printer-dpi" v-model.number="printerForm.dpi" type="number" placeholder="300" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showAddPrinterDialog = false">
            {{ $t("global.cancel") }}
          </Button>
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
  import MdiLoading from "~icons/mdi/loading";
  import MdiPlus from "~icons/mdi/plus";
  import MdiPencil from "~icons/mdi/pencil";
  import MdiDelete from "~icons/mdi/delete";
  import MdiStar from "~icons/mdi/star";
  import MdiTestTube from "~icons/mdi/test-tube";
  import MdiUsb from "~icons/mdi/usb";
  import MdiBluetooth from "~icons/mdi/bluetooth";
  import MdiAlert from "~icons/mdi/alert";
  import { Button } from "@/components/ui/button";
  import { Input } from "@/components/ui/input";
  import { Label } from "@/components/ui/label";
  import { badgeVariants } from "@/components/ui/badge";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
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
    title: "HomeBox | " + t("menu.printers"),
  });

  const confirm = useConfirm();

  // Server printers
  const { printers, pending: printersPending, refresh: refreshPrinters } = usePrinters();
  const { create, update, remove, setDefault, testPrint } = usePrinterActions();

  // Local printers
  const {
    localPrinters,
    isSupported,
    pairUSBPrinter,
    pairBluetoothPrinter,
    removePrinter: removeLocalPrinter,
  } = useLocalPrinters();

  // Dialog state
  const showAddPrinterDialog = ref(false);
  const editingPrinter = ref<PrinterSummary | null>(null);
  const isSaving = ref(false);

  // Form state - label size is now configured per-template, not per-printer
  const printerForm = ref<PrinterCreate>({
    name: "",
    printerType: "ipp",
    address: "",
    description: "",
    dpi: 300,
    isDefault: false,
  });

  function resetForm() {
    printerForm.value = {
      name: "",
      printerType: "ipp",
      address: "",
      description: "",
      dpi: 300,
      isDefault: false,
    };
    editingPrinter.value = null;
  }

  function handleEditPrinter(printer: PrinterSummary) {
    editingPrinter.value = printer;
    printerForm.value = {
      name: printer.name,
      printerType: printer.printerType as "ipp" | "cups" | "brother_raster",
      address: printer.address,
      description: printer.description || "",
      dpi: printer.dpi || 300,
      isDefault: printer.isDefault,
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
    const confirmed = await confirm.reveal(t("pages.printers.delete_confirm"));
    if (!confirmed.data) return;

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
    try {
      const result = await testPrint(id);
      if (result?.success) {
        toast.success(t("pages.printers.toast.test_success"));
      } else {
        toast.error(result?.message || t("pages.printers.toast.test_failed"));
      }
    } catch {
      toast.error(t("pages.printers.toast.test_failed"));
    }
  }

  async function handlePairUSB() {
    try {
      const printer = await pairUSBPrinter();
      if (printer) {
        toast.success(t("pages.printers.toast.paired", { name: printer.name }));
      }
    } catch (error) {
      toast.error((error as Error).message);
    }
  }

  async function handlePairBluetooth() {
    try {
      const printer = await pairBluetoothPrinter();
      if (printer) {
        toast.success(t("pages.printers.toast.paired", { name: printer.name }));
      }
    } catch (error) {
      toast.error((error as Error).message);
    }
  }

  function handleRemoveLocalPrinter(id: string) {
    removeLocalPrinter(id);
    toast.success(t("pages.printers.toast.removed"));
  }

  // Reset form when dialog closes
  watch(showAddPrinterDialog, isOpen => {
    if (!isOpen) {
      resetForm();
    }
  });
</script>
