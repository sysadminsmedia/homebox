<script setup lang="ts">
  import type { Detail } from "~~/components/global/DetailsSection/types";
  import { themes } from "~~/lib/data/themes";
  import type { CurrenciesCurrency, NotifierCreate, NotifierOut } from "~~/lib/api/types/data-contracts";
  import MdiAccount from "~icons/mdi/account";
  import MdiMegaphone from "~icons/mdi/megaphone";
  import MdiDelete from "~icons/mdi/delete";
  import MdiFill from "~icons/mdi/fill";
  import MdiPencil from "~icons/mdi/pencil";
  import MdiAccountMultiple from "~icons/mdi/account-multiple";
import { useI18n } from "vue-i18n";

  definePageMeta({
    middleware: ["auth"],
  });
  useHead({
    title: "Homebox | Profile",
  });

  const api = useUserApi();
  const confirm = useConfirm();
  const notify = useNotifier();
  const { t } = useI18n();

  const currencies = computedAsync(async () => {
    const resp = await api.group.currencies();
    if (resp.error) {
      notify.error("Failed to get currencies");
      return [];
    }

    return resp.data;
  });

  const preferences = useViewPreferences();
  function setDisplayHeader() {
    preferences.value.displayHeaderDecor = !preferences.value.displayHeaderDecor;
  }
  function setLanguage(lang: string) {
    preferences.value.language = lang;
  }

  // Currency Selection
  const currency = ref<CurrenciesCurrency>({
    code: "USD",
    name: "United States Dollar",
    local: "en-US",
    symbol: "$",
  });
  watch(currency, () => {
    if (group.value) {
      group.value.currency = currency.value.code;
    }
  });

  const currencyExample = computed(() => {
    const formatter = new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: currency.value ? currency.value.code : "USD",
    });

    return formatter.format(1000);
  });

  const { data: group } = useAsyncData(async () => {
    const { data } = await api.group.get();
    return data;
  });

  // Sync Initial Currency
  watch(group, () => {
    if (!group.value) {
      return;
    }

    // @ts-expect-error - typescript is stupid, it should know group.value is not null
    const found = currencies.value.find(c => c.code === group.value.currency);
    if (found) {
      currency.value = found;
    }
  });

  async function updateGroup() {
    if (!group.value) {
      return;
    }

    const { data, error } = await api.group.update({
      name: group.value.name,
      currency: group.value.currency,
    });

    if (error) {
      notify.error("Failed to update group");
      return;
    }

    group.value = data;
    notify.success("Group updated");
  }

  const { setTheme } = useTheme();

  const auth = useAuthContext();

  const details = computed(() => {
    console.log(auth.user);
    return [
      {
        name: "global.name",
        text: auth.user?.name || "Unknown",
      },
      {
        name: "global.email",
        text: auth.user?.email || "Unknown",
      },
    ] as Detail[];
  });

  async function deleteProfile() {
    const result = await confirm.open(
      "Are you sure you want to delete your account? If you are the last member in your group all your data will be deleted. This action cannot be undone."
    );

    if (result.isCanceled) {
      return;
    }

    const { response } = await api.user.delete();

    if (response?.status === 204) {
      notify.success("Your account has been deleted.");
      auth.logout(api);
      navigateTo("/");
    }

    notify.error("Failed to delete your account.");
  }

  const token = ref("");
  const tokenUrl = computed(() => {
    if (!window) {
      return "";
    }

    return `${window.location.origin}?token=${token.value}`;
  });

  async function generateToken() {
    const date = new Date();

    const { response, data } = await api.group.createInvitation({
      expiresAt: new Date(date.setDate(date.getDate() + 7)),
      uses: 1,
    });

    if (response?.status === 201) {
      token.value = data.token;
    }
  }

  const passwordChange = reactive({
    loading: false,
    dialog: false,
    current: "",
    new: "",
    isValid: false,
  });

  function openPassChange() {
    passwordChange.dialog = true;
  }

  async function changePassword() {
    passwordChange.loading = true;
    if (!passwordChange.isValid) {
      return;
    }

    const { error } = await api.user.changePassword(passwordChange.current, passwordChange.new);

    if (error) {
      notify.error("Failed to change password.");
      passwordChange.loading = false;
      return;
    }

    notify.success("Password changed successfully.");
    passwordChange.dialog = false;
    passwordChange.new = "";
    passwordChange.current = "";
    passwordChange.loading = false;
  }

  // ===========================================================
  // Notifiers

  const notifiers = useAsyncData(async () => {
    const { data } = await api.notifiers.getAll();

    return data;
  });

  const targetID = ref("");
  const notifier = ref<NotifierCreate | null>(null);
  const notifierDialog = ref(false);

  function openNotifierDialog(v: NotifierOut | null) {
    if (v) {
      targetID.value = v.id;
      notifier.value = {
        name: v.name,
        url: v.url,
        isActive: v.isActive,
      };
    } else {
      notifier.value = {
        name: "",
        url: "",
        isActive: true,
      };
    }

    notifierDialog.value = true;
  }

  async function createNotifier() {
    if (!notifier.value) {
      return;
    }

    if (targetID.value) {
      await editNotifier();
      return;
    }

    const result = await api.notifiers.create({
      name: notifier.value.name,
      url: notifier.value.url || "",
      isActive: notifier.value.isActive,
    });

    if (result.error) {
      notify.error("Failed to create notifier.");
    }

    notifier.value = null;
    notifierDialog.value = false;

    await notifiers.refresh();
  }

  async function editNotifier() {
    if (!notifier.value) {
      return;
    }

    const result = await api.notifiers.update(targetID.value, {
      name: notifier.value.name,
      url: notifier.value.url || "",
      isActive: notifier.value.isActive,
    });

    if (result.error) {
      notify.error("Failed to update notifier.");
    }

    notifier.value = null;
    notifierDialog.value = false;
    targetID.value = "";

    await notifiers.refresh();
  }

  async function deleteNotifier(id: string) {
    const result = await confirm.open("Are you sure you want to delete this notifier?");

    if (result.isCanceled) {
      return;
    }

    const { error } = await api.notifiers.delete(id);

    if (error) {
      notify.error("Failed to delete notifier.");
      return;
    }

    await notifiers.refresh();
  }

  async function testNotifier() {
    if (!notifier.value) {
      return;
    }

    const { error } = await api.notifiers.test(notifier.value.url);

    if (error) {
      notify.error("Failed to test notifier.");
      return;
    }

    notify.success("Notifier test successful.");
  }
</script>

<template>
  <div>
    <BaseModal v-model="passwordChange.dialog">
      <template #title> {{ $t("profile.change_password") }} </template>

      <form @submit.prevent="changePassword">
        <FormPassword v-model="passwordChange.current" :label="$t('profile.current_password')" placeholder="" />
        <FormPassword v-model="passwordChange.new" :label="$t('profile.new_password')" placeholder="" />
        <PasswordScore v-model:valid="passwordChange.isValid" :password="passwordChange.new" />

        <div class="flex">
          <BaseButton
            class="ml-auto"
            :loading="passwordChange.loading"
            :disabled="!passwordChange.isValid"
            type="submit"
          >
            {{ $t("global.submit") }}
          </BaseButton>
        </div>
      </form>
    </BaseModal>

    <BaseModal v-model="notifierDialog">
      <template #title> {{ $t("profile.notifier_modal", { type: notifier != null }) }} </template>

      <form @submit.prevent="createNotifier">
        <template v-if="notifier">
          <FormTextField v-model="notifier.name" :label="$t('global.name')" />
          <FormTextField v-model="notifier.url" :label="$t('profile.url')" />
          <div class="max-w-[100px]">
            <FormCheckbox v-model="notifier.isActive" :label="$t('profile.enabled')" />
          </div>
        </template>
        <div class="mt-4 flex justify-between gap-2">
          <BaseButton :disabled="!(notifier && notifier.url)" type="button" @click="testNotifier">
            {{ $t("profile.test") }}
          </BaseButton>
          <BaseButton type="submit"> {{ $t("global.submit") }} </BaseButton>
        </div>
      </form>
    </BaseModal>

    <BaseContainer class="mb-6 flex flex-col gap-4">
      <BaseCard>
        <template #title>
          <BaseSectionHeader>
            <MdiAccount class="-mt-1 mr-2" />
            <span> {{ $t("profile.user_profile") }} </span>
            <template #description> {{ $t("profile.user_profile_sub") }} </template>
          </BaseSectionHeader>
        </template>

        <DetailsSection :details="details" />

        <div class="p-4">
          <div class="flex gap-2">
            <BaseButton size="sm" @click="openPassChange"> {{ $t("profile.change_password") }} </BaseButton>
            <BaseButton size="sm" @click="generateToken"> {{ $t("profile.gen_invite") }} </BaseButton>
          </div>
          <div v-if="token" class="flex items-center pl-1 pt-4">
            <CopyText class="btn btn-square btn-outline btn-primary btn-sm mr-2" :text="tokenUrl" />
            {{ tokenUrl }}
          </div>
          <div v-if="token" class="flex items-center pl-1 pt-4">
            <CopyText class="btn btn-square btn-outline btn-primary btn-sm mr-2" :text="token" />
            {{ token }}
          </div>
        </div>
        <div class="form-control w-full p-5 pt-0">
          <label class="label">
            <span class="label-text">{{ $t("profile.language") }}</span>
          </label>
          <select v-model="$i18n.locale" @change="(event) => {setLanguage((event.target as HTMLSelectElement).value )}"
            class="select select-bordered">
            <option v-for="lang in $i18n.availableLocales" :key="lang" :value="lang">
              {{ $t(`languages.${lang}`) }} ({{ $t(`languages.${lang}`, 1, { locale: lang }) }})
            </option>
          </select>
        </div>
      </BaseCard>

      <BaseCard>
        <template #title>
          <BaseSectionHeader>
            <MdiMegaphone class="-mt-1 mr-2" />
            <span class=""> {{ $t("profile.notifiers") }} </span>
            <template #description> {{ $t("profile.notifiers_sub") }} </template>
          </BaseSectionHeader>
        </template>

        <div v-if="notifiers.data.value" class="mx-4 divide-y divide-gray-400 rounded-md border border-gray-400">
          <p v-if="notifiers.data.value.length === 0" class="p-2 text-center text-sm">
            {{ $t("profile.no_notifiers") }}
          </p>
          <article v-for="n in notifiers.data.value" v-else :key="n.id" class="p-2">
            <div class="flex flex-wrap items-center gap-2">
              <p class="mr-auto text-lg">{{ n.name }}</p>
              <div class="flex justify-end gap-2">
                <div class="tooltip" data-tip="Delete">
                  <button class="btn btn-square btn-sm" @click="deleteNotifier(n.id)">
                    <MdiDelete />
                  </button>
                </div>
                <div class="tooltip" data-tip="Edit">
                  <button class="btn btn-square btn-sm" @click="openNotifierDialog(n)">
                    <MdiPencil />
                  </button>
                </div>
              </div>
            </div>
            <div class="flex flex-wrap justify-between py-1 text-sm">
              <p>
                <span v-if="n.isActive" class="badge badge-success"> {{ $t("profile.active") }} </span>
                <span v-else class="badge badge-error"> {{ $t("profile.inactive") }} </span>
              </p>
              <p>
                {{ $t("global.created") }}
                <DateTime format="relative" datetime-type="time" :date="n.createdAt" />
              </p>
            </div>
          </article>
        </div>

        <div class="p-4">
          <BaseButton size="sm" @click="openNotifierDialog"> {{ $t("global.create") }} </BaseButton>
        </div>
      </BaseCard>

      <BaseCard>
        <template #title>
          <BaseSectionHeader class="pb-0">
            <MdiAccountMultiple class="-mt-1 mr-2" />
            <span class=""> {{ $t("profile.group_settings") }} </span>
            <template #description>
              {{ $t("profile.group_settings_sub") }}
            </template>
          </BaseSectionHeader>
        </template>

        <div v-if="group && currencies && currencies.length > 0" class="p-5 pt-0">
          <FormSelect v-model="currency" :label="$t('profile.currency_format')" :items="currencies" />
          <p class="m-2 text-sm">{{$t("profile.example")}}: {{ currencyExample }}</p>

          <div class="mt-4">
            <BaseButton size="sm" @click="updateGroup"> {{ $t("profile.update_group") }} </BaseButton>
          </div>
        </div>
      </BaseCard>

      <BaseCard>
        <template #title>
          <BaseSectionHeader>
            <MdiFill class="mr-2" />
            <span class=""> {{ $t("profile.theme_settings") }} </span>
            <template #description>
              {{ $t("profile.theme_settings_sub") }}
            </template>
          </BaseSectionHeader>
        </template>

        <div class="px-4 pb-4">
          <div class="mb-3">
            <BaseButton size="sm" @click="setDisplayHeader">
              {{ $t("profile.display_header", { currentValue: preferences.displayHeaderDecor }) }}
            </BaseButton>
          </div>
          <div class="rounded-box grid grid-cols-1 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
            <div
              v-for="theme in themes"
              :key="theme.value"
              class="overflow-hidden rounded-lg border border-base-content/20 outline-2 outline-offset-2 outline-base-content hover:border-base-content/40"
              :data-theme="theme.value"
              :data-set-theme="theme.value"
              data-act-class="outline"
              @click="setTheme(theme.value)"
            >
              <div :data-theme="theme.value" class="w-full cursor-pointer bg-base-100 font-sans text-base-content">
                <div class="grid grid-cols-5 grid-rows-3">
                  <div class="col-start-1 row-span-2 row-start-1 bg-base-200"></div>
                  <div class="col-start-1 row-start-3 bg-base-300"></div>
                  <div class="col-span-4 col-start-2 row-span-3 row-start-1 flex flex-col gap-1 bg-base-100 p-2">
                    <div class="font-bold">{{ theme.label }}</div>
                    <div class="flex flex-wrap gap-1">
                      <div class="flex aspect-1 w-5 items-center justify-center rounded bg-primary lg:w-6">
                        <div class="text-sm font-bold text-primary-content">A</div>
                      </div>
                      <div class="flex aspect-1 w-5 items-center justify-center rounded bg-secondary lg:w-6">
                        <div class="text-sm font-bold text-secondary-content">A</div>
                      </div>
                      <div class="flex aspect-1 w-5 items-center justify-center rounded bg-accent lg:w-6">
                        <div class="text-sm font-bold text-accent-content">A</div>
                      </div>
                      <div class="flex aspect-1 w-5 items-center justify-center rounded bg-neutral lg:w-6">
                        <div class="text-sm font-bold text-neutral-content">A</div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </BaseCard>

      <BaseCard>
        <template #title>
          <BaseSectionHeader>
            <MdiDelete class="-mt-1 mr-2" />
            <span class=""> {{ $t("profile.delete_account") }} </span>
            <template #description> {{ $t("profile.delete_account_sub") }} </template>
          </BaseSectionHeader>
        </template>
        <div class="border-t-2 border-gray-300 p-4 px-6">
          <BaseButton size="sm" class="btn-error" @click="deleteProfile">
            {{ $t("profile.delete_account") }}
          </BaseButton>
        </div>
      </BaseCard>
    </BaseContainer>
  </div>
</template>

<style scoped></style>
