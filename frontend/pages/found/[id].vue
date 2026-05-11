<script setup lang="ts">
  import { Button } from "@/components/ui/button";
  import { Card } from "@/components/ui/card";
  import AppLogo from "~/components/App/Logo.vue";
  import MdiEmailOutline from "~icons/mdi/email-outline";
  import MdiLogin from "~icons/mdi/login";
  import { useI18n } from "vue-i18n";

  definePageMeta({
    layout: "empty",
  });

  const { t } = useI18n();
  const route = useRoute();
  const api = usePublicApi();
  const itemId = computed<string>(() => {
    const id = route.params.id;
    return typeof id === "string" ? id : (id?.[0] ?? "");
  });

  const {
    data: foundItem,
    pending,
    refresh,
  } = await useAsyncData(`found-item-${itemId.value}`, async () => {
    const { data, error } = await api.foundEntity(itemId.value);
    if (error) {
      return null;
    }

    return data;
  });
  watch(itemId, () => refresh());

  const mailto = computed(() => {
    if (!foundItem.value?.ownerEmail) {
      return "";
    }

    const subject = encodeURIComponent(t("found_item.email_subject", { item: foundItem.value.name }));
    return `mailto:${foundItem.value.ownerEmail}?subject=${subject}`;
  });

  function signIn() {
    const redirectTo = useState("authRedirect");
    redirectTo.value = `/item/${itemId.value}`;
    return navigateTo("/");
  }
</script>

<template>
  <main class="flex min-h-screen items-center justify-center bg-muted/30 p-4">
    <Card class="w-full max-w-md p-6 shadow-sm">
      <div class="mb-6 flex justify-center">
        <AppLogo class="size-16" />
      </div>

      <div v-if="pending" class="text-center text-sm text-muted-foreground">
        {{ $t("global.loading") }}
      </div>

      <div v-else-if="foundItem" class="space-y-6">
        <div class="space-y-2 text-center">
          <h1 class="text-2xl font-semibold">{{ $t("found_item.title") }}</h1>
          <p class="text-muted-foreground">
            {{ $t("found_item.contact_owner") }}
          </p>
        </div>

        <dl class="rounded-md border bg-background p-4 text-sm">
          <div class="flex justify-between gap-4">
            <dt class="text-muted-foreground">{{ $t("items.name") }}</dt>
            <dd class="text-right font-medium">{{ foundItem.name }}</dd>
          </div>
          <div v-if="foundItem.assetId && foundItem.assetId !== '000-000'" class="mt-3 flex justify-between gap-4">
            <dt class="text-muted-foreground">{{ $t("items.asset_id") }}</dt>
            <dd class="text-right font-mono text-xs">{{ foundItem.assetId }}</dd>
          </div>
        </dl>

        <div class="flex flex-col gap-2">
          <Button as-child>
            <a :href="mailto">
              <MdiEmailOutline />
              {{ $t("found_item.email_owner") }}
            </a>
          </Button>
          <Button variant="outline" @click="signIn">
            <MdiLogin />
            {{ $t("found_item.sign_in") }}
          </Button>
        </div>
      </div>

      <div v-else class="space-y-5 text-center">
        <div class="space-y-2">
          <h1 class="text-2xl font-semibold">{{ $t("found_item.not_found_title") }}</h1>
          <p class="text-muted-foreground">{{ $t("found_item.not_found_description") }}</p>
        </div>
        <Button variant="outline" @click="signIn">
          <MdiLogin />
          {{ $t("found_item.sign_in") }}
        </Button>
      </div>
    </Card>
  </main>
</template>
