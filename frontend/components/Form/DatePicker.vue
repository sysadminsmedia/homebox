<template>
  <div v-if="!inline" class="flex w-full flex-col">
    <Label class="cursor-pointer"> {{ label }} </Label>
    <VueDatePicker v-model="selected" :enable-time-picker="false" clearable :dark="isDark" :format="formatDate" />
  </div>
  <div v-else class="sm:flex sm:items-start sm:gap-4">
    <Label class="flex w-full cursor-pointer px-1 py-2"> {{ label }} </Label>
    <VueDatePicker v-model="selected" :enable-time-picker="false" clearable :dark="isDark" :format="formatDate" />
  </div>
</template>

<script setup lang="ts">
  import VueDatePicker from "@vuepic/vue-datepicker";
  import "@vuepic/vue-datepicker/dist/main.css";
  import * as datelib from "~/lib/datelib/datelib";
  import { Label } from "@/components/ui/label";
  import { darkThemes } from "~/lib/data/themes";

  const emit = defineEmits(["update:modelValue", "update:text"]);

  const props = defineProps({
    modelValue: {
      type: Date as () => Date | string | null,
      required: false,
      default: null,
    },
    inline: {
      type: Boolean,
      default: false,
    },
    label: {
      type: String,
      default: "Date",
    },
  });

  const isDark = useIsThemeInList(darkThemes);

  const formatDate = (date: Date | string | number) => fmtDate(date, "human", "date");

  const selected = computed<Date | null>({
    get() {
      // String
      if (props.modelValue === null) {
        return null;
      }

      let date: Date;
      if (typeof props.modelValue === "string") {
        // Empty string
        if (props.modelValue === "") {
          return null;
        }
        date = new Date(props.modelValue);
      } else {
        date = props.modelValue;
      }

      // quick and dirty check for invalid dates
      if (isNaN(date.getTime()) || date.getFullYear() < 1000) {
        return null;
      }

      // Convert UTC date to local date for display
      return new Date(
        date.getUTCFullYear(),
        date.getUTCMonth(),
        date.getUTCDate()
      );
    },
    set(value: Date | null) {
      if (value instanceof Date) {
        // Convert to UTC date without time components
        const utcDate = datelib.zeroTime(value);
        emit("update:modelValue", utcDate);
      } else {
        emit("update:modelValue", null);
      }
    },
  });
</script>

<style class="scoped">
  ::-webkit-calendar-picker-indicator {
    filter: invert(1);
  }
</style>
