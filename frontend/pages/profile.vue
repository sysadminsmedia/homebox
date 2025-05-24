<script setup lang="ts">
  import { toast } from "@/components/ui/sonner";
  import type { Detail } from "~~/components/global/DetailsSection/types";
  import { themes } from "~~/lib/data/themes";
  import type { CurrenciesCurrency, NotifierCreate, NotifierOut } from "~~/lib/api/types/data-contracts";
  import MdiAccount from "~icons/mdi/account";
  import MdiMegaphone from "~icons/mdi/megaphone";
  import MdiDelete from "~icons/mdi/delete";
  import MdiFill from "~icons/mdi/fill";
  import MdiPencil from "~icons/mdi/pencil";
  import MdiAccountMultiple from "~icons/mdi/account-multiple";
  import { getLocaleCode } from "~/composables/use-formatters";
  import { Button } from "@/components/ui/button";
  import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
  import { Label } from "@/components/ui/label";
  import { badgeVariants } from "@/components/ui/badge";
  import LanguageSelector from "~/components/App/LanguageSelector.vue";
  import { Tooltip, TooltipContent, TooltipTrigger, TooltipProvider } from "@/components/ui/tooltip";

  definePageMeta({
    middleware: ["auth"],
  });
  useHead({
    title: "Homebox | Profile",
  });

  const api = useUserApi();
  const confirm = useConfirm();

  const { openDialog, closeDialog } = useDialog();

  const currencies = computedAsync(async () => {
    const resp = await api.group.currencies();
    if (resp.error) {
      toast.error("Failed to get currencies");
      return [];
    }

    return resp.data;
  });

  const preferences = useViewPreferences();
  function setDisplayHeader() {
    preferences.value.displayLegacyHeader = !preferences.value.displayLegacyHeader;
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
    return fmtCurrency(1000, currency.value?.code ?? "USD", getLocaleCode());
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
      toast.error("Failed to update group");
      return;
    }

    group.value = data;
    toast.success("Group updated");
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
      toast.success("Your account has been deleted.");
      auth.logout(api);
      navigateTo("/");
    }

    toast.error("Failed to delete your account.");
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
    current: "",
    new: "",
    isValid: false,
  });

  async function changePassword() {
    passwordChange.loading = true;
    if (!passwordChange.isValid) {
      return;
    }

    const { error } = await api.user.changePassword(passwordChange.current, passwordChange.new);

    if (error) {
      toast.error("Failed to change password.");
      passwordChange.loading = false;
      return;
    }

    toast.success("Password changed successfully.");
    closeDialog("change-password");
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

    openDialog("create-notifier");
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
      toast.error("Failed to create notifier.");
    }

    notifier.value = null;
    closeDialog("create-notifier");

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
      toast.error("Failed to update notifier.");
    }

    notifier.value = null;
    closeDialog("create-notifier");
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
      toast.error("Failed to delete notifier.");
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
      toast.error("Failed to test notifier.");
      return;
    }

    toast.success("Notifier test successful.");
  }
</script>

<template>
  <div>
    <Dialog dialog-id="changePassword">
      <DialogContent>
        <DialogHeader>
          <DialogTitle> {{ $t("profile.change_password") }} </DialogTitle>
        </DialogHeader>

        <FormPassword
          v-model="passwordChange.current"
          :label="$t('profile.current_password')"
          placeholder=""
          class="mb-2"
        />
        <FormPassword v-model="passwordChange.new" :label="$t('profile.new_password')" placeholder="" />
        <PasswordScore v-model:valid="passwordChange.isValid" :password="passwordChange.new" />

        <form @submit.prevent="changePassword">
          <DialogFooter>
            <Button :loading="passwordChange.loading" :disabled="!passwordChange.isValid" type="submit">
              {{ $t("global.submit") }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <Dialog dialog-id="create-notifier">
      <DialogContent>
        <DialogHeader>
          <DialogTitle> {{ $t("profile.notifier_modal", { type: notifier != null }) }} </DialogTitle>
        </DialogHeader>

        <form @submit.prevent="createNotifier">
          <template v-if="notifier">
            <FormTextField v-model="notifier.name" :label="$t('global.name')" class="mb-2" />
            <FormTextField v-model="notifier.url" :label="$t('profile.url')" class="mb-2" />
            <div class="max-w-[100px]">
              <FormCheckbox v-model="notifier.isActive" :label="$t('profile.enabled')" />
            </div>
          </template>
          <div class="mt-4 flex justify-between gap-2">
            <DialogFooter class="flex w-full">
              <Button variant="secondary" :disabled="!(notifier && notifier.url)" type="button" @click="testNotifier">
                {{ $t("profile.test") }}
              </Button>
              <div class="grow"></div>
              <Button type="submit"> {{ $t("global.submit") }} </Button>
            </DialogFooter>
          </div>
        </form>
      </DialogContent>
    </Dialog>

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
            <Button variant="secondary" size="sm" @click="openDialog('changePassword')">
              {{ $t("profile.change_password") }}
            </Button>
            <Button variant="secondary" size="sm" @click="generateToken"> {{ $t("profile.gen_invite") }} </Button>
          </div>
          <div v-if="token" class="flex items-center gap-2 pl-1 pt-4">
            <CopyText :text="tokenUrl" />
            {{ tokenUrl }}
          </div>
          <div v-if="token" class="flex items-center gap-2 pl-1 pt-4">
            <CopyText :text="token" />
            {{ token }}
          </div>
        </div>
        <LanguageSelector />
      </BaseCard>

      <BaseCard>
        <template #title>
          <BaseSectionHeader>
            <MdiMegaphone class="-mt-1 mr-2" />
            <span class=""> {{ $t("profile.notifiers") }} </span>
            <template #description> {{ $t("profile.notifiers_sub") }} </template>
          </BaseSectionHeader>
        </template>

        <div v-if="notifiers.data.value" class="mx-4 divide-y rounded-md border">
          <p v-if="notifiers.data.value.length === 0" class="p-2 text-center text-sm">
            {{ $t("profile.no_notifiers") }}
          </p>
          <article v-for="n in notifiers.data.value" v-else :key="n.id" class="p-2">
            <div class="flex flex-wrap items-center gap-2">
              <p class="mr-auto text-lg">{{ n.name }}</p>
              <TooltipProvider :delay-duration="0" class="flex justify-end gap-2">
                <Tooltip>
                  <TooltipTrigger>
                    <Button variant="destructive" size="icon" @click="deleteNotifier(n.id)">
                      <MdiDelete />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent> Delete </TooltipContent>
                </Tooltip>
                <Tooltip>
                  <TooltipTrigger>
                    <Button variant="outline" size="icon" @click="openNotifierDialog(n)">
                      <MdiPencil />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent> Edit </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
            <div class="flex flex-wrap justify-between py-1 text-sm">
              <p>
                <span v-if="n.isActive" :class="badgeVariants()"> {{ $t("profile.active") }} </span>
                <span v-else :class="badgeVariants({ variant: 'destructive' })"> {{ $t("profile.inactive") }} </span>
              </p>
              <p>
                {{ $t("global.created") }}
                <DateTime format="relative" datetime-type="time" :date="n.createdAt" />
              </p>
            </div>
          </article>
        </div>

        <div class="p-4">
          <Button variant="secondary" size="sm" @click="openNotifierDialog"> {{ $t("global.create") }} </Button>
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
          <Label for="currency"> {{ $t("profile.currency_format") }} </Label>
          <Select
            id="currency"
            :model-value="currency.code"
            @update:model-value="
              event => {
                const newCurrency = currencies?.find(c => c.code === event);
                if (newCurrency) {
                  currency = newCurrency;
                }
              }
            "
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="c in currencies" :key="c.code" :value="c.code">
                {{ c.name }}
              </SelectItem>
            </SelectContent>
          </Select>
          <p class="m-2 text-sm">{{ $t("profile.example") }}: {{ currencyExample }}</p>

          <div class="mt-4">
            <Button variant="secondary" size="sm" @click="updateGroup"> {{ $t("profile.update_group") }} </Button>
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
            <Button variant="secondary" size="sm" @click="setDisplayHeader">
              {{ $t("profile.display_legacy_header", { currentValue: preferences.displayLegacyHeader }) }}
            </Button>
          </div>
          <div class="homebox grid grid-cols-1 gap-4 font-sans sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
            <div
              v-for="theme in themes"
              :key="theme.value"
              :class="'theme-' + theme.value"
              class="overflow-hidden rounded-lg border outline-2 outline-offset-2"
              :data-theme="theme.value"
              :data-set-theme="theme.value"
              data-act-class="outline"
              @click="setTheme(theme.value)"
            >
              <div :data-theme="theme.value" class="w-full cursor-pointer bg-background-accent text-foreground">
                <div class="grid grid-cols-5 grid-rows-3">
                  <div class="col-start-1 row-start-1 bg-background"></div>
                  <div class="col-start-1 row-start-2 bg-sidebar"></div>
                  <div class="col-start-1 row-start-3 bg-background-accent"></div>
                  <div class="col-span-4 col-start-2 row-span-3 row-start-1 flex flex-col gap-1 bg-background p-2">
                    <div class="font-bold">{{ theme.label }}</div>
                    <div class="flex flex-wrap gap-1">
                      <div class="flex size-5 items-center justify-center rounded bg-primary lg:size-6">
                        <div class="text-sm font-bold text-primary-foreground">A</div>
                      </div>
                      <div class="flex size-5 items-center justify-center rounded bg-secondary lg:size-6">
                        <div class="text-sm font-bold text-secondary-foreground">A</div>
                      </div>
                      <div class="flex size-5 items-center justify-center rounded bg-accent lg:size-6">
                        <div class="text-sm font-bold text-accent-foreground">A</div>
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
        <div class="border-t-2 p-4 px-6">
          <Button size="sm" variant="destructive" @click="deleteProfile">
            {{ $t("profile.delete_account") }}
          </Button>
        </div>
      </BaseCard>
    </BaseContainer>
  </div>
</template>

<style scoped></style>
