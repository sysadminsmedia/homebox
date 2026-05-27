<script setup lang="ts">
  import { Label } from "~/components/ui/label";
  import { Button } from "~/components/ui/button";
  import { availableIcons } from "~/lib/icons";

  const props = defineProps({
    modelValue: {
      type: String,
      required: false,
      default: "",
    },
    label: {
      type: String,
      default: "",
    },
  });

  const emits = defineEmits(["update:modelValue"]);

  const id = useId();

  const value = useVModel(props, "modelValue", emits);

  function selectIcon(iconName: string) {
    if (value.value === iconName) {
      value.value = "";
    } else {
      value.value = iconName;
    }
  }
</script>

<template>
  <div class="flex w-full flex-col gap-1.5">
    <Label :for="id" class="flex w-full px-1">
      <span>{{ label }}</span>
    </Label>
    <div class="flex flex-col gap-1">
      <!-- Grid of icon buttons -->
      <div class="grid w-fit grid-cols-6 gap-1">
        <Button
          v-for="icon in availableIcons"
          :key="icon.name"
          size="xl"
          :variant="value === icon.name ? 'default' : 'outline'"
          class="flex size-10 items-center justify-center p-0 text-lg [&_svg]:size-6"
          :aria-label="`Select ${icon.name} icon`"
          @click="selectIcon(icon.name)"
        >
          <component :is="icon.component" />
        </Button>
      </div>
    </div>
  </div>
</template>
