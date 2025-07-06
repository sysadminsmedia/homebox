<template>
  <div v-if="!inline" class="flex w-full items-center gap-1.5">
    <Checkbox :id="id" v-model="value" class="size-6" />
    <Label :for="id" class="cursor-pointer">
      {{ label }}
    </Label>
  </div>
  <div v-else class="sm:grid sm:grid-cols-4 sm:items-start sm:gap-4">
    <Label :for="id" class="flex w-full cursor-pointer px-1 py-2">
      {{ label }}
    </Label>
    <Checkbox :id="id" v-model="value" class="size-6" />
  </div>
</template>

<script setup lang="ts">
  import { Label } from "@/components/ui/label";
  import { Checkbox } from "@/components/ui/checkbox";

  const props = defineProps({
    modelValue: {
      type: Boolean,
      default: false,
    },
    inline: {
      type: Boolean,
      default: false,
    },
    label: {
      type: String,
      default: "",
    },
  });

  const value = useVModel(props, "modelValue");

  const id = useId();
</script>
