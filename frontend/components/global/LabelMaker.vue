<script setup lang="ts">
  import { route } from "../../lib/api/base";
  import { toast } from "@/components/ui/sonner";
  import MdiPrinterPos from "~icons/mdi/printer-pos";
  import MdiFileDownload from "~icons/mdi/file-download";
  import MdiQrcode from "~icons/mdi/qrcode";

  import {
    Dialog,
    DialogContent,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogDescription,
  } from "@/components/ui/dialog";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { Button, ButtonGroup } from "@/components/ui/button";

  const { openDialog, closeDialog } = useDialog();

  const props = defineProps<{
    type: string;
    id: string;
  }>();

  const pubApi = usePublicApi();

  const { data: status } = useAsyncData(async () => {
    const { data, error } = await pubApi.status();
    if (error) {
      toast.error("Failed to load status");
      return;
    }

    return data;
  });

  const serverPrinting = ref(false);

  function openPrint() {
    openDialog("print-label");
  }

  function browserPrint() {
    const printWindow = window.open(getLabelUrl(false), "popup=true");

    if (printWindow !== null) {
      printWindow.onload = () => {
        printWindow.print();
      };
    }
  }

  async function serverPrint() {
    serverPrinting.value = true;
    try {
      await fetch(getLabelUrl(true));
    } catch (err) {
      console.error("Failed to print labels:", err);
      serverPrinting.value = false;
      toast.error("Failed to print label");
      return;
    }

    toast.success("Label printed");
    closeDialog("print-label");
    serverPrinting.value = false;
  }

  function downloadLabel() {
    const link = document.createElement("a");
    link.download = `label-${props.id}.png`;
    link.href = getLabelUrl(false);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }

  function getLabelUrl(print: boolean): string {
    const params = { print };

    if (props.type === "item") {
      return route(`/labelmaker/item/${props.id}`, params);
    } else if (props.type === "location") {
      return route(`/labelmaker/location/${props.id}`, params);
    } else if (props.type === "asset") {
      return route(`/labelmaker/asset/${props.id}`, params);
    } else {
      throw new Error(`Unexpected labelmaker type ${props.type}`);
    }
  }

  function getQRCodeUrl(): string {
    const currentURL = window.location.href;

    return route(`/qrcode`, { data: encodeURIComponent(currentURL) });
  }
</script>

<template>
  <div>
    <Dialog dialog-id="print-label">
      <DialogContent>
        <DialogHeader>
          <DialogTitle> {{ $t("components.global.label_maker.print") }} </DialogTitle>
          <DialogDescription> {{ $t("components.global.label_maker.confirm_description") }} </DialogDescription>
        </DialogHeader>

        <img :src="getLabelUrl(false)" />

        <DialogFooter>
          <ButtonGroup>
            <Button v-if="status?.labelPrinting || false" type="submit" :loading="serverPrinting" @click="serverPrint"
              >{{ $t("components.global.label_maker.server_print") }}
            </Button>
            <Button type="submit" @click="browserPrint">{{ $t("components.global.label_maker.browser_print") }}</Button>
          </ButtonGroup>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog dialog-id="qr-code">
      <DialogContent>
        <DialogHeader>
          <DialogTitle> {{ $t("components.global.page_qr_code.page_url") }} </DialogTitle>
        </DialogHeader>

        <img :src="getQRCodeUrl()" />
      </DialogContent>
    </Dialog>

    <ButtonGroup>
      <Button variant="outline" disabled class="disabled:opacity-100">
        {{ $t("components.global.label_maker.titles") }}
      </Button>
      <Button size="icon" @click="downloadLabel">
        <MdiFileDownload name="mdi-file-download" />
      </Button>
      <Button size="icon" @click="openDialog('print-label')">
        <MdiPrinterPos name="mdi-printer-pos" />
      </Button>
      <Button size="icon" @click="openDialog('qr-code')">
        <MdiQrcode name="mdi-qrcode" />
      </Button>
    </ButtonGroup>
  </div>
</template>
