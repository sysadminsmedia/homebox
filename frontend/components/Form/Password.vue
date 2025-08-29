<template>
  <div class="relative">
    <FormTextField v-model="value" :placeholder="localizedPlaceholder" :label="localizedLabel" :type="inputType" />
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
        <TooltipContent>{{ $t("components.form.password.toggle_show") }}</TooltipContent>
      </Tooltip>
    </TooltipProvider>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import MdiEye from "~icons/mdi/eye";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
  import FormTextField from "@/components/Form/TextField.vue";

  const { t } = useI18n();
  type Props = {
    modelValue: string;
    placeholder?: string;
    label: string;
  };

  const props = defineProps<Props>();

  const [hide, toggle] = useToggle(true);

  const localizedPlaceholder = computed(() => props.placeholder ?? t("global.password"));
  const localizedLabel = computed(() => props.label ?? t("global.password"));

  const inputType = computed(() => {
    return hide.value ? "password" : "text";
  });

  const value = useVModel(props, "modelValue");
</script>
