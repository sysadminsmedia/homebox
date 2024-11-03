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
  const gapY = (availablePageHeight - rows * cardHeight) / (rows - 1);

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
    { label: "Asset Start", ref: "assetRange" },
    { label: "Asset End", ref: "assetRangeMax" },
    { label: "Label Height", ref: "cardHeight" },
    { label: "Label Width", ref: "cardWidth" },
    { label: "Page Width", ref: "pageWidth" },
    { label: "Page Height", ref: "pageHeight" },
    { label: "Page Top Padding", ref: "pageTopPadding" },
    { label: "Page Bottom Padding", ref: "pageBottomPadding" },
    { label: "Page Left Padding", ref: "pageLeftPadding" },
    { label: "Page Right Padding", ref: "pageRightPadding" },
    { label: "Base URL", ref: "baseURL", type: "text" },
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
  let aidStr = aid.padStart(6, "0");
  aidStr = aidStr.slice(0, 3) + "-" + aidStr.slice(3);
  return aidStr;
}

function getQRCodeUrl(assetID: string): string {
  let origin = displayProperties.baseURL.trim();
  if (origin.endsWith("/")) {
    origin = origin.slice(0, -1);
  }
  const data = `${origin}/a/${assetID}`;
  return route(`/qrcode`, { data: encodeURIComponent(data) });
}

function getItem(n: number, item: { name: string; location: { name: string } } | null): LabelData {
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
    return { items: [] };
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
    const item = allFields?.value?.items?.[i];
    items.push(getItem(i, item?.location ? item : null));
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

const out = ref<Output>({
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
    const page: Page = { rows: [] };
    for (let i = 0; i < perPage; i++) {
      const item = itemsCopy.shift();
      if (!item) break;

      if (i % out.value.cols === 0) {
        page.rows.push({ items: [] });
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
    <div class="container prose mx-auto max-w-4xl p-4 pt-6">
      <h1>Homebox Label Generator</h1>
      <p>
        The Homebox Label Generator is a tool to help you print labels for your Homebox inventory. These are intended to
        be print-ahead labels so you can print many labels and have them ready to apply.
      </p>
      <p>
        As such, these labels work by printing a URL QR Code and AssetID information on a label. If you've disabled
        AssetID's in your Homebox settings, you can still use this tool, but the AssetID's won't reference any item.
      </p>
      <p>
        This feature is in early development stages and may change in future releases; if you have feedback please
        provide it in the <a href="https://github.com/sysadminsmedia/homebox/discussions/53">GitHub Discussion</a>.
      </p>
      <h2>Tips</h2>
      <ul>
        <li>
          The defaults here are set up for the
          <a href="https://www.avery.com/templates/5260">Avery 5260 label sheets</a>. If you're using a different sheet,
          you'll need to adjust the settings to match your sheet.
        </li>
        <li>
          If you're customizing your sheet, the dimensions are in inches. When building the 5260 sheet, I found that the
          dimensions used in their template did not match what was needed to print within the boxes.
          <b>Be prepared for some trial and error.</b>
        </li>
        <li>
          When printing, be sure to:
          <ol>
            <li>Set the margins to 0 or None.</li>
            <li>Set the scaling to 100%.</li>
            <li>Disable double-sided printing.</li>
            <li>Use the 'Actual Size' option.</li>
            <li>Use a laser printer for the best results.</li>
          </ol>
        </li>
      </ul>
      <h2>Settings</h2>
      <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div v-for="input in propertyInputs" :key="input.ref" class="form-group">
          <label :for="input.ref">{{ input.label }}</label>
          <component
            :is="input.type === 'number' ? 'input' : 'input'"
            class="form-control"
            v-model="displayProperties[input.ref]"
            :type="input.type ?? 'text'"
            :min="input.type === 'number' ? 0 : undefined"
            :max="input.type === 'number' ? 1000 : undefined"
            :step="input.type === 'number' ? 1 : undefined"
          />
        </div>
      </div>
    </div>
    <div class="container print:hidden mx-auto max-w-4xl px-4 pt-6">
      <h2 class="mt-10 mb-5">Preview</h2>
      <div class="print:hidden">
        <div
          v-for="(page, pageIndex) in pages"
          :key="pageIndex"
          class="print:block border-2 border-gray-400 p-2 mx-auto my-2"
          :style="{ width: `${out.page.width}in`, height: `${out.page.height}in`, display: 'flex', flexDirection: 'column', padding: `${out.page.pt}in ${out.page.pr}in ${out.page.pb}in ${out.page.pl}in` }"
        >
          <div
            v-for="(row, rowIndex) in page.rows"
            :key="rowIndex"
            class="flex"
            :style="{ height: `${out.card.height}in` }"
          >
            <div
              v-for="(item, itemIndex) in row.items"
              :key="itemIndex"
              :style="{ height: `${out.card.height}in`, width: `${out.card.width}in`, paddingRight: `${out.gapX}in`, marginBottom: `${out.gapY}in` }"
              class="flex-none border-2 border-gray-400 flex flex-col justify-center items-center text-center"
            >
              <img :src="item.url" class="w-32 h-32" />
              <div class="mt-2">
                <p class="font-bold text-sm">{{ item.assetID }}</p>
                <p class="text-xs">{{ item.name }}</p>
                <p class="text-xs">{{ item.location }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="container mx-auto max-w-4xl px-4 pt-6">
      <button @click="calcPages" class="btn btn-primary mt-5">
        Generate Labels
      </button>
    </div>
  </div>
</template>
