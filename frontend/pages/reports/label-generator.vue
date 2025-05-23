<script setup lang="ts">
  import { route } from "../../lib/api/base";
  import { toast, Toaster } from "@/components/ui/sonner";
  import { Separator } from "@/components/ui/separator";
  import { Button } from "@/components/ui/button";
  import { Label } from "@/components/ui/label";
  import { Input } from "@/components/ui/input";
  import { Checkbox } from "@/components/ui/checkbox";

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
    measure: "in",
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

  type LabelOptionInput = {
    measure: string;
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
    measure: string;
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

  function calculateGridData(input: LabelOptionInput): Output {
    const { page, cardHeight, cardWidth } = input;

    const measureRegex = /in|cm|mm/;
    const measure = measureRegex.test(input.measure) ? input.measure : "in";

    const availablePageWidth = page.width - page.pageLeftPadding - page.pageRightPadding;
    const availablePageHeight = page.height - page.pageTopPadding - page.pageBottomPadding;

    if (availablePageWidth < cardWidth || availablePageHeight < cardHeight) {
      toast.error("Page size is too small for the card size");
      return out.value;
    }

    const cols = Math.floor(availablePageWidth / cardWidth);
    const rows = Math.floor(availablePageHeight / cardHeight);
    const gapX = (availablePageWidth - cols * cardWidth) / (cols - 1);
    const gapY = (page.height - rows * cardHeight) / (rows - 1);

    return {
      measure,
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
        label: "Measure Type",
        ref: "measure",
        type: "text",
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

  function getItem(n: number, item: { assetId: string; name: string; location: { name: string } } | null): LabelData {
    // format n into - seperated string with leading zeros
    const assetID = fmtAssetID(item?.assetId ?? n + 1);

    return {
      url: getQRCodeUrl(assetID),
      assetID: item?.assetId ?? assetID,
      name: item?.name ?? "_______________",
      location: item?.location?.name ?? "_______________",
    };
  }

  const { data: allFields } = await useAsyncData(async () => {
    const { data, error } = await api.items.getAll({ orderBy: "assetId" });

    if (error) {
      return {
        items: [],
      };
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
    for (let i = displayProperties.assetRange - 1; i < displayProperties.assetRangeMax - 1; i++) {
      const item = allFields?.value?.items?.[i];
      if (item?.location) {
        items.push(getItem(i, item as { assetId: string; location: { name: string }; name: string }));
      } else {
        items.push(getItem(i, null));
      }
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
    measure: "in",
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
      measure: displayProperties.measure,
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
    <Toaster />
    <div class="container prose mx-auto max-w-4xl p-4 pt-6">
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
        provide it in the <a href="https://github.com/sysadminsmedia/homebox/discussions/53">GitHub Discussion</a>
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
      <div class="flex flex-wrap gap-2">
        <NuxtLink href="/tools">Tools</NuxtLink>
        <NuxtLink href="/home">Home</NuxtLink>
      </div>
    </div>
    <Separator class="mx-auto max-w-4xl" />
    <div class="container mx-auto max-w-4xl p-4">
      <div class="mx-auto grid grid-cols-2 gap-3">
        <div v-for="(prop, i) in propertyInputs" :key="i" class="flex w-full max-w-xs flex-col">
          <Label :for="`input-${prop.ref}`">
            {{ prop.label }}
          </Label>
          <Input
            :id="`input-${prop.ref}`"
            v-model="displayProperties[prop.ref]"
            :type="prop.type ? prop.type : 'number'"
            step="0.01"
            placeholder="Type here"
            class="w-full max-w-xs"
          />
        </div>
      </div>
      <div class="max-w-xs">
        <div class="flex items-center gap-2 py-4">
          <Checkbox id="borderedLabels" v-model="bordered" />
          <Label class="cursor-pointer" for="borderedLabels"> Bordered Labels </Label>
        </div>
      </div>

      <div>
        <p>QR Code Example {{ displayProperties.baseURL }}/a/{asset_id}</p>
        <Button size="lg" class="my-4 w-full" @click="calcPages"> Generate Page </Button>
      </div>
    </div>
  </div>
  <div class="flex flex-col items-center">
    <section
      v-for="(page, pi) in pages"
      :key="pi"
      class="border-2 print:border-none"
      :style="{
        paddingTop: `${out.page.pt}${out.measure}`,
        paddingBottom: `${out.page.pb}${out.measure}`,
        paddingLeft: `${out.page.pl}${out.measure}`,
        paddingRight: `${out.page.pr}${out.measure}`,
        width: `${out.page.width}${out.measure}`,
        background: `white`,
        color: `black`,
      }"
    >
      <div
        v-for="(row, ri) in page.rows"
        :key="ri"
        class="flex break-inside-avoid"
        :style="{
          columnGap: `${out.gapX}${out.measure}`,
          rowGap: `${out.gapY}${out.measure}`,
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
            height: `${out.card.height}${out.measure}`,
            width: `${out.card.width}${out.measure}`,
          }"
        >
          <div class="flex items-center">
            <img
              :src="item.url"
              :style="{
                minWidth: `${out.card.height * 0.9}${out.measure}`,
                width: `${out.card.height * 0.9}${out.measure}`,
                height: `${out.card.height * 0.9}${out.measure}`,
              }"
            />
          </div>
          <div class="ml-2 flex flex-col justify-center">
            <div class="font-bold">{{ item.assetID }}</div>
            <div class="text-xs font-light italic">Homebox</div>
            <div class="overflow-hidden text-wrap text-xs">{{ item.name }}</div>
            <div class="text-xs">{{ item.location }}</div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>
