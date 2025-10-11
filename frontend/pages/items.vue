<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import { Input } from "~/components/ui/input";
  import type { ItemSummary, LabelSummary, LocationOutCount } from "~~/lib/api/types/data-contracts";
  import { useLabelStore } from "~~/stores/labels";
  import { useLocationStore } from "~~/stores/locations";
  import MdiLoading from "~icons/mdi/loading";
  import MdiMagnify from "~icons/mdi/magnify";
  import MdiDelete from "~icons/mdi/delete";
  import { Button } from "@/components/ui/button";
  import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
  import { Label } from "@/components/ui/label";
  import { Switch } from "@/components/ui/switch";
  import { Separator } from "@/components/ui/separator";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import BaseContainer from "@/components/Base/Container.vue";
  import SearchFilter from "~/components/Search/Filter.vue";
  import ItemViewSelectable from "~/components/Item/View/Selectable.vue";
  import type { LocationQueryRaw } from "vue-router";

  const { t } = useI18n();

  definePageMeta({
    middleware: ["auth"],
  });

  useHead({
    title: "HomeBox | " + t("global.items"),
  });

  const searchLocked = ref(false);
  const queryParamsInitialized = ref(false);
  const initialSearch = ref(true);

  const api = useUserApi();
  const loading = useMinLoader(500);
  const items = ref<ItemSummary[]>([]);
  const total = ref(0);

  type queryParamValue = string | string[] | number | boolean;
  type queryRef =
    | WritableComputedRef<string>
    | WritableComputedRef<string[]>
    | WritableComputedRef<number>
    | WritableComputedRef<boolean>;
  const queryParamDefaultValues: Record<string, queryParamValue> = {};
  function useOptionalRouteQuery(key: string, defaultValue: string): WritableComputedRef<string>;
  function useOptionalRouteQuery(key: string, defaultValue: string[]): WritableComputedRef<string[]>;
  function useOptionalRouteQuery(key: string, defaultValue: number): WritableComputedRef<number>;
  function useOptionalRouteQuery(key: string, defaultValue: boolean): WritableComputedRef<boolean>;
  function useOptionalRouteQuery(key: string, defaultValue: queryParamValue): queryRef {
    queryParamDefaultValues[key] = defaultValue;
    if (typeof defaultValue === "string") {
      return useRouteQuery(key, defaultValue);
    }
    if (Array.isArray(defaultValue)) {
      return useRouteQuery(key, defaultValue as string[]);
    }
    if (typeof defaultValue === "number") {
      return useRouteQuery(key, defaultValue);
    }
    if (typeof defaultValue === "boolean") {
      return useRouteQuery(key, defaultValue);
    }

    throw Error(`Invalid query value type ${typeof defaultValue}`);
  }

  const page1 = useOptionalRouteQuery("page", 1);

  const page = computed({
    get: () => page1.value,
    set: value => {
      page1.value = value;
    },
  });

  const query = useOptionalRouteQuery("q", "");
  const includeArchived = useOptionalRouteQuery("archived", false);
  const fieldSelector = useOptionalRouteQuery("fieldSelector", false);
  const negateLabels = useOptionalRouteQuery("negateLabels", false);
  const onlyWithoutPhoto = useOptionalRouteQuery("onlyWithoutPhoto", false);
  const onlyWithPhoto = useOptionalRouteQuery("onlyWithPhoto", false);
  const orderBy = useOptionalRouteQuery("orderBy", "name");
  const qLoc = useOptionalRouteQuery("loc", []);
  const qLab = useOptionalRouteQuery("lab", []);

  const preferences = useViewPreferences();
  const pageSize = computed(() => preferences.value.itemsPerTablePage);

  const route = useRoute();
  const router = useRouter();

  onMounted(async () => {
    loading.value = true;
    searchLocked.value = true;
    await Promise.all([locationsStore.ensureLocationsFetched(), labelStore.ensureAllLabelsFetched()]);
    if (qLoc) {
      selectedLocations.value = locations.value.filter(l => qLoc.value.includes(l.id));
    }

    if (qLab) {
      selectedLabels.value = labels.value.filter(l => qLab.value.includes(l.id));
    }

    queryParamsInitialized.value = true;
    searchLocked.value = false;

    const qFields = route.query.fields as string[];
    if (qFields) {
      fieldTuples.value = qFields.map(f => f.split("=") as [string, string]);

      for (const t of fieldTuples.value) {
        if (t[0] && t[1]) {
          await fetchValues(t[0]);
        }
      }
    }

    // trigger search if no changes
    if (!qLab && !qLoc) {
      search();
    }

    loading.value = false;
    window.scroll({
      top: 0,
      left: 0,
      behavior: "smooth",
    });
  });

  const locationsStore = useLocationStore();

  const locationFlatTree = useFlatLocations();

  const locations = computed(() => locationsStore.allLocations);

  const labelStore = useLabelStore();
  const labels = computed(() => labelStore.labels);

  const selectedLocations = ref<LocationOutCount[]>([]);
  const selectedLabels = ref<LabelSummary[]>([]);

  const locIDs = computed(() => selectedLocations.value.map(l => l.id));
  const labIDs = computed(() => selectedLabels.value.map(l => l.id));

  function parseAssetIDString(d: string) {
    d = d.replace(/"/g, "").replace(/-/g, "");

    const aidInt = parseInt(d);
    if (isNaN(aidInt)) {
      return [-1, false];
    }

    return [aidInt, true];
  }

  const byAssetId = computed(() => query.value?.startsWith("#") || false);
  const parsedAssetId = computed(() => {
    if (!byAssetId.value) {
      return "";
    } else {
      const [aid, valid] = parseAssetIDString(query.value.replace("#", ""));
      if (!valid) {
        return t("items.invalid_asset_id");
      } else {
        return aid;
      }
    }
  });

  const fieldTuples = ref<[string, string][]>([]);
  const fieldValuesCache = ref<Record<string, string[]>>({});

  const { data: allFields } = useAsyncData(async () => {
    const { data, error } = await api.items.fields.getAll();

    if (error) {
      return [];
    }

    return data;
  });

  watch(includeArchived, (newV, oldV) => {
    if (newV !== oldV) {
      search();
    }
  });

  watch(fieldSelector, (newV, oldV) => {
    if (newV === false && oldV === true) {
      fieldTuples.value = [];
    }
  });

  watch(negateLabels, (newV, oldV) => {
    if (newV !== oldV) {
      search();
    }
  });

  watch(onlyWithoutPhoto, (newV, oldV) => {
    if (newV && onlyWithPhoto) {
      onlyWithPhoto.value = false;
    }
    if (newV !== oldV) {
      search();
    }
  });

  watch(onlyWithPhoto, (newV, oldV) => {
    if (newV && onlyWithoutPhoto) {
      onlyWithoutPhoto.value = false;
    }
    if (newV !== oldV) {
      search();
    }
  });

  watch(orderBy, (newV, oldV) => {
    if (newV !== oldV) {
      search();
    }
  });

  async function fetchValues(field: string): Promise<string[]> {
    if (fieldValuesCache.value[field]) {
      return fieldValuesCache.value[field];
    }

    const { data, error } = await api.items.fields.getAllValues(field);

    if (error) {
      return [];
    }

    fieldValuesCache.value[field] = data;

    return data;
  }

  async function search() {
    if (searchLocked.value) {
      return;
    }

    loading.value = true;

    const fields = [];

    for (const t of fieldTuples.value) {
      if (t[0] && t[1]) {
        fields.push(`${t[0]}=${t[1]}`);
      }
    }

    const push_query: Record<string, string | string[] | number | boolean | undefined> = {
      archived: includeArchived.value,
      fieldSelector: fieldSelector.value,
      negateLabels: negateLabels.value,
      onlyWithoutPhoto: onlyWithoutPhoto.value,
      onlyWithPhoto: onlyWithPhoto.value,
      orderBy: orderBy.value,
      page: page.value,
      q: query.value,
      loc: locIDs.value,
      lab: labIDs.value,
      fields: fields,
    };

    for (const key in push_query) {
      const val = push_query[key];
      const defaultVal = queryParamDefaultValues[key];
      if (
        (Array.isArray(val) &&
          Array.isArray(defaultVal) &&
          val.length == defaultVal.length &&
          val.every(v => (defaultVal as string[]).includes(v))) ||
        val === queryParamDefaultValues[key]
      ) {
        push_query[key] = undefined;
      }

      // Empirically seen to be unnecessary but according to router.push types,
      // booleans are not supported. This might be more stable.
      if (typeof push_query[key] === "boolean") {
        push_query[key] = String(val);
      }
    }

    await router.push({ query: push_query as LocationQueryRaw });

    const { data, error } = await api.items.getAll({
      q: query.value || "",
      locations: locIDs.value,
      labels: labIDs.value,
      negateLabels: negateLabels.value,
      onlyWithoutPhoto: onlyWithoutPhoto.value,
      onlyWithPhoto: onlyWithPhoto.value,
      includeArchived: includeArchived.value,
      page: page.value,
      pageSize: pageSize.value,
      orderBy: orderBy.value,
      fields,
    });

    function resetItems() {
      page.value = Math.max(1, page.value - 1);
      loading.value = false;
      total.value = 0;
      items.value = [];
    }

    if (error) {
      resetItems();
      toast.error(t("items.toast.failed_search_items"));
      return;
    }

    if (!data.items || data.items.length === 0) {
      resetItems();
      return;
    }

    total.value = data.total;
    items.value = data.items;

    loading.value = false;
    initialSearch.value = false;
  }

  watchDebounced([page, pageSize, query, selectedLabels, selectedLocations], search, { debounce: 250, maxWait: 1000 });

  async function submit() {
    // Set URL Params
    const fields = [];
    for (const t of fieldTuples.value) {
      if (t[0] && t[1]) {
        fields.push(`${t[0]}=${t[1]}`);
      }
    }

    // Reset Pagination
    page.value = 1;

    // Perform Search
    await search();
  }

  async function reset() {
    // Set URL Params
    const fields = [];
    for (const t of fieldTuples.value) {
      if (t[0] && t[1]) {
        fields.push(`${t[0]}=${t[1]}`);
      }
    }

    await search();
  }

  const pagination = proxyRefs({
    page,
    pageSize,
    totalSize: total,
    setPage: (newPage: number) => {
      page.value = newPage;
    },
  });
</script>

<template>
  <BaseContainer>
    <div v-if="locations && labels">
      <div class="flex flex-wrap items-end gap-4 md:flex-nowrap">
        <div class="w-full">
          <Input v-model:model-value="query" :placeholder="$t('global.search')" class="h-12" />
          <div v-if="byAssetId" class="pl-2 pt-2 text-sm">
            <p>{{ $t("items.query_id", { id: parsedAssetId }) }}</p>
          </div>
        </div>
        <Button class="mb-auto h-12 w-full md:w-auto" @click.prevent="submit">
          <MdiLoading v-if="loading" class="animate-spin" />
          <MdiMagnify v-else />
          {{ $t("global.search") }}
        </Button>
      </div>

      <div class="flex w-full flex-wrap gap-2 py-2 md:flex-nowrap">
        <SearchFilter v-model="selectedLocations" :label="$t('global.locations')" :options="locationFlatTree" />
        <SearchFilter v-model="selectedLabels" :label="$t('global.labels')" :options="labels" />
        <Popover>
          <PopoverTrigger as-child>
            <Button size="sm" variant="outline"> {{ $t("items.options") }}</Button>
          </PopoverTrigger>
          <PopoverContent class="z-40 flex flex-col gap-2">
            <Label class="flex cursor-pointer items-center">
              <Switch v-model="includeArchived" class="ml-auto" />
              <div class="grow" />
              {{ $t("items.include_archive") }}
            </Label>
            <Label class="flex cursor-pointer items-center">
              <Switch v-model="fieldSelector" class="ml-auto" />
              <div class="grow" />
              {{ $t("items.field_selector") }}
            </Label>
            <Label class="flex cursor-pointer items-center">
              <Switch v-model="negateLabels" class="ml-auto" />
              <div class="grow" />
              {{ $t("items.negate_labels") }}
            </Label>
            <Label class="flex cursor-pointer items-center">
              <Switch v-model="onlyWithoutPhoto" class="ml-auto" />
              <div class="grow" />
              {{ $t("items.only_without_photo") }}
            </Label>
            <Label class="flex cursor-pointer items-center">
              <Switch v-model="onlyWithPhoto" class="ml-auto" />
              <div class="grow" />
              {{ $t("items.only_with_photo") }}
            </Label>
            <Label class="flex cursor-pointer flex-col gap-2">
              <span class="text-right">
                {{ $t("items.order_by") }}
              </span>

              <Select v-model="orderBy">
                <SelectTrigger>
                  <SelectValue :placeholder="$t('items.order_by')" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="createdAt"> {{ $t("items.created_at") }} </SelectItem>
                  <SelectItem value="updatedAt"> {{ $t("items.updated_at") }} </SelectItem>
                </SelectContent>
              </Select>
            </Label>
            <Separator />
            <Button @click="reset"> {{ $t("items.reset_search") }} </Button>
          </PopoverContent>
        </Popover>
        <div class="grow" />
        <Popover>
          <PopoverTrigger as-child>
            <Button size="sm" variant="outline"> {{ $t("items.tips") }}</Button>
          </PopoverTrigger>
          <PopoverContent class="z-40 w-[325px]" align="end">
            <p class="text-base">{{ $t("items.tips_sub") }}</p>
            <ul class="mt-1 list-disc pl-6 text-sm">
              <li>
                {{ $t("items.tip_1") }}
              </li>
              <li>
                {{ $t("items.tip_2") }}
              </li>
              <li>
                {{ $t("items.tip_3") }}
              </li>
            </ul>
          </PopoverContent>
        </Popover>
      </div>
      <div v-if="fieldSelector" class="flex flex-col gap-2 pb-2">
        <p>{{ $t("items.custom_fields") }}</p>
        <div v-for="(f, idx) in fieldTuples" :key="idx" class="flex flex-wrap gap-2">
          <div class="flex w-full flex-col gap-1 md:w-auto md:grow">
            <Label> Field </Label>
            <Select v-model="fieldTuples[idx]![0]" @update:model-value="fetchValues(f[0])">
              <SelectTrigger>
                <SelectValue :placeholder="$t('items.select_field')" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="field in allFields" :key="field" :value="field"> {{ field }} </SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div class="flex w-full flex-col gap-1 md:w-auto md:grow">
            <Label> {{ $t("items.field_value") }} </Label>
            <Select v-model="fieldTuples[idx]![1]">
              <SelectTrigger>
                <SelectValue placeholder="Select a value" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="value in fieldValuesCache[f[0]]" :key="value" :value="value">
                  {{ value }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
          <Button variant="destructive" type="button" size="icon" class="my-auto" @click="fieldTuples.splice(idx, 1)">
            <MdiDelete />
          </Button>
        </div>
        <Button type="button" size="sm" class="mt-2" @click="() => fieldTuples.push(['', ''])">
          {{ $t("items.add") }}
        </Button>
      </div>
    </div>

    <section>
      <ItemViewSelectable
        :items="items"
        :location-flat-tree="locationFlatTree"
        :pagination="pagination"
        @refresh="async () => search()"
      />
    </section>
  </BaseContainer>
</template>
