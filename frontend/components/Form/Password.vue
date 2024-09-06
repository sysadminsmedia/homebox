<template>
  <div class="relative">
    <FormTextField v-model="value" placeholder="Password" :label="label" :type="inputType"> </FormTextField>
    <button
      type="button"
      class="tooltip absolute right-3 top-11 mb-3 ml-1 mt-auto inline-flex justify-center p-1"
      data-tip="Toggle Password Show"
      @click="toggle()"
    >
      <MdiEye name="mdi-eye" class="size-5" />
    </button>
  </div>
</template>

<script setup lang="ts">
  import MdiEye from "~icons/mdi/eye";

  type Props = {
    modelValue: string;
    placeholder?: string;
    label: string;
  };

  const props = withDefaults(defineProps<Props>(), {
    placeholder: "Password",
    label: "Password",
  });

  const [hide, toggle] = useToggle(true);

  const inputType = computed(() => {
    return hide.value ? "password" : "text";
  });

  const value = useVModel(props, "modelValue");
</script>
