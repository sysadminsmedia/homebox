<script setup lang="ts">
  import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import { route } from "~/lib/api/base";
  import MdiQrcode from "~icons/mdi/qrcode";
  import { Button } from "@/components/ui/button";
  import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
  import { useDialog } from "@/components/ui/dialog-provider";

  const { openDialog } = useDialog();

  function getQRCodeUrl(): string {
    const currentURL = window.location.href;
    // Adjust route import as needed
    return route(`/qrcode`, { data: encodeURIComponent(currentURL) });
  }
</script>

<template>
  <Dialog dialog-id="page-qr-code">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>
          {{ $t("components.global.page_qr_code.page_url") }}
        </DialogTitle>
      </DialogHeader>
      <img :src="getQRCodeUrl()" />
    </DialogContent>
  </Dialog>

  <Tooltip>
    <TooltipTrigger as-child>
      <Button size="icon" @click="openDialog('page-qr-code')">
        <MdiQrcode name="mdi-qrcode" />
      </Button>
    </TooltipTrigger>
    <TooltipContent>
      {{ $t("components.global.page_qr_code.qr_tooltip") }}
    </TooltipContent>
  </Tooltip>
</template>
