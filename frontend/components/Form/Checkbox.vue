<template>
  <div v-if="!inline" class="flex w-full items-center gap-1.5">
    <Checkbox :id="id" v-model="value" class="size-6" />
    <Label :for="id" class="cursor-pointer">
      {{ tag }}
    </Label>
  </div>
  <div v-else class="sm:grid sm:grid-cols-4 sm:items-start sm:gap-4">
    <Label :for="id" class="flex w-full cursor-pointer px-1 py-2">
      {{ tag }}
    </Label>
    <Checkbox :id="id" v-model="value" class="size-6" />
  </div>
</template>

<script setup lang="ts">
  import { Label } from "@/components/ui/tag";
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
    tag: {
      type: String,
      default: "",
    },
  });

  const value = useVModel(props, "modelValue");

  const id = useId();
</script>
