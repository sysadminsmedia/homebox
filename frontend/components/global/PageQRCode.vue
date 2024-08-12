<template>
  <div class="dropdown dropdown-left">
    <slot>
      <label tabindex="0" class="btn btn-circle btn-sm">
        <MdiQrcode />
      </label>
    </slot>
    <div tabindex="0" class="card compact dropdown-content shadow-lg bg-base-100 rounded-box w-64">
      <div class="card-body">
        <h2 class="text-center">{{ $t("components.global.page_qr_code.page_url") }}</h2>
        <img :src="getQRCodeUrl()" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { route } from "../../lib/api/base";
  import MdiQrcode from "~icons/mdi/qrcode";

  function getQRCodeUrl(): string {
    const currentURL = window.location.href;

    return route(`/qrcode`, { data: encodeURIComponent(currentURL) });
  }
</script>

<style lang="scss" scoped></style>
