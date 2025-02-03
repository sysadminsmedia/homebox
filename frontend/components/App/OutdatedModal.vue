<template>
  <AlertDialog v-model:open="open">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>🎉 {{ $t("components.app.outdated.new_version_available") }} 🎉</AlertDialogTitle>
        <AlertDialogDescription>
          <p>{{ $t("components.app.outdated.current_version") }}: {{ current }}</p>
          <p>{{ $t("components.app.outdated.latest_version") }}: {{ latest }}</p>
          <p>
            <a href="https://github.com/sysadminsmedia/homebox/releases" target="_blank" rel="noopener" class="link">
              {{ $t("components.app.outdated.new_version_available_link") }}
            </a>
          </p>
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogAction @click="hide">{{ $t("components.app.outdated.dismiss") }}</AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>

<script setup lang="ts">
  import { lt } from "semver";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogAction,
  } from "~/components/ui/alert-dialog";

  const props = defineProps<{
    status: {
      build: {
        version: string;
      };
      latest: {
        version: string;
      };
    };
  }>();

  const latest = computed(() => props.status.latest.version);
  const current = computed(() => props.status.build.version);

  const isDev = computed(() => import.meta.dev || !current.value?.includes("."));
  const isOutdated = computed(() => current.value && latest.value && lt(current.value, latest.value));
  const hasHiddenLatest = computed(() => localStorage.getItem("latestVersion") === latest.value);

  const displayOutdatedWarning = computed(() => Boolean(!isDev.value && !hasHiddenLatest.value && isOutdated.value));

  const open = ref(false);

  watch(
    displayOutdatedWarning,
    displayOutdatedWarning => {
      console.log("displayOutdatedWarning", displayOutdatedWarning);
      if (displayOutdatedWarning) {
        open.value = true;
      }
    },
    { immediate: true }
  );

  const hide = () => {
    open.value = false;
    localStorage.setItem("latestVersion", latest.value);
  };
</script>
