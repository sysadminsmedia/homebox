<script setup lang="ts">
  import { Button } from "@/components/ui/button";
  const api = useUserApi();

  async function callApi() {
    try {
      // const res = await api.group.addUser(
      //   "45a64854-535e-4474-affc-da2a20ede7ff",
      //   "b348fb6e-45a5-4756-b417-5652cc38dd06"
      // );
      const res1 = await api.user.self();
      // const res = await api.group.delete("45a64854-535e-4474-affc-da2a20ede7ff");
      const res2 = await api.group.get();
      // const res = await api.group.getMembers("45a64854-535e-4474-affc-da2a20ede7ff");

      console.log("API result:", res1, res2);
    } catch (err: any) {
      console.error("API error:", err);
    }
  }
</script>

<template>
  <div class="max-w-3xl p-6">
    <h1 class="mb-4 text-2xl font-semibold">Testing Page</h1>

    <Button @click="callApi">Call API</Button>
  </div>
</template>
