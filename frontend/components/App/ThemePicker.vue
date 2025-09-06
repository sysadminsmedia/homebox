<script setup lang="ts">
  import { themes } from "~~/lib/data/themes";
  import { useTheme } from "~/composables/use-theme";

  const { setTheme } = useTheme();
</script>

<template>
  <div class="homebox grid grid-cols-1 gap-4 font-sans sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
    <div
      v-for="theme in themes"
      :key="theme.value"
      :class="'theme-' + theme.value"
      class="overflow-hidden rounded-lg border outline-2 outline-offset-2"
      :data-theme="theme.value"
      :data-set-theme="theme.value"
      data-act-class="outline"
      @click="setTheme(theme.value)"
    >
      <div :data-theme="theme.value" class="w-full cursor-pointer bg-background-accent text-foreground">
        <div class="grid grid-cols-5 grid-rows-3">
          <div class="col-start-1 row-start-1 bg-background" />
          <div class="col-start-1 row-start-2 bg-sidebar" />
          <div class="col-start-1 row-start-3 bg-background-accent" />
          <div class="col-span-4 col-start-2 row-span-3 row-start-1 flex flex-col gap-1 bg-background p-2">
            <div class="font-bold">{{ theme.label }}</div>
            <div class="flex flex-wrap gap-1">
              <div class="flex size-5 items-center justify-center rounded bg-primary lg:size-6">
                <div class="text-sm font-bold text-primary-foreground">A</div>
              </div>
              <div class="flex size-5 items-center justify-center rounded bg-secondary lg:size-6">
                <div class="text-sm font-bold text-secondary-foreground">A</div>
              </div>
              <div class="flex size-5 items-center justify-center rounded bg-accent lg:size-6">
                <div class="text-sm font-bold text-accent-foreground">A</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped></style>
