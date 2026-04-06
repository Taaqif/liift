import { z } from "zod";
import { i18n } from "@/i18n";

const t = i18n.global.t;

export const planDaySchema = z.object({
  isRest: z.boolean(),
  workoutIds: z.array(z.number().int().positive()),
  description: z.string().optional(),
});

export const planWeekSchema = z.object({
  days: z.array(planDaySchema),
});

export const workoutPlanFormSchema = z.object({
  name: z.string().min(1, t("workoutPlans.validation.nameRequired")),
  description: z.string().optional(),
  numberOfWeeks: z.coerce.number().int().min(1).max(52),
  daysPerWeek: z.coerce.number().int().min(1).max(14),
  weeks: z.array(planWeekSchema),
});

export type PlanDay = z.infer<typeof planDaySchema>;
export type PlanWeek = z.infer<typeof planWeekSchema>;
export type WorkoutPlanFormValues = z.infer<typeof workoutPlanFormSchema>;

export type WorkoutPlan = WorkoutPlanFormValues & {
  id: number;
  created_at?: string;
  updated_at?: string;
};

export type WorkoutPlanProgress = {
  id: number;
  user_id: number;
  plan_id: number;
  plan: WorkoutPlan;
  current_week: number;
  current_day: number;
  started_at: string;
  completed_at: string | null;
};

export function createEmptyDay(): PlanDay {
  return { isRest: false, workoutIds: [] };
}

export function createEmptyWeek(daysPerWeek: number): PlanWeek {
  return {
    days: Array.from({ length: daysPerWeek }, () => createEmptyDay()),
  };
}

export function createEmptyPlan(
  numberOfWeeks: number,
  daysPerWeek: number,
): WorkoutPlanFormValues["weeks"] {
  return Array.from({ length: numberOfWeeks }, () =>
    createEmptyWeek(daysPerWeek),
  );
}

export function resizeWeeks(
  weeks: PlanWeek[],
  newNumberOfWeeks: number,
  newDaysPerWeek: number,
): PlanWeek[] {
  const result: PlanWeek[] = [];
  for (let w = 0; w < newNumberOfWeeks; w++) {
    const existing = weeks[w];
    const existingDays = existing?.days ?? [];
    const days: PlanDay[] = [];
    for (let d = 0; d < newDaysPerWeek; d++) {
      const existing = existingDays[d];
      days.push(existing ? { ...createEmptyDay(), ...existing } : createEmptyDay());
    }
    result.push({ days });
  }
  return result;
}
