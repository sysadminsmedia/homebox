<script setup lang="ts">
  import { statCardData } from "./statistics";
  import { itemsTable } from "./table";
  import { useLabelStore } from "~~/stores/labels";
  import { useLocationStore } from "~~/stores/locations";

  definePageMeta({
    middleware: ["auth"],
  });
  useHead({
    title: "Homebox | Lost Item",
  });

  const api = useUserApi();
  const breakpoints = useBreakpoints();

  const locationStore = useLocationStore();
  const locations = computed(() => locationStore.parentLocations);

  const labelsStore = useLabelStore();
  const labels = computed(() => labelsStore.labels);

  const itemTable = itemsTable(api);
  const stats = statCardData(api);
</script>

<template>
  <template>
  <div class="lost-and-found">
    <h1>Lost and Found</h1>
    <p>You have located an item that may be lost. Please contact the owner here: <a :href="'mailto:' + ownerEmail">{{ ownerEmail }}</a></p>
    <div class="login-option">
      <p>Do you own this item? <router-link to="/login">Login to view or edit.</router-link></p>
    </div>
  </div>
</template>

<script>
export default {
  name: 'LostAndFound',
  data() {
    return {
      ownerEmail: 'katos@creatorswave.com'
    };
  }
};
</script>

<style scoped>
.lost-and-found {
  font-family: Arial, sans-serif;
  margin: 20px;
}

.lost-and-found h1 {
  font-size: 24px;
  color: #333;
}

.lost-and-found p {
  font-size: 16px;
  margin: 10px 0;
}

.login-option p {
  margin-top: 20px;
  font-weight: bold;
}

.login-option a {
  color: #007bff;
  text-decoration: none;
}

.login-option a:hover {
  text-decoration: underline;
}
</style>
</template>
