<template>
  <button @click="copyText">
    <label
      class="swap swap-rotate"
      :class="{
        'swap-active': copied,
      }"
    >
      <MdiContentCopy
        class="swap-off"
        :style="{
          height: `${iconSize}px`,
          width: `${iconSize}px`,
        }"
      />
      <MdiClipboard
        class="swap-on"
        :style="{
          height: `${iconSize}px`,
          width: `${iconSize}px`,
        }"
      />
    </label>
    <Teleport to="#app">
      <BaseModal v-model="copyError">
        <div class="space-y-2">
          <p>
            {{ $t("components.global.copy_text.failed_to_copy") }}
            {{ isNotHttps ? $t("components.global.copy_text.https_required") : "" }}
          </p>
          <p class="text-sm">
            {{ $t("components.global.copy_text.learn_more") }}
            <a
              href="https://homebox.software/en/user-guide/tips-tricks.html#copy-to-clipboard"
              class="text-primary hover:underline"
              target="_blank"
              rel="noopener"
            >
              {{ $t("components.global.copy_text.documentation") }}
            </a>
          </p>
        </div>
      </BaseModal></Teleport
    >
  </button>
</template>

<script setup lang="ts">
  import MdiContentCopy from "~icons/mdi/content-copy";
  import MdiClipboard from "~icons/mdi/clipboard";

  const props = defineProps({
    text: {
      type: String as () => string,
      default: "",
    },
    iconSize: {
      type: Number as () => number,
      default: 20,
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
