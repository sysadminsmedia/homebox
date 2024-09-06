<script setup>
  definePageMeta({
    middleware: ["auth"],
  });

  const show = reactive({
    identification: true,
    purchase: false,
    sold: false,
    extras: false,
  });

  const form = reactive({
    name: "",
    description: "",
    notes: "",

    // Item Identification
    serialNumber: "",
    modelNumber: "",
    manufacturer: "",

    // Purchase Information
    purchaseTime: "",
    purchasePrice: "",
    purchaseFrom: "",

    // Sold Information
    soldTime: "",
    soldPrice: "",
    soldTo: "",
    soldNotes: "",
  });

  function submit() {
    console.log("Submitted!");
  }
</script>

<template>
  <BaseContainer cmp="section">
    <BaseSectionHeader> Add an Item To Your Inventory </BaseSectionHeader>
    <form class="mx-auto my-5 max-w-3xl space-y-6" @submit.prevent="submit">
      <div class="collapse-title divider cursor-pointer px-0">Required Information</div>
      <div class="card bg-base-200">
        <div class="card-body">
          <FormTextField v-model="form.name" label="Name" />
          <FormTextArea v-model="form.description" label="Description" limit="1000" />
        </div>
      </div>

      <div class="divider">
        <button class="btn btn-sm" @click="show.identification = !show.identification">Product Information</button>
      </div>
      <div v-if="show.identification" class="card bg-base-200">
        <div class="card-body grid md:grid-cols-2">
          <FormTextField v-model="form.serialNumber" label="Serial Number" />
          <FormTextField v-model="form.modelNumber" label="Model Number" />
          <FormTextField v-model="form.manufacturer" label="Manufacturer" />
        </div>
      </div>
      <div class="">
        <button class="btn btn-sm" @click="show.purchase = !show.purchase">Purchase Information</button>
        <div class="divider"></div>
      </div>
      <div v-if="show.purchase" class="card bg-base-200">
        <div class="card-body grid md:grid-cols-2">
          <FormTextField v-model="form.purchaseTime" label="Purchase Time" />
          <FormTextField v-model="form.purchasePrice" label="Purchase Price" />
          <FormTextField v-model="form.purchaseFrom" label="Purchase From" />
        </div>
      </div>

      <div class="divider">
        <button class="btn btn-sm" @click="show.sold = !show.sold">Sold Information</button>
      </div>
      <div v-if="show.sold" class="card bg-base-200">
        <div class="card-body">
          <div class="grid gap-2 md:grid-cols-2">
            <FormTextField v-model="form.soldTime" label="Sold Time" />
            <FormTextField v-model="form.soldPrice" label="Sold Price" />
            <FormTextField v-model="form.soldTo" label="Sold To" />
          </div>
          <FormTextArea v-model="form.soldNotes" label="Sold Notes" limit="1000" />
        </div>
      </div>
      <div class="divider">
        <button class="btn btn-sm" @click="show.extras = !show.extras">Extras</button>
      </div>
      <div v-if="show.extras" class="card bg-base-200">
        <div class="card-body">
          <FormTextArea v-model="form.notes" label="Notes" limit="1000" />
        </div>
      </div>
    </form>
  </BaseContainer>
</template>
