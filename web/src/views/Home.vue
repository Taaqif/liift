<script setup lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";
import { useElapsedTimer, formatElapsed } from "@/composables/useElapsedTimer";
import { useActiveWorkoutSession } from "@/features/workout-session/composables/useActiveWorkoutSession";
import { useWorkoutSessions } from "@/features/workout-session/composables/useWorkoutSessions";
import { useActivePlanProgress } from "@/features/workout-plans/composables/useActivePlanProgress";
import { useWeeklyStats } from "@/features/workout-session/composables/useWeeklyStats";
import ActivePlanCard from "@/features/workout-plans/components/ActivePlanCard.vue";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Dumbbell,
  Clock,
  Flame,
  Layers,
  ChevronRight,
  ArrowRight,
  BookOpen,
  CalendarDays,
  BarChart3,
  Target,
  Play,
  TrendingUp,
  CheckCircle2,
} from "lucide-vue-next";

const router = useRouter();

// ── Queries ───────────────────────────────────────────────
const { session: activeSession, loading: activeLoading } = useActiveWorkoutSession();
const { progress: activePlan, loading: planLoading } = useActivePlanProgress();
const { sessions: recentSessions, loading: recentLoading } = useWorkoutSessions(6, 0);

// Weekly stats: Monday → today
const weekBounds = computed(() => {
  const now = new Date();
  const day = now.getDay() || 7;
  const mon = new Date(now);
  mon.setDate(now.getDate() - (day - 1));
  const fmt = (d: Date) =>
    `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
  return { from: fmt(mon), to: fmt(now) };
});

const { sessions: weekSessions } = useWorkoutSessions(
  100,
  0,
  null,
  null,
  computed(() => weekBounds.value.from),
  computed(() => weekBounds.value.to),
);

const { statMap: weekStatMap } = useWeeklyStats(
  computed(() => weekBounds.value.from),
  computed(() => weekBounds.value.to),
);

// ── Greeting ──────────────────────────────────────────────
const greeting = computed(() => {
  const h = new Date().getHours();
  if (h < 12) return "Good morning";
  if (h < 17) return "Good afternoon";
  return "Good evening";
});

const todayLabel = computed(() =>
  new Date().toLocaleDateString(undefined, {
    weekday: "long",
    month: "long",
    day: "numeric",
  }),
);

// ── Active session ────────────────────────────────────────
const { elapsedSeconds } = useElapsedTimer(
  () => activeSession.value?.started_at ?? null,
);

const activeSessionStats = computed(() => {
  const s = activeSession.value;
  if (!s) return null;
  const totalSets = s.exercises.reduce((n, ex) => n + ex.sets.length, 0);
  const completedSets = s.exercises.reduce(
    (n, ex) => n + ex.sets.filter((set) => set.completed_at).length,
    0,
  );
  return {
    totalSets,
    completedSets,
    exerciseCount: s.exercises.length,
    pct: totalSets > 0 ? Math.round((completedSets / totalSets) * 100) : 0,
  };
});

// ── Weekly stats ──────────────────────────────────────────
const weekStats = computed(() => {
  const sessions = weekSessions.value;
  const durationMs = sessions.reduce((sum, s) => {
    if (!s.ended_at) return sum;
    return sum + (new Date(s.ended_at).getTime() - new Date(s.started_at).getTime());
  }, 0);
  const mins = Math.floor(durationMs / 60000);
  const durationLabel =
    mins === 0
      ? "—"
      : mins < 60
        ? `${mins}m`
        : `${Math.floor(mins / 60)}h ${mins % 60 > 0 ? `${mins % 60}m` : ""}`.trim();
  return {
    workouts: sessions.length,
    sets: sessions.reduce((n, s) => n + s.sets_completed, 0),
    durationLabel,
  };
});

// ── Feature stats ─────────────────────────────────────────
const featureStats = computed(() => {
  const m = weekStatMap.value;
  const stats: { label: string; value: string; key: string }[] = [];

  const reps = m.get("rep") ?? 0;
  if (reps > 0) stats.push({ key: "rep", label: "reps", value: reps.toLocaleString() });

  const weight = m.get("weight") ?? 0;
  if (weight > 0) {
    const formatted = weight >= 1000
      ? `${(weight / 1000).toFixed(1).replace(/\.0$/, "")}t`
      : `${weight % 1 === 0 ? weight : weight.toFixed(1)}kg`;
    stats.push({ key: "weight", label: "lifted", value: formatted });
  }

  const distance = m.get("distance") ?? 0;
  if (distance > 0) {
    const formatted = distance >= 1000
      ? `${(distance / 1000).toFixed(1).replace(/\.0$/, "")}km`
      : `${Math.round(distance)}m`;
    stats.push({ key: "distance", label: "distance", value: formatted });
  }

  return stats;
});

// ── Formatters ────────────────────────────────────────────
function formatDuration(startIso: string, endIso: string | null): string {
  if (!endIso) return "";
  const ms = new Date(endIso).getTime() - new Date(startIso).getTime();
  const mins = Math.floor(ms / 60000);
  if (mins < 60) return `${mins}m`;
  const h = Math.floor(mins / 60);
  const m = mins % 60;
  return m > 0 ? `${h}h ${m}m` : `${h}h`;
}

function relativeDate(iso: string): string {
  const d = new Date(iso);
  const today = new Date();
  const yesterday = new Date(today);
  yesterday.setDate(yesterday.getDate() - 1);
  if (d.toDateString() === today.toDateString()) return "Today";
  if (d.toDateString() === yesterday.toDateString()) return "Yesterday";
  return d.toLocaleDateString(undefined, {
    weekday: "short",
    month: "short",
    day: "numeric",
  });
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString(undefined, {
    hour: "numeric",
    minute: "2-digit",
  });
}
</script>

<template>
  <div class="pb-12 space-y-8">
    <!-- Greeting -->
    <div>
      <p class="text-sm text-muted-foreground">{{ todayLabel }}</p>
      <h1 class="text-3xl font-bold tracking-tight mt-0.5">{{ greeting }}</h1>
    </div>

    <!-- ── Active Workout Card ───────────────────────────── -->
    <div v-if="activeLoading" class="space-y-2">
      <Skeleton class="h-[120px] w-full rounded-2xl" />
    </div>

    <div
      v-else-if="activeSession"
      class="relative overflow-hidden rounded-2xl border border-amber-500/40 bg-gradient-to-br from-amber-500/10 via-orange-500/5 to-transparent p-5 shadow-sm"
    >
      <!-- Subtle glow streak -->
      <div
        class="pointer-events-none absolute -top-10 -right-10 h-40 w-40 rounded-full bg-amber-400/10 blur-3xl"
      />

      <div class="flex items-start justify-between gap-4">
        <div class="flex-1 min-w-0">
          <!-- Status badge -->
          <div class="flex items-center gap-2 mb-2">
            <span class="relative flex h-2 w-2">
              <span
                class="animate-ping absolute inline-flex h-full w-full rounded-full bg-amber-400 opacity-75"
              />
              <span class="relative inline-flex h-2 w-2 rounded-full bg-amber-500" />
            </span>
            <span class="text-xs font-semibold uppercase tracking-wider text-amber-600 dark:text-amber-400">
              In Progress
            </span>
          </div>

          <!-- Workout name -->
          <p class="text-xl font-bold leading-tight truncate">
            {{ activeSession.workout?.name ?? "Active Workout" }}
          </p>

          <!-- Timer + progress -->
          <div class="flex items-center gap-4 mt-2 flex-wrap">
            <span class="flex items-center gap-1.5 text-sm text-muted-foreground tabular-nums">
              <Clock class="w-3.5 h-3.5 shrink-0" />
              {{ formatElapsed(elapsedSeconds) }}
            </span>
            <span
              v-if="activeSessionStats && activeSessionStats.exerciseCount > 0"
              class="flex items-center gap-1.5 text-sm text-muted-foreground"
            >
              <Layers class="w-3.5 h-3.5 shrink-0" />
              {{ activeSessionStats.exerciseCount }} exercise{{ activeSessionStats.exerciseCount !== 1 ? "s" : "" }}
            </span>
            <span
              v-if="activeSessionStats && activeSessionStats.totalSets > 0"
              class="flex items-center gap-1.5 text-sm text-muted-foreground"
            >
              <Flame class="w-3.5 h-3.5 shrink-0" />
              {{ activeSessionStats.completedSets }}/{{ activeSessionStats.totalSets }} sets
            </span>
          </div>

          <!-- Sets progress bar -->
          <div
            v-if="activeSessionStats && activeSessionStats.totalSets > 0"
            class="mt-3 h-1.5 w-full rounded-full bg-amber-200/40 dark:bg-amber-900/40 overflow-hidden"
          >
            <div
              class="h-full rounded-full bg-amber-500 transition-all duration-500"
              :style="{ width: `${activeSessionStats.pct}%` }"
            />
          </div>
        </div>

        <Button
          size="sm"
          class="shrink-0 bg-amber-500 hover:bg-amber-600 text-white border-0 shadow-md"
          @click="router.push('/workouts/active')"
        >
          <Play class="w-3.5 h-3.5 mr-1" />
          Continue
        </Button>
      </div>
    </div>

    <!-- ── Active Plan Card ──────────────────────────────── -->
    <div v-if="planLoading && !activeLoading" class="space-y-2">
      <Skeleton class="h-[100px] w-full rounded-2xl" />
    </div>

    <ActivePlanCard v-else-if="activePlan" :progress="activePlan" />

    <!-- ── This week ─────────────────────────────────────── -->
    <div>
      <div class="flex items-center gap-2 mb-3">
        <TrendingUp class="w-4 h-4 text-muted-foreground" />
        <h2 class="text-sm font-semibold text-muted-foreground uppercase tracking-wider">This week</h2>
      </div>

      <div class="grid grid-cols-3 gap-3">
        <div class="rounded-xl border bg-card p-4 text-center">
          <p class="text-2xl font-bold tabular-nums">{{ weekStats.workouts }}</p>
          <p class="text-xs text-muted-foreground mt-0.5">workout{{ weekStats.workouts !== 1 ? "s" : "" }}</p>
        </div>
        <div class="rounded-xl border bg-card p-4 text-center">
          <p class="text-2xl font-bold tabular-nums">{{ weekStats.sets }}</p>
          <p class="text-xs text-muted-foreground mt-0.5">sets</p>
        </div>
        <div class="rounded-xl border bg-card p-4 text-center">
          <p class="text-2xl font-bold tabular-nums">{{ weekStats.durationLabel }}</p>
          <p class="text-xs text-muted-foreground mt-0.5">active</p>
        </div>
      </div>

      <!-- Feature stats — only shown if there's data -->
      <div v-if="featureStats.length > 0" class="flex flex-wrap gap-2 mt-3">
        <div
          v-for="stat in featureStats"
          :key="stat.key"
          class="flex items-baseline gap-1.5 rounded-lg border bg-card px-3 py-2"
        >
          <span class="text-sm font-bold tabular-nums">{{ stat.value }}</span>
          <span class="text-xs text-muted-foreground">{{ stat.label }}</span>
        </div>
      </div>
    </div>

    <!-- ── Bottom: recent + quick links ─────────────────── -->
    <div class="grid grid-cols-1 lg:grid-cols-5 gap-6">

      <!-- Recent sessions (3/5 width on lg) -->
      <div class="lg:col-span-3">
        <div class="flex items-center justify-between mb-3">
          <h2 class="font-semibold">Recent Sessions</h2>
          <button
            class="text-xs text-muted-foreground hover:text-foreground flex items-center gap-1 transition-colors"
            @click="router.push('/workout-history')"
          >
            View all
            <ArrowRight class="w-3 h-3" />
          </button>
        </div>

        <!-- Loading skeletons -->
        <div v-if="recentLoading" class="space-y-2">
          <Skeleton v-for="i in 3" :key="i" class="h-16 w-full rounded-xl" />
        </div>

        <!-- Empty state -->
        <div
          v-else-if="recentSessions.length === 0"
          class="flex flex-col items-center gap-2 py-10 text-muted-foreground/50 text-center"
        >
          <CalendarDays class="w-7 h-7" />
          <p class="text-sm">No sessions yet. Start your first workout.</p>
        </div>

        <!-- Session list -->
        <div v-else class="space-y-2">
          <button
            v-for="session in recentSessions"
            :key="session.id"
            class="w-full text-left group"
            @click="router.push('/workout-history')"
          >
            <div
              class="flex items-center gap-3 rounded-xl border bg-card px-4 py-3 hover:bg-muted/40 active:scale-[0.99] transition-all duration-100"
            >
              <!-- Icon -->
              <div class="shrink-0 w-8 h-8 rounded-lg bg-muted flex items-center justify-center">
                <Dumbbell class="w-4 h-4 text-muted-foreground" />
              </div>

              <!-- Info -->
              <div class="flex-1 min-w-0">
                <p class="text-sm font-semibold truncate leading-tight">
                  {{ session.workout_name || "Unnamed Workout" }}
                </p>
                <p class="text-xs text-muted-foreground flex items-center gap-1.5 mt-0.5">
                  <span>{{ relativeDate(session.started_at) }}</span>
                  <span class="opacity-40">·</span>
                  <span>{{ formatTime(session.started_at) }}</span>
                  <template v-if="session.ended_at">
                    <span class="opacity-40">·</span>
                    <span>{{ formatDuration(session.started_at, session.ended_at) }}</span>
                  </template>
                </p>
              </div>

              <!-- Pills -->
              <div class="shrink-0 flex items-center gap-1.5">
                <span class="inline-flex items-center gap-1 text-xs text-muted-foreground">
                  <Layers class="w-3 h-3" />{{ session.exercise_count }}
                </span>
                <span class="opacity-30">·</span>
                <span class="inline-flex items-center gap-1 text-xs text-muted-foreground">
                  <Flame class="w-3 h-3" />{{ session.sets_completed }}
                </span>
                <ChevronRight
                  class="w-3.5 h-3.5 text-muted-foreground/30 group-hover:text-muted-foreground ml-1 transition-colors"
                />
              </div>
            </div>
          </button>
        </div>
      </div>

      <!-- Quick links (2/5 width on lg) -->
      <div class="lg:col-span-2">
        <h2 class="font-semibold mb-3">Quick Links</h2>
        <div class="grid grid-cols-2 gap-3">
          <button
            class="group flex flex-col items-center gap-2.5 rounded-xl border bg-card p-4 hover:bg-muted/40 active:scale-[0.98] transition-all duration-100 text-center"
            @click="router.push('/workouts')"
          >
            <div class="w-9 h-9 rounded-lg bg-violet-500/10 flex items-center justify-center group-hover:bg-violet-500/20 transition-colors">
              <Dumbbell class="w-4 h-4 text-violet-500" />
            </div>
            <span class="text-xs font-medium leading-tight">Workouts</span>
          </button>

          <button
            class="group flex flex-col items-center gap-2.5 rounded-xl border bg-card p-4 hover:bg-muted/40 active:scale-[0.98] transition-all duration-100 text-center"
            @click="router.push('/workout-plans')"
          >
            <div class="w-9 h-9 rounded-lg bg-blue-500/10 flex items-center justify-center group-hover:bg-blue-500/20 transition-colors">
              <Target class="w-4 h-4 text-blue-500" />
            </div>
            <span class="text-xs font-medium leading-tight">Plans</span>
          </button>

          <button
            class="group flex flex-col items-center gap-2.5 rounded-xl border bg-card p-4 hover:bg-muted/40 active:scale-[0.98] transition-all duration-100 text-center"
            @click="router.push('/workout-history')"
          >
            <div class="w-9 h-9 rounded-lg bg-emerald-500/10 flex items-center justify-center group-hover:bg-emerald-500/20 transition-colors">
              <BarChart3 class="w-4 h-4 text-emerald-500" />
            </div>
            <span class="text-xs font-medium leading-tight">History</span>
          </button>

          <button
            class="group flex flex-col items-center gap-2.5 rounded-xl border bg-card p-4 hover:bg-muted/40 active:scale-[0.98] transition-all duration-100 text-center"
            @click="router.push('/exercises')"
          >
            <div class="w-9 h-9 rounded-lg bg-amber-500/10 flex items-center justify-center group-hover:bg-amber-500/20 transition-colors">
              <BookOpen class="w-4 h-4 text-amber-500" />
            </div>
            <span class="text-xs font-medium leading-tight">Exercises</span>
          </button>
        </div>

        <!-- All done state (no active session or plan) -->
        <div
          v-if="!activeLoading && !planLoading && !activeSession && !activePlan && weekStats.workouts > 0"
          class="mt-3 rounded-xl border border-emerald-500/20 bg-emerald-500/5 p-3 flex items-center gap-2.5"
        >
          <CheckCircle2 class="w-4 h-4 text-emerald-500 shrink-0" />
          <p class="text-xs text-muted-foreground">
            {{ weekStats.workouts }} workout{{ weekStats.workouts !== 1 ? "s" : "" }} done this week. Keep it up!
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
