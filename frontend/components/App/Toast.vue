<template>
  <div class="fixed right-2 top-2 z-[9999] w-[300px]">
    <TransitionGroup name="notify" tag="div">
      <div
        v-for="(notify, index) in notifications.slice(0, 4)"
        :key="notify.id"
        class="my-2 w-[300px] rounded-md p-3 text-sm text-white"
        :class="{
          'bg-primary': notify.type === 'info',
          'bg-red-600': notify.type === 'error',
          'bg-green-600': notify.type === 'success',
        }"
        @click="dropNotification(index)"
      >
        <div class="flex gap-1">
          <template v-if="notify.type == 'success'">
            <MdiCheckboxMarkedCircle class="size-5" />
          </template>
          <template v-if="notify.type == 'info'">
            <MdiInformationSlabCircle class="size-5" />
          </template>

          <template v-if="notify.type == 'error'">
            <MdiAlert class="size-5" />
          </template>
          {{ notify.message }}
        </div>
      </div>
    </TransitionGroup>
  </div>
</template>

<script setup lang="ts">
  import MdiCheckboxMarkedCircle from "~icons/mdi/checkbox-marked-circle";
  import MdiInformationSlabCircle from "~icons/mdi/information-slab-circle";
  import MdiAlert from "~icons/mdi/alert";

  import { useNotifications } from "@/composables/use-notifier";

  const { notifications, dropNotification } = useNotifications();
</script>

<style scoped>
  .notify-move,
  .notify-enter-active,
  .notify-leave-active {
    transition: all 0.5s ease;
  }
  .notify-enter-from,
  .notify-leave-to {
    opacity: 0;
    transform: translateY(-30px);
  }
  .notify-leave-active {
    position: absolute;
    transform: translateY(30px);
  }
</style>
