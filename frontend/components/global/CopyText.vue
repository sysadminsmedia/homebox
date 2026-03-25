<template>
  <TooltipProvider :delay-duration="0">
    <Tooltip>
      <TooltipTrigger as-child>
        <Button size="icon" variant="outline" class="relative" @click="copyText">
          <div
            :data-copied="copied"
            class="group absolute inset-0 flex items-center justify-center transition-transform duration-300 data-[copied=true]:rotate-[360deg]"
          >
            <MdiContentCopy
              class="group-data-[copied=true]:hidden"
              :style="{
                height: `${iconSize}px`,
                width: `${iconSize}px`,
              }"
            />
            <MdiClipboard
              class="hidden group-data-[copied=true]:block"
              :style="{
                height: `${iconSize}px`,
                width: `${iconSize}px`,
              }"
            />
          </div>
        </Button>
      </TooltipTrigger>
      <TooltipContent v-if="tooltip">
        {{ tooltip }}
      </TooltipContent>
    </Tooltip>
  </TooltipProvider>

  <AlertDialog v-model:open="copyError">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle class="space-y-2">
          {{ $t("components.global.copy_text.failed_to_copy") }}
          {{ isNotHttps ? $t("components.global.copy_text.https_required") : "" }}
        </AlertDialogTitle>
        <AlertDialogDescription class="text-sm">
          {{ $t("components.global.copy_text.learn_more") }}
          <a
            href="https://homebox.software/en/user-guide/tips-tricks.html#copy-to-clipboard"
            class="text-primary hover:underline"
            target="_blank"
            rel="noopener"
          >
            {{ $t("components.global.copy_text.documentation") }}
          </a>
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogAction>{{ $t("components.global.copy_text.continue") }}</AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>

<script setup lang="ts">
  import MdiContentCopy from "~icons/mdi/content-copy";
  import MdiClipboard from "~icons/mdi/clipboard";
  import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@/components/ui/alert-dialog";
  import { Button } from "@/components/ui/button";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

  const props = defineProps({
    text: {
      type: String as () => string,
      default: "",
    },
    iconSize: {
      type: Number as () => number,
      default: 20,
    },
    tooltip: {
      type: String as () => string,
      default: "",
    },
  });

  const { copy, copied } = useClipboard({ source: props.text, copiedDuring: 1000 });
  const copyError = ref(false);
  const isNotHttps = window.location.protocol !== "https:";

  async function copyText() {
    await copy(props.text);
    if (!copied.value) {
      console.error(`Failed to copy to clipboard${isNotHttps ? " likely because protocol is not https" : ""}`);
      copyError.value = true;
    }
  }
</script>
