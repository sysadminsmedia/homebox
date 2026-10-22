<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { toast } from "@/components/ui/sonner";
  import MdiPrinter from "~icons/mdi/printer";
  import MdiChevronDown from "~icons/mdi/chevron-down";
  import MdiLoading from "~icons/mdi/loading";
  import { Button, ButtonGroup } from "@/components/ui/button";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
  } from "@/components/ui/dropdown-menu";
  import { useDialog } from "@/components/ui/dialog-provider";
  import { DialogID } from "@/components/ui/dialog-provider/utils";
  import LabelMaker from "~/components/global/LabelMaker.vue";

  const props = withDefaults(
    defineProps<{
      /** What the label is for */
      type: "item" | "location";
      id: string;
      /** Items with an asset id get the legacy asset label instead of the item label */
      assetId?: string;
      variant?: "default" | "outline";
    }>(),
    { assetId: "", variant: "default" }
  );

  const { t } = useI18n();
  const { openDialog } = useDialog();
  const { isQuickPrintAvailable, quickPrintItems, quickPrintLocations, defaultPrinter, defaultTemplateId } =
    useQuickPrint();
  const { templates } = useLabelTemplates();
  const { render, renderLocations } = useLabelTemplateActions();

  const busy = ref(false);

  const defaultTemplate = computed(() => templates.value?.find(tpl => tpl.id === defaultTemplateId.value) ?? null);

  const legacyType = computed(() => (props.type === "item" && props.assetId ? "asset" : props.type));
  const legacyId = computed(() => (legacyType.value === "asset" ? props.assetId : props.id));

  function openChooser() {
    openDialog(DialogID.PrintLabelTemplate, {
      params: props.type === "item" ? { itemIds: [props.id] } : { locationIds: [props.id] },
    });
  }

  async function quickPrint() {
    if (!isQuickPrintAvailable.value) {
      openChooser();
      return;
    }
    busy.value = true;
    try {
      const ok = props.type === "item" ? await quickPrintItems([props.id]) : await quickPrintLocations([props.id]);
      if (ok) {
        toast.success(t("components.label_template.print_button.sent", { printer: defaultPrinter.value?.name ?? "" }));
      } else {
        openChooser();
      }
    } catch {
      toast.error(t("components.label_template.toast.print_failed"));
    } finally {
      busy.value = false;
    }
  }

  async function downloadPng() {
    if (!defaultTemplateId.value) {
      openChooser();
      return;
    }
    busy.value = true;
    try {
      const blob =
        props.type === "item"
          ? await render(defaultTemplateId.value, [props.id])
          : await renderLocations(defaultTemplateId.value, [props.id]);
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `label-${props.id}.png`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      toast.success(t("components.label_template.toast.downloaded"));
    } catch {
      toast.error(t("components.label_template.toast.render_failed"));
    } finally {
      busy.value = false;
    }
  }
</script>

<template>
  <div class="flex items-center">
    <!-- Legacy label maker dialog (the built-in QR card); its buttons are replaced by this control -->
    <LabelMaker :id="legacyId" :type="legacyType" dialog-only />

    <ButtonGroup>
      <Button
        :variant="variant"
        class="w-9 md:w-auto"
        :disabled="busy"
        :aria-label="$t('components.label_template.print_button.print_label')"
        @click="quickPrint"
      >
        <MdiLoading v-if="busy" class="animate-spin" />
        <MdiPrinter v-else />
        <span class="hidden md:inline">{{ $t("components.label_template.print_button.print_label") }}</span>
      </Button>
      <DropdownMenu>
        <DropdownMenuTrigger as-child>
          <Button
            :variant="variant"
            size="icon"
            class="w-8 border-l border-l-white/30"
            :aria-label="$t('components.label_template.print_button.more')"
          >
            <MdiChevronDown class="size-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" class="w-80">
          <template v-if="isQuickPrintAvailable && defaultTemplate && defaultPrinter">
            <DropdownMenuItem class="flex-col items-start gap-0.5" @click="quickPrint">
              <span class="font-medium">
                {{ $t("components.label_template.print_button.quick_with", { template: defaultTemplate.name }) }}
              </span>
              <span class="text-xs text-muted-foreground">
                {{ $t("components.label_template.print_button.quick_hint", { printer: defaultPrinter.name }) }}
              </span>
            </DropdownMenuItem>
          </template>
          <DropdownMenuLabel v-else class="font-normal text-muted-foreground">
            {{ $t("components.label_template.print_button.no_defaults") }}
          </DropdownMenuLabel>
          <DropdownMenuItem @click="openChooser">
            {{ $t("components.label_template.print_button.choose") }}
          </DropdownMenuItem>
          <DropdownMenuItem @click="downloadPng">
            {{ $t("components.label_template.print_button.download_png") }}
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem class="justify-between" @click="openDialog(DialogID.PrintLabel)">
            {{ $t("components.label_template.print_button.basic_label") }}
            <span class="text-xs text-muted-foreground">
              {{ $t("components.label_template.print_button.basic_label_hint") }}
            </span>
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem as-child>
            <NuxtLink to="/collection/printers" class="text-muted-foreground">
              {{ $t("components.label_template.print_button.change_defaults") }}
            </NuxtLink>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </ButtonGroup>
  </div>
</template>
