<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import DOMPurify from "dompurify";
  import { route } from "../../lib/api/base";
  import { Toaster, toast } from "@/components/ui/sonner";
  import { Separator } from "@/components/ui/separator";
  import { Button } from "@/components/ui/button";
  import { Label } from "@/components/ui/label";
  import { Input } from "@/components/ui/input";
  import { Checkbox } from "@/components/ui/checkbox";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

  const { t } = useI18n();

  definePageMeta({
    middleware: ["auth"],
    layout: false,
  });
  useHead({
    title: "HomeBox | " + t("reports.label_generator.title"),
  });

  const api = useUserApi();

  const bordered = ref(false);
  const printLocationRow = ref(true);
  const labelBlankLine = "_______________";

  // Behavior constants for HomeBox text replacement
  const BEHAVIOR_SHOW = "show";
  const BEHAVIOR_ALWAYS_REPLACE = "always_replace";
  const BEHAVIOR_ITEM_NO_NAME_NO_LOCATION = "item_no_name_no_location";
  const BEHAVIOR_ITEM_NO_NAME = "item_no_name";
  const BEHAVIOR_ITEM_NO_LOCATION = "item_no_location";

  const replaceHomeboxBehavior = ref(BEHAVIOR_SHOW);
  const replaceHomeboxText = ref(labelBlankLine);

  const displayProperties = reactive({
    baseURL: window.location.origin,
    assetRange: 1,
    assetRangeMax: 91,
    skipLabels: 0,
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
      toast.error(t("reports.label_generator.toast.page_too_small_card"));
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
    min?: number;
    step?: number;
  }

  const propertyInputs = computed<InputDef[]>(() => {
    return [
      {
        label: t("reports.label_generator.asset_start"),
        ref: "assetRange",
      },
      {
        label: t("reports.label_generator.asset_end"),
        ref: "assetRangeMax",
      },
      {
        label: t("reports.label_generator.skip_first_labels"),
        ref: "skipLabels",
        min: 0,
        step: 1,
      },
      {
        label: t("reports.label_generator.measure_type"),
        ref: "measure",
        type: "text",
      },
      {
        label: t("reports.label_generator.label_height"),
        ref: "cardHeight",
      },
      {
        label: t("reports.label_generator.label_width"),
        ref: "cardWidth",
      },
      {
        label: t("reports.label_generator.page_width"),
        ref: "pageWidth",
      },
      {
        label: t("reports.label_generator.page_height"),
        ref: "pageHeight",
      },
      {
        label: t("reports.label_generator.page_top_padding"),
        ref: "pageTopPadding",
      },
      {
        label: t("reports.label_generator.page_bottom_padding"),
        ref: "pageBottomPadding",
      },
      {
        label: t("reports.label_generator.page_left_padding"),
        ref: "pageLeftPadding",
      },
      {
        label: t("reports.label_generator.page_right_padding"),
        ref: "pageRightPadding",
      },
      {
        label: t("reports.label_generator.base_url"),
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
      name: item?.name ?? labelBlankLine,
      location: item?.location?.name ?? labelBlankLine,
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

  const getHomeBoxLineText = computed(() => {
    return (item: LabelData): string | null => {
      if (replaceHomeboxBehavior.value === BEHAVIOR_SHOW) {
        return "HomeBox";
      }
      if (replaceHomeboxBehavior.value === BEHAVIOR_ALWAYS_REPLACE) {
        return replaceHomeboxText.value;
      }
      if (
        replaceHomeboxBehavior.value === BEHAVIOR_ITEM_NO_NAME_NO_LOCATION &&
        item.name === labelBlankLine &&
        item.location === labelBlankLine
      ) {
        return replaceHomeboxText.value;
      }
      if (replaceHomeboxBehavior.value === BEHAVIOR_ITEM_NO_NAME && item.name === labelBlankLine) {
        return replaceHomeboxText.value;
      }
      if (replaceHomeboxBehavior.value === BEHAVIOR_ITEM_NO_LOCATION && item.location === labelBlankLine) {
        return replaceHomeboxText.value;
      }
      return null;
    };
  });

  type Row = {
    items: Array<LabelData | null>;
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
    const maxSkipLabels = Math.max(0, perPage - 1);

    const skipLabelsRaw = Number(displayProperties.skipLabels);
    const skipLabels = Number.isFinite(skipLabelsRaw)
      ? Math.min(maxSkipLabels, Math.max(0, Math.floor(skipLabelsRaw)))
      : 0;
    if (Number(displayProperties.skipLabels) !== skipLabels) {
      displayProperties.skipLabels = skipLabels;
    }

    const sourceItems = items.value;
    if (sourceItems.length === 0) {
      pages.value = [];
      return;
    }

    const itemsCopy: Array<LabelData | null> = [...sourceItems];
    if (skipLabels > 0) {
      itemsCopy.unshift(...Array.from({ length: skipLabels }, () => null));
    }

    while (itemsCopy.length > 0) {
      const page: Page = {
        rows: [],
      };

      for (let i = 0; i < perPage; i++) {
        const item = itemsCopy.shift();
        if (typeof item === "undefined") {
          break;
        }

        if (i % out.value.cols === 0) {
          page.rows.push({
            items: [],
          });
        }

        page.rows[page.rows.length - 1]!.items.push(item ?? null);
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
      <h1>HomeBox {{ $t("reports.label_generator.title") }}</h1>
      <p>
        {{ $t("reports.label_generator.instruction_1") }}
      </p>
      <p>
        {{ $t("reports.label_generator.instruction_2") }}
      </p>
      <p v-html="DOMPurify.sanitize($t('reports.label_generator.instruction_3'))" />
      <h2>{{ $t("reports.label_generator.tips") }}</h2>
      <ul>
        <li v-html="DOMPurify.sanitize($t('reports.label_generator.tip_1'))" />
        <li v-html="DOMPurify.sanitize($t('reports.label_generator.tip_2'))" />
        <li v-html="DOMPurify.sanitize($t('reports.label_generator.tip_3'))" />
      </ul>
      <div class="flex flex-wrap gap-2">
        <NuxtLink href="/collection/tools">{{ $t("collection.tabs.tools") }}</NuxtLink>
        <NuxtLink href="/home">{{ $t("menu.home") }}</NuxtLink>
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
            :min="prop.min"
            :max="prop.ref === 'skipLabels' ? Math.max(0, out.rows * out.cols - 1) : undefined"
            :step="prop.type === 'text' ? undefined : (prop.step ?? 0.01)"
            :placeholder="$t('reports.label_generator.input_placeholder')"
            class="w-full max-w-xs"
          />
        </div>
        <div class="flex w-full max-w-xs flex-col">
          <Label for="select-replaceHomeboxBehavior">
            {{ $t("reports.label_generator.replace_homebox_behavior") }}
          </Label>
          <Select id="select-replaceHomeboxBehavior" v-model="replaceHomeboxBehavior" class="w-full max-w-xs">
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem :value="BEHAVIOR_SHOW">
                {{ $t("reports.label_generator.replace_homebox_behavior_show_homebox") }}
              </SelectItem>
              <SelectItem :value="BEHAVIOR_ITEM_NO_NAME_NO_LOCATION">
                {{ $t("reports.label_generator.replace_homebox_behavior_item_no_name_no_location") }}
              </SelectItem>
              <SelectItem :value="BEHAVIOR_ITEM_NO_NAME">
                {{ $t("reports.label_generator.replace_homebox_behavior_item_no_name") }}
              </SelectItem>
              <SelectItem :value="BEHAVIOR_ITEM_NO_LOCATION">
                {{ $t("reports.label_generator.replace_homebox_behavior_item_no_location") }}
              </SelectItem>
              <SelectItem :value="BEHAVIOR_ALWAYS_REPLACE">
                {{ $t("reports.label_generator.replace_homebox_behavior_always_replace") }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div v-if="replaceHomeboxBehavior !== BEHAVIOR_SHOW" class="flex w-full max-w-xs flex-col">
          <Label for="input-replaceHomeboxText">
            {{ $t("reports.label_generator.replace_homebox_text") }}
          </Label>
          <Input
            id="input-replaceHomeboxText"
            v-model="replaceHomeboxText"
            type="text"
            :placeholder="$t('reports.label_generator.input_placeholder')"
            class="w-full max-w-xs"
          />
        </div>
      </div>
      <div class="max-w-xs">
        <div class="flex items-center gap-2 py-4">
          <Checkbox id="borderedLabels" v-model="bordered" />
          <Label class="cursor-pointer" for="borderedLabels">
            {{ $t("reports.label_generator.bordered_labels") }}
          </Label>
        </div>
        <div class="flex items-center gap-2 py-4">
          <Checkbox id="printLocationRow" v-model="printLocationRow" />
          <Label class="cursor-pointer" for="printLocationRow">
            {{ $t("reports.label_generator.print_location_row") }}
          </Label>
        </div>
      </div>

      <div>
        <p>{{ $t("reports.label_generator.qr_code_example") }} {{ displayProperties.baseURL }}/a/{asset_id}</p>
        <Button size="lg" class="my-4 w-full" @click="calcPages">
          {{ $t("reports.label_generator.generate_page") }}
        </Button>
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
            'border-black': bordered && !!item,
            'border-transparent': !bordered || !item,
          }"
          :style="{
            height: `${out.card.height}${out.measure}`,
            width: `${out.card.width}${out.measure}`,
          }"
        >
          <template v-if="item">
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
              <div
                v-if="getHomeBoxLineText(item)"
                class="text-xs"
                :class="{ 'font-light italic': getHomeBoxLineText(item) !== labelBlankLine }"
              >
                {{ getHomeBoxLineText(item) }}
              </div>
              <div class="overflow-hidden text-wrap text-xs">{{ item.name }}</div>
              <div v-if="printLocationRow" class="text-xs">{{ item.location }}</div>
            </div>
          </template>
        </div>
      </div>
    </section>
  </div>
</template>
