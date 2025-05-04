<template>
  <div class="relative">
    <FormTextField v-model="value" placeholder="Password" :label="label" :type="inputType"> </FormTextField>
    <TooltipProvider :delay-duration="0">
      <Tooltip>
        <TooltipTrigger as-child>
          <button
            type="button"
            class="absolute right-3 top-6 mb-3 ml-1 mt-auto inline-flex justify-center p-1"
            @click="toggle()"
          >
            <MdiEye name="mdi-eye" class="size-5" />
          </button>
        </TooltipTrigger>
        <TooltipContent>Toggle Password Show</TooltipContent>
      </Tooltip>
    </TooltipProvider>
  </div>
</template>

<script setup lang="ts">
  import MdiEye from "~icons/mdi/eye";
  import { Tooltip, TooltipContent, TooltipTrigger, TooltipProvider } from "@/components/ui/tooltip";

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
