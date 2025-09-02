<script setup lang="ts">
  import { Input } from "@/components/ui/input";
  import { Switch } from "@/components/ui/switch";
  import { Label } from "@/components/ui/label";
  import type { DuplicateSettings } from "~/composables/use-preferences";

  type Props = {
    modelValue: DuplicateSettings;
  };

  type Emits = {
    (e: "update:modelValue", value: DuplicateSettings): void;
  };

  const props = defineProps<Props>();

  const enableCustomPrefix = ref(props.modelValue.copyPrefixOverride !== null);
  const prefix = ref(props.modelValue.copyPrefixOverride ?? "");

  const emit = defineEmits<Emits>();

  const settings = computed({
    get: () => props.modelValue,
    set: value => emit("update:modelValue", value),
  });
</script>

<template>
  <div class="flex flex-col gap-4">
    <div class="flex flex-col gap-3">
      <div class="flex items-center gap-2">
        <Switch id="copy-maintenance" v-model="settings.copyMaintenance" />
        <Label for="copy-maintenance">
          {{ $t("items.duplicate.copy_maintenance") }}
        </Label>
      </div>

      <div class="flex items-center gap-2">
        <Switch id="copy-attachments" v-model="settings.copyAttachments" />
        <Label for="copy-attachments">
          {{ $t("items.duplicate.copy_attachments") }}
        </Label>
      </div>

      <div class="flex items-center gap-2">
        <Switch id="copy-custom-fields" v-model="settings.copyCustomFields" />
        <Label for="copy-custom-fields">
          {{ $t("items.duplicate.copy_custom_fields") }}
        </Label>
      </div>

      <div class="flex items-center gap-2">
        <Switch
          id="copy-prefix"
          v-model="enableCustomPrefix"
          @update:model-value="
            v => {
              settings.copyPrefixOverride = v ? prefix : null;
            }
          "
        />
        <Label for="copy-prefix">{{ $t("items.duplicate.enable_custom_prefix") }}</Label>
      </div>

      <div class="flex flex-col gap-2">
        <Label for="copy-prefix" :class="{ 'opacity-50': !enableCustomPrefix }">
          {{ $t("items.duplicate.custom_prefix") }}
        </Label>
        <Input
          id="copy-prefix"
          v-model="prefix"
          :disabled="!enableCustomPrefix"
          :placeholder="$t('items.duplicate.prefix')"
          class="w-full"
          @input="settings.copyPrefixOverride = prefix"
        />
        <p class="text-sm text-muted-foreground">
          {{ $t("items.duplicate.prefix_instructions") }}
        </p>
      </div>
    </div>
  </div>
</template>
