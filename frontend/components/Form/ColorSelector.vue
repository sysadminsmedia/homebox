<script setup lang="ts">
  import { computed, onMounted } from "vue";
  import { useI18n } from "vue-i18n";
  import { Label } from "~/components/ui/label";
  import { Button } from "~/components/ui/button";
  import MdiClose from "~icons/mdi/close";
  import MdiDiceMultiple from "~icons/mdi/dice-multiple";

  const { t } = useI18n();

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
    inline: {
      type: Boolean,
      default: false,
    },
    showHex: {
      type: Boolean,
      default: false,
    },
    size: {
      type: [String, Number],
      default: 24,
    },
    startingColor: {
      type: String,
      default: "",
    },
  });

  const emits = defineEmits(["update:modelValue"]);

  const id = useId();

  const swatchStyle = computed(() => ({
    backgroundColor: props.modelValue || "hsl(var(--muted))",
    width: typeof props.size === "number" ? `${props.size}px` : props.size,
    height: typeof props.size === "number" ? `${props.size}px` : props.size,
  }));

  const value = useVModel(props, "modelValue", emits);

  // Initialize with starting color if provided and current value is empty
  onMounted(() => {
    if (props.startingColor && (!value.value || value.value === "")) {
      value.value = props.startingColor;
    }
  });

  function clearColor() {
    value.value = "";
  }

  function randomizeColor() {
    const randomColor =
      "#" +
      Math.floor(Math.random() * 16777215)
        .toString(16)
        .padStart(6, "0");
    value.value = randomColor;
  }
</script>

<template>
  <div v-if="!inline" class="flex w-full flex-col gap-1.5">
    <Label :for="id" class="flex w-full px-1">
      <span>{{ label }}</span>
    </Label>
    <div class="flex items-center gap-2">
      <span
        :style="swatchStyle"
        class="inline-block cursor-pointer rounded-full border ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
        :aria-label="`${t('components.color_selector.color')}: ${modelValue || t('components.color_selector.no_color_selected')}`"
        role="button"
        tabindex="0"
        @click="($refs.colorInput as HTMLInputElement).click()"
      />
      <span v-if="showHex" class="font-mono text-xs text-muted-foreground">{{
        modelValue || t("components.color_selector.no_color")
      }}</span>
      <div class="flex gap-1">
        <Button
          type="button"
          variant="outline"
          size="sm"
          class="size-6 p-0"
          :aria-label="t('components.color_selector.randomize')"
          @click="randomizeColor"
        >
          <MdiDiceMultiple class="size-3" />
        </Button>
        <Button
          type="button"
          variant="outline"
          size="sm"
          class="size-6 p-0"
          :aria-label="t('components.color_selector.clear')"
          @click="clearColor"
        >
          <MdiClose class="size-3" />
        </Button>
      </div>
      <input :id="id" ref="colorInput" v-model="value" type="color" class="sr-only" tabindex="-1" />
    </div>
  </div>
  <div v-else class="sm:grid sm:grid-cols-4 sm:items-start sm:gap-4">
    <Label class="flex w-full px-1 py-2" :for="id">
      <span>{{ label }}</span>
    </Label>
    <div class="col-span-3 mt-2 flex items-center gap-2">
      <span
        :style="swatchStyle"
        class="inline-block cursor-pointer rounded-full border ring-offset-background focus:outline-none focus:outline-primary focus:ring-2 focus:ring-ring focus:ring-offset-2"
        :aria-label="`${t('components.color_selector.color')}: ${modelValue || t('components.color_selector.no_color_selected')}`"
        role="button"
        tabindex="0"
        @click="($refs.colorInput as HTMLInputElement).click()"
      />
      <span v-if="showHex" class="font-mono text-xs text-muted-foreground">{{
        modelValue || t("components.color_selector.no_color")
      }}</span>
      <div class="flex gap-1">
        <Button
          type="button"
          variant="outline"
          size="sm"
          class="size-6 p-0"
          :aria-label="t('components.color_selector.randomize')"
          @click="randomizeColor"
        >
          <MdiDiceMultiple class="size-3" />
        </Button>
        <Button
          type="button"
          variant="outline"
          size="sm"
          class="size-6 p-0"
          :aria-label="t('components.color_selector.clear')"
          @click="clearColor"
        >
          <MdiClose class="size-3" />
        </Button>
      </div>
      <input :id="id" ref="colorInput" v-model="value" type="color" class="sr-only" tabindex="-1" />
    </div>
  </div>
</template>
