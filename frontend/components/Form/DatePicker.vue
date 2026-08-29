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
  import { parseDateOnly, toDateOnlyString } from "~/lib/datelib/dateOnly";
  import { Label } from "@/components/ui/label";
  import { darkThemes } from "~/lib/data/themes";

  const emit = defineEmits(["update:modelValue", "update:text"]);

  const props = defineProps({
    modelValue: {
      type: [Date, String] as unknown as () => Date | string | null,
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
      if (typeof props.modelValue === "string") {
        // Empty string
        if (props.modelValue === "") {
          return null;
        }

        // Invalid Date string
        if (props.modelValue === "Invalid Date") {
          return null;
        }

        // YYYY-MM-DD is read through local components. `new Date("2026-04-18")`
        // would be UTC midnight, which the calendar then highlights as the
        // 17th for anyone west of Greenwich. Timestamps from older records
        // still fall back to the plain constructor.
        const dateOnly = parseDateOnly(props.modelValue);
        if (dateOnly) {
          return dateOnly;
        }

        const parsed = new Date(props.modelValue);
        return isNaN(parsed.getTime()) ? null : parsed;
      }

      // Date
      if (props.modelValue instanceof Date) {
        if (props.modelValue.getFullYear() < 1000) {
          return null;
        }

        if (isNaN(props.modelValue.getTime())) {
          return null;
        }

        // Valid Date
        return props.modelValue;
      }

      return null;
    },
    set(value: Date | null) {
      // Always a YYYY-MM-DD string built from local components. Emitting a
      // Date instead would let JSON.stringify serialize it as a UTC instant,
      // which lands on the previous day for every user west of Greenwich and
      // shifts again on each subsequent save. Every field this picker drives
      // is a calendar date (types.Date on the backend), so there is no case
      // where a Date object is the right thing to emit — a timestamp field
      // should use VueDatePicker directly.
      emit("update:modelValue", value ? toDateOnlyString(value) : "");
    },
  });
</script>

<style class="scoped">
  ::-webkit-calendar-picker-indicator {
    filter: invert(1);
  }
</style>
