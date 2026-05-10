<script setup lang="ts">
  import MdiEmailOutline from "~icons/mdi/email-outline";
  import MdiLogin from "~icons/mdi/login";
  import AppLogo from "~/components/App/Logo.vue";
  import { Button } from "@/components/ui/button";

  definePageMeta({
    layout: "empty",
  });

  useHead({
    title: "HomeBox | Found Item",
  });

  const route = useRoute();
  const kind = computed(() => route.params.kind as string);
  const labelId = computed(() => route.params.id as string);
  const api = usePublicApi();

  const isAssetLabel = computed(() => kind.value === "asset");
  const isSupportedLabel = computed(() => kind.value === "item" || isAssetLabel.value);
  const labelPath = computed(() => `${isAssetLabel.value ? "/a" : "/item"}/${labelId.value}`);

  const { data: contact, pending } = useAsyncData(`found-${kind.value}-${labelId.value}`, async () => {
    if (!isSupportedLabel.value) {
      return null;
    }

    const { data, error } = isAssetLabel.value
      ? await api.foundAssetContact(labelId.value)
      : await api.foundEntityContact(labelId.value);
    if (error) {
      return null;
    }

    return data;
  });

  const mailtoHref = computed(() => {
    if (!contact.value?.ownerEmail) {
      return "";
    }

    const email = encodeURIComponent(contact.value.ownerEmail);
    const subject = encodeURIComponent("Found item");
    const body = encodeURIComponent(`I found an item with a HomeBox label: ${labelPath.value}`);
    return `mailto:${email}?subject=${subject}&body=${body}`;
  });

  function signIn() {
    const redirectTo = useState("authRedirect");
    redirectTo.value = labelPath.value;
    return navigateTo("/");
  }
</script>

<template>
  <main class="min-h-screen bg-background text-foreground">
    <section class="mx-auto flex min-h-screen w-full max-w-xl flex-col justify-center px-6 py-12">
      <AppLogo class="mb-8 size-16" />

      <div class="space-y-4">
        <p class="text-sm font-medium text-muted-foreground">HomeBox</p>
        <h1 class="text-3xl font-semibold tracking-normal">Did you find this item?</h1>
        <p class="text-base leading-7 text-muted-foreground">
          Contact the owner directly so they can arrange its return.
        </p>
      </div>

      <div class="mt-8 flex flex-col gap-3 sm:flex-row">
        <Button v-if="contact?.ownerEmail" as="a" :href="mailtoHref" class="w-full sm:w-auto">
          <MdiEmailOutline />
          Email owner
        </Button>

        <Button variant="outline" class="w-full sm:w-auto" @click="signIn">
          <MdiLogin />
          Sign in
        </Button>
      </div>

      <p v-if="pending" class="mt-6 text-sm text-muted-foreground">Loading contact details...</p>
      <p v-else-if="!contact?.ownerEmail" class="mt-6 text-sm text-muted-foreground">
        This label could not be resolved. Sign in to open HomeBox.
      </p>
    </section>
  </main>
</template>
