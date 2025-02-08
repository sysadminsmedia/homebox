<script setup lang="ts">
  import { route } from "../../lib/api/base";
  import MdiPrinterPos from "~icons/mdi/printer-pos";
  import MdiFileDownload from "~icons/mdi/file-download";

  const props = defineProps<{
    type: string;
    id: string;
  }>();

  const pubApi = usePublicApi();
  const toast = useNotifier();

  const { data: status } = useAsyncData(async () => {
    const { data, error } = await pubApi.status();
    if (error) {
      toast.error("Failed to load status");
      return;
    }

    return data;
  });

  const printModal = ref(false);
  const printing = ref(false);

  function openPrint() {
    printModal.value = true;
  }

  async function print() {
    printing.value = true;
    const { error } = await fetch(getLabelUrl(true));

    if (error) {
      printing.value = false;
      toast.error("Failed to print label");
      return;
    }

    toast.success("Label printed");
    printModal.value = false;
    printing.value = false;
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
</script>

<template>
  <div>
    <BaseModal v-model="printModal">
      <template #title>{{ $t("components.global.label_maker.print") }}</template>
      <p>
        {{ $t("components.global.label_maker.confirm_description") }}
      </p>
      <img :src="getLabelUrl(false)" />
      <div class="modal-action">
        <BaseButton type="submit" :loading="printing" @click="print">{{
          $t("components.global.label_maker.print")
        }}</BaseButton>
      </div>
    </BaseModal>

    <div class="dropdown dropdown-left">
      <slot>
        <label tabindex="0" class="btn btn-sm">
          {{ $t("components.global.label_maker.titles") }}
        </label>
      </slot>
      <ul class="menu dropdown-content compact rounded-box bg-base-100 w-52 shadow-lg">
        <li v-if="status?.labelPrinting || false">
          <button @click="openPrint">
            <MdiPrinterPos name="mdi-printer-pos" class="mr-2" />
            {{ $t("components.global.label_maker.print") }}
          </button>
        </li>
        <li>
          <button @click="downloadLabel">
            <MdiFileDownload name="mdi-file-download" class="mr-2" />
            {{ $t("components.global.label_maker.download") }}
          </button>
        </li>
      </ul>
    </div>
  </div>
</template>
