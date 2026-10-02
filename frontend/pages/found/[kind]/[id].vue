<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import MdiLoading from "~icons/mdi/loading";
  import MdiMapMarkerAlert from "~icons/mdi/map-marker-alert";
  import MdiEmailOutline from "~icons/mdi/email-outline";
  import MdiArrowLeft from "~icons/mdi/arrow-left";
  import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
  import { Button } from "@/components/ui/button";
  import AppLogo from "~/components/App/Logo.vue";
  import FormTextArea from "~/components/Form/TextArea.vue";
  import FormTextField from "~/components/Form/TextField.vue";
  import type { FoundItemKind, FoundItemResult } from "~~/lib/api/public";
  import { utf8Length } from "@/lib/utils";

  const { t } = useI18n();

  definePageMeta({
    layout: "empty",
  });

  useHead({
    title: "HomeBox | " + t("found.title"),
  });

  const route = useRoute();
  const api = usePublicApi();

  const kind = computed<FoundItemKind>(() => (route.params.kind === "item" ? "item" : "asset"));
  const id = computed(() => String(route.params.id));

  const item = ref<FoundItemResult | null>(null);
  const pending = ref(true);

  onMounted(async () => {
    try {
      const { data, error } = await api.foundItem(kind.value, id.value);
      item.value = error ? null : (data ?? null);
    } catch {
      item.value = null;
    } finally {
      pending.value = false;
    }
  });

  const message = ref("");
  const replyTo = ref("");
  const sending = ref(false);
  const sent = ref(false);
  const sendError = ref(false);

  async function submit() {
    if (!message.value.trim()) {
      return;
    }

    sending.value = true;
    sendError.value = false;

    const { error } = await api.foundContact(kind.value, id.value, message.value.trim(), replyTo.value.trim());

    sending.value = false;

    if (error) {
      sendError.value = true;
      return;
    }

    sent.value = true;
  }
</script>

<template>
  <div class="flex min-h-screen flex-col items-center justify-center p-6">
    <div class="mb-6 flex items-center gap-2 whitespace-nowrap text-3xl font-bold tracking-tight">
      HomeB
      <AppLogo class="-mb-2 w-10" />
      x
    </div>

    <div v-if="pending" class="flex flex-col items-center gap-2 text-sm text-muted-foreground">
      <div class="flex items-center gap-2">
        <MdiLoading class="size-5 animate-spin" />
        {{ $t("global.loading") }}
      </div>
      <NuxtLink to="/" class="text-sm text-muted-foreground hover:underline">
        <span class="inline-flex items-center gap-1">
          <MdiArrowLeft class="size-4" />
          {{ $t("found.sign_in") }}
        </span>
      </NuxtLink>
    </div>

    <Card v-else-if="!item" class="md:w-[460px]">
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <MdiMapMarkerAlert class="size-6" />
          {{ $t("found.title") }}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p class="text-sm">{{ $t("found.not_available") }}</p>
      </CardContent>
      <CardFooter>
        <NuxtLink to="/" class="text-sm text-muted-foreground hover:underline">
          <span class="inline-flex items-center gap-1">
            <MdiArrowLeft class="size-4" />
            {{ $t("found.sign_in") }}
          </span>
        </NuxtLink>
      </CardFooter>
    </Card>

    <Card v-else class="md:w-[460px]">
      <CardHeader>
        <CardTitle class="flex items-center gap-2">
          <MdiMapMarkerAlert class="size-6" />
          {{ $t("found.title") }}
        </CardTitle>
      </CardHeader>

      <CardContent class="flex flex-col gap-4">
        <p class="text-sm text-muted-foreground">{{ $t("found.description") }}</p>

        <div v-if="item.message" class="rounded-md border border-border bg-muted/50 p-3">
          <p class="mb-1 text-xs font-medium text-muted-foreground">{{ $t("found.owner_message") }}</p>
          <p class="whitespace-pre-wrap text-sm">{{ item.message }}</p>
        </div>

        <template v-if="item.mode === 'mailto'">
          <Button as-child class="w-full">
            <a :href="`mailto:${item.email}`" class="inline-flex items-center justify-center gap-2">
              <MdiEmailOutline class="size-4" />
              {{ $t("found.mailto") }}
            </a>
          </Button>
        </template>

        <template v-else>
          <template v-if="sent">
            <p class="text-sm">{{ $t("found.contact_form.sent") }}</p>
          </template>
          <form v-else id="found-contact-form" class="flex flex-col gap-3" @submit.prevent="submit">
            <FormTextArea
              v-model="message"
              :label="$t('found.contact_form.label')"
              :placeholder="$t('found.contact_form.placeholder')"
              :max-length="2000"
            />
            <FormTextField
              id="found-reply-to"
              v-model="replyTo"
              :label="$t('found.contact_form.reply_to')"
              type="email"
              name="email"
              autocomplete="email"
            />
            <p v-if="sendError" class="text-sm text-destructive">{{ $t("found.contact_form.error") }}</p>
          </form>
        </template>
      </CardContent>

      <CardFooter class="flex flex-col gap-2">
        <Button
          v-if="item.mode === 'form' && !sent"
          form="found-contact-form"
          type="submit"
          class="w-full"
          :disabled="sending || !message.trim() || utf8Length(message) > 2000"
        >
          {{ sending ? $t("found.contact_form.sending") : $t("found.contact_form.send") }}
        </Button>
        <NuxtLink to="/" class="text-sm text-muted-foreground hover:underline">
          <span class="inline-flex items-center gap-1">
            <MdiArrowLeft class="size-4" />
            {{ $t("found.sign_in") }}
          </span>
        </NuxtLink>
      </CardFooter>
    </Card>
  </div>
</template>
