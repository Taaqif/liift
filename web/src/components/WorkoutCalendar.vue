<script setup lang="ts">
import { ChevronLeftIcon, ChevronRightIcon } from "lucide-vue-next";
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { Button } from "./ui/button";
import { Card, CardContent } from "./ui/card";

const { t, d } = useI18n();

interface Props {
  selectedDate: Date;
}

interface Emits {
  (e: "update:selectedDate", date: Date): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

const today = new Date();
today.setHours(0, 0, 0, 0);

// Get the Monday of the week for a given date
const getMonday = (date: Date): Date => {
  const dateCopy = new Date(date);
  dateCopy.setHours(0, 0, 0, 0);
  const day = dateCopy.getDay();
  const diff = dateCopy.getDate() - day + (day === 0 ? -6 : 1); // Adjust when day is Sunday
  return new Date(dateCopy.setDate(diff));
};

// Get all days in the week starting from Monday
const weekDays = computed(() => {
  const monday = getMonday(props.selectedDate);
  const days: Date[] = [];

  for (let i = 0; i < 7; i++) {
    const day = new Date(monday);
    day.setDate(monday.getDate() + i);
    days.push(day);
  }

  return days;
});

// Check if a date is today
const isToday = (date: Date): boolean => {
  return (
    date.getDate() === today.getDate() &&
    date.getMonth() === today.getMonth() &&
    date.getFullYear() === today.getFullYear()
  );
};

// Check if a date is selected
const isSelected = (date: Date): boolean => {
  return (
    date.getDate() === props.selectedDate.getDate() &&
    date.getMonth() === props.selectedDate.getMonth() &&
    date.getFullYear() === props.selectedDate.getFullYear()
  );
};

// Navigate to previous week
const goToPreviousWeek = (): void => {
  const monday = getMonday(props.selectedDate);
  monday.setDate(monday.getDate() - 7);
  emit("update:selectedDate", monday);
};

// Navigate to next week
const goToNextWeek = (): void => {
  const monday = getMonday(props.selectedDate);
  monday.setDate(monday.getDate() + 7);
  emit("update:selectedDate", monday);
};

// Select a date
const selectDate = (date: Date): void => {
  emit("update:selectedDate", new Date(date));
};

// Format date for display
const formatDate = (date: Date): string => {
  return date.getDate().toString();
};

// Get day name abbreviation using i18n
const getDayName = (date: Date): string => {
  const dayIndex = date.getDay() === 0 ? 6 : date.getDay() - 1;
  const dayKeys: string[] = [
    "calendar.dayNames.monday",
    "calendar.dayNames.tuesday",
    "calendar.dayNames.wednesday",
    "calendar.dayNames.thursday",
    "calendar.dayNames.friday",
    "calendar.dayNames.saturday",
    "calendar.dayNames.sunday",
  ];
  const dayKey = dayKeys[dayIndex];
  if (!dayKey) {
    return "";
  }
  return t(dayKey);
};
</script>

<template>
  <Card>
    <CardContent>
      <div class="flex items-center justify-between mb-6">
        <Button
          @click="goToPreviousWeek"
          variant="secondary"
          size="icon"
          :aria-label="t('calendar.previousWeek')"
        >
          <ChevronLeftIcon />
        </Button>
        <div class="font-semibold text-gray-900 text-base">
          <template v-if="weekDays[0] && weekDays[6]">
            {{ d(weekDays[0], "monthDay") }} -
            {{ d(weekDays[6], "monthDayYear") }}
          </template>
        </div>
        <Button
          @click="goToNextWeek"
          variant="secondary"
          size="icon"
          :aria-label="t('calendar.nextWeek')"
        >
          <ChevronRightIcon />
        </Button>
      </div>
      <div class="grid grid-cols-7 gap-2">
        <Button
          v-for="(day, index) in weekDays"
          :key="index"
          variant="outline"
          class="h-auto flex flex-col items-center justify-center"
          :class="{
            'border-primary bg-secondary': isSelected(day),
            'border-primary/30': isToday(day) && !isSelected(day),
          }"
          @click="selectDate(day)"
        >
          <div class="text-xs text-gray-500 mb-1 uppercase tracking-wider">
            {{ getDayName(day) }}
          </div>
          <div class="text-lg font-medium">
            {{ formatDate(day) }}
          </div>
        </Button>
      </div>
    </CardContent>
  </Card>
</template>
