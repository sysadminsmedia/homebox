<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import MdiLockReset from "~icons/mdi/lock-reset";
  import MdiArrowLeft from "~icons/mdi/arrow-left";
  import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
  import { Button } from "@/components/ui/button";
  import AppLogo from "~/components/App/Logo.vue";
  import FormPassword from "~/components/Form/Password.vue";
  import PasswordScore from "~/components/global/PasswordScore.vue";
  import { PASSWORD_MIN_LENGTH, PASSWORD_RULES } from "~/lib/passwords";

  const { t } = useI18n();

  useHead({
    title: "HomeBox | " + t("index.reset_password_title"),
  });

  definePageMeta({
    layout: "empty",
  });

  const route = useRoute();
  const router = useRouter();
  const api = usePublicApi();

  const token = computed(() => {
    const t = route.query.token;
    return typeof t === "string" ? t : "";
  });

  const password = ref("");
  const passwordValid = ref(false);
  const loading = ref(false);

  onMounted(() => {
    if (!token.value) {
      toast.error(t("index.reset_password_missing_token"));
    }
  });

  async function submit() {
    if (!token.value || !passwordValid.value) return;

    loading.value = true;
    const { error } = await api.resetPassword(token.value, password.value);
    loading.value = false;

    if (error) {
      // 400 means the token is invalid/expired/used; anything else is also
      // unactionable for the user. Show one message either way.
      toast.error(t("index.reset_password_invalid"));
      return;
    }

    toast.success(t("index.reset_password_success"));
    router.replace("/");
  }
</script>

<template>
  <div class="flex min-h-screen flex-col items-center justify-center p-6">
    <div class="mb-6 flex items-center gap-2 text-3xl font-bold tracking-tight">
      HomeB
      <AppLogo class="-mb-2 w-10" />
      x
    </div>

    <Card class="md:w-[460px]">
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <MdiLockReset class="size-6" />
          {{ $t("index.reset_password_title") }}
        </CardTitle>
      </CardHeader>

      <CardContent>
        <p class="mb-4 text-sm text-muted-foreground">{{ $t("index.reset_password_subtitle") }}</p>
        <form id="reset-password-form" @submit.prevent="submit">
          <FormPassword
            id="reset-password-new"
            v-model="password"
            :label="$t('index.set_password')"
            name="new-password"
            autocomplete="new-password"
            :min-length="PASSWORD_MIN_LENGTH"
            :passwordrules="PASSWORD_RULES"
            :required="true"
          />
          <PasswordScore v-model:valid="passwordValid" :password="password" />
        </form>
      </CardContent>

      <CardFooter class="flex flex-col gap-2">
        <Button form="reset-password-form" type="submit" class="w-full" :disabled="loading || !passwordValid || !token">
          {{ $t("index.reset_password_set") }}
        </Button>
        <NuxtLink to="/" class="text-sm text-muted-foreground hover:underline">
          <span class="inline-flex items-center gap-1">
            <MdiArrowLeft class="size-4" />
            {{ $t("index.back_to_login") }}
          </span>
        </NuxtLink>
      </CardFooter>
    </Card>
  </div>
</template>
