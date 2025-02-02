<script setup lang="ts">
  import { useMagicKeys } from "@vueuse/core";

  import { ref, watch } from "vue";
  import { useI18n } from "vue-i18n";
  import {
    CommandDialog,
    CommandInput,
    CommandList,
    CommandEmpty,
    CommandGroup,
    CommandItem,
    CommandSeparator,
  } from "~/components/ui/command";
  import { Shortcut } from "~/components/ui/shortcut";

  export type QuickMenuAction =
    | { text: string; action: () => void; type: "navigate" }
    | { text: string; action: () => void; shortcut: string; type: "create" };

  const props = defineProps({
    actions: {
      type: Array as PropType<QuickMenuAction[]>,
      required: false,
      default: () => [],
    },
  });

  const open = ref(false);
  const { t } = useI18n();

  const keys = useMagicKeys();
  const CtrlBackquote = keys.control_Backquote;

  function handleOpenChange() {
    open.value = !open.value;
  }

  watch(CtrlBackquote, v => {
    if (v) handleOpenChange();
  });
</script>

<template>
  <CommandDialog :open="open" @update:open="handleOpenChange">
    <CommandInput
      :placeholder="t('global.search')"
      @keydown="
        (e: KeyboardEvent) => {
          const action = props.actions.filter(item => 'shortcut' in item).find(item => item.shortcut === e.key);
          if (action) {
            open = false;
            action.action();
          }
        }
      "
    />
    <CommandList>
      <CommandEmpty>No results found.</CommandEmpty>
      <CommandGroup :heading="t('global.create')">
        <CommandItem
          v-for="(create, i) in props.actions.filter(item => item.type === 'create')"
          :key="`$global.create_${i + 1}`"
          :value="create.text"
          @select="
            () => {
              open = false;
              create.action();
            }
          "
        >
          {{ create.text }}
          <Shortcut v-if="'shortcut' in create" class="ml-auto" size="sm" :keys="[create.shortcut]" />
        </CommandItem>
      </CommandGroup>
      <CommandSeparator />
      <CommandGroup :heading="t('global.navigate')">
        <CommandItem
          v-for="(navigate, i) in props.actions.filter(item => item.type === 'navigate')"
          :key="navigate.text"
          :value="`global.navigate_${i + 1}`"
          @select="
            () => {
              open = false;
              navigate.action();
            }
          "
        >
          {{ navigate.text }}
        </CommandItem>
      </CommandGroup>
    </CommandList>
  </CommandDialog>
</template>
