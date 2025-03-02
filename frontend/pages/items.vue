<script setup lang="ts">
  import type { ItemSummary, LabelSummary, LocationOutCount } from "~~/lib/api/types/data-contracts";
  import { useLabelStore } from "~~/stores/labels";
  import { useLocationStore } from "~~/stores/locations";
  import MdiLoading from "~icons/mdi/loading";
  import MdiSelectSearch from "~icons/mdi/select-search";
  import MdiMagnify from "~icons/mdi/magnify";
  import MdiDelete from "~icons/mdi/delete";
  import MdiChevronRight from "~icons/mdi/chevron-right";
  import MdiChevronLeft from "~icons/mdi/chevron-left";

  definePageMeta({
    middleware: ["auth"],
  });

  useHead({
    title: "Homebox | Items",
  });

  const searchLocked = ref(false);
  const queryParamsInitialized = ref(false);
  const initialSearch = ref(true);

  const api = useUserApi();
  const loading = useMinLoader(500);
  const items = ref<ItemSummary[]>([]);
  const total = ref(0);

  const page1 = useRouteQuery("page", 1);

  const page = computed({
    get: () => page1.value,
    set: value => {
      page1.value = value;
    },
  });

  const pageSize = useRouteQuery("pageSize", 30);
  const query = ref("");
  const fuzzySearch = ref(false);
  const advanced = useRouteQuery("advanced", false);
  const includeArchived = useRouteQuery("archived", false);
  const fieldSelector = useRouteQuery("fieldSelector", false);
  const negateLabels = useRouteQuery("negateLabels", false);
  const onlyWithoutPhoto = useRouteQuery("onlyWithoutPhoto", false);
  const onlyWithPhoto = useRouteQuery("onlyWithPhoto", false);
  const orderBy = useRouteQuery("orderBy", "name");

  const totalPages = computed(() => Math.ceil(total.value / pageSize.value));
  const hasNext = computed(() => page.value * pageSize.value < total.value);
  const hasPrev = computed(() => page.value > 1);

  function prev() {
    page.value = Math.max(1, page.value - 1);
  }

  function next() {
    page.value = Math.min(Math.ceil(total.value / pageSize.value), page.value + 1);
  }

  const route = useRoute();
  const router = useRouter();

  onMounted(async () => {
    loading.value = true;
    // Wait until locations and labels are loaded
    let maxRetry = 10;
    while (!labels.value || !locations.value) {
      await new Promise(resolve => setTimeout(resolve, 100));
      if (maxRetry-- < 0) {
        break;
      }
    }
    searchLocked.value = true;
    
    // Handle query parameters from URL
    if (route.query.q) {
      query.value = route.query.q as string;
    }
    
    if (route.query.fuzzySearch) {
      fuzzySearch.value = (route.query.fuzzySearch === 'true');
    }
    
    const qLoc = route.query.loc as string[];
    if (qLoc) {
      selectedLocations.value = locations.value.filter(l => qLoc.includes(l.id));
    }

    const qLab = route.query.lab as string[];
    if (qLab) {
      selectedLabels.value = labels.value.filter(l => qLab.includes(l.id));
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

  const locationFlatTree = await useFlatLocations();

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
        return "Invalid Asset ID";
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

  watch(advanced, (v, lv) => {
    if (v === false && lv === true) {
      selectedLocations.value = [];
      selectedLabels.value = [];
      fieldTuples.value = [];

      console.log("advanced", advanced.value);

      router.push({
        query: {
          advanced: route.query.advanced,
          q: query.value,
          page: page.value,
          pageSize: pageSize.value,
          includeArchived: includeArchived.value ? "true" : "false",
          negateLabels: negateLabels.value ? "true" : "false",
          onlyWithoutPhoto: onlyWithoutPhoto.value ? "true" : "false",
          onlyWithPhoto: onlyWithPhoto.value ? "true" : "false",
          orderBy: orderBy.value,
        },
      });
    }
  });

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

    const toast = useNotifier();

    const { data, error } = await api.items.getAll({
      q: query.value || "",
      fuzzySearch: fuzzySearch.value,
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
      toast.error("Failed to search items");
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

  watchDebounced([page, pageSize, query, fuzzySearch, selectedLabels, selectedLocations], search, { debounce: 250, maxWait: 1000 });

  async function submit() {
    // Set URL Params
    const fields = [];
    for (const t of fieldTuples.value) {
      if (t[0] && t[1]) {
        fields.push(`${t[0]}=${t[1]}`);
      }
    }

    await router.push({
      query: {
        archived: includeArchived.value.toString(),
        advanced: advanced.value.toString(),
        fieldSelector: fieldSelector.value.toString(),
        pageSize: pageSize.value.toString(),
        page: page.value.toString(),
        orderBy: orderBy.value,
        q: query.value,
        fuzzySearch: fuzzySearch.value.toString(),
        loc: locIDs.value,
        lab: labIDs.value,
        fields,
      },
    });

    await search();
  }

  async function reset() {
    // Reset all the filters
    query.value = "";
    fuzzySearch.value = false;
    selectedLabels.value = [];
    selectedLocations.value = [];
    negateLabels.value = false;
    onlyWithoutPhoto.value = false;
    onlyWithPhoto.value = false;
    includeArchived.value = false;
    advanced.value = false;
    fieldSelector.value = false;
    orderBy.value = "name";
    fieldTuples.value = [["", ""]];
    page.value = 1;

    // Trigger a search with the reset values
    search();
  }
</script>

<template>
  <BaseContainer class="mb-16">
    <div v-if="locations && labels">
      <div class="flex flex-wrap items-end gap-4 md:flex-nowrap">
        <div class="w-full">
          <FormTextField v-model="query" :placeholder="$t('global.search')" />
          <div v-if="byAssetId" class="pl-2 pt-2 text-sm">
            <p>{{ $t("items.query_id", { id: parsedAssetId }) }}</p>
          </div>
        </div>
        <BaseButton class="btn-block md:w-auto" @click.prevent="submit">
          <template #icon>
            <MdiLoading v-if="loading" class="animate-spin" />
            <MdiMagnify v-else />
          </template>
          {{ $t("global.search") }}
        </BaseButton>
      </div>

      <div class="flex w-full flex-wrap gap-2 py-2 md:flex-nowrap">
        <SearchFilter v-model="selectedLocations" :label="$t('global.locations')" :options="locationFlatTree">
          <template #display="{ item }">
            <div>
              <div class="flex w-full">
                {{ item.name }}
              </div>
              <div v-if="item.name != item.treeString" class="mt-1 text-xs">
                {{ item.treeString }}
              </div>
            </div>
          </template>
        </SearchFilter>
        <SearchFilter v-model="selectedLabels" :label="$t('global.labels')" :options="labels" />
        <div class="dropdown">
          <label tabindex="0" class="btn btn-xs">{{ $t("items.options") }}</label>
          <div
            tabindex="0"
            class="dropdown-content mt-1 w-72 -translate-x-24 overflow-auto rounded-md bg-base-100 p-4 shadow"
          >
            <label class="label mr-auto cursor-pointer">
              <input v-model="includeArchived" type="checkbox" class="toggle toggle-primary toggle-sm" />
              <span class="label-text ml-4 text-right"> {{ $t("items.include_archive") }} </span>
            </label>
            <label class="label mr-auto cursor-pointer">
              <input v-model="fieldSelector" type="checkbox" class="toggle toggle-primary toggle-sm" />
              <span class="label-text ml-4 text-right"> {{ $t("items.field_selector") }} </span>
            </label>
            <label class="label mr-auto cursor-pointer">
              <input v-model="negateLabels" type="checkbox" class="toggle toggle-primary toggle-sm" />
              <span class="label-text ml-4 text-right"> {{ $t("items.negate_labels") }} </span>
            </label>
            <label class="label mr-auto cursor-pointer">
              <input v-model="onlyWithoutPhoto" type="checkbox" class="toggle toggle-primary toggle-sm" />
              <span class="label-text ml-4 text-right"> {{ $t("items.only_without_photo") }} </span>
            </label>
            <label class="label mr-auto cursor-pointer">
              <input v-model="onlyWithPhoto" type="checkbox" class="toggle toggle-primary toggle-sm" />
              <span class="label-text ml-4 text-right"> {{ $t("items.only_with_photo") }} </span>
            </label>
            <label class="label mr-auto cursor-pointer">
              <input v-model="fuzzySearch" type="checkbox" class="toggle toggle-primary toggle-sm" />
              <span class="label-text ml-4 text-right"> {{ $t("items.fuzzy_search") }} </span>
            </label>
            <label class="label mr-auto cursor-pointer">
              <select v-model="orderBy" class="select select-bordered select-sm">
                <option value="name" selected>{{ $t("global.name") }}</option>
                <option value="createdAt">{{ $t("items.created_at") }}</option>
                <option value="updatedAt">{{ $t("items.updated_at") }}</option>
              </select>
              <span class="label-text ml-4 text-right"> {{ $t("items.order_by") }} </span>
            </label>
            <hr class="my-2" />
            <BaseButton class="btn-sm btn-block" @click="reset"> {{ $t("items.reset_search") }} </BaseButton>
          </div>
        </div>
        <div class="dropdown dropdown-end ml-auto">
          <label tabindex="0" class="btn btn-xs">{{ $t("items.tips") }}</label>
          <div
            tabindex="0"
            class="dropdown-content mt-1 w-[325px] overflow-auto rounded-md bg-base-100 p-4 text-sm shadow"
          >
            <p class="text-base">{{ $t("items.tips_sub") }}</p>
            <ul class="mt-1 list-disc pl-6">
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
          </div>
        </div>
      </div>
      <div v-if="fieldSelector" class="space-y-2 py-4">
        <p>{{ $t("items.custom_fields") }}</p>
        <div v-for="(f, idx) in fieldTuples" :key="idx" class="flex flex-wrap gap-2">
          <div class="form-control w-full max-w-xs">
            <label class="label">
              <span class="label-text">Field</span>
            </label>
            <select
              v-model="fieldTuples[idx][0]"
              class="select select-bordered"
              :items="allFields ?? []"
              @change="fetchValues(f[0])"
            >
              <option v-for="(fv, _, i) in allFields" :key="i" :value="fv">{{ fv }}</option>
            </select>
          </div>
          <div class="form-control w-full max-w-xs">
            <label class="label">
              <span class="label-text">{{ $t("items.field_value") }}</span>
            </label>
            <select v-model="fieldTuples[idx][1]" class="select select-bordered" :items="fieldValuesCache[f[0]]">
              <option v-for="v in fieldValuesCache[f[0]]" :key="v" :value="v">{{ v }}</option>
            </select>
          </div>
          <button
            type="button"
            class="btn btn-square btn-sm mb-2 ml-auto mt-auto md:ml-0"
            @click="fieldTuples.splice(idx, 1)"
          >
            <MdiDelete class="size-5" />
          </button>
        </div>
        <BaseButton type="button" class="btn-sm mt-2" @click="() => fieldTuples.push(['', ''])">
          {{ $t("items.add") }}
        </BaseButton>
      </div>
    </div>

    <section class="mt-10">
      <BaseSectionHeader ref="itemsTitle"> {{ $t("global.items") }} </BaseSectionHeader>
      <p v-if="items.length > 0" class="flex items-center text-base font-medium">
        {{ $t("items.results", { total: total }) }}
        <span class="ml-auto text-base"> {{ $t("items.pages", { page: page, totalPages: totalPages }) }} </span>
      </p>

      <div v-if="items.length === 0" class="flex flex-col items-center gap-2">
        <MdiSelectSearch class="size-10" />
        <p>{{ $t("items.no_results") }}</p>
      </div>
      <div v-else ref="cardgrid" class="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-5">
        <ItemCard v-for="item in items" :key="item.id" :item="item" :location-flat-tree="locationFlatTree" />
      </div>
      <div v-if="items.length > 0 && (hasNext || hasPrev)" class="mt-10 flex flex-col items-center gap-2">
        <div class="flex">
          <div class="btn-group">
            <button :disabled="!hasPrev" class="text-no-transform btn" @click="prev">
              <MdiChevronLeft class="mr-1 size-6" name="mdi-chevron-left" />
              {{ $t("items.prev_page") }}
            </button>
            <button v-if="hasPrev" class="text-no-transform btn" @click="page = 1">{{ $t("items.first") }}</button>
            <button v-if="hasNext" class="text-no-transform btn" @click="page = totalPages">
              {{ $t("items.last") }}
            </button>
            <button :disabled="!hasNext" class="text-no-transform btn" @click="next">
              {{ $t("items.next_page") }}
              <MdiChevronRight class="ml-1 size-6" name="mdi-chevron-right" />
            </button>
          </div>
        </div>
        <p class="text-sm font-bold">{{ $t("items.pages", { page: page, totalPages: totalPages }) }}</p>
      </div>
    </section>
  </BaseContainer>
</template>

<style lang="css">
  .list-move,
  .list-enter-active,
  .list-leave-active {
    transition: all 0.25s ease;
  }

  .list-enter-from,
  .list-leave-to {
    opacity: 0;
    transform: translateY(30px);
  }

  .list-leave-active {
    position: absolute;
  }
</style>
