<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
  import { Button } from "@/components/ui/button";
  import AppLogo from "~/components/App/Logo.vue";
  import MdiAlertCircle from "~icons/mdi/alert-circle";
  import MdiEmail from "~icons/mdi/email";
  import MdiLogin from "~icons/mdi/login";
  import MdiPackageVariant from "~icons/mdi/package-variant";

  definePageMeta({
    layout: "empty",
  });

  const { t } = useI18n();
  const route = useRoute();
  const api = usePublicApi();

  const kind = computed(() => route.params.kind as string);
  const id = computed(() => route.params.id as string);
  const isAsset = computed(() => kind.value === "asset");
  const isItem = computed(() => kind.value === "item");
  const hasInvalidKind = computed(() => !isAsset.value && !isItem.value);
  const originalPath = computed(() => {
    if (isAsset.value) {
      return `/a/${id.value}`;
    }
    if (isItem.value) {
      return `/item/${id.value}`;
    }
    return "/";
  });

  useHead({
    title: `HomeBox | ${t("found.title")}`,
  });

  const {
    data: found,
    pending,
    error,
  } = useAsyncData(`found-${kind.value}-${id.value}`, async () => {
    if (hasInvalidKind.value) {
      throw createError({
        statusCode: 404,
        statusMessage: t("found.invalid_kind"),
      });
    }

    const { data, error: apiError } = isAsset.value ? await api.foundAsset(id.value) : await api.foundItem(id.value);
    if (apiError) {
      return null;
    }
    return data;
  });

  const canContactOwner = computed(() => Boolean(found.value?.contactEmail && !found.value?.multipleMatches));
  const hasContactName = computed(() => Boolean(found.value?.contactName));
  const ownerNameText = computed(() => t("found.owner_name", { name: found.value?.contactName ?? "" }));
  const hasMultipleMatches = computed(() => Boolean(found.value?.multipleMatches));
  const notFoundText = computed(() =>
    hasInvalidKind.value || error.value ? t("found.invalid_kind") : t("found.not_found")
  );

  const contactHref = computed(() => {
    if (!found.value?.contactEmail) {
      return "";
    }

    const subject = t("found.email_subject", { assetId: found.value.assetId });
    const body = t("found.email_body");
    return `mailto:${found.value.contactEmail}?subject=${encodeURIComponent(subject)}&body=${encodeURIComponent(body)}`;
  });

  function signIn() {
    const redirectTo = useState("authRedirect");
    redirectTo.value = originalPath.value;
    navigateTo("/");
  }
</script>

<template>
  <main class="flex min-h-screen bg-background px-4 py-8 text-foreground">
    <div class="mx-auto flex w-full max-w-xl flex-col justify-center">
      <div class="mb-6 flex items-center justify-center gap-2 text-primary">
        <h1 class="flex items-center text-4xl font-bold">
          HomeB
          <AppLogo class="-mb-3 w-10" />
          x
        </h1>
      </div>

      <Card>
        <CardHeader>
          <CardTitle class="flex items-center gap-3 text-2xl">
            <MdiPackageVariant class="size-8 text-primary" />
            {{ $t("found.title") }}
          </CardTitle>
        </CardHeader>

        <CardContent class="space-y-4 text-sm text-muted-foreground">
          <p v-if="pending">{{ $t("global.loading") }}</p>

          <template v-else-if="canContactOwner">
            <p>{{ $t("found.contact_body") }}</p>
            <p v-if="hasContactName">
              {{ ownerNameText }}
            </p>
          </template>

          <template v-else-if="hasMultipleMatches">
            <p class="flex gap-2">
              <MdiAlertCircle class="mt-0.5 size-5 shrink-0" />
              <span>{{ $t("found.multiple_matches") }}</span>
            </p>
          </template>

          <template v-else-if="hasInvalidKind || error">
            <p class="flex gap-2">
              <MdiAlertCircle class="mt-0.5 size-5 shrink-0" />
              <span>{{ notFoundText }}</span>
            </p>
          </template>

          <template v-else>
            <p class="flex gap-2">
              <MdiAlertCircle class="mt-0.5 size-5 shrink-0" />
              <span>{{ $t("found.not_found") }}</span>
            </p>
          </template>
        </CardContent>

        <CardFooter class="flex flex-col gap-2 sm:flex-row">
          <Button v-if="canContactOwner" as-child class="w-full sm:w-auto">
            <a :href="contactHref">
              <MdiEmail />
              {{ $t("found.contact_owner") }}
            </a>
          </Button>

          <Button variant="outline" class="w-full sm:w-auto" @click="signIn">
            <MdiLogin />
            {{ $t("found.sign_in") }}
          </Button>
        </CardFooter>
      </Card>
    </div>
  </main>
</template>
