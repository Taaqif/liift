<script setup lang="ts">
import { ref } from "vue";
import WorkoutCalendar from "../components/WorkoutCalendar.vue";

const selectedDate = ref<Date>(new Date());

const handleDateChange = (date: Date): void => {
  selectedDate.value = date;
  console.log("Selected date:", date.toLocaleDateString());
};
</script>

<template>
  <div class="p-8 max-w-[800px] mx-auto">
    <WorkoutCalendar
      :selected-date="selectedDate"
      @update:selected-date="handleDateChange"
    />
    <div class="mt-8 p-4 bg-gray-50 rounded-lg border border-gray-200">
      <p class="m-0 text-gray-700 text-base">
        Selected Date:
        {{
          selectedDate.toLocaleDateString("en-US", {
            weekday: "long",
            year: "numeric",
            month: "long",
            day: "numeric",
          })
        }}
      </p>
    </div>
  </div>
</template>
