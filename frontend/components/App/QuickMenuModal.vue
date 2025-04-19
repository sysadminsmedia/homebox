<script setup lang="ts">
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
  import { useDialog, useDialogHotkey } from "~/components/ui/dialog-provider";

  export type QuickMenuAction =
    | { text: string; href: string; type: "navigate" }
    | { text: string; dialogId: string; shortcut: string; type: "create" };

  const props = defineProps({
    actions: {
      type: Array as PropType<QuickMenuAction[]>,
      required: false,
      default: () => [],
    },
  });

  const { t } = useI18n();
  const { closeDialog, openDialog } = useDialog();

  useDialogHotkey("quick-menu", { code: "Backquote", ctrl: true });
</script>

<template>
  <CommandDialog dialog-id="quick-menu">
    <CommandInput
      :placeholder="t('components.quick_menu.shortcut_hint')"
      @keydown="
        (e: KeyboardEvent) => {
          const item = props.actions.filter(item => 'shortcut' in item).find(item => item.shortcut === e.key);
          if (item) {
            e.preventDefault();
            openDialog(item.dialogId);
          }
          // if esc is pressed, close the dialog
          if (e.key === 'Escape') {
            e.preventDefault();
            closeDialog('quick-menu');
          }
        }
      "
    />
    <CommandList>
      <CommandSeparator />
      <CommandEmpty>{{ t("components.quick_menu.no_results") }}</CommandEmpty>
      <CommandGroup :heading="t('global.create')">
        <CommandItem
          v-for="(create, i) in props.actions.filter(item => item.type === 'create')"
          :key="`$global.create_${i + 1}`"
          :value="create.text"
          @select="
            e => {
              e.preventDefault();
              openDialog(create.dialogId);
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
              closeDialog('quick-menu');
              navigateTo(navigate.href);
            }
          "
        >
          {{ navigate.text }}
        </CommandItem>
      </CommandGroup>
    </CommandList>
  </CommandDialog>
</template>
