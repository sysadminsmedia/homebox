<script setup lang="ts">
  import type { ItemSummary, LabelSummary, LocationOutCount } from "~~/lib/api/types/data-contracts";
  import { useLabelStore } from "~~/stores/labels";
  import { useLocationStore } from "~~/stores/locations";
  import MdiLoading from "~icons/mdi/loading";
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

  const pageSize = useRouteQuery("pageSize", 21);
  const query = useRouteQuery("q", "");
  const advanced = useRouteQuery("advanced", false);
  const includeArchived = useRouteQuery("archived", false);
  const fieldSelector = useRouteQuery("fieldSelector", false);
  const negateLabels = useRouteQuery("negateLabels", false);
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
      locations: locIDs.value,
      labels: labIDs.value,
      negateLabels: negateLabels.value,
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

  watchDebounced([page, pageSize, query, selectedLabels, selectedLocations], search, { debounce: 250, maxWait: 1000 });

  async function submit() {
    // Set URL Params
    const fields = [];
    for (const t of fieldTuples.value) {
      if (t[0] && t[1]) {
        fields.push(`${t[0]}=${t[1]}`);
      }
    }

    // Push non-reactive query fields
    await router.push({
      query: {
        // Reactive
        advanced: "true",
        archived: includeArchived.value ? "true" : "false",
        fieldSelector: fieldSelector.value ? "true" : "false",
        negateLabels: negateLabels.value ? "true" : "false",
        orderBy: orderBy.value,
        pageSize: pageSize.value,
        page: page.value,
        q: query.value,

        // Non-reactive
        loc: locIDs.value,
        lab: labIDs.value,
        fields,
      },
    });

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

    await router.push({
      query: {
        archived: "false",
        fieldSelector: "false",
        pageSize: 10,
        page: 1,
        orderBy: "name",
        q: "",
        loc: [],
        lab: [],
        fields,
      },
    });

    await search();
  }
</script>

<template>
  <BaseContainer class="mb-16">
    <div v-if="locations && labels">
      <div class="flex flex-wrap md:flex-nowrap gap-4 items-end">
        <div class="w-full">
          <FormTextField v-model="query" placeholder="Search" />
          <div v-if="byAssetId" class="text-sm pl-2 pt-2">
            <p>Querying Asset ID Number: {{ parsedAssetId }}</p>
          </div>
        </div>
        <BaseButton class="btn-block md:w-auto" @click.prevent="submit">
          <template #icon>
            <MdiLoading v-if="loading" class="animate-spin" />
            <MdiMagnify v-else />
          </template>
          Search
        </BaseButton>
      </div>

      <div class="flex flex-wrap md:flex-nowrap gap-2 w-full py-2">
        <SearchFilter v-model="selectedLocations" label="Locations" :options="locationFlatTree">
          <template #display="{ item }">
            <div>
              <div class="flex w-full">
                {{ item.name }}
              </div>
              <div v-if="item.name != item.treeString" class="text-xs mt-1">
                {{ item.treeString }}
              </div>
            </div>
          </template>
        </SearchFilter>
        <SearchFilter v-model="selectedLabels" label="Labels" :options="labels" />
        <div class="dropdown">
          <label tabindex="0" class="btn btn-xs">Options</label>
          <div
            tabindex="0"
            class="dropdown-content mt-1 max-h-72 p-4 w-64 overflow-auto shadow bg-base-100 rounded-md -translate-x-24"
          >
            <label class="label cursor-pointer mr-auto">
              <input v-model="includeArchived" type="checkbox" class="toggle toggle-sm toggle-primary" />
              <span class="label-text ml-4"> Include Archived Items </span>
            </label>
            <label class="label cursor-pointer mr-auto">
              <input v-model="fieldSelector" type="checkbox" class="toggle toggle-sm toggle-primary" />
              <span class="label-text ml-4"> Field Selector </span>
            </label>
            <label class="label cursor-pointer mr-auto">
              <input v-model="negateLabels" type="checkbox" class="toggle toggle-sm toggle-primary" />
              <span class="label-text ml-4"> Negate selected labels </span>
            </label>
            <label class="label cursor-pointer mr-auto">
              <select v-model="orderBy" class="select select-bordered select-sm">
                <option value="name" selected>Name</option>
                <option value="createdAt">Created At</option>
                <option value="updatedAt">Updated At</option>
              </select>
              <span class="label-text ml-4"> Order By </span>
            </label>
            <hr class="my-2" />
            <BaseButton class="btn-block btn-sm" @click="reset"> Reset Search</BaseButton>
          </div>
        </div>
        <div class="dropdown ml-auto dropdown-end">
          <label tabindex="0" class="btn btn-xs">Tips</label>
          <div
            tabindex="0"
            class="dropdown-content mt-1 p-4 w-[325px] text-sm overflow-auto shadow bg-base-100 rounded-md"
          >
            <p class="text-base">Search Tips</p>
            <ul class="mt-1 list-disc pl-6">
              <li>
                Location and label filters use the 'OR' operation. If more than one is selected only one will be
                required for a match.
              </li>
              <li>Searches prefixed with '#'' will query for a asset ID (example '#000-001')</li>
              <li>
                Field filters use the 'OR' operation. If more than one is selected only one will be required for a
                match.
              </li>
            </ul>
          </div>
        </div>
      </div>
      <div v-if="fieldSelector" class="py-4 space-y-2">
        <p>Custom Fields</p>
        <div v-for="(f, idx) in fieldTuples" :key="idx" class="flex flex-wrap gap-2">
          <div class="form-control w-full max-w-xs">
            <label class="label">
              <span class="label-text">Field</span>
            </label>
            <select
              v-model="fieldTuples[idx][0]"
              class="select-bordered select"
              :items="allFields ?? []"
              @change="fetchValues(f[0])"
            >
              <option v-for="(fv, _, i) in allFields" :key="i" :value="fv">{{ fv }}</option>
            </select>
          </div>
          <div class="form-control w-full max-w-xs">
            <label class="label">
              <span class="label-text">Field Value</span>
            </label>
            <select v-model="fieldTuples[idx][1]" class="select-bordered select" :items="fieldValuesCache[f[0]]">
              <option v-for="v in fieldValuesCache[f[0]]" :key="v" :value="v">{{ v }}</option>
            </select>
          </div>
          <button
            type="button"
            class="btn btn-square btn-sm md:ml-0 ml-auto mt-auto mb-2"
            @click="fieldTuples.splice(idx, 1)"
          >
            <MdiDelete class="w-5 h-5" />
          </button>
        </div>
        <BaseButton type="button" class="btn-sm mt-2" @click="() => fieldTuples.push(['', ''])"> Add</BaseButton>
      </div>
    </div>

    <section class="mt-10">
      <BaseSectionHeader ref="itemsTitle"> Items </BaseSectionHeader>
      <p class="text-base font-medium flex items-center">
        {{ total }} Results
        <span class="text-base ml-auto"> Page {{ page }} of {{ totalPages }}</span>
      </p>

      <div ref="cardgrid" class="grid mt-4 grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
        <ItemCard v-for="item in items" :key="item.id" :item="item" />

        <div class="hidden first:inline text-xl">No Items Found</div>
      </div>
      <div v-if="items.length > 0 && (hasNext || hasPrev)" class="mt-10 flex gap-2 flex-col items-center">
        <div class="flex">
          <div class="btn-group">
            <button :disabled="!hasPrev" class="btn text-no-transform" @click="prev">
              <MdiChevronLeft class="mr-1 h-6 w-6" name="mdi-chevron-left" />
              Prev
            </button>
            <button v-if="hasPrev" class="btn text-no-transform" @click="page = 1">First</button>
            <button v-if="hasNext" class="btn text-no-transform" @click="page = totalPages">Last</button>
            <button :disabled="!hasNext" class="btn text-no-transform" @click="next">
              Next
              <MdiChevronRight class="ml-1 h-6 w-6" name="mdi-chevron-right" />
            </button>
          </div>
        </div>
        <p class="text-sm font-bold">Page {{ page }} of {{ totalPages }}</p>
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
