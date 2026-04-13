<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from "vue";
import { useRouter } from "vue-router";
import { VueQueryDevtools } from "@tanstack/vue-query-devtools";
import {
  NavigationMenu,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
} from "@/components/ui/navigation-menu";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Sheet, SheetContent } from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { Toaster } from "@/components/ui/sonner";
import { useAuth } from "@/lib/auth/composables/useAuth";
import { useActiveWorkoutSession } from "@/features/workout-session/composables/useActiveWorkoutSession";
import { Dumbbell, Timer } from "lucide-vue-next";

const { user, isAuthenticated, logout, initAuth } = useAuth();
const router = useRouter();
const mobileMenuOpen = ref(false);
const { session: activeSession } = useActiveWorkoutSession();

// Live elapsed timer for the active workout pill
const elapsedSeconds = ref(0);
let elapsedTimerId: ReturnType<typeof setInterval> | null = null;

function tickElapsed() {
  if (!activeSession.value?.started_at) { elapsedSeconds.value = 0; return; }
  elapsedSeconds.value = Math.floor((Date.now() - new Date(activeSession.value.started_at).getTime()) / 1000);
}

function formatElapsed(s: number): string {
  const m = Math.floor(s / 60);
  const h = Math.floor(m / 60);
  if (h > 0) return `${h}h ${(m % 60)}m`;
  return `${m}m`;
}

// Start/stop timer based on active session
import { watch } from "vue";
watch(activeSession, (s) => {
  if (s) {
    tickElapsed();
    if (!elapsedTimerId) elapsedTimerId = setInterval(tickElapsed, 10000);
  } else {
    if (elapsedTimerId) { clearInterval(elapsedTimerId); elapsedTimerId = null; }
    elapsedSeconds.value = 0;
  }
}, { immediate: true });

onUnmounted(() => { if (elapsedTimerId) clearInterval(elapsedTimerId); });

const activeSetsCompleted = computed(() => {
  if (!activeSession.value) return 0;
  return activeSession.value.exercises.flatMap((e) => e.sets).filter((s) => s.completed_at).length;
});

const activeTotalSets = computed(() => {
  if (!activeSession.value) return 0;
  return activeSession.value.exercises.flatMap((e) => e.sets).length;
});

onMounted(() => {
  initAuth();
});

function closeMobileMenu() {
  mobileMenuOpen.value = false;
}

function handleLogout() {
  logout();
  closeMobileMenu();
}

const navLinks = [
  { to: "/", labelKey: "nav.home" },
  { to: "/workouts", labelKey: "nav.workouts" },
  { to: "/workout-plans", labelKey: "nav.workoutPlans" },
  { to: "/workout-plans/active", labelKey: "workoutPlans.progress.title" },
  { to: "/exercises", labelKey: "nav.exercises" },
  { to: "/workout-history", labelKey: "nav.workoutHistory" },
];
</script>

<template>
  <div class="h-dvh flex flex-col overflow-hidden">
    <!-- Header -->
    <header class="shrink-0 px-4 md:px-8 py-3 border-b flex items-center justify-between gap-2">
      <!-- Desktop nav -->
      <NavigationMenu class="hidden md:flex">
        <NavigationMenuList>
          <NavigationMenuItem v-for="link in navLinks" :key="link.to">
            <NavigationMenuLink as-child>
              <router-link :to="link.to" active-class="text-accent-foreground bg-accent">
                {{ $t(link.labelKey) }}
              </router-link>
            </NavigationMenuLink>
          </NavigationMenuItem>
        </NavigationMenuList>
      </NavigationMenu>

      <!-- Mobile: app name -->
      <router-link to="/" class="md:hidden font-semibold text-lg">Liift</router-link>

      <!-- Right side: active workout pill + auth + hamburger -->
      <div class="flex items-center gap-2 ml-auto">

      <!-- Active workout pill -->
      <router-link
        v-if="activeSession"
        to="/workouts/active"
        class="flex items-center gap-2 px-3 py-1.5 rounded-full bg-green-500/10 border border-green-500/30 text-green-700 dark:text-green-400 hover:bg-green-500/20 transition-colors text-sm font-medium shrink-0"
      >
        <span class="relative flex size-2 shrink-0">
          <span class="absolute inline-flex h-full w-full rounded-full bg-green-500 opacity-75 animate-ping" />
          <span class="relative inline-flex size-2 rounded-full bg-green-500" />
        </span>
        <Dumbbell class="size-3.5 shrink-0" />
        <!-- Name (sm+) -->
        <span class="hidden sm:inline font-semibold truncate max-w-28 md:max-w-36">
          {{ activeSession.workout?.name ?? $t("workoutSession.activeWorkout") }}
        </span>
        <!-- Timer (always) -->
        <span class="flex items-center gap-0.5 text-xs opacity-80 tabular-nums">
          <Timer class="size-3 shrink-0" />
          {{ formatElapsed(elapsedSeconds) }}
        </span>
        <!-- Sets ratio (sm+) -->
        <span class="hidden sm:inline text-xs opacity-80 tabular-nums">
          · {{ activeSetsCompleted }}/{{ activeTotalSets }} sets
        </span>
      </router-link>
        <!-- Desktop auth -->
        <template v-if="isAuthenticated && user">
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <Button variant="ghost" class="hidden md:flex gap-2">
                <span>{{ user.username }}</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>
                <div class="flex flex-col">
                  <span>{{ user.username }}</span>
                  <span v-if="user.email" class="text-xs text-muted-foreground">{{ user.email }}</span>
                </div>
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem @click="logout">{{ $t("auth.logOut") }}</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </template>
        <template v-else>
          <div class="hidden md:flex items-center gap-2">
            <Button variant="ghost" as-child>
              <router-link to="/login">{{ $t("auth.login") }}</router-link>
            </Button>
            <Button as-child>
              <router-link to="/register">{{ $t("auth.register") }}</router-link>
            </Button>
          </div>
        </template>

        <!-- Mobile hamburger -->
        <Button variant="ghost" size="icon" class="md:hidden" @click="mobileMenuOpen = true">
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="4" y1="6" x2="20" y2="6" />
            <line x1="4" y1="12" x2="20" y2="12" />
            <line x1="4" y1="18" x2="20" y2="18" />
          </svg>
        </Button>
      </div>
    </header>

    <!-- Mobile sheet menu -->
    <Sheet v-model:open="mobileMenuOpen">
      <SheetContent side="left" class="flex flex-col gap-0 p-0 w-72">
        <nav class="flex flex-col py-4">
          <router-link
            v-for="link in navLinks"
            :key="link.to"
            :to="link.to"
            class="px-6 py-3 text-sm font-medium hover:bg-accent hover:text-accent-foreground transition-colors"
            active-class="bg-accent text-accent-foreground"
            @click="closeMobileMenu"
          >
            {{ $t(link.labelKey) }}
          </router-link>
        </nav>
        <div class="mt-auto border-t px-6 py-4">
          <template v-if="isAuthenticated && user">
            <p class="text-sm font-medium">{{ user.username }}</p>
            <p v-if="user.email" class="text-xs text-muted-foreground mb-3">{{ user.email }}</p>
            <Button variant="outline" class="w-full" @click="handleLogout">{{ $t("auth.logOut") }}</Button>
          </template>
          <template v-else>
            <div class="flex flex-col gap-2">
              <Button as-child @click="closeMobileMenu">
                <router-link to="/register">{{ $t("auth.register") }}</router-link>
              </Button>
              <Button variant="outline" as-child @click="closeMobileMenu">
                <router-link to="/login">{{ $t("auth.login") }}</router-link>
              </Button>
            </div>
          </template>
        </div>
      </SheetContent>
    </Sheet>

    <!-- Main content -->
    <main class="flex-1 min-h-0 overflow-y-auto">
      <div class="max-w-7xl mx-auto w-full px-4 py-6 md:px-8 md:py-8">
        <router-view />
      </div>
    </main>

    <VueQueryDevtools />
    <Toaster />
  </div>
</template>
