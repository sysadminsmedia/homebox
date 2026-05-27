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
  import { toDateOnlyString } from "~/lib/datelib/dateOnly";
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
    // When true, this field represents a date-only value (no time-of-day or
    // timezone semantics). modelValue is bound as a YYYY-MM-DD string and the
    // component emits a YYYY-MM-DD string. Use this for any field stored as
    // types.Date on the backend (purchaseDate, scheduledDate, etc.) — it
    // prevents the timezone day-shift bug that occurs when JSON.stringify
    // converts a Date object to a UTC ISO string.
    //
    // Leave false (the default) for true timestamp fields like invite expiry,
    // which need full Date-object precision.
    dateOnly: {
      type: Boolean,
      default: false,
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

        return datelib.parse(props.modelValue);
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
      if (props.dateOnly) {
        // Always emit YYYY-MM-DD strings, derived from local components, so
        // the user's calendar day is preserved across the API round-trip.
        emit("update:modelValue", value ? toDateOnlyString(value) : "");
        return;
      }

      if (value instanceof Date) {
        value = datelib.zeroTime(value);
        emit("update:modelValue", value);
      } else {
        value = value ? datelib.zeroTime(new Date(value)) : null;
        emit("update:modelValue", value);
      }
    },
  });
</script>

<style class="scoped">
  ::-webkit-calendar-picker-indicator {
    filter: invert(1);
  }
</style>
