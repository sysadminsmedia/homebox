<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import BaseContainer from "@/components/Base/Container.vue";
  import BaseSectionHeader from "@/components/Base/SectionHeader.vue";
  import ItemCard from "~/components/Item/Card.vue";

  const { t } = useI18n();

  definePageMeta({
    middleware: ["auth"],
  });

  const route = useRoute();
  const api = useUserApi();

  const assetId = computed<string>(() => route.params.id as string);

  const { pending, data: items } = useLazyAsyncData(`asset/${assetId.value}`, async () => {
    const { data, error } = await api.assets.get(assetId.value);
    if (error) {
      toast.error(t("items.toast.failed_to_load_asset"));
      navigateTo("/home");
      return;
    }
    switch (data.total) {
      case 0:
        toast.error(t("items.toast.asset_not_found"));
        navigateTo("/home");
        break;
      case 1:
        navigateTo(`/item/${data.items[0]!.id}`, { replace: true, redirectCode: 302 });
        break;
      default:
        return data.items;
    }
  });
</script>

<template>
  <BaseContainer>
    <section v-if="!pending">
      <BaseSectionHeader class="mb-5"> {{ $t("items.associated_with_multiple") }} </BaseSectionHeader>
      <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
        <ItemCard v-for="item in items" :key="item.id" :item="item" />
      </div>
    </section>
  </BaseContainer>
</template>
