<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import MdiEmailOutline from "~icons/mdi/email-outline";
  import MdiArrowLeft from "~icons/mdi/arrow-left";
  import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
  import { Button } from "@/components/ui/button";
  import AppLogo from "~/components/App/Logo.vue";
  import FormTextField from "~/components/Form/TextField.vue";

  const { t } = useI18n();

  useHead({
    title: "HomeBox | " + t("index.forgot_password_title"),
  });

  definePageMeta({
    layout: "empty",
  });

  const api = usePublicApi();

  const email = ref("");
  const loading = ref(false);
  const submitted = ref(false);

  async function submit() {
    loading.value = true;
    const { error } = await api.forgotPassword(email.value.trim());
    loading.value = false;

    if (error) {
      // Server deliberately collapses configuration state and missing-account
      // cases into the same 204 — see HandleForgotPassword. So real errors
      // here are limited to rate limiting (429), demo mode (403), or network
      // failures, none of which we want to detail to an unauthenticated user.
      toast.error(t("index.toast.invalid_email"));
      return;
    }

    submitted.value = true;
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
          <MdiEmailOutline class="size-6" />
          {{ $t("index.forgot_password_title") }}
        </CardTitle>
      </CardHeader>

      <CardContent>
        <template v-if="submitted">
          <p class="text-sm">{{ $t("index.forgot_password_check_email") }}</p>
        </template>
        <template v-else>
          <p class="mb-4 text-sm text-muted-foreground">{{ $t("index.forgot_password_subtitle") }}</p>
          <form id="forgot-password-form" @submit.prevent="submit">
            <FormTextField
              id="forgot-email"
              v-model="email"
              :label="$t('global.email')"
              type="email"
              name="email"
              autocomplete="username"
              :required="true"
            />
          </form>
        </template>
      </CardContent>

      <CardFooter class="flex flex-col gap-2">
        <Button
          v-if="!submitted"
          form="forgot-password-form"
          type="submit"
          class="w-full"
          :disabled="loading || !email"
        >
          {{ $t("index.forgot_password_send") }}
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
