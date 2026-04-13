<script setup lang="ts">
import {
  CalendarRoot,
  CalendarHeader,
  CalendarHeading,
  CalendarNext,
  CalendarPrev,
  CalendarGrid,
  CalendarGridHead,
  CalendarGridRow,
  CalendarHeadCell,
  CalendarGridBody,
  CalendarCell,
  CalendarCellTrigger,
} from "reka-ui";
import { ChevronLeft, ChevronRight } from "lucide-vue-next";
import type { AnyCalendarDate } from "@internationalized/date";
import { cn } from "@/lib/utils";

interface Props {
  class?: string;
  activityDates?: Set<string>;
  locale?: string;
  weekStartsOn?: 0 | 1 | 2 | 3 | 4 | 5 | 6;
  weekdayFormat?: "narrow" | "short" | "long";
  fixedWeeks?: boolean;
  numberOfMonths?: number;
  disabled?: boolean;
  readonly?: boolean;
  pagedNavigation?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  activityDates: () => new Set(),
});

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const model = defineModel<any>();
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const placeholder = defineModel<any>("placeholder");

function dateKey(date: AnyCalendarDate): string {
  return `${date.year}-${String(date.month).padStart(2, "0")}-${String(date.day).padStart(2, "0")}`;
}
</script>

<template>
  <CalendarRoot
    v-model="model"
    v-model:placeholder="placeholder"
    :locale="props.locale"
    :week-starts-on="props.weekStartsOn"
    :weekday-format="props.weekdayFormat"
    :fixed-weeks="props.fixedWeeks"
    :number-of-months="props.numberOfMonths"
    :disabled="props.disabled"
    :readonly="props.readonly"
    :paged-navigation="props.pagedNavigation"
    :class="cn('w-full select-none p-4', props.class)"
    v-slot="{ grid, weekDays }"
  >
    <!-- Header: prev / month-year / next -->
    <CalendarHeader class="flex items-center justify-between mb-4">
      <CalendarPrev
        class="inline-flex items-center justify-center rounded-lg w-9 h-9 text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors disabled:opacity-30 disabled:pointer-events-none"
      >
        <ChevronLeft class="w-4 h-4" />
      </CalendarPrev>
      <CalendarHeading class="text-sm font-semibold" />
      <CalendarNext
        class="inline-flex items-center justify-center rounded-lg w-9 h-9 text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors disabled:opacity-30 disabled:pointer-events-none"
      >
        <ChevronRight class="w-4 h-4" />
      </CalendarNext>
    </CalendarHeader>

    <div v-for="month in grid" :key="month.value.toString()">
      <CalendarGrid class="w-full table-fixed border-collapse">
        <!-- Day-of-week headings -->
        <CalendarGridHead>
          <CalendarGridRow class="flex w-full">
            <CalendarHeadCell
              v-for="day in weekDays"
              :key="day"
              class="flex-1 text-[0.75rem] font-medium text-muted-foreground text-center pb-2"
            >
              {{ day }}
            </CalendarHeadCell>
          </CalendarGridRow>
        </CalendarGridHead>

        <!-- Weeks -->
        <CalendarGridBody>
          <CalendarGridRow
            v-for="(weekDates, index) in month.rows"
            :key="index"
            class="flex w-full mt-1"
          >
            <CalendarCell
              v-for="weekDate in weekDates"
              :key="weekDate.toString()"
              :date="weekDate"
              class="flex-1 p-0 text-center"
            >
              <CalendarCellTrigger
                :day="weekDate"
                :month="month.value"
                :class="cn(
                  'group w-full py-2 flex flex-col items-center justify-center gap-1',
                  'rounded-lg text-sm font-normal transition-colors',
                  'hover:bg-accent hover:text-accent-foreground',
                  'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
                  'data-[selected]:bg-primary data-[selected]:text-primary-foreground data-[selected]:hover:bg-primary/90',
                  'data-[today]:ring-1 data-[today]:ring-primary data-[today]:font-semibold',
                  'data-[outside-view]:text-muted-foreground/30 data-[outside-view]:pointer-events-none',
                  'data-[disabled]:pointer-events-none data-[disabled]:opacity-40',
                )"
              >
                <span class="leading-none tabular-nums">{{ weekDate.day }}</span>
                <!-- Activity dot -->
                <span
                  :class="[
                    'block w-1 h-1 rounded-full transition-colors',
                    activityDates.has(dateKey(weekDate))
                      ? 'bg-primary group-data-[selected]:bg-primary-foreground group-data-[outside-view]:opacity-30'
                      : 'bg-transparent',
                  ]"
                  aria-hidden="true"
                />
              </CalendarCellTrigger>
            </CalendarCell>
          </CalendarGridRow>
        </CalendarGridBody>
      </CalendarGrid>
    </div>
  </CalendarRoot>
</template>
