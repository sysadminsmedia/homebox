<template>
  <div v-if="!inline" v-bind="wrapperAttrs()" class="flex w-full flex-col gap-1.5" :class="wrapperClass()">
    <Label :for="fieldId" class="flex w-full px-1">
      <span> {{ label }} </span>
      <span class="grow" />
      <span
        :class="{
          'text-destructive': hasLengthError,
        }"
      >
        {{ characterCountText }}
      </span>
    </Label>
    <Input
      v-bind="inputAttrs()"
      :id="fieldId"
      ref="input"
      v-model="value"
      :name="name"
      :placeholder="placeholder"
      :type="type"
      :autocomplete="autocomplete"
      :min="min"
      :max="max"
      :step="step"
      :minlength="minLength !== -1 ? minLength : undefined"
      :maxlength="maxLength !== -1 ? maxLength : undefined"
      :passwordrules="passwordrules"
      :required="required"
      class="w-full"
      @input="markDirty"
    />
  </div>
  <div v-else v-bind="wrapperAttrs()" class="sm:grid sm:grid-cols-4 sm:items-start sm:gap-4" :class="wrapperClass()">
    <Label class="flex w-full px-1 py-2" :for="fieldId">
      <span> {{ label }} </span>
      <span class="grow" />
      <span
        :class="{
          'text-destructive': hasLengthError,
        }"
      >
        {{ characterCountText }}
      </span>
    </Label>
    <Input
      v-bind="inputAttrs()"
      :id="fieldId"
      ref="input"
      v-model="value"
      :name="name"
      :placeholder="placeholder"
      :type="type"
      :autocomplete="autocomplete"
      :min="min"
      :max="max"
      :step="step"
      :minlength="minLength !== -1 ? minLength : undefined"
      :maxlength="maxLength !== -1 ? maxLength : undefined"
      :passwordrules="passwordrules"
      :required="required"
      class="col-span-3 mt-2 w-full"
      @input="markDirty"
    />
  </div>
</template>

<script lang="ts" setup>
  import { useI18n } from "vue-i18n";
  import { utf8Length } from "@/lib/utils";
  import { Label } from "~/components/ui/label";
  import { Input } from "~/components/ui/input";

  defineOptions({
    inheritAttrs: false,
  });

  const attrs = useAttrs();

  const props = defineProps({
    id: {
      type: String,
      default: undefined,
    },
    label: {
      type: String,
      default: "",
    },
    modelValue: {
      type: [String, Number],
      default: null,
    },
    required: {
      type: [Boolean],
      default: null,
    },
    type: {
      type: String,
      default: "text",
    },
    name: {
      type: String,
      default: undefined,
    },
    autocomplete: {
      type: String,
      default: undefined,
    },
    passwordrules: {
      type: String,
      default: undefined,
    },
    triggerFocus: {
      type: Boolean,
      default: null,
    },
    inline: {
      type: Boolean,
      default: false,
    },
    placeholder: {
      type: String,
      default: "",
    },
    maxLength: {
      type: Number,
      default: -1,
      required: false,
    },
    minLength: {
      type: Number,
      default: -1,
      required: false,
    },
    min: {
      type: [String, Number],
      default: undefined,
    },
    max: {
      type: [String, Number],
      default: undefined,
    },
    step: {
      type: [String, Number],
      default: undefined,
    },
  });

  const { t } = useI18n();

  const generatedId = useId();
  const fieldId = computed(() => props.id ?? generatedId);

  function wrapperClass() {
    return attrs.class;
  }

  function wrapperAttrs() {
    const testId = attrs["data-testid"];
    return testId ? { "data-testid": testId } : {};
  }

  function inputAttrs() {
    const { class: _class, "data-testid": _testId, ...rest } = attrs;
    return rest;
  }

  const input = ref<HTMLElement | null>(null);

  whenever(
    () => props.triggerFocus,
    () => {
      if (input.value) {
        input.value.focus();
      }
    }
  );

  const value = useVModel(props, "modelValue");
  const isDirty = ref(false);

  function markDirty() {
    isDirty.value = true;
  }

  const hasLengthError = computed(() => {
    if (typeof value.value !== "string" || !isDirty.value) {
      return false;
    }

    return (
      (props.maxLength !== -1 && utf8Length(value.value) > props.maxLength) ||
      (props.minLength !== -1 && utf8Length(value.value) < props.minLength)
    );
  });

  const characterCountText = computed(() => {
    if (typeof value.value !== "string") {
      return "";
    }

    if (props.maxLength !== -1) {
      const minText = props.minLength !== -1 ? ` (${t("components.form.min_length", { min: props.minLength })})` : "";
      return `${utf8Length(value.value)}/${props.maxLength}${minText}`;
    }

    if (props.minLength !== -1) {
      return `${utf8Length(value.value)} (${t("components.form.min_length", { min: props.minLength })})`;
    }

    return "";
  });
</script>
