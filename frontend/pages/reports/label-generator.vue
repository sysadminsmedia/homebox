<script setup lang="ts">
  import { route } from "../../lib/api/base";

  definePageMeta({
    middleware: ["auth"],
    layout: false,
  });
  useHead({
    title: "Homebox | Printer",
  });

  const api = useUserApi();

  const bordered = ref(false);

  const displayProperties = reactive({
    baseURL: window.location.origin,
    assetRange: 1,
    assetRangeMax: 91,
    gapY: 0.25,
    columns: 3,
    cardHeight: 1,
    cardWidth: 2.63,
    pageWidth: 8.5,
    pageHeight: 11,
    pageTopPadding: 0.52,
    pageBottomPadding: 0.42,
    pageLeftPadding: 0.25,
    pageRightPadding: 0.1,
  });

  type Input = {
    page: {
      height: number;
      width: number;
      pageTopPadding: number;
      pageBottomPadding: number;
      pageLeftPadding: number;
      pageRightPadding: number;
    };
    cardHeight: number;
    cardWidth: number;
  };

  type Output = {
    cols: number;
    rows: number;
    gapY: number;
    gapX: number;
    card: {
      width: number;
      height: number;
    };
    page: {
      width: number;
      height: number;
      pt: number;
      pb: number;
      pl: number;
      pr: number;
    };
  };

  const notifier = useNotifier();

  function calculateGridData(input: Input): Output {
    const { page, cardHeight, cardWidth } = input;

    const availablePageWidth = page.width - page.pageLeftPadding - page.pageRightPadding;
    const availablePageHeight = page.height - page.pageTopPadding - page.pageBottomPadding;

    if (availablePageWidth < cardWidth || availablePageHeight < cardHeight) {
      notifier.error("Page size is too small for the card size");
      return out.value;
    }

    const cols = Math.floor(availablePageWidth / cardWidth);
    const rows = Math.floor(availablePageHeight / cardHeight);
    const gapX = (availablePageWidth - cols * cardWidth) / (cols - 1);
    const gapY = (page.height - rows * cardHeight) / (rows - 1);

    return {
      cols,
      rows,
      gapX,
      gapY,
      card: {
        width: cardWidth,
        height: cardHeight,
      },
      page: {
        width: page.width,
        height: page.height,
        pt: page.pageTopPadding,
        pb: page.pageBottomPadding,
        pl: page.pageLeftPadding,
        pr: page.pageRightPadding,
      },
    };
  }

  interface InputDef {
    label: string;
    ref: keyof typeof displayProperties;
    type?: "number" | "text";
  }

  const propertyInputs = computed<InputDef[]>(() => {
    return [
      {
        label: "Asset Start",
        ref: "assetRange",
      },
      {
        label: "Asset End",
        ref: "assetRangeMax",
      },
      {
        label: "Label Height",
        ref: "cardHeight",
      },
      {
        label: "Label Width",
        ref: "cardWidth",
      },
      {
        label: "Page Width",
        ref: "pageWidth",
      },
      {
        label: "Page Height",
        ref: "pageHeight",
      },
      {
        label: "Page Top Padding",
        ref: "pageTopPadding",
      },
      {
        label: "Page Bottom Padding",
        ref: "pageBottomPadding",
      },
      {
        label: "Page Left Padding",
        ref: "pageLeftPadding",
      },
      {
        label: "Page Right Padding",
        ref: "pageRightPadding",
      },
      {
        label: "Base URL",
        ref: "baseURL",
        type: "text",
      },
    ];
  });

  type LabelData = {
    url: string;
    name: string;
    assetID: string;
    location: string;
  };

  function fmtAssetID(aid: number | string) {
    aid = aid.toString();

    let aidStr = aid.toString().padStart(6, "0");
    aidStr = aidStr.slice(0, 3) + "-" + aidStr.slice(3);
    return aidStr;
  }

  function getQRCodeUrl(assetID: string): string {
    let origin = displayProperties.baseURL.trim();

    // remove trailing slash
    if (origin.endsWith("/")) {
      origin = origin.slice(0, -1);
    }

    const data = `${origin}/a/${assetID}`;

    return route(`/qrcode`, { data: encodeURIComponent(data) });
  }

  function getItem(n: number, item: { value: { name: string; location: string } } | null): LabelData {
    // format n into - seperated string with leading zeros

    const assetID = fmtAssetID(n);

    return {
      url: getQRCodeUrl(assetID),
      assetID,
      name: item?.name ?? "_______________",
      location: item?.location?.name ?? "_______________",
    };
  }

  const { data: allFields } = await useAsyncData(async () => {
    const { data, error } = await api.items.getAll();

    if (error) {
      return [];
    }

    return data;
  });

  const items = computed(() => {
    if (displayProperties.assetRange > displayProperties.assetRangeMax) {
      return [];
    }

    const diff = displayProperties.assetRangeMax - displayProperties.assetRange;

    if (diff > 999) {
      return [];
    }

    const items: LabelData[] = [];
    for (let i = displayProperties.assetRange; i < displayProperties.assetRangeMax; i++) {
      items.push(getItem(i, allFields?.value?.items?.[i] ?? null));
    }
    return items;
  });

  type Row = {
    items: LabelData[];
  };

  type Page = {
    rows: Row[];
  };

  const pages = ref<Page[]>([]);

  const out = ref({
    cols: 0,
    rows: 0,
    gapY: 0,
    gapX: 0,
    card: {
      width: 0,
      height: 0,
    },
    page: {
      width: 0,
      height: 0,
      pt: 0,
      pb: 0,
      pl: 0,
      pr: 0,
    },
  });

  function calcPages() {
    // Set Out Dimensions
    out.value = calculateGridData({
      page: {
        height: displayProperties.pageHeight,
        width: displayProperties.pageWidth,
        pageTopPadding: displayProperties.pageTopPadding,
        pageBottomPadding: displayProperties.pageBottomPadding,
        pageLeftPadding: displayProperties.pageLeftPadding,
        pageRightPadding: displayProperties.pageRightPadding,
      },
      cardHeight: displayProperties.cardHeight,
      cardWidth: displayProperties.cardWidth,
    });

    const calc: Page[] = [];

    const perPage = out.value.rows * out.value.cols;

    const itemsCopy = [...items.value];

    while (itemsCopy.length > 0) {
      const page: Page = {
        rows: [],
      };

      for (let i = 0; i < perPage; i++) {
        const item = itemsCopy.shift();
        if (!item) {
          break;
        }

        if (i % out.value.cols === 0) {
          page.rows.push({
            items: [],
          });
        }

        page.rows[page.rows.length - 1].items.push(item);
      }

      calc.push(page);
    }

    pages.value = calc;
  }

  onMounted(() => {
    calcPages();
  });
</script>

<template>
  <div class="print:hidden">
    <AppToast />
    <div class="container max-w-4xl mx-auto p-4 pt-6 prose">
      <h1>Homebox Label Generator</h1>
      <p>
        The Homebox Label Generator is a tool to help you print labels for your Homebox inventory. These are intended to
        be print-ahead labels so you can print many labels and have them ready to apply
      </p>
      <p>
        As such, these labels work by printing a URL QR Code and AssetID information on a label. If you've disabled
        AssetID's in your Homebox settings, you can still use this tool, but the AssetID's won't reference any item
      </p>
      <p>
        This feature is in early development stages and may change in future releases, if you have feedback please
        provide it in the <a href="https://github.com/sysadminsmedia/homebox/discussions/273">GitHub Discussion</a>
      </p>
      <h2>Tips</h2>
      <ul>
        <li>
          The defaults here are setup for the
          <a href="https://www.avery.com/templates/5260">Avery 5260 label sheets</a>. If you're using a different sheet,
          you'll need to adjust the settings to match your sheet.
        </li>
        <li>
          If you're customizing your sheet the dimensions are in inches. When building the 5260 sheet, I found that the
          dimensions used in their template, did not match what was needed to print within the boxes.
          <b>Be prepared for some trial and error</b>
        </li>
        <li>
          When printing be sure to:
          <ol>
            <li>Set the margins to 0 or None</li>
            <li>Set the scaling to 100%</li>
            <li>Disable double-sided printing</li>
            <li>Print a test page before printing multiple pages</li>
          </ol>
        </li>
      </ul>
      <div class="flex gap-2 flex-wrap">
        <NuxtLink href="/tools">Tools</NuxtLink>
        <NuxtLink href="/home">Home</NuxtLink>
      </div>
    </div>
    <div class="divider max-w-4xl mx-auto"></div>
    <div class="container max-w-4xl mx-auto p-4">
      <div class="grid grid-cols-2 mx-auto gap-3">
        <div v-for="(prop, i) in propertyInputs" :key="i" class="form-control w-full max-w-xs">
          <label class="label">
            <span class="label-text">{{ prop.label }}</span>
          </label>
          <input
            v-model="displayProperties[prop.ref]"
            :type="prop.type ? prop.type : 'number'"
            step="0.01"
            placeholder="Type here"
            class="input input-bordered w-full max-w-xs"
          />
        </div>
      </div>
      <div class="max-w-xs">
        <div class="form-control">
          <label class="cursor-pointer label">
            <input v-model="bordered" type="checkbox" class="checkbox checkbox-secondary" />
            <span class="label-text">Bordered Labels</span>
          </label>
        </div>
      </div>

      <div>
        <p>QR Code Example {{ displayProperties.baseURL }}/a/{asset_id}</p>
        <BaseButton class="btn-block my-4" @click="calcPages"> Generate Page </BaseButton>
      </div>
    </div>
  </div>
  <div class="flex flex-col items-center print-show">
    <section
      v-for="(page, pi) in pages"
      :key="pi"
      class="border-2 print:border-none"
      :style="{
        paddingTop: `${out.page.pt}in`,
        paddingBottom: `${out.page.pb}in`,
        paddingLeft: `${out.page.pl}in`,
        paddingRight: `${out.page.pr}in`,
        width: `${out.page.width}in`,
      }"
    >
      <div
        v-for="(row, ri) in page.rows"
        :key="ri"
        class="flex break-inside-avoid"
        :style="{
          columnGap: `${out.gapX}in`,
          rowGap: `${out.gapY}in`,
        }"
      >
        <div
          v-for="(item, idx) in row.items"
          :key="idx"
          class="flex border-2"
          :class="{
            'border-black': bordered,
            'border-transparent': !bordered,
          }"
          :style="{
            height: `${out.card.height}in`,
            width: `${out.card.width}in`,
          }"
        >
          <div class="flex items-center">
            <img
              :src="item.url"
              :style="{
                width: `${out.card.height * 0.9}in`,
                height: `${out.card.height * 0.9}in`,
              }"
            />
          </div>
          <div class="ml-2 flex flex-col justify-center">
            <div class="font-bold">{{ item.assetID }}</div>
            <div class="text-xs font-light italic">Homebox</div>
            <div>{{ item.name }}</div>
            <div>{{ item.location }}</div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<style lang="css">
  .letter-size {
    width: 8.5in;
    height: 11in;
    padding: 0.5in;
  }
</style>
